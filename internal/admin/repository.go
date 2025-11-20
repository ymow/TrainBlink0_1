package admin

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrAdminNotFound     = errors.New("admin not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// Repository handles admin database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new admin repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new admin
func (r *Repository) Create(ctx context.Context, admin *Admin) error {
	// Check if email already exists
	var count int64
	if err := r.db.WithContext(ctx).Model(&Admin{}).Where("email = ?", admin.Email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrEmailAlreadyExists
	}

	return r.db.WithContext(ctx).Create(admin).Error
}

// GetByID retrieves admin by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Admin, error) {
	var admin Admin
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminNotFound
		}
		return nil, err
	}
	return &admin, nil
}

// GetByEmail retrieves admin by email
func (r *Repository) GetByEmail(ctx context.Context, email string) (*Admin, error) {
	var admin Admin
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminNotFound
		}
		return nil, err
	}
	return &admin, nil
}

// Update updates admin fields
func (r *Repository) Update(ctx context.Context, admin *Admin) error {
	return r.db.WithContext(ctx).Save(admin).Error
}

// UpdateLastLogin updates last login timestamp and IP
func (r *Repository) UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&Admin{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"last_login_ip": ip,
			"failed_login_attempts": 0, // Reset failed attempts on successful login
			"locked_until": nil,
		}).Error
}

// IncrementFailedLoginAttempts increments failed login counter
func (r *Repository) IncrementFailedLoginAttempts(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Admin{}).
		Where("id = ?", id).
		UpdateColumn("failed_login_attempts", gorm.Expr("failed_login_attempts + ?", 1)).
		Error
}

// LockAccount locks admin account until specified time
func (r *Repository) LockAccount(ctx context.Context, id uuid.UUID, until time.Time) error {
	return r.db.WithContext(ctx).Model(&Admin{}).
		Where("id = ?", id).
		Update("locked_until", until).
		Error
}

// List retrieves all admins with pagination
func (r *Repository) List(ctx context.Context, offset, limit int) ([]*Admin, int64, error) {
	var admins []*Admin
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).Model(&Admin{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&admins).Error

	return admins, total, err
}

// Delete soft deletes an admin (sets is_active to false)
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Admin{}).
		Where("id = ?", id).
		Update("is_active", false).
		Error
}

// HardDelete permanently deletes an admin
func (r *Repository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Admin{}, id).Error
}

// GetRole retrieves role by ID
func (r *Repository) GetRole(ctx context.Context, roleID string) (*Role, error) {
	var role Role
	err := r.db.WithContext(ctx).Where("id = ?", roleID).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &role, nil
}

// ListRoles retrieves all roles
func (r *Repository) ListRoles(ctx context.Context) ([]*Role, error) {
	var roles []*Role
	err := r.db.WithContext(ctx).Order("created_at ASC").Find(&roles).Error
	return roles, err
}
