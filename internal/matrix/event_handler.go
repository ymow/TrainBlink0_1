package matrix

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/gorm"
)

// MessageEventHandler handles Matrix message events
type MessageEventHandler struct {
	db *gorm.DB
}

// NewMessageEventHandler creates a new message event handler
func NewMessageEventHandler(db *gorm.DB) *MessageEventHandler {
	return &MessageEventHandler{db: db}
}

// GetHandlerName returns the name of the handler
func (h *MessageEventHandler) GetHandlerName() string {
	return "MessageEventHandler"
}

// HandleEvent processes a Matrix event
func (h *MessageEventHandler) HandleEvent(ctx context.Context, event *Event, roomID string) error {
	switch event.Type {
	case EventTypeRoomMessage:
		return h.handleMessage(ctx, event, roomID)
	case EventTypeRoomMember:
		return h.handleMembership(ctx, event, roomID)
	case EventTypeReceipt:
		return h.handleReceipt(ctx, event, roomID)
	case EventTypeTyping:
		return h.handleTyping(ctx, event, roomID)
	case EventTypePresence:
		return h.handlePresence(ctx, event)
	case EventTypeRoomName:
		return h.handleRoomName(ctx, event, roomID)
	case EventTypeRoomTopic:
		return h.handleRoomTopic(ctx, event, roomID)
	default:
		// Log unknown event types for debugging
		if event.Type != "" {
			log.Printf("🔍 Unhandled event type: %s in room %s", event.Type, roomID)
		}
	}

	return nil
}

// handleMessage processes m.room.message events
func (h *MessageEventHandler) handleMessage(ctx context.Context, event *Event, roomID string) error {
	msgType, ok := event.Content["msgtype"].(string)
	if !ok {
		return fmt.Errorf("missing msgtype in message event")
	}

	body, ok := event.Content["body"].(string)
	if !ok {
		return fmt.Errorf("missing body in message event")
	}

	log.Printf("💬 Message in %s from %s [%s]: %s",
		roomID, event.Sender, msgType, truncateString(body, 100))

	// Store message in database if db is available
	if h.db != nil {
		contentJSON, _ := json.Marshal(event.Content)

		message := &MatrixMessage{
			EventID:        event.EventID,
			RoomID:         roomID,
			Sender:         event.Sender,
			MsgType:        msgType,
			Body:           body,
			OriginServerTS: event.OriginServerTS,
			Content:        contentJSON,
		}

		if err := h.db.WithContext(ctx).Create(message).Error; err != nil {
			log.Printf("⚠️  Failed to store message: %v", err)
			// Don't return error - continue processing
		}
	}

	// TODO: Additional processing
	// - Content moderation (AI safety engine)
	// - Analytics tracking
	// - Push notifications
	// - Webhook triggers

	return nil
}

// handleMembership processes m.room.member events
func (h *MessageEventHandler) handleMembership(ctx context.Context, event *Event, roomID string) error {
	membership, ok := event.Content["membership"].(string)
	if !ok {
		return fmt.Errorf("missing membership in member event")
	}

	var stateKey string
	if event.StateKey != nil {
		stateKey = *event.StateKey
	}

	displayName := ""
	if name, ok := event.Content["displayname"].(string); ok {
		displayName = name
	}

	switch membership {
	case MembershipJoin:
		log.Printf("👋 %s joined %s (display: %s)", stateKey, roomID, displayName)
		// TODO: Update room member count in cache
		// TODO: Send welcome message if configured
	case MembershipLeave:
		log.Printf("👋 %s left %s", stateKey, roomID)
		// TODO: Update room member count in cache
	case MembershipInvite:
		log.Printf("📨 %s invited to %s", stateKey, roomID)
	case MembershipBan:
		reason := ""
		if r, ok := event.Content["reason"].(string); ok {
			reason = r
		}
		log.Printf("🚫 %s banned from %s (reason: %s)", stateKey, roomID, reason)
	}

	return nil
}

// handleReceipt processes m.receipt events (read receipts)
func (h *MessageEventHandler) handleReceipt(ctx context.Context, event *Event, roomID string) error {
	// Extract receipt data: event.Content is a map of event_id -> receipt type -> user -> data
	for eventID, receiptTypes := range event.Content {
		if receiptTypesMap, ok := receiptTypes.(map[string]interface{}); ok {
			if readReceipts, ok := receiptTypesMap["m.read"].(map[string]interface{}); ok {
				for userID, receiptData := range readReceipts {
					if dataMap, ok := receiptData.(map[string]interface{}); ok {
						if ts, ok := dataMap["ts"].(float64); ok {
							log.Printf("📖 Read receipt: %s read %s in %s at %d",
								userID, eventID, roomID, int64(ts))

							// TODO: Store read receipts for analytics
							// TODO: Update unread message counts
						}
					}
				}
			}
		}
	}

	return nil
}

// handleTyping processes m.typing events
func (h *MessageEventHandler) handleTyping(ctx context.Context, event *Event, roomID string) error {
	userIDs, ok := event.Content["user_ids"].([]interface{})
	if !ok {
		return nil
	}

	if len(userIDs) > 0 {
		log.Printf("⌨️  Typing in %s: %d users", roomID, len(userIDs))
	}

	// TODO: Broadcast typing indicators to connected clients via WebSocket
	// TODO: Cache typing status in Redis with TTL

	return nil
}

// handlePresence processes m.presence events
func (h *MessageEventHandler) handlePresence(ctx context.Context, event *Event) error {
	presence, ok := event.Content["presence"].(string)
	if !ok {
		return nil
	}

	statusMsg := ""
	if msg, ok := event.Content["status_msg"].(string); ok {
		statusMsg = msg
	}

	log.Printf("🟢 Presence: %s is %s (%s)", event.Sender, presence, statusMsg)

	// TODO: Update presence cache in Redis
	// TODO: Broadcast presence updates to relevant users

	return nil
}

// handleRoomName processes m.room.name events
func (h *MessageEventHandler) handleRoomName(ctx context.Context, event *Event, roomID string) error {
	name, ok := event.Content["name"].(string)
	if !ok {
		return nil
	}

	log.Printf("📝 Room %s renamed to: %s", roomID, name)

	// TODO: Update room metadata in database/cache

	return nil
}

// handleRoomTopic processes m.room.topic events
func (h *MessageEventHandler) handleRoomTopic(ctx context.Context, event *Event, roomID string) error {
	topic, ok := event.Content["topic"].(string)
	if !ok {
		return nil
	}

	log.Printf("📝 Room %s topic changed to: %s", roomID, truncateString(topic, 100))

	// TODO: Update room metadata in database/cache

	return nil
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ============================================================
// Database Model for Matrix Messages
// ============================================================

// MatrixMessage represents a stored Matrix message
type MatrixMessage struct {
	ID             int64  `gorm:"primaryKey;autoIncrement"`
	EventID        string `gorm:"uniqueIndex;not null"`
	RoomID         string `gorm:"index;not null"`
	Sender         string `gorm:"index;not null"`
	MsgType        string `gorm:"not null"`
	Body           string `gorm:"type:text;not null"`
	OriginServerTS int64  `gorm:"index;not null"`
	Content        []byte `gorm:"type:jsonb"`
	CreatedAt      int64  `gorm:"autoCreateTime:milli"`
}

// TableName specifies the table name for MatrixMessage
func (MatrixMessage) TableName() string {
	return "matrix_messages"
}

// ============================================================
// Additional Event Handlers
// ============================================================

// LoggingEventHandler logs all events for debugging
type LoggingEventHandler struct {
	verbose bool
}

// NewLoggingEventHandler creates a logging event handler
func NewLoggingEventHandler(verbose bool) *LoggingEventHandler {
	return &LoggingEventHandler{verbose: verbose}
}

// GetHandlerName returns the name of the handler
func (h *LoggingEventHandler) GetHandlerName() string {
	return "LoggingEventHandler"
}

// HandleEvent logs the event
func (h *LoggingEventHandler) HandleEvent(ctx context.Context, event *Event, roomID string) error {
	if !h.verbose {
		return nil
	}

	if event.Type == "" {
		return nil
	}

	log.Printf("📋 Event: type=%s, room=%s, sender=%s, event_id=%s",
		event.Type, roomID, event.Sender, event.EventID)

	return nil
}

// ============================================================
// Analytics Event Handler
// ============================================================

// AnalyticsEventHandler tracks analytics for events
type AnalyticsEventHandler struct {
	// TODO: Add analytics service dependency
}

// NewAnalyticsEventHandler creates an analytics event handler
func NewAnalyticsEventHandler() *AnalyticsEventHandler {
	return &AnalyticsEventHandler{}
}

// GetHandlerName returns the name of the handler
func (h *AnalyticsEventHandler) GetHandlerName() string {
	return "AnalyticsEventHandler"
}

// HandleEvent tracks analytics for the event
func (h *AnalyticsEventHandler) HandleEvent(ctx context.Context, event *Event, roomID string) error {
	// TODO: Track event metrics
	// - Message count per room
	// - Active users
	// - Message types distribution
	// - Peak activity times
	// - User engagement metrics

	return nil
}

// ============================================================
// Moderation Event Handler
// ============================================================

// ModerationEventHandler handles content moderation
type ModerationEventHandler struct {
	// TODO: Add AI safety engine dependency
}

// NewModerationEventHandler creates a moderation event handler
func NewModerationEventHandler() *ModerationEventHandler {
	return &ModerationEventHandler{}
}

// GetHandlerName returns the name of the handler
func (h *ModerationEventHandler) GetHandlerName() string {
	return "ModerationEventHandler"
}

// HandleEvent performs content moderation on the event
func (h *ModerationEventHandler) HandleEvent(ctx context.Context, event *Event, roomID string) error {
	// Only moderate message events
	if event.Type != EventTypeRoomMessage {
		return nil
	}

	// TODO: Implement content moderation
	// - Spam detection
	// - Profanity filtering
	// - Harmful content detection (violence, etc.)
	// - Automatic actions (warn, kick, ban)

	return nil
}
