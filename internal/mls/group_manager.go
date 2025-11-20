package mls

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// GroupManager handles MLS group operations
type GroupManager struct {
	db     *gorm.DB
	redis  *redis.Client
	config *MLSConfig
}

// NewGroupManager creates a new group manager
func NewGroupManager(db *gorm.DB, redis *redis.Client, config *MLSConfig) *GroupManager {
	if config == nil {
		config = DefaultMLSConfig()
	}

	return &GroupManager{
		db:     db,
		redis:  redis,
		config: config,
	}
}

// CreateGroup creates a new MLS group
func (gm *GroupManager) CreateGroup(
	ctx context.Context,
	creatorUserID string,
	roomID *string,
	stationID *string,
	cipherSuite uint16,
) (*MLSGroup, error) {
	// Validate cipher suite
	if !gm.isValidCipherSuite(cipherSuite) {
		return nil, fmt.Errorf("unsupported cipher suite: 0x%04x", cipherSuite)
	}

	// Generate unique group ID
	groupID := fmt.Sprintf("mls_group_%s", uuid.New().String())

	// Initialize with empty state (will be set on first commit)
	group := &MLSGroup{
		ID:              uuid.New(),
		GroupID:         groupID,
		RoomID:          roomID,
		StationID:       stationID,
		CreatorUserID:   creatorUserID,
		Epoch:           0,
		TreeHash:        []byte{}, // Empty until first commit
		ConfirmationTag: []byte{}, // Empty until first commit
		CipherSuite:     cipherSuite,
		ProtocolVersion: gm.config.ProtocolVersion,
		MemberCount:     0,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		LastEpochAt:     time.Now(),
	}

	// Insert to database
	if err := gm.db.WithContext(ctx).Create(group).Error; err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	// Initialize Redis counters
	seqKey := fmt.Sprintf("mls:group:%s:sequence", groupID)
	if err := gm.redis.Set(ctx, seqKey, 0, 0).Err(); err != nil {
		log.Printf("⚠️  Failed to initialize sequence counter: %v", err)
	}

	// Initialize epoch history
	history := &MLSEpochHistory{
		GroupID:        groupID,
		Epoch:          0,
		TreeHash:       group.TreeHash,
		ConfirmationTag: group.ConfirmationTag,
		MemberCount:    0,
		ChangeType:     "create",
		StartedAt:      time.Now(),
	}
	gm.db.WithContext(ctx).Create(history)

	log.Printf("✅ Created MLS group: %s (cipher suite: %s)",
		groupID, CipherSuiteName(cipherSuite))

	return group, nil
}

// AddMember adds a member to the group
func (gm *GroupManager) AddMember(
	ctx context.Context,
	groupID string,
	userID string,
	clientID string,
	credentialID string,
	signaturePublicKey []byte,
) error {
	// Get current group state
	group, err := gm.GetGroup(ctx, groupID)
	if err != nil {
		return err
	}

	if !group.IsActive {
		return fmt.Errorf("group is not active")
	}

	// Check member limit
	if group.MemberCount >= gm.config.MaxMembersPerGroup {
		return fmt.Errorf("group has reached maximum member count")
	}

	// Check if already a member
	exists, _ := gm.IsMember(ctx, groupID, clientID)
	if exists {
		return fmt.Errorf("client is already a member")
	}

	// Insert member record
	member := &MLSGroupMember{
		ID:                 uuid.New(),
		GroupID:            groupID,
		UserID:             userID,
		ClientID:           clientID,
		CredentialID:       credentialID,
		SignaturePublicKey: signaturePublicKey,
		JoinedAtEpoch:      group.Epoch + 1, // Will join in next epoch
		IsActive:           true,
		JoinedAt:           time.Now(),
	}

	if err := gm.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// Update member count
	if err := gm.db.WithContext(ctx).Model(&MLSGroup{}).
		Where("group_id = ?", groupID).
		Updates(map[string]interface{}{
			"member_count": gorm.Expr("member_count + ?", 1),
			"updated_at":   time.Now(),
		}).Error; err != nil {
		return err
	}

	// Update Redis cache
	membersKey := fmt.Sprintf("mls:group:%s:members", groupID)
	if err := gm.redis.SAdd(ctx, membersKey, clientID).Err(); err != nil {
		log.Printf("⚠️  Failed to update Redis members cache: %v", err)
	}
	gm.redis.Expire(ctx, membersKey, 1*time.Hour)

	log.Printf("👥 Added member %s to group %s (epoch %d → %d)",
		clientID, groupID, group.Epoch, group.Epoch+1)

	return nil
}

// RemoveMember removes a member from the group
func (gm *GroupManager) RemoveMember(
	ctx context.Context,
	groupID string,
	clientID string,
) error {
	group, err := gm.GetGroup(ctx, groupID)
	if err != nil {
		return err
	}

	// Mark as removed
	now := time.Now()
	removedAtEpoch := group.Epoch + 1

	result := gm.db.WithContext(ctx).Model(&MLSGroupMember{}).
		Where("group_id = ? AND client_id = ? AND is_active = ?", groupID, clientID, true).
		Updates(map[string]interface{}{
			"removed_at_epoch": removedAtEpoch,
			"is_active":        false,
			"last_seen_at":     now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("member not found or already removed")
	}

	// Update member count
	gm.db.WithContext(ctx).Model(&MLSGroup{}).
		Where("group_id = ?", groupID).
		Updates(map[string]interface{}{
			"member_count": gorm.Expr("member_count - ?", 1),
			"updated_at":   now,
		})

	// Update Redis
	membersKey := fmt.Sprintf("mls:group:%s:members", groupID)
	gm.redis.SRem(ctx, membersKey, clientID)

	log.Printf("👋 Removed member %s from group %s", clientID, groupID)

	return nil
}

// IncrementEpoch advances the group epoch
func (gm *GroupManager) IncrementEpoch(
	ctx context.Context,
	groupID string,
	treeHash []byte,
	confirmationTag []byte,
	changeType string,
	changedByClientID *string,
) error {
	now := time.Now()

	// Get current epoch
	var currentEpoch int64
	if err := gm.db.WithContext(ctx).Model(&MLSGroup{}).
		Where("group_id = ?", groupID).
		Pluck("epoch", &currentEpoch).Error; err != nil {
		return err
	}

	newEpoch := currentEpoch + 1

	// Update group
	if err := gm.db.WithContext(ctx).Model(&MLSGroup{}).
		Where("group_id = ?", groupID).
		Updates(map[string]interface{}{
			"epoch":            newEpoch,
			"tree_hash":        treeHash,
			"confirmation_tag": confirmationTag,
			"last_epoch_at":    now,
			"updated_at":       now,
		}).Error; err != nil {
		return err
	}

	// Close previous epoch history
	if currentEpoch > 0 {
		gm.db.WithContext(ctx).Model(&MLSEpochHistory{}).
			Where("group_id = ? AND epoch = ? AND ended_at IS NULL", groupID, currentEpoch).
			Update("ended_at", now)
	}

	// Get member count
	var memberCount int
	gm.db.WithContext(ctx).Model(&MLSGroup{}).
		Where("group_id = ?", groupID).
		Pluck("member_count", &memberCount)

	// Create new epoch history
	history := &MLSEpochHistory{
		GroupID:           groupID,
		Epoch:             newEpoch,
		TreeHash:          treeHash,
		ConfirmationTag:   confirmationTag,
		MemberCount:       memberCount,
		ChangeType:        changeType,
		ChangedByClientID: changedByClientID,
		StartedAt:         now,
	}
	gm.db.WithContext(ctx).Create(history)

	log.Printf("📈 Group %s advanced to epoch %d (change: %s)",
		groupID, newEpoch, changeType)

	return nil
}

// GetGroup retrieves group by ID
func (gm *GroupManager) GetGroup(ctx context.Context, groupID string) (*MLSGroup, error) {
	var group MLSGroup

	if err := gm.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("group not found: %s", groupID)
		}
		return nil, err
	}

	return &group, nil
}

// GetActiveMembers retrieves all active members
func (gm *GroupManager) GetActiveMembers(ctx context.Context, groupID string) ([]*MLSGroupMember, error) {
	var members []*MLSGroupMember

	if err := gm.db.WithContext(ctx).
		Where("group_id = ? AND is_active = ?", groupID, true).
		Order("joined_at ASC").
		Find(&members).Error; err != nil {
		return nil, err
	}

	return members, nil
}

// GetActiveMemberCount returns the count of active members
func (gm *GroupManager) GetActiveMemberCount(ctx context.Context, groupID string) (int, error) {
	var count int64

	if err := gm.db.WithContext(ctx).Model(&MLSGroupMember{}).
		Where("group_id = ? AND is_active = ?", groupID, true).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

// IsMember checks if client is a member (active or removed)
func (gm *GroupManager) IsMember(ctx context.Context, groupID string, clientID string) (bool, error) {
	var count int64

	if err := gm.db.WithContext(ctx).Model(&MLSGroupMember{}).
		Where("group_id = ? AND client_id = ?", groupID, clientID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// IsActiveMember checks if client is an active member
func (gm *GroupManager) IsActiveMember(ctx context.Context, groupID string, clientID string) (bool, error) {
	var count int64

	if err := gm.db.WithContext(ctx).Model(&MLSGroupMember{}).
		Where("group_id = ? AND client_id = ? AND is_active = ?", groupID, clientID, true).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeactivateGroup deactivates a group
func (gm *GroupManager) DeactivateGroup(ctx context.Context, groupID string) error {
	if err := gm.db.WithContext(ctx).Model(&MLSGroup{}).
		Where("group_id = ?", groupID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return err
	}

	log.Printf("🚫 Deactivated group %s", groupID)
	return nil
}

// GetGroupStats retrieves group statistics
func (gm *GroupManager) GetGroupStats(ctx context.Context, groupID string) (*GroupStatsResponse, error) {
	group, err := gm.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Get message count
	var messageCount int64
	gm.db.WithContext(ctx).Model(&MLSMessage{}).
		Where("group_id = ?", groupID).
		Count(&messageCount)

	return &GroupStatsResponse{
		GroupID:      group.GroupID,
		MemberCount:  group.MemberCount,
		MessageCount: messageCount,
		Epoch:        group.Epoch,
		LastActive:   group.LastEpochAt,
	}, nil
}

// ListGroupsByStation retrieves all groups for a station
func (gm *GroupManager) ListGroupsByStation(ctx context.Context, stationID string) ([]*MLSGroup, error) {
	var groups []*MLSGroup

	if err := gm.db.WithContext(ctx).
		Where("station_id = ? AND is_active = ?", stationID, true).
		Order("created_at DESC").
		Find(&groups).Error; err != nil {
		return nil, err
	}

	return groups, nil
}

// isValidCipherSuite checks if a cipher suite is supported
func (gm *GroupManager) isValidCipherSuite(suite uint16) bool {
	for _, s := range gm.config.CipherSuites {
		if s == suite {
			return true
		}
	}
	return false
}
