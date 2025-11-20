package model

import "time"

// MatrixRoom represents a Matrix room (station room)
type MatrixRoom struct {
	ID                string    `json:"id" db:"id"`                                   // !abc123:trainblink.org
	RoomAlias         string    `json:"room_alias" db:"room_alias"`                   // #tokyo-station:trainblink.org
	StationID         string    `json:"station_id" db:"station_id"`                   // Associated station
	Name              string    `json:"name" db:"name"`                               // Room display name
	Topic             string    `json:"topic" db:"topic"`                             // Room topic/description
	EncryptionEnabled bool      `json:"encryption_enabled" db:"encryption_enabled"`   // MLS encryption status
	MLSGroupID        string    `json:"mls_group_id,omitempty" db:"mls_group_id"`     // MLS group identifier
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	LastActiveAt      time.Time `json:"last_active_at" db:"last_active_at"`
	MemberCount       int       `json:"member_count" db:"member_count"`
}

// MatrixUser represents a Matrix user (anonymous temporary user)
type MatrixUser struct {
	ID           string    `json:"id" db:"id"`                       // @anon_xxx:trainblink.org
	UserID       string    `json:"user_id" db:"user_id"`             // TrainBlink user ID
	DisplayName  string    `json:"display_name" db:"display_name"`   // Display name
	AccessToken  string    `json:"access_token" db:"access_token"`   // Matrix access token (encrypted)
	DeviceID     string    `json:"device_id" db:"device_id"`         // Device identifier
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`       // Token expiration
	IsActive     bool      `json:"is_active" db:"is_active"`
}

// MatrixRoomMember represents room membership
type MatrixRoomMember struct {
	ID           string    `json:"id" db:"id"`
	RoomID       string    `json:"room_id" db:"room_id"`
	MatrixUserID string    `json:"matrix_user_id" db:"matrix_user_id"`
	Membership   string    `json:"membership" db:"membership"`       // join, leave, invite, ban
	JoinedAt     time.Time `json:"joined_at" db:"joined_at"`
	LeftAt       *time.Time `json:"left_at,omitempty" db:"left_at"`
}

// MatrixEvent represents a Matrix event
type MatrixEvent struct {
	ID             string    `json:"id" db:"id"`                           // $event-123
	Type           string    `json:"type" db:"type"`                       // m.room.message, m.room.member, etc.
	RoomID         string    `json:"room_id" db:"room_id"`
	Sender         string    `json:"sender" db:"sender"`                   // Matrix user ID
	Content        string    `json:"content" db:"content"`                 // JSON content
	StateKey       *string   `json:"state_key,omitempty" db:"state_key"`   // For state events
	Timestamp      time.Time `json:"timestamp" db:"timestamp"`
	TransactionID  string    `json:"transaction_id,omitempty" db:"transaction_id"`
}

// MatrixRoomInfo - Response structure for room information
type MatrixRoomInfo struct {
	RoomID            string `json:"room_id"`
	RoomAlias         string `json:"room_alias"`
	Name              string `json:"name"`
	Topic             string `json:"topic"`
	MemberCount       int    `json:"member_count"`
	EncryptionEnabled bool   `json:"encryption_enabled"`
	MLSEnabled        bool   `json:"mls_enabled"`
	MLSGroupID        string `json:"mls_group_id,omitempty"`
}

// MatrixResources - Matrix connection resources for client
type MatrixResources struct {
	RoomID            string `json:"room_id"`
	RoomAlias         string `json:"room_alias"`
	MatrixUserID      string `json:"matrix_user_id"`
	HomeserverURL     string `json:"homeserver_url"`
	AccessToken       string `json:"access_token"`
	EncryptionEnabled bool   `json:"encryption_enabled"`
	MLSGroupID        string `json:"mls_group_id,omitempty"`
	MemberCount       int    `json:"member_count"`
}
