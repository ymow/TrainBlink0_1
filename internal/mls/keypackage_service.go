package mls

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// KeyPackageService handles KeyPackage operations
type KeyPackageService struct {
	db     *gorm.DB
	redis  *redis.Client
	config *MLSConfig
}

// NewKeyPackageService creates a new KeyPackage service
func NewKeyPackageService(db *gorm.DB, redis *redis.Client, config *MLSConfig) *KeyPackageService {
	if config == nil {
		config = DefaultMLSConfig()
	}

	return &KeyPackageService{
		db:     db,
		redis:  redis,
		config: config,
	}
}

// UploadKeyPackage stores a KeyPackage for a user
func (kps *KeyPackageService) UploadKeyPackage(
	ctx context.Context,
	userID string,
	clientID string,
	credentialID string,
	keyPackageData []byte,
	cipherSuite uint16,
	clientIP string,
) (*MLSKeyPackage, error) {
	// 1. Rate limiting check
	allowed, err := kps.checkRateLimit(ctx, userID)
	if err != nil || !allowed {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// 2. Check user quota
	count, err := kps.GetAvailableCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	if count >= kps.config.KeyPackageConfig.MaxPerUser {
		return nil, fmt.Errorf("KeyPackage quota exceeded (max: %d)", kps.config.KeyPackageConfig.MaxPerUser)
	}

	// 3. Generate KeyPackage ID (SHA256 hash of data)
	hash := sha256.Sum256(keyPackageData)
	keyPackageID := base64.URLEncoding.EncodeToString(hash[:])

	// 4. Create KeyPackage record
	kp := &MLSKeyPackage{
		ID:             uuid.New(),
		KeyPackageID:   keyPackageID,
		UserID:         userID,
		ClientID:       clientID,
		CredentialID:   credentialID,
		KeyPackageData: keyPackageData,
		CipherSuite:    cipherSuite,
		CreatedAt:      time.Now(),
		IsConsumed:     false,
		CreatedByIP:    clientIP,
	}

	// 5. Insert to database
	if err := kps.db.WithContext(ctx).Create(kp).Error; err != nil {
		return nil, fmt.Errorf("failed to upload KeyPackage: %w", err)
	}

	// 6. Add to Redis cache
	cacheKey := fmt.Sprintf("keypackage:%s:available", userID)
	if err := kps.redis.SAdd(ctx, cacheKey, keyPackageID).Err(); err != nil {
		log.Printf("⚠️  Failed to cache KeyPackage: %v", err)
	}
	kps.redis.Expire(ctx, cacheKey, time.Duration(kps.config.KeyPackageConfig.TTLDays)*24*time.Hour)

	log.Printf("📦 Uploaded KeyPackage for user %s (client: %s, suite: %s)",
		userID, clientID, CipherSuiteName(cipherSuite))

	return kp, nil
}

// ClaimKeyPackage retrieves and marks a KeyPackage as consumed
func (kps *KeyPackageService) ClaimKeyPackage(
	ctx context.Context,
	userID string,
	groupID string,
	cipherSuite uint16,
) (*MLSKeyPackage, error) {
	// 1. Start transaction
	tx := kps.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 2. Find available KeyPackage with row lock
	var kp MLSKeyPackage
	err := tx.Where("user_id = ? AND cipher_suite = ? AND is_consumed = ?", userID, cipherSuite, false).
		Order("created_at ASC").
		Limit(1).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		First(&kp).Error

	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no available KeyPackage for user %s (cipher suite: %s)",
				userID, CipherSuiteName(cipherSuite))
		}
		return nil, fmt.Errorf("failed to find KeyPackage: %w", err)
	}

	// 3. Mark as consumed
	now := time.Now()
	if err := tx.Model(&MLSKeyPackage{}).
		Where("id = ?", kp.ID).
		Updates(map[string]interface{}{
			"consumed_at":          now,
			"consumed_by_group_id": groupID,
			"is_consumed":          true,
		}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 4. Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// 5. Update Redis cache
	cacheKey := fmt.Sprintf("keypackage:%s:available", userID)
	kps.redis.SRem(ctx, cacheKey, kp.KeyPackageID)

	kp.ConsumedAt = &now
	kp.ConsumedByGroupID = &groupID
	kp.IsConsumed = true

	log.Printf("🔑 Claimed KeyPackage %s for group %s", kp.KeyPackageID[:16], groupID)

	return &kp, nil
}

// GetAvailableKeyPackages returns available KeyPackages for a user
func (kps *KeyPackageService) GetAvailableKeyPackages(
	ctx context.Context,
	userID string,
	cipherSuite uint16,
	limit int,
) ([]*MLSKeyPackage, error) {
	var keyPackages []*MLSKeyPackage

	if err := kps.db.WithContext(ctx).
		Where("user_id = ? AND cipher_suite = ? AND is_consumed = ?", userID, cipherSuite, false).
		Order("created_at ASC").
		Limit(limit).
		Find(&keyPackages).Error; err != nil {
		return nil, err
	}

	return keyPackages, nil
}

// GetAvailableCount returns count of available KeyPackages
func (kps *KeyPackageService) GetAvailableCount(ctx context.Context, userID string) (int, error) {
	var count int64

	if err := kps.db.WithContext(ctx).Model(&MLSKeyPackage{}).
		Where("user_id = ? AND is_consumed = ?", userID, false).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

// DeleteExpiredKeyPackages removes old unconsumed KeyPackages
func (kps *KeyPackageService) DeleteExpiredKeyPackages(ctx context.Context) (int64, error) {
	expiryDays := kps.config.KeyPackageConfig.TTLDays
	if expiryDays <= 0 {
		expiryDays = 30
	}

	expiredTime := time.Now().AddDate(0, 0, -expiryDays)

	result := kps.db.WithContext(ctx).
		Where("is_consumed = ? AND created_at < ?", false, expiredTime).
		Delete(&MLSKeyPackage{})

	if result.Error != nil {
		return 0, result.Error
	}

	if result.RowsAffected > 0 {
		log.Printf("🗑️  Deleted %d expired KeyPackages", result.RowsAffected)
	}

	return result.RowsAffected, nil
}

// DeleteConsumedKeyPackages removes old consumed KeyPackages (cleanup)
func (kps *KeyPackageService) DeleteConsumedKeyPackages(ctx context.Context, olderThanDays int) (int64, error) {
	if olderThanDays <= 0 {
		olderThanDays = 90 // Default 90 days
	}

	expiredTime := time.Now().AddDate(0, 0, -olderThanDays)

	result := kps.db.WithContext(ctx).
		Where("is_consumed = ? AND consumed_at < ?", true, expiredTime).
		Delete(&MLSKeyPackage{})

	if result.Error != nil {
		return 0, result.Error
	}

	if result.RowsAffected > 0 {
		log.Printf("🗑️  Deleted %d old consumed KeyPackages", result.RowsAffected)
	}

	return result.RowsAffected, nil
}

// checkRateLimit checks if user can upload more KeyPackages
func (kps *KeyPackageService) checkRateLimit(ctx context.Context, userID string) (bool, error) {
	// Rate limit: configured uploads per hour per user
	window := time.Now().Format("2006-01-02-15")
	key := fmt.Sprintf("ratelimit:keypackage:%s:%s", userID, window)

	count, err := kps.redis.Incr(ctx, key).Result()
	if err != nil {
		log.Printf("⚠️  Rate limit check failed: %v", err)
		return true, nil // Allow on Redis errors
	}

	if count == 1 {
		kps.redis.Expire(ctx, key, time.Hour)
	}

	limit := int64(kps.config.KeyPackageConfig.UploadRateLimit)
	if count > limit {
		log.Printf("⚠️  User %s exceeded KeyPackage upload rate limit (%d/%d)",
			userID, count, limit)
		return false, nil
	}

	return true, nil
}

// GetStats returns KeyPackage statistics
func (kps *KeyPackageService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	var totalCount, availableCount, consumedCount int64

	kps.db.WithContext(ctx).Model(&MLSKeyPackage{}).Count(&totalCount)
	kps.db.WithContext(ctx).Model(&MLSKeyPackage{}).Where("is_consumed = ?", false).Count(&availableCount)
	kps.db.WithContext(ctx).Model(&MLSKeyPackage{}).Where("is_consumed = ?", true).Count(&consumedCount)

	return map[string]interface{}{
		"total_keypackages":     totalCount,
		"available_keypackages": availableCount,
		"consumed_keypackages":  consumedCount,
	}, nil
}
