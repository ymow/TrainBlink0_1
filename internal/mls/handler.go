package mls

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Handler handles MLS HTTP API endpoints
type Handler struct {
	db              *gorm.DB
	redis           *redis.Client
	groupMgr        *GroupManager
	keyPackageSvc   *KeyPackageService
	deliverySvc     *DeliveryService
	config          *MLSConfig
}

// NewHandler creates a new MLS HTTP handler
func NewHandler(
	db *gorm.DB,
	redis *redis.Client,
	groupMgr *GroupManager,
	keyPackageSvc *KeyPackageService,
	deliverySvc *DeliveryService,
	config *MLSConfig,
) *Handler {
	if config == nil {
		config = DefaultMLSConfig()
	}

	return &Handler{
		db:            db,
		redis:         redis,
		groupMgr:      groupMgr,
		keyPackageSvc: keyPackageSvc,
		deliverySvc:   deliverySvc,
		config:        config,
	}
}

// RegisterRoutes registers MLS API routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	mls := router.Group("/mls/v1")
	{
		// Group Management
		mls.POST("/groups", h.CreateGroup)
		mls.GET("/groups/:group_id", h.GetGroup)
		mls.POST("/groups/:group_id/members", h.AddMember)
		mls.DELETE("/groups/:group_id/members/:client_id", h.RemoveMember)
		mls.GET("/groups/:group_id/members", h.GetMembers)

		// KeyPackage Management
		mls.POST("/keypackages", h.UploadKeyPackage)
		mls.POST("/keypackages/claim", h.ClaimKeyPackage)
		mls.GET("/keypackages/stats", h.GetKeyPackageStats)

		// Message Delivery
		mls.POST("/groups/:group_id/messages", h.SendMessage)
		mls.GET("/groups/:group_id/messages", h.GetPendingMessages)
		mls.POST("/messages/:message_id/ack", h.AcknowledgeMessage)

		// Statistics
		mls.GET("/stats", h.GetStats)
	}

	log.Println("✅ Registered MLS API routes at /mls/v1")
}

// ============================================================
// Group Management Handlers
// ============================================================

// CreateGroup creates a new MLS group
func (h *Handler) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Create group
	group, err := h.groupMgr.CreateGroup(
		c.Request.Context(),
		userID.(string),
		req.RoomID,
		req.StationID,
		req.CipherSuite,
	)
	if err != nil {
		log.Printf("❌ Failed to create group: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	// Add creator as first member
	if req.CreatorKeyPackage != "" {
		keyPackageData, err := base64.StdEncoding.DecodeString(req.CreatorKeyPackage)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key package encoding"})
			return
		}

		// For now, use the keyPackageData as the signaturePublicKey
		// In production, this should extract the actual signature key from the KeyPackage
		if err := h.groupMgr.AddMember(c.Request.Context(), group.GroupID, userID.(string), userID.(string)+"_client", req.CreatorCredentialID, keyPackageData); err != nil {
			log.Printf("⚠️  Failed to add creator as member: %v", err)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"group_id":      group.GroupID,
		"epoch":         group.Epoch,
		"cipher_suite":  group.CipherSuite,
		"member_count":  group.MemberCount,
		"created_at":    group.CreatedAt,
	})
}

// GetGroup retrieves group information
func (h *Handler) GetGroup(c *gin.Context) {
	groupID := c.Param("group_id")

	group, err := h.groupMgr.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"group_id":          group.GroupID,
		"station_id":        group.StationID,
		"epoch":             group.Epoch,
		"cipher_suite":      group.CipherSuite,
		"protocol_version":  group.ProtocolVersion,
		"member_count":      group.MemberCount,
		"is_active":         group.IsActive,
		"created_at":        group.CreatedAt,
		"updated_at":        group.UpdatedAt,
	})
}

// AddMember adds a member to a group
func (h *Handler) AddMember(c *gin.Context) {
	groupID := c.Param("group_id")

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Decode key package
	keyPackageData, err := base64.StdEncoding.DecodeString(req.KeyPackage)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key package encoding"})
		return
	}

	// Add member
	// Note: In production, extract the actual signature public key from the KeyPackage
	// For now, use the keyPackageData as the signaturePublicKey parameter
	if err := h.groupMgr.AddMember(c.Request.Context(), groupID, req.UserID, req.ClientID, req.CredentialID, keyPackageData); err != nil {
		log.Printf("❌ Failed to add member: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	// Get updated group
	group, _ := h.groupMgr.GetGroup(c.Request.Context(), groupID)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Member added successfully",
		"group_id":     groupID,
		"new_epoch":    group.Epoch,
		"member_count": group.MemberCount,
	})
}

// RemoveMember removes a member from a group
func (h *Handler) RemoveMember(c *gin.Context) {
	groupID := c.Param("group_id")
	clientID := c.Param("client_id")

	if err := h.groupMgr.RemoveMember(c.Request.Context(), groupID, clientID); err != nil {
		log.Printf("❌ Failed to remove member: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	// Get updated group
	group, _ := h.groupMgr.GetGroup(c.Request.Context(), groupID)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Member removed successfully",
		"group_id":     groupID,
		"new_epoch":    group.Epoch,
		"member_count": group.MemberCount,
	})
}

// GetMembers retrieves all members of a group
func (h *Handler) GetMembers(c *gin.Context) {
	groupID := c.Param("group_id")

	members, err := h.groupMgr.GetActiveMembers(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get members"})
		return
	}

	// Convert to response format
	memberList := make([]gin.H, len(members))
	for i, member := range members {
		memberList[i] = gin.H{
			"client_id":       member.ClientID,
			"user_id":         member.UserID,
			"credential_id":   member.CredentialID,
			"joined_at_epoch": member.JoinedAtEpoch,
			"is_active":       member.IsActive,
			"joined_at":       member.JoinedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"group_id":     groupID,
		"member_count": len(members),
		"members":      memberList,
	})
}

// ============================================================
// KeyPackage Management Handlers
// ============================================================

// UploadKeyPackage uploads a KeyPackage
func (h *Handler) UploadKeyPackage(c *gin.Context) {
	var req UploadKeyPackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Decode key package
	keyPackageData, err := base64.StdEncoding.DecodeString(req.KeyPackage)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key package encoding"})
		return
	}

	// Get client IP
	clientIP := c.ClientIP()

	// Upload KeyPackage
	kp, err := h.keyPackageSvc.UploadKeyPackage(
		c.Request.Context(),
		userID.(string),
		req.ClientID,
		req.CredentialID,
		keyPackageData,
		req.CipherSuite,
		clientIP,
	)
	if err != nil {
		log.Printf("❌ Failed to upload KeyPackage: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"keypackage_id": kp.KeyPackageID,
		"client_id":     kp.ClientID,
		"cipher_suite":  kp.CipherSuite,
		"created_at":    kp.CreatedAt,
	})
}

// ClaimKeyPackage claims a KeyPackage for adding a user to a group
func (h *Handler) ClaimKeyPackage(c *gin.Context) {
	var req ClaimKeyPackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	kp, err := h.keyPackageSvc.ClaimKeyPackage(
		c.Request.Context(),
		req.UserID,
		req.GroupID,
		req.CipherSuite,
	)
	if err != nil {
		log.Printf("❌ Failed to claim KeyPackage: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "No available KeyPackage"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keypackage_id": kp.KeyPackageID,
		"client_id":     kp.ClientID,
		"keypackage":    base64.StdEncoding.EncodeToString(kp.KeyPackageData),
		"cipher_suite":  kp.CipherSuite,
	})
}

// GetKeyPackageStats retrieves KeyPackage statistics
func (h *Handler) GetKeyPackageStats(c *gin.Context) {
	stats, err := h.keyPackageSvc.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ============================================================
// Message Delivery Handlers
// ============================================================

// SendMessage sends an MLS message to a group
func (h *Handler) SendMessage(c *gin.Context) {
	groupID := c.Param("group_id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Decode ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(req.Ciphertext)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ciphertext encoding"})
		return
	}

	// Decode authenticated data
	authenticatedData, err := base64.StdEncoding.DecodeString(req.AuthenticatedData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authenticated data encoding"})
		return
	}

	// Send message
	message, err := h.deliverySvc.SendMessage(
		c.Request.Context(),
		req.SenderClientID,
		groupID,
		req.MessageType,
		ciphertext,
		authenticatedData,
	)
	if err != nil {
		log.Printf("❌ Failed to send message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message_id":      message.MessageID,
		"group_id":        message.GroupID,
		"epoch":           message.Epoch,
		"sequence_number": message.SequenceNumber,
		"sent_at":         message.SentAt,
		"recipient_count": message.RecipientCount,
	})
}

// GetPendingMessages retrieves pending messages for a client
func (h *Handler) GetPendingMessages(c *gin.Context) {
	groupID := c.Param("group_id")
	clientID := c.Query("client_id")

	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	// Parse query parameters
	var afterSequence int64 = 0
	if seq := c.Query("after_sequence"); seq != "" {
		fmt.Sscanf(seq, "%d", &afterSequence)
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	// Get pending messages
	messages, err := h.deliverySvc.GetPendingMessages(
		c.Request.Context(),
		clientID,
		groupID,
		afterSequence,
		limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	// Convert to response format
	messageList := make([]gin.H, len(messages))
	for i, msg := range messages {
		messageList[i] = gin.H{
			"message_id":         msg.MessageID,
			"group_id":           msg.GroupID,
			"sender_client_id":   msg.SenderClientID,
			"message_type":       msg.MessageType,
			"epoch":              msg.Epoch,
			"sequence_number":    msg.SequenceNumber,
			"ciphertext":         base64.StdEncoding.EncodeToString(msg.Ciphertext),
			"authenticated_data": base64.StdEncoding.EncodeToString(msg.AuthenticatedData),
			"sent_at":            msg.SentAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"group_id": groupID,
		"count":    len(messages),
		"messages": messageList,
	})
}

// AcknowledgeMessage acknowledges message delivery
func (h *Handler) AcknowledgeMessage(c *gin.Context) {
	messageID := c.Param("message_id")

	var req AcknowledgeMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if err := h.deliverySvc.AcknowledgeDelivery(c.Request.Context(), messageID, req.ClientID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acknowledge message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message acknowledged",
	})
}

// ============================================================
// Statistics Handler
// ============================================================

// GetStats retrieves MLS statistics
func (h *Handler) GetStats(c *gin.Context) {
	// Get group stats
	var totalGroups, activeGroups int64
	h.db.Model(&MLSGroup{}).Count(&totalGroups)
	h.db.Model(&MLSGroup{}).Where("is_active = ?", true).Count(&activeGroups)

	var totalMembers int64
	h.db.Model(&MLSGroupMember{}).Where("is_active = ?", true).Count(&totalMembers)

	// Get delivery stats
	deliveryStats, _ := h.deliverySvc.GetStats(c.Request.Context())

	// Get KeyPackage stats
	var totalKeyPackages, availableKeyPackages int64
	h.db.Model(&MLSKeyPackage{}).Count(&totalKeyPackages)
	h.db.Model(&MLSKeyPackage{}).Where("is_consumed = ?", false).Count(&availableKeyPackages)

	stats := gin.H{
		"groups": map[string]interface{}{
			"total":         totalGroups,
			"active":        activeGroups,
			"total_members": totalMembers,
		},
		"delivery": deliveryStats,
		"keypackages": map[string]interface{}{
			"total":     totalKeyPackages,
			"available": availableKeyPackages,
			"consumed":  totalKeyPackages - availableKeyPackages,
		},
	}

	c.JSON(http.StatusOK, stats)
}

// ============================================================
// Request/Response Types (already defined in types.go)
// ============================================================

// These types are already defined in types.go:
// - CreateGroupRequest
// - AddMemberRequest
// - UploadKeyPackageRequest
// - ClaimKeyPackageRequest
// - SendMessageRequest
// - AcknowledgeMessageRequest
