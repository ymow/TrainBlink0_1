package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ymow/messenger_protocol_research/internal/model"
)

// PermissionService defines the interface for permission operations
type PermissionService interface {
	BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID, reason *string) error
	UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error
	GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]model.UserBlock, error)
}

// PermissionHandler handles permission and blocking operations
type PermissionHandler struct {
	service PermissionService
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(service PermissionService) *PermissionHandler {
	return &PermissionHandler{
		service: service,
	}
}

// BlockUser handles POST /api/v1/permissions/block/:user_id
func (h *PermissionHandler) BlockUser(c *gin.Context) {
	blockerID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	blockedIDStr := c.Param("user_id")
	blockedID, err := uuid.Parse(blockedIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
			"code":  "INVALID_USER_ID",
		})
		return
	}

	// Optional reason from request body
	var req struct {
		Reason *string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	err = h.service.BlockUser(c.Request.Context(), blockerID, blockedID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "BLOCK_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "User blocked successfully",
		"blocker_id": blockerID,
		"blocked_id": blockedID,
	})
}

// UnblockUser handles DELETE /api/v1/permissions/block/:user_id
func (h *PermissionHandler) UnblockUser(c *gin.Context) {
	blockerID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	blockedIDStr := c.Param("user_id")
	blockedID, err := uuid.Parse(blockedIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
			"code":  "INVALID_USER_ID",
		})
		return
	}

	err = h.service.UnblockUser(c.Request.Context(), blockerID, blockedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "UNBLOCK_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User unblocked successfully",
	})
}

// GetBlockedUsers handles GET /api/v1/permissions/blocked
func (h *PermissionHandler) GetBlockedUsers(c *gin.Context) {
	blockerID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	blocks, err := h.service.GetBlockedUsers(c.Request.Context(), blockerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   blocks,
		"count":  len(blocks),
	})
}
