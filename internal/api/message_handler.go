package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ymow/messenger_protocol_research/internal/message"
	"github.com/ymow/messenger_protocol_research/internal/model"
)

// MessageHandler handles chat message operations
type MessageHandler struct {
	db                *gorm.DB
	service           *message.Service
	permissionService interface {
		CanSendMessage(ctx context.Context, senderID, receiverID uuid.UUID) (bool, string, error)
	}
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(db *gorm.DB, service *message.Service, permService interface {
	CanSendMessage(ctx context.Context, senderID, receiverID uuid.UUID) (bool, string, error)
}) *MessageHandler {
	return &MessageHandler{
		db:                db,
		service:           service,
		permissionService: permService,
	}
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

	// PERMISSION CHECK: Verify sender can message receiver
	if h.permissionService != nil {
		canSend, reason, err := h.permissionService.CanSendMessage(
			c.Request.Context(),
			req.SenderID,
			req.ReceiverID,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check messaging permission",
				"code":  "PERMISSION_CHECK_FAILED",
			})
			return
		}

		if !canSend {
			// Map reason to HTTP status and user message
			statusCode := http.StatusForbidden
			errorMsg := "You cannot send messages to this user"

			switch reason {
			case "NO_VALID_DISCOVERY":
				errorMsg = "No valid BLE discovery. You must be within range (50-100m) and discovery must be less than 10 minutes old."
			case "SENDER_BLOCKED_BY_RECEIVER":
				errorMsg = "This user has blocked you"
			case "SELF_MESSAGE_NOT_ALLOWED":
				statusCode = http.StatusBadRequest
				errorMsg = "You cannot message yourself"
			}

			c.JSON(statusCode, gin.H{
				"error": errorMsg,
				"code":  reason,
			})
			return
		}
	}
	// END PERMISSION CHECK

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

// GetMessages handles GET /api/v1/messages - retrieve message history with pagination
func (h *MessageHandler) GetMessages(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Parse query params with validation
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Build filters
	filters := &message.MessageFilters{
		PeerID:    c.Query("peer_id"),
		Since:     c.Query("since"),
		Status:    c.Query("status"),
		Direction: c.Query("direction"),
	}

	// Call service
	messages, total, err := h.service.GetUserMessages(
		c.Request.Context(),
		userID,
		limit,
		offset,
		filters,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "FETCH_FAILED",
		})
		return
	}

	// Return paginated response
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"messages": messages,
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"has_more": (offset + limit) < int(total),
		},
	})
}

// GetConversation handles GET /api/v1/messages/conversation/:peer_id
func (h *MessageHandler) GetConversation(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Parse peer ID
	peerIDStr := c.Param("peer_id")
	peerID, err := uuid.Parse(peerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid peer ID",
			"code":  "INVALID_PEER_ID",
		})
		return
	}

	// Parse pagination
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Parse time filters
	var beforeTime, afterTime *time.Time
	if before := c.Query("before"); before != "" {
		if t, err := time.Parse(time.RFC3339, before); err == nil {
			beforeTime = &t
		}
	}
	if after := c.Query("after"); after != "" {
		if t, err := time.Parse(time.RFC3339, after); err == nil {
			afterTime = &t
		}
	}

	// Call service
	messages, total, err := h.service.GetConversation(
		c.Request.Context(),
		userID,
		peerID,
		limit,
		offset,
		beforeTime,
		afterTime,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"messages": messages,
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"peer_id":  peerID,
		},
	})
}

// MarkAsRead handles PATCH /api/v1/messages/:id/read
func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid message ID",
			"code":  "INVALID_MESSAGE_ID",
		})
		return
	}

	message, err := h.service.MarkMessageAsRead(c.Request.Context(), messageID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "MARK_READ_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   message,
	})
}