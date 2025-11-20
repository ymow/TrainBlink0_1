package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ymow/messenger_protocol_research/internal/trip"
)

// TripHandler handles trip-related HTTP requests
type TripHandler struct {
	service *trip.Service
}

// NewTripHandler creates a new trip handler
func NewTripHandler(service *trip.Service) *TripHandler {
	return &TripHandler{
		service: service,
	}
}

// StartTrip handles POST /api/v1/trips/start
func (h *TripHandler) StartTrip(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Parse request
	var req trip.StartTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Set user ID from auth context
	req.UserID = userID

	// Start trip
	tripData, err := h.service.StartTrip(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "START_TRIP_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   tripData,
	})
}

// EndTrip handles POST /api/v1/trips/end
func (h *TripHandler) EndTrip(c *gin.Context) {
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
	var req trip.EndTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Set user ID
	req.UserID = userID

	// End trip
	tripData, err := h.service.EndTrip(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "END_TRIP_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   tripData,
	})
}

// GetActiveTrip handles GET /api/v1/trips/active
func (h *TripHandler) GetActiveTrip(c *gin.Context) {
	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Get active trip
	tripData, err := h.service.GetActiveTrip(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GET_TRIP_FAILED",
		})
		return
	}

	if tripData == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No active trip found",
			"code":  "NO_ACTIVE_TRIP",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   tripData,
	})
}

// GetTripByID handles GET /api/v1/trips/:id
func (h *TripHandler) GetTripByID(c *gin.Context) {
	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Parse trip ID
	tripIDStr := c.Param("id")
	tripID, err := uuid.Parse(tripIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid trip ID",
			"code":  "INVALID_TRIP_ID",
		})
		return
	}

	// Get trip
	tripData, err := h.service.GetTripByID(c.Request.Context(), tripID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
			"code":  "TRIP_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   tripData,
	})
}

// GetUserTrips handles GET /api/v1/trips
func (h *TripHandler) GetUserTrips(c *gin.Context) {
	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Parse pagination parameters
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 {
				limit = 100 // Max limit
			}
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get trips
	trips, total, err := h.service.GetUserTrips(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GET_TRIPS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"trips":  trips,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// CancelTrip handles DELETE /api/v1/trips/:id
func (h *TripHandler) CancelTrip(c *gin.Context) {
	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Parse trip ID
	tripIDStr := c.Param("id")
	tripID, err := uuid.Parse(tripIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid trip ID",
			"code":  "INVALID_TRIP_ID",
		})
		return
	}

	// Cancel trip
	if err := h.service.CancelTrip(c.Request.Context(), tripID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "CANCEL_TRIP_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Trip cancelled successfully",
	})
}

// UpdateDiscoveryEnabled handles PATCH /api/v1/trips/:id/discovery
func (h *TripHandler) UpdateDiscoveryEnabled(c *gin.Context) {
	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Parse trip ID
	tripIDStr := c.Param("id")
	tripID, err := uuid.Parse(tripIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid trip ID",
			"code":  "INVALID_TRIP_ID",
		})
		return
	}

	// Parse request
	var req struct {
		Enabled bool `json:"enabled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Update discovery enabled
	if err := h.service.UpdateDiscoveryEnabled(c.Request.Context(), tripID, userID, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "UPDATE_DISCOVERY_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Discovery setting updated",
	})
}

// GetActiveTripStats handles GET /api/v1/trips/stats
func (h *TripHandler) GetActiveTripStats(c *gin.Context) {
	// Get stats (admin-only endpoint in production)
	stats, err := h.service.GetActiveTripStats(c.Request.Context())
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

// getUserIDFromContext extracts user ID from Gin context
// In production, this would be set by Firebase auth middleware
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	// Try to get from context (set by auth middleware)
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uuid.UUID); ok {
			return uid, nil
		}
		if uid, ok := userID.(string); ok {
			return uuid.Parse(uid)
		}
	}

	// For development: try X-User-ID header
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr != "" {
		return uuid.Parse(userIDStr)
	}

	return uuid.Nil, http.ErrNoCookie
}
