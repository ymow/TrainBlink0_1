package matrix

// ============================================================
// Matrix Client-Server API Types
// Comprehensive type definitions for Matrix protocol
// ============================================================

// ============================================================
// Authentication Types
// ============================================================

type LoginRequest struct {
	Type                     string      `json:"type"`
	Identifier               interface{} `json:"identifier,omitempty"`
	Password                 string      `json:"password,omitempty"`
	Token                    string      `json:"token,omitempty"`
	InitialDeviceDisplayName string      `json:"initial_device_display_name,omitempty"`
}

type LoginResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	DeviceID     string `json:"device_id"`
	HomeServer   string `json:"home_server"`
	WellKnown    interface{} `json:"well_known,omitempty"`
}

type RegisterRequest struct {
	Username                 string `json:"username,omitempty"`
	Password                 string `json:"password,omitempty"`
	InitialDeviceDisplayName string `json:"initial_device_display_name,omitempty"`
	InhibitLogin             bool   `json:"inhibit_login"`
}

type RegisterResponse struct {
	UserID      string `json:"user_id"`
	AccessToken string `json:"access_token,omitempty"`
	DeviceID    string `json:"device_id,omitempty"`
	HomeServer  string `json:"home_server"`
}

// ============================================================
// Room Types
// ============================================================

type CreateRoomRequest struct {
	RoomAliasName             string                 `json:"room_alias_name,omitempty"`
	Name                      string                 `json:"name,omitempty"`
	Topic                     string                 `json:"topic,omitempty"`
	Visibility                string                 `json:"visibility"` // "public" or "private"
	Preset                    string                 `json:"preset,omitempty"`
	InitialState              []StateEvent           `json:"initial_state,omitempty"`
	PowerLevelContentOverride map[string]interface{} `json:"power_level_content_override,omitempty"`
	CreationContent           map[string]interface{} `json:"creation_content,omitempty"`
	Invite                    []string               `json:"invite,omitempty"`
}

type CreateRoomResponse struct {
	RoomID string `json:"room_id"`
}

type JoinRoomResponse struct {
	RoomID string `json:"room_id"`
}

type JoinedRoomsResponse struct {
	JoinedRooms []string `json:"joined_rooms"`
}

// ============================================================
// Event Types
// ============================================================

type Event struct {
	Type           string                 `json:"type"`
	EventID        string                 `json:"event_id,omitempty"`
	Sender         string                 `json:"sender,omitempty"`
	RoomID         string                 `json:"room_id,omitempty"`
	StateKey       *string                `json:"state_key,omitempty"`
	OriginServerTS int64                  `json:"origin_server_ts,omitempty"`
	Content        map[string]interface{} `json:"content"`
	Unsigned       map[string]interface{} `json:"unsigned,omitempty"`
	PrevContent    map[string]interface{} `json:"prev_content,omitempty"`
}

type StateEvent struct {
	Type     string                 `json:"type"`
	StateKey string                 `json:"state_key"`
	Content  map[string]interface{} `json:"content"`
}

// ============================================================
// Message Types
// ============================================================

type SendMessageRequest struct {
	MsgType       string      `json:"msgtype"`
	Body          string      `json:"body"`
	Format        string      `json:"format,omitempty"`
	FormattedBody string      `json:"formatted_body,omitempty"`
	URL           string      `json:"url,omitempty"`
	Info          *MediaInfo  `json:"info,omitempty"`
	GeoURI        string      `json:"geo_uri,omitempty"`
	Filename      string      `json:"filename,omitempty"`
}

type MediaInfo struct {
	MimeType      string      `json:"mimetype,omitempty"`
	Size          int64       `json:"size,omitempty"`
	Width         int         `json:"w,omitempty"`
	Height        int         `json:"h,omitempty"`
	ThumbnailURL  string      `json:"thumbnail_url,omitempty"`
	ThumbnailInfo *MediaInfo  `json:"thumbnail_info,omitempty"`
}

type SendMessageResponse struct {
	EventID string `json:"event_id"`
}

// ============================================================
// Media Types
// ============================================================

type UploadResponse struct {
	ContentURI string `json:"content_uri"`
}

// ============================================================
// Sync Types
// ============================================================

type SyncResponse struct {
	NextBatch   string          `json:"next_batch"`
	Rooms       RoomsData       `json:"rooms"`
	Presence    PresenceData    `json:"presence"`
	AccountData AccountDataList `json:"account_data"`
	ToDevice    ToDeviceData    `json:"to_device,omitempty"`
	DeviceLists DeviceLists     `json:"device_lists,omitempty"`
}

type RoomsData struct {
	Join   map[string]JoinedRoomData  `json:"join"`
	Invite map[string]InvitedRoomData `json:"invite"`
	Leave  map[string]LeftRoomData    `json:"leave"`
}

type JoinedRoomData struct {
	Timeline            TimelineData        `json:"timeline"`
	State               StateData           `json:"state"`
	Ephemeral           EphemeralData       `json:"ephemeral"`
	AccountData         AccountDataList     `json:"account_data"`
	UnreadNotifications UnreadNotifications `json:"unread_notifications"`
	Summary             RoomSummary         `json:"summary,omitempty"`
}

type TimelineData struct {
	Events    []Event `json:"events"`
	Limited   bool    `json:"limited"`
	PrevBatch string  `json:"prev_batch"`
}

type StateData struct {
	Events []Event `json:"events"`
}

type EphemeralData struct {
	Events []Event `json:"events"`
}

type AccountDataList struct {
	Events []Event `json:"events"`
}

type PresenceData struct {
	Events []Event `json:"events"`
}

type InvitedRoomData struct {
	InviteState StateData `json:"invite_state"`
}

type LeftRoomData struct {
	Timeline TimelineData `json:"timeline"`
	State    StateData    `json:"state"`
}

type UnreadNotifications struct {
	HighlightCount    int `json:"highlight_count"`
	NotificationCount int `json:"notification_count"`
}

type RoomSummary struct {
	Heroes             []string `json:"m.heroes,omitempty"`
	JoinedMemberCount  int      `json:"m.joined_member_count,omitempty"`
	InvitedMemberCount int      `json:"m.invited_member_count,omitempty"`
}

type ToDeviceData struct {
	Events []Event `json:"events"`
}

type DeviceLists struct {
	Changed []string `json:"changed,omitempty"`
	Left    []string `json:"left,omitempty"`
}

// ============================================================
// Room State Types
// ============================================================

type MembersResponse struct {
	Chunk []Event `json:"chunk"`
}

type MessagesResponse struct {
	Start string  `json:"start"`
	End   string  `json:"end"`
	Chunk []Event `json:"chunk"`
}

// ============================================================
// Encryption Types (E2EE)
// ============================================================

type DeviceKeys struct {
	UserID     string                       `json:"user_id"`
	DeviceID   string                       `json:"device_id"`
	Algorithms []string                     `json:"algorithms"`
	Keys       map[string]string            `json:"keys"`
	Signatures map[string]map[string]string `json:"signatures"`
}

type UploadKeysRequest struct {
	DeviceKeys  *DeviceKeys            `json:"device_keys,omitempty"`
	OneTimeKeys map[string]interface{} `json:"one_time_keys,omitempty"`
}

type UploadKeysResponse struct {
	OneTimeKeyCounts map[string]int `json:"one_time_key_counts"`
}

type QueryKeysRequest struct {
	DeviceKeys map[string][]string `json:"device_keys"`
	Timeout    int                 `json:"timeout,omitempty"`
}

type QueryKeysResponse struct {
	DeviceKeys map[string]map[string]DeviceKeys `json:"device_keys"`
	Failures   map[string]interface{}           `json:"failures,omitempty"`
}

// ============================================================
// Typing Indicator Types
// ============================================================

type TypingRequest struct {
	Typing  bool  `json:"typing"`
	Timeout int64 `json:"timeout,omitempty"`
}

// ============================================================
// Read Receipt Types
// ============================================================

type ReadReceiptRequest struct {
	// Empty body for read receipt
}

// ============================================================
// Presence Types
// ============================================================

type PresenceRequest struct {
	Presence  string `json:"presence"` // "online", "offline", "unavailable"
	StatusMsg string `json:"status_msg,omitempty"`
}

type PresenceResponse struct {
	Presence       string `json:"presence"`
	LastActiveAgo  int64  `json:"last_active_ago,omitempty"`
	StatusMsg      string `json:"status_msg,omitempty"`
	CurrentlyActive bool   `json:"currently_active,omitempty"`
}

// ============================================================
// Error Types
// ============================================================

type MatrixError struct {
	ErrCode string `json:"errcode"`
	Error   string `json:"error"`
}

func (e *MatrixError) String() string {
	return e.ErrCode + ": " + e.Error
}

// ============================================================
// Filter Types (for /sync)
// ============================================================

type Filter struct {
	Room          *RoomFilter         `json:"room,omitempty"`
	Presence      *EventFilter        `json:"presence,omitempty"`
	AccountData   *EventFilter        `json:"account_data,omitempty"`
	EventFormat   string              `json:"event_format,omitempty"`
	EventFields   []string            `json:"event_fields,omitempty"`
}

type RoomFilter struct {
	NotRooms       []string            `json:"not_rooms,omitempty"`
	Rooms          []string            `json:"rooms,omitempty"`
	Ephemeral      *RoomEventFilter    `json:"ephemeral,omitempty"`
	IncludeLeave   bool                `json:"include_leave"`
	State          *StateFilter        `json:"state,omitempty"`
	Timeline       *RoomEventFilter    `json:"timeline,omitempty"`
	AccountData    *RoomEventFilter    `json:"account_data,omitempty"`
}

type EventFilter struct {
	Limit      int      `json:"limit,omitempty"`
	NotSenders []string `json:"not_senders,omitempty"`
	NotTypes   []string `json:"not_types,omitempty"`
	Senders    []string `json:"senders,omitempty"`
	Types      []string `json:"types,omitempty"`
}

type RoomEventFilter struct {
	Limit                int      `json:"limit,omitempty"`
	NotSenders           []string `json:"not_senders,omitempty"`
	NotTypes             []string `json:"not_types,omitempty"`
	Senders              []string `json:"senders,omitempty"`
	Types                []string `json:"types,omitempty"`
	LazyLoadMembers      bool     `json:"lazy_load_members,omitempty"`
	IncludeRedundantMembers bool  `json:"include_redundant_members,omitempty"`
	NotRooms             []string `json:"not_rooms,omitempty"`
	Rooms                []string `json:"rooms,omitempty"`
	ContainsURL          *bool    `json:"contains_url,omitempty"`
}

type StateFilter struct {
	Limit                int      `json:"limit,omitempty"`
	NotSenders           []string `json:"not_senders,omitempty"`
	NotTypes             []string `json:"not_types,omitempty"`
	Senders              []string `json:"senders,omitempty"`
	Types                []string `json:"types,omitempty"`
	LazyLoadMembers      bool     `json:"lazy_load_members,omitempty"`
	IncludeRedundantMembers bool  `json:"include_redundant_members,omitempty"`
	NotRooms             []string `json:"not_rooms,omitempty"`
	Rooms                []string `json:"rooms,omitempty"`
	ContainsURL          *bool    `json:"contains_url,omitempty"`
}

// ============================================================
// Power Level Types
// ============================================================

type PowerLevelContent struct {
	Ban           int                    `json:"ban,omitempty"`
	Events        map[string]int         `json:"events,omitempty"`
	EventsDefault int                    `json:"events_default,omitempty"`
	Invite        int                    `json:"invite,omitempty"`
	Kick          int                    `json:"kick,omitempty"`
	Redact        int                    `json:"redact,omitempty"`
	StateDefault  int                    `json:"state_default,omitempty"`
	Users         map[string]int         `json:"users,omitempty"`
	UsersDefault  int                    `json:"users_default,omitempty"`
}

// ============================================================
// Room Member Types
// ============================================================

type MemberContent struct {
	Membership  string `json:"membership"` // "invite", "join", "leave", "ban"
	DisplayName string `json:"displayname,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

// ============================================================
// Constants
// ============================================================

const (
	// Event Types
	EventTypeRoomMessage       = "m.room.message"
	EventTypeRoomMember        = "m.room.member"
	EventTypeRoomCreate        = "m.room.create"
	EventTypeRoomName          = "m.room.name"
	EventTypeRoomTopic         = "m.room.topic"
	EventTypeRoomAvatar        = "m.room.avatar"
	EventTypeRoomPowerLevels   = "m.room.power_levels"
	EventTypeRoomEncryption    = "m.room.encryption"
	EventTypeRoomEncrypted     = "m.room.encrypted"
	EventTypeRoomKey           = "m.room_key"
	EventTypeTyping            = "m.typing"
	EventTypeReceipt           = "m.receipt"
	EventTypePresence          = "m.presence"

	// Message Types
	MsgTypeText     = "m.text"
	MsgTypeEmote    = "m.emote"
	MsgTypeNotice   = "m.notice"
	MsgTypeImage    = "m.image"
	MsgTypeFile     = "m.file"
	MsgTypeAudio    = "m.audio"
	MsgTypeVideo    = "m.video"
	MsgTypeLocation = "m.location"

	// Membership States
	MembershipInvite = "invite"
	MembershipJoin   = "join"
	MembershipLeave  = "leave"
	MembershipBan    = "ban"
	MembershipKnock  = "knock"

	// Presence States
	PresenceOnline      = "online"
	PresenceOffline     = "offline"
	PresenceUnavailable = "unavailable"

	// Room Visibility
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"

	// Room Presets
	PresetPrivateChat        = "private_chat"
	PresetPublicChat         = "public_chat"
	PresetTrustedPrivateChat = "trusted_private_chat"
)
