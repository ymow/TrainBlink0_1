package websocket

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ymow/messenger_protocol_research/internal/model"
)

// MessageHandler handles different types of WebSocket messages
type MessageHandler struct {
	hub *Hub
}

// NewMessageHandler creates a new MessageHandler
func NewMessageHandler(hub *Hub) *MessageHandler {
	return &MessageHandler{
		hub: hub,
	}
}

// HandleMessage processes incoming client messages
func (h *MessageHandler) HandleMessage(client *Client, msg *model.ClientMessage) {
	switch msg.Type {
	case model.MessageTypeMessage:
		h.handleChatMessage(client, msg)
	case model.MessageTypeBroadcast:
		h.handleBroadcastMessage(client, msg)
	case model.MessageTypeTyping:
		h.handleTypingIndicator(client, msg)
	case model.MessageTypePresence:
		h.handlePresenceUpdate(client, msg)
	case model.MessageTypePing:
		h.handlePing(client)
	case model.MessageTypeReadReceipt:
		h.handleReadReceipt(client, msg)
	case model.MessageTypeDeliveryAck:
		h.handleDeliveryAck(client, msg)
	default:
		fmt.Printf("Unknown message type: %s\n", msg.Type)
		client.sendError("UNKNOWN_MESSAGE_TYPE", fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}

// handleChatMessage handles regular chat messages
func (h *MessageHandler) handleChatMessage(client *Client, msg *model.ClientMessage) {
	// Check rate limit
	if !client.checkRateLimit(false) {
		client.sendError("RATE_LIMIT_EXCEEDED", "Message rate limit exceeded. Please slow down.")
		fmt.Printf("Rate limit exceeded for user %s\n", client.UserID)
		return
	}

	// Validate message content
	if len(msg.Content) == 0 {
		client.sendError("EMPTY_MESSAGE", "Message content cannot be empty")
		return
	}

	if len(msg.Content) > maxMessageSize {
		client.sendError("MESSAGE_TOO_LARGE", "Message content exceeds maximum size")
		return
	}

	// Create server message
	serverMsg := &model.ServerMessage{
		ID:        generateMessageID(),
		Type:      model.MessageTypeMessage,
		From:      client.UserID,
		To:        msg.To,
		StationID: client.StationID,
		Content:   msg.Content,
		Timestamp: time.Now(),
		Metadata:  msg.Metadata,
	}

	// If it's a direct message (To is specified)
	if msg.To != "" {
		// PERMISSION CHECK: Verify sender can message receiver
		receiverUUID, err := uuid.Parse(msg.To)
		if err != nil {
			client.sendError("INVALID_RECEIVER", "Invalid receiver ID")
			return
		}

		senderUUID, err := uuid.Parse(client.UserID)
		if err != nil {
			client.sendError("INVALID_SENDER", "Invalid sender ID")
			return
		}

		// Check if sender has permission to message receiver
		canSend, reason, err := h.hub.permissionService.CanSendMessage(
			context.Background(),
			senderUUID,
			receiverUUID,
		)

		if err != nil {
			client.sendError("PERMISSION_CHECK_FAILED",
				fmt.Sprintf("Failed to check permission: %v", err))
			fmt.Printf("⚠️  Permission check error: %v\n", err)
			return
		}

		if !canSend {
			// Send user-friendly error based on reason
			errorMsg := "You cannot send messages to this user"
			switch reason {
			case "NO_VALID_DISCOVERY":
				errorMsg = "You must be within BLE range (50-100m) to message this user. Discovery expires after 10 minutes."
			case "SENDER_BLOCKED_BY_RECEIVER":
				errorMsg = "This user has blocked you"
			case "SELF_MESSAGE_NOT_ALLOWED":
				errorMsg = "You cannot message yourself"
			}

			client.sendError(reason, errorMsg)
			fmt.Printf("🚫 Message blocked: %s -> %s (reason: %s)\n",
				client.UserID, msg.To, reason)
			return
		}
		// END PERMISSION CHECK

		// Send to specific user
		h.hub.SendToUser(msg.To, serverMsg)

		// Send acknowledgment to sender
		client.sendAck(serverMsg.ID)

		fmt.Printf("Direct message from %s to %s: %s\n", client.UserID, msg.To, msg.Content)
	} else {
		// Broadcast to station
		h.hub.BroadcastToStation(client.StationID, serverMsg, client)

		// Send acknowledgment to sender
		client.sendAck(serverMsg.ID)

		fmt.Printf("Station message from %s in %s: %s\n", client.UserID, client.StationID, msg.Content)
	}
}

// handleBroadcastMessage handles station-wide broadcast messages
func (h *MessageHandler) handleBroadcastMessage(client *Client, msg *model.ClientMessage) {
	// Check rate limit (broadcasts have stricter limits)
	if !client.checkRateLimit(true) {
		client.sendError("RATE_LIMIT_EXCEEDED", "Broadcast rate limit exceeded. Please slow down.")
		fmt.Printf("Broadcast rate limit exceeded for user %s\n", client.UserID)
		return
	}

	// Validate message content
	if len(msg.Content) == 0 {
		client.sendError("EMPTY_MESSAGE", "Broadcast content cannot be empty")
		return
	}

	if len(msg.Content) > maxMessageSize {
		client.sendError("MESSAGE_TOO_LARGE", "Broadcast content exceeds maximum size")
		return
	}

	// Create server message
	serverMsg := &model.ServerMessage{
		ID:        generateMessageID(),
		Type:      model.MessageTypeBroadcast,
		From:      client.UserID,
		StationID: client.StationID,
		Content:   msg.Content,
		Timestamp: time.Now(),
		Metadata:  msg.Metadata,
	}

	// Broadcast to entire station
	h.hub.BroadcastToStation(client.StationID, serverMsg, client)

	// Send acknowledgment to sender
	client.sendAck(serverMsg.ID)

	fmt.Printf("Broadcast from %s in %s: %s\n", client.UserID, client.StationID, msg.Content)
}

// handleTypingIndicator handles typing indicator events
func (h *MessageHandler) handleTypingIndicator(client *Client, msg *model.ClientMessage) {
	// Extract typing status from metadata
	isTyping := false
	if msg.Metadata != nil {
		if val, ok := msg.Metadata["is_typing"].(bool); ok {
			isTyping = val
		}
	}

	// Create typing indicator event
	typingIndicator := &model.TypingIndicator{
		UserID:    client.UserID,
		StationID: client.StationID,
		IsTyping:  isTyping,
		Timestamp: time.Now(),
	}

	// Create server message
	serverMsg := &model.ServerMessage{
		ID:        generateMessageID(),
		Type:      model.MessageTypeTyping,
		From:      client.UserID,
		To:        msg.To,
		StationID: client.StationID,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"typing_indicator": typingIndicator,
		},
	}

	// If it's a direct typing indicator (To is specified)
	if msg.To != "" {
		// Send to specific user
		h.hub.SendToUser(msg.To, serverMsg)
	} else {
		// Broadcast to station (exclude sender)
		h.hub.BroadcastToStation(client.StationID, serverMsg, client)
	}

	fmt.Printf("Typing indicator from %s in %s: %v\n", client.UserID, client.StationID, isTyping)
}

// handlePresenceUpdate handles user presence updates
func (h *MessageHandler) handlePresenceUpdate(client *Client, msg *model.ClientMessage) {
	// Extract presence status from metadata
	var status model.PresenceStatus = model.PresenceOnline
	if msg.Metadata != nil {
		if val, ok := msg.Metadata["status"].(string); ok {
			status = model.PresenceStatus(val)
		}
	}

	// Create presence update event
	presenceUpdate := &model.PresenceUpdate{
		UserID:    client.UserID,
		StationID: client.StationID,
		Status:    status,
		Timestamp: time.Now(),
	}

	// Create server message
	serverMsg := &model.ServerMessage{
		ID:        generateMessageID(),
		Type:      model.MessageTypePresence,
		From:      client.UserID,
		StationID: client.StationID,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"presence_update": presenceUpdate,
		},
	}

	// Broadcast to station (exclude sender)
	h.hub.BroadcastToStation(client.StationID, serverMsg, client)

	fmt.Printf("Presence update from %s in %s: %s\n", client.UserID, client.StationID, status)
}

// handlePing handles ping messages (heartbeat)
func (h *MessageHandler) handlePing(client *Client) {
	// Update last ping time
	client.LastPingAt = time.Now()

	// Send pong response
	pongMsg := &model.ServerMessage{
		ID:        generateMessageID(),
		Type:      model.MessageTypePong,
		Timestamp: time.Now(),
	}

	select {
	case client.send <- pongMsg:
	default:
		fmt.Printf("Failed to send pong to client: send channel full\n")
	}
}

// handleReadReceipt handles read receipt events from clients
func (h *MessageHandler) handleReadReceipt(client *Client, msg *model.ClientMessage) {
	// Parse message IDs from metadata
	var messageIDs []uuid.UUID
	if msg.Metadata != nil {
		if ids, ok := msg.Metadata["message_ids"].([]interface{}); ok {
			for _, id := range ids {
				if idStr, ok := id.(string); ok {
					if msgID, err := uuid.Parse(idStr); err == nil {
						messageIDs = append(messageIDs, msgID)
					}
				}
			}
		}
	}

	if len(messageIDs) == 0 {
		client.sendError("INVALID_READ_RECEIPT", "No valid message IDs provided")
		return
	}

	// Mark as read in database
	ctx := context.Background()
	userID, err := uuid.Parse(client.UserID)
	if err != nil {
		client.sendError("INVALID_USER_ID", "Invalid user ID")
		return
	}

	updatedMessages, err := h.hub.messageService.MarkMessagesAsRead(ctx, messageIDs, userID)
	if err != nil {
		client.sendError("READ_RECEIPT_FAILED", err.Error())
		fmt.Printf("⚠️  Failed to mark messages as read: %v\n", err)
		return
	}

	// Send read acknowledgments to original senders
	for _, message := range updatedMessages {
		readAckMsg := &model.ServerMessage{
			ID:        uuid.New().String(),
			Type:      model.MessageTypeReadAck,
			From:      client.UserID,
			To:        message.SenderID.String(),
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"message_id": message.ID.String(),
				"read_at":    message.ReadAt,
			},
		}

		// Send to original sender
		h.hub.SendToUser(message.SenderID.String(), readAckMsg)
	}

	fmt.Printf("📖 Read receipt: user %s read %d messages\n", client.UserID, len(updatedMessages))
}

// handleDeliveryAck handles delivery acknowledgment from client
func (h *MessageHandler) handleDeliveryAck(client *Client, msg *model.ClientMessage) {
	var messageID uuid.UUID
	if msg.Metadata != nil {
		if idStr, ok := msg.Metadata["message_id"].(string); ok {
			messageID, _ = uuid.Parse(idStr)
		}
	}

	if messageID == uuid.Nil {
		client.sendError("INVALID_DELIVERY_ACK", "No valid message ID provided")
		return
	}

	// Update to DELIVERED status
	ctx := context.Background()
	if err := h.hub.messageService.UpdateDeliveryStatus(ctx, messageID, model.MessageDelivered); err != nil {
		fmt.Printf("⚠️  Failed to update delivery status: %v\n", err)
		client.sendError("DELIVERY_ACK_FAILED", err.Error())
	} else {
		fmt.Printf("✅ Message %s delivered to user %s\n", messageID, client.UserID)
	}
}
