package matrix

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// SyncManager manages the Matrix /sync loop
type SyncManager struct {
	client    *Client
	since     string
	handlers  []EventHandler
	stopChan  chan struct{}
	isRunning bool
	mu        sync.RWMutex

	// Sync statistics
	syncCount   int64
	errorCount  int64
	lastSyncAt  time.Time
	avgDuration time.Duration
}

// EventHandler processes Matrix events
type EventHandler interface {
	// HandleEvent processes a single event
	HandleEvent(ctx context.Context, event *Event, roomID string) error

	// GetHandlerName returns the name of the handler for logging
	GetHandlerName() string
}

// NewSyncManager creates a new sync manager
func NewSyncManager(client *Client) *SyncManager {
	return &SyncManager{
		client:   client,
		handlers: make([]EventHandler, 0),
		stopChan: make(chan struct{}),
	}
}

// RegisterHandler registers an event handler
func (sm *SyncManager) RegisterHandler(handler EventHandler) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.handlers = append(sm.handlers, handler)
	log.Printf("Registered event handler: %s", handler.GetHandlerName())
}

// Start starts the sync loop
func (sm *SyncManager) Start(ctx context.Context) error {
	sm.mu.Lock()
	if sm.isRunning {
		sm.mu.Unlock()
		return fmt.Errorf("sync manager already running")
	}
	sm.isRunning = true
	sm.mu.Unlock()

	log.Println("🔄 Starting Matrix sync loop...")
	go sm.syncLoop(ctx)

	return nil
}

// Stop stops the sync loop
func (sm *SyncManager) Stop() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.isRunning {
		return
	}

	log.Println("🛑 Stopping Matrix sync loop...")
	close(sm.stopChan)
	sm.isRunning = false
}

// IsRunning returns whether the sync loop is running
func (sm *SyncManager) IsRunning() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.isRunning
}

// GetStats returns sync statistics
func (sm *SyncManager) GetStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return map[string]interface{}{
		"is_running":      sm.isRunning,
		"sync_count":      sm.syncCount,
		"error_count":     sm.errorCount,
		"last_sync_at":    sm.lastSyncAt,
		"avg_duration_ms": sm.avgDuration.Milliseconds(),
		"handler_count":   len(sm.handlers),
		"since_token":     sm.since,
	}
}

// syncLoop is the main sync loop
func (sm *SyncManager) syncLoop(ctx context.Context) {
	log.Println("✅ Matrix sync loop started")

	// Initial sync with smaller timeout
	if sm.since == "" {
		log.Println("📥 Performing initial sync...")
		if err := sm.doSync(ctx, 5000); err != nil {
			log.Printf("❌ Initial sync failed: %v", err)
			sm.handleSyncError(ctx, err)
		} else {
			log.Println("✅ Initial sync complete")
		}
	}

	// Main sync loop with 30s timeout
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Sync loop stopped by context")
			return

		case <-sm.stopChan:
			log.Println("🛑 Sync loop stopped")
			return

		case <-ticker.C:
			if err := sm.doSync(ctx, 30000); err != nil {
				log.Printf("⚠️  Sync error: %v", err)
				sm.handleSyncError(ctx, err)
			}
		}
	}
}

// doSync performs a single sync request
func (sm *SyncManager) doSync(ctx context.Context, timeout int) error {
	startTime := time.Now()

	// Execute sync
	syncResp, err := sm.client.Sync(ctx, sm.since, timeout)
	if err != nil {
		sm.mu.Lock()
		sm.errorCount++
		sm.mu.Unlock()
		return fmt.Errorf("sync request failed: %w", err)
	}

	// Update statistics
	duration := time.Since(startTime)
	sm.mu.Lock()
	sm.syncCount++
	sm.lastSyncAt = time.Now()
	sm.since = syncResp.NextBatch

	// Calculate average duration
	if sm.avgDuration == 0 {
		sm.avgDuration = duration
	} else {
		sm.avgDuration = (sm.avgDuration + duration) / 2
	}
	sm.mu.Unlock()

	// Process sync response
	if err := sm.processSyncResponse(ctx, syncResp); err != nil {
		log.Printf("⚠️  Error processing sync response: %v", err)
	}

	// Log sync stats periodically
	if sm.syncCount%100 == 0 {
		log.Printf("📊 Sync stats: count=%d, avg_duration=%dms, errors=%d",
			sm.syncCount, sm.avgDuration.Milliseconds(), sm.errorCount)
	}

	return nil
}

// processSyncResponse processes the sync response
func (sm *SyncManager) processSyncResponse(ctx context.Context, resp *SyncResponse) error {
	// Process joined rooms
	for roomID, roomData := range resp.Rooms.Join {
		// Process timeline events
		for i := range roomData.Timeline.Events {
			sm.processEvent(ctx, &roomData.Timeline.Events[i], roomID)
		}

		// Process state events
		for i := range roomData.State.Events {
			sm.processEvent(ctx, &roomData.State.Events[i], roomID)
		}

		// Process ephemeral events (typing, receipts)
		for i := range roomData.Ephemeral.Events {
			sm.processEvent(ctx, &roomData.Ephemeral.Events[i], roomID)
		}

		// Process account data
		for i := range roomData.AccountData.Events {
			sm.processEvent(ctx, &roomData.AccountData.Events[i], roomID)
		}
	}

	// Process invited rooms
	for roomID, roomData := range resp.Rooms.Invite {
		for i := range roomData.InviteState.Events {
			sm.processEvent(ctx, &roomData.InviteState.Events[i], roomID)
		}
	}

	// Process left rooms
	for roomID, roomData := range resp.Rooms.Leave {
		for i := range roomData.Timeline.Events {
			sm.processEvent(ctx, &roomData.Timeline.Events[i], roomID)
		}

		for i := range roomData.State.Events {
			sm.processEvent(ctx, &roomData.State.Events[i], roomID)
		}
	}

	// Process presence events
	for i := range resp.Presence.Events {
		sm.processEvent(ctx, &resp.Presence.Events[i], "")
	}

	// Process to-device events
	for i := range resp.ToDevice.Events {
		sm.processEvent(ctx, &resp.ToDevice.Events[i], "")
	}

	return nil
}

// processEvent processes a single event through all handlers
func (sm *SyncManager) processEvent(ctx context.Context, event *Event, roomID string) {
	if event.RoomID == "" {
		event.RoomID = roomID
	}

	sm.mu.RLock()
	handlers := sm.handlers
	sm.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler.HandleEvent(ctx, event, roomID); err != nil {
			log.Printf("⚠️  Handler %s error for event type %s: %v",
				handler.GetHandlerName(), event.Type, err)
		}
	}
}

// handleSyncError handles sync errors with backoff
func (sm *SyncManager) handleSyncError(ctx context.Context, err error) {
	sm.mu.Lock()
	errorCount := sm.errorCount
	sm.mu.Unlock()

	// Exponential backoff
	backoff := time.Duration(min(errorCount, 10)) * time.Second
	if backoff < 1*time.Second {
		backoff = 1 * time.Second
	}
	if backoff > 30*time.Second {
		backoff = 30 * time.Second
	}

	log.Printf("⏱️  Retrying in %v... (error count: %d)", backoff, errorCount)

	select {
	case <-time.After(backoff):
		// Continue
	case <-ctx.Done():
		return
	case <-sm.stopChan:
		return
	}
}

// SetSinceToken sets the since token for the next sync
func (sm *SyncManager) SetSinceToken(since string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.since = since
}

// GetSinceToken gets the current since token
func (sm *SyncManager) GetSinceToken() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.since
}

// min returns the minimum of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
