package model

import (
	"time"

	"github.com/google/uuid"
)

// PeerConnectionState represents the connection state of a peer
type PeerConnectionState string

const (
	PeerNotConnected PeerConnectionState = "NOT_CONNECTED"
	PeerConnecting   PeerConnectionState = "CONNECTING"
	PeerConnected    PeerConnectionState = "CONNECTED"
)

// Peer represents a nearby peer discovered via P2P or server
type Peer struct {
	ID             uuid.UUID           `json:"id" gorm:"type:uuid;primaryKey"`
	DisplayName    string              `json:"display_name" gorm:"size:50;not null"`
	ConnectionState PeerConnectionState `json:"connection_state" gorm:"type:varchar(20);default:'NOT_CONNECTED'"`
	DiscoveredAt   time.Time           `json:"discovered_at" gorm:"not null"`
	LastSeenAt     time.Time           `json:"last_seen_at" gorm:"not null"`
	SignalStrength *float64            `json:"signal_strength,omitempty" gorm:"type:decimal(5,2)"`
	StationID      string              `json:"station_id" gorm:"size:20"`
	EndpointID     *string             `json:"endpoint_id,omitempty" gorm:"size:100"`
	Metadata       map[string]string   `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// TableName overrides the table name
func (Peer) TableName() string {
	return "peers"
}

// IsConnected returns true if peer is connected
func (p *Peer) IsConnected() bool {
	return p.ConnectionState == PeerConnected
}

// IsConnecting returns true if peer is connecting
func (p *Peer) IsConnecting() bool {
	return p.ConnectionState == PeerConnecting
}

// CanConnect returns true if peer can be connected to
func (p *Peer) CanConnect() bool {
	return p.ConnectionState == PeerNotConnected
}

// SecondsSinceLastSeen returns seconds since peer was last seen
func (p *Peer) SecondsSinceLastSeen() int64 {
	return int64(time.Since(p.LastSeenAt).Seconds())
}

// IsStale returns true if peer hasn't been seen in 60 seconds
func (p *Peer) IsStale() bool {
	return p.SecondsSinceLastSeen() > 60
}

// UpdateLastSeen updates the last seen timestamp
func (p *Peer) UpdateLastSeen() {
	p.LastSeenAt = time.Now()
}

// UpdateConnectionState updates the connection state
func (p *Peer) UpdateConnectionState(state PeerConnectionState) {
	p.ConnectionState = state
	p.UpdatedAt = time.Now()
}

// NewPeer creates a new peer instance
func NewPeer(displayName, stationID string) *Peer {
	now := time.Now()
	return &Peer{
		ID:              uuid.New(),
		DisplayName:     displayName,
		ConnectionState: PeerNotConnected,
		DiscoveredAt:    now,
		LastSeenAt:      now,
		StationID:       stationID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
