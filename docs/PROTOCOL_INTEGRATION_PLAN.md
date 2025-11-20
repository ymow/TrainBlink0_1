# TrainBlink 通訊協議整合方案

**文檔版本**: 1.0  
**創建日期**: 2025-11-20  
**狀態**: 規劃中 → 執行

---

## 📋 執行摘要

本文檔提供了 TrainBlink 通訊協議的完整整合方案，從當前的 HTTP REST API 逐步演進到 Matrix + MLS + WebTransport 的現代化通訊架構。

### 核心目標

1. **跨平台互通**: iOS ↔ Android 無縫通訊
2. **即時性**: 從 HTTP 輪詢到 WebSocket/WebTransport
3. **安全性**: 從明文到 MLS 端到端加密
4. **可演進**: 每個階段都能獨立運行和測試
5. **保持特性**: TrainBlink 的車站、臨時性、隱私特性

### 演進路線（12 週）

```
Phase 0 (Week 1):   HTTP REST API ✅ [已完成]
Phase 1 (Week 2):   WebSocket 即時通訊
Phase 2 (Week 3):   JWT 認證 + 持久化
Phase 3 (Week 4):   車站地理圍欄整合
Phase 4 (Week 5-6): Matrix 協議基礎
Phase 5 (Week 7-8): MLS 端到端加密
Phase 6 (Week 9-10): WebTransport 升級
Phase 7 (Week 11-12): 完整功能整合
```

### 當前狀態

- ✅ **Phase 0 完成**: HTTP REST API, 客戶端管理, 消息存儲
- 🎯 **下一步**: Phase 1 - WebSocket 即時通訊
- 📊 **進度**: 8% (1/12 週)

---

## 📚 文檔結構

本整合方案包含以下文檔：

1. **PROTOCOL_INTEGRATION_PLAN.md** (本文檔)
   - 執行摘要
   - 整體路線圖
   - 文檔索引

2. **PROTOCOL_SPECIFICATION.md**
   - 詳細協議規範（每個階段）
   - 消息格式定義
   - API 接口設計
   - 錯誤處理規範

3. **CLIENT_INTEGRATION_GUIDE.md**
   - iOS 客戶端實現指南
   - Android 客戶端實現指南
   - 通用連接流程
   - 代碼示例

4. **MATRIX_MLS_INTEGRATION.md**
   - Matrix 協議整合路徑
   - MLS 加密實現方案
   - WebTransport 升級策略
   - 技術選型與決策

5. **TESTING_STRATEGY.md**
   - 測試計劃（每個階段）
   - Postman 測試集
   - 跨平台互通測試
   - 性能測試指標

6. **RISK_ASSESSMENT.md**
   - 技術風險評估
   - 隱私與安全考量
   - 決策理由與權衡
   - 應急方案

---

## 🎯 分階段路線圖

### Phase 0: HTTP REST API ✅ (Week 1) - 已完成

**目標**: 建立基礎 HTTP 通訊

**已完成項目**:
- ✅ HTTP Server (Gin framework)
- ✅ 客戶端註冊 API
- ✅ 消息發送/接收 API
- ✅ 點對點消息
- ✅ 廣播消息
- ✅ 線程安全存儲
- ✅ CORS 支持
- ✅ Postman 測試環境

**當前限制**:
- 僅 HTTP，無即時推送
- 內存存儲，無持久化
- 無認證機制
- 無地理圍欄

**測試覆蓋**:
- ✅ 12 個 Postman 測試
- ✅ 自動化測試腳本
- ✅ curl 命令測試

---

### Phase 1: WebSocket 即時通訊 (Week 2)

**目標**: 添加 WebSocket 支持，實現即時消息推送

**核心功能**:
1. **WebSocket 服務器**
   - 連接管理（hub 模式）
   - 心跳保活
   - 自動重連
   - 連接池管理

2. **即時消息推送**
   - 點對點即時送達
   - 廣播消息
   - 在線狀態同步
   - 輸入狀態 (typing indicator)

3. **混合架構**
   - HTTP 用於初始註冊
   - WebSocket 用於消息傳輸
   - 優雅降級（WebSocket → HTTP 輪詢）

**API 設計**:
```javascript
// WebSocket 連接
ws://localhost:8080/ws?client_id={client_id}

// 消息格式
{
  "type": "message",
  "payload": {
    "id": "msg-123",
    "from": "client-456",
    "to": "client-789",
    "text": "Hello!",
    "timestamp": "2025-11-20T10:00:00Z"
  }
}
```

**交付物**:
- [ ] WebSocket hub 實現
- [ ] 客戶端連接管理
- [ ] 即時消息轉發
- [ ] Postman WebSocket 測試
- [ ] iOS/Android 示例代碼

**成功標準**:
- ✅ WebSocket 連接穩定（30 秒心跳）
- ✅ 消息延遲 < 100ms
- ✅ 支持 100+ 並發連接
- ✅ 自動重連成功率 > 95%

---

### Phase 2: JWT 認證 + 持久化 (Week 3)

**目標**: 添加安全認證和數據持久化

**核心功能**:
1. **JWT 認證系統**
   - 匿名 token 生成
   - Token 驗證中間件
   - Token 刷新機制
   - WebSocket 認證

2. **PostgreSQL 持久化**
   - 客戶端表
   - 消息表
   - 索引優化
   - 自動清理過期數據

3. **Redis 緩存**
   - 在線用戶緩存
   - 消息緩存
   - Token 黑名單

**數據庫設計**:
```sql
-- 客戶端表
CREATE TABLE clients (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    device_id VARCHAR(100),
    station_id VARCHAR(20),
    token_hash VARCHAR(64),
    connected_at TIMESTAMP,
    last_seen_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 消息表
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    from_client_id UUID REFERENCES clients(id),
    to_client_id UUID,
    text TEXT NOT NULL,
    delivery_status VARCHAR(20),
    is_ephemeral BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    delivered_at TIMESTAMP,
    read_at TIMESTAMP
);

-- 索引
CREATE INDEX idx_messages_to_client ON messages(to_client_id, created_at DESC);
CREATE INDEX idx_messages_from_client ON messages(from_client_id, created_at DESC);
CREATE INDEX idx_clients_station ON clients(station_id, last_seen_at DESC);
```

**交付物**:
- [ ] JWT 中間件
- [ ] PostgreSQL 集成
- [ ] Redis 緩存層
- [ ] 數據遷移腳本
- [ ] 離線消息推送

**成功標準**:
- ✅ Token 驗證成功率 100%
- ✅ 數據庫查詢 < 10ms (P95)
- ✅ 離線消息正確保存
- ✅ 過期數據自動清理

---

### Phase 3: 車站地理圍欄整合 (Week 4)

**目標**: 整合 TrainBlink 的車站系統

**核心功能**:
1. **車站管理**
   - 34 個車站數據載入
   - 車站 API
   - 地理圍欄驗證

2. **車站房間**
   - 進入車站自動加入房間
   - 離開車站自動清理
   - 同車站用戶發現
   - 車站內廣播

3. **臨時性支持**
   - 離站刪除所有數據
   - 臨時 UUID 管理
   - 會話生命週期

**API 設計**:
```http
# 獲取車站列表
GET /api/v1/stations
Response: {
  "stations": [
    {
      "id": "1001",
      "name": "台北車站",
      "lat": 25.0478,
      "lng": 121.5170,
      "active_users": 15
    }
  ]
}

# 進入車站
POST /api/v1/stations/1001/join
Request: {
  "client_id": "client-123",
  "lat": 25.0478,
  "lng": 121.5170
}

# 離開車站
POST /api/v1/stations/1001/leave
Request: {
  "client_id": "client-123"
}

# 獲取車站內用戶
GET /api/v1/stations/1001/peers
Response: {
  "station_id": "1001",
  "peers": [
    {
      "id": "peer-456",
      "display_name": "User-1234",
      "last_seen": "2025-11-20T10:00:00Z"
    }
  ]
}
```

**交付物**:
- [ ] 車站數據載入
- [ ] 地理圍欄驗證 API
- [ ] 車站房間管理
- [ ] 自動清理機制
- [ ] iOS/Android 地理定位集成

**成功標準**:
- ✅ 車站數據正確載入（34 個）
- ✅ 地理圍欄驗證準確
- ✅ 進出車站事件正確觸發
- ✅ 數據自動清理成功

---

### Phase 4: Matrix 協議基礎 (Week 5-6)

**目標**: 整合 Matrix 協議，為 MLS 加密做準備

**核心功能**:
1. **Matrix Room 概念**
   - 車站 → Matrix Room 映射
   - Room 創建/加入/離開
   - Room 成員管理
   - Room 狀態同步

2. **Matrix 事件系統**
   - m.room.message 事件
   - m.room.member 事件
   - m.typing 事件
   - m.receipt 事件

3. **Matrix Client-Server API**
   - 簡化版實現
   - 必要端點支持
   - 事件同步機制

**Matrix Room 結構**:
```json
{
  "room_id": "!taipei-station:trainblink.local",
  "room_alias": "#taipei-1001:trainblink.local",
  "name": "台北車站",
  "topic": "台北車站 TrainBlink 聊天室",
  "members": [
    "@user-123:trainblink.local",
    "@user-456:trainblink.local"
  ],
  "state": {
    "m.room.create": {
      "creator": "@system:trainblink.local",
      "room_version": "10"
    },
    "m.room.encryption": null  // Phase 5 才啟用
  }
}
```

**Matrix 事件格式**:
```json
{
  "type": "m.room.message",
  "sender": "@user-123:trainblink.local",
  "room_id": "!taipei-station:trainblink.local",
  "event_id": "$event-789",
  "origin_server_ts": 1700000000000,
  "content": {
    "msgtype": "m.text",
    "body": "Hello from TrainBlink!"
  }
}
```

**交付物**:
- [ ] Matrix Room 管理器
- [ ] Matrix 事件處理器
- [ ] 車站 → Room 映射
- [ ] Matrix 消息格式轉換
- [ ] 測試 Matrix 客戶端

**成功標準**:
- ✅ 車站正確映射到 Matrix Room
- ✅ 消息格式符合 Matrix 規範
- ✅ 支持基本 Matrix 事件
- ✅ 可用標準 Matrix 客戶端測試

---

### Phase 5: MLS 端到端加密 (Week 7-8)

**目標**: 實現 MLS 群組加密

**核心功能**:
1. **MLS 基礎設施**
   - MLS 庫集成 (OpenMLS / Cisco MLS)
   - 密鑰管理
   - 群組生命週期

2. **MLS 群組管理**
   - 車站 MLS 群組創建
   - 成員加入/移除
   - 密鑰輪換
   - 狀態同步

3. **端到端加密**
   - 消息加密/解密
   - 前向安全性
   - 未來保密性
   - 認證加密

**MLS 工作流程**:
```
1. 用戶進入車站
   → 生成 MLS KeyPackage
   → 上傳到服務器

2. Server 創建車站 MLS 群組
   → 使用第一個用戶的 KeyPackage
   → 生成 GroupInfo

3. 新用戶加入
   → 獲取 GroupInfo
   → 提交 Commit 消息
   → Server 廣播 Welcome 消息

4. 發送加密消息
   → 客戶端：明文 → MLS 加密
   → Server：轉發密文（無法解密）
   → 接收者：MLS 解密 → 明文

5. 用戶離開
   → 提交 Remove Commit
   → Server 更新群組狀態
   → 重新生成密鑰（前向安全）
```

**API 設計**:
```http
# 上傳 KeyPackage
POST /api/v1/mls/key-packages
Request: {
  "client_id": "client-123",
  "key_package": "base64-encoded-key-package"
}

# 獲取車站 MLS 群組信息
GET /api/v1/stations/1001/mls-group
Response: {
  "group_id": "group-taipei-1001",
  "epoch": 5,
  "members": ["client-123", "client-456"],
  "group_info": "base64-encoded-group-info"
}

# 提交 MLS Commit
POST /api/v1/mls/commit
Request: {
  "group_id": "group-taipei-1001",
  "commit": "base64-encoded-commit",
  "welcome": "base64-encoded-welcome"  // 如果添加成員
}

# 發送加密消息
POST /api/v1/mls/messages
Request: {
  "group_id": "group-taipei-1001",
  "ciphertext": "base64-encoded-mls-ciphertext"
}
```

**交付物**:
- [ ] MLS 庫集成
- [ ] MLS 群組管理器
- [ ] 密鑰包管理
- [ ] 加密消息處理
- [ ] iOS/Android MLS 客戶端

**成功標準**:
- ✅ MLS 群組成功創建
- ✅ 成員加入/移除正常
- ✅ 消息正確加密/解密
- ✅ Server 無法讀取消息內容
- ✅ 前向安全性驗證通過

---

### Phase 6: WebTransport 升級 (Week 9-10)

**目標**: 升級傳輸層到 WebTransport（可選）

**核心功能**:
1. **WebTransport Server**
   - HTTP/3 支持
   - 雙向流
   - 單向流
   - 數據報模式

2. **性能優化**
   - 多路復用
   - 0-RTT 連接恢復
   - 流優先級
   - 擁塞控制

3. **優雅降級**
   - WebTransport → WebSocket
   - 自動協議協商
   - 透明切換

**WebTransport vs WebSocket**:
```
性能對比（100 條消息）:

WebSocket:
- 建立連接: ~50ms (TLS handshake)
- 平均延遲: 120ms
- P99 延遲: 250ms
- 丟包重傳: 整個 TCP 連接阻塞

WebTransport:
- 建立連接: ~30ms (QUIC 0-RTT)
- 平均延遲: 60ms
- P99 延遲: 100ms
- 丟包重傳: 僅阻塞單個流

改善: ~50% 延遲降低
```

**交付物**:
- [ ] WebTransport Go 服務器
- [ ] iOS WebTransport 客戶端
- [ ] Android WebTransport 客戶端
- [ ] 協議協商邏輯
- [ ] 性能基準測試

**成功標準**:
- ✅ WebTransport 連接成功
- ✅ 延遲降低 > 40%
- ✅ 降級到 WebSocket 正常
- ✅ 支持 iOS 16.4+ / Android 12+

**注意**: WebTransport 支持有限，可能需要調整優先級。

---

### Phase 7: 完整功能整合 (Week 11-12)

**目標**: 整合所有 TrainBlink 特色功能

**核心功能**:
1. **臨時消息**
   - 10 秒自動刪除
   - 倒計時顯示
   - 服務器自動清理

2. **內容分享**
   - 照片上傳/下載
   - 圖片壓縮
   - AI 安全審核（NSFW/人臉）

3. **用戶管理**
   - 封鎖用戶
   - 舉報功能
   - 管理介面

4. **相遇追蹤**
   - 相遇記錄
   - 統計分析
   - 歷史查詢

5. **Analytics**
   - 使用統計
   - 性能監控
   - 錯誤追蹤

**交付物**:
- [ ] 臨時消息系統
- [ ] 內容分享 API
- [ ] AI 審核服務
- [ ] 封鎖/舉報功能
- [ ] 相遇追蹤系統
- [ ] Analytics 儀表板

**成功標準**:
- ✅ 所有 TrainBlink 功能正常
- ✅ iOS + Android 功能對等
- ✅ 端到端測試通過
- ✅ 性能滿足 SLA

---

## 📊 技術棧總結

### Server (Go)

```
階段     | 技術棧
--------|------------------------------------------
Phase 0 | net/http, Gin (或 Echo)
Phase 1 | + Gorilla WebSocket
Phase 2 | + PostgreSQL, Redis, JWT
Phase 3 | + 地理圍欄邏輯
Phase 4 | + Matrix Room/Event 管理
Phase 5 | + OpenMLS / Cisco MLS
Phase 6 | + WebTransport (quic-go)
Phase 7 | + AI 服務, S3, Analytics
```

### iOS Client (Swift)

```
階段     | 技術棧
--------|------------------------------------------
Phase 0 | URLSession
Phase 1 | + Starscream (WebSocket)
Phase 2 | + Keychain (Token 存儲)
Phase 3 | + CoreLocation, GeofenceManager
Phase 4 | + Matrix iOS SDK (或自定義)
Phase 5 | + MLS Swift 庫
Phase 6 | + WebTransport (iOS 16.4+)
Phase 7 | + Vision, CoreML, Firebase
```

### Android Client (Kotlin)

```
階段     | 技術棧
--------|------------------------------------------
Phase 0 | Retrofit / OkHttp
Phase 1 | + OkHttp WebSocket
Phase 2 | + EncryptedSharedPreferences
Phase 3 | + Google Location Services
Phase 4 | + Matrix Android SDK (或自定義)
Phase 5 | + MLS Kotlin 庫
Phase 6 | + WebTransport (Android 12+)
Phase 7 | + ML Kit, Firebase
```

---

## 📈 進度追蹤

### 里程碑

| 階段 | 開始日期 | 結束日期 | 狀態 | 進度 |
|------|---------|---------|------|------|
| Phase 0 | 2025-11-18 | 2025-11-20 | ✅ 完成 | 100% |
| Phase 1 | 2025-11-21 | 2025-11-27 | ⏳ 計劃中 | 0% |
| Phase 2 | 2025-11-28 | 2025-12-04 | ⏳ 計劃中 | 0% |
| Phase 3 | 2025-12-05 | 2025-12-11 | ⏳ 計劃中 | 0% |
| Phase 4 | 2025-12-12 | 2025-12-25 | ⏳ 計劃中 | 0% |
| Phase 5 | 2025-12-26 | 2026-01-08 | ⏳ 計劃中 | 0% |
| Phase 6 | 2026-01-09 | 2026-01-22 | ⏳ 計劃中 | 0% |
| Phase 7 | 2026-01-23 | 2026-02-05 | ⏳ 計劃中 | 0% |

**總體進度**: 8% (1/12 週完成)

---

## 🎯 下一步行動

### 立即開始：Phase 1 - WebSocket 即時通訊

**本週目標**:
1. 實現 WebSocket Hub
2. 客戶端連接管理
3. 即時消息轉發
4. 測試與文檔

**詳細計劃**:
參見 [PROTOCOL_SPECIFICATION.md](./PROTOCOL_SPECIFICATION.md) - Phase 1 章節

---

## 📚 相關文檔

- [協議規範](./PROTOCOL_SPECIFICATION.md)
- [客戶端整合指南](./CLIENT_INTEGRATION_GUIDE.md)
- [Matrix/MLS 整合](./MATRIX_MLS_INTEGRATION.md)
- [測試策略](./TESTING_STRATEGY.md)
- [風險評估](./RISK_ASSESSMENT.md)
- [TrainBlink 系統分析](../TRAINBLINK_ANALYSIS.md)
- [12 週路線圖](../ROADMAP.md)

---

**文檔維護者**: TrainBlink Team  
**最後更新**: 2025-11-20  
**版本**: 1.0
