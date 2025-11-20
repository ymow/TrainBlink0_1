package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ymow/messenger_protocol_research/internal/cleanup"
)

// CleanupHandler handles cleanup-related HTTP requests (admin endpoints)
type CleanupHandler struct {
	cleanupService *cleanup.Service
}

// NewCleanupHandler creates a new cleanup handler
func NewCleanupHandler(cleanupService *cleanup.Service) *CleanupHandler {
	return &CleanupHandler{
		cleanupService: cleanupService,
	}
}

// GetCleanupStats handles GET /api/v1/cleanup/stats
func (h *CleanupHandler) GetCleanupStats(c *gin.Context) {
	stats, err := h.cleanupService.GetCleanupStats(c.Request.Context())
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

// RunManualCleanup handles POST /api/v1/cleanup/run
func (h *CleanupHandler) RunManualCleanup(c *gin.Context) {
	stats, err := h.cleanupService.RunManualCleanup(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "CLEANUP_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Manual cleanup completed",
		"data":    stats,
	})
}

// GetExpiringRoomWarnings handles GET /api/v1/cleanup/expiring-rooms
func (h *CleanupHandler) GetExpiringRoomWarnings(c *gin.Context) {
	// Parse hours parameter (default 1 hour)
	hours := 1
	if hoursStr := c.Query("hours"); hoursStr != "" {
		if parsedHours, err := strconv.Atoi(hoursStr); err == nil && parsedHours > 0 {
			hours = parsedHours
			if hours > 24 {
				hours = 24 // Max 24 hours
			}
		}
	}

	within := time.Duration(hours) * time.Hour
	warnings, err := h.cleanupService.WarnExpiringRooms(c.Request.Context(), within)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "GET_WARNINGS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"warnings": warnings,
			"total":    len(warnings),
			"within_hours": hours,
		},
	})
}
