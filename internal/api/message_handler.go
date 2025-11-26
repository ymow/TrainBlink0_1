package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ymow/messenger_protocol_research/internal/model"
)

// MessageHandler handles chat message operations
type MessageHandler struct {
	db *gorm.DB
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(db *gorm.DB) *MessageHandler {
	return &MessageHandler{db: db}
}

// MessageRequest represents the request format for sending a message
type MessageRequest struct {
	Text       string    `json:"text" binding:"required"`
	SenderID   uuid.UUID `json:"sender_id" binding:"required"`
	ReceiverID uuid.UUID `json:"receiver_id" binding:"required"`
	Timestamp  time.Time `json:"timestamp"`
}

// MessageResponse represents the response format for message operations
type MessageResponse struct {
	ID             uuid.UUID                     `json:"id"`
	Text           string                        `json:"text"`
	SenderID       uuid.UUID                     `json:"sender_id"`
	ReceiverID     uuid.UUID                     `json:"receiver_id"`
	Timestamp      time.Time                     `json:"timestamp"`
	DeliveryStatus model.MessageDeliveryStatus   `json:"delivery_status"`
	CreatedAt      time.Time                     `json:"created_at"`
}

// PostMessage handles POST /api/v1/message - persist a chat message
func (h *MessageHandler) PostMessage(c *gin.Context) {
	var req MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set timestamp if not provided
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	// Create new message
	message := model.NewTextMessage(req.Text, req.SenderID, req.ReceiverID)
	message.Timestamp = req.Timestamp

	// For now, handle without database if db is nil
	if h.db == nil {
		// Return success response even without database persistence
		response := MessageResponse{
			ID:             message.ID,
			Text:           message.Text,
			SenderID:       message.SenderID,
			ReceiverID:     message.ReceiverID,
			Timestamp:      message.Timestamp,
			DeliveryStatus: message.DeliveryStatus,
			CreatedAt:      message.CreatedAt,
		}
		c.JSON(http.StatusCreated, response)
		return
	}

	// Save to database
	if err := h.db.Create(message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	response := MessageResponse{
		ID:             message.ID,
		Text:           message.Text,
		SenderID:       message.SenderID,
		ReceiverID:     message.ReceiverID,
		Timestamp:      message.Timestamp,
		DeliveryStatus: message.DeliveryStatus,
		CreatedAt:      message.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}