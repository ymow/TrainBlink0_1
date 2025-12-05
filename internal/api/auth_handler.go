package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ymow/messenger_protocol_research/internal/auth"
	"github.com/ymow/messenger_protocol_research/internal/user"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	db         *gorm.DB
	jwtService *auth.JWTService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtService: jwtService,
	}
}

// AnonymousLoginRequest represents anonymous login request
type AnonymousLoginRequest struct {
	DeviceID   string `json:"device_id" binding:"required"`
	DeviceType string `json:"device_type"` // "ios", "android", "web", etc.
}

// AnonymousLoginResponse represents anonymous login response
type AnonymousLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	UserID       string `json:"user_id"`
}

// AnonymousLogin handles POST /api/v1/auth/anonymous
// Allows mobile apps to get JWT tokens without account creation
func (h *AuthHandler) AnonymousLogin(c *gin.Context) {
	var req AnonymousLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Generate anonymous Firebase UID using device_id
	// This allows us to reuse the existing User model without schema changes
	firebaseUID := "anonymous:" + req.DeviceID

	// Find or create user
	var existingUser user.User
	result := h.db.Where("firebase_uid = ?", firebaseUID).First(&existingUser)

	var userID uuid.UUID

	if result.Error == gorm.ErrRecordNotFound {
		// Create new anonymous user
		newUser := user.User{
			ID:          uuid.New(),
			FirebaseUID: firebaseUID,
			IsActive:    true,
		}

		// Set device type as display name if provided
		if req.DeviceType != "" {
			deviceType := "Anonymous (" + req.DeviceType + ")"
			newUser.DisplayName = &deviceType
		}

		if err := h.db.Create(&newUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create user",
				"details": err.Error(),
			})
			return
		}

		userID = newUser.ID
	} else if result.Error != nil {
		// Database error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
			"details": result.Error.Error(),
		})
		return
	} else {
		// User exists
		userID = existingUser.ID
	}

	// Generate JWT token pair
	tokenPair, err := h.jwtService.GenerateUserTokenPair(userID, req.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate tokens",
			"details": err.Error(),
		})
		return
	}

	// Return tokens
	c.JSON(http.StatusOK, AnonymousLoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		UserID:       userID.String(),
	})
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse represents refresh token response
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// RefreshToken handles POST /api/v1/auth/refresh
// Issues a new access token using a valid refresh token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Validate refresh token
	claims, err := h.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired refresh token",
		})
		return
	}

	// Verify it's a refresh token
	if claims.TokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token is not a refresh token",
		})
		return
	}

	// Parse user_id
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user_id in token",
		})
		return
	}

	// Generate new access token
	newTokenPair, err := h.jwtService.GenerateUserTokenPair(userID, claims.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate new token",
			"details": err.Error(),
		})
		return
	}

	// Return new access token
	c.JSON(http.StatusOK, RefreshTokenResponse{
		AccessToken: newTokenPair.AccessToken,
		ExpiresIn:   newTokenPair.ExpiresIn,
	})
}
