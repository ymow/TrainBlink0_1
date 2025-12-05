package message

import (
	"context"
	"log"
	"time"

	"github.com/ymow/messenger_protocol_research/internal/model"
)

// StatusWorker handles automatic status updates for stale messages
// Messages stuck in PENDING or SENDING for too long are marked as FAILED
type StatusWorker struct {
	service  *Service
	interval time.Duration // Check interval (default: 5 minutes)
	timeout  time.Duration // Message timeout (default: 5 minutes)
}

// NewStatusWorker creates a new status worker
func NewStatusWorker(service *Service) *StatusWorker {
	return &StatusWorker{
		service:  service,
		interval: 5 * time.Minute,
		timeout:  5 * time.Minute,
	}
}

// SetInterval sets the check interval
func (w *StatusWorker) SetInterval(interval time.Duration) {
	w.interval = interval
}

// SetTimeout sets the message timeout duration
func (w *StatusWorker) SetTimeout(timeout time.Duration) {
	w.timeout = timeout
}

// Start starts the background worker
func (w *StatusWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Println("⏰ Message status worker started")

	// Run immediately on start
	w.processStaleMessages(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Message status worker stopped")
			return
		case <-ticker.C:
			w.processStaleMessages(ctx)
		}
	}
}

// processStaleMessages marks stale messages as FAILED
func (w *StatusWorker) processStaleMessages(ctx context.Context) {
	cutoffTime := time.Now().Add(-w.timeout)

	// Find messages stuck in PENDING or SENDING for too long
	result := w.service.db.WithContext(ctx).
		Model(&model.ChatMessage{}).
		Where("delivery_status IN ? AND updated_at < ?",
			[]model.MessageDeliveryStatus{model.MessagePending, model.MessageSending},
			cutoffTime,
		).
		Updates(map[string]interface{}{
			"delivery_status": model.MessageFailed,
			"updated_at":      time.Now(),
		})

	if result.Error != nil {
		log.Printf("⚠️  Failed to mark stale messages: %v", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf("⚠️  Marked %d stale messages as FAILED", result.RowsAffected)
	}
}
