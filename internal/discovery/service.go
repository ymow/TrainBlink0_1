package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/ymow/messenger_protocol_research/internal/model"
	"gorm.io/gorm"
)

// Service handles discovery tracking operations (anonymized analytics)
type Service struct {
	db *gorm.DB
}

// NewService creates a new discovery service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// LogDiscoveryRequest contains parameters for logging a discovery event
type LogDiscoveryRequest struct {
	TripRoute                 string  `json:"trip_route" binding:"required"` // "Tokyo → Osaka"
	DiscoveredUserAnonymousID string  `json:"discovered_user_anonymous_id" binding:"required"`
	DistanceEstimate          *string `json:"distance_estimate,omitempty"`
	RSSI                      *int    `json:"rssi,omitempty"`

	// Optional demographics (anonymized)
	DiscovererAgeRange *string `json:"discoverer_age_range,omitempty"`
	DiscoveredAgeRange *string `json:"discovered_age_range,omitempty"`
	DiscovererGender   *string `json:"discoverer_gender,omitempty"`
	DiscoveredGender   *string `json:"discovered_gender,omitempty"`
}

// LogDiscovery records an anonymized discovery event for analytics
func (s *Service) LogDiscovery(ctx context.Context, req *LogDiscoveryRequest) error {
	discovery := &model.Discovery{
		TripRoute:                 req.TripRoute,
		DiscoveredUserAnonymousID: req.DiscoveredUserAnonymousID,
		DistanceEstimate:          req.DistanceEstimate,
		RSSI:                      req.RSSI,
		DiscovererAgeRange:        req.DiscovererAgeRange,
		DiscoveredAgeRange:        req.DiscoveredAgeRange,
		DiscovererGender:          req.DiscovererGender,
		DiscoveredGender:          req.DiscoveredGender,
		DiscoveredAt:              time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(discovery).Error; err != nil {
		return fmt.Errorf("failed to log discovery: %w", err)
	}

	return nil
}

// GetDiscoveriesByRoute retrieves discoveries for a specific route (analytics)
func (s *Service) GetDiscoveriesByRoute(ctx context.Context, route string, limit, offset int) ([]model.Discovery, int64, error) {
	var discoveries []model.Discovery
	var total int64

	// Count total
	if err := s.db.WithContext(ctx).
		Model(&model.Discovery{}).
		Where("trip_route = ?", route).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count discoveries: %w", err)
	}

	// Fetch discoveries
	if err := s.db.WithContext(ctx).
		Where("trip_route = ?", route).
		Order("discovered_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&discoveries).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch discoveries: %w", err)
	}

	return discoveries, total, nil
}

// GetDiscoveryStats returns discovery statistics
func (s *Service) GetDiscoveryStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total discoveries
	var totalDiscoveries int64
	if err := s.db.WithContext(ctx).
		Model(&model.Discovery{}).
		Count(&totalDiscoveries).Error; err != nil {
		return nil, err
	}
	stats["total_discoveries"] = totalDiscoveries

	// Discoveries last 24 hours
	var discoveries24h int64
	if err := s.db.WithContext(ctx).
		Model(&model.Discovery{}).
		Where("discovered_at > NOW() - INTERVAL '24 hours'").
		Count(&discoveries24h).Error; err != nil {
		return nil, err
	}
	stats["discoveries_24h"] = discoveries24h

	// Top routes
	type RouteStats struct {
		Route string `json:"route"`
		Count int64  `json:"count"`
	}
	var topRoutes []RouteStats
	s.db.WithContext(ctx).
		Model(&model.Discovery{}).
		Select("trip_route as route, COUNT(*) as count").
		Group("trip_route").
		Order("count DESC").
		Limit(10).
		Find(&topRoutes)
	stats["top_routes"] = topRoutes

	// Distance distribution
	type DistanceStats struct {
		Distance string `json:"distance"`
		Count    int64  `json:"count"`
	}
	var distanceDistribution []DistanceStats
	s.db.WithContext(ctx).
		Model(&model.Discovery{}).
		Select("distance_estimate as distance, COUNT(*) as count").
		Where("distance_estimate IS NOT NULL").
		Group("distance_estimate").
		Order("count DESC").
		Find(&distanceDistribution)
	stats["distance_distribution"] = distanceDistribution

	// Average RSSI by distance
	type RSSIStats struct {
		Distance string  `json:"distance"`
		AvgRSSI  float64 `json:"avg_rssi"`
		Count    int64   `json:"count"`
	}
	var rssiStats []RSSIStats
	s.db.WithContext(ctx).
		Model(&model.Discovery{}).
		Select("distance_estimate as distance, AVG(rssi) as avg_rssi, COUNT(*) as count").
		Where("distance_estimate IS NOT NULL AND rssi IS NOT NULL").
		Group("distance_estimate").
		Order("avg_rssi DESC").
		Find(&rssiStats)
	stats["rssi_by_distance"] = rssiStats

	return stats, nil
}

// GetRouteStats returns detailed statistics for a specific route
func (s *Service) GetRouteStats(ctx context.Context, route string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	stats["route"] = route

	// Total discoveries for this route
	var totalDiscoveries int64
	if err := s.db.WithContext(ctx).
		Model(&model.Discovery{}).
		Where("trip_route = ?", route).
		Count(&totalDiscoveries).Error; err != nil {
		return nil, err
	}
	stats["total_discoveries"] = totalDiscoveries

	// Discoveries by time of day
	type HourlyStats struct {
		Hour  int   `json:"hour"`
		Count int64 `json:"count"`
	}
	var hourlyStats []HourlyStats
	s.db.WithContext(ctx).
		Model(&model.Discovery{}).
		Select("EXTRACT(HOUR FROM discovered_at) as hour, COUNT(*) as count").
		Where("trip_route = ?", route).
		Group("hour").
		Order("hour").
		Find(&hourlyStats)
	stats["discoveries_by_hour"] = hourlyStats

	// Distance distribution for this route
	type DistanceCount struct {
		Distance string `json:"distance"`
		Count    int64  `json:"count"`
	}
	var distanceCounts []DistanceCount
	s.db.WithContext(ctx).
		Model(&model.Discovery{}).
		Select("distance_estimate as distance, COUNT(*) as count").
		Where("trip_route = ? AND distance_estimate IS NOT NULL", route).
		Group("distance_estimate").
		Order("count DESC").
		Find(&distanceCounts)
	stats["distance_distribution"] = distanceCounts

	// Demographics breakdown (if available)
	if s.hasDemo(ctx, route) {
		var genderStats []struct {
			Gender string `json:"gender"`
			Count  int64  `json:"count"`
		}
		s.db.WithContext(ctx).
			Model(&model.Discovery{}).
			Select("discovered_gender as gender, COUNT(*) as count").
			Where("trip_route = ? AND discovered_gender IS NOT NULL", route).
			Group("discovered_gender").
			Find(&genderStats)
		stats["gender_distribution"] = genderStats

		var ageStats []struct {
			AgeRange string `json:"age_range"`
			Count    int64  `json:"count"`
		}
		s.db.WithContext(ctx).
			Model(&model.Discovery{}).
			Select("discovered_age_range as age_range, COUNT(*) as count").
			Where("trip_route = ? AND discovered_age_range IS NOT NULL", route).
			Group("discovered_age_range").
			Order("count DESC").
			Find(&ageStats)
		stats["age_distribution"] = ageStats
	}

	return stats, nil
}

// GetDiscoveriesByDateRange retrieves discoveries within a date range
func (s *Service) GetDiscoveriesByDateRange(ctx context.Context, startDate, endDate time.Time) ([]model.Discovery, error) {
	var discoveries []model.Discovery

	if err := s.db.WithContext(ctx).
		Where("discovered_at BETWEEN ? AND ?", startDate, endDate).
		Order("discovered_at DESC").
		Find(&discoveries).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch discoveries: %w", err)
	}

	return discoveries, nil
}

// GetPopularRoutes returns the most popular routes by discovery count
func (s *Service) GetPopularRoutes(ctx context.Context, limit int) ([]RoutePopularity, error) {
	var routes []RoutePopularity

	err := s.db.WithContext(ctx).
		Model(&model.Discovery{}).
		Select(`
			trip_route as route,
			COUNT(*) as discovery_count,
			COUNT(DISTINCT discovered_user_anonymous_id) as unique_users,
			AVG(rssi) as avg_rssi,
			MAX(discovered_at) as last_discovered_at
		`).
		Group("trip_route").
		Order("discovery_count DESC").
		Limit(limit).
		Find(&routes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch popular routes: %w", err)
	}

	return routes, nil
}

type RoutePopularity struct {
	Route            string    `json:"route"`
	DiscoveryCount   int64     `json:"discovery_count"`
	UniqueUsers      int64     `json:"unique_users"`
	AvgRSSI          *float64  `json:"avg_rssi,omitempty"`
	LastDiscoveredAt time.Time `json:"last_discovered_at"`
}

// GetDiscoveryTrends returns discovery trends over time
func (s *Service) GetDiscoveryTrends(ctx context.Context, days int) ([]DailyTrend, error) {
	var trends []DailyTrend

	err := s.db.WithContext(ctx).
		Model(&model.Discovery{}).
		Select("DATE(discovered_at) as date, COUNT(*) as count").
		Where("discovered_at > NOW() - INTERVAL '? days'", days).
		Group("date").
		Order("date DESC").
		Find(&trends).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch discovery trends: %w", err)
	}

	return trends, nil
}

type DailyTrend struct {
	Date  time.Time `json:"date"`
	Count int64     `json:"count"`
}

// CleanupOldDiscoveries removes discovery records older than specified days (GDPR compliance)
func (s *Service) CleanupOldDiscoveries(ctx context.Context, daysToKeep int) (int, error) {
	cutoffDate := time.Now().AddDate(0, 0, -daysToKeep)

	result := s.db.WithContext(ctx).
		Where("discovered_at < ?", cutoffDate).
		Delete(&model.Discovery{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old discoveries: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

// hasDemo checks if route has demographic data
func (s *Service) hasDemo(ctx context.Context, route string) bool {
	var count int64
	s.db.WithContext(ctx).
		Model(&model.Discovery{}).
		Where("trip_route = ? AND (discovered_gender IS NOT NULL OR discovered_age_range IS NOT NULL)", route).
		Count(&count)
	return count > 0
}
