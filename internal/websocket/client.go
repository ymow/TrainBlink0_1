package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ymow/messenger_protocol_research/internal/model"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10 // 54 seconds

	// Maximum message size allowed from peer
	maxMessageSize = 8192 // 8KB

	// Rate limiting
	maxMessagesPerMinute   = 60  // Regular messages
	maxBroadcastsPerMinute = 10  // Broadcast messages
	rateLimitWindow        = time.Minute
)

// Client is a middleman between the websocket connection and the hub
type Client struct {
	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan *model.ServerMessage

	// Hub reference
	hub *Hub

	// User identification
	UserID    string
	DeviceID  string
	SessionID string
	StationID string

	// Client metadata
	ConnectedAt time.Time
	LastPingAt  time.Time

	// Rate limiting
	messageCount     int       // Messages sent in current window
	broadcastCount   int       // Broadcasts sent in current window
	rateLimitReset   time.Time // When to reset counters
	totalMessages    int64     // Total messages sent (lifetime)
	totalBroadcasts  int64     // Total broadcasts sent (lifetime)

	// Message handler
	messageHandler *MessageHandler
}

// NewClient creates a new Client
func NewClient(conn *websocket.Conn, hub *Hub, userID, deviceID, sessionID, stationID string, messageHandler *MessageHandler) *Client {
	now := time.Now()
	return &Client{
		conn:           conn,
		send:           make(chan *model.ServerMessage, 256),
		hub:            hub,
		UserID:         userID,
		DeviceID:       deviceID,
		SessionID:      sessionID,
		StationID:      stationID,
		ConnectedAt:    now,
		LastPingAt:     now,
		rateLimitReset: now.Add(rateLimitWindow),
		messageHandler: messageHandler,
	}
}

// readPump pumps messages from the websocket connection to the hub
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.LastPingAt = time.Now()
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		// Parse the client message
		var clientMsg model.ClientMessage
		if err := json.Unmarshal(messageBytes, &clientMsg); err != nil {
			fmt.Printf("Failed to parse message: %v\n", err)
			// Send error back to client
			c.sendError("INVALID_MESSAGE", "Failed to parse message")
			continue
		}

		// Handle the message
		c.messageHandler.HandleMessage(c, &clientMsg)
	}
}

// writePump pumps messages from the hub to the websocket connection
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write message as JSON
			if err := c.conn.WriteJSON(message); err != nil {
				fmt.Printf("Failed to write message: %v\n", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(code, message string) {
	errorMsg := &model.ServerMessage{
		ID:        generateMessageID(),
		Type:      model.MessageTypeError,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"error": &model.ErrorMessage{
				Code:      code,
				Message:   message,
				Timestamp: time.Now(),
			},
		},
	}

	select {
	case c.send <- errorMsg:
	default:
		fmt.Printf("Failed to send error to client: send channel full\n")
	}
}

// sendAck sends an acknowledgment message to the client
func (c *Client) sendAck(originalMessageID string) {
	ackMsg := &model.ServerMessage{
		ID:        generateMessageID(),
		Type:      model.MessageTypeAck,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"ack_for": originalMessageID,
		},
	}

	select {
	case c.send <- ackMsg:
	default:
		fmt.Printf("Failed to send ack to client: send channel full\n")
	}
}

// checkRateLimit checks if the client has exceeded rate limits
// Returns true if allowed, false if rate limit exceeded
func (c *Client) checkRateLimit(isBroadcast bool) bool {
	now := time.Now()

	// Reset counters if window has expired
	if now.After(c.rateLimitReset) {
		c.messageCount = 0
		c.broadcastCount = 0
		c.rateLimitReset = now.Add(rateLimitWindow)
	}

	// Check broadcast rate limit
	if isBroadcast {
		if c.broadcastCount >= maxBroadcastsPerMinute {
			return false
		}
		c.broadcastCount++
		c.totalBroadcasts++
		return true
	}

	// Check regular message rate limit
	if c.messageCount >= maxMessagesPerMinute {
		return false
	}
	c.messageCount++
	c.totalMessages++
	return true
}

// GetStats returns client statistics
func (c *Client) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"user_id":          c.UserID,
		"station_id":       c.StationID,
		"connected_at":     c.ConnectedAt,
		"uptime_seconds":   time.Since(c.ConnectedAt).Seconds(),
		"total_messages":   c.totalMessages,
		"total_broadcasts": c.totalBroadcasts,
		"last_ping_at":     c.LastPingAt,
	}
}
