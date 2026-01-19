package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/encoding/protojson"

	{{.Web.APIImportPath}}
)

const (
	topicName  = "b-dms.bookmark.events.v1"
	brokerAddr = "localhost:9092"
)

func main() {
	// Set up structured JSON logging as the default.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	opts := []kgo.Opt{
		kgo.SeedBrokers(brokerAddr),
		kgo.AllowAutoTopicCreation(),
	}
	client, err := kgo.NewClient(opts...)
	if err != nil {
		slog.Error("failed to create kafka client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	ctx := context.Background()
	var wg sync.WaitGroup

	testCases := []struct {
		name  string
		event *{{.D.APIAlias}}.BookmarkEventEnvelope
	}{
		{"CreatedEvent", newCreatedEvent()},
		{"UpdatedEvent", newUpdatedEvent()},
		{"DeletedEvent", newDeletedEvent()},
	}

	for _, tc := range testCases {
		// Generate 16 bytes (32 hex chars) for Trace ID.
		traceID, err := generateRandomHex(16)
		if err != nil {
			slog.Error("failed to generate traceID", "error", err)
			continue
		}

		// Generate 8 bytes (16 hex chars) for Span ID.
		spanID, err := generateRandomHex(8)
		if err != nil {
			slog.Error("failed to generate spanID", "error", err)
			continue
		}

		// Construct W3C standard Traceparent string: version-traceid-spanid-flags.
		// Example: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
		traceParent := fmt.Sprintf("00-%s-%s-01", traceID, spanID)

		// Increment WaitGroup only after successful setup to avoid deadlocks on continue.
		wg.Add(1)
		go func(name string, evt *{{.D.APIAlias}}.BookmarkEventEnvelope, traceStr string) {
			defer wg.Done()
			sendEvent(ctx, client, name, evt, traceStr)
		}(tc.name, tc.event, traceParent)
	}

	wg.Wait()
	slog.Info("all test events sent successfully")
}

// sendEvent serializes the protobuf message and produces it to Kafka synchronously.
func sendEvent(ctx context.Context, client *kgo.Client, name string, event *{{.D.APIAlias}}.BookmarkEventEnvelope, traceParent string) {
	marshaler := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}

	jsonBytes, err := marshaler.Marshal(event)
	if err != nil {
		slog.Error("serialization failed", "name", name, "error", err)
		return
	}

	record := &kgo.Record{
		Topic: topicName,
		Key:   []byte(event.EventID),
		Value: jsonBytes,
		Headers: []kgo.RecordHeader{
			{Key: "traceparent", Value: []byte(traceParent)},
			{Key: "content-type", Value: []byte("application/json")},
		},
	}

	slog.Info("sending event",
		"name", name,
		"trace_parent", traceParent,
		"payload_size", len(jsonBytes),
	)

	results := client.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		slog.Error("send failed", "name", name, "error", err)
		return
	}

	// Since we only sent one message, the result is at index 0.
	rec := results[0].Record
	slog.Info("send successful",
		"name", name,
		"partition", rec.Partition,
		"offset", rec.Offset,
	)
}

// newCreatedEvent constructs a sample bookmark creation event.
func newCreatedEvent() *{{.D.APIAlias}}.BookmarkEventEnvelope {
	return &{{.D.APIAlias}}.BookmarkEventEnvelope{
		EventID:   uuid.NewString(),
		Timestamp: time.Now().UnixMilli(),
		Type:      {{.D.APIAlias}}.BookmarkEventType_BOOKMARK_EVENT_TYPE_CREATED,
		Payload: &{{.D.APIAlias}}.BookmarkEventEnvelope_Created{
			Created: &{{.D.APIAlias}}.CreateBookmarkRequest{},
		},
	}
}

// newUpdatedEvent constructs a sample bookmark update event.
func newUpdatedEvent() *{{.D.APIAlias}}.BookmarkEventEnvelope {
	return &{{.D.APIAlias}}.BookmarkEventEnvelope{
		EventID:   uuid.NewString(),
		Timestamp: time.Now().UnixMilli(),
		Type:      {{.D.APIAlias}}.BookmarkEventType_BOOKMARK_EVENT_TYPE_UPDATED,
		Payload: &{{.D.APIAlias}}.BookmarkEventEnvelope_Updated{
			Updated: &{{.D.APIAlias}}.UpdateBookmarkRequest{
				BookmarkID: "bm-2024-001",
			},
		},
	}
}

// newDeletedEvent constructs a sample bookmark deletion event.
func newDeletedEvent() *{{.D.APIAlias}}.BookmarkEventEnvelope {
	return &{{.D.APIAlias}}.BookmarkEventEnvelope{
		EventID:   uuid.NewString(),
		Timestamp: time.Now().UnixMilli(),
		Type:      {{.D.APIAlias}}.BookmarkEventType_BOOKMARK_EVENT_TYPE_DELETED,
		Payload: &{{.D.APIAlias}}.BookmarkEventEnvelope_Deleted{
			Deleted: &{{.D.APIAlias}}.DeleteBookmarkRequest{
				BookmarkID: "bm-2024-001",
			},
		},
	}
}

// generateRandomHex generates a random hexadecimal string of length 2*n.
func generateRandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
