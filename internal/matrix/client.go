package matrix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is a Matrix Client-Server API client
type Client struct {
	httpClient    *http.Client
	homeserverURL string
	accessToken   string
	userID        string
	deviceID      string
}

// NewClient creates a new Matrix client
func NewClient(homeserverURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 90 * time.Second, // Longer timeout for /sync
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
				MaxConnsPerHost:     50,
			},
		},
		homeserverURL: homeserverURL,
	}
}

// SetAccessToken sets the access token for authenticated requests
func (c *Client) SetAccessToken(token, userID, deviceID string) {
	c.accessToken = token
	c.userID = userID
	c.deviceID = deviceID
}

// ============================================================
// Authentication
// ============================================================

// Register registers a new user
func (c *Client) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	var resp RegisterResponse
	err := c.doRequest(ctx, "POST", "/_matrix/client/v3/register", req, &resp, false)
	if err != nil {
		return nil, err
	}

	// Set access token if login not inhibited
	if !req.InhibitLogin && resp.AccessToken != "" {
		c.SetAccessToken(resp.AccessToken, resp.UserID, resp.DeviceID)
	}

	return &resp, nil
}

// Login authenticates with the homeserver
func (c *Client) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	var resp LoginResponse
	err := c.doRequest(ctx, "POST", "/_matrix/client/v3/login", req, &resp, false)
	if err != nil {
		return nil, err
	}

	// Set access token
	c.SetAccessToken(resp.AccessToken, resp.UserID, resp.DeviceID)

	return &resp, nil
}

// Logout invalidates the access token
func (c *Client) Logout(ctx context.Context) error {
	return c.doRequest(ctx, "POST", "/_matrix/client/v3/logout", nil, nil, true)
}

// ============================================================
// Room Operations
// ============================================================

// CreateRoom creates a new room
func (c *Client) CreateRoom(ctx context.Context, req *CreateRoomRequest) (*CreateRoomResponse, error) {
	var resp CreateRoomResponse
	err := c.doRequest(ctx, "POST", "/_matrix/client/v3/createRoom", req, &resp, true)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// JoinRoom joins a room by ID or alias
func (c *Client) JoinRoom(ctx context.Context, roomIDOrAlias string) (*JoinRoomResponse, error) {
	var resp JoinRoomResponse
	path := fmt.Sprintf("/_matrix/client/v3/join/%s", url.PathEscape(roomIDOrAlias))
	err := c.doRequest(ctx, "POST", path, map[string]interface{}{}, &resp, true)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// JoinRoomByID joins a room by its internal ID
func (c *Client) JoinRoomByID(ctx context.Context, roomID string) (*JoinRoomResponse, error) {
	var resp JoinRoomResponse
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/join", url.PathEscape(roomID))
	err := c.doRequest(ctx, "POST", path, map[string]interface{}{}, &resp, true)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// LeaveRoom leaves a room
func (c *Client) LeaveRoom(ctx context.Context, roomID string) error {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/leave", url.PathEscape(roomID))
	return c.doRequest(ctx, "POST", path, map[string]interface{}{}, nil, true)
}

// InviteUser invites a user to a room
func (c *Client) InviteUser(ctx context.Context, roomID, userID string) error {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/invite", url.PathEscape(roomID))
	body := map[string]string{"user_id": userID}
	return c.doRequest(ctx, "POST", path, body, nil, true)
}

// KickUser kicks a user from a room
func (c *Client) KickUser(ctx context.Context, roomID, userID, reason string) error {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/kick", url.PathEscape(roomID))
	body := map[string]string{
		"user_id": userID,
		"reason":  reason,
	}
	return c.doRequest(ctx, "POST", path, body, nil, true)
}

// BanUser bans a user from a room
func (c *Client) BanUser(ctx context.Context, roomID, userID, reason string) error {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/ban", url.PathEscape(roomID))
	body := map[string]string{
		"user_id": userID,
		"reason":  reason,
	}
	return c.doRequest(ctx, "POST", path, body, nil, true)
}

// UnbanUser unbans a user from a room
func (c *Client) UnbanUser(ctx context.Context, roomID, userID string) error {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/unban", url.PathEscape(roomID))
	body := map[string]string{"user_id": userID}
	return c.doRequest(ctx, "POST", path, body, nil, true)
}

// GetJoinedRooms returns the list of rooms the user is joined to
func (c *Client) GetJoinedRooms(ctx context.Context) (*JoinedRoomsResponse, error) {
	var resp JoinedRoomsResponse
	err := c.doRequest(ctx, "GET", "/_matrix/client/v3/joined_rooms", nil, &resp, true)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ============================================================
// Messaging
// ============================================================

// SendMessage sends a message to a room
func (c *Client) SendMessage(ctx context.Context, roomID string, req *SendMessageRequest) (*SendMessageResponse, error) {
	var resp SendMessageResponse
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/send/m.room.message", url.PathEscape(roomID))
	err := c.doRequest(ctx, "POST", path, req, &resp, true)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// SendTextMessage sends a text message (convenience method)
func (c *Client) SendTextMessage(ctx context.Context, roomID, body string) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, roomID, &SendMessageRequest{
		MsgType: MsgTypeText,
		Body:    body,
	})
}

// SendHTMLMessage sends an HTML-formatted message
func (c *Client) SendHTMLMessage(ctx context.Context, roomID, body, htmlBody string) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, roomID, &SendMessageRequest{
		MsgType:       MsgTypeText,
		Body:          body,
		Format:        "org.matrix.custom.html",
		FormattedBody: htmlBody,
	})
}

// SendImageMessage sends an image message
func (c *Client) SendImageMessage(ctx context.Context, roomID, body, mxcURL string, info *MediaInfo) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, roomID, &SendMessageRequest{
		MsgType: MsgTypeImage,
		Body:    body,
		URL:     mxcURL,
		Info:    info,
	})
}

// SendFileMessage sends a file message
func (c *Client) SendFileMessage(ctx context.Context, roomID, filename, mxcURL string, info *MediaInfo) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, roomID, &SendMessageRequest{
		MsgType:  MsgTypeFile,
		Body:     filename,
		Filename: filename,
		URL:      mxcURL,
		Info:     info,
	})
}

// SendLocationMessage sends a location message
func (c *Client) SendLocationMessage(ctx context.Context, roomID, body, geoURI string, info *MediaInfo) (*SendMessageResponse, error) {
	return c.SendMessage(ctx, roomID, &SendMessageRequest{
		MsgType: MsgTypeLocation,
		Body:    body,
		GeoURI:  geoURI,
		Info:    info,
	})
}

// SendStateEvent sends a state event
func (c *Client) SendStateEvent(ctx context.Context, roomID, eventType, stateKey string, content map[string]interface{}) error {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/state/%s/%s",
		url.PathEscape(roomID),
		url.PathEscape(eventType),
		url.PathEscape(stateKey))
	return c.doRequest(ctx, "PUT", path, content, nil, true)
}

// ============================================================
// Media
// ============================================================

// UploadMedia uploads media to the homeserver
func (c *Client) UploadMedia(ctx context.Context, data []byte, contentType, filename string) (*UploadResponse, error) {
	path := "/_matrix/media/v3/upload"
	if filename != "" {
		path += "?filename=" + url.QueryEscape(filename)
	}

	fullURL := c.homeserverURL + path
	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed: %d - %s", resp.StatusCode, string(body))
	}

	var result UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// DownloadMedia downloads media from the homeserver
func (c *Client) DownloadMedia(ctx context.Context, mxcURL string) ([]byte, string, error) {
	// Parse mxc:// URL: mxc://server/mediaID
	if len(mxcURL) < 6 || mxcURL[:6] != "mxc://" {
		return nil, "", fmt.Errorf("invalid mxc URL: %s", mxcURL)
	}

	parts := mxcURL[6:] // Remove "mxc://"
	path := "/_matrix/media/v3/download/" + parts

	fullURL := c.homeserverURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("download failed: %d - %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	return data, contentType, nil
}

// ============================================================
// Sync
// ============================================================

// Sync performs a sync request
func (c *Client) Sync(ctx context.Context, since string, timeout int) (*SyncResponse, error) {
	path := "/_matrix/client/v3/sync"
	query := fmt.Sprintf("?timeout=%d", timeout)
	if since != "" {
		query += "&since=" + url.QueryEscape(since)
	}

	var resp SyncResponse
	err := c.doRequest(ctx, "GET", path+query, nil, &resp, true)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ============================================================
// Room State & Members
// ============================================================

// GetRoomState gets all state events for a room
func (c *Client) GetRoomState(ctx context.Context, roomID string) ([]Event, error) {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/state", url.PathEscape(roomID))

	var events []Event
	err := c.doRequest(ctx, "GET", path, nil, &events, true)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// GetRoomStateEvent gets a specific state event
func (c *Client) GetRoomStateEvent(ctx context.Context, roomID, eventType, stateKey string) (*Event, error) {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/state/%s/%s",
		url.PathEscape(roomID),
		url.PathEscape(eventType),
		url.PathEscape(stateKey))

	var event Event
	err := c.doRequest(ctx, "GET", path, nil, &event, true)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

// GetRoomMembers gets all members in a room
func (c *Client) GetRoomMembers(ctx context.Context, roomID string) ([]Event, error) {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/members", url.PathEscape(roomID))

	var resp MembersResponse
	err := c.doRequest(ctx, "GET", path, nil, &resp, true)
	if err != nil {
		return nil, err
	}

	return resp.Chunk, nil
}

// GetRoomMessages gets paginated messages from a room
func (c *Client) GetRoomMessages(ctx context.Context, roomID, from, dir string, limit int) (*MessagesResponse, error) {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/messages?from=%s&dir=%s&limit=%d",
		url.PathEscape(roomID),
		url.QueryEscape(from),
		url.QueryEscape(dir),
		limit)

	var resp MessagesResponse
	err := c.doRequest(ctx, "GET", path, nil, &resp, true)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ============================================================
// Typing Indicators
// ============================================================

// SetTyping sets the typing status for a user in a room
func (c *Client) SetTyping(ctx context.Context, roomID string, typing bool, timeout int64) error {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/typing/%s",
		url.PathEscape(roomID),
		url.PathEscape(c.userID))

	body := &TypingRequest{
		Typing:  typing,
		Timeout: timeout,
	}

	return c.doRequest(ctx, "PUT", path, body, nil, true)
}

// ============================================================
// Read Receipts
// ============================================================

// SendReadReceipt sends a read receipt for an event
func (c *Client) SendReadReceipt(ctx context.Context, roomID, eventID string) error {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/receipt/m.read/%s",
		url.PathEscape(roomID),
		url.PathEscape(eventID))

	return c.doRequest(ctx, "POST", path, map[string]interface{}{}, nil, true)
}

// ============================================================
// Presence
// ============================================================

// SetPresence sets the user's presence
func (c *Client) SetPresence(ctx context.Context, presence, statusMsg string) error {
	path := fmt.Sprintf("/_matrix/client/v3/presence/%s/status", url.PathEscape(c.userID))

	body := &PresenceRequest{
		Presence:  presence,
		StatusMsg: statusMsg,
	}

	return c.doRequest(ctx, "PUT", path, body, nil, true)
}

// GetPresence gets a user's presence
func (c *Client) GetPresence(ctx context.Context, userID string) (*PresenceResponse, error) {
	path := fmt.Sprintf("/_matrix/client/v3/presence/%s/status", url.PathEscape(userID))

	var resp PresenceResponse
	err := c.doRequest(ctx, "GET", path, nil, &resp, true)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ============================================================
// Encryption (E2EE)
// ============================================================

// UploadKeys uploads device keys and one-time keys
func (c *Client) UploadKeys(ctx context.Context, req *UploadKeysRequest) (*UploadKeysResponse, error) {
	var resp UploadKeysResponse
	err := c.doRequest(ctx, "POST", "/_matrix/client/v3/keys/upload", req, &resp, true)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// QueryKeys queries device keys for users
func (c *Client) QueryKeys(ctx context.Context, userIDs []string) (*QueryKeysResponse, error) {
	req := QueryKeysRequest{
		DeviceKeys: make(map[string][]string),
		Timeout:    10000,
	}

	for _, userID := range userIDs {
		req.DeviceKeys[userID] = []string{} // Query all devices
	}

	var resp QueryKeysResponse
	err := c.doRequest(ctx, "POST", "/_matrix/client/v3/keys/query", req, &resp, true)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ============================================================
// Helper Methods
// ============================================================

// doRequest performs an HTTP request to the Matrix homeserver
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}, requireAuth bool) error {
	fullURL := c.homeserverURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if requireAuth && c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var matrixErr MatrixError
		if err := json.Unmarshal(respBody, &matrixErr); err == nil {
			return fmt.Errorf("matrix error: %s", matrixErr.String())
		}
		return fmt.Errorf("http error: %d - %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// GetHomeserverURL returns the homeserver URL
func (c *Client) GetHomeserverURL() string {
	return c.homeserverURL
}

// GetAccessToken returns the current access token
func (c *Client) GetAccessToken() string {
	return c.accessToken
}

// GetUserID returns the current user ID
func (c *Client) GetUserID() string {
	return c.userID
}

// GetDeviceID returns the current device ID
func (c *Client) GetDeviceID() string {
	return c.deviceID
}
