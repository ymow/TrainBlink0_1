package model

import (
	"time"

	"github.com/google/uuid"
)

// BLEDiscovery represents a BLE proximity discovery event
// Grants 10-minute messaging permission between users
type BLEDiscovery struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	// Participants
	DiscovererID uuid.UUID `gorm:"type:uuid;not null;index:idx_permission_check,priority:1" json:"discoverer_id"`
	DiscoveredID uuid.UUID `gorm:"type:uuid;not null;index:idx_permission_check,priority:2" json:"discovered_id"`

	// Context
	TripID *uuid.UUID `gorm:"type:uuid;index" json:"trip_id,omitempty"`

	// BLE metadata
	RSSI             *int    `json:"rssi,omitempty"`
	DistanceEstimate *string `gorm:"type:varchar(20)" json:"distance_estimate,omitempty"`

	// Timing (CRITICAL for permission validation)
	DiscoveredAt time.Time `gorm:"not null;default:now()" json:"discovered_at"`
	ExpiresAt    time.Time `gorm:"not null;index:idx_permission_check,priority:3" json:"expires_at"`

	// Timestamps
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
}

// TableName specifies the table name
func (BLEDiscovery) TableName() string {
	return "ble_discoveries"
}

// IsExpired checks if discovery has expired (10 minutes)
func (d *BLEDiscovery) IsExpired() bool {
	return time.Now().After(d.ExpiresAt)
}

// IsValid checks if discovery is still valid for messaging
func (d *BLEDiscovery) IsValid() bool {
	return !d.IsExpired()
}

// TimeRemainingSeconds returns seconds until expiry
func (d *BLEDiscovery) TimeRemainingSeconds() int64 {
	if d.IsExpired() {
		return 0
	}
	return int64(time.Until(d.ExpiresAt).Seconds())
}
