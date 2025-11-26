package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketHandler manages WebSocket connections for real-time chat
type WebSocketHandler struct {
	connectionsHandler *ConnectionsHandler
	upgrader           websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(connectionsHandler *ConnectionsHandler) *WebSocketHandler {
	return &WebSocketHandler{
		connectionsHandler: connectionsHandler,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from localhost and development origins
				origin := r.Header.Get("Origin")
				return origin == "http://localhost:5173" || origin == "http://localhost:3000"
			},
		},
	}
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type       string      `json:"type"`
	Content    interface{} `json:"content"`
	SenderID   string      `json:"sender_id,omitempty"`
	ReceiverID string      `json:"receiver_id,omitempty"`
	Timestamp  string      `json:"timestamp,omitempty"`
}

// HandleWebSocket handles WebSocket connections
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Generate client ID
	clientID := uuid.New().String()
	log.Printf("WebSocket client connected: %s", clientID)

	// Add connection to tracking
	h.connectionsHandler.AddConnection(clientID)
	defer h.connectionsHandler.RemoveConnection(clientID)

	// Send welcome message
	welcomeMsg := WebSocketMessage{
		Type: "welcome",
		Content: map[string]interface{}{
			"client_id": clientID,
			"message":   "Connected to TrainBlink chat",
		},
	}
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Error sending welcome message: %v", err)
		return
	}

	// Handle incoming messages
	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle different message types
		switch msg.Type {
		case "ping":
			// Respond with pong
			pongMsg := WebSocketMessage{
				Type:    "pong",
				Content: "pong",
			}
			if err := conn.WriteJSON(pongMsg); err != nil {
				log.Printf("Error sending pong: %v", err)
				return
			}

		case "chat_message":
			// Echo the message back (real-time chat simulation)
			// In a full implementation, this would broadcast to the receiver
			echoMsg := WebSocketMessage{
				Type:       "message_received",
				Content:    msg.Content,
				SenderID:   msg.SenderID,
				ReceiverID: msg.ReceiverID,
				Timestamp:  msg.Timestamp,
			}
			if err := conn.WriteJSON(echoMsg); err != nil {
				log.Printf("Error echoing message: %v", err)
				return
			}

		case "get_connections":
			// Return current connection count
			connectionsMsg := WebSocketMessage{
				Type: "connections_update",
				Content: map[string]interface{}{
					"count":      len(h.connectionsHandler.activeConnections),
					"client_ids": h.connectionsHandler.clientIDs,
				},
			}
			if err := conn.WriteJSON(connectionsMsg); err != nil {
				log.Printf("Error sending connections update: %v", err)
				return
			}

		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}

	log.Printf("WebSocket client disconnected: %s", clientID)
}
