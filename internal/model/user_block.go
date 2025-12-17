package model

import (
	"time"

	"github.com/google/uuid"
)

// UserBlock represents a user blocking relationship
// Blocked users cannot send messages to their blocker
type UserBlock struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	// Participants
	BlockerID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_unique_block,priority:1" json:"blocker_id"`
	BlockedID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_unique_block,priority:2" json:"blocked_id"`

	// Optional metadata
	Reason *string `gorm:"type:text" json:"reason,omitempty"`

	// Timestamps
	BlockedAt time.Time `gorm:"default:now()" json:"blocked_at"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
}

// TableName specifies the table name
func (UserBlock) TableName() string {
	return "user_blocks"
}
