package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a mobile app user
type User struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`

	// Authentication
	FirebaseUID   string  `json:"firebase_uid" gorm:"uniqueIndex;not null"`
	MatrixUserID  *string `json:"matrix_user_id,omitempty" gorm:"index"`

	// Profile
	DisplayName *string `json:"display_name,omitempty"`
	AvatarEmoji *string `json:"avatar_emoji,omitempty"`
	AvatarColor *string `json:"avatar_color,omitempty"`
	StatusText  *string `json:"status_text,omitempty"`

	// Settings
	Preferences map[string]interface{} `json:"preferences" gorm:"type:jsonb;default:'{}'"`

	// Status
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	IsBanned  bool       `json:"is_banned" gorm:"default:false;index"`
	BanReason *string    `json:"ban_reason,omitempty"`
	BannedAt  *time.Time `json:"banned_at,omitempty"`
	BannedBy  *uuid.UUID `json:"banned_by,omitempty"`
	BanUntil  *time.Time `json:"ban_until,omitempty"`

	// Statistics
	TotalSessions      int `json:"total_sessions" gorm:"default:0"`
	TotalEncounters    int `json:"total_encounters" gorm:"default:0"`
	TotalMessagesSent  int `json:"total_messages_sent" gorm:"default:0"`
	TotalContentShared int `json:"total_content_shared" gorm:"default:0"`

	// Timestamps
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty" gorm:"index"`

	// Privacy
	DataRetentionDays int `json:"data_retention_days" gorm:"default:1"`
}

// TableName specifies the table name
func (User) TableName() string {
	return "users"
}

// UpdateProfileRequest represents user profile update request
type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name"`
	AvatarEmoji *string `json:"avatar_emoji"`
	AvatarColor *string `json:"avatar_color"`
	StatusText  *string `json:"status_text"`
}

// UserResponse is the public user response
type UserResponse struct {
	ID            uuid.UUID              `json:"id"`
	FirebaseUID   string                 `json:"firebase_uid"`
	DisplayName   *string                `json:"display_name,omitempty"`
	AvatarEmoji   *string                `json:"avatar_emoji,omitempty"`
	AvatarColor   *string                `json:"avatar_color,omitempty"`
	StatusText    *string                `json:"status_text,omitempty"`
	Preferences   map[string]interface{} `json:"preferences"`
	IsActive      bool                   `json:"is_active"`
	IsBanned      bool                   `json:"is_banned"`
	TotalSessions int                    `json:"total_sessions"`
	CreatedAt     time.Time              `json:"created_at"`
	LastSeenAt    *time.Time             `json:"last_seen_at,omitempty"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:            u.ID,
		FirebaseUID:   u.FirebaseUID,
		DisplayName:   u.DisplayName,
		AvatarEmoji:   u.AvatarEmoji,
		AvatarColor:   u.AvatarColor,
		StatusText:    u.StatusText,
		Preferences:   u.Preferences,
		IsActive:      u.IsActive,
		IsBanned:      u.IsBanned,
		TotalSessions: u.TotalSessions,
		CreatedAt:     u.CreatedAt,
		LastSeenAt:    u.LastSeenAt,
	}
}
