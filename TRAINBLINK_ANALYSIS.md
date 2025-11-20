# TrainBlink 系統分析報告

**分析日期**: 2025-11-20
**原始倉庫**: https://github.com/ymow/trainblink
**目的**: 為實作 Go Server 做準備

---

## 1. 專案概述

### 1.1 什麼是 TrainBlink？

TrainBlink 是一個針對火車通勤者的 1對1 匿名社交 App，主要特點：

- **地理圍欄**：僅在火車站內啟動（34 個台灣車站）
- **P2P 連接**：無需服務器，使用 MultipeerConnectivity / Nearby Connections
- **完全匿名**：無註冊、臨時 UUID、無數據持久化
- **AI 安全**：NSFW 檢測、人臉檢測
- **自動清理**：離開車站時自動刪除所有數據

### 1.2 技術棧

#### iOS (主要實現)
- **語言**: Swift 5.9+
- **UI**: SwiftUI
- **最低版本**: iOS 15.0+
- **架構**: MVVM + Combine
- **P2P**: MultipeerConnectivity (Bluetooth LE + WiFi Direct)
- **AI**: Vision + CoreML
- **位置**: CoreLocation
- **監控**: Firebase (Analytics, Crashlytics, Performance)

#### Android (進行中)
- **語言**: Kotlin 1.9.20
- **UI**: Jetpack Compose + Material3
- **最低版本**: Android 8.0 (SDK 26)
- **P2P**: Google Nearby Connections API
- **AI**: ML Kit

---

## 2. 核心數據模型

### 2.1 Station (車站)

```json
{
  "id": "1001",
  "name": "台北車站",
  "name_en": "Taipei",
  "lat": 25.0478,
  "lng": 121.5170,
  "type": "TRA",
  "radius": 500,
  "lines": ["西部幹線", "東部幹線"]
}
```

**字段說明**：
- `id`: 車站唯一標識
- `name`: 繁體中文名稱
- `name_en`: 英文名稱
- `lat/lng`: GPS 坐標（精確到 4 位小數）
- `type`: 車站類型（TRA=台鐵, THSR=高鐵, MRT=捷運）
- `radius`: 地理圍欄半徑（米）
- `lines`: 所屬路線

**統計**：
- 總共 34 個車站（22 TRA + 12 THSR）
- 默認半徑：500 米

### 2.2 Peer (對等節點)

```kotlin
data class Peer(
    val id: String,                              // UUID
    val displayName: String,                     // 顯示名稱
    val connectionState: PeerConnectionState,    // 連接狀態
    val discoveredAt: Date,                      // 發現時間
    var lastSeenAt: Date,                        // 最後看到時間
    var signalStrength: Double?,                 // 信號強度
    var metadata: Map<String, String>?,          // 元數據
    var endpointId: String?                      // Nearby Connections ID
)

enum class PeerConnectionState {
    NOT_CONNECTED,
    CONNECTING,
    CONNECTED
}
```

**關鍵邏輯**：
- 60 秒未見視為過期（stale）
- 最多顯示 20 個附近節點
- 支持信號強度排序

### 2.3 ChatMessage (聊天消息)

```kotlin
data class ChatMessage(
    val id: String,                              // UUID
    val text: String,                            // 消息內容
    val senderId: String,                        // 發送者 ID
    val receiverId: String,                      // 接收者 ID
    val timestamp: Date,                         // 時間戳
    var isEncrypted: Boolean = false,            // 是否加密
    var isRead: Boolean = false,                 // 是否已讀
    var deliveryStatus: MessageDeliveryStatus,   // 送達狀態
    var deliveredAt: Date?,                      // 送達時間
    var readAt: Date?,                           // 已讀時間
    // 臨時消息 (Feature 6)
    var isEphemeral: Boolean = false,            // 是否臨時
    var expiresAt: Date?,                        // 過期時間
    var isExpired: Boolean = false               // 是否已過期
)

enum class MessageDeliveryStatus {
    PENDING,    // 等待中
    SENDING,    // 發送中
    SENT,       // 已發送
    DELIVERED,  // 已送達
    READ,       // 已讀
    FAILED      // 失敗
}
```

**重要特性**：
- **臨時消息**：默認 10 秒自動刪除
- **送達狀態**：6 種狀態追蹤
- **已讀回執**：支持已讀時間戳

### 2.4 ContentItem (分享內容)

```kotlin
data class ContentItem(
    val id: String,                              // UUID
    val type: ContentType,                       // 內容類型
    var state: ContentState,                     // 內容狀態
    val createdAt: Date,                         // 創建時間
    var textContent: String?,                    // 文本內容
    var imageData: ByteArray?,                   // 圖片數據
    var originalImageSize: Int?,                 // 原始大小
    var compressedImageSize: Int?,               // 壓縮後大小
    var fileSize: Int,                           // 文件大小
    var progress: Double,                        // 傳輸進度
    var aiReviewedAt: Date?,                     // AI 審核時間
    var aiRejectionReason: AIRejectionReason?,   // 拒絕原因
    val senderId: String,                        // 發送者 ID
    var recipientId: String?                     // 接收者 ID
)

enum class ContentType {
    PHOTO,
    TEXT
}

enum class ContentState {
    PENDING,     // 等待中
    REVIEWING,   // AI 審核中
    APPROVED,    // 已批准
    REJECTED,    // 已拒絕
    SENDING,     // 發送中
    SENT,        // 已發送
    FAILED       // 失敗
}

enum class AIRejectionReason {
    NSFW,              // 不適當內容
    FACE_DETECTED,     // 檢測到人臉
    INAPPROPRIATE      // 違反準則
}
```

**AI 安全機制**：
- **NSFW 檢測**：置信度 >= 0.3 自動拒絕
- **人臉檢測**：檢測到人臉時警告用戶
- **圖片壓縮**：最大 10MB，自動壓縮（0.7 質量）

### 2.5 Encounter (相遇記錄)

```kotlin
data class Encounter(
    val id: String,                              // UUID
    val peerId: String,                          // 對等節點 ID
    val timestamp: Date,                         // 時間戳
    val stationId: String,                       // 車站 ID
    val stationName: String,                     // 車站名稱
    val interactionType: InteractionType         // 互動類型
)

enum class InteractionType {
    DISCOVERY,       // 發現
    CHAT,            // 聊天
    CONTENT_SHARING  // 內容分享
}
```

**統計功能**：
- 相遇次數
- 最常見車站
- 連續天數追蹤
- 頻率標籤（每日/每週/偶爾）

### 2.6 ChatRoom (聊天室)

```kotlin
data class ChatRoom(
    val id: String,                              // UUID
    val peerId: String,                          // 對等節點 ID
    var messages: List<ChatMessage>,             // 消息列表
    var isActive: Boolean,                       // 是否活躍
    var unreadCount: Int,                        // 未讀計數
    var encounterCount: Int                      // 相遇次數
)
```

**限制**：
- 最多 500 條消息/聊天室
- 30 秒發送超時
- 離開車站時自動清理

---

## 3. 通訊協議

### 3.1 P2P 發現協議

**iOS - MultipeerConnectivity**:
- **Service Type**: `trainblink-chat` (最多 15 字符，小寫，僅連字符)
- **傳輸**: Bluetooth LE + WiFi Direct
- **加密**: MCSession encryptionPreference = `.required`
- **範圍**: 車站內（通過地理圍欄控制）

**Discovery Info** (元數據):
```swift
[
  "version": "1.0",
  "platform": "ios"
]
```

**連接流程**:
1. 進入車站 → 啟動 Advertiser + Browser
2. 發現節點 → 發送邀請（30 秒超時）
3. 接受邀請 → 建立 MCSession
4. 離開車站 → 斷開所有連接

**Android - Nearby Connections**:
- **策略**: P2P_CLUSTER
- **傳輸**: 類似 MultipeerConnectivity
- **加密**: 自動加密

### 3.2 消息傳輸協議

**傳輸方式**:
- MultipeerConnectivity MCSession
- JSON 編碼
- Base64 編碼圖片數據

**消息結構** (推測):
```json
{
  "type": "chat_message",
  "id": "uuid",
  "senderId": "peer-id",
  "receiverId": "peer-id",
  "text": "Hello",
  "timestamp": 1700000000,
  "isEphemeral": false,
  "deliveryStatus": "sent"
}
```

**內容分享結構** (推測):
```json
{
  "type": "content_item",
  "id": "uuid",
  "contentType": "photo",
  "senderId": "peer-id",
  "imageData": "base64...",
  "fileSize": 102400,
  "timestamp": 1700000000
}
```

### 3.3 狀態同步

**節點狀態更新**:
- 每 60 秒清理過期節點
- 信號強度實時更新
- 連接狀態變化立即通知

**消息狀態同步**:
- PENDING → SENDING → SENT → DELIVERED → READ
- 失敗時標記為 FAILED
- 臨時消息倒計時

---

## 4. 系統架構

### 4.1 整體架構

```
┌─────────────────────────────────────────────────────┐
│                   SwiftUI Views                     │
│              (ContentView, ChatView)                │
└────────────────────┬────────────────────────────────┘
                     │ @EnvironmentObject
                     ↓
┌─────────────────────────────────────────────────────┐
│                   AppState                          │
│              (@ObservableObject)                    │
│     - currentStation                                │
│     - isInStation                                   │
│     - discoveredPeers                               │
└────────────────────┬────────────────────────────────┘
                     │ uses
                     ↓
┌─────────────────────────────────────────────────────┐
│                  Managers (Services)                │
│  ┌──────────────────┬──────────────────┬────────┐  │
│  │ GeofenceManager  │ MultipeerManager │ Chat   │  │
│  │ ContentSharing   │ EncounterTracking│ AI     │  │
│  └──────────────────┴──────────────────┴────────┘  │
└────────────────────┬────────────────────────────────┘
                     │ publishes
                     ↓
┌─────────────────────────────────────────────────────┐
│              Firebase Analytics                     │
│     Analytics | Crashlytics | Performance          │
└─────────────────────────────────────────────────────┘
```

### 4.2 核心管理器

#### GeofenceManager
- **職責**: 車站進出檢測
- **延遲**: 進入 30 秒，退出 3 分鐘
- **監控**: 動態監控最近 20 個車站
- **事件**: station_entered, station_exited

#### MultipeerManager
- **職責**: P2P 發現與連接
- **服務**: Advertiser + Browser
- **超時**: 60 秒節點過期，30 秒連接超時
- **事件**: peer_discovered, peer_connected

#### ChatManager
- **職責**: 聊天室和消息管理
- **限制**: 500 條消息/聊天室
- **特性**: 臨時消息、已讀回執
- **持久化**: UserDefaults

#### ContentSharingManager
- **職責**: 內容分享與傳輸
- **審核**: AI 預審（NSFW + 人臉）
- **壓縮**: 自動壓縮大圖片
- **傳輸**: MCSession 數據流

#### EncounterTrackingManager
- **職責**: 相遇記錄與統計
- **去重**: 5 分鐘窗口
- **保留**: 90 天
- **統計**: 頻率、地點、連續天數

### 4.3 數據流

**進入車站流程**:
```
1. 用戶進入車站範圍
2. GeofenceManager 檢測 → 30 秒確認
3. 觸發 station_entered 事件
4. MultipeerManager 啟動發現
5. 發現節點 → 記錄 Encounter
6. 用戶連接 → 創建 ChatRoom
```

**發送消息流程**:
```
1. 用戶輸入消息
2. ChatManager 創建 ChatMessage (PENDING)
3. MultipeerManager 發送 JSON 數據
4. 更新狀態: SENDING → SENT → DELIVERED
5. 對方接收 → 顯示並標記 READ
6. (如果是臨時消息) 10 秒後自動刪除
```

**發送內容流程**:
```
1. 用戶選擇照片/文本
2. ContentSharingManager 創建 ContentItem
3. AI 審核 (REVIEWING):
   - NSFW 檢測 (< 500ms)
   - 人臉檢測 (< 300ms)
4. 審核通過 (APPROVED) → 壓縮圖片
5. 通過 MultipeerManager 發送
6. 實時進度更新 (0.0 → 1.0)
7. 發送完成 (SENT)
```

**離開車站流程**:
```
1. 用戶離開車站範圍
2. GeofenceManager 檢測 → 3 分鐘確認
3. 觸發 station_exited 事件
4. 自動清理:
   - 關閉所有 ChatRoom
   - 刪除所有 ContentItem
   - 清空臨時文件
   - 斷開所有連接
   - 停止 MultipeerManager
5. 重置 AppState
```

---

## 5. 已實現功能

### ✅ 已完成 (iOS)

1. **Feature 12**: Firebase 監控與分析
   - Analytics (40+ 事件)
   - Crashlytics (自動崩潰報告)
   - Performance Monitoring (AI 推理、傳輸時間)

2. **Feature 1**: 地理圍欄系統
   - 34 個車站數據庫
   - 進出檢測（30s / 3min 延遲）
   - 動態區域監控（最近 20 個）

3. **Feature 2**: P2P 發現
   - MultipeerConnectivity 實現
   - 自動發現與連接
   - 節點過期清理（60s）

4. **Feature 3**: 內容分享
   - 照片 + 文本分享
   - 自動圖片壓縮（10MB 限制）
   - 實時傳輸進度

5. **Feature 4**: AI 安全引擎
   - NSFW 檢測（Core ML，占位模型）
   - 人臉檢測（Vision framework，生產就緒）
   - < 1 秒總處理時間

6. **Feature 5**: 1對1 聊天室
   - 聊天室管理
   - P2P 消息收發
   - 送達狀態與已讀回執
   - UI 完整實現

7. **Feature 6**: 臨時消息
   - 10 秒自動刪除
   - 倒計時顯示
   - 線程安全追蹤

8. **Feature 7**: 封鎖與舉報
   - 封鎖/解封節點
   - 舉報 9 種違規類型
   - 自動過濾封鎖用戶

9. **Feature 8**: 相遇追蹤
   - 自動記錄相遇
   - 地點與時間統計
   - 頻率檢測（每日/每週）

### 🚧 進行中 (Android)

- 基礎架構完成（1200+ 行）
- 所有數據模型完成（8 個文件）
- AnalyticsManager 生產就緒
- Manager 實現進行中

---

## 6. 隱私與安全

### 6.1 核心原則

**必須實現**:
1. ✅ 無服務器通訊（僅 P2P）
2. ✅ 無需互聯網（離線優先）
3. ✅ 無數據持久化（會話期間）
4. ✅ 離站自動刪除所有數據
5. ✅ 臨時 UUID（無持久標識）
6. ✅ AI 內容審核
7. ✅ 不分享位置（僅車站名稱）
8. ✅ 不訪問聯繫人

**禁止實現**:
1. ❌ 服務器 API 調用
2. ❌ 雲存儲
3. ❌ 用戶帳號
4. ❌ 登錄系統
5. ❌ 包含 PII 的分析
6. ❌ 社交媒體集成
7. ❌ 位置歷史追蹤

### 6.2 數據生命週期

```
進入車站
    ↓
生成臨時 UUID
初始化 P2P 會話
啟動發現
    ↓
車站內（僅內存）
- 節點列表
- 聊天消息
- 分享內容（臨時文件）
- 連接狀態
    ↓
離開車站（自動清理）
- 刪除所有聊天消息
- 刪除所有分享內容
- 刪除臨時文件
- 斷開所有連接
- 重置 UUID
- 清空內存
```

### 6.3 AI 安全流程

```swift
// 1. NSFW 檢測（必需）
let nsfwResult = try await NSFWDetector.analyze(content)
if nsfwResult.confidence >= 0.3 {
    return .rejected(.nsfwDetected)
}

// 2. 人臉檢測（警告）
let faceResult = try await FaceDetector.analyze(content)
if faceResult.faceCount > 0 {
    let confirmed = await showFaceWarning()
    if !confirmed {
        return .rejected(.userCancelled)
    }
}

// 3. 記錄審核
AnalyticsManager.shared.logContentReviewedByAI(
    contentType: contentType,
    reviewResult: .approved,
    reviewDurationMs: Int(duration * 1000)
)

return .approved
```

---

## 7. 性能指標

### 7.1 地理圍欄性能

| 指標 | 目標 | 測量方式 |
|------|------|----------|
| 進入檢測 | < 30s | GeofenceManager 計時器 |
| 退出檢測 | < 3min | GeofenceManager 計時器 |
| 電池影響 | < 5% / 小時 | iOS 電池設置 |
| CPU 使用 | < 10% 平均 | Xcode Instruments |

### 7.2 P2P 性能

| 指標 | 目標 | 測量方式 |
|------|------|----------|
| 發現時間 | < 5 秒 | Firebase Performance |
| 連接時間 | < 3 秒 | Firebase Performance |
| 傳輸速度 | > 100 KB/s | Firebase Performance |

### 7.3 AI 推理性能

| 指標 | 目標 | 測量方式 |
|------|------|----------|
| NSFW 檢測 | < 500ms | PerformanceTracker |
| 人臉檢測 | < 300ms | PerformanceTracker |
| 模型加載 | < 2 秒 | App 啟動追蹤 |

### 7.4 內存需求

| 組件 | 最大內存 |
|------|----------|
| 地理圍欄 | < 10 MB |
| P2P 發現 | < 20 MB |
| AI 模型 | < 100 MB |
| 內容緩存 | < 50 MB |
| 總應用 | < 200 MB |

---

## 8. 測試覆蓋率

### 8.1 當前覆蓋率

| 組件類型 | 目標覆蓋率 | 當前覆蓋率 |
|---------|-----------|-----------|
| Models | 95% | ✅ 95% |
| Services | 90% | ✅ 90% |
| ViewModels | 85% | 🚧 進行中 |
| 總體 | 90% | ✅ 92% |

### 8.2 測試統計

- **單元測試**: 100+ 測試方法
- **集成測試**: 完整用戶流程
- **性能測試**: AI 推理、批量操作
- **所有測試**: ✅ 通過

---

## 9. 為 Go Server 做準備

### 9.1 當前架構限制

TrainBlink 當前設計為**完全 P2P**，無服務器：

**優點**:
- 完全隱私
- 無需互聯網
- 無服務器成本
- 數據不離開設備

**限制**:
- 無法跨車站通訊
- 無法保存歷史記錄
- 無法實現用戶帳號
- 無法進行內容審核（僅本地 AI）

### 9.2 潛在 Server 功能

如果要實作 Go Server，可以考慮以下功能（**需要 PRD 修改**）：

#### A. 增強功能（可選）

1. **內容審核服務**
   - 雲端 AI 審核（更準確）
   - 人工審核隊列
   - 全局封禁列表

2. **匹配與推薦**
   - 基於車站的匹配
   - 興趣推薦
   - 熱門內容

3. **歷史與同步**
   - 跨設備同步（可選）
   - 相遇歷史備份
   - 聊天記錄（加密）

4. **通知服務**
   - 推送通知
   - 離線消息
   - 相遇提醒

5. **分析與監控**
   - 實時分析儀表板
   - 用戶行為分析
   - 濫用檢測

#### B. Server 架構建議

```
┌─────────────────────────────────────────────────────┐
│                   Go Server                         │
│                                                     │
│  ┌──────────────────────────────────────────────┐  │
│  │         HTTP/WebSocket Gateway              │  │
│  │    (Gin/Echo + Gorilla WebSocket)           │  │
│  └────────────────┬────────────────────────────┘  │
│                   │                                │
│  ┌────────────────┴────────────────┬─────────────┐ │
│  │                                 │             │ │
│  ▼                                 ▼             ▼ │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐   │ │
│  │  Auth    │    │ Matching │    │ Content  │   │ │
│  │  Service │    │  Service │    │ Moderation│  │ │
│  └──────────┘    └──────────┘    └──────────┘   │ │
│                                                   │ │
│  ┌──────────────────────────────────────────────┐ │
│  │           Message Broker (NATS/Redis)        │ │
│  └──────────────────────────────────────────────┘ │
│                                                   │ │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐   │ │
│  │ PostgreSQL│   │  Redis   │    │   S3     │   │ │
│  │ (Metadata)│   │ (Cache)  │    │ (Content)│   │ │
│  └──────────┘    └──────────┘    └──────────┘   │ │
└─────────────────────────────────────────────────────┘
```

#### C. API 設計建議

**認證**:
```
POST /api/v1/auth/anonymous
- 生成臨時 token
- 無需註冊
- 車站綁定

Response:
{
  "token": "jwt-token",
  "userId": "temp-uuid",
  "expiresAt": 1700000000
}
```

**發現**:
```
GET /api/v1/stations/{stationId}/peers
- 獲取車站內的節點
- 實時更新（WebSocket）

Response:
{
  "peers": [
    {
      "id": "peer-uuid",
      "displayName": "User-1234",
      "discoveredAt": "2025-11-20T08:00:00Z",
      "signalStrength": 0.8
    }
  ]
}
```

**消息**:
```
POST /api/v1/messages
- 發送消息
- P2P 優先，Server 作為中繼

Request:
{
  "receiverId": "peer-uuid",
  "text": "Hello",
  "isEphemeral": true,
  "lifetimeSeconds": 10
}

Response:
{
  "messageId": "msg-uuid",
  "status": "sent",
  "sentAt": "2025-11-20T08:00:00Z"
}
```

**內容審核**:
```
POST /api/v1/content/review
- 提交內容進行審核
- 返回審核結果

Request:
{
  "contentId": "content-uuid",
  "contentType": "photo",
  "imageData": "base64..."
}

Response:
{
  "approved": true,
  "confidence": 0.95,
  "reasons": [],
  "reviewedAt": "2025-11-20T08:00:00Z"
}
```

### 9.3 數據庫設計建議

**PostgreSQL 表結構**:

```sql
-- 用戶（臨時）
CREATE TABLE users (
    id UUID PRIMARY KEY,
    display_name VARCHAR(50),
    current_station_id VARCHAR(20),
    created_at TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN
);

-- 車站
CREATE TABLE stations (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(100),
    name_en VARCHAR(100),
    latitude DECIMAL(10, 7),
    longitude DECIMAL(10, 7),
    type VARCHAR(10),
    radius INTEGER,
    lines JSONB
);

-- 消息（臨時存儲）
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    sender_id UUID REFERENCES users(id),
    receiver_id UUID REFERENCES users(id),
    text TEXT,
    is_encrypted BOOLEAN,
    is_ephemeral BOOLEAN,
    expires_at TIMESTAMP,
    delivery_status VARCHAR(20),
    created_at TIMESTAMP,
    delivered_at TIMESTAMP,
    read_at TIMESTAMP
);

-- 內容項
CREATE TABLE content_items (
    id UUID PRIMARY KEY,
    sender_id UUID REFERENCES users(id),
    content_type VARCHAR(20),
    state VARCHAR(20),
    text_content TEXT,
    image_url VARCHAR(500),
    file_size INTEGER,
    ai_reviewed_at TIMESTAMP,
    ai_rejection_reason VARCHAR(50),
    created_at TIMESTAMP
);

-- 相遇記錄
CREATE TABLE encounters (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    peer_id UUID REFERENCES users(id),
    station_id VARCHAR(20) REFERENCES stations(id),
    interaction_type VARCHAR(20),
    created_at TIMESTAMP
);

-- 封鎖記錄
CREATE TABLE blocked_peers (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    blocked_peer_id UUID REFERENCES users(id),
    reason VARCHAR(50),
    blocked_at TIMESTAMP
);

-- 舉報記錄
CREATE TABLE reports (
    id UUID PRIMARY KEY,
    reporter_id UUID REFERENCES users(id),
    reported_peer_id UUID REFERENCES users(id),
    reason VARCHAR(50),
    description TEXT,
    context_type VARCHAR(20),
    status VARCHAR(20),
    created_at TIMESTAMP,
    reviewed_at TIMESTAMP
);
```

### 9.4 WebSocket 協議建議

**連接**:
```
ws://server/ws?token=jwt-token&stationId=1001
```

**消息類型**:

```json
// 1. 節點發現
{
  "type": "peer_discovered",
  "peer": {
    "id": "peer-uuid",
    "displayName": "User-1234",
    "discoveredAt": "2025-11-20T08:00:00Z"
  }
}

// 2. 節點離開
{
  "type": "peer_lost",
  "peerId": "peer-uuid"
}

// 3. 收到消息
{
  "type": "message_received",
  "message": {
    "id": "msg-uuid",
    "senderId": "peer-uuid",
    "text": "Hello",
    "timestamp": "2025-11-20T08:00:00Z"
  }
}

// 4. 送達回執
{
  "type": "message_delivered",
  "messageId": "msg-uuid",
  "deliveredAt": "2025-11-20T08:00:00Z"
}

// 5. 已讀回執
{
  "type": "message_read",
  "messageId": "msg-uuid",
  "readAt": "2025-11-20T08:00:00Z"
}
```

### 9.5 Go Server 技術棧建議

```go
// 主要框架
- Web: Gin / Echo
- WebSocket: Gorilla WebSocket
- Database: GORM / sqlx
- Cache: go-redis
- Queue: NATS / RabbitMQ
- Auth: golang-jwt
- Config: Viper
- Logging: Zap / Logrus

// 項目結構
project/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── routes.go
│   ├── service/
│   │   ├── auth.go
│   │   ├── matching.go
│   │   └── moderation.go
│   ├── repository/
│   │   ├── user.go
│   │   ├── message.go
│   │   └── station.go
│   ├── model/
│   │   ├── peer.go
│   │   ├── message.go
│   │   └── content.go
│   └── ws/
│       ├── hub.go
│       └── client.go
├── pkg/
│   ├── jwt/
│   ├── logger/
│   └── validator/
├── config/
│   └── config.yaml
└── migrations/
    └── 001_initial.sql
```

---

## 10. 下一步行動

### 10.1 立即可做

1. ✅ **分析完成**：已充分理解 TrainBlink 架構和數據模型
2. 🚧 **環境準備**：準備 Go Server 開發環境
3. ⏳ **決策**：確定 Server 功能範圍（需與 PRD 協調）

### 10.2 Go Server 實作步驟（建議）

如果決定實作 Server：

**Phase 1: 基礎架構**
1. 設置 Go 項目結構
2. 實現基本 HTTP/WebSocket 服務器
3. 數據庫設計與遷移
4. JWT 認證系統

**Phase 2: 核心功能**
1. 車站節點發現 API
2. 消息中繼服務
3. WebSocket 實時通訊
4. Redis 緩存層

**Phase 3: 增強功能**
1. 內容審核服務（AI 集成）
2. 相遇記錄聚合
3. 封鎖與舉報管理
4. 分析儀表板

**Phase 4: 優化與部署**
1. 性能優化
2. 負載測試
3. Docker 容器化
4. K8s 部署

### 10.3 關鍵考量

**隱私影響**:
- Server 引入會改變隱私模型
- 需要透明的數據政策
- 考慮端到端加密
- GDPR/隱私法合規

**技術挑戰**:
- P2P 與 Server 混合架構
- 實時同步複雜度
- 臨時數據管理
- 擴展性設計

**用戶體驗**:
- 保持離線優先
- Server 作為增強，非必需
- 優雅降級策略
- 透明的數據使用

---

## 11. 總結

### 11.1 TrainBlink 核心特徵

- ✅ **完全 P2P**：無服務器依賴
- ✅ **匿名隱私**：臨時 ID，無追蹤
- ✅ **地理圍欄**：僅車站內活動
- ✅ **AI 安全**：本地內容審核
- ✅ **自動清理**：離站自動刪除

### 11.2 數據模型完整性

已完整分析：
- ✅ Station: 車站信息
- ✅ Peer: 對等節點
- ✅ ChatMessage: 聊天消息
- ✅ ContentItem: 分享內容
- ✅ Encounter: 相遇記錄
- ✅ ChatRoom: 聊天室
- ✅ BlockedPeer: 封鎖記錄
- ✅ Report: 舉報記錄

### 11.3 就緒狀態

**為 Go Server 實作準備就緒**：
- ✅ 數據模型清晰
- ✅ 協議理解完整
- ✅ 架構分析透徹
- ✅ API 設計建議完成
- ✅ 技術棧推薦確定

**待確認**：
- ⏳ Server 功能範圍
- ⏳ 隱私政策更新
- ⏳ PRD 修訂（如需要）

---

**報告完成日期**: 2025-11-20
**下一步**: 準備 Go Server 開發環境並等待功能需求確認
