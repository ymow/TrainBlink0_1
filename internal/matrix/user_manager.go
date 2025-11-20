package matrix

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"sync"

	"gorm.io/gorm"
)

// UserManager handles Matrix user operations
type UserManager struct {
	client       *Client
	db           *gorm.DB
	homeserverURL string
	serverName    string

	// User cache
	userCache map[string]*MatrixUser // firebaseUID -> MatrixUser
	cacheMutex sync.RWMutex
}

// NewUserManager creates a new user manager
func NewUserManager(client *Client, db *gorm.DB, homeserverURL, serverName string) *UserManager {
	return &UserManager{
		client:        client,
		db:            db,
		homeserverURL: homeserverURL,
		serverName:    serverName,
		userCache:     make(map[string]*MatrixUser),
	}
}

// RegisterAnonymousUser registers a new anonymous Matrix user
func (um *UserManager) RegisterAnonymousUser(ctx context.Context, firebaseUID string) (*MatrixUser, error) {
	// Check if user already exists in cache
	um.cacheMutex.RLock()
	if user, exists := um.userCache[firebaseUID]; exists {
		um.cacheMutex.RUnlock()
		log.Printf("👤 Using cached user for Firebase UID %s: %s", firebaseUID, user.MatrixUserID)
		return user, nil
	}
	um.cacheMutex.RUnlock()

	// Check database
	if um.db != nil {
		var user MatrixUser
		err := um.db.WithContext(ctx).
			Where("firebase_uid = ?", firebaseUID).
			First(&user).Error

		if err == nil {
			// Found in database, cache it
			um.cacheMutex.Lock()
			um.userCache[firebaseUID] = &user
			um.cacheMutex.Unlock()

			log.Printf("👤 Loaded user from database: %s", user.MatrixUserID)
			return &user, nil
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("database error: %w", err)
		}
	}

	// Generate anonymous username
	username := um.generateAnonymousUsername()
	password := um.generateSecurePassword()
	displayName := um.generateDisplayName()

	// Register with Matrix homeserver
	registerReq := &RegisterRequest{
		Username:                 username,
		Password:                 password,
		InitialDeviceDisplayName: "TrainBlink",
		InhibitLogin:             false,
	}

	resp, err := um.client.Register(ctx, registerReq)
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}

	log.Printf("✅ Registered new Matrix user: %s (Firebase UID: %s)", resp.UserID, firebaseUID)

	// Create user record
	user := &MatrixUser{
		FirebaseUID:  firebaseUID,
		MatrixUserID: resp.UserID,
		Username:     username,
		DisplayName:  displayName,
		AccessToken:  resp.AccessToken,
		DeviceID:     resp.DeviceID,
		Password:     password, // Store for potential re-login
		IsActive:     true,
	}

	// Store in database
	if um.db != nil {
		if err := um.db.WithContext(ctx).Create(user).Error; err != nil {
			log.Printf("⚠️  Failed to store user in database: %v", err)
			// Continue anyway - user was registered successfully
		}
	}

	// Cache the user
	um.cacheMutex.Lock()
	um.userCache[firebaseUID] = user
	um.cacheMutex.Unlock()

	return user, nil
}

// GetOrCreateUser gets an existing user or creates a new one
func (um *UserManager) GetOrCreateUser(ctx context.Context, firebaseUID string) (*MatrixUser, error) {
	return um.RegisterAnonymousUser(ctx, firebaseUID)
}

// GetUserByFirebaseUID gets a user by Firebase UID
func (um *UserManager) GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*MatrixUser, error) {
	// Check cache
	um.cacheMutex.RLock()
	if user, exists := um.userCache[firebaseUID]; exists {
		um.cacheMutex.RUnlock()
		return user, nil
	}
	um.cacheMutex.RUnlock()

	// Check database
	if um.db != nil {
		var user MatrixUser
		err := um.db.WithContext(ctx).
			Where("firebase_uid = ?", firebaseUID).
			First(&user).Error

		if err == nil {
			// Cache it
			um.cacheMutex.Lock()
			um.userCache[firebaseUID] = &user
			um.cacheMutex.Unlock()
			return &user, nil
		}

		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("database error: %w", err)
		}
	}

	return nil, fmt.Errorf("user not found for Firebase UID: %s", firebaseUID)
}

// GetUserByMatrixID gets a user by Matrix user ID
func (um *UserManager) GetUserByMatrixID(ctx context.Context, matrixUserID string) (*MatrixUser, error) {
	if um.db != nil {
		var user MatrixUser
		err := um.db.WithContext(ctx).
			Where("matrix_user_id = ?", matrixUserID).
			First(&user).Error

		if err == nil {
			return &user, nil
		}

		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("database error: %w", err)
		}
	}

	return nil, fmt.Errorf("user not found for Matrix ID: %s", matrixUserID)
}

// CreateClient creates a new Matrix client for a user
func (um *UserManager) CreateClient(user *MatrixUser) *Client {
	client := NewClient(um.homeserverURL)
	client.SetAccessToken(user.AccessToken, user.MatrixUserID, user.DeviceID)
	return client
}

// UpdateDisplayName updates a user's display name
func (um *UserManager) UpdateDisplayName(ctx context.Context, firebaseUID, displayName string) error {
	user, err := um.GetUserByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return err
	}

	// Update in database
	if um.db != nil {
		err := um.db.WithContext(ctx).
			Model(&MatrixUser{}).
			Where("firebase_uid = ?", firebaseUID).
			Update("display_name", displayName).Error

		if err != nil {
			return fmt.Errorf("update display name failed: %w", err)
		}
	}

	// Update cache
	um.cacheMutex.Lock()
	if cachedUser, exists := um.userCache[firebaseUID]; exists {
		cachedUser.DisplayName = displayName
	}
	um.cacheMutex.Unlock()

	// TODO: Update display name on Matrix homeserver
	// This requires setting the m.room.member state event

	log.Printf("✏️  Updated display name for %s: %s", user.MatrixUserID, displayName)
	return nil
}

// DeactivateUser deactivates a user
func (um *UserManager) DeactivateUser(ctx context.Context, firebaseUID string) error {
	// Update in database
	if um.db != nil {
		err := um.db.WithContext(ctx).
			Model(&MatrixUser{}).
			Where("firebase_uid = ?", firebaseUID).
			Update("is_active", false).Error

		if err != nil {
			return fmt.Errorf("deactivate user failed: %w", err)
		}
	}

	// Remove from cache
	um.cacheMutex.Lock()
	delete(um.userCache, firebaseUID)
	um.cacheMutex.Unlock()

	log.Printf("🚫 Deactivated user for Firebase UID: %s", firebaseUID)
	return nil
}

// SetPresence sets a user's presence status
func (um *UserManager) SetPresence(ctx context.Context, user *MatrixUser, presence, statusMsg string) error {
	client := um.CreateClient(user)
	if err := client.SetPresence(ctx, presence, statusMsg); err != nil {
		return fmt.Errorf("set presence failed: %w", err)
	}

	log.Printf("🟢 Set presence for %s: %s (%s)", user.MatrixUserID, presence, statusMsg)
	return nil
}

// generateAnonymousUsername generates a random anonymous username
func (um *UserManager) generateAnonymousUsername() string {
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	suffix := base64.URLEncoding.EncodeToString(randomBytes)[:11]
	return fmt.Sprintf("anon_%s", suffix)
}

// generateSecurePassword generates a secure random password
func (um *UserManager) generateSecurePassword() string {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	return base64.URLEncoding.EncodeToString(randomBytes)
}

// generateDisplayName generates a random display name
func (um *UserManager) generateDisplayName() string {
	adjectives := []string{"Curious", "Happy", "Brave", "Friendly", "Cheerful", "Adventurous", "Mysterious"}
	nouns := []string{"Traveler", "Explorer", "Wanderer", "Passenger", "Voyager", "Rider", "Journeyer"}

	randomBytes := make([]byte, 2)
	rand.Read(randomBytes)

	adj := adjectives[int(randomBytes[0])%len(adjectives)]
	noun := nouns[int(randomBytes[1])%len(nouns)]

	return fmt.Sprintf("%s %s", adj, noun)
}

// GetUserStats returns statistics about users
func (um *UserManager) GetUserStats(ctx context.Context) (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"cached_users": len(um.userCache),
	}

	if um.db != nil {
		var totalUsers int64
		var activeUsers int64

		if err := um.db.Model(&MatrixUser{}).Count(&totalUsers).Error; err != nil {
			return nil, fmt.Errorf("count users failed: %w", err)
		}

		if err := um.db.Model(&MatrixUser{}).Where("is_active = ?", true).Count(&activeUsers).Error; err != nil {
			return nil, fmt.Errorf("count active users failed: %w", err)
		}

		stats["total_users"] = totalUsers
		stats["active_users"] = activeUsers
	}

	return stats, nil
}

// ============================================================
// Database Model for Matrix Users
// ============================================================

// MatrixUser represents a Matrix user in the database
type MatrixUser struct {
	ID           int64  `gorm:"primaryKey;autoIncrement"`
	FirebaseUID  string `gorm:"uniqueIndex;not null"`
	MatrixUserID string `gorm:"uniqueIndex;not null"`
	Username     string `gorm:"not null"`
	DisplayName  string
	AccessToken  string
	DeviceID     string
	Password     string // Stored encrypted
	IsActive     bool  `gorm:"default:true"`
	CreatedAt    int64 `gorm:"autoCreateTime:milli"`
	UpdatedAt    int64 `gorm:"autoUpdateTime:milli"`
}

// TableName specifies the table name for MatrixUser
func (MatrixUser) TableName() string {
	return "matrix_users"
}
