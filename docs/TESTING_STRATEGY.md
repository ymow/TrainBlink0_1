# TrainBlink 測試策略

**版本**: 1.0  
**最後更新**: 2025-11-20

---

## 📋 概述

本文檔定義了 TrainBlink 通訊協議的完整測試策略，涵蓋單元測試、集成測試、端到端測試和性能測試。

---

## Phase 0: HTTP REST API 測試

### Postman 測試集

已提供完整的 Postman Collection：
- **文件**: `tests/TrainBlink_Complete.postman_collection.json`
- **測試數**: 12 個
- **覆蓋率**: 100%

### 測試流程

```bash
# 1. 啟動服務器
go run cmd/server/main_with_post.go

# 2. 執行自動化測試
./tests/test_all_endpoints.sh

# 3. 或使用 Postman
# 導入 TrainBlink_Complete.postman_collection.json
# 按順序執行所有請求
```

### 測試場景

| 測試 | 端點 | 預期結果 |
|------|------|----------|
| 1. 健康檢查 | GET /health | 200 OK, 服務器狀態 |
| 2. 註冊 Alice | POST /api/v1/clients | 201 Created, client_id |
| 3. 註冊 Bob | POST /api/v1/clients | 201 Created, client_id |
| 4. 列出客戶端 | GET /api/v1/clients | 200 OK, 2 客戶端 |
| 5. Alice → Bob | POST /api/v1/messages | 201 Created, message_id |
| 6. Bob → Alice | POST /api/v1/messages | 201 Created, message_id |
| 7. 系統廣播 | POST /api/v1/messages | 201 Created, to=all |
| 8. 獲取所有消息 | GET /api/v1/messages | 200 OK, 3 條消息 |
| 9. Alice 的消息 | GET /api/v1/messages/user?id=alice_id | 200 OK, 3 條 |
| 10. Bob 的消息 | GET /api/v1/messages/user?id=bob_id | 200 OK, 3 條 |

### curl 測試示例

```bash
# 1. 註冊客戶端
CLIENT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/clients \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","device_id":"iOS-1234"}')

CLIENT_ID=$(echo $CLIENT_RESPONSE | jq -r '.client.id')

# 2. 發送消息
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d "{\"from\":\"$CLIENT_ID\",\"to\":\"all\",\"text\":\"Hello World\"}"

# 3. 獲取消息
curl "http://localhost:8080/api/v1/messages/user?id=$CLIENT_ID"
```

---

## Phase 1: WebSocket 測試

### WebSocket 測試工具

**wscat** (命令行):
```bash
# 安裝
npm install -g wscat

# 連接
wscat -c "ws://localhost:8080/ws?client_id=client-123"

# 發送消息
> {"type":"send_message","payload":{"to":"all","text":"Hello"}}

# 接收響應
< {"type":"message_ack","payload":{...}}
```

**瀏覽器測試**:
```html
<!DOCTYPE html>
<html>
<head>
    <title>WebSocket Test</title>
</head>
<body>
    <script>
        const ws = new WebSocket('ws://localhost:8080/ws?client_id=test-client');
        
        ws.onopen = () => {
            console.log('✅ Connected');
            
            // 發送消息
            ws.send(JSON.stringify({
                type: 'send_message',
                payload: {
                    to: 'all',
                    text: 'Hello from browser!'
                }
            }));
        };
        
        ws.onmessage = (event) => {
            console.log('📨 Received:', JSON.parse(event.data));
        };
        
        ws.onerror = (error) => {
            console.error('❌ Error:', error);
        };
        
        ws.onclose = () => {
            console.log('👋 Disconnected');
        };
    </script>
</body>
</html>
```

### 測試場景

| 測試 | 描述 | 預期結果 |
|------|------|----------|
| 連接建立 | WebSocket 握手 | 收到 welcome 消息 |
| 心跳 | 每 30 秒 ping | 收到 pong 響應 |
| 發送消息 | send_message | 收到 message_ack |
| 接收消息 | 等待 new_message | 正確解析消息 |
| 送達回執 | message_delivered | 原發送者收到 delivery_receipt |
| 已讀回執 | message_read | 原發送者收到 read_receipt |
| 輸入狀態 | typing 事件 | 對方收到 typing_indicator |
| 斷線重連 | 模擬網絡斷開 | 自動重連成功 |
| 消息同步 | 重連後 sync | 收到缺失的消息 |

---

## Phase 2: 認證測試

### JWT Token 測試

```bash
# 1. 獲取 Token
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/anonymous \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","device_id":"iOS-1234"}')

TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.auth.access_token')

# 2. 使用 Token 訪問 API
curl http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer $TOKEN"

# 3. 測試無效 Token
curl http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer invalid-token"
# 預期: 401 Unauthorized

# 4. 測試過期 Token
# (等待 12 小時後測試)
# 預期: 401 Unauthorized, "Token expired"

# 5. 刷新 Token
REFRESH_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.auth.refresh_token')
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}"
```

### 測試矩陣

| 場景 | Token 狀態 | 預期結果 |
|------|-----------|----------|
| 正常請求 | 有效 Token | 200 OK |
| 無 Token | 缺失 | 401 Unauthorized |
| 無效 Token | 格式錯誤 | 401 Unauthorized |
| 過期 Token | 已過期 | 401 Unauthorized |
| 篡改 Token | 簽名錯誤 | 401 Unauthorized |
| 刷新成功 | 有效 Refresh | 200 OK, 新 Token |
| 刷新失敗 | 無效 Refresh | 401 Unauthorized |

---

## Phase 3: 地理圍欄測試

### 測試數據

```json
// 台北車站（1001）
{
  "lat": 25.0478,
  "lng": 121.5170,
  "radius": 500  // 米
}

// 測試位置
{
  "inside": {"lat": 25.0478, "lng": 121.5170},      // 0m, 內部
  "edge": {"lat": 25.0523, "lng": 121.5170},         // ~500m, 邊緣
  "outside": {"lat": 25.0578, "lng": 121.5170}       // ~1100m, 外部
}
```

### 測試場景

```bash
# 1. 成功進入車站（內部）
curl -X POST http://localhost:8080/api/v1/stations/1001/join \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"lat":25.0478,"lng":121.5170}'
# 預期: 200 OK, "joined"

# 2. 邊緣測試（約 500m）
curl -X POST http://localhost:8080/api/v1/stations/1001/join \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"lat":25.0523,"lng":121.5170}'
# 預期: 200 OK 或 400 Bad Request（取決於容差）

# 3. 拒絕進入（外部）
curl -X POST http://localhost:8080/api/v1/stations/1001/join \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"lat":25.0578,"lng":121.5170}'
# 預期: 400 Bad Request, "out_of_range"
```

### 距離計算驗證

使用 Haversine 公式驗證距離計算：

```go
// 測試：台北車站到測試點的距離
func TestDistanceCalculation(t *testing.T) {
    station := Station{
        Lat: 25.0478,
        Lng: 121.5170,
    }
    
    testCases := []struct {
        name     string
        lat      float64
        lng      float64
        expected float64
        delta    float64
    }{
        {"Same location", 25.0478, 121.5170, 0, 1},
        {"500m north", 25.0523, 121.5170, 500, 10},
        {"1km north", 25.0578, 121.5170, 1100, 50},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            distance := calculateDistance(
                station.Lat, station.Lng,
                tc.lat, tc.lng,
            )
            
            if math.Abs(distance-tc.expected) > tc.delta {
                t.Errorf("Expected ~%.0fm, got %.0fm", tc.expected, distance)
            }
        })
    }
}
```

---

## 跨平台互通測試

### 測試矩陣

| 發送者 | 接收者 | 協議 | 預期結果 |
|-------|-------|------|----------|
| iOS | Android | HTTP | ✅ 消息送達 |
| Android | iOS | HTTP | ✅ 消息送達 |
| iOS | Android | WebSocket | ✅ 即時送達 |
| Android | iOS | WebSocket | ✅ 即時送達 |
| iOS | iOS | WebSocket | ✅ 即時送達 |
| Android | Android | WebSocket | ✅ 即時送達 |

### 測試流程

```
1. 啟動 Go Server
2. 啟動 iOS Simulator / 真機
3. 啟動 Android Emulator / 真機
4. iOS 註冊 (Alice)
5. Android 註冊 (Bob)
6. Alice 發送消息給 Bob
7. 驗證 Bob 收到消息
8. Bob 回覆 Alice
9. 驗證 Alice 收到消息
10. 測試廣播消息
11. 驗證雙方都收到
```

### 自動化測試腳本

**test_cross_platform.sh**:
```bash
#!/bin/bash

SERVER_URL="http://localhost:8080"

echo "🚀 Starting Cross-Platform Test"

# 1. 註冊 iOS 客戶端
echo "📱 Registering iOS client..."
IOS_RESPONSE=$(curl -s -X POST $SERVER_URL/api/v1/clients \
  -H "Content-Type: application/json" \
  -d '{"name":"iOS-Alice","device_id":"iOS-1234"}')

IOS_ID=$(echo $IOS_RESPONSE | jq -r '.client.id')
echo "✅ iOS Client: $IOS_ID"

# 2. 註冊 Android 客戶端
echo "🤖 Registering Android client..."
ANDROID_RESPONSE=$(curl -s -X POST $SERVER_URL/api/v1/clients \
  -H "Content-Type: application/json" \
  -d '{"name":"Android-Bob","device_id":"Android-5678"}')

ANDROID_ID=$(echo $ANDROID_RESPONSE | jq -r '.client.id')
echo "✅ Android Client: $ANDROID_ID"

# 3. iOS → Android
echo "📤 iOS sending to Android..."
curl -s -X POST $SERVER_URL/api/v1/messages \
  -H "Content-Type: application/json" \
  -d "{\"from\":\"$IOS_ID\",\"to\":\"$ANDROID_ID\",\"text\":\"Hello from iOS!\"}" > /dev/null

# 4. Android → iOS
echo "📤 Android sending to iOS..."
curl -s -X POST $SERVER_URL/api/v1/messages \
  -H "Content-Type: application/json" \
  -d "{\"from\":\"$ANDROID_ID\",\"to\":\"$IOS_ID\",\"text\":\"Hello from Android!\"}" > /dev/null

# 5. 驗證消息
echo "🔍 Verifying messages..."
IOS_MESSAGES=$(curl -s "$SERVER_URL/api/v1/messages/user?id=$IOS_ID" | jq '.count')
ANDROID_MESSAGES=$(curl -s "$SERVER_URL/api/v1/messages/user?id=$ANDROID_ID" | jq '.count')

if [ "$IOS_MESSAGES" -eq 2 ] && [ "$ANDROID_MESSAGES" -eq 2 ]; then
    echo "✅ Cross-platform test PASSED"
    exit 0
else
    echo "❌ Cross-platform test FAILED"
    echo "iOS messages: $IOS_MESSAGES (expected 2)"
    echo "Android messages: $ANDROID_MESSAGES (expected 2)"
    exit 1
fi
```

---

## 性能測試

### 負載測試工具

**使用 Apache Bench**:
```bash
# 1. 註冊客戶端 (100 個請求, 10 並發)
ab -n 100 -c 10 -p client.json -T application/json \
  http://localhost:8080/api/v1/clients

# 2. 發送消息 (1000 個請求, 50 並發)
ab -n 1000 -c 50 -p message.json -T application/json \
  http://localhost:8080/api/v1/messages
```

**使用 wrk**:
```bash
# 安裝 wrk
brew install wrk  # macOS
apt install wrk   # Ubuntu

# 負載測試
wrk -t12 -c400 -d30s --latency http://localhost:8080/api/v1/clients
```

### 性能指標

| 階段 | 指標 | 目標值 | 測試方法 |
|------|------|--------|----------|
| Phase 0 | HTTP 響應時間 (P95) | < 10ms | Apache Bench |
| Phase 0 | 吞吐量 | > 1000 req/s | wrk |
| Phase 1 | WebSocket 延遲 | < 100ms | 自定義腳本 |
| Phase 1 | 並發連接數 | > 10,000 | wscat 循環 |
| Phase 2 | 數據庫查詢 (P95) | < 10ms | PostgreSQL explain |
| Phase 2 | Redis 緩存命中率 | > 90% | Redis INFO |
| Phase 5 | MLS 加密延遲 | < 50ms | 基準測試 |
| Phase 6 | WebTransport 延遲 | < 60ms | 基準測試 |

### WebSocket 性能測試

**test_websocket_load.js** (Node.js):
```javascript
const WebSocket = require('ws');

const CONNECTIONS = 100;
const MESSAGES_PER_CONNECTION = 10;

let connectedCount = 0;
let sentCount = 0;
let receivedCount = 0;
const startTime = Date.now();

for (let i = 0; i < CONNECTIONS; i++) {
    const ws = new WebSocket(`ws://localhost:8080/ws?client_id=test-${i}`);
    
    ws.on('open', () => {
        connectedCount++;
        
        // 發送消息
        for (let j = 0; j < MESSAGES_PER_CONNECTION; j++) {
            ws.send(JSON.stringify({
                type: 'send_message',
                payload: {
                    to: 'all',
                    text: `Message ${j} from client ${i}`
                }
            }));
            sentCount++;
        }
    });
    
    ws.on('message', (data) => {
        receivedCount++;
        
        // 完成後報告
        if (receivedCount === CONNECTIONS * MESSAGES_PER_CONNECTION) {
            const duration = (Date.now() - startTime) / 1000;
            console.log(`✅ Test completed in ${duration}s`);
            console.log(`📊 Stats:`);
            console.log(`   - Connections: ${connectedCount}`);
            console.log(`   - Messages sent: ${sentCount}`);
            console.log(`   - Messages received: ${receivedCount}`);
            console.log(`   - Throughput: ${(receivedCount / duration).toFixed(2)} msg/s`);
            process.exit(0);
        }
    });
    
    ws.on('error', (error) => {
        console.error(`❌ Error:`, error.message);
    });
}
```

---

## 持續集成 (CI)

### GitHub Actions 配置

**.github/workflows/test.yml**:
```yaml
name: TrainBlink Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run unit tests
        run: go test -v -cover ./...
      
      - name: Start server
        run: |
          go build -o server cmd/server/main_with_post.go
          ./server &
          sleep 5
      
      - name: Run integration tests
        run: ./tests/test_all_endpoints.sh
      
      - name: Run cross-platform tests
        run: ./tests/test_cross_platform.sh
```

---

## 測試覆蓋率目標

| 組件 | 目標覆蓋率 |
|------|-----------|
| API Handlers | 90% |
| Business Logic | 95% |
| Data Models | 95% |
| WebSocket Hub | 85% |
| Authentication | 95% |
| 整體 | 90% |

---

**文檔維護者**: TrainBlink Team  
**最後更新**: 2025-11-20  
**版本**: 1.0
