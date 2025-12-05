package model

import "time"

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeMessage      MessageType = "message"       // Regular chat message
	MessageTypeBroadcast    MessageType = "broadcast"     // Station-wide broadcast
	MessageTypeTyping       MessageType = "typing"        // Typing indicator
	MessageTypePresence     MessageType = "presence"      // User presence update
	MessageTypeJoin         MessageType = "join"          // User joined station
	MessageTypeLeave        MessageType = "leave"         // User left station
	MessageTypePing         MessageType = "ping"          // Heartbeat ping
	MessageTypePong         MessageType = "pong"          // Heartbeat pong
	MessageTypeError        MessageType = "error"         // Error message
	MessageTypeAck          MessageType = "ack"           // Acknowledgment
	MessageTypeReadReceipt  MessageType = "read_receipt"  // Client → Server: mark as read (Phase 1)
	MessageTypeReadAck      MessageType = "read_ack"      // Server → Client: read acknowledgment (Phase 1)
	MessageTypeDeliveryAck  MessageType = "delivery_ack"  // Client → Server: delivery acknowledgment (Phase 1)
)

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	ID        string                 `json:"id"`                   // Unique message ID
	Type      MessageType            `json:"type"`                 // Message type
	From      string                 `json:"from,omitempty"`       // Sender user ID
	To        string                 `json:"to,omitempty"`         // Recipient user ID (optional, for DM)
	StationID string                 `json:"station_id,omitempty"` // Station ID (for room messages)
	Content   string                 `json:"content,omitempty"`    // Message content
	Timestamp time.Time              `json:"timestamp"`            // Message timestamp
	Metadata  map[string]interface{} `json:"metadata,omitempty"`   // Additional metadata
}

// ClientMessage represents a message sent from client to server
type ClientMessage struct {
	Type     MessageType            `json:"type"`               // Message type
	To       string                 `json:"to,omitempty"`       // Recipient (optional)
	Content  string                 `json:"content,omitempty"`  // Message content
	Metadata map[string]interface{} `json:"metadata,omitempty"` // Additional metadata
}

// ServerMessage represents a message sent from server to client
type ServerMessage struct {
	ID        string                 `json:"id"`                   // Message ID
	Type      MessageType            `json:"type"`                 // Message type
	From      string                 `json:"from,omitempty"`       // Sender
	To        string                 `json:"to,omitempty"`         // Recipient
	StationID string                 `json:"station_id,omitempty"` // Station
	Content   string                 `json:"content,omitempty"`    // Content
	Timestamp time.Time              `json:"timestamp"`            // Timestamp
	Metadata  map[string]interface{} `json:"metadata,omitempty"`   // Metadata
}

// PresenceStatus represents user presence status
type PresenceStatus string

const (
	PresenceOnline  PresenceStatus = "online"
	PresenceTyping  PresenceStatus = "typing"
	PresenceOffline PresenceStatus = "offline"
)

// PresenceUpdate represents a user presence update
type PresenceUpdate struct {
	UserID    string         `json:"user_id"`
	StationID string         `json:"station_id"`
	Status    PresenceStatus `json:"status"`
	Timestamp time.Time      `json:"timestamp"`
}

// TypingIndicator represents a typing indicator event
type TypingIndicator struct {
	UserID    string    `json:"user_id"`
	StationID string    `json:"station_id"`
	IsTyping  bool      `json:"is_typing"`
	Timestamp time.Time `json:"timestamp"`
}

// JoinEvent represents a user joining a station
type JoinEvent struct {
	UserID    string    `json:"user_id"`
	DeviceID  string    `json:"device_id,omitempty"`
	StationID string    `json:"station_id"`
	Timestamp time.Time `json:"timestamp"`
}

// LeaveEvent represents a user leaving a station
type LeaveEvent struct {
	UserID    string    `json:"user_id"`
	StationID string    `json:"station_id"`
	Duration  int       `json:"duration_seconds,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorMessage represents an error message
type ErrorMessage struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
