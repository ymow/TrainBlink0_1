# TrainBlink Matrix + MLS + WebTransport 實作路線圖

**最終目標**: 使用 Matrix + MLS + WebTransport 實現 iOS ↔ Android 跨平台通訊

**策略**: 從簡單到複雜，循序漸進，每個階段都能運行

---

## 📊 整體架構圖

```
最終架構:

┌──────────────────┐         ┌──────────────────┐
│   iOS Client     │         │ Android Client   │
│                  │         │                  │
│ Matrix SDK       │         │ Matrix SDK       │
│ + MLS            │         │ + MLS            │
│ + WebTransport   │         │ + WebTransport   │
└────────┬─────────┘         └─────────┬────────┘
         │                             │
         │    WebTransport/WebSocket   │
         │                             │
         └──────────┬──────────────────┘
                    │
         ┌──────────▼──────────┐
         │   Go Server         │
         │                     │
         │  Matrix Homeserver  │
         │  (Synapse Proxy)    │
         │  + MLS Handler      │
         │  + WebTransport     │
         └─────────────────────┘
```

---

## 🎯 Phase 0: Hello World (Week 1) - **當前階段**

### 目標
建立最基本的 HTTP/WebSocket 通訊，確認環境正常。

### 任務清單

#### Day 1: Go HTTP Server
- [ ] 實作基本 Gin HTTP server
- [ ] 實作 `/ping` endpoint
- [ ] 實作 `/api/v1/hello` endpoint
- [ ] 添加 CORS 支持
- [ ] 測試 curl 請求

**交付物**:
```bash
curl http://localhost:8080/ping
# 回應: {"message": "pong"}

curl http://localhost:8080/api/v1/hello
# 回應: {"message": "Hello from TrainBlink Server!"}
```

#### Day 2: WebSocket Echo Server
- [ ] 實作基本 WebSocket handler
- [ ] 實作 echo 功能（收到什麼就返回什麼）
- [ ] 添加連接管理（連接/斷開日誌）
- [ ] 測試 WebSocket 連接

**交付物**:
```bash
# 使用 wscat 測試
wscat -c ws://localhost:8080/ws
> {"message": "Hello"}
< {"message": "Echo: Hello"}
```

#### Day 3: iOS 簡單客戶端
- [ ] 創建 iOS 項目
- [ ] 實作 HTTP 請求到 `/api/v1/hello`
- [ ] 實作 WebSocket 連接
- [ ] 實作發送/接收消息
- [ ] UI 顯示消息

**交付物**:
- iOS app 可以連接到 Go server
- 可以發送消息並收到 echo

#### Day 4: Android 簡單客戶端
- [ ] 創建 Android 項目
- [ ] 實作 HTTP 請求到 `/api/v1/hello`
- [ ] 實作 WebSocket 連接（OkHttp）
- [ ] 實作發送/接收消息
- [ ] UI 顯示消息

**交付物**:
- Android app 可以連接到 Go server
- 可以發送消息並收到 echo

#### Day 5: 跨平台消息轉發
- [ ] Server 維護連接列表
- [ ] iOS 發送 → Server 轉發 → Android 接收
- [ ] Android 發送 → Server 轉發 → iOS 接收
- [ ] 添加消息 ID 追蹤

**交付物**:
- ✅ iOS ↔ Server ↔ Android 可以互相收發消息
- ✅ 第一個跨平台 Hello World 完成！

### 成功標準
- ✅ Go server 運行正常
- ✅ iOS app 可以連接並發送消息
- ✅ Android app 可以連接並發送消息
- ✅ iOS ↔ Android 可以互相通訊（通過 server）

---

## 🚀 Phase 1: 基本消息系統 (Week 2)

### 目標
實作基本的用戶識別、消息路由和持久化。

### 任務
- [ ] 用戶 UUID 系統（臨時匿名 ID）
- [ ] 消息路由（點對點）
- [ ] 消息歷史（內存存儲）
- [ ] 已讀回執
- [ ] 消息送達狀態

### 數據結構
```go
type Message struct {
    ID         string    `json:"id"`
    From       string    `json:"from"`
    To         string    `json:"to"`
    Text       string    `json:"text"`
    Timestamp  time.Time `json:"timestamp"`
    Status     string    `json:"status"` // pending, sent, delivered, read
}
```

### 交付物
- ✅ 用戶有唯一 ID
- ✅ 消息可以指定接收者
- ✅ 接收者離線時消息保存
- ✅ 接收者上線時收到離線消息

---

## 🔐 Phase 2: JWT 認證系統 (Week 3)

### 目標
實作安全的匿名認證系統。

### 任務
- [ ] JWT token 生成
- [ ] `/auth/anonymous` endpoint
- [ ] Token 驗證中間件
- [ ] WebSocket 認證
- [ ] Token 刷新機制

### API
```http
POST /api/v1/auth/anonymous
Request:
{
  "device_id": "iOS-1234",
  "station_id": "1001"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": "temp-uuid-1234",
  "expires_at": 1700000000
}
```

### 交付物
- ✅ 匿名用戶可以獲取 token
- ✅ WebSocket 連接需要 token
- ✅ 非法 token 被拒絕

---

## 📍 Phase 3: 車站系統整合 (Week 4)

### 目標
整合地理圍欄和車站房間概念。

### 任務
- [ ] 載入 34 個車站數據
- [ ] 車站 API (`/api/v1/stations`)
- [ ] 車站內用戶列表
- [ ] 進入/離開車站事件
- [ ] 車站房間邏輯

### 數據流
```
1. iOS/Android 檢測進入車站 (Geofence)
2. 調用 /api/v1/stations/1001/join
3. Server 將用戶加入 "台北車站" 房間
4. 廣播新用戶進入給房間內其他用戶
5. 用戶可以看到同車站內的其他用戶
```

### 交付物
- ✅ 車站數據載入成功
- ✅ 用戶可以加入車站
- ✅ 同車站用戶可以看到彼此
- ✅ 離開車站自動清理

---

## 🎭 Phase 4: Matrix 協議基礎 (Week 5-6)

### 目標
開始整合 Matrix 協議，先不用 MLS。

### 任務
- [ ] 研究 Matrix Client-Server API
- [ ] 實作 Matrix 房間概念
- [ ] 實作 Matrix 事件系統
- [ ] 車站 → Matrix Room 映射
- [ ] 基本的 Matrix 消息格式

### Matrix Room 結構
```json
{
  "room_id": "!taipei-station:trainblink.org",
  "room_alias": "#taipei-station:trainblink.org",
  "name": "台北車站",
  "topic": "Chat at Taipei Station",
  "members": ["@user1", "@user2"],
  "encryption": null  // Phase 5 才加入
}
```

### 交付物
- ✅ 車站對應到 Matrix Room
- ✅ 消息格式符合 Matrix 規範
- ✅ 可以查詢房間成員
- ✅ 可以查詢房間歷史

---

## 🔒 Phase 5: MLS 加密整合 (Week 7-8)

### 目標
添加 MLS 端到端加密。

### 任務
- [ ] 研究 MLS 規範
- [ ] 整合 MLS 庫（Go）
- [ ] 實作 MLS 群組創建
- [ ] 實作 MLS 成員管理
- [ ] 實作 MLS 消息加密/解密

### MLS 流程
```
1. 用戶 A 進入車站
2. Server 為該車站創建 MLS 群組（如果不存在）
3. 用戶 A 加入 MLS 群組
4. 用戶 B 進入同車站
5. 用戶 B 加入同 MLS 群組
6. A 發送消息 → MLS 加密 → Server 轉發 → B MLS 解密
```

### 交付物
- ✅ MLS 群組自動創建
- ✅ 用戶自動加入車站 MLS 群組
- ✅ 消息端到端加密
- ✅ Server 無法讀取消息內容

---

## ⚡ Phase 6: WebTransport 整合 (Week 9-10)

### 目標
升級傳輸層到 WebTransport（如果可行）。

### 任務
- [ ] 研究 WebTransport Go 實現
- [ ] 實作 WebTransport server
- [ ] iOS WebTransport 客戶端
- [ ] Android WebTransport 客戶端
- [ ] 優雅降級到 WebSocket

### 性能對比
```
測試場景: 發送 100 條消息

WebSocket:
- 平均延遲: 120ms
- P99 延遲: 250ms

WebTransport:
- 平均延遲: 60ms
- P99 延遲: 100ms

改善: ~50% 延遲降低
```

### 交付物
- ✅ WebTransport 連接成功
- ✅ 延遲明顯降低
- ✅ 自動降級到 WebSocket（如果不支持）

---

## 🎨 Phase 7: 完整功能整合 (Week 11-12)

### 目標
整合 TrainBlink 的所有特色功能。

### 任務
- [ ] 臨時消息（10 秒自動刪除）
- [ ] 內容分享（照片/文本）
- [ ] AI 安全審核
- [ ] 封鎖與舉報
- [ ] 相遇追蹤
- [ ] Analytics 整合

### 交付物
- ✅ 所有 TrainBlink 功能實現
- ✅ iOS + Android 功能對等
- ✅ 完整的端到端測試

---

## 📊 技術選型總結

| 階段 | 技術棧 | 複雜度 | 時間 |
|------|--------|--------|------|
| Phase 0 | HTTP + WebSocket | ⭐☆☆☆☆ | 1 週 |
| Phase 1 | 基本消息系統 | ⭐⭐☆☆☆ | 1 週 |
| Phase 2 | JWT 認證 | ⭐⭐☆☆☆ | 1 週 |
| Phase 3 | 車站系統 | ⭐⭐⭐☆☆ | 1 週 |
| Phase 4 | Matrix 基礎 | ⭐⭐⭐☆☆ | 2 週 |
| Phase 5 | MLS 加密 | ⭐⭐⭐⭐☆ | 2 週 |
| Phase 6 | WebTransport | ⭐⭐⭐⭐☆ | 2 週 |
| Phase 7 | 完整功能 | ⭐⭐⭐☆☆ | 2 週 |
| **總計** | | | **12 週** |

---

## 🎯 當前狀態

✅ **已完成**:
- TrainBlink 系統分析
- Go 項目結構建立
- 數據模型定義（Station, Peer, Message）

🚧 **進行中**:
- Phase 0: Hello World 實作

⏳ **待完成**:
- Phase 1-7

---

## 📝 下一步行動

### 立即開始：Phase 0 - Day 1

**任務**: 實作 Go HTTP Hello World Server

**步驟**:
1. 創建 `cmd/server/main.go`
2. 實作基本 Gin server
3. 添加 `/ping` 和 `/api/v1/hello` endpoints
4. 測試 HTTP 請求
5. 添加日誌和錯誤處理

**預期時間**: 2-3 小時

**成功標準**:
```bash
# Server 啟動
go run cmd/server/main.go
# 輸出: Server running on :8080

# 測試 ping
curl http://localhost:8080/ping
# 回應: {"message":"pong","timestamp":"2025-11-20T09:00:00Z"}

# 測試 hello
curl http://localhost:8080/api/v1/hello
# 回應: {"message":"Hello from TrainBlink Server!","version":"0.1.0"}
```

---

## 💡 關鍵決策點

### 為什麼從 Hello World 開始？

1. **降低風險**: 先確認基礎設施正常
2. **快速驗證**: 每個階段都能運行，獲得即時反饋
3. **團隊信心**: 看到進展會提升士氣
4. **學習曲線**: 逐步學習複雜技術
5. **靈活調整**: 遇到問題可以早期調整架構

### 為什麼最後才整合 Matrix + MLS？

1. **技術成熟度**: Matrix MLS 還在 beta，先用簡單方案
2. **可測試性**: 先確保基本通訊正常
3. **可替換性**: 如果 Matrix 有問題，可以換其他方案
4. **漸進式**: 先實現功能，再優化協議

---

## 🚀 準備好開始了嗎？

我建議我們現在就開始 **Phase 0 - Day 1**：

```bash
# 我將為你創建：
1. cmd/server/main.go          # HTTP server 入口
2. internal/api/routes.go      # 路由定義
3. internal/api/handlers/ping.go   # Ping handler
4. internal/api/handlers/hello.go  # Hello handler
5. pkg/logger/logger.go        # 日誌工具
```

**準備好開始了嗎？** 我們從最簡單的 HTTP server 開始！🎉
