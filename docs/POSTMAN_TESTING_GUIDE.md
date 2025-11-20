# TrainBlink Server - Postman Testing Guide

**版本**: 0.1.0
**階段**: Phase 0 - Day 2
**日期**: 2025-11-20

---

## 📋 目錄

1. [快速開始](#快速開始)
2. [導入 Postman Collection](#導入-postman-collection)
3. [HTTP Endpoints 測試](#http-endpoints-測試)
4. [WebSocket 測試](#websocket-測試)
5. [測試場景](#測試場景)
6. [curl 測試命令](#curl-測試命令)

---

## 🚀 快速開始

### 1. 啟動 Server

```bash
# 方法 1: 直接運行
go run cmd/server/main_websocket.go

# 方法 2: 編譯後運行
go build -o bin/server cmd/server/main_websocket.go
./bin/server
```

### 2. 驗證 Server 運行

```bash
curl http://localhost:8080/ping
```

**預期響應**:
```json
{
  "message": "pong",
  "timestamp": "2025-11-20T10:00:00Z",
  "server": "TrainBlink Server v0.1.0"
}
```

---

## 📥 導入 Postman Collection

### 方法 1: 導入文件

1. 打開 Postman
2. 點擊 **Import** 按鈕
3. 選擇文件：`tests/TrainBlink_Server.postman_collection.json`
4. 點擊 **Import**

### 方法 2: 手動創建

如果無法導入，可以手動創建 Collection 並添加以下 endpoints。

---

## 🔍 HTTP Endpoints 測試

### 1. Health & Status

#### A. Root - API Info

**Request**:
```
GET http://localhost:8080/
```

**Expected Response** (200 OK):
```json
{
  "message": "TrainBlink Server API",
  "version": "0.1.0",
  "endpoints": {
    "GET /ping": "Health check",
    "GET /health": "Detailed health status",
    "GET /api/v1/hello": "Hello World",
    "GET /api/v1/welcome": "Welcome message",
    "POST /api/v1/message": "Send message to client",
    "GET /api/v1/connections": "Get active connections",
    "WS /ws": "WebSocket connection"
  }
}
```

---

#### B. Ping

**Request**:
```
GET http://localhost:8080/ping
```

**Expected Response** (200 OK):
```json
{
  "message": "pong",
  "timestamp": "2025-11-20T10:00:00Z",
  "server": "TrainBlink Server v0.1.0"
}
```

---

#### C. Health Check

**Request**:
```
GET http://localhost:8080/health
```

**Expected Response** (200 OK):
```json
{
  "status": "healthy",
  "timestamp": "2025-11-20T10:00:00Z",
  "uptime": 123.456,
  "connections": 2,
  "total_requests": 45
}
```

**Fields**:
- `uptime`: Server 運行時間（秒）
- `connections`: 當前 WebSocket 連接數
- `total_requests`: 總請求數

---

### 2. API v1

#### A. Hello World

**Request**:
```
GET http://localhost:8080/api/v1/hello
```

**Expected Response** (200 OK):
```json
{
  "message": "Hello from TrainBlink Server!",
  "version": "0.1.0",
  "timestamp": "2025-11-20T10:00:00Z",
  "features": [
    "http",
    "websocket",
    "echo",
    "broadcast",
    "message-relay"
  ],
  "metadata": {
    "protocol": "TrainBlink Protocol v1.0",
    "stage": "Phase 0 - Day 2 - WebSocket"
  }
}
```

---

#### B. Welcome

**Request**:
```
GET http://localhost:8080/api/v1/welcome?name=Alice&device_id=iOS-1234
```

**Query Parameters**:
- `name` (optional): 用戶名稱，默認 "Anonymous"
- `device_id` (optional): 設備 ID，默認 "unknown"

**Expected Response** (200 OK):
```json
{
  "message": "Welcome to TrainBlink!",
  "user": "Alice",
  "device_id": "iOS-1234",
  "timestamp": "2025-11-20T10:00:00Z",
  "tip": "Connect via WebSocket at ws://localhost:8080/ws"
}
```

---

#### C. Get Active Connections

**Request**:
```
GET http://localhost:8080/api/v1/connections
```

**Expected Response** (200 OK):
```json
{
  "count": 2,
  "client_ids": [
    "client-1732096800123456789",
    "client-1732096800987654321"
  ],
  "timestamp": "2025-11-20T10:00:00Z"
}
```

**用途**: 獲取所有在線 WebSocket 客戶端的 ID 列表，用於發送定向消息。

---

### 3. Messaging (POST)

#### A. Send Broadcast Message

**Request**:
```
POST http://localhost:8080/api/v1/message
Content-Type: application/json
```

**Body**:
```json
{
  "from": "postman-client",
  "to": "all",
  "message": "Hello everyone from Postman!"
}
```

**Expected Response** (200 OK):
```json
{
  "status": "sent",
  "from": "postman-client",
  "to": "all",
  "message": "Hello everyone from Postman!",
  "timestamp": "2025-11-20T10:00:00Z"
}
```

**行為**: 消息會被廣播到所有連接的 WebSocket 客戶端。

---

#### B. Send Message to Specific Client

**Request**:
```
POST http://localhost:8080/api/v1/message
Content-Type: application/json
```

**Body**:
```json
{
  "from": "postman-client",
  "to": "client-1732096800123456789",
  "message": "Private message from Postman"
}
```

**Expected Response** (200 OK):
```json
{
  "status": "sent",
  "from": "postman-client",
  "to": "client-1732096800123456789",
  "message": "Private message from Postman",
  "timestamp": "2025-11-20T10:00:00Z"
}
```

**行為**: 消息只會發送給指定的客戶端。

**如果客戶端不存在** (404 Not Found):
```json
{
  "error": "Client client-999 not found"
}
```

---

## 🔌 WebSocket 測試

### 使用瀏覽器測試

**HTML Test Page** (`tests/websocket_test.html`):

```html
<!DOCTYPE html>
<html>
<head>
    <title>TrainBlink WebSocket Test</title>
</head>
<body>
    <h1>TrainBlink WebSocket Test</h1>
    <div>
        <button onclick="connect()">Connect</button>
        <button onclick="disconnect()">Disconnect</button>
        <button onclick="sendEcho()">Send Echo</button>
        <button onclick="sendBroadcast()">Send Broadcast</button>
    </div>
    <div>
        <input type="text" id="clientId" placeholder="Target Client ID" />
        <button onclick="sendMessage()">Send to Client</button>
    </div>
    <pre id="log"></pre>

    <script>
        let ws;
        let myClientId;

        function log(msg) {
            document.getElementById('log').textContent += msg + '\n';
        }

        function connect() {
            ws = new WebSocket('ws://localhost:8080/ws');

            ws.onopen = () => {
                log('✅ Connected to server');
            };

            ws.onmessage = (event) => {
                const msg = JSON.parse(event.data);
                log('📩 Received: ' + JSON.stringify(msg, null, 2));

                if (msg.type === 'welcome' && msg.data) {
                    myClientId = msg.data.client_id;
                    log('🆔 My Client ID: ' + myClientId);
                }
            };

            ws.onclose = () => {
                log('❌ Disconnected from server');
            };

            ws.onerror = (error) => {
                log('⚠️ Error: ' + error);
            };
        }

        function disconnect() {
            if (ws) {
                ws.close();
            }
        }

        function sendEcho() {
            const msg = {
                type: 'echo',
                message: 'Hello from browser!'
            };
            ws.send(JSON.stringify(msg));
            log('📤 Sent: ' + JSON.stringify(msg));
        }

        function sendBroadcast() {
            const msg = {
                type: 'broadcast',
                message: 'Broadcasting from browser!'
            };
            ws.send(JSON.stringify(msg));
            log('📤 Sent: ' + JSON.stringify(msg));
        }

        function sendMessage() {
            const targetId = document.getElementById('clientId').value;
            const msg = {
                type: 'message',
                to: targetId,
                message: 'Direct message from browser'
            };
            ws.send(JSON.stringify(msg));
            log('📤 Sent: ' + JSON.stringify(msg));
        }
    </script>
</body>
</html>
```

保存並用瀏覽器打開此文件進行測試。

---

### 使用 wscat 測試

**安裝 wscat**:
```bash
npm install -g wscat
```

**連接 WebSocket**:
```bash
wscat -c ws://localhost:8080/ws
```

**測試命令**:

1. **Echo 消息**:
```json
{"type":"echo","message":"Hello World"}
```

**預期響應**:
```json
{
  "type": "echo",
  "message": "Echo: Hello World",
  "timestamp": "2025-11-20T10:00:00Z",
  "data": {
    "original": "Hello World"
  }
}
```

2. **廣播消息**:
```json
{"type":"broadcast","message":"Hello everyone!"}
```

**預期響應**:
```json
{
  "type": "sent",
  "message": "Message sent to all",
  "timestamp": "2025-11-20T10:00:00Z"
}
```

3. **發送給特定客戶端**:
```json
{"type":"message","to":"client-123456","message":"Private message"}
```

---

## 📋 測試場景

### 場景 1: 基本健康檢查

1. 啟動 server
2. 測試 `/ping` - 確認 server 運行
3. 測試 `/health` - 查看詳細狀態
4. 測試 `/api/v1/hello` - 確認 API 版本

**成功標準**: 所有 endpoints 返回 200 OK

---

### 場景 2: WebSocket Echo

1. 打開 WebSocket 連接 (`ws://localhost:8080/ws`)
2. 接收 welcome 消息並記錄 client ID
3. 發送 echo 消息
4. 驗證收到正確的 echo 響應

**成功標準**: Echo 消息正確返回

---

### 場景 3: 跨客戶端廣播

1. 打開兩個 WebSocket 連接（Client A, Client B）
2. 從 Client A 發送 broadcast 消息
3. 驗證 Client B 收到廣播消息
4. 驗證 Client A 也收到確認

**成功標準**: Client B 收到來自 Client A 的消息

---

### 場景 4: HTTP → WebSocket 消息中繼

1. 打開一個 WebSocket 連接
2. 記錄 client ID
3. 使用 Postman 發送 POST `/api/v1/message` 到該 client ID
4. 驗證 WebSocket 客戶端收到消息

**成功標準**: HTTP POST 的消息成功送達 WebSocket 客戶端

---

### 場景 5: 連接管理

1. 打開 3 個 WebSocket 連接
2. 調用 GET `/api/v1/connections` 查看連接列表
3. 關閉 1 個連接
4. 再次調用 GET `/api/v1/connections` 驗證數量減少

**成功標準**: 連接數正確追蹤

---

## 💻 curl 測試命令

### 完整測試腳本

```bash
#!/bin/bash

# 1. Ping
curl -X GET http://localhost:8080/ping

# 2. Health
curl -X GET http://localhost:8080/health

# 3. Hello
curl -X GET http://localhost:8080/api/v1/hello

# 4. Welcome
curl -X GET "http://localhost:8080/api/v1/welcome?name=Alice&device_id=iOS-1234"

# 5. Connections
curl -X GET http://localhost:8080/api/v1/connections

# 6. Send broadcast message
curl -X POST http://localhost:8080/api/v1/message \
  -H "Content-Type: application/json" \
  -d '{
    "from": "curl-client",
    "to": "all",
    "message": "Hello from curl!"
  }'

# 7. Send message to specific client
curl -X POST http://localhost:8080/api/v1/message \
  -H "Content-Type: application/json" \
  -d '{
    "from": "curl-client",
    "to": "client-12345",
    "message": "Private message from curl"
  }'
```

---

## 🧪 自動化測試腳本

我們提供了完整的自動化測試腳本：

```bash
# 運行所有測試
./tests/test_postman.sh
```

**測試內容**:
- ✅ Root endpoint
- ✅ Ping
- ✅ Health check
- ✅ Hello endpoint
- ✅ Welcome with parameters
- ✅ Connections list
- ✅ Send broadcast message
- ✅ Send to specific client (404 expected)

---

## 📊 預期結果總覽

| Endpoint | Method | 預期狀態 | 響應時間 |
|----------|--------|----------|----------|
| `/` | GET | 200 | < 1ms |
| `/ping` | GET | 200 | < 1ms |
| `/health` | GET | 200 | < 1ms |
| `/api/v1/hello` | GET | 200 | < 1ms |
| `/api/v1/welcome` | GET | 200 | < 1ms |
| `/api/v1/connections` | GET | 200 | < 1ms |
| `/api/v1/message` | POST | 200 | < 5ms |
| `/ws` | WebSocket | 101 | < 10ms |

---

## 🐛 常見問題

### Q1: 連接被拒絕 (Connection Refused)

**原因**: Server 未啟動

**解決方案**:
```bash
go run cmd/server/main_websocket.go
```

---

### Q2: WebSocket 升級失敗

**原因**: 可能是 CORS 或 Upgrade header 問題

**解決方案**: 確保使用正確的 WebSocket URL (`ws://localhost:8080/ws`)

---

### Q3: 消息未送達特定客戶端

**原因**: Client ID 錯誤或客戶端已斷開

**解決方案**:
1. 調用 GET `/api/v1/connections` 獲取正確的 client ID
2. 確認客戶端仍在線

---

### Q4: CORS 錯誤

**原因**: 瀏覽器安全策略

**解決方案**: Server 已配置 CORS 允許所有源，確認使用正確的協議 (http:// 或 ws://)

---

## 🎯 下一步

完成所有測試後，你可以：

1. ✅ **Phase 0 Day 2 完成** - WebSocket 基礎功能正常
2. ⏭️ **Phase 1** - 實作用戶識別和消息歷史
3. ⏭️ **Phase 2** - 添加 JWT 認證
4. ⏭️ **Phase 3** - 整合車站系統

---

**文檔版本**: 1.0
**最後更新**: 2025-11-20
**作者**: TrainBlink Team
