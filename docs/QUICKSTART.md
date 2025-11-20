# TrainBlink 通訊協議整合 - 快速開始指南

**版本**: 1.0  
**最後更新**: 2025-11-20

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

### 2. 測試 API

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
# 1. 註冊客戶端
curl -X POST http://localhost:8080/api/v1/clients \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","device_id":"iOS-1234"}'

# 保存返回的 client_id

# 2. 發送消息
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"from":"CLIENT_ID_HERE","to":"all","text":"Hello!"}'

# 3. 獲取消息
curl http://localhost:8080/api/v1/messages/user?id=CLIENT_ID_HERE
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

3. 連接測試:
```bash
wscat -c "ws://localhost:8080/ws?client_id=test-client"

# 發送消息
> {"type":"send_message","payload":{"to":"all","text":"Hello WebSocket!"}}

# 收到響應
< {"type":"message_ack","payload":{...}}
```

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
**最後更新**: 2025-11-20
