package mls

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// DeliveryService handles MLS message delivery
type DeliveryService struct {
	db       *gorm.DB
	redis    *redis.Client
	groupMgr *GroupManager
	config   *MLSConfig
}

// NewDeliveryService creates a new delivery service
func NewDeliveryService(db *gorm.DB, redis *redis.Client, groupMgr *GroupManager, config *MLSConfig) *DeliveryService {
	if config == nil {
		config = DefaultMLSConfig()
	}

	return &DeliveryService{
		db:       db,
		redis:    redis,
		groupMgr: groupMgr,
		config:   config,
	}
}

// SendMessage handles MLS message delivery
func (ds *DeliveryService) SendMessage(
	ctx context.Context,
	senderClientID string,
	groupID string,
	messageType string,
	ciphertext []byte,
	authenticatedData []byte,
) (*MLSMessage, error) {
	// 1. Validate message size
	if len(ciphertext) > ds.config.MaxMessageSize {
		return nil, fmt.Errorf("message size exceeds maximum (%d bytes)", ds.config.MaxMessageSize)
	}

	// 2. Validate group and membership
	group, err := ds.groupMgr.GetGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	if !group.IsActive {
		return nil, fmt.Errorf("group is not active")
	}

	isMember, err := ds.groupMgr.IsActiveMember(ctx, groupID, senderClientID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("sender is not an active member")
	}

	// 3. Get next sequence number (atomic)
	seqKey := fmt.Sprintf("mls:group:%s:sequence", groupID)
	seqNum, err := ds.redis.Incr(ctx, seqKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get sequence number: %w", err)
	}

	// 4. Create message record
	messageID := uuid.New().String()
	message := &MLSMessage{
		ID:                uuid.New(),
		MessageID:         messageID,
		GroupID:           groupID,
		SenderClientID:    senderClientID,
		MessageType:       messageType,
		Epoch:             group.Epoch,
		SequenceNumber:    seqNum,
		Ciphertext:        ciphertext,
		AuthenticatedData: authenticatedData,
		SentAt:            time.Now(),
		RecipientCount:    group.MemberCount - 1, // Exclude sender
		DeliveredCount:    0,
	}

	// 5. Persist to database
	if err := ds.db.WithContext(ctx).Create(message).Error; err != nil {
		return nil, fmt.Errorf("failed to persist message: %w", err)
	}

	// 6. Get active recipients
	recipients, err := ds.groupMgr.GetActiveMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipients: %w", err)
	}

	// 7. Fan-out to recipients (async)
	go ds.deliverToRecipients(context.Background(), message, recipients)

	log.Printf("✉️  Sent MLS message %s to group %s (epoch: %d, seq: %d)",
		messageID[:16], groupID, message.Epoch, message.SequenceNumber)

	return message, nil
}

// deliverToRecipients sends message to all active members
func (ds *DeliveryService) deliverToRecipients(
	ctx context.Context,
	message *MLSMessage,
	recipients []*MLSGroupMember,
) {
	deliveredCount := 0

	for _, recipient := range recipients {
		// Skip sender
		if recipient.ClientID == message.SenderClientID {
			continue
		}

		// Check if client is connected (WebSocket)
		sessionKey := fmt.Sprintf("mls:session:%s", recipient.ClientID)
		isConnected, _ := ds.redis.Exists(ctx, sessionKey).Result()

		if isConnected > 0 {
			// Send via WebSocket (placeholder - integrate with WebSocket system)
			if err := ds.sendViaWebSocket(ctx, recipient.ClientID, message); err == nil {
				deliveredCount++
			}
		} else {
			// Queue for later delivery
			ds.queueForDelivery(ctx, recipient.ClientID, message.MessageID)
		}
	}

	// Update delivery count
	ds.db.WithContext(ctx).Model(&MLSMessage{}).
		Where("id = ?", message.ID).
		Updates(map[string]interface{}{
			"delivered_count": deliveredCount,
			"processed_at":    time.Now(),
		})
}

// GetPendingMessages retrieves messages for a client
func (ds *DeliveryService) GetPendingMessages(
	ctx context.Context,
	clientID string,
	groupID string,
	afterSequence int64,
	limit int,
) ([]*MLSMessage, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	var messages []*MLSMessage

	if err := ds.db.WithContext(ctx).
		Where("group_id = ? AND sequence_number > ? AND sender_client_id != ?", groupID, afterSequence, clientID).
		Order("sequence_number ASC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

// AcknowledgeDelivery marks message as delivered to a client
func (ds *DeliveryService) AcknowledgeDelivery(
	ctx context.Context,
	messageID string,
	clientID string,
) error {
	// Increment delivered count
	result := ds.db.WithContext(ctx).Model(&MLSMessage{}).
		Where("message_id = ?", messageID).
		Update("delivered_count", gorm.Expr("delivered_count + ?", 1))

	return result.Error
}

// GetStats returns delivery statistics
func (ds *DeliveryService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	var totalMessages, pendingMessages int64

	ds.db.WithContext(ctx).Model(&MLSMessage{}).Count(&totalMessages)
	ds.db.WithContext(ctx).Model(&MLSMessage{}).
		Where("delivered_count < recipient_count").
		Count(&pendingMessages)

	return map[string]interface{}{
		"total_messages":   totalMessages,
		"pending_messages": pendingMessages,
	}, nil
}

// Placeholder functions for WebSocket integration
func (ds *DeliveryService) sendViaWebSocket(ctx context.Context, clientID string, message *MLSMessage) error {
	// TODO: Implement WebSocket delivery
	// This should integrate with the existing WebSocket hub from Phase 1
	log.Printf("📬 Would deliver message %s to client %s via WebSocket", message.MessageID[:16], clientID)
	return nil
}

func (ds *DeliveryService) queueForDelivery(ctx context.Context, clientID string, messageID string) {
	// Queue message for later delivery
	queueKey := fmt.Sprintf("mls:queue:%s", clientID)
	if err := ds.redis.RPush(ctx, queueKey, messageID).Err(); err != nil {
		log.Printf("⚠️  Failed to queue message: %v", err)
	}
}
