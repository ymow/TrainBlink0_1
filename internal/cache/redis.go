package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ymow/messenger_protocol_research/internal/config"
)

// RedisService provides Redis caching and pub/sub functionality
type RedisService struct {
	client *redis.Client
}

// NewRedisService creates a new Redis service
func NewRedisService(cfg config.RedisConfig) (*RedisService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("✅ Redis connected: %s", cfg.Addr)

	return &RedisService{client: client}, nil
}

// Close closes the Redis connection
func (s *RedisService) Close() error {
	return s.client.Close()
}

// GetClient returns the underlying Redis client
func (s *RedisService) GetClient() *redis.Client {
	return s.client
}

// ============================================================
// Session Management
// ============================================================

// AddUserToStation adds user to station's active users set
func (s *RedisService) AddUserToStation(ctx context.Context, stationID, userID string) error {
	key := fmt.Sprintf("station:%s:active_users", stationID)
	return s.client.SAdd(ctx, key, userID).Err()
}

// RemoveUserFromStation removes user from station's active users set
func (s *RedisService) RemoveUserFromStation(ctx context.Context, stationID, userID string) error {
	key := fmt.Sprintf("station:%s:active_users", stationID)
	return s.client.SRem(ctx, key, userID).Err()
}

// GetStationActiveUsers returns all active users in a station
func (s *RedisService) GetStationActiveUsers(ctx context.Context, stationID string) ([]string, error) {
	key := fmt.Sprintf("station:%s:active_users", stationID)
	return s.client.SMembers(ctx, key).Result()
}

// GetStationUserCount returns count of active users in station
func (s *RedisService) GetStationUserCount(ctx context.Context, stationID string) (int64, error) {
	key := fmt.Sprintf("station:%s:active_users", stationID)
	return s.client.SCard(ctx, key).Result()
}

// SetUserSession stores user's current session info
func (s *RedisService) SetUserSession(ctx context.Context, userID string, session map[string]interface{}) error {
	key := fmt.Sprintf("user:%s:session", userID)

	// Convert to map[string]string for HSET
	strMap := make(map[string]string)
	for k, v := range session {
		strMap[k] = fmt.Sprintf("%v", v)
	}

	if err := s.client.HSet(ctx, key, strMap).Err(); err != nil {
		return err
	}

	// Set TTL
	return s.client.Expire(ctx, key, 4*time.Hour).Err()
}

// GetUserSession retrieves user's current session
func (s *RedisService) GetUserSession(ctx context.Context, userID string) (map[string]string, error) {
	key := fmt.Sprintf("user:%s:session", userID)
	return s.client.HGetAll(ctx, key).Result()
}

// DeleteUserSession removes user's session
func (s *RedisService) DeleteUserSession(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:%s:session", userID)
	return s.client.Del(ctx, key).Err()
}

// ============================================================
// Rate Limiting
// ============================================================

// IncrementRateLimit increments rate limit counter
func (s *RedisService) IncrementRateLimit(
	ctx context.Context,
	key string,
	limit int64,
	window time.Duration,
) (int64, error) {
	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Set expiration only on first increment
	if count == 1 {
		s.client.Expire(ctx, key, window)
	}

	return count, nil
}

// CheckRateLimit checks if rate limit is exceeded
func (s *RedisService) CheckRateLimit(
	ctx context.Context,
	userID, endpoint string,
	limit int64,
	window time.Duration,
) (bool, error) {
	key := fmt.Sprintf("rate_limit:%s:%s", userID, endpoint)

	count, err := s.IncrementRateLimit(ctx, key, limit, window)
	if err != nil {
		return false, err
	}

	return count <= limit, nil
}

// ============================================================
// Data Caching
// ============================================================

// CacheJSON caches any JSON-serializable data
func (s *RedisService) CacheJSON(
	ctx context.Context,
	key string,
	data interface{},
	ttl time.Duration,
) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, key, jsonData, ttl).Err()
}

// GetCachedJSON retrieves and unmarshals cached JSON data
func (s *RedisService) GetCachedJSON(
	ctx context.Context,
	key string,
	dest interface{},
) error {
	jsonData, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, dest)
}

// InvalidateCache deletes cached data
func (s *RedisService) InvalidateCache(ctx context.Context, pattern string) error {
	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.client.Del(ctx, keys...).Err()
	}

	return nil
}

// ============================================================
// Pub/Sub
// ============================================================

// PublishMessage publishes message to channel
func (s *RedisService) PublishMessage(ctx context.Context, channel string, message interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return s.client.Publish(ctx, channel, jsonData).Err()
}

// Subscribe subscribes to channel
func (s *RedisService) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return s.client.Subscribe(ctx, channels...)
}

// ============================================================
// Token Management
// ============================================================

// StoreRefreshToken stores refresh token with admin ID
func (s *RedisService) StoreRefreshToken(ctx context.Context, tokenHash, adminID string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", tokenHash)
	return s.client.Set(ctx, key, adminID, ttl).Err()
}

// GetRefreshToken retrieves admin ID by refresh token hash
func (s *RedisService) GetRefreshToken(ctx context.Context, tokenHash string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", tokenHash)
	return s.client.Get(ctx, key).Result()
}

// DeleteRefreshToken deletes refresh token (logout)
func (s *RedisService) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	key := fmt.Sprintf("refresh_token:%s", tokenHash)
	return s.client.Del(ctx, key).Err()
}

// RevokeAccessToken adds access token to revocation list
func (s *RedisService) RevokeAccessToken(ctx context.Context, tokenHash string, ttl time.Duration) error {
	key := "revoked_tokens"
	// Add to set with TTL
	if err := s.client.SAdd(ctx, key, tokenHash).Err(); err != nil {
		return err
	}
	// Set expiration
	return s.client.Expire(ctx, key, ttl).Err()
}

// IsAccessTokenRevoked checks if access token is revoked
func (s *RedisService) IsAccessTokenRevoked(ctx context.Context, tokenHash string) (bool, error) {
	key := "revoked_tokens"
	return s.client.SIsMember(ctx, key, tokenHash).Result()
}
