package model

import (
	"time"

	"github.com/google/uuid"
)

// Trip represents a user's travel journey
type Trip struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`

	// Trip details
	Route       string  `gorm:"type:varchar(200);not null;index" json:"route"`  // "Tokyo → Osaka"
	TrainNumber *string `gorm:"type:varchar(50)" json:"train_number,omitempty"` // "Nozomi 123"

	// Timing
	DepartureTime    time.Time  `gorm:"type:timestamp with time zone;not null;index" json:"departure_time"`
	EstimatedArrival time.Time  `gorm:"type:timestamp with time zone;not null" json:"estimated_arrival"`
	ActualEndTime    *time.Time `gorm:"type:timestamp with time zone" json:"actual_end_time,omitempty"`

	// BLE Discovery
	DiscoveryEnabled bool   `gorm:"default:true" json:"discovery_enabled"`
	BLEAnonymousID   string `gorm:"type:varchar(50);not null;uniqueIndex" json:"ble_anonymous_id"` // "TB_abc123"

	// Status
	Status string `gorm:"type:varchar(20);default:'active';index;check:status IN ('active', 'ended', 'cancelled')" json:"status"`

	// Timestamps
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`
}

// TableName specifies the table name for Trip
func (Trip) TableName() string {
	return "trips"
}

// TripStatus constants
const (
	TripStatusActive    = "active"
	TripStatusEnded     = "ended"
	TripStatusCancelled = "cancelled"
)

// IsActive checks if trip is currently active
func (t *Trip) IsActive() bool {
	return t.Status == TripStatusActive
}

// HasEnded checks if trip has ended
func (t *Trip) HasEnded() bool {
	return t.Status == TripStatusEnded || t.Status == TripStatusCancelled
}

// Duration returns the estimated trip duration
func (t *Trip) Duration() time.Duration {
	return t.EstimatedArrival.Sub(t.DepartureTime)
}

// ActualDuration returns the actual trip duration if ended
func (t *Trip) ActualDuration() *time.Duration {
	if t.ActualEndTime == nil {
		return nil
	}
	duration := t.ActualEndTime.Sub(t.DepartureTime)
	return &duration
}

// Discovery represents an anonymized BLE discovery event
type Discovery struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	// Anonymized trip information (NO user_id for privacy!)
	TripRoute string `gorm:"type:varchar(200);not null;index" json:"trip_route"` // "Tokyo → Osaka"

	// Discovery metadata
	DiscoveredUserAnonymousID string  `gorm:"type:varchar(50);not null" json:"discovered_user_anonymous_id"`
	DistanceEstimate          *string `gorm:"type:varchar(20)" json:"distance_estimate,omitempty"` // "Close (2-10m)"
	RSSI                      *int    `json:"rssi,omitempty"`                                      // BLE signal strength

	// Optional demographics (anonymized, opt-in)
	DiscovererAgeRange *string `gorm:"type:varchar(20)" json:"discoverer_age_range,omitempty"` // "25-30"
	DiscoveredAgeRange *string `gorm:"type:varchar(20)" json:"discovered_age_range,omitempty"` // "25-30"
	DiscovererGender   *string `gorm:"type:char(1)" json:"discoverer_gender,omitempty"`        // M/F/O
	DiscoveredGender   *string `gorm:"type:char(1)" json:"discovered_gender,omitempty"`        // M/F/O

	// Timestamp
	DiscoveredAt time.Time `gorm:"default:now();index" json:"discovered_at"`
}

// TableName specifies the table name for Discovery
func (Discovery) TableName() string {
	return "discoveries"
}

// MatrixEphemeralRoom represents a temporary 1-on-1 DM room
type MatrixEphemeralRoom struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	RoomID string `gorm:"type:varchar(255);not null;uniqueIndex" json:"room_id"` // !dm_xyz:trainblink.org

	// Participants (via trips, not direct user_id for privacy)
	Trip1ID uuid.UUID `gorm:"type:uuid;not null;index" json:"trip1_id"`
	Trip2ID uuid.UUID `gorm:"type:uuid;not null;index" json:"trip2_id"`

	// Anonymous IDs for tracking (no user_id!)
	AnonymousID1 string `gorm:"type:varchar(50);not null" json:"anonymous_id_1"` // TB_abc123
	AnonymousID2 string `gorm:"type:varchar(50);not null" json:"anonymous_id_2"` // TB_def456

	// MLS encryption group
	MLSGroupID *string `gorm:"type:varchar(255);index" json:"mls_group_id,omitempty"`

	// Ephemeral settings
	ExpiresAt        time.Time  `gorm:"type:timestamp with time zone;not null;index" json:"expires_at"`
	AutoDeleteQueued bool       `gorm:"default:false;index" json:"auto_delete_queued"`
	DeletedAt        *time.Time `gorm:"type:timestamp with time zone" json:"deleted_at,omitempty"`

	// Metadata
	MessageCount  int        `gorm:"default:0" json:"message_count"`
	LastMessageAt *time.Time `gorm:"type:timestamp with time zone" json:"last_message_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
}

// TableName specifies the table name for MatrixEphemeralRoom
func (MatrixEphemeralRoom) TableName() string {
	return "matrix_ephemeral_rooms"
}

// IsExpired checks if room has expired
func (r *MatrixEphemeralRoom) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// IsDeleted checks if room has been deleted
func (r *MatrixEphemeralRoom) IsDeleted() bool {
	return r.DeletedAt != nil
}

// ShouldAutoDelete checks if room should be auto-deleted
func (r *MatrixEphemeralRoom) ShouldAutoDelete() bool {
	return r.IsExpired() && !r.IsDeleted() && !r.AutoDeleteQueued
}
