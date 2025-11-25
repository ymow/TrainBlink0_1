package geofence

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ymow/messenger_protocol_research/internal/matrix"
	"github.com/ymow/messenger_protocol_research/internal/model"
)

// Service handles geofencing operations
type Service struct {
	matrixBridge *matrix.BridgeService

	// In-memory storage (will be replaced with PostgreSQL/Redis in production)
	stations       map[string]*model.Station
	sessions       map[uuid.UUID]*model.UserSession
	activeSessions map[string]uuid.UUID // userID -> sessionID
	stationMutex   sync.RWMutex
	sessionMutex   sync.RWMutex
}

// NewService creates a new geofencing service
func NewService(matrixBridge *matrix.BridgeService) *Service {
	return &Service{
		matrixBridge:   matrixBridge,
		stations:       make(map[string]*model.Station),
		sessions:       make(map[uuid.UUID]*model.UserSession),
		activeSessions: make(map[string]uuid.UUID),
	}
}

// LoadStations loads station data (in production, this would load from database)
func (s *Service) LoadStations(stations []*model.Station) {
	s.stationMutex.Lock()
	defer s.stationMutex.Unlock()

	for _, station := range stations {
		s.stations[station.ID] = station
	}
}

// GetStationByID retrieves a station by ID
func (s *Service) GetStationByID(ctx context.Context, stationID string) (*model.Station, error) {
	s.stationMutex.RLock()
	defer s.stationMutex.RUnlock()

	station, exists := s.stations[stationID]
	if !exists {
		return nil, fmt.Errorf("station not found: %s", stationID)
	}

	return station, nil
}

// HandleEnterStation handles user entering a station
func (s *Service) HandleEnterStation(
	ctx context.Context,
	userID, deviceID string,
	req *model.EnterStationRequest,
) (*model.EnterStationResponse, error) {

	// 1. Validate station exists
	station, err := s.GetStationByID(ctx, req.StationID)
	if err != nil {
		return nil, fmt.Errorf("invalid station: %w", err)
	}

	// 2. Check if user already has active session
	s.sessionMutex.RLock()
	existingSessionID, hasSession := s.activeSessions[userID]
	s.sessionMutex.RUnlock()

	if hasSession {
		// User already in a station, exit first
		existingSession := s.sessions[existingSessionID]
		if existingSession != nil && existingSession.ExitedAt == nil {
			return nil, fmt.Errorf("user already in station %s, please exit first", existingSession.StationID)
		}
	}

	// 3. Create new session
	sessionID := uuid.New()
	now := time.Now()

	capabilitiesJSON, _ := json.Marshal(req.Capabilities)

	session := &model.UserSession{
		ID:               sessionID,
		UserID:           userID,
		DeviceID:         deviceID,
		StationID:        req.StationID,
		EnteredAt:        now,
		EntryLatitude:    req.Coordinates.Latitude,
		EntryLongitude:   req.Coordinates.Longitude,
		ClientVersion:    req.ClientVersion,
		CapabilitiesJSON: string(capabilitiesJSON),
		CreatedAt:        now,
	}

	s.sessionMutex.Lock()
	s.sessions[sessionID] = session
	s.activeSessions[userID] = sessionID
	s.sessionMutex.Unlock()

	// 4. Prepare P2P Resources (if enabled)
	var p2pResources *model.P2PResources
	if req.Capabilities.P2PEnabled {
		activePeers := s.getActiveUsersInStation(req.StationID)
		p2pResources = &model.P2PResources{
			ActivePeersNearby: activePeers - 1, // Exclude self
			SignalingServer:   "wss://signal.trainblink.org",
			ICEServers:        s.getICEServers(),
		}
	}

	// 5. Prepare Matrix Resources (if enabled) ⭐
	var matrixResources *model.MatrixResources
	if req.Capabilities.MatrixEnabled {
		matrixRes, err := s.setupMatrixForUser(ctx, userID, station, session)
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Matrix setup failed: %v\n", err)
			// Continue with P2P only
		} else {
			matrixResources = matrixRes
		}
	}

	// 6. Return response
	return &model.EnterStationResponse{
		SessionID:        sessionID,
		Station:          station,
		P2PResources:     p2pResources,
		MatrixResources:  matrixResources,
		ContentAvailable: s.getContentCount(ctx, req.StationID),
		Recommendations:  s.getRecommendations(ctx, userID, req.StationID),
	}, nil
}

// setupMatrixForUser sets up Matrix for a user entering a station
func (s *Service) setupMatrixForUser(
	ctx context.Context,
	userID string,
	station *model.Station,
	session *model.UserSession,
) (*model.MatrixResources, error) {

	// 1. Generate anonymous Matrix User ID
	matrixUserID := s.matrixBridge.GenerateMatrixUserID(userID)

	// 2. Get or create station room
	roomInfo, err := s.matrixBridge.GetOrCreateStationRoom(ctx, station.ID, station.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create room: %w", err)
	}

	// 3. Join room
	accessToken, err := s.matrixBridge.JoinRoom(ctx, matrixUserID, roomInfo.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to join room: %w", err)
	}

	// 4. Update session with Matrix info
	now := time.Now()
	session.MatrixUserID = matrixUserID
	session.MatrixRoomID = roomInfo.RoomID
	session.MatrixJoinedAt = &now

	// 5. Return Matrix resources
	return &model.MatrixResources{
		RoomID:            roomInfo.RoomID,
		RoomAlias:         roomInfo.RoomAlias,
		MatrixUserID:      matrixUserID,
		HomeserverURL:     s.matrixBridge.GetHomeserverURL(),
		AccessToken:       accessToken,
		EncryptionEnabled: roomInfo.EncryptionEnabled,
		MLSGroupID:        roomInfo.MLSGroupID,
		MemberCount:       roomInfo.MemberCount,
	}, nil
}

// HandleExitStation handles user exiting a station
func (s *Service) HandleExitStation(
	ctx context.Context,
	sessionID uuid.UUID,
	userID, stationID string,
	activity *model.ActivitySummary,
) (*model.ExitStationResponse, error) {

	// 1. Find session
	s.sessionMutex.RLock()
	session, exists := s.sessions[sessionID]
	s.sessionMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if session.UserID != userID {
		return nil, fmt.Errorf("session does not belong to user")
	}

	// 2. Update session with exit information
	now := time.Now()
	duration := int(now.Sub(session.EnteredAt).Seconds())

	session.ExitedAt = &now
	session.DurationSeconds = &duration
	session.P2PChatsCreated = activity.P2PChatsCreated
	session.P2PMessagesSent = activity.P2PMessagesSent
	session.P2PContentShared = activity.ContentShared
	session.MatrixMessagesSent = activity.MatrixMessagesSent

	// 3. Leave Matrix room if joined
	matrixLeft := false
	if session.MatrixRoomID != "" && session.MatrixUserID != "" {
		err := s.matrixBridge.LeaveRoom(ctx, session.MatrixUserID, session.MatrixRoomID)
		if err != nil {
			fmt.Printf("Failed to leave Matrix room: %v\n", err)
		} else {
			matrixLeft = true
			session.MatrixLeftAt = &now
		}
	}

	// 4. Remove from active sessions
	s.sessionMutex.Lock()
	delete(s.activeSessions, userID)
	s.sessionMutex.Unlock()

	// 5. Return response
	return &model.ExitStationResponse{
		SessionSummary: *activity,
		CleanupStatus: model.CleanupStatus{
			P2PClosed:        true,
			MatrixLeft:       matrixLeft,
			LocalDataCleared: true,
		},
	}, nil
}

// GetActiveUsersInStation returns count of active users in a station
func (s *Service) getActiveUsersInStation(stationID string) int {
	s.sessionMutex.RLock()
	defer s.sessionMutex.RUnlock()

	count := 0
	for _, sessionID := range s.activeSessions {
		if session, exists := s.sessions[sessionID]; exists {
			if session.StationID == stationID && session.ExitedAt == nil {
				count++
			}
		}
	}

	return count
}

// Helper functions

func (s *Service) getICEServers() []model.ICEServer {
	return []model.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
		{
			URLs:       []string{"turn:turn.trainblink.org:3478"},
			Username:   "trainblink",
			Credential: "temporary-credential",
		},
	}
}

func (s *Service) getContentCount(ctx context.Context, stationID string) int {
	// TODO: Implement content counting
	return 0
}

func (s *Service) getRecommendations(ctx context.Context, userID, stationID string) *model.Recommendations {
	// TODO: Implement recommendation engine
	return &model.Recommendations{
		Peers:           []string{},
		IcebreakerCards: []string{"card_001", "card_002"},
	}
}

// GetSessionByID retrieves a session by ID
func (s *Service) GetSessionByID(sessionID uuid.UUID) (*model.UserSession, error) {
	s.sessionMutex.RLock()
	defer s.sessionMutex.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

// GetStats returns service statistics
func (s *Service) GetStats() map[string]interface{} {
	s.sessionMutex.RLock()
	s.stationMutex.RLock()
	defer s.sessionMutex.RUnlock()
	defer s.stationMutex.RUnlock()

	activeSessions := 0
	for _, sessionID := range s.activeSessions {
		if session, exists := s.sessions[sessionID]; exists && session.ExitedAt == nil {
			activeSessions++
		}
	}

	matrixStats := s.matrixBridge.GetStats()

	return map[string]interface{}{
		"total_stations":  len(s.stations),
		"total_sessions":  len(s.sessions),
		"active_sessions": activeSessions,
		"matrix_stats":    matrixStats,
	}
}
