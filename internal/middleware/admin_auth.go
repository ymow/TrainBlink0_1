package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ymow/messenger_protocol_research/internal/auth"
	"github.com/ymow/messenger_protocol_research/internal/cache"
)

// ContextKey is a custom type for context keys
type ContextKey string

const (
	// AdminClaimsKey is the context key for admin JWT claims
	AdminClaimsKey ContextKey = "admin_claims"
)

// AdminAuthMiddleware validates JWT tokens for admin routes
func AdminAuthMiddleware(jwtService *auth.JWTService, redisService *cache.RedisService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				respondError(w, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization token is required")
				return
			}

			// Verify JWT token
			claims, err := jwtService.VerifyToken(token)
			if err != nil {
				respondError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token")
				return
			}

			// Check if token is revoked (optional - requires Redis)
			if redisService != nil {
				tokenHash := auth.HashToken(token)
				revoked, err := redisService.IsAccessTokenRevoked(r.Context(), tokenHash)
				if err == nil && revoked {
					respondError(w, http.StatusUnauthorized, "TOKEN_REVOKED", "Token has been revoked")
					return
				}
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), AdminClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission middleware checks if admin has required permission
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(AdminClaimsKey).(*auth.AdminClaims)
			if !ok {
				respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}

			if !claims.HasPermission(permission) {
				respondError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole middleware checks if admin has required role
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(AdminClaimsKey).(*auth.AdminClaims)
			if !ok {
				respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}

			hasRole := false
			for _, role := range roles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				respondError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient role")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetAdminClaims retrieves admin claims from request context
func GetAdminClaims(r *http.Request) (*auth.AdminClaims, bool) {
	claims, ok := r.Context().Value(AdminClaimsKey).(*auth.AdminClaims)
	return claims, ok
}

// respondError sends JSON error response
func respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "error",
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
