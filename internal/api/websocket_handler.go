package api

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/ymow/messenger_protocol_research/internal/geofence"
	"github.com/ymow/messenger_protocol_research/internal/websocket"
)

// WebSocketHandler handles WebSocket connection requests
type WebSocketHandler struct {
	hub             *websocket.Hub
	geofenceService *geofence.Service
	messageHandler  *websocket.MessageHandler
}

// NewWebSocketHandler creates a new WebSocketHandler
func NewWebSocketHandler(hub *websocket.Hub, geofenceService *geofence.Service) *WebSocketHandler {
	return &WebSocketHandler{
		hub:             hub,
		geofenceService: geofenceService,
		messageHandler:  websocket.NewMessageHandler(hub),
	}
}

// HandleWebSocket handles WebSocket connection upgrades
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get session ID from query parameter
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Missing session_id parameter", http.StatusBadRequest)
		return
	}

	// Parse session ID
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		http.Error(w, "Invalid session_id format", http.StatusBadRequest)
		return
	}

	// Validate session with geofencing service
	session, err := h.geofenceService.GetSessionByID(sessionUUID)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Get station information
	station, err := h.geofenceService.GetStationByID(r.Context(), session.StationID)
	if err != nil {
		http.Error(w, "Station not found", http.StatusNotFound)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := websocket.UpgradeConnection(w, r)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	// Create new client
	client := websocket.NewClient(
		conn,
		h.hub,
		session.UserID,
		session.DeviceID,
		session.ID.String(),
		station.ID,
		h.messageHandler,
	)

	// Register client with hub via channel
	h.hub.RegisterClient(client)

	// Start client pumps in separate goroutines
	go client.WritePump()
	go client.ReadPump()

	fmt.Printf("WebSocket connection established: UserID=%s, StationID=%s, SessionID=%s\n",
		session.UserID, station.ID, sessionID)
}
