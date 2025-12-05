# TrainBlink Authentication Implementation

## Overview

TrainBlink uses JWT (JSON Web Token) based authentication for both HTTP API and WebSocket connections. The system supports anonymous authentication, allowing mobile apps to obtain tokens without requiring user account creation.

## Architecture

### Components

1. **JWT Service** (`internal/auth/jwt.go`)
   - Token generation and validation
   - Supports both admin and user tokens
   - Configurable TTLs for access and refresh tokens

2. **Authentication Handler** (`internal/api/auth_handler.go`)
   - Anonymous login endpoint
   - Token refresh endpoint
   - User creation and retrieval

3. **JWT Middleware** (`internal/api/middleware/auth.go`)
   - Bearer token validation
   - Request authentication
   - Context injection for user info

4. **Route Protection** (`internal/api/routes.go`)
   - Public routes (no auth required)
   - Protected routes (JWT required)
   - WebSocket token-based authentication

## Authentication Flow

```
┌─────────────┐                    ┌─────────────┐
│ Mobile App  │                    │   Server    │
└──────┬──────┘                    └──────┬──────┘
       │                                  │
       │ POST /api/v1/auth/anonymous     │
       │ {device_id, device_type}        │
       ├─────────────────────────────────>│
       │                                  │
       │  {access_token, refresh_token}  │
       │<─────────────────────────────────┤
       │                                  │
       │ GET /api/v1/messages             │
       │ Authorization: Bearer <token>   │
       ├─────────────────────────────────>│
       │                                  │
       │  {messages: [...]}              │
       │<─────────────────────────────────┤
       │                                  │
```

## API Endpoints

### Anonymous Login

Create an anonymous user and obtain JWT tokens.

**Endpoint:** `POST /api/v1/auth/anonymous`

**Request:**
```json
{
  "device_id": "unique-device-identifier",
  "device_type": "ios" | "android" | "web"
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 3600,
  "user_id": "3db81e4b-7b6c-40dc-bcf6-c8540322bf0d"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request format
- `500 Internal Server Error` - Database or server error

**Notes:**
- `device_id` should be a persistent identifier (e.g., iOS `identifierForVendor`, Android `ANDROID_ID`)
- Anonymous users are created with `firebase_uid` = "anonymous:" + device_id
- Calling with same device_id returns tokens for existing user

---

### Refresh Token

Obtain a new access token using a valid refresh token.

**Endpoint:** `POST /api/v1/auth/refresh`

**Request:**
```json
{
  "refresh_token": "eyJhbGc..."
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGc...",
  "expires_in": 3600
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request format or user_id
- `401 Unauthorized` - Invalid, expired, or non-refresh token

**Notes:**
- Refresh tokens have longer TTL (default: 7 days)
- Access tokens expire after 1 hour (default)
- Only refresh tokens can be used to obtain new access tokens

---

## Using JWT Tokens

### HTTP Requests

All protected endpoints require the `Authorization` header with a Bearer token:

```bash
curl -H "Authorization: Bearer eyJhbGc..." \
  http://localhost:8080/api/v1/messages
```

**Protected Endpoints:**
- All `/api/v1/*` endpoints except `/api/v1/auth/*`
- Examples: `/api/v1/messages`, `/api/v1/trips/start`, `/api/v1/discoveries/log`

**Public Endpoints:**
- `/ping` - Health check
- `/health` - Detailed health status
- `/api/v1/hello` - Test endpoint
- `/api/v1/welcome` - Test endpoint
- `/api/v1/auth/anonymous` - Anonymous login
- `/api/v1/auth/refresh` - Token refresh

### WebSocket Connections

WebSocket connections use token-based authentication via query parameter:

```
ws://localhost:8080/ws?token=eyJhbGc...&station_id=shibuya
```

**Required Parameters:**
- `token` - Valid JWT access token
- `station_id` - Station identifier

**Connection Flow:**
1. Client obtains access token via anonymous login
2. Client connects to WebSocket with token in query param
3. Server validates token before upgrading connection
4. Connection rejected with 401 if token invalid

---

## Token Claims

### User Token Claims

```json
{
  "user_id": "3db81e4b-7b6c-40dc-bcf6-c8540322bf0d",
  "device_id": "test-device-001",
  "token_type": "access" | "refresh",
  "iss": "trainblink-api",
  "exp": 1764956293,
  "iat": 1764952693
}
```

### Admin Token Claims

```json
{
  "admin_id": "admin-uuid",
  "email": "admin@example.com",
  "role": "admin",
  "permissions": ["read", "write", "delete"],
  "token_type": "access" | "refresh",
  "iss": "trainblink-api",
  "exp": 1764956293,
  "iat": 1764952693
}
```

---

## Token Storage

### iOS (Keychain)

```swift
import Security

class TokenManager {
    private let service = "com.trainblink.app"

    func saveToken(_ token: String, forKey key: String) {
        let data = token.data(using: .utf8)!
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecValueData as String: data
        ]

        SecItemDelete(query as CFDictionary)
        SecItemAdd(query as CFDictionary, nil)
    }

    func getToken(forKey key: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        guard status == errSecSuccess,
              let data = result as? Data,
              let token = String(data: data, encoding: .utf8) else {
            return nil
        }

        return token
    }

    func clearTokens() {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service
        ]
        SecItemDelete(query as CFDictionary)
    }
}
```

### Android (EncryptedSharedPreferences)

```kotlin
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey

class TokenManager(context: Context) {
    private val masterKey = MasterKey.Builder(context)
        .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
        .build()

    private val sharedPreferences = EncryptedSharedPreferences.create(
        context,
        "trainblink_secure_prefs",
        masterKey,
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
    )

    fun saveToken(key: String, token: String) {
        sharedPreferences.edit()
            .putString(key, token)
            .apply()
    }

    fun getToken(key: String): String? {
        return sharedPreferences.getString(key, null)
    }

    fun clearTokens() {
        sharedPreferences.edit().clear().apply()
    }
}
```

---

## Security Considerations

### Token Security

1. **Access Token Expiration**: Default 1 hour
   - Short TTL limits damage if token compromised
   - Forces periodic refresh for long-running sessions

2. **Refresh Token Expiration**: Default 7 days
   - Longer TTL for better UX
   - Stored securely in Keychain/EncryptedSharedPreferences

3. **Token Type Validation**
   - Access tokens only for API requests
   - Refresh tokens only for token refresh
   - Prevents token misuse

4. **HTTPS Only (Production)**
   - All tokens transmitted over HTTPS
   - Development: HTTP allowed for localhost
   - Production: Enforce HTTPS

### Best Practices

1. **Token Storage**
   - Never store tokens in UserDefaults/SharedPreferences (unencrypted)
   - Use Keychain (iOS) or EncryptedSharedPreferences (Android)
   - Clear tokens on logout

2. **Token Refresh**
   - Implement automatic token refresh before expiration
   - Handle 401 responses by refreshing token
   - Retry original request with new token

3. **Error Handling**
   - Handle token expiration gracefully
   - Prompt re-authentication if refresh fails
   - Clear invalid tokens immediately

4. **WebSocket Authentication**
   - Validate token before upgrading connection
   - Close connection if token expires during session
   - Implement reconnection with fresh token

---

## Configuration

### Environment Variables

```bash
# JWT Secret Key (REQUIRED)
export JWT_SECRET_KEY="your-secret-key-min-32-chars"

# Token TTLs (Optional, defaults shown)
export JWT_ACCESS_TOKEN_TTL="1h"
export JWT_REFRESH_TOKEN_TTL="168h"  # 7 days
```

### Configuration File

```go
// internal/config/config.go
type JWTConfig struct {
    SecretKey       string        `env:"JWT_SECRET_KEY" required:"true"`
    AccessTokenTTL  time.Duration `env:"JWT_ACCESS_TOKEN_TTL" default:"1h"`
    RefreshTokenTTL time.Duration `env:"JWT_REFRESH_TOKEN_TTL" default:"168h"`
}
```

---

## Testing

### Manual Testing with cURL

```bash
# 1. Anonymous Login
curl -X POST http://localhost:8080/api/v1/auth/anonymous \
  -H "Content-Type: application/json" \
  -d '{"device_id":"test-device","device_type":"cli"}'

# Response:
# {
#   "access_token": "eyJhbGc...",
#   "refresh_token": "eyJhbGc...",
#   "expires_in": 3600,
#   "user_id": "..."
# }

# 2. Test Protected Endpoint (with token)
TOKEN="eyJhbGc..."
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/messages

# Response: 200 OK with data

# 3. Test Protected Endpoint (without token)
curl http://localhost:8080/api/v1/messages

# Response: 401 UNAUTHORIZED

# 4. Refresh Token
REFRESH_TOKEN="eyJhbGc..."
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}"

# Response:
# {
#   "access_token": "eyJhbGc...",
#   "expires_in": 3600
# }
```

### Unit Testing

```go
// internal/auth/jwt_test.go
func TestGenerateUserTokenPair(t *testing.T) {
    jwtService := auth.NewJWTService("test-secret-key", time.Hour, time.Hour*24*7)
    userID := uuid.New()
    deviceID := "test-device"

    tokenPair, err := jwtService.GenerateUserTokenPair(userID, deviceID)
    assert.NoError(t, err)
    assert.NotEmpty(t, tokenPair.AccessToken)
    assert.NotEmpty(t, tokenPair.RefreshToken)
    assert.Equal(t, int64(3600), tokenPair.ExpiresIn)
}

func TestValidateToken(t *testing.T) {
    jwtService := auth.NewJWTService("test-secret-key", time.Hour, time.Hour*24*7)
    userID := uuid.New()
    deviceID := "test-device"

    tokenPair, _ := jwtService.GenerateUserTokenPair(userID, deviceID)

    claims, err := jwtService.ValidateToken(tokenPair.AccessToken)
    assert.NoError(t, err)
    assert.Equal(t, userID.String(), claims.UserID)
    assert.Equal(t, deviceID, claims.DeviceID)
    assert.Equal(t, "access", claims.TokenType)
}
```

---

## Troubleshooting

### Common Issues

**Issue: "Missing Authorization header"**
- **Cause**: No Authorization header in request
- **Solution**: Add `Authorization: Bearer <token>` header

**Issue: "Invalid Authorization header format"**
- **Cause**: Incorrect header format (missing "Bearer " prefix)
- **Solution**: Ensure header is `Authorization: Bearer <token>`

**Issue: "Token has expired"**
- **Cause**: Access token expired (> 1 hour old)
- **Solution**: Use refresh token to obtain new access token

**Issue: "Invalid token type. Use access token for API requests"**
- **Cause**: Attempting to use refresh token for API requests
- **Solution**: Use access token, reserve refresh token for `/auth/refresh`

**Issue: "Invalid or expired refresh token"**
- **Cause**: Refresh token expired (> 7 days old) or invalid
- **Solution**: Re-authenticate with `/auth/anonymous`

---

## Implementation Details

### Database Schema

```sql
-- Users table (migrations/001_create_users.up.sql)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    firebase_uid VARCHAR(255) UNIQUE NOT NULL,  -- "anonymous:<device_id>"
    matrix_user_id VARCHAR(255),
    display_name VARCHAR(255),
    avatar_emoji VARCHAR(10),
    avatar_color VARCHAR(7),
    status_text TEXT,
    preferences JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    is_banned BOOLEAN DEFAULT false,
    ban_reason TEXT,
    banned_at TIMESTAMP,
    banned_by UUID,
    ban_until TIMESTAMP,
    total_sessions INTEGER DEFAULT 0,
    total_encounters INTEGER DEFAULT 0,
    total_messages_sent INTEGER DEFAULT 0,
    total_content_shared INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP,
    data_retention_days INTEGER DEFAULT 1
);

CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);
CREATE INDEX idx_users_matrix_user_id ON users(matrix_user_id);
CREATE INDEX idx_users_is_banned ON users(is_banned);
CREATE INDEX idx_users_last_seen_at ON users(last_seen_at);
```

### Anonymous User Creation

Anonymous users are created with a special `firebase_uid` format:

```
firebase_uid = "anonymous:" + device_id
```

This allows:
- Persistent user identity across app reinstalls (same device_id)
- Future migration to Firebase authentication (replace firebase_uid)
- No user registration required

---

## Future Enhancements

### Planned Features

1. **Token Revocation**
   - Implement token blacklist in Redis
   - Invalidate tokens on logout
   - Admin endpoint to revoke user tokens

2. **Firebase Authentication Integration**
   - Support Firebase Auth in addition to anonymous auth
   - Migrate anonymous users to Firebase users
   - Link multiple devices to single Firebase account

3. **OAuth Integration**
   - Google Sign-In
   - Apple Sign-In
   - Email/Password authentication

4. **Rate Limiting**
   - Limit login attempts per device
   - Prevent token farming
   - DDoS protection

5. **Session Management**
   - Track active sessions per user
   - Force logout from other devices
   - Session expiration policies

---

## References

- **JWT Specification**: https://tools.ietf.org/html/rfc7519
- **GORM Documentation**: https://gorm.io/docs/
- **Gin Framework**: https://gin-gonic.com/docs/
- **iOS Keychain**: https://developer.apple.com/documentation/security/keychain_services
- **Android EncryptedSharedPreferences**: https://developer.android.com/reference/androidx/security/crypto/EncryptedSharedPreferences

---

## Support

For issues or questions about authentication:
- Check logs in server output
- Review this documentation
- Test with cURL to isolate client vs server issues
- Report bugs at: https://github.com/anthropics/claude-code/issues
