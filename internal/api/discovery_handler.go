package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ymow/messenger_protocol_research/internal/discovery"
)

// DiscoveryHandler handles discovery-related HTTP requests
type DiscoveryHandler struct {
	service           *discovery.Service
	permissionService interface {
		RecordDiscovery(ctx context.Context, discovererID, discoveredID uuid.UUID, tripID *uuid.UUID, rssi *int, distanceEstimate *string) error
	}
}

// NewDiscoveryHandler creates a new discovery handler
func NewDiscoveryHandler(service *discovery.Service, permService interface {
	RecordDiscovery(ctx context.Context, discovererID, discoveredID uuid.UUID, tripID *uuid.UUID, rssi *int, distanceEstimate *string) error
}) *DiscoveryHandler {
	return &DiscoveryHandler{
		service:           service,
		permissionService: permService,
	}
}

// LogDiscovery handles POST /api/v1/discoveries/log
func (h *DiscoveryHandler) LogDiscovery(c *gin.Context) {
	// Parse request
	var req discovery.LogDiscoveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Log discovery
	if err := h.service.LogDiscovery(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "LOG_DISCOVERY_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Discovery logged successfully",
	})
}

// GetDiscoveriesByRoute handles GET /api/v1/discoveries/route/:route
func (h *DiscoveryHandler) GetDiscoveriesByRoute(c *gin.Context) {
	// Get route from URL
	route := c.Param("route")
	if route == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Route parameter is required",
			"code":  "MISSING_ROUTE",
		})
		return
	}

	// Parse pagination parameters
	limit := 50
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 200 {
				limit = 200 // Max limit
			}
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get discoveries
	discoveries, total, err := h.service.GetDiscoveriesByRoute(c.Request.Context(), route, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GET_DISCOVERIES_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"discoveries": discoveries,
			"total":       total,
			"limit":       limit,
			"offset":      offset,
		},
	})
}

// GetDiscoveryStats handles GET /api/v1/discoveries/stats
func (h *DiscoveryHandler) GetDiscoveryStats(c *gin.Context) {
	// Get discovery statistics
	stats, err := h.service.GetDiscoveryStats(c.Request.Context())
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

// GetRouteStats handles GET /api/v1/discoveries/route/:route/stats
func (h *DiscoveryHandler) GetRouteStats(c *gin.Context) {
	// Get route from URL
	route := c.Param("route")
	if route == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Route parameter is required",
			"code":  "MISSING_ROUTE",
		})
		return
	}

	// Get route statistics
	stats, err := h.service.GetRouteStats(c.Request.Context(), route)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GET_ROUTE_STATS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// GetDiscoveriesByDateRange handles GET /api/v1/discoveries/range
func (h *DiscoveryHandler) GetDiscoveriesByDateRange(c *gin.Context) {
	// Parse date parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start_date and end_date query parameters are required",
			"code":  "MISSING_DATE_PARAMS",
		})
		return
	}

	// Parse dates (expecting RFC3339 format)
	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start_date format. Use RFC3339 (e.g., 2024-01-01T00:00:00Z)",
			"code":  "INVALID_DATE_FORMAT",
		})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end_date format. Use RFC3339 (e.g., 2024-01-31T23:59:59Z)",
			"code":  "INVALID_DATE_FORMAT",
		})
		return
	}

	// Get discoveries
	discoveries, err := h.service.GetDiscoveriesByDateRange(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GET_DISCOVERIES_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"discoveries": discoveries,
			"total":       len(discoveries),
			"start_date":  startDate,
			"end_date":    endDate,
		},
	})
}

// GetPopularRoutes handles GET /api/v1/discoveries/popular-routes
func (h *DiscoveryHandler) GetPopularRoutes(c *gin.Context) {
	// Parse limit parameter
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 50 {
				limit = 50 // Max limit
			}
		}
	}

	// Get popular routes
	routes, err := h.service.GetPopularRoutes(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GET_POPULAR_ROUTES_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"routes": routes,
			"total":  len(routes),
		},
	})
}

// GetDiscoveryTrends handles GET /api/v1/discoveries/trends
func (h *DiscoveryHandler) GetDiscoveryTrends(c *gin.Context) {
	// Parse days parameter
	days := 7 // Default to 7 days
	if daysStr := c.Query("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
			if days > 90 {
				days = 90 // Max 90 days
			}
		}
	}

	// Get discovery trends
	trends, err := h.service.GetDiscoveryTrends(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GET_TRENDS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"trends": trends,
			"days":   days,
		},
	})
}

// RecordDiscovery handles POST /api/v1/discoveries/record
// Mobile apps call this when BLE discovery occurs to grant messaging permission
func (h *DiscoveryHandler) RecordDiscovery(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	var req struct {
		DiscoveredUserID string  `json:"discovered_user_id" binding:"required"`
		TripID           *string `json:"trip_id"`
		RSSI             *int    `json:"rssi"`
		DistanceEstimate *string `json:"distance_estimate"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "INVALID_REQUEST",
		})
		return
	}

	discoveredID, err := uuid.Parse(req.DiscoveredUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid discovered user ID",
			"code":  "INVALID_USER_ID",
		})
		return
	}

	var tripID *uuid.UUID
	if req.TripID != nil {
		parsed, err := uuid.Parse(*req.TripID)
		if err == nil {
			tripID = &parsed
		}
	}

	if h.permissionService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Permission service not available",
			"code":  "SERVICE_UNAVAILABLE",
		})
		return
	}

	err = h.permissionService.RecordDiscovery(
		c.Request.Context(),
		userID,
		discoveredID,
		tripID,
		req.RSSI,
		req.DistanceEstimate,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "RECORD_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Discovery recorded - 10-minute messaging permission granted",
	})
}
