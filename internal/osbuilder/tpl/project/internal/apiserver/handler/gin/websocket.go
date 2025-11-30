package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/errorsx"
	"google.golang.org/protobuf/types/known/timestamppb"

	"{{.D.ModuleName}}/internal/pkg/contextx"
	"{{.D.ModuleName}}/internal/pkg/errno"
	{{.Web.APIImportPath}}
)

// Configuration constants
const (
	writeWait       = 10 * time.Second
	pongWait        = 60 * time.Second
	pingPeriod      = 54 * time.Second
	maxMessageSize  = 512
	bufferSize      = 256
	readBufferSize  = 1024
	writeBufferSize = 1024
)

// Message types
const (
	MessageTypePing          = "ws.ping"
	MessageTypeCount         = "ws.count"
	MessageTypePong          = "ws.pong"
	MessageTypeError         = "ws.error"
	MessageTypeCountResponse = "ws.count_response"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Add proper origin check for production
	},
}

// WSHub manages all WebSocket client connections
type WSHub struct {
	clients     sync.Map // map[*WSClient]bool - using sync.Map for better concurrency
	clientCount int32    // atomic counter for client count
	mu          sync.RWMutex
}

// WSClient represents a WebSocket client connection
type WSClient struct {
	conn   *websocket.Conn
	send   chan []byte
	hub    *WSHub
	userID string
}

// NewWSHub creates a new WebSocket hub
func NewWSHub() *WSHub {
	return &WSHub{}
}

// Register adds a client to the hub
func (h *WSHub) Register(client *WSClient) {
	h.clients.Store(client, true)

	h.mu.Lock()
	h.clientCount++
	count := h.clientCount
	h.mu.Unlock()

	slog.Info("client connected", "total_clients", count, "user_id", client.userID)
}

// Unregister removes a client from the hub
func (h *WSHub) Unregister(client *WSClient) {
	if _, exists := h.clients.LoadAndDelete(client); exists {
		defer func() {
			if recover() != nil {
				// Channel already closed
			}
		}()
		close(client.send)

		h.mu.Lock()
		h.clientCount--
		h.mu.Unlock()

		slog.Info("client disconnected", "total_clients", h.clientCount, "user_id", client.userID)
	}
}

// Count returns the number of connected clients
func (h *WSHub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return int(h.clientCount)
}

// BroadcastMessage broadcasts a message to all connected clients
func (h *WSHub) BroadcastMessage(msg *v1.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("broadcast marshal error", "error", err)
		return
	}

	h.clients.Range(func(key, value interface{}) bool {
		client := key.(*WSClient)
		select {
		case client.send <- data:
		default:
			// Client buffer full, remove client
			h.Unregister(client)
			slog.Warn("client buffer full during broadcast, removed", "user_id", client.userID)
		}
		return true
	})
}

// NewWSClient creates a new WebSocket client
func NewWSClient(conn *websocket.Conn, hub *WSHub) *WSClient {
	client := &WSClient{
		conn:   conn,
		send:   make(chan []byte, bufferSize),
		hub:    hub,
		userID: generateUserID(),
	}

	// Configure connection settings
	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	return client
}

// ServeWebSocket handles WebSocket upgrade and connection management
func (h *Handler) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	// Get user ID from query parameter or generate one
	userID := contextx.UserID(r.Context())
	client := NewWSClient(conn, h.hub)
	if userID != "" {
		client.userID = userID
	}

	h.hub.Register(client)
	defer h.hub.Unregister(client)

	// Start write pump in goroutine
	go h.writePump(client)

	// Read pump blocks until connection closes
	h.readPump(client)
}

// readPump handles incoming messages from WebSocket client and routes to appropriate handlers.
func (h *Handler) readPump(client *WSClient) {
	defer client.conn.Close()
	for {
		var msg v1.WSMessage
		if err := client.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket read error", "error", err, "user_id", client.userID)
			}
			break
		}

		msg.Timestamp = timestamppb.Now()

		switch msg.Type {
		case MessageTypePing:
			h.handlePing(client, &msg)
		case MessageTypeCount:
			h.handleCount(client, &msg)
		default:
			h.sendError(client, msg.ID, errno.ErrInvalidMessageType)
			slog.Warn("unknown message type", "type", msg.Type, "user_id", client.userID)
		}
	}
}

// writePump handles outgoing messages to WebSocket client
func (h *Handler) writePump(client *WSClient) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			if !ok {
				h.writeCloseMessage(client)
				return
			}

			if err := h.writeMessageWithBatch(client, message); err != nil {
				slog.Debug("write message failed",
					"error", err,
					"user_id", client.userID)
				return
			}

		case <-ticker.C:
			if err := h.writePingMessage(client); err != nil {
				return
			}
		}
	}
}

// writeMessageWithBatch writes message with batching support
func (h *Handler) writeMessageWithBatch(client *WSClient, message []byte) error {
	client.conn.SetWriteDeadline(time.Now().Add(writeWait))

	w, err := client.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	if _, err := w.Write(message); err != nil {
		return err
	}

	// Write batched messages from queue
	return h.writeBatchedMessages(w, client)
}

// writeBatchedMessages writes any queued messages in batch
func (h *Handler) writeBatchedMessages(w io.WriteCloser, client *WSClient) error {
	n := len(client.send)
	for i := 0; i < n; i++ {
		select {
		case msg := <-client.send:
			if _, err := w.Write([]byte{'\n'}); err != nil {
				return err
			}
			if _, err := w.Write(msg); err != nil {
				return err
			}
		default:
			return nil
		}
	}
	return nil
}

// writeCloseMessage sends close message to client
func (h *Handler) writeCloseMessage(client *WSClient) {
	client.conn.SetWriteDeadline(time.Now().Add(writeWait))
	client.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

// writePingMessage sends ping message to client
func (h *Handler) writePingMessage(client *WSClient) error {
	client.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return client.conn.WriteMessage(websocket.PingMessage, nil)
}

func (h *Handler) handleMessage(client *WSClient, msg *v1.WSMessage) {
	msg.Timestamp = timestamppb.Now()

	switch msg.Type {
	case MessageTypePing:
		h.handlePing(client, msg)
	case MessageTypeCount:
		h.handleCount(client, msg)
	default:
		h.sendError(client, msg.ID, errno.ErrInvalidMessageType)
		slog.Warn("unknown message type", "type", msg.Type, "user_id", client.userID)
	}
}

// handlePing processes ping requests
func (h *Handler) handlePing(client *WSClient, msg *v1.WSMessage) {
	rq := msg.GetPingRequest()
	if rq == nil {
		h.sendError(client, msg.ID, errno.ErrPayloadInvalid)
		return
	}

	slog.Debug("handling ping request", "sequence", rq.Sequence, "user_id", client.userID)

	response, err := h.biz.WSV1().Ping(context.Background(), rq)
	if err != nil {
		h.sendError(client, msg.ID, errno.ErrPing.WithMessage(err.Error()))
		return
	}

	message := &v1.WSMessage{
		Type: MessageTypePong,
		ID:   msg.ID,
		Payload: &v1.WSMessage_PingResponse{
			PingResponse: response,
		},
		Timestamp: timestamppb.Now(),
	}

	h.sendMessage(client, message)

	slog.Debug("Sent pong response", "sequence", response.Sequence, "processing_time_us", response.ProcessingTimeUS)
}

// handleCount processes client count requests
func (h *Handler) handleCount(client *WSClient, msg *v1.WSMessage) {
	count := h.hub.Count()

	slog.Debug("client count requested", "count", count, "user_id", client.userID)

	// Use PingResponse structure for count data
	// TODO: Define dedicated CountResponse message type in proto
	response := &v1.PingResponse{
		Sequence:         uint64(count),
		ProcessingTimeUS: time.Now().UnixMicro(),
	}

	message := &v1.WSMessage{
		Type: MessageTypeCountResponse,
		ID:   msg.ID,
		Payload: &v1.WSMessage_PingResponse{
			PingResponse: response,
		},
		Timestamp: timestamppb.Now(),
	}

	h.sendMessage(client, message)
}

// sendMessage sends a message to specific client
func (h *Handler) sendMessage(client *WSClient, msg *v1.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("message marshal failed", "error", err, "user_id", client.userID)
		return
	}

	select {
	case client.send <- data:
		// Message sent successfully
	default:
		// Buffer full, unregister client
		h.hub.Unregister(client)
		slog.Warn("send buffer full, client removed", "user_id", client.userID)
	}
}

// sendError sends error response to client
func (h *Handler) sendError(client *WSClient, msgID string, errx *errorsx.ErrorX) {
	errMessage := &v1.WSMessage{
		Type: MessageTypeError,
		ID:   msgID,
		Payload: &v1.WSMessage_ErrorResponse{
			ErrorResponse: &v1.Error{
				Code:    int32(errx.Code),
				Reason:  errx.Reason,
				Message: errx.Message,
			},
		},
		Timestamp: timestamppb.Now(),
	}

	h.sendMessage(client, errMessage)

	slog.Debug("sent error response", "user_id", client.userID, "code", errx.Code, "reason", errx.Reason)
}

// generateUserID creates a unique user identifier
func generateUserID() string {
	return "anonymous_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

func init() {
	Register(func(v1 *gin.RouterGroup, handler *Handler) {
		v1.GET("/ws", gin.WrapH(http.HandlerFunc(handler.ServeWebSocket)))
	})
}

