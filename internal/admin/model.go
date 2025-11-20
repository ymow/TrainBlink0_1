package admin

import (
	"time"

	"github.com/google/uuid"
)

// Admin represents an admin user
type Admin struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`

	// Authentication
	Email        string `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string `json:"-" gorm:"not null"`

	// Profile
	Name string `json:"name" gorm:"not null"`

	// Role & Permissions
	Role        string   `json:"role" gorm:"not null;default:'moderator'"`
	Permissions []string `json:"permissions" gorm:"type:jsonb;default:'[]'"`

	// Status
	IsActive bool `json:"is_active" gorm:"default:true;index"`

	// Security
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP         *string    `json:"last_login_ip,omitempty"`
	FailedLoginAttempts int        `json:"-" gorm:"default:0"`
	LockedUntil         *time.Time `json:"-"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`
}

// TableName specifies the table name
func (Admin) TableName() string {
	return "admins"
}

// LoginRequest represents admin login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents admin login response
type LoginResponse struct {
	Admin        *AdminResponse `json:"admin"`
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresIn    int64          `json:"expires_in"`
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// CreateAdminRequest represents create admin request
type CreateAdminRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=super_admin admin moderator"`
}

// AdminResponse is the public admin response
type AdminResponse struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Role        string     `json:"role"`
	Permissions []string   `json:"permissions"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ToResponse converts Admin to AdminResponse
func (a *Admin) ToResponse() *AdminResponse {
	return &AdminResponse{
		ID:          a.ID,
		Email:       a.Email,
		Name:        a.Name,
		Role:        a.Role,
		Permissions: a.Permissions,
		IsActive:    a.IsActive,
		LastLoginAt: a.LastLoginAt,
		CreatedAt:   a.CreatedAt,
	}
}

// Role represents a permission role
type Role struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"not null"`
	Description *string                `json:"description"`
	Permissions map[string]interface{} `json:"permissions" gorm:"type:jsonb;not null"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TableName specifies the table name
func (Role) TableName() string {
	return "roles"
}
