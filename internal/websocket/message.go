package websocket

import (
	"fmt"
	"time"

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
	default:
		fmt.Printf("Unknown message type: %s\n", msg.Type)
		client.sendError("UNKNOWN_MESSAGE_TYPE", fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}

// handleChatMessage handles regular chat messages
func (h *MessageHandler) handleChatMessage(client *Client, msg *model.ClientMessage) {
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
