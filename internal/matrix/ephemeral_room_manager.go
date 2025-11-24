package matrix

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/ymow/messenger_protocol_research/internal/model"
	"gorm.io/gorm"
)

// EphemeralRoomManager handles ephemeral 1-on-1 DM room operations
type EphemeralRoomManager struct {
	client       *Client
	db           *gorm.DB
	roomLifetime time.Duration // Default 24 hours
}

// NewEphemeralRoomManager creates a new ephemeral room manager
func NewEphemeralRoomManager(client *Client, db *gorm.DB) *EphemeralRoomManager {
	return &EphemeralRoomManager{
		client:       client,
		db:           db,
		roomLifetime: 24 * time.Hour, // Default 24 hours
	}
}

// SetRoomLifetime sets the default room lifetime
func (erm *EphemeralRoomManager) SetRoomLifetime(lifetime time.Duration) {
	erm.roomLifetime = lifetime
}

// CreateEphemeralDMRequest contains parameters for creating ephemeral DM
type CreateEphemeralDMRequest struct {
	Trip1ID      uuid.UUID `json:"trip1_id"`
	Trip2ID      uuid.UUID `json:"trip2_id"`
	AnonymousID1 string    `json:"anonymous_id_1"`
	AnonymousID2 string    `json:"anonymous_id_2"`
	MLSGroupID   *string   `json:"mls_group_id,omitempty"`
}

// CreateEphemeralDM creates a new ephemeral 1-on-1 DM room
func (erm *EphemeralRoomManager) CreateEphemeralDM(ctx context.Context, req *CreateEphemeralDMRequest) (*model.MatrixEphemeralRoom, error) {
	// Check if room already exists between these two trips
	var existingRoom model.MatrixEphemeralRoom
	err := erm.db.WithContext(ctx).
		Where("(trip1_id = ? AND trip2_id = ?) OR (trip1_id = ? AND trip2_id = ?)",
			req.Trip1ID, req.Trip2ID, req.Trip2ID, req.Trip1ID).
		Where("deleted_at IS NULL").
		First(&existingRoom).Error

	if err == nil {
		// Room already exists and not deleted
		log.Printf("🏠 Using existing ephemeral DM: %s", existingRoom.RoomID)
		return &existingRoom, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Create new Matrix room
	roomReq := &CreateRoomRequest{
		Name:       fmt.Sprintf("TrainBlink DM (%s ↔ %s)", req.AnonymousID1, req.AnonymousID2),
		Visibility: VisibilityPrivate,
		Preset:     PresetTrustedPrivateChat,
		InitialState: []StateEvent{
			{
				Type:     "m.room.history_visibility",
				StateKey: "",
				Content: map[string]interface{}{
					"history_visibility": "shared",
				},
			},
			{
				Type:     "m.room.encryption",
				StateKey: "",
				Content: map[string]interface{}{
					"algorithm": "m.megolm.v1.aes-sha2",
				},
			},
		},
	}

	// Add MLS encryption if group ID provided
	if req.MLSGroupID != nil && *req.MLSGroupID != "" {
		roomReq.InitialState = append(roomReq.InitialState, StateEvent{
			Type:     "org.trainblink.mls_group",
			StateKey: "",
			Content: map[string]interface{}{
				"group_id": *req.MLSGroupID,
			},
		})
	}

	resp, err := erm.client.CreateRoom(ctx, roomReq)
	if err != nil {
		return nil, fmt.Errorf("create ephemeral room failed: %w", err)
	}

	log.Printf("✅ Created ephemeral DM room: %s", resp.RoomID)

	// Store in database
	expiresAt := time.Now().Add(erm.roomLifetime)
	ephemeralRoom := &model.MatrixEphemeralRoom{
		RoomID:       resp.RoomID,
		Trip1ID:      req.Trip1ID,
		Trip2ID:      req.Trip2ID,
		AnonymousID1: req.AnonymousID1,
		AnonymousID2: req.AnonymousID2,
		MLSGroupID:   req.MLSGroupID,
		ExpiresAt:    expiresAt,
	}

	if err := erm.db.WithContext(ctx).Create(ephemeralRoom).Error; err != nil {
		return nil, fmt.Errorf("failed to store ephemeral room: %w", err)
	}

	log.Printf("⏰ Ephemeral room will expire at: %s", expiresAt.Format(time.RFC3339))

	return ephemeralRoom, nil
}

// GetEphemeralRoomByID retrieves an ephemeral room by ID
func (erm *EphemeralRoomManager) GetEphemeralRoomByID(ctx context.Context, roomID uuid.UUID) (*model.MatrixEphemeralRoom, error) {
	var room model.MatrixEphemeralRoom
	err := erm.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", roomID).
		First(&room).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("ephemeral room not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ephemeral room: %w", err)
	}

	return &room, nil
}

// GetEphemeralRoomByMatrixRoomID retrieves an ephemeral room by Matrix room ID
func (erm *EphemeralRoomManager) GetEphemeralRoomByMatrixRoomID(ctx context.Context, matrixRoomID string) (*model.MatrixEphemeralRoom, error) {
	var room model.MatrixEphemeralRoom
	err := erm.db.WithContext(ctx).
		Where("room_id = ? AND deleted_at IS NULL", matrixRoomID).
		First(&room).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("ephemeral room not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ephemeral room: %w", err)
	}

	return &room, nil
}

// GetEphemeralRoomsBetweenTrips gets ephemeral room between two trips
func (erm *EphemeralRoomManager) GetEphemeralRoomsBetweenTrips(ctx context.Context, trip1ID, trip2ID uuid.UUID) (*model.MatrixEphemeralRoom, error) {
	var room model.MatrixEphemeralRoom
	err := erm.db.WithContext(ctx).
		Where("((trip1_id = ? AND trip2_id = ?) OR (trip1_id = ? AND trip2_id = ?)) AND deleted_at IS NULL",
			trip1ID, trip2ID, trip2ID, trip1ID).
		First(&room).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("no ephemeral room found between trips")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ephemeral room: %w", err)
	}

	return &room, nil
}

// GetActiveEphemeralRoomsForTrip gets all active ephemeral rooms for a trip
func (erm *EphemeralRoomManager) GetActiveEphemeralRoomsForTrip(ctx context.Context, tripID uuid.UUID) ([]model.MatrixEphemeralRoom, error) {
	var rooms []model.MatrixEphemeralRoom
	now := time.Now()
	err := erm.db.WithContext(ctx).
		Where("(trip1_id = ? OR trip2_id = ?) AND deleted_at IS NULL AND expires_at > ?", tripID, tripID, now).
		Find(&rooms).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch active ephemeral rooms: %w", err)
	}

	return rooms, nil
}

// IncrementMessageCount increments the message count for a room
func (erm *EphemeralRoomManager) IncrementMessageCount(ctx context.Context, roomID string) error {
	now := time.Now()
	err := erm.db.WithContext(ctx).
		Model(&model.MatrixEphemeralRoom{}).
		Where("room_id = ?", roomID).
		Updates(map[string]interface{}{
			"message_count":   gorm.Expr("message_count + 1"),
			"last_message_at": &now,
		}).Error

	if err != nil {
		return fmt.Errorf("failed to increment message count: %w", err)
	}

	return nil
}

// DeleteExpiredRooms finds and deletes expired ephemeral rooms
func (erm *EphemeralRoomManager) DeleteExpiredRooms(ctx context.Context) (int, error) {
	// Find expired rooms that haven't been deleted yet
	var expiredRooms []model.MatrixEphemeralRoom
	now := time.Now()
	err := erm.db.WithContext(ctx).
		Where("expires_at < ? AND deleted_at IS NULL AND auto_delete_queued = false", now).
		Find(&expiredRooms).Error

	if err != nil {
		return 0, fmt.Errorf("failed to find expired rooms: %w", err)
	}

	if len(expiredRooms) == 0 {
		return 0, nil
	}

	log.Printf("🧹 Found %d expired ephemeral rooms to delete", len(expiredRooms))

	deletedCount := 0
	for _, room := range expiredRooms {
		// Mark as queued for deletion
		if err := erm.db.WithContext(ctx).
			Model(&room).
			Update("auto_delete_queued", true).Error; err != nil {
			log.Printf("⚠️  Failed to mark room %s as queued: %v", room.RoomID, err)
			continue
		}

		// Try to delete from Matrix (best effort)
		// Note: This requires admin API or bot to be room admin
		// For MVP, we just mark as deleted in our database
		// TODO: Implement actual Matrix room deletion via admin API

		// Mark as deleted in database
		now := time.Now()
		if err := erm.db.WithContext(ctx).
			Model(&room).
			Update("deleted_at", &now).Error; err != nil {
			log.Printf("⚠️  Failed to mark room %s as deleted: %v", room.RoomID, err)
			continue
		}

		log.Printf("🗑️  Deleted expired ephemeral room: %s (expired at %s)",
			room.RoomID, room.ExpiresAt.Format(time.RFC3339))
		deletedCount++
	}

	return deletedCount, nil
}

// ExtendRoomLifetime extends the lifetime of an ephemeral room
func (erm *EphemeralRoomManager) ExtendRoomLifetime(ctx context.Context, roomID uuid.UUID, extension time.Duration) error {
	var room model.MatrixEphemeralRoom
	if err := erm.db.WithContext(ctx).First(&room, roomID).Error; err != nil {
		return fmt.Errorf("room not found: %w", err)
	}

	newExpiresAt := room.ExpiresAt.Add(extension)
	if err := erm.db.WithContext(ctx).
		Model(&room).
		Update("expires_at", newExpiresAt).Error; err != nil {
		return fmt.Errorf("failed to extend room lifetime: %w", err)
	}

	log.Printf("⏰ Extended room %s lifetime to %s", room.RoomID, newExpiresAt.Format(time.RFC3339))
	return nil
}

// MarkRoomForDeletion manually marks a room for deletion
func (erm *EphemeralRoomManager) MarkRoomForDeletion(ctx context.Context, roomID uuid.UUID) error {
	now := time.Now()
	err := erm.db.WithContext(ctx).
		Model(&model.MatrixEphemeralRoom{}).
		Where("id = ?", roomID).
		Updates(map[string]interface{}{
			"deleted_at":         &now,
			"auto_delete_queued": true,
		}).Error

	if err != nil {
		return fmt.Errorf("failed to mark room for deletion: %w", err)
	}

	log.Printf("🗑️  Marked ephemeral room for deletion: %s", roomID)
	return nil
}

// GetEphemeralRoomStats returns statistics about ephemeral rooms
func (erm *EphemeralRoomManager) GetEphemeralRoomStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total ephemeral rooms
	var totalRooms int64
	if err := erm.db.WithContext(ctx).
		Model(&model.MatrixEphemeralRoom{}).
		Count(&totalRooms).Error; err != nil {
		return nil, err
	}
	stats["total_rooms"] = totalRooms

	// Active rooms (not expired, not deleted)
	var activeRooms int64
	now := time.Now()
	if err := erm.db.WithContext(ctx).
		Model(&model.MatrixEphemeralRoom{}).
		Where("expires_at > ? AND deleted_at IS NULL", now).
		Count(&activeRooms).Error; err != nil {
		return nil, err
	}
	stats["active_rooms"] = activeRooms

	// Expired rooms (not yet deleted)
	var expiredRooms int64
	if err := erm.db.WithContext(ctx).
		Model(&model.MatrixEphemeralRoom{}).
		Where("expires_at < ? AND deleted_at IS NULL", now).
		Count(&expiredRooms).Error; err != nil {
		return nil, err
	}
	stats["expired_rooms"] = expiredRooms

	// Deleted rooms
	var deletedRooms int64
	if err := erm.db.WithContext(ctx).
		Model(&model.MatrixEphemeralRoom{}).
		Where("deleted_at IS NOT NULL").
		Count(&deletedRooms).Error; err != nil {
		return nil, err
	}
	stats["deleted_rooms"] = deletedRooms

	// Total messages
	type MessageStats struct {
		TotalMessages int64 `json:"total_messages"`
	}
	var msgStats MessageStats
	if err := erm.db.WithContext(ctx).
		Model(&model.MatrixEphemeralRoom{}).
		Select("SUM(message_count) as total_messages").
		Scan(&msgStats).Error; err != nil {
		return nil, err
	}
	stats["total_messages"] = msgStats.TotalMessages

	return stats, nil
}

// GetExpiringSoonRooms gets rooms expiring within the given duration
func (erm *EphemeralRoomManager) GetExpiringSoonRooms(ctx context.Context, within time.Duration) ([]model.MatrixEphemeralRoom, error) {
	now := time.Now()
	cutoffTime := now.Add(within)

	var rooms []model.MatrixEphemeralRoom
	err := erm.db.WithContext(ctx).
		Where("expires_at < ? AND expires_at > ? AND deleted_at IS NULL", cutoffTime, now).
		Order("expires_at ASC").
		Find(&rooms).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch expiring rooms: %w", err)
	}

	return rooms, nil
}
