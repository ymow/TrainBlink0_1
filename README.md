# TrainBlink Messenger Protocol Research

基於 [trainblink](https://github.com/ymow/trainblink) 的研究項目，用於實作複雜的 Go server。

## 專案概述

本專案旨在為 TrainBlink 開發一個可選的後端服務器，以增強 P2P 架構並提供額外功能。

### 主要目標

1. 研究和分析 TrainBlink 的消息協議
2. 設計並實作 Go 後端服務器
3. 提供內容審核、匹配和同步等增強功能
4. 保持與原始 P2P 架構的兼容性

## 文檔

- **[TrainBlink 分析報告](TRAINBLINK_ANALYSIS.md)** - 完整的系統分析
  - 數據模型
  - 通訊協議
  - 系統架構
  - Server 設計建議

## 項目結構

```
messenger_protocol_research/
├── cmd/
│   └── server/              # 主服務器入口
├── internal/
│   ├── api/                 # HTTP/WebSocket API
│   │   ├── handlers/        # 請求處理器
│   │   └── middleware/      # 中間件
│   ├── service/             # 業務邏輯
│   ├── repository/          # 數據訪問層
│   ├── model/               # 數據模型
│   └── ws/                  # WebSocket 管理
├── pkg/                     # 可重用包
│   ├── jwt/                 # JWT 認證
│   ├── logger/              # 日誌
│   └── validator/           # 驗證
├── config/                  # 配置文件
├── migrations/              # 數據庫遷移
└── scripts/                 # 工具腳本
```

## 技術棧

- **語言**: Go 1.21+
- **Web 框架**: Gin
- **WebSocket**: Gorilla WebSocket
- **數據庫**: PostgreSQL + GORM
- **緩存**: Redis
- **認證**: JWT
- **日誌**: Zap
- **配置**: Viper

## 快速開始

### 前置需求

- Go 1.21 或更高版本
- PostgreSQL 15+
- Redis 7+
- Docker (可選)

### 安裝依賴

```bash
go mod download
```

### 配置

1. 複製配置範例：
```bash
cp config/config.example.yaml config/config.yaml
```

2. 編輯 `config/config.yaml` 並設置你的數據庫和 Redis 連接。

### 運行

```bash
# 開發模式
go run cmd/server/main.go

# 編譯
go build -o bin/server cmd/server/main.go

# 運行編譯後的二進制
./bin/server
```

### 使用 Docker

```bash
# 構建鏡像
docker build -t trainblink-server .

# 運行容器
docker-compose up -d
```

## 開發

### 運行測試

```bash
# 所有測試
go test ./...

# 帶覆蓋率
go test -cover ./...

# 生成覆蓋率報告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 代碼格式化

```bash
# 格式化代碼
go fmt ./...

# 靜態檢查
go vet ./...

# 使用 golangci-lint
golangci-lint run
```

## API 文檔

### 認證

```http
POST /api/v1/auth/anonymous
```

生成臨時匿名 token。

**響應**:
```json
{
  "token": "jwt-token",
  "userId": "temp-uuid",
  "expiresAt": 1700000000
}
```

### 車站

```http
GET /api/v1/stations
```

獲取所有車站列表。

```http
GET /api/v1/stations/{stationId}/peers
```

獲取指定車站內的在線節點。

### 消息

```http
POST /api/v1/messages
```

發送消息到指定節點。

```http
GET /api/v1/messages/{peerId}
```

獲取與指定節點的消息歷史。

### WebSocket

```
ws://server/ws?token=jwt-token&stationId=1001
```

實時消息和節點發現。

詳細 API 文檔請參考 [API.md](docs/API.md)（待完成）。

## 路線圖

### Phase 1: 基礎架構 (當前)
- [x] 項目結構設置
- [x] 分析 TrainBlink 協議
- [ ] 基本 HTTP 服務器
- [ ] WebSocket 支持
- [ ] 數據庫設計和遷移
- [ ] JWT 認證

### Phase 2: 核心功能
- [ ] 車站節點發現
- [ ] 消息中繼服務
- [ ] 實時通訊
- [ ] Redis 緩存

### Phase 3: 增強功能
- [ ] 內容審核服務
- [ ] 相遇記錄聚合
- [ ] 封鎖與舉報管理
- [ ] 分析儀表板

### Phase 4: 優化與部署
- [ ] 性能優化
- [ ] 負載測試
- [ ] Docker 容器化
- [ ] Kubernetes 部署

## 貢獻

歡迎貢獻！請遵循以下步驟：

1. Fork 此倉庫
2. 創建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 開啟 Pull Request

## 許可證

Copyright © 2025. All rights reserved.

## 致謝

- 基於 [TrainBlink](https://github.com/ymow/trainblink) 項目
- 感謝 TrainBlink 團隊提供的設計靈感

---

**最後更新**: 2025-11-20
**狀態**: 🚧 開發中
