package websocket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/ymow/messenger_protocol_research/internal/cache"
	"github.com/ymow/messenger_protocol_research/internal/message"
	"github.com/ymow/messenger_protocol_research/internal/model"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Clients by user ID
	clientsByUser map[string]*Client

	// Clients by station ID (station rooms)
	clientsByStation map[string]map[*Client]bool

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to all clients in a station
	broadcast chan *BroadcastMessage

	// Direct message to a specific user
	directMessage chan *DirectMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Statistics
	stats *HubStats

	// Phase 1: Offline message queue support
	redis          *redis.Client
	messageService *message.Service
}

// BroadcastMessage represents a message to broadcast to a station
type BroadcastMessage struct {
	StationID string
	Message   *model.ServerMessage
	Exclude   *Client // Optional: exclude this client from broadcast
}

// DirectMessage represents a message to send to a specific user
type DirectMessage struct {
	UserID  string
	Message *model.ServerMessage
}

// HubStats contains hub statistics
type HubStats struct {
	TotalClients     int            `json:"total_clients"`
	ClientsByStation map[string]int `json:"clients_by_station"`
	MessagesHandled  int64          `json:"messages_handled"`
	mu               sync.RWMutex
}

// NewHub creates a new Hub
func NewHub(redis *redis.Client, msgService *message.Service) *Hub {
	return &Hub{
		clients:          make(map[*Client]bool),
		clientsByUser:    make(map[string]*Client),
		clientsByStation: make(map[string]map[*Client]bool),
		register:         make(chan *Client, 10),
		unregister:       make(chan *Client, 10),
		broadcast:        make(chan *BroadcastMessage, 256),
		directMessage:    make(chan *DirectMessage, 256),
		stats: &HubStats{
			ClientsByStation: make(map[string]int),
		},
		redis:          redis,
		messageService: msgService,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	fmt.Println("WebSocket Hub started")
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case broadcastMsg := <-h.broadcast:
			h.broadcastToStation(broadcastMsg)

		case directMsg := <-h.directMessage:
			h.sendDirectMessage(directMsg)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add to clients map
	h.clients[client] = true

	// Add to user map
	if client.UserID != "" {
		h.clientsByUser[client.UserID] = client
	}

	// Add to station room
	if client.StationID != "" {
		if _, ok := h.clientsByStation[client.StationID]; !ok {
			h.clientsByStation[client.StationID] = make(map[*Client]bool)
		}
		h.clientsByStation[client.StationID][client] = true

		// Update stats
		h.stats.mu.Lock()
		h.stats.ClientsByStation[client.StationID]++
		h.stats.TotalClients++
		h.stats.mu.Unlock()
	}

	fmt.Printf("Client registered: UserID=%s, StationID=%s, Total clients=%d\n",
		client.UserID, client.StationID, len(h.clients))

	// Notify station about new join
	if client.StationID != "" {
		joinEvent := &model.JoinEvent{
			UserID:    client.UserID,
			DeviceID:  client.DeviceID,
			StationID: client.StationID,
			Timestamp: time.Now(),
		}

		h.broadcastToStation(&BroadcastMessage{
			StationID: client.StationID,
			Message: &model.ServerMessage{
				ID:        generateMessageID(),
				Type:      model.MessageTypeJoin,
				From:      client.UserID,
				StationID: client.StationID,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"join_event": joinEvent,
				},
			},
			Exclude: client, // Don't send join event to the joining user
		})
	}

	// Phase 1: Deliver offline messages after registration
	if h.redis != nil && h.messageService != nil && client.UserID != "" {
		go h.DeliverOfflineMessages(client)
	}
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		// Remove from clients map
		delete(h.clients, client)

		// Remove from user map
		if client.UserID != "" {
			delete(h.clientsByUser, client.UserID)
		}

		// Remove from station room
		if client.StationID != "" {
			if stationClients, ok := h.clientsByStation[client.StationID]; ok {
				delete(stationClients, client)
				if len(stationClients) == 0 {
					delete(h.clientsByStation, client.StationID)
				}

				// Update stats
				h.stats.mu.Lock()
				h.stats.ClientsByStation[client.StationID]--
				if h.stats.ClientsByStation[client.StationID] <= 0 {
					delete(h.stats.ClientsByStation, client.StationID)
				}
				h.stats.TotalClients--
				h.stats.mu.Unlock()
			}

			// Notify station about leave
			leaveEvent := &model.LeaveEvent{
				UserID:    client.UserID,
				StationID: client.StationID,
				Timestamp: time.Now(),
			}

			h.broadcastToStation(&BroadcastMessage{
				StationID: client.StationID,
				Message: &model.ServerMessage{
					ID:        generateMessageID(),
					Type:      model.MessageTypeLeave,
					From:      client.UserID,
					StationID: client.StationID,
					Timestamp: time.Now(),
					Metadata: map[string]interface{}{
						"leave_event": leaveEvent,
					},
				},
			})
		}

		// Close the client's send channel
		close(client.send)

		fmt.Printf("Client unregistered: UserID=%s, StationID=%s, Total clients=%d\n",
			client.UserID, client.StationID, len(h.clients))
	}
}

// broadcastToStation broadcasts a message to all clients in a station
func (h *Hub) broadcastToStation(broadcastMsg *BroadcastMessage) {
	h.mu.RLock()
	stationClients, ok := h.clientsByStation[broadcastMsg.StationID]
	h.mu.RUnlock()

	if !ok || len(stationClients) == 0 {
		return
	}

	// Update stats
	h.stats.mu.Lock()
	h.stats.MessagesHandled++
	h.stats.mu.Unlock()

	// Send to all clients in the station
	for client := range stationClients {
		// Skip excluded client (e.g., sender)
		if broadcastMsg.Exclude != nil && client == broadcastMsg.Exclude {
			continue
		}

		select {
		case client.send <- broadcastMsg.Message:
		default:
			// Client's send channel is full, unregister it
			h.unregister <- client
		}
	}
}

// sendDirectMessage sends a message to a specific user
func (h *Hub) sendDirectMessage(directMsg *DirectMessage) {
	h.mu.RLock()
	client, ok := h.clientsByUser[directMsg.UserID]
	h.mu.RUnlock()

	if !ok {
		fmt.Printf("User not connected: %s\n", directMsg.UserID)
		return
	}

	// Extract message ID for status tracking
	var messageID uuid.UUID
	if directMsg.Message.Metadata != nil {
		if idStr, ok := directMsg.Message.Metadata["message_id"].(string); ok {
			messageID, _ = uuid.Parse(idStr)
		}
	}

	// Update to SENDING status
	if messageID != uuid.Nil && h.messageService != nil {
		ctx := context.Background()
		h.messageService.UpdateDeliveryStatus(ctx, messageID, model.MessageSending)
	}

	// Update stats
	h.stats.mu.Lock()
	h.stats.MessagesHandled++
	h.stats.mu.Unlock()

	select {
	case client.send <- directMsg.Message:
		// Update to SENT status after successful send
		if messageID != uuid.Nil && h.messageService != nil {
			ctx := context.Background()
			h.messageService.UpdateDeliveryStatus(ctx, messageID, model.MessageSent)
		}
	default:
		// Send failed - mark as FAILED
		if messageID != uuid.Nil && h.messageService != nil {
			ctx := context.Background()
			h.messageService.UpdateDeliveryStatus(ctx, messageID, model.MessageFailed)
		}
		// Client's send channel is full, unregister it
		h.unregister <- client
	}
}

// BroadcastToStation sends a broadcast message to all clients in a station
func (h *Hub) BroadcastToStation(stationID string, message *model.ServerMessage, excludeClient *Client) {
	h.broadcast <- &BroadcastMessage{
		StationID: stationID,
		Message:   message,
		Exclude:   excludeClient,
	}
}

// SendToUser sends a direct message to a specific user
func (h *Hub) SendToUser(userID string, message *model.ServerMessage) {
	h.directMessage <- &DirectMessage{
		UserID:  userID,
		Message: message,
	}
}

// GetStats returns hub statistics
func (h *Hub) GetStats() *HubStats {
	h.stats.mu.RLock()
	defer h.stats.mu.RUnlock()

	// Create a copy to avoid race conditions
	statsCopy := &HubStats{
		TotalClients:     h.stats.TotalClients,
		MessagesHandled:  h.stats.MessagesHandled,
		ClientsByStation: make(map[string]int),
	}

	for k, v := range h.stats.ClientsByStation {
		statsCopy.ClientsByStation[k] = v
	}

	return statsCopy
}

// GetClientsInStation returns the number of clients in a station
func (h *Hub) GetClientsInStation(stationID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if stationClients, ok := h.clientsByStation[stationID]; ok {
		return len(stationClients)
	}
	return 0
}

// GetDetailedStats returns detailed WebSocket statistics including client info
func (h *Hub) GetDetailedStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.stats.mu.RLock()
	defer h.stats.mu.RUnlock()

	// Collect client statistics
	clientStats := make([]map[string]interface{}, 0)
	totalMessages := int64(0)
	totalBroadcasts := int64(0)

	for client := range h.clients {
		stats := client.GetStats()
		clientStats = append(clientStats, stats)
		totalMessages += stats["total_messages"].(int64)
		totalBroadcasts += stats["total_broadcasts"].(int64)
	}

	// Calculate average uptime
	avgUptime := 0.0
	if len(clientStats) > 0 {
		totalUptime := 0.0
		for _, stats := range clientStats {
			totalUptime += stats["uptime_seconds"].(float64)
		}
		avgUptime = totalUptime / float64(len(clientStats))
	}

	// Copy clients by station map
	clientsByStation := make(map[string]int)
	for k, v := range h.stats.ClientsByStation {
		clientsByStation[k] = v
	}

	return map[string]interface{}{
		"connections": map[string]interface{}{
			"total":      len(h.clients),
			"by_station": clientsByStation,
			"avg_uptime": avgUptime,
		},
		"messages": map[string]interface{}{
			"handled":         h.stats.MessagesHandled,
			"total_sent":      totalMessages,
			"total_broadcast": totalBroadcasts,
		},
		"rate_limits": map[string]interface{}{
			"max_messages_per_minute":   maxMessagesPerMinute,
			"max_broadcasts_per_minute": maxBroadcastsPerMinute,
		},
		"clients": clientStats,
	}
}

// RegisterClient sends a client to the register channel
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient sends a client to the unregister channel
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// DeliverOfflineMessages sends queued messages when user connects
func (h *Hub) DeliverOfflineMessages(client *Client) {
	ctx := context.Background()

	// Create Redis service wrapper
	redisService := cache.NewRedisServiceFromClient(h.redis)

	// Dequeue messages (max 100 at once)
	messageIDStrs, err := redisService.DequeueOfflineMessages(ctx, client.UserID, 100)
	if err != nil {
		fmt.Printf("⚠️  Failed to dequeue messages: %v\n", err)
		return
	}

	if len(messageIDStrs) == 0 {
		return
	}

	fmt.Printf("📬 Delivering %d offline messages to user %s\n", len(messageIDStrs), client.UserID)

	// Parse message IDs
	messageIDs := make([]uuid.UUID, 0, len(messageIDStrs))
	for _, idStr := range messageIDStrs {
		if msgID, err := uuid.Parse(idStr); err == nil {
			messageIDs = append(messageIDs, msgID)
		}
	}

	// Fetch messages from database
	messages, err := h.messageService.GetMessagesByIDs(ctx, messageIDs)
	if err != nil {
		fmt.Printf("⚠️  Failed to fetch messages: %v\n", err)
		return
	}

	// Send each message
	deliveredCount := 0
	for _, msg := range messages {
		serverMsg := &model.ServerMessage{
			ID:        msg.ID.String(),
			Type:      model.MessageTypeMessage,
			From:      msg.SenderID.String(),
			Content:   msg.Text,
			Timestamp: msg.Timestamp,
			Metadata: map[string]interface{}{
				"message_id":      msg.ID.String(),
				"delivery_status": msg.DeliveryStatus,
				"is_offline":      true, // Flag for client
			},
		}

		select {
		case client.send <- serverMsg:
			// Mark as delivered
			h.messageService.UpdateDeliveryStatus(ctx, msg.ID, model.MessageDelivered)
			deliveredCount++
		default:
			// Re-enqueue if send fails
			redisService.EnqueueOfflineMessage(ctx, client.UserID, msg.ID.String())
		}
	}

	fmt.Printf("✅ Delivered %d/%d offline messages to user %s\n",
		deliveredCount, len(messages), client.UserID)
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
