package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog" // Use standard library slog for structured logging
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	// Assuming you have the protobuf generated files
	{{.Web.APIImportPath}}
)

// Message types (matching server constants for clarity)
const (
	PingMessageType      = "ws.ping"           // Message type for client-initiated ping.
	CountMessageType     = "ws.count"          // Message type for client-initiated count request.
	PongMessageType      = "ws.pong"           // Message type for server-initiated pong response.
	ErrorMessageType     = "ws.error"          // Message type for server-initiated error response.
	CountResponseMessage = "ws.count_response" // Message type for server-initiated count response.
)

// client represents a WebSocket client for testing purposes.
type client struct {
	conn         *websocket.Conn
	serverURL    string
	userID       string
	pingInterval time.Duration
	done         chan struct{}  // Channel to signal goroutines to shut down.
	wg           sync.WaitGroup // WaitGroup to wait for all goroutines to finish.
	writeMutex   sync.Mutex     // Mutex to protect concurrent writes to the WebSocket connection.
	logger       *slog.Logger   // Structured logger for this client instance.
}

// NewClient creates a new WebSocket test client instance.
// It initializes the client with the server URL, a unique user ID, and ping interval.
func NewClient(serverURL, userID string, pingInterval time.Duration) *client {
	return &client{
		serverURL:    serverURL,
		userID:       userID,
		pingInterval: pingInterval,
		done:         make(chan struct{}),
		logger:       slog.With("userID", userID), // Logger with client-specific context.
	}
}

// Connect establishes a WebSocket connection to the server.
// It constructs the URL, optionally adds the user ID as a query parameter,
// and attempts to dial the WebSocket server.
func (c *client) Connect() error {
	u, err := url.Parse(c.serverURL)
	if err != nil {
		c.logger.Error("Failed to parse server URL", "url", c.serverURL, "error", err)
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Add user_id as query parameter if provided.
	if c.userID != "" {
		q := u.Query()
		q.Set("user_id", c.userID)
		u.RawQuery = q.Encode()
	}

	c.logger.Info("Attempting to connect to WebSocket server", "targetURL", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		c.logger.Error("WebSocket dial failed", "error", err)
		return fmt.Errorf("dial failed: %w", err)
	}

	c.conn = conn
	c.logger.Info("Connected successfully to WebSocket server")
	return nil
}

// Start initiates the client's operations by launching goroutines for
// reading incoming messages, sending pings, and requesting counts.
// It blocks until all these goroutines have completed.
func (c *client) Start() {
	// Add 3 to WaitGroup for readPump, pingPump, and countPump goroutines.
	c.wg.Add(3)

	go c.readPump()
	go c.pingPump()
	// go c.countPump()

	// Wait for all client goroutines to finish. This call will block.
	c.wg.Wait()
	c.logger.Info("Client operations stopped")
}

// Stop gracefully closes the client connection and signals all its goroutines to terminate.
// It attempts to send a close message to the server before closing the underlying connection.
func (c *client) Stop() {
	c.logger.Info("Stopping client operations")
	close(c.done) // Signal goroutines to stop.

	if c.conn != nil {
		// Attempt to gracefully close the connection by sending a close message.
		c.writeMutex.Lock() // Acquire lock before writing close message
		err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.writeMutex.Unlock() // Release lock
		if err != nil {
			// Log error but proceed to close, as it might be a broken pipe already.
			c.logger.Error("Failed to send close message to WebSocket server", "error", err)
		}
		c.conn.Close() // Close the underlying network connection.
		c.conn = nil   // Mark connection as closed to prevent further use.
	}
}

// readPump handles incoming messages from the WebSocket server.
// It continuously reads messages, resets the read deadline on pong,
// and dispatches messages to the appropriate handler.
func (c *client) readPump() {
	defer c.wg.Done()

	c.conn.SetReadLimit(1024)                                // Maximum message size in bytes.
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // Initial read deadline.
	c.conn.SetPongHandler(func(string) error {
		// Reset read deadline on receiving a pong message to keep the connection alive.
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-c.done:
			c.logger.Info("Read pump exiting due to shutdown signal")
			return
		default:
			// ReadJSON is safe for concurrent reads.
			_, messageBytes, err := c.conn.ReadMessage()
			if err != nil {
				return
			}

			var msg v1.WSMessage
			// Use protojson to unmarshal the received bytes into the protobuf message.
			// protojson.Unmarshal will correctly handle oneof fields.
			if err := protojson.Unmarshal(messageBytes, &msg); err != nil {
				c.logger.Error("Failed to unmarshal protobuf message from JSON", "error", err, "rawMessage", string(messageBytes))
				// Decide if this is a critical error leading to client shutdown
				c.Stop()
				return
			}

			c.handleMessage(&msg)
		}
	}
}

// handleMessage processes messages received from the server, dispatching them
// to appropriate handlers based on their type.
func (c *client) handleMessage(msg *v1.WSMessage) {
	switch msg.Type {
	case PongMessageType:
		c.handlePong(msg)
	case CountResponseMessage:
		c.handleCountResponse(msg)
	case ErrorMessageType:
		c.handleError(msg)
	default:
		c.logger.Warn("Received unknown message type", "messageID", msg.ID, "messageType", msg.Type)
	}
}

// handlePong processes pong responses from the server.
// It logs details from the PingResponse payload.
func (c *client) handlePong(msg *v1.WSMessage) {
	pong := msg.GetPingResponse() // Assuming PingResponse is used for pong data.
	if pong != nil {
		c.logger.Info("Received PONG response",
			"messageID", msg.ID,
			"sequence", pong.Sequence,
			"processingTimeUS", pong.ProcessingTimeUS)
	} else {
		c.logger.Warn("Received PONG message with empty PingResponse payload", "messageID", msg.ID)
	}
}

// handleCountResponse processes count responses from the server.
// ASSUMPTION: v1.WSMessage has a `GetCountResponse()` method.
func (c *client) handleCountResponse(msg *v1.WSMessage) {
	countResp := msg.GetPingResponse() // Using PingResponse for count data
	if countResp != nil {
		c.logger.Info("Received Client Count response", "messageID", msg.ID, "count", countResp.Sequence, "processingTimeUS", countResp.ProcessingTimeUS)
	} else {
		c.logger.Warn("Received Count Response message with empty CountResponse payload", "messageID", msg.ID)
	}
}

func (c *client) handleError(msg *v1.WSMessage) {
	errorResp := msg.GetErrorResponse()
	if errorResp != nil {
		c.logger.Error("Received server error",
			"messageID", msg.ID,
			"errorCode", errorResp.Code,
			"errorReason", errorResp.Reason,
			"errorMessage", errorResp.Message)
	} else {
		c.logger.Error("Received error message with empty ErrorResponse payload", "messageID", msg.ID)
	}
}

// pingPump sends ping messages to the server at regular intervals.
func (c *client) pingPump() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()

	var sequence uint64 = 1

	for {
		select {
		case <-c.done:
			c.logger.Info("Ping pump exiting due to shutdown signal")
			return
		case <-ticker.C:
			c.sendPing(sequence)
			sequence++
		}
	}
}

// countPump periodically requests the current client count from the server.
// The request interval is currently hardcoded to 10 seconds.
func (c *client) countPump() {
	defer c.wg.Done()

	ticker := time.NewTicker(10 * time.Second) // Request count every 10 seconds.
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			c.logger.Info("Count pump exiting due to shutdown signal")
			return
		case <-ticker.C:
			c.sendCountRequest()
		}
	}
}

// sendPing constructs and sends a ping message to the server.
// It includes a unique message ID and sequence number.
func (c *client) sendPing(sequence uint64) {
	pingReq := &v1.PingRequest{
		Sequence: sequence,
	}

	msg := &v1.WSMessage{
		Type: PingMessageType,
		// Include userID in message ID for better tracing across multiple clients.
		ID: fmt.Sprintf("ping_%s_%d_%d", c.userID, sequence, time.Now().UnixNano()),
		Payload: &v1.WSMessage_PingRequest{ // Assuming PingRequest_ is the oneof field name.
			PingRequest: pingReq,
		},
		Timestamp: timestamppb.Now(),
	}

	c.sendMessage(msg)
	c.logger.Debug("Sent PING request", "messageID", msg.ID, "sequence", sequence)
}

// sendCountRequest constructs and sends a count request message to the server.
// It includes a unique message ID.
func (c *client) sendCountRequest() {
	// Assuming CountRequest is a simple empty message or has minimal fields.
	// If CountRequest has specific fields, they should be populated here.
	msg := &v1.WSMessage{
		Type:      CountMessageType,
		ID:        fmt.Sprintf("count_%s_%d", c.userID, time.Now().UnixNano()), // Include userID for tracing.
		Timestamp: timestamppb.Now(),
	}

	c.sendMessage(msg)
	c.logger.Debug("Sent COUNT request", "messageID", msg.ID)
}

// sendMessage sends a protobuf message to the WebSocket server.
// It ensures that only one goroutine writes to the connection at a time using a mutex.
func (c *client) sendMessage(msg *v1.WSMessage) {
	c.writeMutex.Lock()         // Acquire lock before writing.
	defer c.writeMutex.Unlock() // Release lock after writing.

	// Set a write deadline to prevent writes from blocking indefinitely.
	err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		c.logger.Error("Failed to set write deadline for WebSocket connection", "error", err)
		c.Stop() // Initiate client shutdown if write deadline cannot be set.
		return
	}

	// Use protojson to marshal the protobuf message into JSON bytes.
	// protojson.Marshal will correctly handle oneof fields.
	messageBytes, err := protojson.Marshal(msg)
	if err != nil {
		c.logger.Error("Failed to marshal protobuf message to JSON",
			"messageType", msg.Type, "messageID", msg.ID, "error", err)
		c.Stop()
		return
	}

	// Send the JSON bytes as a text message over WebSocket.
	if err := c.conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
		c.logger.Error("Failed to write JSON message to WebSocket connection",
			"messageType", msg.Type, "messageID", msg.ID, "error", err)
		c.Stop()
		return
	}
}

// sendInvalidMessage sends a message with an intentionally invalid type for testing server error handling.
func (c *client) sendInvalidMessage() {
	msg := &v1.WSMessage{
		Type:      "invalid.message.type",
		ID:        fmt.Sprintf("invalid_%s_%d", c.userID, time.Now().UnixNano()),
		Timestamp: timestamppb.Now(),
	}

	c.sendMessage(msg)
	c.logger.Info("Sent INVALID message for server error handling test", "messageID", msg.ID)
}

func main() {
	// Define command-line flags for client configuration.
	var (
		serverURL    = flag.String("url", "ws://localhost:8080/v1/ws", "WebSocket server URL")
		userIDPrefix = flag.String("user-prefix", "test_user", "Prefix for generated user IDs. If empty, auto-generated IDs are used.")
		pingInterval = flag.Duration("ping", 5*time.Second, "Interval at which ping messages are sent to the server.")
		numClients   = flag.Int("clients", 1, "Number of concurrent WebSocket clients to simulate.")
		testDuration = flag.Duration("duration", 30*time.Second, "Total duration for the client test run.")
		testInvalid  = flag.Bool("invalid", false, "If true, one client will send an invalid message after 5 seconds to test server error handling.")
	)
	flag.Parse()

	// Configure default structured logger for the main function.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug, // Adjust logging level as needed (e.g., slog.LevelInfo for production)
	})))

	slog.Info("Starting WebSocket client test",
		"serverURL", *serverURL,
		"numClients", *numClients,
		"pingInterval", *pingInterval,
		"testDuration", *testDuration,
		"testInvalidMessage", *testInvalid,
	)

	// Context for graceful shutdown, listening for OS interrupt signals.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel() // Ensure cancel function is called to release resources.

	// Use a WaitGroup to ensure all clients are initialized and started before proceeding.
	var clientSetupWg sync.WaitGroup
	clients := make([]*client, *numClients) // Pre-allocate slice for clients.

	for i := 0; i < *numClients; i++ {
		clientSetupWg.Add(1)
		go func(i int) {
			defer clientSetupWg.Done()

			// Generate a unique user ID for each client.
			currentUserID := *userIDPrefix
			if currentUserID == "" {
				currentUserID = fmt.Sprintf("auto_gen_user_%d", i+1)
			} else {
				currentUserID = fmt.Sprintf("%s_%d", currentUserID, i+1)
			}

			cli := NewClient(*serverURL, currentUserID, *pingInterval)
			clients[i] = cli // Store the client instance.

			if err := cli.Connect(); err != nil {
				slog.Error("Failed to connect client, initiating graceful shutdown",
					"clientIndex", i+1, "userID", currentUserID, "error", err)
				cancel() // Signal all operations to stop on first client connection failure.
				return
			}
			go cli.Start() // Start client operations in a new goroutine.
		}(i)
	}
	clientSetupWg.Wait() // Wait for all clients to complete their connection and start attempts.

	// Check if the context was canceled during client setup (e.g., connection failure or interrupt).
	select {
	case <-ctx.Done():
		slog.Warn("Test interrupted or client setup failed. Shutting down.", "reason", ctx.Err())
		// Stop any clients that might have connected before the cancellation.
		for _, cli := range clients {
			if cli != nil {
				cli.Stop()
			}
		}
		os.Exit(1) // Exit with error code if setup failed.
	default:
		// Continue if all clients started successfully.
	}

	// If requested, send an invalid message using the first client after a delay.
	if *testInvalid && *numClients > 0 && clients[0] != nil {
		go func() {
			select {
			case <-ctx.Done():
				return // Do not send if the context is already cancelled.
			case <-time.After(5 * time.Second):
				slog.Info("Triggering invalid message send for testing server error handling", "targetClient", clients[0].userID)
				clients[0].sendInvalidMessage()
			}
		}()
	}

	// Wait for the specified test duration or an interrupt signal.
	select {
	case <-time.After(*testDuration):
		slog.Info("Test duration completed. Initiating graceful shutdown.")
	case <-ctx.Done():
		slog.Info("Interrupt signal received. Initiating graceful shutdown.", "signal", ctx.Err())
	}

	// Initiate clean shutdown for all active clients.
	slog.Info("Shutting down all active clients...")
	var clientShutdownWg sync.WaitGroup
	for i, cli := range clients {
		if cli != nil {
			clientShutdownWg.Add(1)
			go func(index int, clientToStop *client) {
				defer clientShutdownWg.Done()
				slog.Debug("Stopping client", "clientIndex", index+1, "userID", clientToStop.userID)
				clientToStop.Stop()
			}(i, cli)
		}
	}
	clientShutdownWg.Wait() // Wait for all clients to finish stopping.

	slog.Info("WebSocket client test completed.")
}
