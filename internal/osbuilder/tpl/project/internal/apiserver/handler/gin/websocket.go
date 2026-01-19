package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/onexstack/onexstack/pkg/errorsx"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

    "{{.D.ModuleName}}/internal/pkg/contextx"
    "{{.D.ModuleName}}/internal/pkg/errno"
    {{.Web.APIImportPath}}
)

// WebSocket configuration constants.
const (
	writeWait       = 10 * time.Second
	pongWait        = 60 * time.Second
	pingPeriod      = 54 * time.Second
	maxMessageSize  = 512
	bufferSize      = 256
	readBufferSize  = 1024
	writeBufferSize = 1024
)

// WebSocket message types.
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
		// TODO: Add proper origin check for production environment.
		return true
	},
}

// WSHub manages the set of active clients and broadcasts messages to the clients.
type WSHub struct {
	// clients uses a map with a RWMutex instead of sync.Map for better type safety
	// and performance in this specific usage pattern (frequent reads/broadcasts).
	clientsMu sync.RWMutex
	clients   map[*WSClient]struct{}

	clientCount int64
}

// WSClient is a middleman between the websocket connection and the hub.
type WSClient struct {
	hub    *WSHub
	conn   *websocket.Conn
	send   chan []byte
	userID string
	// logger pre-binds context fields like user_id to simplify logging.
	logger *slog.Logger
}

// NewWSHub creates a new WebSocket hub instance.
func NewWSHub() *WSHub {
	return &WSHub{
		clients: make(map[*WSClient]struct{}),
	}
}

// Register registers a new client to the hub.
func (h *WSHub) Register(client *WSClient) {
	h.clientsMu.Lock()
	if _, ok := h.clients[client]; !ok {
		h.clients[client] = struct{}{}
		atomic.AddInt64(&h.clientCount, 1)
	}
	h.clientsMu.Unlock()

	client.logger.Info("client connected", "total_clients", atomic.LoadInt64(&h.clientCount))
}

// Unregister removes a client from the hub and closes the connection.
func (h *WSHub) Unregister(client *WSClient) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
		atomic.AddInt64(&h.clientCount, -1)
		client.logger.Info("client disconnected", "total_clients", atomic.LoadInt64(&h.clientCount))
	}
}

// Count returns the number of connected clients.
func (h *WSHub) Count() int64 {
	return atomic.LoadInt64(&h.clientCount)
}

// BroadcastMessage sends a message to all connected clients.
func (h *WSHub) BroadcastMessage(msg *v1.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal broadcast message", "error", err)
		return
	}

	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			// If the client's send buffer is full, we assume the connection is dead or stalled.
			// We handle the unregister in a non-blocking way or let the read/write pumps handle the closure.
			// Note: Avoid calling Unregister here directly if it locks Mutex to prevent deadlock,
			// but since Unregister locks the same mutex we are holding (RLock), we must NOT call it here.
			// Instead, just log warning. The pumps will eventually fail and clean up.
			client.logger.Warn("client buffer full during broadcast, skipping message")
		}
	}
}

// NewWSClient creates a new WebSocket client instance.
func NewWSClient(conn *websocket.Conn, hub *WSHub, userID string) *WSClient {
	if userID == "" {
		userID = generateUserID()
	}

	client := &WSClient{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, bufferSize),
		userID: userID,
		logger: slog.With("user_id", userID),
	}

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	return client
}

// ServeWebSocket handles websocket requests from the peer.
func (h *Handler) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}

	// We don't defer conn.Close() here because the connection ownership is transferred
	// to the client pumps which manage the lifecycle.

	userID := contextx.UserID(r.Context())
	client := NewWSClient(conn, h.hub, userID)

	h.hub.Register(client)

	// Allow collection of memory referencing the caller by starting goroutines.
	go h.writePump(client)
	go h.readPump(client)
}

// readPump pumps messages from the websocket connection to the hub.
// The application runs readPump in a per-connection goroutine.
func (h *Handler) readPump(client *WSClient) {
	defer func() {
		h.hub.Unregister(client)
		client.conn.Close()
	}()

	for {
		_, messageBytes, err := client.conn.ReadMessage() // Read raw bytes
		if err != nil {
			slog.Error("Failed to read message from connection", "error", err)
			// h.sendError(client, msg.ID, errno.ErrPayloadInvalid)
			break
		}

		var msg v1.WSMessage
		if err := protojson.Unmarshal(messageBytes, &msg); err != nil {
			slog.Error("Failed to unmarshal protobuf message from JSON", "error", err, "rawMessage", string(messageBytes))
			// h.sendError(client, msg.ID, errno.ErrPayloadInvalid)
			break
		}

		msg.Timestamp = timestamppb.Now()

		switch msg.Type {
		case MessageTypePing:
			h.handlePing(client, &msg)
		case MessageTypeCount:
			h.handleCount(client, &msg)
		default:
			client.logger.Warn("unknown message type received", "type", msg.Type)
			h.sendError(client, msg.ID, errno.ErrInvalidMessageType)
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
// A goroutine running writePump is started for each connection.
func (h *Handler) writePump(client *WSClient) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := h.writeWithBatch(client, message); err != nil {
				client.logger.Debug("failed to write message", "error", err)
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// writeWithBatch writes the initial message and any queued messages to the websocket connection.
func (h *Handler) writeWithBatch(client *WSClient, message []byte) error {
	w, err := client.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	if _, err := w.Write(message); err != nil {
		return err
	}

	// Add queued chat messages to the current websocket message.
	n := len(client.send)
	for i := 0; i < n; i++ {
		select {
		case msg := <-client.send:
			w.Write([]byte{'\n'})
			w.Write(msg)
		default:
			// Channel is empty.
		}
	}

	return w.Close()
}

// handlePing processes the ping request business logic.
func (h *Handler) handlePing(client *WSClient, msg *v1.WSMessage) {
	rq := msg.GetPingRequest()
	client.logger.Debug("handling ping request", "sequence", rq.Sequence)

	// Note: using context.Background() as the request context is closed after upgrade.
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
	client.logger.Debug("sent pong response", "sequence", response.Sequence, "processing_time_us", response.ProcessingTimeUS)
}

// handleCount processes the count request logic.
func (h *Handler) handleCount(client *WSClient, msg *v1.WSMessage) {
	count := h.hub.Count()
	client.logger.Debug("client count requested", "count", count)

	// TODO: Define a dedicated CountResponse message type in protobuf.
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

// sendMessage marshals and queues a message for delivery to the client.
func (h *Handler) sendMessage(client *WSClient, msg *v1.WSMessage) {
	messageBytes, err := protojson.Marshal(msg)
	if err != nil {
		client.logger.Error("Failed to marshal protobuf message to JSON", "error", err)
		return
	}

	select {
	case client.send <- messageBytes:
	default:
		h.hub.Unregister(client)
		client.logger.Warn("client buffer full, unregistering")
	}
}

// sendError helper sends a standardized error response.
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
	client.logger.Debug("sent error response", "code", errx.Code, "reason", errx.Reason)
}

// generateUserID creates a random anonymous user ID.
func generateUserID() string {
	return "anonymous_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

func init() {
	Register(func(v1 *gin.RouterGroup, handler *Handler, mws ...gin.HandlerFunc) {
		v1.GET("/ws", gin.WrapH(http.HandlerFunc(handler.ServeWebSocket)))
	})
}
