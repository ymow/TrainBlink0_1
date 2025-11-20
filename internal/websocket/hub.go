package websocket

import (
	"fmt"
	"sync"
	"time"

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
func NewHub() *Hub {
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

	// Update stats
	h.stats.mu.Lock()
	h.stats.MessagesHandled++
	h.stats.mu.Unlock()

	select {
	case client.send <- directMsg.Message:
	default:
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

// RegisterClient sends a client to the register channel
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient sends a client to the unregister channel
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
