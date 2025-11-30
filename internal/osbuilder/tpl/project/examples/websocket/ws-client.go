package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/types/known/timestamppb"

	// Assuming you have the protobuf generated files
	v1 "github.com/onexstack/b-dms/pkg/api/apiserver/v1"
)

// Message types (matching server constants)
const (
	MessageTypePing          = "ws.ping"
	MessageTypeCount         = "ws.count"
	MessageTypePong          = "ws.pong"
	MessageTypeError         = "ws.error"
	MessageTypeCountResponse = "ws.count_response"
)

// Client represents a WebSocket client for testing
type Client struct {
	conn         *websocket.Conn
	url          string
	userID       string
	pingInterval time.Duration
	done         chan struct{}
	wg           sync.WaitGroup
}

// NewClient creates a new WebSocket test client
func NewClient(serverURL, userID string, pingInterval time.Duration) *Client {
	return &Client{
		url:          serverURL,
		userID:       userID,
		pingInterval: pingInterval,
		done:         make(chan struct{}),
	}
}

// Connect establishes WebSocket connection to the server
func (c *Client) Connect() error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	// Add user_id as query parameter if provided
	if c.userID != "" {
		q := u.Query()
		q.Set("user_id", c.userID)
		u.RawQuery = q.Encode()
	}

	log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial failed: %v", err)
	}

	c.conn = conn
	log.Printf("Connected successfully as user: %s", c.userID)
	return nil
}

// Start begins the client operations (reading, writing, pinging)
func (c *Client) Start() {
	c.wg.Add(3)

	// Start message reader
	go c.readPump()

	// Start ping sender
	go c.pingPump()

	// Start count requester
	go c.countPump()

	c.wg.Wait()
}

// Stop gracefully closes the client connection
func (c *Client) Stop() {
	log.Printf("Stopping client...")
	close(c.done)
	if c.conn != nil {
		c.conn.Close()
	}
}

// readPump handles incoming messages from the server
func (c *Client) readPump() {
	defer c.wg.Done()
	defer c.conn.Close()

	c.conn.SetReadLimit(1024)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-c.done:
			return
		default:
			var msg v1.WSMessage
			err := c.conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error: %v", err)
				}
				return
			}
			c.handleMessage(&msg)
		}
	}
}

// handleMessage processes messages received from the server
func (c *Client) handleMessage(msg *v1.WSMessage) {
	switch msg.Type {
	case MessageTypePong:
		c.handlePong(msg)
	case MessageTypeCountResponse:
		c.handleCountResponse(msg)
	case MessageTypeError:
		c.handleError(msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handlePong processes pong responses from the server
func (c *Client) handlePong(msg *v1.WSMessage) {
	pong := msg.GetPingResponse()
	if pong != nil {
		log.Printf("Received PONG - ID: %s, Sequence: %d, Processing Time: %d μs",
			msg.ID, pong.Sequence, pong.ProcessingTimeUS)
	}
}

// handleCountResponse processes count responses from the server
func (c *Client) handleCountResponse(msg *v1.WSMessage) {
	countResp := msg.GetPingResponse() // Using PingResponse for count data
	if countResp != nil {
		log.Printf("Client Count - ID: %s, Count: %d, Server Time: %d μs",
			msg.ID, countResp.Sequence, countResp.ProcessingTimeUS)
	}
}

// handleError processes error messages from the server
func (c *Client) handleError(msg *v1.WSMessage) {
	errorResp := msg.GetErrorResponse()
	if errorResp != nil {
		log.Printf("Server Error - ID: %s, Code: %d, Reason: %s, Message: %s",
			msg.ID, errorResp.Code, errorResp.Reason, errorResp.Message)
	}
}

// pingPump sends ping messages to the server at regular intervals
func (c *Client) pingPump() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()

	sequence := uint64(1)

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.sendPing(sequence)
			sequence++
		}
	}
}

// countPump requests client count from server periodically
func (c *Client) countPump() {
	defer c.wg.Done()

	ticker := time.NewTicker(10 * time.Second) // Request count every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.sendCountRequest()
		}
	}
}

// sendPing sends a ping message to the server
func (c *Client) sendPing(sequence uint64) {
	pingReq := &v1.PingRequest{
		Sequence: sequence,
	}

	msg := &v1.WSMessage{
		Type: MessageTypePing,
		ID:   fmt.Sprintf("ping_%d_%d", sequence, time.Now().UnixNano()),
		Payload: &v1.WSMessage_PingRequest{
			PingRequest: pingReq,
		},
		Timestamp: timestamppb.Now(),
	}

	c.sendMessage(msg)
	log.Printf("Sent PING - Sequence: %d, ID: %s", sequence, msg.ID)
}

// sendCountRequest requests the current client count from server
func (c *Client) sendCountRequest() {
	msg := &v1.WSMessage{
		Type:      MessageTypeCount,
		ID:        fmt.Sprintf("count_%d", time.Now().UnixNano()),
		Timestamp: timestamppb.Now(),
	}

	c.sendMessage(msg)
	log.Printf("Sent COUNT request - ID: %s", msg.ID)
}

// sendMessage sends a message to the server
func (c *Client) sendMessage(msg *v1.WSMessage) {
	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	if err := c.conn.WriteJSON(msg); err != nil {
		log.Printf("Write failed: %v", err)
		return
	}
}

// sendInvalidMessage sends an invalid message type for testing error handling
func (c *Client) sendInvalidMessage() {
	msg := &v1.WSMessage{
		Type:      "invalid.message.type",
		ID:        fmt.Sprintf("invalid_%d", time.Now().UnixNano()),
		Timestamp: timestamppb.Now(),
	}

	c.sendMessage(msg)
	log.Printf("Sent INVALID message - ID: %s", msg.ID)
}

func main() {
	var (
		serverURL    = flag.String("url", "ws://localhost:8080/api/v1/ws", "WebSocket server URL")
		userID       = flag.String("user", "", "User ID (optional, will be generated if empty)")
		pingInterval = flag.Duration("ping", 5*time.Second, "Ping interval")
		numClients   = flag.Int("clients", 1, "Number of concurrent clients")
		testDuration = flag.Duration("duration", 30*time.Second, "Test duration")
		testInvalid  = flag.Bool("invalid", false, "Send invalid message after 5 seconds")
	)
	flag.Parse()

	log.Printf("Starting WebSocket client test...")
	log.Printf("Server URL: %s", *serverURL)
	log.Printf("Number of clients: %d", *numClients)
	log.Printf("Ping interval: %v", *pingInterval)
	log.Printf("Test duration: %v", *testDuration)

	// Create interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Create and start clients
	clients := make([]*Client, *numClients)
	for i := 0; i < *numClients; i++ {
		userIDForClient := *userID
		if userIDForClient == "" {
			userIDForClient = fmt.Sprintf("test_user_%d", i+1)
		} else if *numClients > 1 {
			userIDForClient = fmt.Sprintf("%s_%d", *userID, i+1)
		}

		client := NewClient(*serverURL, userIDForClient, *pingInterval)
		clients[i] = client

		if err := client.Connect(); err != nil {
			log.Fatalf("Failed to connect client %d: %v", i+1, err)
		}

		go client.Start()
	}

	// Send invalid message for testing if requested
	if *testInvalid && len(clients) > 0 {
		go func() {
			time.Sleep(5 * time.Second)
			log.Printf("Sending invalid message for testing...")
			clients[0].sendInvalidMessage()
		}()
	}

	// Wait for test duration or interrupt
	select {
	case <-time.After(*testDuration):
		log.Printf("Test duration completed")
	case <-interrupt:
		log.Printf("Interrupt received")
	}

	// Clean shutdown
	log.Printf("Shutting down clients...")
	for i, client := range clients {
		log.Printf("Stopping client %d...", i+1)
		client.Stop()
	}

	log.Printf("Test completed")
}
