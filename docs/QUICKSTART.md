# TrainBlink 通訊協議整合 - 快速開始指南

**版本**: 1.0  
**最後更新**: 2025-12-06 (新增 JWT 認證說明)

---

## 🚀 5 分鐘快速上手

### 前提條件

- Go 1.21+
- iOS Xcode 15+ / Android Studio
- Postman (用於 API 測試)

---

## Phase 0: HTTP REST API (當前階段)

### 1. 啟動服務器

```bash
cd /home/user/messenger_protocol_research

# 方式 1: 直接運行
go run cmd/server/main_with_post.go

# 方式 2: 編譯後運行
go build -o trainblink-server cmd/server/main_with_post.go
./trainblink-server

# 輸出:
# ╔══════════════════════════════════════════════════════╗
# ║     🚄 TrainBlink Server - Phase 0: Day 2           ║
# ║     Version: 0.1.0                                  ║
# ╚══════════════════════════════════════════════════════╝
# 🚀 Server is running on http://localhost:8080
```

### 2. 獲取認證令牌 (JWT Authentication)

所有 API 端點（除了公開端點）都需要 JWT 認證。首先獲取訪問令牌：

```bash
# 匿名登錄獲取 JWT 令牌
curl -X POST http://localhost:8080/api/v1/auth/anonymous \
  -H "Content-Type: application/json" \
  -d '{"device_id":"test-device-001","device_type":"cli"}'

# 響應:
# {
#   "access_token": "eyJhbGc...",
#   "refresh_token": "eyJhbGc...",
#   "expires_in": 3600,
#   "user_id": "3db81e4b-7b6c-40dc-bcf6-c8540322bf0d"
# }

# 保存 access_token 用於後續請求
export TOKEN="eyJhbGc..."
```

**公開端點（無需認證）**:
- `GET /ping` - 健康檢查
- `GET /health` - 詳細健康狀態
- `GET /api/v1/hello` - 測試端點
- `POST /api/v1/auth/anonymous` - 匿名登錄
- `POST /api/v1/auth/refresh` - 刷新令牌

**受保護端點（需要認證）**:
- 所有其他 `/api/v1/*` 端點

**令牌刷新**:
```bash
# 當 access_token 過期（1小時後）時，使用 refresh_token 獲取新令牌
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"YOUR_REFRESH_TOKEN"}'
```

詳細文檔: [認證實現指南](./AUTHENTICATION_IMPLEMENTATION.md)

---

### 3. 測試 API

#### 選項 A: Postman

```bash
# 導入 Postman Collection
tests/TrainBlink_Complete.postman_collection.json

# 按順序執行：
1. Health Check
2. Register Alice
3. Register Bob
4. Alice → Bob
5. Bob → Alice
6. Get Messages
```

#### 選項 B: curl

```bash
# 1. 獲取認證令牌（見上方步驟 2）
export TOKEN="YOUR_ACCESS_TOKEN_HERE"

# 2. 發送消息
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"content":"Hello from TrainBlink!","to_user_id":"recipient-uuid"}'

# 3. 獲取消息列表
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/messages

# 4. 開始旅程 (Trip)
curl -X POST http://localhost:8080/api/v1/trips/start \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"train_line":"山手線","direction":"外回り","starting_station":"渋谷"}'

# 5. 獲取活動旅程
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/trips/active
```

#### 選項 C: 自動化腳本

```bash
./tests/test_all_endpoints.sh
```

---

## Phase 1: WebSocket (下一階段)

### 準備工作

1. 安裝 wscat (WebSocket 測試工具):
```bash
npm install -g wscat
```

2. 啟動 WebSocket 服務器:
```bash
go run cmd/server/main_websocket.go
```

3. 連接測試（需要認證令牌）:
```bash
# 首先獲取 JWT 令牌（見步驟 2）
export TOKEN="YOUR_ACCESS_TOKEN_HERE"

# 使用令牌連接 WebSocket
wscat -c "ws://localhost:8080/ws?token=$TOKEN&station_id=shibuya"

# 發送消息
> {"type":"message","content":"Hello WebSocket!","to":"recipient-user-id"}

# 收到響應
< {"id":"msg-123","type":"message","from":"sender-id","content":"Hello!","timestamp":"2025-12-06T..."}
```

**注意**: WebSocket 連接需要在 URL 查詢參數中提供有效的 JWT 令牌和 station_id。

---

## iOS 客戶端快速集成

### 1. 創建項目

```bash
# 使用 Xcode 創建新項目
File → New → Project → iOS App
- Product Name: TrainBlink
- Interface: SwiftUI
- Language: Swift
```

### 2. 添加網絡代碼

複製以下文件到項目：
- `NetworkManager.swift` (見 CLIENT_INTEGRATION_GUIDE.md)
- `TrainBlinkAPI.swift`
- `Models.swift`

### 3. 更新 Server URL

```swift
// NetworkManager.swift
private let baseURL = "http://YOUR_MACHINE_IP:8080"  // 不要用 localhost
```

### 4. 運行

```swift
// ContentView.swift
struct ContentView: View {
    @StateObject private var viewModel = ChatViewModel()
    
    var body: some View {
        ChatScreen(viewModel: viewModel)
    }
}
```

---

## Android 客戶端快速集成

### 1. 添加依賴

```kotlin
// build.gradle.kts
dependencies {
    implementation("com.squareup.retrofit2:retrofit:2.9.0")
    implementation("com.squareup.retrofit2:converter-gson:2.9.0")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.7.3")
}
```

### 2. 添加網絡代碼

複製以下文件到項目：
- `ApiService.kt` (見 CLIENT_INTEGRATION_GUIDE.md)
- `RetrofitClient.kt`
- `Models.kt`
- `TrainBlinkRepository.kt`

### 3. 更新 Server URL

```kotlin
// RetrofitClient.kt
private const val BASE_URL = "http://YOUR_MACHINE_IP:8080"
// Android 模擬器使用 10.0.2.2 代表 host 的 localhost
```

### 4. 運行

```kotlin
// MainActivity.kt
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            ChatScreen()
        }
    }
}
```

---

## 常見問題

### 服務器無法啟動

**問題**: `address already in use`

**解決**:
```bash
# 查找占用 8080 端口的進程
lsof -i :8080

# 殺死進程
kill -9 PID
```

### iOS 無法連接

**問題**: `The resource could not be loaded because the App Transport Security policy requires the use of a secure connection`

**解決**: 添加到 `Info.plist`:
```xml
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsArbitraryLoads</key>
    <true/>
</dict>
```

### Android 無法連接

**問題**: `Failed to connect to localhost/127.0.0.1:8080`

**解決**: 
```kotlin
// 使用模擬器時，用 10.0.2.2 代替 localhost
private const val BASE_URL = "http://10.0.2.2:8080"

// 或使用真機時，用電腦的實際 IP
private const val BASE_URL = "http://192.168.1.100:8080"
```

---

## 下一步

1. ✅ 完成 Phase 0 測試
2. 📖 閱讀 [協議規範](./PROTOCOL_SPECIFICATION.md)
3. 📱 實現 iOS/Android 客戶端
4. 🔌 進入 Phase 1: WebSocket
5. 🔐 進入 Phase 2: JWT 認證

---

## 資源鏈接

- [完整整合方案](./PROTOCOL_INTEGRATION_PLAN.md)
- [協議規範](./PROTOCOL_SPECIFICATION.md)
- [客戶端指南](./CLIENT_INTEGRATION_GUIDE.md)
- [測試策略](./TESTING_STRATEGY.md)
- [Matrix/MLS 路線圖](./MATRIX_MLS_ROADMAP.md)

---

**需要幫助?** 查看詳細文檔或聯繫團隊。

**文檔維護者**: TrainBlink Team  
**最後更新**: 2025-12-06 (新增 JWT 認證說明)
