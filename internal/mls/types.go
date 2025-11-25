package mls

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// MLS Core Types (RFC 9420)
// ============================================================

// MLSGroup represents an MLS group
type MLSGroup struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GroupID       string    `gorm:"uniqueIndex;not null" json:"group_id"`
	RoomID        *string   `gorm:"index" json:"room_id,omitempty"`
	StationID     *string   `gorm:"index" json:"station_id,omitempty"`
	CreatorUserID string    `gorm:"not null" json:"creator_user_id"`

	// MLS State
	Epoch           int64  `gorm:"not null;default:0" json:"epoch"`
	TreeHash        []byte `gorm:"not null" json:"tree_hash"`
	ConfirmationTag []byte `gorm:"not null" json:"confirmation_tag"`

	// Encryption Config
	CipherSuite     uint16 `gorm:"not null" json:"cipher_suite"`
	ProtocolVersion string `gorm:"not null" json:"protocol_version"`

	// Metadata
	MemberCount int       `gorm:"default:0" json:"member_count"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:now()" json:"updated_at"`
	LastEpochAt time.Time `gorm:"default:now()" json:"last_epoch_at"`
}

// TableName specifies the table name for MLSGroup
func (MLSGroup) TableName() string {
	return "mls_groups"
}

// MLSGroupMember represents a member in an MLS group
type MLSGroupMember struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GroupID            string    `gorm:"index;not null" json:"group_id"`
	UserID             string    `gorm:"index;not null" json:"user_id"`
	ClientID           string    `gorm:"not null" json:"client_id"`
	CredentialID       string    `gorm:"not null" json:"credential_id"`
	SignaturePublicKey []byte    `gorm:"not null" json:"signature_public_key"`

	// Membership State
	JoinedAtEpoch  int64  `gorm:"not null" json:"joined_at_epoch"`
	RemovedAtEpoch *int64 `json:"removed_at_epoch,omitempty"`
	IsActive       bool   `gorm:"default:true;index" json:"is_active"`

	// Timestamps
	JoinedAt   time.Time  `gorm:"default:now()" json:"joined_at"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
}

// TableName specifies the table name for MLSGroupMember
func (MLSGroupMember) TableName() string {
	return "mls_group_members"
}

// MLSCredential represents a client credential
type MLSCredential struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CredentialID string    `gorm:"uniqueIndex;not null" json:"credential_id"`
	UserID       string    `gorm:"index;not null" json:"user_id"`
	ClientID     string    `gorm:"index;not null" json:"client_id"`

	// Credential Data
	CredentialType     string `gorm:"not null" json:"credential_type"` // "basic" or "x509"
	CredentialData     []byte `gorm:"not null" json:"credential_data"`
	SignaturePublicKey []byte `gorm:"not null" json:"signature_public_key"`
	CertificateChain   []byte `json:"certificate_chain,omitempty"` // For X.509

	// Lifecycle
	IssuedAt  time.Time  `gorm:"not null;default:now()" json:"issued_at"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
}

// TableName specifies the table name for MLSCredential
func (MLSCredential) TableName() string {
	return "mls_credentials"
}

// MLSKeyPackage represents a KeyPackage
type MLSKeyPackage struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	KeyPackageID string    `gorm:"uniqueIndex;not null" json:"key_package_id"`
	UserID       string    `gorm:"index;not null" json:"user_id"`
	ClientID     string    `gorm:"not null" json:"client_id"`
	CredentialID string    `gorm:"not null" json:"credential_id"`

	// KeyPackage Data
	KeyPackageData []byte `gorm:"not null" json:"key_package_data"`
	CipherSuite    uint16 `gorm:"not null" json:"cipher_suite"`

	// Lifecycle
	CreatedAt         time.Time  `gorm:"default:now()" json:"created_at"`
	ConsumedAt        *time.Time `json:"consumed_at,omitempty"`
	ConsumedByGroupID *string    `json:"consumed_by_group_id,omitempty"`
	IsConsumed        bool       `gorm:"default:false" json:"is_consumed"`

	// Rate Limiting
	CreatedByIP string `json:"created_by_ip,omitempty"`
}

// TableName specifies the table name for MLSKeyPackage
func (MLSKeyPackage) TableName() string {
	return "mls_key_packages"
}

// MLSMessage represents an MLS protocol message
type MLSMessage struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MessageID      string    `gorm:"uniqueIndex;not null" json:"message_id"`
	GroupID        string    `gorm:"index;not null" json:"group_id"`
	SenderClientID string    `gorm:"not null" json:"sender_client_id"`

	// Message Type
	MessageType string `gorm:"not null" json:"message_type"` // "application" | "proposal" | "commit"

	// Epoch & Ordering
	Epoch          int64 `gorm:"not null;index" json:"epoch"`
	SequenceNumber int64 `gorm:"not null" json:"sequence_number"`

	// Message Data (encrypted)
	Ciphertext        []byte `gorm:"not null" json:"ciphertext"`
	AuthenticatedData []byte `json:"authenticated_data,omitempty"`

	// Timestamps
	SentAt      time.Time  `gorm:"default:now();index" json:"sent_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`

	// Delivery Tracking
	RecipientCount int `json:"recipient_count"`
	DeliveredCount int `gorm:"default:0" json:"delivered_count"`
}

// TableName specifies the table name for MLSMessage
func (MLSMessage) TableName() string {
	return "mls_messages"
}

// MLSEpochHistory represents epoch change history
type MLSEpochHistory struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GroupID string    `gorm:"index;not null" json:"group_id"`
	Epoch   int64     `gorm:"not null" json:"epoch"`

	// State Snapshot
	TreeHash        []byte `gorm:"not null" json:"tree_hash"`
	ConfirmationTag []byte `gorm:"not null" json:"confirmation_tag"`
	MemberCount     int    `gorm:"not null" json:"member_count"`

	// Change Metadata
	ChangeType        string  `gorm:"not null" json:"change_type"` // "member_add" | "member_remove" | "update"
	ChangedByClientID *string `json:"changed_by_client_id,omitempty"`

	// Timestamps
	StartedAt time.Time  `gorm:"not null" json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

// TableName specifies the table name for MLSEpochHistory
func (MLSEpochHistory) TableName() string {
	return "mls_epoch_history"
}

// ============================================================
// MLS Cipher Suites (RFC 9420)
// ============================================================

const (
	// Recommended cipher suite for TrainBlink
	MLS_128_DHKEMX25519_AES128GCM_SHA256_Ed25519        uint16 = 0x0001
	MLS_128_DHKEMP256_AES128GCM_SHA256_P256             uint16 = 0x0002
	MLS_128_DHKEMX25519_CHACHA20POLY1305_SHA256_Ed25519 uint16 = 0x0003
	MLS_256_DHKEMX448_AES256GCM_SHA512_Ed448            uint16 = 0x0004
	MLS_256_DHKEMP521_AES256GCM_SHA512_P521             uint16 = 0x0005
	MLS_256_DHKEMX448_CHACHA20POLY1305_SHA512_Ed448     uint16 = 0x0006
	MLS_256_DHKEMP384_AES256GCM_SHA384_P384             uint16 = 0x0007
)

// CipherSuiteName returns the human-readable name of a cipher suite
func CipherSuiteName(suite uint16) string {
	names := map[uint16]string{
		0x0001: "MLS_128_DHKEMX25519_AES128GCM_SHA256_Ed25519",
		0x0002: "MLS_128_DHKEMP256_AES128GCM_SHA256_P256",
		0x0003: "MLS_128_DHKEMX25519_CHACHA20POLY1305_SHA256_Ed25519",
		0x0004: "MLS_256_DHKEMX448_AES256GCM_SHA512_Ed448",
		0x0005: "MLS_256_DHKEMP521_AES256GCM_SHA512_P521",
		0x0006: "MLS_256_DHKEMX448_CHACHA20POLY1305_SHA512_Ed448",
		0x0007: "MLS_256_DHKEMP384_AES256GCM_SHA384_P384",
	}
	if name, ok := names[suite]; ok {
		return name
	}
	return "Unknown"
}

// ============================================================
// MLS Message Types
// ============================================================

const (
	MessageTypeApplication = "application" // Encrypted application data
	MessageTypeProposal    = "proposal"    // Group operation proposal
	MessageTypeCommit      = "commit"      // Commit to group changes
)

// ============================================================
// MLS Credential Types
// ============================================================

const (
	CredentialTypeBasic = "basic" // Basic credential (username)
	CredentialTypeX509  = "x509"  // X.509 certificate
)

// ============================================================
// MLS Change Types
// ============================================================

const (
	ChangeTypeMemberAdd    = "member_add"
	ChangeTypeMemberRemove = "member_remove"
	ChangeTypeUpdate       = "update"
)

// ============================================================
// MLS Protocol Version
// ============================================================

const (
	ProtocolVersionMLS10 = "mls10" // RFC 9420
)

// ============================================================
// MLS Configuration
// ============================================================

// MLSConfig holds MLS system configuration
type MLSConfig struct {
	Enabled            bool             `json:"enabled"`
	CipherSuites       []uint16         `json:"cipher_suites"`
	ProtocolVersion    string           `json:"protocol_version"`
	MaxMembersPerGroup int              `json:"max_members_per_group"`
	MaxMessageSize     int              `json:"max_message_size"`
	KeyPackageConfig   KeyPackageConfig `json:"keypackage_config"`
}

// KeyPackageConfig holds KeyPackage configuration
type KeyPackageConfig struct {
	MaxPerUser      int `json:"max_per_user"`
	TTLDays         int `json:"ttl_days"`
	UploadRateLimit int `json:"upload_rate_limit"` // Per hour
}

// DefaultMLSConfig returns default MLS configuration
func DefaultMLSConfig() *MLSConfig {
	return &MLSConfig{
		Enabled: true,
		CipherSuites: []uint16{
			MLS_128_DHKEMX25519_AES128GCM_SHA256_Ed25519,
			MLS_128_DHKEMX25519_CHACHA20POLY1305_SHA256_Ed25519,
		},
		ProtocolVersion:    ProtocolVersionMLS10,
		MaxMembersPerGroup: 10000,
		MaxMessageSize:     100000, // 100KB
		KeyPackageConfig: KeyPackageConfig{
			MaxPerUser:      100,
			TTLDays:         30,
			UploadRateLimit: 100, // Per hour
		},
	}
}

// ============================================================
// Request/Response Types
// ============================================================

// CreateGroupRequest is the request to create a new MLS group
type CreateGroupRequest struct {
	StationID           *string `json:"station_id,omitempty"`
	RoomID              *string `json:"room_id,omitempty"`
	CipherSuite         uint16  `json:"cipher_suite"`
	CreatorCredentialID string  `json:"creator_credential_id"`
	CreatorKeyPackage   string  `json:"creator_key_package,omitempty"` // Base64 encoded, optional
}

// CreateGroupResponse is the response after creating a group
type CreateGroupResponse struct {
	GroupID         string    `json:"group_id"`
	Epoch           int64     `json:"epoch"`
	CipherSuite     uint16    `json:"cipher_suite"`
	ProtocolVersion string    `json:"protocol_version"`
	CreatedAt       time.Time `json:"created_at"`
}

// UploadKeyPackageRequest is the request to upload a KeyPackage
type UploadKeyPackageRequest struct {
	ClientID     string `json:"client_id"`
	CredentialID string `json:"credential_id"`
	CipherSuite  uint16 `json:"cipher_suite"`
	KeyPackage   string `json:"key_package"` // Base64 encoded
}

// UploadKeyPackageResponse is the response after uploading a KeyPackage
type UploadKeyPackageResponse struct {
	KeyPackageID string    `json:"key_package_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// SendMessageRequest is the request to send an MLS message
type SendMessageRequest struct {
	MessageType       string `json:"message_type"`
	SenderClientID    string `json:"sender_client_id"`
	Ciphertext        string `json:"ciphertext"`                   // Base64 encoded
	AuthenticatedData string `json:"authenticated_data,omitempty"` // Base64 encoded
}

// SendMessageResponse is the response after sending a message
type SendMessageResponse struct {
	MessageID      string    `json:"message_id"`
	Epoch          int64     `json:"epoch"`
	SequenceNumber int64     `json:"sequence_number"`
	SentAt         time.Time `json:"sent_at"`
}

// GetMessagesResponse is the response for getting pending messages
type GetMessagesResponse struct {
	Messages     []MessageData `json:"messages"`
	NextSequence int64         `json:"next_sequence"`
}

// MessageData represents a message in the response
type MessageData struct {
	MessageID         string    `json:"message_id"`
	SenderClientID    string    `json:"sender_client_id"`
	MessageType       string    `json:"message_type"`
	Epoch             int64     `json:"epoch"`
	SequenceNumber    int64     `json:"sequence_number"`
	Ciphertext        string    `json:"ciphertext"`                   // Base64 encoded
	AuthenticatedData string    `json:"authenticated_data,omitempty"` // Base64 encoded
	SentAt            time.Time `json:"sent_at"`
}

// GroupStatsResponse is the response for group statistics
type GroupStatsResponse struct {
	GroupID      string    `json:"group_id"`
	MemberCount  int       `json:"member_count"`
	MessageCount int64     `json:"message_count"`
	Epoch        int64     `json:"epoch"`
	LastActive   time.Time `json:"last_active"`
}

// AddMemberRequest is the request to add a member to a group
type AddMemberRequest struct {
	UserID       string `json:"user_id"`
	ClientID     string `json:"client_id"`
	CredentialID string `json:"credential_id"`
	KeyPackage   string `json:"key_package"` // Base64 encoded
}

// ClaimKeyPackageRequest is the request to claim a KeyPackage
type ClaimKeyPackageRequest struct {
	UserID      string `json:"user_id"`
	GroupID     string `json:"group_id"`
	CipherSuite uint16 `json:"cipher_suite"`
}

// AcknowledgeMessageRequest is the request to acknowledge message delivery
type AcknowledgeMessageRequest struct {
	ClientID string `json:"client_id"`
}
