package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/onexstack/onexstack/pkg/logger"                 // Keep custom logger interface
	sloglogger "github.com/onexstack/onexstack/pkg/logger/slog" // Keep custom slog implementation
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// HandlerFunc defines the function signature for processing records of a specific topic.
type HandlerFunc func(ctx context.Context, record *kgo.Record) error

// Engine is a generic Kafka consumer engine responsible for polling messages
// and dispatching them to registered topic handlers.
type Engine struct {
	client    *kgo.Client
	tracer    trace.Tracer
	handlers  map[string]HandlerFunc
	mu        sync.RWMutex   // Protects the handlers map from concurrent access
	logger    logger.Logger  // Retain custom logger.Logger interface
	wg        sync.WaitGroup // Used to wait for all in-flight processing goroutines to finish.
	isRunning bool           // Indicates if the engine's polling loop is active
	stopChan  chan struct{}  // Channel to signal graceful shutdown
}

// Option defines the functional option pattern for configuring the Engine.
type Option func(*Engine)

// WithLogger allows providing a custom logger.Logger.
func WithLogger(customLogger logger.Logger) Option {
	return func(e *Engine) {
		if customLogger != nil {
			e.logger = customLogger
		}
	}
}

// WithTracer allows providing a custom OTEL Tracer.
func WithTracer(tracer trace.Tracer) Option {
	return func(e *Engine) {
		if tracer != nil {
			e.tracer = tracer
		}
	}
}

// NewEngine initializes the engine by creating a new internal Kafka client.
// This is the standard way to start if an existing client is not provided.
func NewEngine(brokers []string, groupID string, opts ...Option) (*Engine, error) {
	e := initEngine(opts...) // Initialize base engine with defaults

	// Configure Kafka client options
	kgoOpts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.DisableAutoCommit(), // Manual offset management is required when disabled
		kgo.OnPartitionsRevoked(func(ctx context.Context, _ *kgo.Client, _ map[string][]int32) {
			e.logger.Info("Kafka partitions revoked", "context_canceled", errors.Is(ctx.Err(), context.Canceled))
		}),
		kgo.OnPartitionsAssigned(func(ctx context.Context, _ *kgo.Client, _ map[string][]int32) {
			e.logger.Info("Kafka partitions assigned", "context_canceled", errors.Is(ctx.Err(), context.Canceled))
		}),
	}

	client, err := kgo.NewClient(kgoOpts...)
	if err != nil {
		e.logger.Error("Failed to create Kafka client", "error", err)
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}
	e.client = client

	return e, nil
}

// NewEngineFromClient initializes the engine using an existing Kafka client.
// Use this if there is a need to share a client instance or require advanced client
// configuration not covered by NewEngine.
func NewEngineFromClient(client *kgo.Client, opts ...Option) (*Engine, error) {
	if client == nil {
		return nil, errors.New("kafka client cannot be nil")
	}

	e := initEngine(opts...)
	e.client = client
	return e, nil
}

// initEngine helps initializing the Engine structure with default values and applying options.
func initEngine(opts ...Option) *Engine {
	e := &Engine{
		handlers: make(map[string]HandlerFunc),
		logger:   sloglogger.Default(),                 // Retain custom sloglogger.Default()
		tracer:   otel.Tracer("kafka-consumer-engine"), // Default OpenTelemetry tracer
		stopChan: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Register dynamically registers a topic and its handler function with the engine.
func (e *Engine) Register(topic string, handler HandlerFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.handlers[topic]; exists {
		e.logger.Warn("Overwriting handler for topic", "topic", topic)
	}
	e.handlers[topic] = handler

	// Dynamically add the topic to the subscription list.
	// AddConsumeTopics is thread-safe and can be called while the client is running.
	e.client.AddConsumeTopics(topic)
	e.logger.Info("Registered handler for topic", "topic", topic)
}

// Start initiates the message consumption loop. It is a blocking call
// that continues until the provided context is canceled or Stop is called.
// It returns an error if the polling loop exits due to an unrecoverable issue.
func (e *Engine) Start(ctx context.Context) error {
	e.logger.Info("Kafka consumer engine started polling for messages...")

	e.isRunning = true
	defer func() {
		e.isRunning = false
		// Ensure stopChan is closed only once after the polling loop exits.
		// This prevents panics if Stop is called multiple times.
		select {
		case <-e.stopChan:
			// Already closed or signal received, do nothing
		default:
			close(e.stopChan) // Signal that the polling loop has stopped
		}
	}()

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("Polling stopped", "reason", "context canceled")
			return nil
		case <-e.stopChan: // Manual stop signal
			e.logger.Info("Polling stopped", "reason", "manual stop requested")
			return nil
		default:
			// Continue polling
		}

		fetches := e.client.PollFetches(ctx)

		if err := fetches.Err(); err != nil {
			// kgo.ErrClientClosed is usually a clean shutdown indicator.
			if errors.Is(err, kgo.ErrClientClosed) {
				e.logger.Info("Kafka client closed, stopping polling")
				return nil
			}
			// Context canceled is also a clean shutdown.
			if errors.Is(err, context.Canceled) {
				e.logger.Info("Polling context canceled, stopping polling")
				return nil
			}
			e.logger.Error("Kafka poll error", "error", err)
			time.Sleep(100 * time.Millisecond) // Simple backoff for transient errors
			continue
		}

		// Iterate over records and dispatch to handlers in separate goroutines
		fetches.EachRecord(func(r *kgo.Record) {
			e.wg.Add(1)
			go func(record *kgo.Record) {
				defer e.wg.Done()
				// Pass the main context, it will be augmented by tracing in process method
				e.process(ctx, record)
			}(r)
		})

		// Offset commit logic:
		// When DisableAutoCommit() is used, the application is responsible for committing offsets.
		// kgo.CommitUncommittedOffsets commits all offsets for records that have been fetched
		// but whose offsets have not yet been committed.
		// WARNING: This call here commits offsets before individual record processing might be complete.
		// This can lead to data loss if a record fails processing AFTER its offset is committed,
		// or duplicate processing if the consumer crashes before a batch of records is fully processed
		// but their offsets are committed.
		// For exactly-once or at-least-once processing, a more robust commit strategy is needed,
		// e.g., committing individual record offsets after successful processing,
		// or batch committing only when all records in a batch are successfully processed.
		if err := e.client.CommitUncommittedOffsets(ctx); err != nil {
			if !errors.Is(err, kgo.ErrClientClosed) {
				e.logger.Error("Failed to commit uncommitted Kafka offsets", "error", err)
			}
		}
	}
}

// Stop gracefully shuts down the engine. It closes the Kafka client
// and waits for all in-flight message processing goroutines to complete.
func (e *Engine) Stop() {
	e.logger.Info("Initiating Kafka consumer engine shutdown...")

	// Signal the polling loop to stop if it's running
	// Use a select to avoid blocking if the channel is already closed or
	// if the polling loop has already stopped.
	if e.isRunning {
		select {
		case e.stopChan <- struct{}{}:
			// Signal sent successfully
		default:
			// Channel already closed or no receiver, polling loop likely stopped
		}
		// Wait for the polling loop to actually stop and close its stopChan
		// This ensures Start's defer block runs and closes its side.
		// If stopChan was already closed, this will immediately return.
		<-e.stopChan
	}

	if e.client != nil {
		e.client.Close() // This also signals kgo.ErrClientClosed to PollFetches
		e.logger.Info("Kafka client closed")
	}

	e.logger.Info("Waiting for all in-flight message handlers to finish...")
	e.wg.Wait() // Wait for all goroutines started by process to finish
	e.logger.Info("Kafka consumer engine stopped gracefully.")
}

// process handles the execution of a single Kafka record's handler function.
// It extracts OpenTelemetry trace context, creates a new span, and
// logs any errors that occur during handler execution.
func (e *Engine) process(ctx context.Context, r *kgo.Record) {
	e.mu.RLock()
	handler, ok := e.handlers[r.Topic]
	e.mu.RUnlock()

	if !ok {
		e.logger.Debug("No handler registered for topic, skipping record", "topic", r.Topic, "partition", r.Partition, "offset", r.Offset)
		return // No handler, simply return
	}

	// Extract OpenTelemetry trace context from Kafka headers
	carrier := propagation.MapCarrier{}
	for _, h := range r.Headers {
		carrier[h.Key] = string(h.Value)
	}
	parentCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)

	// Create a new span for processing this Kafka record
	spanCtx, span := e.tracer.Start(parentCtx, fmt.Sprintf("kafka.receive.%s", r.Topic),
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.operation", "receive"),
			attribute.String("messaging.destination_kind", "topic"),
			attribute.String("messaging.destination_name", r.Topic),
		),
	)
	defer span.End()

	// Execute the topic handler
	if err := handler(spanCtx, r); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		e.logger.Error("Kafka record handler failed", "topic", r.Topic, "partition", r.Partition, "offset", r.Offset, "error", err)
	} else {
		// Optionally, log successful processing for debugging/auditing
		e.logger.Debug("Kafka record handled successfully", "topic", r.Topic, "partition", r.Partition, "offset", r.Offset)
	}
}
