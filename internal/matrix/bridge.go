package matrix

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ymow/messenger_protocol_research/internal/model"
)

// BridgeService provides Matrix protocol integration
type BridgeService struct {
	homeserverURL string
	domain        string

	// In-memory storage (will be replaced with PostgreSQL in production)
	rooms       map[string]*model.MatrixRoom
	users       map[string]*model.MatrixUser
	members     map[string][]string // roomID -> []userID
	roomMutex   sync.RWMutex
	userMutex   sync.RWMutex
	memberMutex sync.RWMutex

	// Station to Room mapping
	stationRooms map[string]string // stationID -> roomID
	stationMutex sync.RWMutex
}

// NewBridgeService creates a new Matrix Bridge Service
func NewBridgeService(homeserverURL, domain string) *BridgeService {
	if domain == "" {
		domain = "trainblink.org"
	}
	if homeserverURL == "" {
		homeserverURL = "https://matrix.trainblink.org"
	}

	return &BridgeService{
		homeserverURL: homeserverURL,
		domain:        domain,
		rooms:         make(map[string]*model.MatrixRoom),
		users:         make(map[string]*model.MatrixUser),
		members:       make(map[string][]string),
		stationRooms:  make(map[string]string),
	}
}

// GenerateMatrixUserID creates an anonymous Matrix user ID
func (b *BridgeService) GenerateMatrixUserID(userID string) string {
	// Generate random anonymous ID
	randomBytes := make([]byte, 6)
	rand.Read(randomBytes)
	anonID := base64.URLEncoding.EncodeToString(randomBytes)[:8]

	return fmt.Sprintf("@anon_%s:%s", anonID, b.domain)
}

// GenerateAccessToken creates a temporary access token
func (b *BridgeService) GenerateAccessToken() string {
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	return "syt_" + base64.URLEncoding.EncodeToString(tokenBytes)
}

// GetOrCreateStationRoom gets existing room or creates new one for a station
func (b *BridgeService) GetOrCreateStationRoom(ctx context.Context, stationID, stationName string) (*model.MatrixRoomInfo, error) {
	b.stationMutex.RLock()
	roomID, exists := b.stationRooms[stationID]
	b.stationMutex.RUnlock()

	if exists {
		// Room already exists, return it
		return b.getRoomInfo(roomID)
	}

	// Create new room
	return b.createStationRoom(ctx, stationID, stationName)
}

// createStationRoom creates a new Matrix room for a station
func (b *BridgeService) createStationRoom(ctx context.Context, stationID, stationName string) (*model.MatrixRoomInfo, error) {
	b.roomMutex.Lock()
	defer b.roomMutex.Unlock()

	// Generate room ID and alias
	roomUUID := uuid.New().String()[:8]
	roomID := fmt.Sprintf("!%s:%s", roomUUID, b.domain)
	roomAlias := fmt.Sprintf("#%s:%s", stationID, b.domain)

	now := time.Now()

	room := &model.MatrixRoom{
		ID:                roomID,
		RoomAlias:         roomAlias,
		StationID:         stationID,
		Name:              fmt.Sprintf("%s TrainBlink Room", stationName),
		Topic:             fmt.Sprintf("Anonymous chat for %s station", stationName),
		EncryptionEnabled: false, // Will be enabled in Phase 5 (MLS)
		MLSGroupID:        "",
		CreatedAt:         now,
		UpdatedAt:         now,
		LastActiveAt:      now,
		MemberCount:       0,
	}

	b.rooms[roomID] = room
	b.stationRooms[stationID] = roomID
	b.members[roomID] = []string{}

	return &model.MatrixRoomInfo{
		RoomID:            roomID,
		RoomAlias:         roomAlias,
		Name:              room.Name,
		Topic:             room.Topic,
		MemberCount:       0,
		EncryptionEnabled: false,
		MLSEnabled:        false,
		MLSGroupID:        "",
	}, nil
}

// JoinRoom adds a user to a Matrix room
func (b *BridgeService) JoinRoom(ctx context.Context, matrixUserID, roomID string) (string, error) {
	b.roomMutex.RLock()
	room, exists := b.rooms[roomID]
	b.roomMutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("room not found: %s", roomID)
	}

	// Check if user already joined
	b.memberMutex.RLock()
	members := b.members[roomID]
	for _, member := range members {
		if member == matrixUserID {
			b.memberMutex.RUnlock()
			// Already joined, return existing token
			return b.getUserAccessToken(matrixUserID)
		}
	}
	b.memberMutex.RUnlock()

	// Create or update user
	accessToken := b.GenerateAccessToken()
	now := time.Now()

	b.userMutex.Lock()
	b.users[matrixUserID] = &model.MatrixUser{
		ID:          matrixUserID,
		UserID:      extractUserIDFromMatrixID(matrixUserID),
		DisplayName: generateDisplayName(matrixUserID),
		AccessToken: accessToken,
		DeviceID:    generateDeviceID(),
		CreatedAt:   now,
		ExpiresAt:   now.Add(24 * time.Hour), // Token valid for 24 hours
		IsActive:    true,
	}
	b.userMutex.Unlock()

	// Add user to room members
	b.memberMutex.Lock()
	b.members[roomID] = append(b.members[roomID], matrixUserID)
	b.memberMutex.Unlock()

	// Update room member count
	b.roomMutex.Lock()
	room.MemberCount = len(b.members[roomID])
	room.LastActiveAt = now
	b.roomMutex.Unlock()

	return accessToken, nil
}

// LeaveRoom removes a user from a Matrix room
func (b *BridgeService) LeaveRoom(ctx context.Context, matrixUserID, roomID string) error {
	b.roomMutex.RLock()
	room, exists := b.rooms[roomID]
	b.roomMutex.RUnlock()

	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	// Remove user from room members
	b.memberMutex.Lock()
	members := b.members[roomID]
	newMembers := []string{}
	for _, member := range members {
		if member != matrixUserID {
			newMembers = append(newMembers, member)
		}
	}
	b.members[roomID] = newMembers
	b.memberMutex.Unlock()

	// Update room member count
	b.roomMutex.Lock()
	room.MemberCount = len(newMembers)
	room.LastActiveAt = time.Now()
	b.roomMutex.Unlock()

	// Deactivate user
	b.userMutex.Lock()
	if user, exists := b.users[matrixUserID]; exists {
		user.IsActive = false
	}
	b.userMutex.Unlock()

	return nil
}

// GetRoomMemberCount returns the number of members in a room
func (b *BridgeService) GetRoomMemberCount(roomID string) int {
	b.memberMutex.RLock()
	defer b.memberMutex.RUnlock()

	return len(b.members[roomID])
}

// getRoomInfo retrieves room information
func (b *BridgeService) getRoomInfo(roomID string) (*model.MatrixRoomInfo, error) {
	b.roomMutex.RLock()
	defer b.roomMutex.RUnlock()

	room, exists := b.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	memberCount := b.GetRoomMemberCount(roomID)

	return &model.MatrixRoomInfo{
		RoomID:            room.ID,
		RoomAlias:         room.RoomAlias,
		Name:              room.Name,
		Topic:             room.Topic,
		MemberCount:       memberCount,
		EncryptionEnabled: room.EncryptionEnabled,
		MLSEnabled:        room.MLSGroupID != "",
		MLSGroupID:        room.MLSGroupID,
	}, nil
}

// getUserAccessToken retrieves access token for a user
func (b *BridgeService) getUserAccessToken(matrixUserID string) (string, error) {
	b.userMutex.RLock()
	defer b.userMutex.RUnlock()

	user, exists := b.users[matrixUserID]
	if !exists {
		return "", fmt.Errorf("user not found: %s", matrixUserID)
	}

	return user.AccessToken, nil
}

// Helper functions

func extractUserIDFromMatrixID(matrixID string) string {
	// Extract user ID from @anon_xxx:domain
	// For now, use the matrix ID as-is
	return matrixID
}

func generateDisplayName(matrixID string) string {
	// Generate random display name
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	suffix := base64.URLEncoding.EncodeToString(randomBytes)[:6]
	return fmt.Sprintf("User-%s", suffix)
}

func generateDeviceID() string {
	// Generate random device ID
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	return base64.URLEncoding.EncodeToString(randomBytes)
}

// GetHomeserverURL returns the Matrix homeserver URL
func (b *BridgeService) GetHomeserverURL() string {
	return b.homeserverURL
}

// GetStats returns bridge statistics
func (b *BridgeService) GetStats() map[string]interface{} {
	b.roomMutex.RLock()
	b.userMutex.RLock()
	b.memberMutex.RLock()
	defer b.roomMutex.RUnlock()
	defer b.userMutex.RUnlock()
	defer b.memberMutex.RUnlock()

	totalMembers := 0
	for _, members := range b.members {
		totalMembers += len(members)
	}

	activeUsers := 0
	for _, user := range b.users {
		if user.IsActive {
			activeUsers++
		}
	}

	return map[string]interface{}{
		"total_rooms":       len(b.rooms),
		"total_users":       len(b.users),
		"active_users":      activeUsers,
		"total_memberships": totalMembers,
		"homeserver_url":    b.homeserverURL,
	}
}
