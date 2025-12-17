package permission

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ymow/messenger_protocol_research/internal/model"
)

const (
	// Discovery TTL - user can message for 10 minutes after discovery
	DiscoveryTTL = 10 * time.Minute

	// Cache TTLs
	PermissionCacheTTL = 2 * time.Minute // Cache permission checks
	BlockCacheTTL      = 5 * time.Minute // Cache block checks
)

// Service handles messaging permission validation
type Service struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewService creates a new permission service
func NewService(db *gorm.DB, redis *redis.Client) *Service {
	return &Service{
		db:    db,
		redis: redis,
	}
}

// CanSendMessage checks if sender can message receiver
// Returns (canSend bool, reason string, error)
func (s *Service) CanSendMessage(
	ctx context.Context,
	senderID, receiverID uuid.UUID,
) (bool, string, error) {
	// Edge case: prevent self-messaging
	if senderID == receiverID {
		return false, "SELF_MESSAGE_NOT_ALLOWED", nil
	}

	// Check 1: Is sender blocked by receiver?
	blocked, err := s.IsUserBlocked(ctx, receiverID, senderID)
	if err != nil {
		return false, "BLOCK_CHECK_FAILED", err
	}
	if blocked {
		return false, "SENDER_BLOCKED_BY_RECEIVER", nil
	}

	// Check 2: Has sender discovered receiver via BLE within last 10 minutes?
	discovered, err := s.HasValidDiscovery(ctx, senderID, receiverID)
	if err != nil {
		return false, "DISCOVERY_CHECK_FAILED", err
	}
	if !discovered {
		return false, "NO_VALID_DISCOVERY", nil
	}

	// All checks passed
	return true, "PERMISSION_GRANTED", nil
}

// HasValidDiscovery checks if A has discovered B within the last 10 minutes
func (s *Service) HasValidDiscovery(
	ctx context.Context,
	discovererID, discoveredID uuid.UUID,
) (bool, error) {
	// Try Redis cache first
	if s.redis != nil {
		cacheKey := fmt.Sprintf("discovery:%s:%s", discovererID, discoveredID)
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil && cached == "1" {
			return true, nil
		}
		if err == nil && cached == "0" {
			return false, nil
		}
	}

	// Query database for valid discovery
	var count int64
	err := s.db.WithContext(ctx).
		Model(&model.BLEDiscovery{}).
		Where("discoverer_id = ? AND discovered_id = ? AND expires_at > NOW()",
			discovererID, discoveredID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check discovery: %w", err)
	}

	hasDiscovery := count > 0

	// Cache result
	if s.redis != nil {
		cacheKey := fmt.Sprintf("discovery:%s:%s", discovererID, discoveredID)
		value := "0"
		if hasDiscovery {
			value = "1"
		}
		s.redis.Set(ctx, cacheKey, value, PermissionCacheTTL)
	}

	return hasDiscovery, nil
}

// RecordDiscovery records a BLE discovery event
func (s *Service) RecordDiscovery(
	ctx context.Context,
	discovererID, discoveredID uuid.UUID,
	tripID *uuid.UUID,
	rssi *int,
	distanceEstimate *string,
) error {
	// Prevent self-discovery
	if discovererID == discoveredID {
		return fmt.Errorf("cannot discover yourself")
	}

	now := time.Now()
	expiresAt := now.Add(DiscoveryTTL)

	discovery := &model.BLEDiscovery{
		DiscovererID:     discovererID,
		DiscoveredID:     discoveredID,
		TripID:           tripID,
		RSSI:             rssi,
		DistanceEstimate: distanceEstimate,
		DiscoveredAt:     now,
		ExpiresAt:        expiresAt,
	}

	// Use ON CONFLICT to update if exists
	err := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "discoverer_id"},
				{Name: "discovered_id"},
				{Name: "trip_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"rssi",
				"distance_estimate",
				"discovered_at",
				"expires_at",
			}),
		}).
		Create(discovery).Error

	if err != nil {
		return fmt.Errorf("failed to record discovery: %w", err)
	}

	// Invalidate cache
	if s.redis != nil {
		cacheKey := fmt.Sprintf("discovery:%s:%s", discovererID, discoveredID)
		s.redis.Del(ctx, cacheKey)
	}

	return nil
}

// IsUserBlocked checks if blockedID is blocked by blockerID
func (s *Service) IsUserBlocked(
	ctx context.Context,
	blockerID, blockedID uuid.UUID,
) (bool, error) {
	// Try Redis cache first
	if s.redis != nil {
		cacheKey := fmt.Sprintf("block:%s:%s", blockerID, blockedID)
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil && cached == "1" {
			return true, nil
		}
		if err == nil && cached == "0" {
			return false, nil
		}
	}

	// Query database
	var count int64
	err := s.db.WithContext(ctx).
		Model(&model.UserBlock{}).
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check block: %w", err)
	}

	isBlocked := count > 0

	// Cache result
	if s.redis != nil {
		cacheKey := fmt.Sprintf("block:%s:%s", blockerID, blockedID)
		value := "0"
		if isBlocked {
			value = "1"
		}
		s.redis.Set(ctx, cacheKey, value, BlockCacheTTL)
	}

	return isBlocked, nil
}

// BlockUser blocks a user
func (s *Service) BlockUser(
	ctx context.Context,
	blockerID, blockedID uuid.UUID,
	reason *string,
) error {
	// Prevent self-blocking
	if blockerID == blockedID {
		return fmt.Errorf("cannot block yourself")
	}

	block := &model.UserBlock{
		BlockerID: blockerID,
		BlockedID: blockedID,
		Reason:    reason,
	}

	// Use ON CONFLICT to ignore if already exists
	err := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "blocker_id"}, {Name: "blocked_id"}},
			DoNothing: true,
		}).
		Create(block).Error

	if err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}

	// Invalidate cache
	if s.redis != nil {
		cacheKey := fmt.Sprintf("block:%s:%s", blockerID, blockedID)
		s.redis.Del(ctx, cacheKey)
	}

	return nil
}

// UnblockUser unblocks a user
func (s *Service) UnblockUser(
	ctx context.Context,
	blockerID, blockedID uuid.UUID,
) error {
	result := s.db.WithContext(ctx).
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Delete(&model.UserBlock{})

	if result.Error != nil {
		return fmt.Errorf("failed to unblock user: %w", result.Error)
	}

	// Invalidate cache
	if s.redis != nil {
		cacheKey := fmt.Sprintf("block:%s:%s", blockerID, blockedID)
		s.redis.Del(ctx, cacheKey)
	}

	return nil
}

// GetBlockedUsers returns list of users blocked by blockerID
func (s *Service) GetBlockedUsers(
	ctx context.Context,
	blockerID uuid.UUID,
) ([]model.UserBlock, error) {
	var blocks []model.UserBlock

	err := s.db.WithContext(ctx).
		Where("blocker_id = ?", blockerID).
		Order("blocked_at DESC").
		Find(&blocks).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get blocked users: %w", err)
	}

	return blocks, nil
}

// CleanupExpiredDiscoveries removes expired discovery records
func (s *Service) CleanupExpiredDiscoveries(ctx context.Context) (int64, error) {
	result := s.db.WithContext(ctx).
		Where("expires_at <= NOW()").
		Delete(&model.BLEDiscovery{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup discoveries: %w", result.Error)
	}

	return result.RowsAffected, nil
}
