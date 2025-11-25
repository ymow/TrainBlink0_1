package trip

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/ymow/messenger_protocol_research/internal/model"
	"gorm.io/gorm"
)

// Service handles trip management operations
type Service struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewService creates a new trip service
func NewService(db *gorm.DB, redis *redis.Client) *Service {
	return &Service{
		db:    db,
		redis: redis,
	}
}

// StartTripRequest contains parameters for starting a new trip
type StartTripRequest struct {
	UserID           uuid.UUID `json:"user_id"`
	Route            string    `json:"route" binding:"required"` // "Tokyo → Osaka"
	TrainNumber      *string   `json:"train_number,omitempty"`   // "Nozomi 123"
	DepartureTime    time.Time `json:"departure_time" binding:"required"`
	EstimatedArrival time.Time `json:"estimated_arrival" binding:"required"`
}

// StartTrip creates a new active trip for a user
func (s *Service) StartTrip(ctx context.Context, req *StartTripRequest) (*model.Trip, error) {
	// Validate times
	if req.EstimatedArrival.Before(req.DepartureTime) {
		return nil, fmt.Errorf("estimated arrival must be after departure time")
	}

	// Check if user already has an active trip
	var activeTrip model.Trip
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", req.UserID, model.TripStatusActive).
		First(&activeTrip).Error

	if err == nil {
		return nil, fmt.Errorf("user already has an active trip: %s", activeTrip.Route)
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check active trips: %w", err)
	}

	// Generate anonymous BLE ID
	bleID, err := s.generateBLEAnonymousID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate BLE ID: %w", err)
	}

	// Create trip
	trip := &model.Trip{
		UserID:           req.UserID,
		Route:            req.Route,
		TrainNumber:      req.TrainNumber,
		DepartureTime:    req.DepartureTime,
		EstimatedArrival: req.EstimatedArrival,
		DiscoveryEnabled: true,
		BLEAnonymousID:   bleID,
		Status:           model.TripStatusActive,
	}

	if err := s.db.WithContext(ctx).Create(trip).Error; err != nil {
		return nil, fmt.Errorf("failed to create trip: %w", err)
	}

	// Cache in Redis for fast lookup (BLE ID → Trip ID)
	if s.redis != nil {
		cacheKey := fmt.Sprintf("trip:ble:%s", bleID)
		if err := s.redis.Set(ctx, cacheKey, trip.ID.String(), 24*time.Hour).Err(); err != nil {
			// Log but don't fail
			fmt.Printf("Warning: failed to cache trip: %v\n", err)
		}

		// Cache user's active trip
		userCacheKey := fmt.Sprintf("trip:user:%s:active", req.UserID.String())
		if err := s.redis.Set(ctx, userCacheKey, trip.ID.String(), 24*time.Hour).Err(); err != nil {
			fmt.Printf("Warning: failed to cache user trip: %v\n", err)
		}
	}

	return trip, nil
}

// EndTripRequest contains parameters for ending a trip
type EndTripRequest struct {
	TripID uuid.UUID `json:"trip_id" binding:"required"`
	UserID uuid.UUID `json:"user_id"`
}

// EndTrip marks a trip as ended
func (s *Service) EndTrip(ctx context.Context, req *EndTripRequest) (*model.Trip, error) {
	var trip model.Trip
	err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", req.TripID, req.UserID).
		First(&trip).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("trip not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trip: %w", err)
	}

	if trip.HasEnded() {
		return &trip, nil // Already ended
	}

	// Update trip
	now := time.Now()
	trip.Status = model.TripStatusEnded
	trip.ActualEndTime = &now

	if err := s.db.WithContext(ctx).Save(&trip).Error; err != nil {
		return nil, fmt.Errorf("failed to end trip: %w", err)
	}

	// Remove from Redis cache
	if s.redis != nil {
		cacheKey := fmt.Sprintf("trip:ble:%s", trip.BLEAnonymousID)
		s.redis.Del(ctx, cacheKey)

		userCacheKey := fmt.Sprintf("trip:user:%s:active", req.UserID.String())
		s.redis.Del(ctx, userCacheKey)
	}

	return &trip, nil
}

// GetActiveTrip retrieves user's current active trip
func (s *Service) GetActiveTrip(ctx context.Context, userID uuid.UUID) (*model.Trip, error) {
	// Try cache first (if Redis is available)
	if s.redis != nil {
		userCacheKey := fmt.Sprintf("trip:user:%s:active", userID.String())
		tripIDStr, err := s.redis.Get(ctx, userCacheKey).Result()

		if err == nil && tripIDStr != "" {
			// Found in cache, fetch from DB
			tripID, err := uuid.Parse(tripIDStr)
			if err == nil {
				var trip model.Trip
				if err := s.db.WithContext(ctx).First(&trip, tripID).Error; err == nil {
					return &trip, nil
				}
			}
		}
	}

	// Not in cache or cache miss, query DB
	var trip model.Trip
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, model.TripStatusActive).
		Order("created_at DESC").
		First(&trip).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil // No active trip
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active trip: %w", err)
	}

	// Update cache
	if s.redis != nil {
		userCacheKey := fmt.Sprintf("trip:user:%s:active", userID.String())
		s.redis.Set(ctx, userCacheKey, trip.ID.String(), 24*time.Hour)
	}

	return &trip, nil
}

// GetTripByID retrieves a trip by ID
func (s *Service) GetTripByID(ctx context.Context, tripID, userID uuid.UUID) (*model.Trip, error) {
	var trip model.Trip
	err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", tripID, userID).
		First(&trip).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("trip not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trip: %w", err)
	}

	return &trip, nil
}

// GetTripByBLEID retrieves a trip by BLE anonymous ID (for discovery)
func (s *Service) GetTripByBLEID(ctx context.Context, bleID string) (*model.Trip, error) {
	// Try cache first (if Redis is available)
	if s.redis != nil {
		cacheKey := fmt.Sprintf("trip:ble:%s", bleID)
		tripIDStr, err := s.redis.Get(ctx, cacheKey).Result()

		if err == nil && tripIDStr != "" {
			tripID, err := uuid.Parse(tripIDStr)
			if err == nil {
				var trip model.Trip
				if err := s.db.WithContext(ctx).First(&trip, tripID).Error; err == nil {
					return &trip, nil
				}
			}
		}
	}

	// Not in cache, query DB
	var trip model.Trip
	err := s.db.WithContext(ctx).
		Where("ble_anonymous_id = ? AND status = ?", bleID, model.TripStatusActive).
		First(&trip).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("trip not found for BLE ID: %s", bleID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trip: %w", err)
	}

	// Update cache
	if s.redis != nil {
		cacheKey := fmt.Sprintf("trip:ble:%s", bleID)
		s.redis.Set(ctx, cacheKey, trip.ID.String(), 24*time.Hour)
	}

	return &trip, nil
}

// GetUserTrips retrieves all trips for a user
func (s *Service) GetUserTrips(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Trip, int64, error) {
	var trips []model.Trip
	var total int64

	// Count total
	if err := s.db.WithContext(ctx).
		Model(&model.Trip{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count trips: %w", err)
	}

	// Fetch trips
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&trips).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch trips: %w", err)
	}

	return trips, total, nil
}

// CancelTrip cancels an active trip
func (s *Service) CancelTrip(ctx context.Context, tripID, userID uuid.UUID) error {
	var trip model.Trip
	err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", tripID, userID).
		First(&trip).Error

	if err == gorm.ErrRecordNotFound {
		return fmt.Errorf("trip not found")
	}
	if err != nil {
		return fmt.Errorf("failed to fetch trip: %w", err)
	}

	if trip.HasEnded() {
		return fmt.Errorf("trip already ended")
	}

	// Update status
	trip.Status = model.TripStatusCancelled
	now := time.Now()
	trip.ActualEndTime = &now

	if err := s.db.WithContext(ctx).Save(&trip).Error; err != nil {
		return fmt.Errorf("failed to cancel trip: %w", err)
	}

	// Clear cache
	if s.redis != nil {
		cacheKey := fmt.Sprintf("trip:ble:%s", trip.BLEAnonymousID)
		s.redis.Del(ctx, cacheKey)

		userCacheKey := fmt.Sprintf("trip:user:%s:active", userID.String())
		s.redis.Del(ctx, userCacheKey)
	}

	return nil
}

// UpdateDiscoveryEnabled enables/disables discovery for a trip
func (s *Service) UpdateDiscoveryEnabled(ctx context.Context, tripID, userID uuid.UUID, enabled bool) error {
	result := s.db.WithContext(ctx).
		Model(&model.Trip{}).
		Where("id = ? AND user_id = ? AND status = ?", tripID, userID, model.TripStatusActive).
		Update("discovery_enabled", enabled)

	if result.Error != nil {
		return fmt.Errorf("failed to update discovery: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("trip not found or not active")
	}

	return nil
}

// GetActiveTripStats returns statistics about active trips
func (s *Service) GetActiveTripStats(ctx context.Context) (map[string]interface{}, error) {
	var totalActive int64
	if err := s.db.WithContext(ctx).
		Model(&model.Trip{}).
		Where("status = ?", model.TripStatusActive).
		Count(&totalActive).Error; err != nil {
		return nil, err
	}

	// Get route distribution
	type RouteCount struct {
		Route string
		Count int64
	}
	var routeCounts []RouteCount
	s.db.WithContext(ctx).
		Model(&model.Trip{}).
		Select("route, COUNT(*) as count").
		Where("status = ?", model.TripStatusActive).
		Group("route").
		Order("count DESC").
		Limit(10).
		Find(&routeCounts)

	return map[string]interface{}{
		"total_active": totalActive,
		"top_routes":   routeCounts,
	}, nil
}

// generateBLEAnonymousID generates a unique anonymous ID for BLE broadcasting
func (s *Service) generateBLEAnonymousID() (string, error) {
	// Generate 6 random bytes
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode to base64 and take first 8 characters
	encoded := base64.URLEncoding.EncodeToString(bytes)
	if len(encoded) > 8 {
		encoded = encoded[:8]
	}

	// Format as TB_xxxxx
	return fmt.Sprintf("TB_%s", encoded), nil
}

// CleanupExpiredTrips marks trips as ended if they've passed estimated arrival + buffer
func (s *Service) CleanupExpiredTrips(ctx context.Context) (int, error) {
	// Mark trips as ended if estimated_arrival + 2 hours has passed
	bufferTime := 2 * time.Hour
	cutoffTime := time.Now().Add(-bufferTime)

	result := s.db.WithContext(ctx).
		Model(&model.Trip{}).
		Where("status = ? AND estimated_arrival < ?", model.TripStatusActive, cutoffTime).
		Updates(map[string]interface{}{
			"status":          model.TripStatusEnded,
			"actual_end_time": time.Now(),
		})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup trips: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}
