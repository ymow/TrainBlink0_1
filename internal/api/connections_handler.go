package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ConnectionsHandler handles WebSocket connection tracking
type ConnectionsHandler struct {
	// In-memory connection tracking for now
	activeConnections map[string]time.Time
	clientIDs         []string
}

// NewConnectionsHandler creates a new connections handler
func NewConnectionsHandler() *ConnectionsHandler {
	return &ConnectionsHandler{
		activeConnections: make(map[string]time.Time),
		clientIDs:         []string{},
	}
}

// ConnectionsResponse represents the response format for connections endpoint
type ConnectionsResponse struct {
	Count     int      `json:"count"`
	ClientIDs []string `json:"client_ids"`
	Timestamp string   `json:"timestamp"`
}

// GetConnections returns the current active connection count
func (h *ConnectionsHandler) GetConnections(c *gin.Context) {
	// Clean up stale connections (older than 30 seconds)
	h.cleanupStaleConnections()

	response := ConnectionsResponse{
		Count:     len(h.activeConnections),
		ClientIDs: h.clientIDs,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// AddConnection adds a new connection
func (h *ConnectionsHandler) AddConnection(clientID string) {
	h.activeConnections[clientID] = time.Now()
	h.updateClientIDs()
}

// RemoveConnection removes a connection
func (h *ConnectionsHandler) RemoveConnection(clientID string) {
	delete(h.activeConnections, clientID)
	h.updateClientIDs()
}

// cleanupStaleConnections removes connections older than 30 seconds
func (h *ConnectionsHandler) cleanupStaleConnections() {
	now := time.Now()
	for clientID, lastSeen := range h.activeConnections {
		if now.Sub(lastSeen) > 30*time.Second {
			delete(h.activeConnections, clientID)
		}
	}
	h.updateClientIDs()
}

// updateClientIDs updates the slice of client IDs
func (h *ConnectionsHandler) updateClientIDs() {
	h.clientIDs = make([]string, 0, len(h.activeConnections))
	for clientID := range h.activeConnections {
		h.clientIDs = append(h.clientIDs, clientID)
	}
}