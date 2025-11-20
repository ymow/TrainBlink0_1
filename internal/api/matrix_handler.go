package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ymow/messenger_protocol_research/internal/matrix"
	"github.com/ymow/messenger_protocol_research/internal/trip"
)

// MatrixHandler handles Matrix-related HTTP requests
type MatrixHandler struct {
	ephemeralRoomMgr *matrix.EphemeralRoomManager
	tripService      *trip.Service
}

// NewMatrixHandler creates a new matrix handler
func NewMatrixHandler(ephemeralRoomMgr *matrix.EphemeralRoomManager, tripService *trip.Service) *MatrixHandler {
	return &MatrixHandler{
		ephemeralRoomMgr: ephemeralRoomMgr,
		tripService:      tripService,
	}
}

// CreateEphemeralDM handles POST /api/v1/matrix/dm/create
func (h *MatrixHandler) CreateEphemeralDM(c *gin.Context) {
	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Parse request
	var req struct {
		DiscoveredUserBLEID string  `json:"discovered_user_ble_id" binding:"required"`
		MLSGroupID          *string `json:"mls_group_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Get user's active trip
	myTrip, err := h.tripService.GetActiveTrip(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GET_TRIP_FAILED",
		})
		return
	}

	if myTrip == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You must have an active trip to create a DM",
			"code":  "NO_ACTIVE_TRIP",
		})
		return
	}

	// Get discovered user's trip by BLE ID
	discoveredTrip, err := h.tripService.GetTripByBLEID(c.Request.Context(), req.DiscoveredUserBLEID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Discovered user not found or not on an active trip",
			"code":  "DISCOVERED_USER_NOT_FOUND",
		})
		return
	}

	// Create ephemeral DM room
	createReq := &matrix.CreateEphemeralDMRequest{
		Trip1ID:      myTrip.ID,
		Trip2ID:      discoveredTrip.ID,
		AnonymousID1: myTrip.BLEAnonymousID,
		AnonymousID2: discoveredTrip.BLEAnonymousID,
		MLSGroupID:   req.MLSGroupID,
	}

	room, err := h.ephemeralRoomMgr.CreateEphemeralDM(c.Request.Context(), createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "CREATE_DM_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   room,
	})
}

// GetEphemeralDMByID handles GET /api/v1/matrix/dm/:id
func (h *MatrixHandler) GetEphemeralDMByID(c *gin.Context) {
	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Parse room ID
	roomIDStr := c.Param("id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid room ID",
			"code":  "INVALID_ROOM_ID",
		})
		return
	}

	// Get room
	room, err := h.ephemeralRoomMgr.GetEphemeralRoomByID(c.Request.Context(), roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "ROOM_NOT_FOUND",
		})
		return
	}

	// Verify user is part of this room
	myTrip, err := h.tripService.GetActiveTrip(c.Request.Context(), userID)
	if err != nil || myTrip == nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not part of this room",
			"code":  "FORBIDDEN",
		})
		return
	}

	if room.Trip1ID != myTrip.ID && room.Trip2ID != myTrip.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not part of this room",
			"code":  "FORBIDDEN",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   room,
	})
}

// GetMyActiveEphemeralDMs handles GET /api/v1/matrix/dm/active
func (h *MatrixHandler) GetMyActiveEphemeralDMs(c *gin.Context) {
	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Get user's active trip
	myTrip, err := h.tripService.GetActiveTrip(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GET_TRIP_FAILED",
		})
		return
	}

	if myTrip == nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"rooms": []interface{}{},
				"total": 0,
			},
		})
		return
	}

	// Get active ephemeral rooms for user's trip
	rooms, err := h.ephemeralRoomMgr.GetActiveEphemeralRoomsForTrip(c.Request.Context(), myTrip.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GET_ROOMS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"rooms": rooms,
			"total": len(rooms),
		},
	})
}

// IncrementMessageCount handles POST /api/v1/matrix/dm/:roomId/message
func (h *MatrixHandler) IncrementMessageCount(c *gin.Context) {
	// Parse room ID from URL
	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Room ID is required",
			"code":  "MISSING_ROOM_ID",
		})
		return
	}

	// Increment message count
	if err := h.ephemeralRoomMgr.IncrementMessageCount(c.Request.Context(), roomID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "INCREMENT_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Message count incremented",
	})
}

// GetEphemeralRoomStats handles GET /api/v1/matrix/dm/stats
func (h *MatrixHandler) GetEphemeralRoomStats(c *gin.Context) {
	// Get ephemeral room statistics
	stats, err := h.ephemeralRoomMgr.GetEphemeralRoomStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GET_STATS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// ExtendRoomLifetime handles PATCH /api/v1/matrix/dm/:id/extend
func (h *MatrixHandler) ExtendRoomLifetime(c *gin.Context) {
	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Parse room ID
	roomIDStr := c.Param("id")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid room ID",
			"code":  "INVALID_ROOM_ID",
		})
		return
	}

	// Parse request
	var req struct {
		ExtensionHours int `json:"extension_hours" binding:"required,min=1,max=24"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Get room
	room, err := h.ephemeralRoomMgr.GetEphemeralRoomByID(c.Request.Context(), roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "ROOM_NOT_FOUND",
		})
		return
	}

	// Verify user is part of this room
	myTrip, err := h.tripService.GetActiveTrip(c.Request.Context(), userID)
	if err != nil || myTrip == nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not part of this room",
			"code":  "FORBIDDEN",
		})
		return
	}

	if room.Trip1ID != myTrip.ID && room.Trip2ID != myTrip.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not part of this room",
			"code":  "FORBIDDEN",
		})
		return
	}

	// Extend room lifetime
	extension := time.Duration(req.ExtensionHours) * time.Hour
	if err := h.ephemeralRoomMgr.ExtendRoomLifetime(c.Request.Context(), roomID, extension); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "EXTEND_LIFETIME_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Room lifetime extended by %d hours", req.ExtensionHours),
	})
}
