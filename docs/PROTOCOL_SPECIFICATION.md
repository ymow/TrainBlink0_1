# TrainBlink 通訊協議規範

**版本**: 1.0  
**最後更新**: 2025-11-20  
**狀態**: 規範制定中

---

## 📖 概述

本文檔定義了 TrainBlink 通訊協議的詳細規範，涵蓋從 Phase 0 到 Phase 7 的所有協議演進階段。

### 協議版本歷史

| 版本 | 階段 | 發布日期 | 主要特性 |
|------|------|---------|----------|
| v0.1 | Phase 0 | 2025-11-20 | HTTP REST API |
| v0.2 | Phase 1 | TBD | + WebSocket |
| v0.3 | Phase 2 | TBD | + JWT + 持久化 |
| v0.4 | Phase 3 | TBD | + 車站系統 |
| v1.0 | Phase 4 | TBD | + Matrix 協議 |
| v1.1 | Phase 5 | TBD | + MLS 加密 |
| v1.2 | Phase 6 | TBD | + WebTransport |
| v2.0 | Phase 7 | TBD | 完整功能 |

---

## Phase 0: HTTP REST API 協議 (v0.1) ✅

### 連接建立

#### 1. 客戶端註冊

**Endpoint**: `POST /api/v1/clients`

**Request**:
```json
{
  "name": "Alice",           // 顯示名稱 (必填，1-50 字符)
  "device_id": "iOS-1234"    // 設備 ID (可選)
}
```

**Response** (201 Created):
```json
{
  "status": "registered",
  "client": {
    "id": "client-1763630597712548056",
    "name": "Alice",
    "device_id": "iOS-1234",
    "connected_at": "2025-11-20T09:23:17.712Z",
    "last_seen": "2025-11-20T09:23:17.712Z"
  },
  "message": "Welcome Alice! Your client ID is client-1763630597712548056"
}
```

**錯誤響應**:
```json
// 400 Bad Request - 缺少必填字段
{
  "error": "name is required",
  "code": "MISSING_FIELD"
}

// 400 Bad Request - 無效的 JSON
{
  "error": "Invalid JSON",
  "code": "INVALID_JSON"
}
```

#### 2. 獲取客戶端列表

**Endpoint**: `GET /api/v1/clients`

**Response** (200 OK):
```json
{
  "count": 2,
  "clients": [
    {
      "id": "client-1763630597712548056",
      "name": "Alice",
      "device_id": "iOS-1234",
      "connected_at": "2025-11-20T09:23:17.712Z",
      "last_seen": "2025-11-20T09:23:17.712Z"
    },
    {
      "id": "client-1763630597868313260",
      "name": "Bob",
      "device_id": "Android-5678",
      "connected_at": "2025-11-20T09:23:17.868Z",
      "last_seen": "2025-11-20T09:23:17.868Z"
    }
  ],
  "timestamp": "2025-11-20T09:23:18.027Z"
}
```

### 消息傳輸

#### 1. 發送點對點消息

**Endpoint**: `POST /api/v1/messages`

**Request**:
```json
{
  "from": "client-1763630597712548056",     // 發送者 ID (必填)
  "to": "client-1763630597868313260",       // 接收者 ID (必填)
  "text": "Hello Bob, this is Alice!"       // 消息內容 (必填，1-5000 字符)
}
```

**Response** (201 Created):
```json
{
  "status": "sent",
  "message": {
    "id": "msg-1763630598105918581",
    "from": "client-1763630597712548056",
    "to": "client-1763630597868313260",
    "text": "Hello Bob, this is Alice!",
    "timestamp": "2025-11-20T09:23:18.106Z",
    "status": "sent"
  }
}
```

**錯誤響應**:
```json
// 400 Bad Request - 發送者不存在
{
  "error": "Sender client client-123 not found",
  "code": "SENDER_NOT_FOUND"
}

// 404 Not Found - 接收者不存在
{
  "error": "Recipient client client-456 not found",
  "code": "RECIPIENT_NOT_FOUND"
}

// 400 Bad Request - 消息過長
{
  "error": "Text exceeds maximum length (5000 characters)",
  "code": "TEXT_TOO_LONG"
}
```

#### 2. 發送廣播消息

**Endpoint**: `POST /api/v1/messages`

**Request**:
```json
{
  "from": "system",                        // 可以是 "system" 或任何 client ID
  "to": "all",                             // 廣播到所有人
  "text": "Welcome everyone to TrainBlink!"
}
```

**Response** (201 Created):
```json
{
  "status": "sent",
  "message": {
    "id": "msg-1763630598263961631",
    "from": "system",
    "to": "all",
    "text": "Welcome everyone to TrainBlink!",
    "timestamp": "2025-11-20T09:23:18.264Z",
    "status": "sent"
  }
}
```

#### 3. 獲取用戶消息

**Endpoint**: `GET /api/v1/messages/user?id={client_id}`

**Response** (200 OK):
```json
{
  "user_id": "client-1763630597712548056",
  "count": 3,
  "messages": [
    {
      "id": "msg-1763630598105918581",
      "from": "client-1763630597712548056",
      "to": "client-1763630597868313260",
      "text": "Hello Bob, this is Alice!",
      "timestamp": "2025-11-20T09:23:18.106Z",
      "status": "sent"
    },
    {
      "id": "msg-1763630598190188696",
      "from": "client-1763630597868313260",
      "to": "client-1763630597712548056",
      "text": "Hi Alice, nice to meet you!",
      "timestamp": "2025-11-20T09:23:18.190Z",
      "status": "sent"
    },
    {
      "id": "msg-1763630598263961631",
      "from": "system",
      "to": "all",
      "text": "Welcome everyone to TrainBlink!",
      "timestamp": "2025-11-20T09:23:18.264Z",
      "status": "sent"
    }
  ],
  "timestamp": "2025-11-20T09:23:20.000Z"
}
```

**過濾邏輯**:
- 返回用戶發送的消息 (`from` = user_id)
- 返回發給用戶的消息 (`to` = user_id)
- 返回廣播消息 (`to` = "all")

### 健康檢查

#### 1. Ping

**Endpoint**: `GET /ping`

**Response** (200 OK):
```json
{
  "message": "pong",
  "timestamp": "2025-11-20T09:23:17.000Z",
  "server": "TrainBlink Server v0.1.0"
}
```

#### 2. 詳細健康狀態

**Endpoint**: `GET /health`

**Response** (200 OK):
```json
{
  "status": "healthy",
  "timestamp": "2025-11-20T09:23:17.000Z",
  "uptime": 3600.5,              // 秒
  "clients": 15,                 // 當前在線客戶端數
  "messages": 234,               // 總消息數
  "total_requests": 1523         // 總請求數
}
```

### CORS 支持

所有 API 端點都支持 CORS，允許跨域請求：

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

---

## Phase 1: WebSocket 協議 (v0.2)

### 連接建立

#### 1. WebSocket 握手

**URL**: `ws://localhost:8080/ws?client_id={client_id}`

**Query Parameters**:
- `client_id` (必填): 客戶端 ID（通過 HTTP API 註冊獲得）

**連接流程**:
```
1. 客戶端通過 HTTP 註冊 → 獲得 client_id
2. 客戶端使用 client_id 建立 WebSocket 連接
3. Server 驗證 client_id 有效性
4. 連接建立成功 → 發送歡迎消息
5. 開始心跳（每 30 秒）
```

#### 2. 歡迎消息

**Server → Client** (連接建立時):
```json
{
  "type": "welcome",
  "payload": {
    "client_id": "client-123",
    "server_time": "2025-11-20T10:00:00Z",
    "session_id": "session-456",
    "heartbeat_interval": 30
  }
}
```

#### 3. 心跳機制

**Client → Server** (每 30 秒):
```json
{
  "type": "ping",
  "timestamp": "2025-11-20T10:00:30Z"
}
```

**Server → Client**:
```json
{
  "type": "pong",
  "timestamp": "2025-11-20T10:00:30Z"
}
```

**超時處理**:
- 如果 60 秒內沒有收到心跳 → Server 關閉連接
- Client 收到關閉 → 自動重連（退避重試：1s, 2s, 4s, 8s, 16s, 30s max）

### 消息格式

#### 基礎消息結構

所有 WebSocket 消息都遵循以下結構：

```json
{
  "type": "message_type",      // 消息類型
  "payload": {                 // 消息負載（根據類型不同）
    ...
  },
  "timestamp": "ISO8601",      // 可選：客戶端時間戳
  "id": "unique-id"            // 可選：消息 ID（用於追蹤）
}
```

#### 1. 發送消息

**Client → Server**:
```json
{
  "type": "send_message",
  "payload": {
    "to": "client-456",                    // 接收者 ID 或 "all"
    "text": "Hello!",                      // 消息內容
    "is_ephemeral": false,                 // 是否臨時消息
    "lifetime_seconds": 10                 // 如果是臨時消息，生命週期（秒）
  },
  "id": "client-msg-789"                   // 客戶端生成的消息 ID
}
```

**Server → Client (ACK)**:
```json
{
  "type": "message_ack",
  "payload": {
    "client_message_id": "client-msg-789",
    "server_message_id": "msg-1763630598105918581",
    "status": "sent",
    "timestamp": "2025-11-20T10:00:00Z"
  }
}
```

**Server → Recipient**:
```json
{
  "type": "new_message",
  "payload": {
    "id": "msg-1763630598105918581",
    "from": "client-123",
    "from_name": "Alice",                  // 可選：發送者名稱
    "to": "client-456",
    "text": "Hello!",
    "timestamp": "2025-11-20T10:00:00Z",
    "is_ephemeral": false,
    "expires_at": null
  }
}
```

#### 2. 送達回執

**Client → Server** (收到消息後):
```json
{
  "type": "message_delivered",
  "payload": {
    "message_id": "msg-1763630598105918581"
  }
}
```

**Server → Original Sender**:
```json
{
  "type": "delivery_receipt",
  "payload": {
    "message_id": "msg-1763630598105918581",
    "delivered_to": "client-456",
    "delivered_at": "2025-11-20T10:00:05Z"
  }
}
```

#### 3. 已讀回執

**Client → Server** (閱讀消息後):
```json
{
  "type": "message_read",
  "payload": {
    "message_id": "msg-1763630598105918581"
  }
}
```

**Server → Original Sender**:
```json
{
  "type": "read_receipt",
  "payload": {
    "message_id": "msg-1763630598105918581",
    "read_by": "client-456",
    "read_at": "2025-11-20T10:00:10Z"
  }
}
```

#### 4. 輸入狀態

**Client → Server** (開始/停止輸入):
```json
{
  "type": "typing",
  "payload": {
    "to": "client-456",        // 對方 ID
    "is_typing": true          // true = 開始輸入, false = 停止輸入
  }
}
```

**Server → Recipient**:
```json
{
  "type": "typing_indicator",
  "payload": {
    "from": "client-123",
    "from_name": "Alice",
    "is_typing": true
  }
}
```

**輸入狀態規則**:
- 客戶端在開始輸入時發送 `is_typing: true`
- 在停止輸入時發送 `is_typing: false`
- Server 不存儲輸入狀態，僅轉發
- 如果 5 秒內沒有收到新的輸入狀態，自動視為停止

#### 5. 在線狀態

**Server → All Clients** (當用戶上線時):
```json
{
  "type": "user_online",
  "payload": {
    "client_id": "client-789",
    "client_name": "Charlie",
    "timestamp": "2025-11-20T10:00:00Z"
  }
}
```

**Server → All Clients** (當用戶離線時):
```json
{
  "type": "user_offline",
  "payload": {
    "client_id": "client-789",
    "client_name": "Charlie",
    "timestamp": "2025-11-20T10:05:00Z",
    "reason": "disconnected"    // disconnected, timeout, error
  }
}
```

### 錯誤處理

#### 錯誤消息格式

**Server → Client**:
```json
{
  "type": "error",
  "payload": {
    "code": "INVALID_RECIPIENT",
    "message": "Recipient client-456 not found",
    "details": {
      "attempted_recipient": "client-456"
    },
    "original_message_id": "client-msg-789"  // 如果有關聯消息
  }
}
```

#### 錯誤代碼

| 錯誤代碼 | 說明 | HTTP 等價 |
|---------|------|-----------|
| `INVALID_CLIENT_ID` | 客戶端 ID 無效或不存在 | 400 |
| `INVALID_RECIPIENT` | 接收者不存在 | 404 |
| `MESSAGE_TOO_LONG` | 消息超過最大長度 | 400 |
| `RATE_LIMIT_EXCEEDED` | 發送速率過快 | 429 |
| `UNAUTHORIZED` | 未授權（Phase 2+） | 401 |
| `SERVER_ERROR` | 服務器內部錯誤 | 500 |
| `INVALID_MESSAGE_FORMAT` | 消息格式錯誤 | 400 |

#### 連接關閉碼

| 關閉碼 | 說明 | 是否可重連 |
|-------|------|-----------|
| 1000 | 正常關閉 | ✅ |
| 1001 | 客戶端離開 | ✅ |
| 1002 | 協議錯誤 | ❌ |
| 1003 | 不支持的數據類型 | ❌ |
| 1006 | 異常關閉 | ✅ |
| 1008 | 違反政策 | ❌ |
| 1011 | 服務器錯誤 | ✅ |
| 4000 | 無效的 client_id | ❌ |
| 4001 | 認證失敗 | ❌ |
| 4002 | 會話過期 | ✅ |

### 重連策略

#### 自動重連流程

```
1. WebSocket 連接斷開
2. 清理本地狀態（心跳計時器等）
3. 等待退避時間（指數退避）
4. 嘗試重連
5. 如果成功 → 重新訂閱/同步狀態
6. 如果失敗 → 增加退避時間，返回步驟 3
```

#### 退避時間表

| 重試次數 | 等待時間 |
|---------|---------|
| 1 | 1 秒 |
| 2 | 2 秒 |
| 3 | 4 秒 |
| 4 | 8 秒 |
| 5 | 16 秒 |
| 6+ | 30 秒 (最大) |

#### 重連成功後

**Client → Server**:
```json
{
  "type": "reconnect",
  "payload": {
    "last_message_id": "msg-123",    // 最後收到的消息 ID
    "timestamp": "2025-11-20T09:55:00Z"
  }
}
```

**Server → Client** (同步缺失的消息):
```json
{
  "type": "sync",
  "payload": {
    "messages": [
      {
        "id": "msg-124",
        "from": "client-456",
        "to": "client-123",
        "text": "Are you there?",
        "timestamp": "2025-11-20T09:56:00Z"
      }
    ],
    "events": [
      {
        "type": "user_offline",
        "payload": {...}
      }
    ]
  }
}
```

---

## Phase 2: JWT 認證協議 (v0.3)

### 認證流程

#### 1. 匿名註冊

**Endpoint**: `POST /api/v1/auth/anonymous`

**Request**:
```json
{
  "name": "Alice",              // 顯示名稱 (必填)
  "device_id": "iOS-1234",      // 設備 ID (可選)
  "station_id": "1001"          // 當前車站 ID (可選，Phase 3 必填)
}
```

**Response** (201 Created):
```json
{
  "status": "success",
  "client": {
    "id": "550e8400-e29b-41d4-a716-446655440000",  // UUID v4
    "name": "Alice",
    "device_id": "iOS-1234",
    "station_id": "1001",
    "created_at": "2025-11-20T10:00:00Z",
    "expires_at": "2025-11-20T22:00:00Z"           // 12 小時後過期
  },
  "auth": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 43200,                            // 秒（12 小時）
    "refresh_token": "refresh-token-xyz"            // 用於刷新 token
  }
}
```

#### 2. JWT Token 結構

**Header**:
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

**Payload**:
```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",   // client_id
  "name": "Alice",
  "device_id": "iOS-1234",
  "station_id": "1001",
  "iat": 1700000000,                               // issued at
  "exp": 1700043200,                               // expires at (12 小時後)
  "jti": "unique-token-id"                         // token ID
}
```

**Signature**:
```
HMACSHA256(
  base64UrlEncode(header) + "." +
  base64UrlEncode(payload),
  secret
)
```

#### 3. 刷新 Token

**Endpoint**: `POST /api/v1/auth/refresh`

**Request**:
```json
{
  "refresh_token": "refresh-token-xyz"
}
```

**Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 43200
}
```

### 使用 Token

#### HTTP 請求

**Header**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**示例**:
```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{"to":"client-456","text":"Hello!"}'
```

#### WebSocket 連接

**方法 1: Query Parameter**
```
ws://localhost:8080/ws?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**方法 2: 連接後認證**
```javascript
// 1. 建立連接
ws = new WebSocket('ws://localhost:8080/ws')

// 2. 發送認證消息
ws.send(JSON.stringify({
  type: 'auth',
  payload: {
    token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
  }
}))

// 3. 等待認證響應
{
  "type": "auth_success",
  "payload": {
    "client_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Alice"
  }
}
```

### 錯誤處理

**401 Unauthorized**:
```json
{
  "error": "unauthorized",
  "message": "Invalid or expired token",
  "code": "INVALID_TOKEN"
}
```

**403 Forbidden**:
```json
{
  "error": "forbidden",
  "message": "You do not have permission to access this resource",
  "code": "INSUFFICIENT_PERMISSIONS"
}
```

---

## Phase 3: 車站系統協議 (v0.4)

### 車站管理

#### 1. 獲取車站列表

**Endpoint**: `GET /api/v1/stations`

**Query Parameters**:
- `lat` (可選): 緯度
- `lng` (可選): 經度
- `radius` (可選): 搜索半徑（米），默認 5000

**Response** (200 OK):
```json
{
  "count": 34,
  "stations": [
    {
      "id": "1001",
      "name": "台北車站",
      "name_en": "Taipei",
      "lat": 25.0478,
      "lng": 121.5170,
      "type": "TRA",
      "radius": 500,
      "lines": ["西部幹線", "東部幹線"],
      "active_users": 15,         // 當前在線用戶數
      "distance": 1234            // 如果提供了 lat/lng，返回距離（米）
    }
  ]
}
```

#### 2. 獲取車站詳情

**Endpoint**: `GET /api/v1/stations/{station_id}`

**Response** (200 OK):
```json
{
  "id": "1001",
  "name": "台北車站",
  "name_en": "Taipei",
  "lat": 25.0478,
  "lng": 121.5170,
  "type": "TRA",
  "radius": 500,
  "lines": ["西部幹線", "東部幹線"],
  "active_users": 15,
  "metadata": {
    "address": "台北市中正區北平西路3號",
    "postal_code": "100"
  }
}
```

#### 3. 進入車站

**Endpoint**: `POST /api/v1/stations/{station_id}/join`

**Request**:
```json
{
  "lat": 25.0478,               // 當前位置緯度
  "lng": 121.5170               // 當前位置經度
}
```

**Response** (200 OK):
```json
{
  "status": "joined",
  "station": {
    "id": "1001",
    "name": "台北車站"
  },
  "room": {
    "room_id": "room-taipei-1001",
    "active_users": 15,
    "joined_at": "2025-11-20T10:00:00Z"
  }
}
```

**地理圍欄驗證**:
- Server 計算客戶端位置與車站位置的距離
- 如果距離 > 車站半徑 → 拒絕進入
- 拒絕響應:
```json
{
  "error": "out_of_range",
  "message": "You are 1234 meters away from the station",
  "code": "GEOFENCE_VIOLATION",
  "details": {
    "distance": 1234,
    "max_distance": 500
  }
}
```

#### 4. 離開車站

**Endpoint**: `POST /api/v1/stations/{station_id}/leave`

**Response** (200 OK):
```json
{
  "status": "left",
  "station": {
    "id": "1001",
    "name": "台北車站"
  },
  "cleanup": {
    "messages_deleted": 12,
    "data_cleared": true
  }
}
```

**自動清理**:
- 刪除該用戶在該車站的所有消息
- 清理臨時文件
- 通知其他用戶該用戶已離線

#### 5. 獲取車站內用戶

**Endpoint**: `GET /api/v1/stations/{station_id}/peers`

**Response** (200 OK):
```json
{
  "station_id": "1001",
  "count": 15,
  "peers": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "display_name": "User-1234",
      "discovered_at": "2025-11-20T09:50:00Z",
      "last_seen_at": "2025-11-20T10:00:00Z",
      "connection_state": "CONNECTED"
    }
  ]
}
```

### WebSocket 車站事件

#### 1. 用戶進入車站

**Server → All Peers in Station**:
```json
{
  "type": "peer_joined_station",
  "payload": {
    "peer": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "display_name": "User-1234",
      "joined_at": "2025-11-20T10:00:00Z"
    },
    "station": {
      "id": "1001",
      "name": "台北車站"
    },
    "active_users": 16
  }
}
```

#### 2. 用戶離開車站

**Server → All Peers in Station**:
```json
{
  "type": "peer_left_station",
  "payload": {
    "peer_id": "550e8400-e29b-41d4-a716-446655440000",
    "station": {
      "id": "1001",
      "name": "台北車站"
    },
    "left_at": "2025-11-20T10:30:00Z",
    "active_users": 15
  }
}
```

---

## 消息狀態機

### 狀態轉換

```
PENDING → SENDING → SENT → DELIVERED → READ
   ↓          ↓       ↓         ↓
FAILED ←─────┴───────┴─────────┘
```

### 狀態定義

| 狀態 | 說明 | 觸發條件 |
|------|------|---------|
| `PENDING` | 消息創建，等待發送 | 客戶端創建消息 |
| `SENDING` | 正在發送到服務器 | 開始網絡請求 |
| `SENT` | 服務器已接收 | 服務器返回 ACK |
| `DELIVERED` | 已送達接收者 | 接收者確認收到 |
| `READ` | 接收者已閱讀 | 接收者打開消息 |
| `FAILED` | 發送失敗 | 網絡錯誤、驗證失敗等 |

---

## 錯誤處理規範

### HTTP 錯誤碼

| 狀態碼 | 說明 | 使用場景 |
|-------|------|---------|
| 200 | OK | 成功獲取資源 |
| 201 | Created | 成功創建資源 |
| 400 | Bad Request | 請求格式錯誤 |
| 401 | Unauthorized | 未認證 |
| 403 | Forbidden | 無權限 |
| 404 | Not Found | 資源不存在 |
| 429 | Too Many Requests | 速率限制 |
| 500 | Internal Server Error | 服務器錯誤 |
| 503 | Service Unavailable | 服務不可用 |

### 錯誤響應格式

```json
{
  "error": "error_type",           // 錯誤類型
  "message": "Human readable message",
  "code": "ERROR_CODE",            // 機器可讀的錯誤代碼
  "details": {                     // 可選：詳細信息
    "field": "email",
    "reason": "invalid format"
  },
  "timestamp": "2025-11-20T10:00:00Z"
}
```

---

**下一部分**: Phase 4-7 協議規範（Matrix、MLS、WebTransport）將在後續更新。

---

**文檔維護者**: TrainBlink Team  
**最後更新**: 2025-11-20  
**版本**: 1.0
