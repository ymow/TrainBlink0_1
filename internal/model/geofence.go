package model

import (
	"time"

	"github.com/google/uuid"
)

// Coordinates represents GPS coordinates
type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  float64 `json:"accuracy"`
}

// Capabilities represents client capabilities
type Capabilities struct {
	P2PEnabled    bool `json:"p2p_enabled"`
	MatrixEnabled bool `json:"matrix_enabled"`
	MLSSupported  bool `json:"mls_supported"`
}

// UserSession represents a geofence session
type UserSession struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	DeviceID  string    `json:"device_id" db:"device_id"`
	StationID string    `json:"station_id" db:"station_id"`

	// Timeline
	EnteredAt       time.Time  `json:"entered_at" db:"entered_at"`
	ExitedAt        *time.Time `json:"exited_at,omitempty" db:"exited_at"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" db:"duration_seconds"`

	// Entry/Exit Coordinates
	EntryLatitude  float64  `json:"entry_latitude" db:"entry_latitude"`
	EntryLongitude float64  `json:"entry_longitude" db:"entry_longitude"`
	ExitLatitude   *float64 `json:"exit_latitude,omitempty" db:"exit_latitude"`
	ExitLongitude  *float64 `json:"exit_longitude,omitempty" db:"exit_longitude"`

	// P2P Activity Stats
	P2PChatsCreated  int `json:"p2p_chats_created" db:"p2p_chats_created"`
	P2PMessagesSent  int `json:"p2p_messages_sent" db:"p2p_messages_sent"`
	P2PContentShared int `json:"p2p_content_shared" db:"p2p_content_shared"`
	Encounters       int `json:"encounters" db:"encounters"`

	// Matrix Activity Stats
	MatrixUserID       string     `json:"matrix_user_id" db:"matrix_user_id"`
	MatrixRoomID       string     `json:"matrix_room_id" db:"matrix_room_id"`
	MatrixMessagesSent int        `json:"matrix_messages_sent" db:"matrix_messages_sent"`
	MatrixJoinedAt     *time.Time `json:"matrix_joined_at,omitempty" db:"matrix_joined_at"`
	MatrixLeftAt       *time.Time `json:"matrix_left_at,omitempty" db:"matrix_left_at"`

	// Metadata
	ClientVersion    string `json:"client_version" db:"client_version"`
	CapabilitiesJSON string `json:"capabilities" db:"capabilities"` // JSON stored

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// EnterStationRequest - Request to enter a station
type EnterStationRequest struct {
	StationID     string       `json:"station_id" binding:"required"`
	Coordinates   Coordinates  `json:"coordinates" binding:"required"`
	Timestamp     time.Time    `json:"timestamp"`
	ClientVersion string       `json:"client_version"`
	Capabilities  Capabilities `json:"capabilities"`
}

// ExitStationRequest - Request to exit a station
type ExitStationRequest struct {
	SessionID       string          `json:"session_id" binding:"required"`
	StationID       string          `json:"station_id" binding:"required"`
	Timestamp       time.Time       `json:"timestamp"`
	DurationSeconds int             `json:"duration_seconds"`
	Activity        ActivitySummary `json:"activity"`
}

// ActivitySummary - Summary of user activity during session
type ActivitySummary struct {
	P2PChatsCreated    int `json:"p2p_chats_created"`
	P2PMessagesSent    int `json:"p2p_messages_sent"`
	ContentShared      int `json:"content_shared"`
	MatrixMessagesSent int `json:"matrix_messages_sent"`
}

// P2PResources - P2P connection resources
type P2PResources struct {
	ActivePeersNearby int         `json:"active_peers_nearby"`
	SignalingServer   string      `json:"signaling_server"`
	ICEServers        []ICEServer `json:"ice_servers"`
}

// ICEServer - ICE server configuration
type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

// Recommendations - Content recommendations
type Recommendations struct {
	Peers           []string `json:"peers"`
	IcebreakerCards []string `json:"icebreaker_cards"`
}

// EnterStationResponse - Response when entering a station
type EnterStationResponse struct {
	SessionID        uuid.UUID        `json:"session_id"`
	Station          *Station         `json:"station"`
	P2PResources     *P2PResources    `json:"p2p,omitempty"`
	MatrixResources  *MatrixResources `json:"matrix,omitempty"`
	ContentAvailable int              `json:"content_available"`
	Recommendations  *Recommendations `json:"recommendations,omitempty"`
}

// ExitStationResponse - Response when exiting a station
type ExitStationResponse struct {
	SessionSummary ActivitySummary `json:"session_summary"`
	CleanupStatus  CleanupStatus   `json:"cleanup_status"`
}

// CleanupStatus - Status of cleanup operations
type CleanupStatus struct {
	P2PClosed        bool `json:"p2p_closed"`
	MatrixLeft       bool `json:"matrix_left"`
	LocalDataCleared bool `json:"local_data_cleared"`
}
