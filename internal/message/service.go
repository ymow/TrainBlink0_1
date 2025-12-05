package message

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/ymow/messenger_protocol_research/internal/model"
)

// Service handles message business logic
type Service struct {
	db    *gorm.DB
	redis *redis.Client
}

// MessageFilters contains filtering options for message queries
type MessageFilters struct {
	PeerID    string // Filter by conversation peer
	Since     string // RFC3339 timestamp - messages after this time
	Status    string // Filter by delivery status
	Direction string // "sent", "received", or "all" (default)
}

// NewService creates a new message service
func NewService(db *gorm.DB, redis *redis.Client) *Service {
	return &Service{
		db:    db,
		redis: redis,
	}
}

// GetUserMessages retrieves all messages for a user with pagination and filters
func (s *Service) GetUserMessages(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
	filters *MessageFilters,
) ([]model.ChatMessage, int64, error) {
	var messages []model.ChatMessage
	var total int64

	// Base query
	query := s.db.WithContext(ctx).Model(&model.ChatMessage{})

	// Apply direction filter
	if filters != nil && filters.Direction != "" {
		switch filters.Direction {
		case "sent":
			query = query.Where("sender_id = ?", userID)
		case "received":
			query = query.Where("receiver_id = ?", userID)
		default: // "all"
			query = query.Where("sender_id = ? OR receiver_id = ?", userID, userID)
		}
	} else {
		// Default: all messages
		query = query.Where("sender_id = ? OR receiver_id = ?", userID, userID)
	}

	// Apply peer filter
	if filters != nil && filters.PeerID != "" {
		if peerID, err := uuid.Parse(filters.PeerID); err == nil {
			query = query.Where(
				"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
				userID, peerID, peerID, userID,
			)
		}
	}

	// Apply time filter
	if filters != nil && filters.Since != "" {
		if sinceTime, err := time.Parse(time.RFC3339, filters.Since); err == nil {
			query = query.Where("timestamp > ?", sinceTime)
		}
	}

	// Apply status filter
	if filters != nil && filters.Status != "" {
		query = query.Where("delivery_status = ?", filters.Status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	// Fetch with pagination
	if err := query.
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch messages: %w", err)
	}

	return messages, total, nil
}

// GetConversation retrieves bidirectional conversation between two users
func (s *Service) GetConversation(
	ctx context.Context,
	userID, peerID uuid.UUID,
	limit, offset int,
	beforeTime, afterTime *time.Time,
) ([]model.ChatMessage, int64, error) {
	var messages []model.ChatMessage
	var total int64

	// Base query: bidirectional conversation
	query := s.db.WithContext(ctx).
		Model(&model.ChatMessage{}).
		Where(
			"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID, peerID, peerID, userID,
		)

	// Apply time filters
	if beforeTime != nil {
		query = query.Where("timestamp < ?", beforeTime)
	}
	if afterTime != nil {
		query = query.Where("timestamp > ?", afterTime)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	// Fetch with pagination
	if err := query.
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch messages: %w", err)
	}

	return messages, total, nil
}

// GetMessagesByIDs retrieves messages by ID list (for offline queue delivery)
func (s *Service) GetMessagesByIDs(
	ctx context.Context,
	messageIDs []uuid.UUID,
) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage

	if err := s.db.WithContext(ctx).
		Where("id IN ?", messageIDs).
		Order("timestamp ASC"). // Deliver in chronological order
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch messages by IDs: %w", err)
	}

	return messages, nil
}

// CreateAndSendMessage creates message and handles delivery/queueing
func (s *Service) CreateAndSendMessage(
	ctx context.Context,
	text string,
	senderID, receiverID uuid.UUID,
) (*model.ChatMessage, error) {
	// Create message
	message := model.NewTextMessage(text, senderID, receiverID)

	// Save to database
	if err := s.db.WithContext(ctx).Create(message).Error; err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Check if receiver is online (via Redis session check)
	if s.redis != nil {
		sessionKey := fmt.Sprintf("user:%s:session", receiverID.String())
		isOnline, _ := s.redis.Exists(ctx, sessionKey).Result()

		if isOnline > 0 {
			// Mark as sent (WebSocket will update to DELIVERED)
			message.DeliveryStatus = model.MessageSent
			s.db.WithContext(ctx).Model(message).Update("delivery_status", model.MessageSent)
		} else {
			// Will be enqueued by caller
			message.DeliveryStatus = model.MessagePending
			s.db.WithContext(ctx).Model(message).Update("delivery_status", model.MessagePending)
		}
	}

	return message, nil
}

// MarkMessagesAsRead marks multiple messages as read (batch operation)
func (s *Service) MarkMessagesAsRead(
	ctx context.Context,
	messageIDs []uuid.UUID,
	receiverID uuid.UUID,
) ([]model.ChatMessage, error) {
	now := time.Now()

	// Batch update - only update messages that are not already read
	result := s.db.WithContext(ctx).
		Model(&model.ChatMessage{}).
		Where("id IN ? AND receiver_id = ? AND is_read = false", messageIDs, receiverID).
		Updates(map[string]interface{}{
			"is_read":         true,
			"read_at":         now,
			"delivery_status": model.MessageRead,
			"updated_at":      now,
		})

	if result.Error != nil {
		return nil, fmt.Errorf("failed to mark messages as read: %w", result.Error)
	}

	// Fetch updated messages
	var messages []model.ChatMessage
	if err := s.db.WithContext(ctx).
		Where("id IN ?", messageIDs).
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch updated messages: %w", err)
	}

	return messages, nil
}

// MarkMessageAsRead marks single message as read (convenience method)
func (s *Service) MarkMessageAsRead(
	ctx context.Context,
	messageID, receiverID uuid.UUID,
) (*model.ChatMessage, error) {
	messages, err := s.MarkMessagesAsRead(ctx, []uuid.UUID{messageID}, receiverID)
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		return nil, fmt.Errorf("message not found or already read")
	}
	return &messages[0], nil
}

// UpdateDeliveryStatus updates message delivery status
func (s *Service) UpdateDeliveryStatus(
	ctx context.Context,
	messageID uuid.UUID,
	status model.MessageDeliveryStatus,
) error {
	updates := map[string]interface{}{
		"delivery_status": status,
		"updated_at":      time.Now(),
	}

	// Add timestamp for specific statuses
	if status == model.MessageDelivered {
		updates["delivered_at"] = time.Now()
	}

	result := s.db.WithContext(ctx).
		Model(&model.ChatMessage{}).
		Where("id = ?", messageID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("message not found: %s", messageID)
	}

	return nil
}

// BatchUpdateDeliveryStatus updates multiple messages (for batch delivery)
func (s *Service) BatchUpdateDeliveryStatus(
	ctx context.Context,
	messageIDs []uuid.UUID,
	status model.MessageDeliveryStatus,
) error {
	updates := map[string]interface{}{
		"delivery_status": status,
		"updated_at":      time.Now(),
	}

	if status == model.MessageDelivered {
		updates["delivered_at"] = time.Now()
	}

	return s.db.WithContext(ctx).
		Model(&model.ChatMessage{}).
		Where("id IN ?", messageIDs).
		Updates(updates).Error
}

// GetMessageStatus retrieves current status of a message
func (s *Service) GetMessageStatus(
	ctx context.Context,
	messageID uuid.UUID,
) (*model.MessageDeliveryStatus, error) {
	var message model.ChatMessage
	if err := s.db.WithContext(ctx).
		Select("delivery_status").
		First(&message, messageID).Error; err != nil {
		return nil, err
	}

	return &message.DeliveryStatus, nil
}
