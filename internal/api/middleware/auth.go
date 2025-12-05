package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ymow/messenger_protocol_research/internal/auth"
)

// JWTAuthMiddleware validates JWT tokens on protected routes
func JWTAuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  "UNAUTHORIZED",
				"error": "Missing Authorization header",
			})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  "UNAUTHORIZED",
				"error": "Invalid Authorization header format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  "UNAUTHORIZED",
				"error": "Missing token",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			var errorMessage string
			if err == auth.ErrExpiredToken {
				errorMessage = "Token has expired"
			} else if err == auth.ErrInvalidToken {
				errorMessage = "Invalid token"
			} else {
				errorMessage = "Token validation failed"
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  "UNAUTHORIZED",
				"error": errorMessage,
			})
			c.Abort()
			return
		}

		// Check token type (should be access token, not refresh)
		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  "UNAUTHORIZED",
				"error": "Invalid token type. Use access token for API requests",
			})
			c.Abort()
			return
		}

		// Set user info in context for handlers to use
		// Support both admin and user tokens
		if claims.AdminID != "" {
			// Admin token
			c.Set("admin_id", claims.AdminID)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Set("permissions", claims.Permissions)
		}

		if claims.UserID != "" {
			// User token
			c.Set("user_id", claims.UserID)
			c.Set("device_id", claims.DeviceID)
		}

		// Continue to next handler
		c.Next()
	}
}

// OptionalJWTAuthMiddleware validates JWT tokens if present, but doesn't require them
// Useful for endpoints that have different behavior for authenticated vs unauthenticated users
func OptionalJWTAuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Try to validate token if provided
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtService.ValidateToken(token)

			if err == nil && claims.TokenType == "access" {
				// Valid token, set context
				if claims.AdminID != "" {
					c.Set("admin_id", claims.AdminID)
					c.Set("email", claims.Email)
					c.Set("role", claims.Role)
					c.Set("permissions", claims.Permissions)
				}

				if claims.UserID != "" {
					c.Set("user_id", claims.UserID)
					c.Set("device_id", claims.DeviceID)
				}
			}
			// If token is invalid, continue anyway (optional auth)
		}

		c.Next()
	}
}
