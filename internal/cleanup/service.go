package cleanup

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/ymow/messenger_protocol_research/internal/discovery"
	"github.com/ymow/messenger_protocol_research/internal/matrix"
	"github.com/ymow/messenger_protocol_research/internal/trip"
)

// Service handles automatic cleanup of expired data
type Service struct {
	db                *gorm.DB
	redis             *redis.Client
	tripService       *trip.Service
	discoveryService  *discovery.Service
	ephemeralRoomMgr  *matrix.EphemeralRoomManager

	// Cleanup intervals
	tripCleanupInterval      time.Duration
	roomCleanupInterval      time.Duration
	discoveryCleanupInterval time.Duration

	// Retention periods
	discoveryRetentionDays int
}

// NewService creates a new cleanup service
func NewService(
	db *gorm.DB,
	redisClient *redis.Client,
	tripService *trip.Service,
	discoveryService *discovery.Service,
	ephemeralRoomMgr *matrix.EphemeralRoomManager,
) *Service {
	return &Service{
		db:                       db,
		redis:                    redisClient,
		tripService:              tripService,
		discoveryService:         discoveryService,
		ephemeralRoomMgr:         ephemeralRoomMgr,
		tripCleanupInterval:      1 * time.Hour,  // Run hourly
		roomCleanupInterval:      30 * time.Minute, // Run every 30 minutes
		discoveryCleanupInterval: 24 * time.Hour,   // Run daily
		discoveryRetentionDays:   90,               // Keep 90 days of discovery data
	}
}

// SetTripCleanupInterval sets the trip cleanup interval
func (s *Service) SetTripCleanupInterval(interval time.Duration) {
	s.tripCleanupInterval = interval
}

// SetRoomCleanupInterval sets the room cleanup interval
func (s *Service) SetRoomCleanupInterval(interval time.Duration) {
	s.roomCleanupInterval = interval
}

// SetDiscoveryRetentionDays sets the discovery data retention period
func (s *Service) SetDiscoveryRetentionDays(days int) {
	s.discoveryRetentionDays = days
}

// Start starts the cleanup service with background workers
func (s *Service) Start(ctx context.Context) {
	log.Println("🧹 Starting cleanup service...")

	// Start trip cleanup worker
	go s.tripCleanupWorker(ctx)

	// Start ephemeral room cleanup worker
	go s.roomCleanupWorker(ctx)

	// Start discovery cleanup worker
	go s.discoveryCleanupWorker(ctx)

	log.Println("✅ Cleanup service started")
}

// tripCleanupWorker periodically cleans up expired trips
func (s *Service) tripCleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(s.tripCleanupInterval)
	defer ticker.Stop()

	// Run immediately on start
	s.cleanupExpiredTrips(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Trip cleanup worker stopped")
			return
		case <-ticker.C:
			s.cleanupExpiredTrips(ctx)
		}
	}
}

// roomCleanupWorker periodically cleans up expired ephemeral rooms
func (s *Service) roomCleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(s.roomCleanupInterval)
	defer ticker.Stop()

	// Run immediately on start
	s.cleanupExpiredRooms(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Room cleanup worker stopped")
			return
		case <-ticker.C:
			s.cleanupExpiredRooms(ctx)
		}
	}
}

// discoveryCleanupWorker periodically cleans up old discovery data
func (s *Service) discoveryCleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(s.discoveryCleanupInterval)
	defer ticker.Stop()

	// Run immediately on start
	s.cleanupOldDiscoveries(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Discovery cleanup worker stopped")
			return
		case <-ticker.C:
			s.cleanupOldDiscoveries(ctx)
		}
	}
}

// cleanupExpiredTrips cleans up trips that have passed their estimated arrival + buffer
func (s *Service) cleanupExpiredTrips(ctx context.Context) {
	count, err := s.tripService.CleanupExpiredTrips(ctx)
	if err != nil {
		log.Printf("❌ Failed to cleanup expired trips: %v", err)
		return
	}

	if count > 0 {
		log.Printf("🧹 Cleaned up %d expired trips", count)
	}
}

// cleanupExpiredRooms cleans up expired ephemeral rooms
func (s *Service) cleanupExpiredRooms(ctx context.Context) {
	count, err := s.ephemeralRoomMgr.DeleteExpiredRooms(ctx)
	if err != nil {
		log.Printf("❌ Failed to cleanup expired rooms: %v", err)
		return
	}

	if count > 0 {
		log.Printf("🧹 Cleaned up %d expired ephemeral rooms", count)
	}
}

// cleanupOldDiscoveries cleans up old discovery data (GDPR compliance)
func (s *Service) cleanupOldDiscoveries(ctx context.Context) {
	count, err := s.discoveryService.CleanupOldDiscoveries(ctx, s.discoveryRetentionDays)
	if err != nil {
		log.Printf("❌ Failed to cleanup old discoveries: %v", err)
		return
	}

	if count > 0 {
		log.Printf("🧹 Cleaned up %d old discovery records (older than %d days)",
			count, s.discoveryRetentionDays)
	}
}

// RunManualCleanup runs all cleanup tasks immediately
func (s *Service) RunManualCleanup(ctx context.Context) (*CleanupStats, error) {
	stats := &CleanupStats{}

	// Cleanup expired trips
	tripCount, err := s.tripService.CleanupExpiredTrips(ctx)
	if err != nil {
		return nil, fmt.Errorf("trip cleanup failed: %w", err)
	}
	stats.TripsDeleted = tripCount

	// Cleanup expired rooms
	roomCount, err := s.ephemeralRoomMgr.DeleteExpiredRooms(ctx)
	if err != nil {
		return nil, fmt.Errorf("room cleanup failed: %w", err)
	}
	stats.RoomsDeleted = roomCount

	// Cleanup old discoveries
	discoveryCount, err := s.discoveryService.CleanupOldDiscoveries(ctx, s.discoveryRetentionDays)
	if err != nil {
		return nil, fmt.Errorf("discovery cleanup failed: %w", err)
	}
	stats.DiscoveriesDeleted = discoveryCount

	log.Printf("🧹 Manual cleanup completed: %d trips, %d rooms, %d discoveries",
		tripCount, roomCount, discoveryCount)

	return stats, nil
}

// GetCleanupStats returns cleanup statistics
func (s *Service) GetCleanupStats(ctx context.Context) (*CleanupStats, error) {
	stats := &CleanupStats{}

	// Trip stats
	var expiredTrips int64
	twoHoursAgo := time.Now().Add(-2 * time.Hour)
	if err := s.db.WithContext(ctx).
		Table("trips").
		Where("status = ? AND estimated_arrival < ?", "active", twoHoursAgo).
		Count(&expiredTrips).Error; err != nil {
		return nil, fmt.Errorf("failed to count expired trips: %w", err)
	}
	stats.ExpiredTripsCount = int(expiredTrips)

	// Room stats
	var expiredRooms int64
	now := time.Now()
	if err := s.db.WithContext(ctx).
		Table("matrix_ephemeral_rooms").
		Where("expires_at < ? AND deleted_at IS NULL", now).
		Count(&expiredRooms).Error; err != nil {
		return nil, fmt.Errorf("failed to count expired rooms: %w", err)
	}
	stats.ExpiredRoomsCount = int(expiredRooms)

	// Discovery stats
	var oldDiscoveries int64
	cutoffDate := time.Now().AddDate(0, 0, -s.discoveryRetentionDays)
	if err := s.db.WithContext(ctx).
		Table("discoveries").
		Where("discovered_at < ?", cutoffDate).
		Count(&oldDiscoveries).Error; err != nil {
		return nil, fmt.Errorf("failed to count old discoveries: %w", err)
	}
	stats.OldDiscoveriesCount = int(oldDiscoveries)

	return stats, nil
}

// WarnExpiringRooms sends warnings for rooms expiring soon (useful for notifications)
func (s *Service) WarnExpiringRooms(ctx context.Context, within time.Duration) ([]ExpiringRoomWarning, error) {
	rooms, err := s.ephemeralRoomMgr.GetExpiringSoonRooms(ctx, within)
	if err != nil {
		return nil, fmt.Errorf("failed to get expiring rooms: %w", err)
	}

	warnings := make([]ExpiringRoomWarning, len(rooms))
	for i, room := range rooms {
		timeUntilExpiry := time.Until(room.ExpiresAt)
		warnings[i] = ExpiringRoomWarning{
			RoomID:         room.RoomID,
			Trip1ID:        room.Trip1ID,
			Trip2ID:        room.Trip2ID,
			ExpiresAt:      room.ExpiresAt,
			TimeRemaining:  timeUntilExpiry,
			MessageCount:   room.MessageCount,
		}
	}

	if len(warnings) > 0 {
		log.Printf("⚠️  %d rooms expiring within %s", len(warnings), within)
	}

	return warnings, nil
}

// CleanupStats contains cleanup statistics
type CleanupStats struct {
	TripsDeleted         int `json:"trips_deleted"`
	RoomsDeleted         int `json:"rooms_deleted"`
	DiscoveriesDeleted   int `json:"discoveries_deleted"`
	ExpiredTripsCount    int `json:"expired_trips_count"`
	ExpiredRoomsCount    int `json:"expired_rooms_count"`
	OldDiscoveriesCount  int `json:"old_discoveries_count"`
}

// ExpiringRoomWarning contains information about an expiring room
type ExpiringRoomWarning struct {
	RoomID        string        `json:"room_id"`
	Trip1ID       interface{}   `json:"trip1_id"`
	Trip2ID       interface{}   `json:"trip2_id"`
	ExpiresAt     time.Time     `json:"expires_at"`
	TimeRemaining time.Duration `json:"time_remaining"`
	MessageCount  int           `json:"message_count"`
}
