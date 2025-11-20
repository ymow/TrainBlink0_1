package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ymow/messenger_protocol_research/internal/geofence"
	"github.com/ymow/messenger_protocol_research/internal/model"
)

// GeofenceHandler handles geofencing HTTP requests
type GeofenceHandler struct {
	service *geofence.Service
}

// NewGeofenceHandler creates a new geofence handler
func NewGeofenceHandler(service *geofence.Service) *GeofenceHandler {
	return &GeofenceHandler{
		service: service,
	}
}

// EnterStation handles POST /api/v1/geofence/enter
func (h *GeofenceHandler) EnterStation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req model.EnterStationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
			"code":  "INVALID_JSON",
		})
		return
	}

	// Get user ID from header (in production, this would come from JWT)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous-" + uuid.New().String()[:8]
	}

	deviceID := r.Header.Get("X-Device-ID")
	if deviceID == "" {
		deviceID = "device-" + uuid.New().String()[:8]
	}

	// Set timestamp if not provided
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	// Handle enter station
	resp, err := h.service.HandleEnterStation(r.Context(), userID, deviceID, &req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
			"code":  "ENTER_STATION_FAILED",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   resp,
	})
}

// ExitStation handles POST /api/v1/geofence/exit
func (h *GeofenceHandler) ExitStation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req model.ExitStationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
			"code":  "INVALID_JSON",
		})
		return
	}

	// Get user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error": "Missing user ID",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Parse session ID
	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid session ID",
			"code":  "INVALID_SESSION_ID",
		})
		return
	}

	// Set timestamp if not provided
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	// Handle exit station
	resp, err := h.service.HandleExitStation(r.Context(), sessionID, userID, req.StationID, &req.Activity)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
			"code":  "EXIT_STATION_FAILED",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   resp,
	})
}

// GetStats handles GET /api/v1/geofence/stats
func (h *GeofenceHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.service.GetStats()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   stats,
	})
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
