package matrix

import (
	"context"
	"fmt"
	"log"
	"sync"

	"gorm.io/gorm"
)

// RoomManager handles Matrix room operations
type RoomManager struct {
	client *Client
	db     *gorm.DB

	// In-memory cache of station rooms
	stationRooms map[string]string // stationID -> roomID
	roomsMutex   sync.RWMutex
}

// NewRoomManager creates a new room manager
func NewRoomManager(client *Client, db *gorm.DB) *RoomManager {
	return &RoomManager{
		client:       client,
		db:           db,
		stationRooms: make(map[string]string),
	}
}

// GetOrCreateStationRoom gets or creates a room for a station
func (rm *RoomManager) GetOrCreateStationRoom(ctx context.Context, stationID, stationName string) (string, error) {
	// Check cache first
	rm.roomsMutex.RLock()
	roomID, exists := rm.stationRooms[stationID]
	rm.roomsMutex.RUnlock()

	if exists {
		log.Printf("🏠 Using existing room for station %s: %s", stationID, roomID)
		return roomID, nil
	}

	// Check database
	if rm.db != nil {
		var room MatrixRoom
		err := rm.db.WithContext(ctx).
			Where("station_id = ? AND is_active = ?", stationID, true).
			First(&room).Error

		if err == nil {
			// Found in database, cache it
			rm.roomsMutex.Lock()
			rm.stationRooms[stationID] = room.RoomID
			rm.roomsMutex.Unlock()

			log.Printf("🏠 Loaded room for station %s from database: %s", stationID, room.RoomID)
			return room.RoomID, nil
		} else if err != gorm.ErrRecordNotFound {
			return "", fmt.Errorf("database error: %w", err)
		}
	}

	// Create new room
	log.Printf("🏗️  Creating new room for station %s", stationID)
	return rm.createStationRoom(ctx, stationID, stationName)
}

// createStationRoom creates a new Matrix room for a station
func (rm *RoomManager) createStationRoom(ctx context.Context, stationID, stationName string) (string, error) {
	roomAlias := stationID // Use station ID as room alias

	req := &CreateRoomRequest{
		RoomAliasName: roomAlias,
		Name:          fmt.Sprintf("%s - TrainBlink", stationName),
		Topic:         fmt.Sprintf("Chat with travelers at %s station", stationName),
		Visibility:    VisibilityPublic,
		Preset:        PresetPublicChat,
		InitialState: []StateEvent{
			{
				Type:     "m.room.guest_access",
				StateKey: "",
				Content: map[string]interface{}{
					"guest_access": "can_join",
				},
			},
			{
				Type:     "m.room.history_visibility",
				StateKey: "",
				Content: map[string]interface{}{
					"history_visibility": "shared",
				},
			},
		},
		PowerLevelContentOverride: map[string]interface{}{
			"users": map[string]int{
				rm.client.GetUserID(): 100, // Bot has admin power
			},
			"events": map[string]int{
				"m.room.name":         50,
				"m.room.power_levels": 100,
				"m.room.encryption":   100,
			},
			"events_default": 0,
			"state_default":  50,
			"ban":            50,
			"kick":           50,
			"redact":         50,
			"invite":         0,
		},
	}

	resp, err := rm.client.CreateRoom(ctx, req)
	if err != nil {
		return "", fmt.Errorf("create room failed: %w", err)
	}

	log.Printf("✅ Created room %s for station %s", resp.RoomID, stationID)

	// Store in database
	if rm.db != nil {
		room := &MatrixRoom{
			RoomID:    resp.RoomID,
			RoomAlias: fmt.Sprintf("#%s:%s", roomAlias, extractServerName(rm.client.GetHomeserverURL())),
			StationID: stationID,
			Name:      req.Name,
			Topic:     req.Topic,
			IsActive:  true,
		}

		if err := rm.db.WithContext(ctx).Create(room).Error; err != nil {
			log.Printf("⚠️  Failed to store room in database: %v", err)
			// Continue anyway - room was created successfully
		}
	}

	// Cache the room
	rm.roomsMutex.Lock()
	rm.stationRooms[stationID] = resp.RoomID
	rm.roomsMutex.Unlock()

	return resp.RoomID, nil
}

// JoinRoom joins a user to a room
func (rm *RoomManager) JoinRoom(ctx context.Context, roomID string) error {
	resp, err := rm.client.JoinRoomByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("join room failed: %w", err)
	}

	log.Printf("✅ Joined room: %s", resp.RoomID)
	return nil
}

// LeaveRoom leaves a room
func (rm *RoomManager) LeaveRoom(ctx context.Context, roomID string) error {
	if err := rm.client.LeaveRoom(ctx, roomID); err != nil {
		return fmt.Errorf("leave room failed: %w", err)
	}

	log.Printf("👋 Left room: %s", roomID)
	return nil
}

// SendMessage sends a text message to a room
func (rm *RoomManager) SendMessage(ctx context.Context, roomID, body string) (string, error) {
	resp, err := rm.client.SendTextMessage(ctx, roomID, body)
	if err != nil {
		return "", fmt.Errorf("send message failed: %w", err)
	}

	return resp.EventID, nil
}

// SendHTMLMessage sends an HTML message to a room
func (rm *RoomManager) SendHTMLMessage(ctx context.Context, roomID, body, htmlBody string) (string, error) {
	resp, err := rm.client.SendHTMLMessage(ctx, roomID, body, htmlBody)
	if err != nil {
		return "", fmt.Errorf("send HTML message failed: %w", err)
	}

	return resp.EventID, nil
}

// SendImageMessage sends an image to a room
func (rm *RoomManager) SendImageMessage(ctx context.Context, roomID, body, mxcURL string, info *MediaInfo) (string, error) {
	resp, err := rm.client.SendImageMessage(ctx, roomID, body, mxcURL, info)
	if err != nil {
		return "", fmt.Errorf("send image failed: %w", err)
	}

	return resp.EventID, nil
}

// UploadAndSendImage uploads an image and sends it to a room
func (rm *RoomManager) UploadAndSendImage(ctx context.Context, roomID, filename string, data []byte, contentType string) (string, error) {
	// Upload to Matrix media repository
	uploadResp, err := rm.client.UploadMedia(ctx, data, contentType, filename)
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	// Send image message
	info := &MediaInfo{
		MimeType: contentType,
		Size:     int64(len(data)),
	}

	resp, err := rm.client.SendImageMessage(ctx, roomID, filename, uploadResp.ContentURI, info)
	if err != nil {
		return "", fmt.Errorf("send image failed: %w", err)
	}

	log.Printf("📷 Uploaded and sent image to %s: %s (event: %s)",
		roomID, uploadResp.ContentURI, resp.EventID)

	return resp.EventID, nil
}

// GetRoomMembers gets all members in a room
func (rm *RoomManager) GetRoomMembers(ctx context.Context, roomID string) ([]Event, error) {
	members, err := rm.client.GetRoomMembers(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("get members failed: %w", err)
	}

	return members, nil
}

// GetRoomMemberCount gets the count of members in a room
func (rm *RoomManager) GetRoomMemberCount(ctx context.Context, roomID string) (int, error) {
	members, err := rm.GetRoomMembers(ctx, roomID)
	if err != nil {
		return 0, err
	}

	// Count only joined members
	count := 0
	for _, member := range members {
		if membership, ok := member.Content["membership"].(string); ok {
			if membership == MembershipJoin {
				count++
			}
		}
	}

	return count, nil
}

// GetRoomState gets the state of a room
func (rm *RoomManager) GetRoomState(ctx context.Context, roomID string) ([]Event, error) {
	state, err := rm.client.GetRoomState(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("get state failed: %w", err)
	}

	return state, nil
}

// SetRoomName sets the name of a room
func (rm *RoomManager) SetRoomName(ctx context.Context, roomID, name string) error {
	content := map[string]interface{}{"name": name}
	return rm.client.SendStateEvent(ctx, roomID, EventTypeRoomName, "", content)
}

// SetRoomTopic sets the topic of a room
func (rm *RoomManager) SetRoomTopic(ctx context.Context, roomID, topic string) error {
	content := map[string]interface{}{"topic": topic}
	return rm.client.SendStateEvent(ctx, roomID, EventTypeRoomTopic, "", content)
}

// SetTyping sets the typing indicator
func (rm *RoomManager) SetTyping(ctx context.Context, roomID string, typing bool, timeout int64) error {
	return rm.client.SetTyping(ctx, roomID, typing, timeout)
}

// SendReadReceipt sends a read receipt
func (rm *RoomManager) SendReadReceipt(ctx context.Context, roomID, eventID string) error {
	return rm.client.SendReadReceipt(ctx, roomID, eventID)
}

// InviteUser invites a user to a room
func (rm *RoomManager) InviteUser(ctx context.Context, roomID, userID string) error {
	if err := rm.client.InviteUser(ctx, roomID, userID); err != nil {
		return fmt.Errorf("invite failed: %w", err)
	}

	log.Printf("📨 Invited %s to room %s", userID, roomID)
	return nil
}

// KickUser kicks a user from a room
func (rm *RoomManager) KickUser(ctx context.Context, roomID, userID, reason string) error {
	if err := rm.client.KickUser(ctx, roomID, userID, reason); err != nil {
		return fmt.Errorf("kick failed: %w", err)
	}

	log.Printf("🚪 Kicked %s from room %s (reason: %s)", userID, roomID, reason)
	return nil
}

// BanUser bans a user from a room
func (rm *RoomManager) BanUser(ctx context.Context, roomID, userID, reason string) error {
	if err := rm.client.BanUser(ctx, roomID, userID, reason); err != nil {
		return fmt.Errorf("ban failed: %w", err)
	}

	log.Printf("🚫 Banned %s from room %s (reason: %s)", userID, roomID, reason)
	return nil
}

// extractServerName extracts server name from homeserver URL
func extractServerName(homeserverURL string) string {
	// Simple extraction - in production, parse URL properly
	// For now, return a default
	return "trainblink.org"
}

// ============================================================
// Database Model for Matrix Rooms
// ============================================================

// MatrixRoom represents a Matrix room in the database
type MatrixRoom struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	RoomID    string `gorm:"uniqueIndex;not null"`
	RoomAlias string `gorm:"index"`
	StationID string `gorm:"index"`
	Name      string
	Topic     string
	IsActive  bool  `gorm:"default:true"`
	CreatedAt int64 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli"`
}

// TableName specifies the table name for MatrixRoom
func (MatrixRoom) TableName() string {
	return "matrix_rooms"
}
