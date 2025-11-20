package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ymow/messenger_protocol_research/internal/cache"
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	RequestsPerWindow int64
	WindowDuration    time.Duration
	KeyPrefix         string
}

// DefaultRateLimitConfig returns sensible defaults
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerWindow: 100,
		WindowDuration:    time.Minute,
		KeyPrefix:         "rate_limit",
	}
}

// RateLimitMiddleware implements rate limiting using Redis
func RateLimitMiddleware(redisService *cache.RedisService, config RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get identifier (IP address or user ID)
			identifier := getClientIdentifier(r)
			endpoint := r.URL.Path

			// Construct rate limit key
			key := fmt.Sprintf("%s:%s:%s", config.KeyPrefix, identifier, endpoint)

			// Check rate limit
			count, err := redisService.IncrementRateLimit(
				r.Context(),
				key,
				config.RequestsPerWindow,
				config.WindowDuration,
			)

			if err != nil {
				// If Redis fails, allow the request but log error
				next.ServeHTTP(w, r)
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerWindow))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, config.RequestsPerWindow-count)))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(config.WindowDuration).Unix()))

			// Check if limit exceeded
			if count > config.RequestsPerWindow {
				respondError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
					fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %v",
						config.RequestsPerWindow, config.WindowDuration))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IPRateLimitMiddleware creates a rate limiter based on IP address
func IPRateLimitMiddleware(redisService *cache.RedisService, requestsPerMinute int64) func(http.Handler) http.Handler {
	config := RateLimitConfig{
		RequestsPerWindow: requestsPerMinute,
		WindowDuration:    time.Minute,
		KeyPrefix:         "ip_rate_limit",
	}
	return RateLimitMiddleware(redisService, config)
}

// UserRateLimitMiddleware creates a rate limiter based on authenticated user
func UserRateLimitMiddleware(redisService *cache.RedisService, requestsPerMinute int64) func(http.Handler) http.Handler {
	config := RateLimitConfig{
		RequestsPerWindow: requestsPerMinute,
		WindowDuration:    time.Minute,
		KeyPrefix:         "user_rate_limit",
	}
	return RateLimitMiddleware(redisService, config)
}

// getClientIdentifier extracts client identifier from request
func getClientIdentifier(r *http.Request) string {
	// Try to get admin claims first (if authenticated)
	if claims, ok := GetAdminClaims(r); ok {
		return fmt.Sprintf("admin_%s", claims.AdminID)
	}

	// Fall back to IP address
	return getClientIP(r)
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (if behind proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// max returns the maximum of two int64 values
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
