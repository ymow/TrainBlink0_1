# Matrix + MLS + WebTransport 整合路線圖

**版本**: 1.0  
**最後更新**: 2025-11-20

---

## 概述

本文檔概述如何從當前的自定義協議逐步演進到基於 Matrix + MLS + WebTransport 的現代化通訊架構。

---

## 為什麼選擇 Matrix + MLS?

### Matrix 的優勢

1. **開放標準**: 去中心化、開放協議
2. **互操作性**: 可以與其他 Matrix 服務器互通
3. **豐富生態**: 現成的客戶端 SDK、服務器實現
4. **成熟度**: 被 Element、Discord、Mozilla 等使用

### MLS 的優勢

1. **IETF 標準**: RFC 9420（2023 年發布）
2. **群組加密**: 專為群組通訊設計
3. **前向安全**: 即使密鑰洩露，過去的消息仍安全
4. **高效性**: 比 Signal Protocol 更適合大型群組

### 與 TrainBlink 的契合度

| TrainBlink 需求 | Matrix 支持 | MLS 支持 |
|----------------|------------|----------|
| 車站房間 | ✅ Matrix Room | ✅ MLS Group |
| 臨時性 | ✅ 可配置 | ✅ 群組可銷毀 |
| 端到端加密 | ✅ m.room.encryption | ✅ 原生支持 |
| 跨平台 | ✅ 多平台 SDK | ✅ 多語言庫 |
| 即時通訊 | ✅ 實時同步 | ✅ 不影響 |

---

## 演進路徑

### 階段 1: 準備期（Phase 0-3）

**目標**: 建立穩固的基礎，為 Matrix 整合做準備

**已完成**:
- ✅ HTTP REST API
- ✅ 客戶端管理
- ✅ 消息系統

**進行中**:
- 🚧 WebSocket 即時通訊
- 🚧 JWT 認證
- 🚧 車站系統

**關鍵決策**:
- 數據模型與 Matrix 兼容（Room、Event、Member）
- API 設計考慮未來遷移
- 保持協議靈活性

---

### 階段 2: Matrix 基礎（Phase 4）

**目標**: 整合 Matrix 協議核心概念

#### 車站 → Matrix Room 映射

```
台北車站（1001）
    ↓
Matrix Room: !taipei-1001:trainblink.local
    - Room Alias: #taipei-1001:trainblink.local
    - Room Name: 台北車站
    - Visibility: Public (車站內可見)
    - History Visibility: Joined (僅成員可見歷史)
```

#### Matrix 事件類型

| TrainBlink 功能 | Matrix 事件 |
|----------------|------------|
| 發送消息 | m.room.message |
| 用戶進入車站 | m.room.member (join) |
| 用戶離開車站 | m.room.member (leave) |
| 輸入狀態 | m.typing |
| 已讀回執 | m.receipt |
| 臨時消息 | 自定義: m.trainblink.ephemeral |
| 相遇記錄 | 自定義: m.trainblink.encounter |

#### 實現方案

**選項 A: 輕量級 Matrix 實現**
```
優點:
+ 完全控制
+ 輕量級，易於優化
+ 快速迭代

缺點:
- 需要自己實現所有功能
- 維護成本高
- 可能不完全兼容
```

**選項 B: Matrix Synapse 集成**
```
優點:
+ 完整的 Matrix 功能
+ 社區支持
+ 與其他 Matrix 服務器互通

缺點:
- 重量級（Python）
- 學習曲線陡峭
- 可能有不需要的功能
```

**推薦**: **選項 A** → 先實現輕量級版本，驗證可行性後再考慮完整 Synapse

#### 最小可行 Matrix 服務器

**必須實現的端點**:
```
客戶端-服務器 API:
- POST /_matrix/client/v3/register           # 註冊
- POST /_matrix/client/v3/login              # 登錄
- POST /_matrix/client/v3/createRoom         # 創建房間
- POST /_matrix/client/v3/rooms/{roomId}/join  # 加入房間
- POST /_matrix/client/v3/rooms/{roomId}/send/{eventType}  # 發送事件
- GET  /_matrix/client/v3/sync               # 同步事件

房間管理:
- m.room.create                              # 房間創建
- m.room.member                              # 成員管理
- m.room.message                             # 消息
```

**可選端點**（Phase 5+ 添加）:
```
- m.room.encryption                          # 加密設置
- m.room.power_levels                        # 權限管理
- m.room.topic                               # 房間主題
```

---

### 階段 3: MLS 加密（Phase 5）

**目標**: 實現端到端加密

#### MLS 工作原理

```
基本概念:
1. Group: 車站對應一個 MLS 群組
2. Epoch: 每次成員變化時遞增
3. KeyPackage: 用戶的公鑰材料
4. Commit: 群組狀態變更提案
5. Welcome: 邀請新成員加入
```

#### 車站 MLS 群組生命週期

```
1. 創建車站群組
   用戶 A 進入台北車站
   → Server 創建 MLS Group
   → A 上傳 KeyPackage
   → Group 初始化（Epoch 0）

2. 新用戶加入
   用戶 B 進入台北車站
   → B 上傳 KeyPackage
   → A 提交 Add Commit
   → Server 生成 Welcome 消息
   → B 接受 Welcome 並同步群組狀態
   → Group Epoch 遞增（Epoch 1）

3. 發送加密消息
   用戶 A 發送消息
   → A: 明文 → MLS 加密 → 密文
   → Server: 轉發密文（無法解密）
   → B: 密文 → MLS 解密 → 明文

4. 用戶離開
   用戶 A 離開車站
   → Server 提交 Remove Commit
   → 所有成員更新群組狀態
   → 重新生成密鑰（前向安全）
   → Group Epoch 遞增（Epoch 2）
```

#### MLS 庫選擇

| 語言 | 庫 | 成熟度 | 推薦度 |
|------|-----|--------|--------|
| Go | github.com/cisco/go-mls | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| Swift | Swift-MLS (Apple) | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| Kotlin | MLS-Kotlin | ⭐⭐ | ⭐⭐⭐ |
| Rust | openmls | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ (可用 FFI) |

**推薦方案**:
- **Go Server**: Cisco go-mls（企業級，穩定）
- **iOS**: Swift-MLS（Apple 官方支持）
- **Android**: OpenMLS via JNI（Rust FFI 綁定）

#### MLS API 設計

```http
# 1. 上傳 KeyPackage
POST /api/v1/mls/key-packages
Content-Type: application/json
Authorization: Bearer {token}

Request:
{
  "key_package": "base64-encoded-key-package",
  "client_id": "550e8400-e29b-41d4-a716-446655440000"
}

Response (201):
{
  "status": "uploaded",
  "key_package_ref": "kp-1234567890"
}

---

# 2. 加入車站 MLS 群組
POST /api/v1/stations/{station_id}/mls/join
Authorization: Bearer {token}

Request:
{
  "key_package_ref": "kp-1234567890"
}

Response (200):
{
  "group_id": "group-taipei-1001",
  "epoch": 5,
  "welcome_message": "base64-encoded-welcome",  // 如果是新加入
  "group_info": "base64-encoded-group-info"
}

---

# 3. 發送加密消息
POST /api/v1/mls/messages
Content-Type: application/octet-stream
Authorization: Bearer {token}

Request Body:
<binary MLS ciphertext>

Response (201):
{
  "message_id": "msg-1234567890",
  "timestamp": "2025-11-20T10:00:00Z"
}

---

# 4. 提交 MLS Commit
POST /api/v1/mls/commits
Authorization: Bearer {token}

Request:
{
  "group_id": "group-taipei-1001",
  "commit": "base64-encoded-commit",
  "welcome": "base64-encoded-welcome"  // 如果添加新成員
}

Response (200):
{
  "status": "accepted",
  "new_epoch": 6
}
```

#### 客戶端 MLS 流程（iOS 示例）

```swift
import MLS

class MLSManager {
    private var client: MLSClient
    private var groups: [String: MLSGroup] = [:]
    
    // 1. 初始化 MLS 客戶端
    func initialize() async throws {
        client = try await MLSClient.create()
        
        // 生成 KeyPackage
        let keyPackage = try await client.generateKeyPackage()
        
        // 上傳到服務器
        try await uploadKeyPackage(keyPackage)
    }
    
    // 2. 加入車站群組
    func joinStationGroup(stationID: String) async throws {
        let response = try await api.joinStationMLS(stationID)
        
        if let welcomeData = response.welcomeMessage {
            // 新加入：處理 Welcome 消息
            let group = try await client.processWelcome(welcomeData)
            groups[stationID] = group
        } else {
            // 已存在：同步群組狀態
            let group = groups[stationID]
            try await group?.sync(response.groupInfo)
        }
    }
    
    // 3. 發送加密消息
    func sendEncryptedMessage(to stationID: String, text: String) async throws {
        guard let group = groups[stationID] else {
            throw MLSError.groupNotFound
        }
        
        // MLS 加密
        let plaintext = text.data(using: .utf8)!
        let ciphertext = try await group.encrypt(plaintext)
        
        // 發送到服務器
        try await api.sendMLSMessage(ciphertext)
    }
    
    // 4. 接收加密消息
    func receiveEncryptedMessage(_ ciphertext: Data, from stationID: String) async throws -> String {
        guard let group = groups[stationID] else {
            throw MLSError.groupNotFound
        }
        
        // MLS 解密
        let plaintext = try await group.decrypt(ciphertext)
        return String(data: plaintext, encoding: .utf8) ?? ""
    }
}
```

---

### 階段 4: WebTransport（Phase 6）

**目標**: 升級傳輸層，降低延遲

#### WebTransport vs WebSocket

| 特性 | WebSocket | WebTransport |
|------|-----------|--------------|
| 協議 | TCP | QUIC (UDP) |
| 連接建立 | ~50ms (TLS) | ~30ms (0-RTT) |
| 多路復用 | 單一流 | 多個流 |
| 丟包影響 | 整個連接阻塞 | 僅影響單個流 |
| 優先級 | 不支持 | 支持流優先級 |
| 瀏覽器支持 | 100% | ~70% (2024) |
| 移動支持 | iOS/Android 全支持 | iOS 16.4+, Android 12+ |

#### 實現策略

```
階段 1: WebSocket（Phase 1）
- 完整實現 WebSocket
- 驗證協議和消息格式

階段 2: WebTransport（Phase 6）
- 添加 WebTransport 支持
- 保留 WebSocket 作為後備

階段 3: 智能選擇
- 客戶端檢測 WebTransport 支持
- 自動選擇最佳協議
- 透明降級
```

#### 協議協商

```javascript
// 客戶端自動選擇

async function connectToServer() {
    // 1. 檢測 WebTransport 支持
    if (typeof WebTransport !== 'undefined') {
        try {
            return await connectWebTransport();
        } catch (error) {
            console.warn('WebTransport failed, falling back to WebSocket');
        }
    }
    
    // 2. 降級到 WebSocket
    return connectWebSocket();
}

async function connectWebTransport() {
    const url = 'https://trainblink.local:4433';
    const transport = new WebTransport(url);
    
    await transport.ready;
    
    // 創建雙向流
    const stream = await transport.createBidirectionalStream();
    
    return {
        type: 'webtransport',
        transport,
        stream
    };
}

function connectWebSocket() {
    const ws = new WebSocket('wss://trainblink.local:8080/ws');
    
    return new Promise((resolve, reject) => {
        ws.onopen = () => resolve({ type: 'websocket', ws });
        ws.onerror = reject;
    });
}
```

---

## 技術棧總結

### Server (Go)

```
Phase 4: Matrix 基礎
├── Matrix Room Manager
├── Event Handler
├── Sync Engine
└── State Resolution

Phase 5: MLS 加密
├── github.com/cisco/go-mls
├── KeyPackage 管理
├── Group 生命週期
└── Commit 處理

Phase 6: WebTransport
├── github.com/quic-go/quic-go
├── github.com/quic-go/webtransport-go
└── Protocol Negotiation
```

### iOS (Swift)

```
Phase 4: Matrix 基礎
├── Matrix iOS SDK (或自定義)
├── Room Manager
└── Event Sync

Phase 5: MLS 加密
├── Swift-MLS (Apple)
├── Keychain (密鑰存儲)
└── CryptoKit (輔助)

Phase 6: WebTransport
├── Network.framework
└── 降級邏輯
```

### Android (Kotlin)

```
Phase 4: Matrix 基礎
├── Matrix Android SDK (或自定義)
├── Room Manager
└── Event Sync

Phase 5: MLS 加密
├── OpenMLS (Rust FFI)
├── AndroidKeyStore
└── 密鑰管理

Phase 6: WebTransport
├── Cronet (Chrome 網絡庫)
└── 降級邏輯
```

---

## 風險與挑戰

### 技術風險

| 風險 | 影響 | 緩解措施 |
|------|------|----------|
| Matrix 複雜度高 | 高 | 先實現輕量級版本 |
| MLS 庫不成熟 | 中 | 選擇企業級庫（Cisco） |
| WebTransport 支持有限 | 低 | 保留 WebSocket 後備 |
| 跨平台兼容性 | 中 | 完整測試矩陣 |

### 架構風險

| 風險 | 影響 | 緩解措施 |
|------|------|----------|
| 過度工程 | 高 | 循序漸進，每階段驗證 |
| 性能下降 | 中 | 持續性能測試 |
| 向後兼容 | 低 | 版本協商機制 |

---

## 成功標準

### Phase 4: Matrix 基礎

- ✅ 車站正確映射到 Matrix Room
- ✅ 可用標準 Matrix 客戶端測試
- ✅ 事件格式符合 Matrix 規範
- ✅ 同步機制正常工作

### Phase 5: MLS 加密

- ✅ MLS 群組成功創建
- ✅ 成員加入/移除正常
- ✅ 消息正確加密/解密
- ✅ Server 無法讀取消息內容
- ✅ 前向安全性驗證通過

### Phase 6: WebTransport

- ✅ WebTransport 連接成功
- ✅ 延遲降低 > 40%
- ✅ 降級到 WebSocket 正常
- ✅ iOS/Android 正常工作

---

## 下一步行動

1. **立即**: 完成 Phase 1 (WebSocket)
2. **本週**: 開始 Phase 2 (JWT + 持久化)
3. **下週**: 開始 Phase 3 (車站系統)
4. **2 週後**: 開始研究 Matrix 協議
5. **1 個月後**: 開始 Phase 4 (Matrix 基礎)

---

**文檔維護者**: TrainBlink Team  
**最後更新**: 2025-11-20  
**版本**: 1.0
