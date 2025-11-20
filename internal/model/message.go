package model

import (
	"time"

	"github.com/google/uuid"
)

// MessageDeliveryStatus represents the delivery status of a message
type MessageDeliveryStatus string

const (
	MessagePending   MessageDeliveryStatus = "PENDING"
	MessageSending   MessageDeliveryStatus = "SENDING"
	MessageSent      MessageDeliveryStatus = "SENT"
	MessageDelivered MessageDeliveryStatus = "DELIVERED"
	MessageRead      MessageDeliveryStatus = "READ"
	MessageFailed    MessageDeliveryStatus = "FAILED"
)

// ChatMessage represents a chat message between peers
type ChatMessage struct {
	ID             uuid.UUID             `json:"id" gorm:"type:uuid;primaryKey"`
	Text           string                `json:"text" gorm:"type:text;not null"`
	SenderID       uuid.UUID             `json:"sender_id" gorm:"type:uuid;not null"`
	ReceiverID     uuid.UUID             `json:"receiver_id" gorm:"type:uuid;not null"`
	Timestamp      time.Time             `json:"timestamp" gorm:"not null"`
	IsEncrypted    bool                  `json:"is_encrypted" gorm:"default:false"`
	IsRead         bool                  `json:"is_read" gorm:"default:false"`
	DeliveryStatus MessageDeliveryStatus `json:"delivery_status" gorm:"type:varchar(20);default:'PENDING'"`
	DeliveredAt    *time.Time            `json:"delivered_at,omitempty"`
	ReadAt         *time.Time            `json:"read_at,omitempty"`

	// Ephemeral message support (Feature 6)
	IsEphemeral bool       `json:"is_ephemeral" gorm:"default:false"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsExpired   bool       `json:"is_expired" gorm:"default:false"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName overrides the table name
func (ChatMessage) TableName() string {
	return "chat_messages"
}

// IsSentByMe returns true if message was sent by given user ID
func (m *ChatMessage) IsSentByMe(userID uuid.UUID) bool {
	return m.SenderID == userID
}

// IsReceivedByMe returns true if message was received by given user ID
func (m *ChatMessage) IsReceivedByMe(userID uuid.UUID) bool {
	return m.ReceiverID == userID
}

// TimeRemainingSeconds returns seconds until message expires
func (m *ChatMessage) TimeRemainingSeconds() *int64 {
	if !m.IsEphemeral || m.ExpiresAt == nil {
		return nil
	}
	remaining := int64(time.Until(*m.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return &remaining
}

// ShouldDelete returns true if ephemeral message should be deleted
func (m *ChatMessage) ShouldDelete() bool {
	if !m.IsEphemeral {
		return false
	}
	if m.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*m.ExpiresAt) || m.IsExpired
}

// MarkAsDelivered marks message as delivered
func (m *ChatMessage) MarkAsDelivered() {
	now := time.Now()
	m.DeliveryStatus = MessageDelivered
	m.DeliveredAt = &now
	m.UpdatedAt = now
}

// MarkAsRead marks message as read
func (m *ChatMessage) MarkAsRead() {
	now := time.Now()
	m.IsRead = true
	m.ReadAt = &now
	if m.DeliveryStatus == MessageDelivered {
		m.DeliveryStatus = MessageRead
	}
	m.UpdatedAt = now
}

// MarkAsFailed marks message as failed
func (m *ChatMessage) MarkAsFailed() {
	m.DeliveryStatus = MessageFailed
	m.UpdatedAt = time.Now()
}

// MarkAsExpired marks ephemeral message as expired
func (m *ChatMessage) MarkAsExpired() {
	m.IsExpired = true
	m.UpdatedAt = time.Now()
}

// NewTextMessage creates a new text message
func NewTextMessage(text string, senderID, receiverID uuid.UUID) *ChatMessage {
	now := time.Now()
	return &ChatMessage{
		ID:             uuid.New(),
		Text:           text,
		SenderID:       senderID,
		ReceiverID:     receiverID,
		Timestamp:      now,
		DeliveryStatus: MessagePending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// NewEphemeralMessage creates a new ephemeral message (auto-deletes after lifetimeSeconds)
func NewEphemeralMessage(text string, senderID, receiverID uuid.UUID, lifetimeSeconds int64) *ChatMessage {
	now := time.Now()
	expiresAt := now.Add(time.Duration(lifetimeSeconds) * time.Second)

	return &ChatMessage{
		ID:             uuid.New(),
		Text:           text,
		SenderID:       senderID,
		ReceiverID:     receiverID,
		Timestamp:      now,
		DeliveryStatus: MessagePending,
		IsEphemeral:    true,
		ExpiresAt:      &expiresAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
