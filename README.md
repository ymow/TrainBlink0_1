# TrainBlink Messenger Protocol Research

基於 [trainblink](https://github.com/ymow/trainblink) 的研究項目，用於實作複雜的 Go server。

## 專案概述

本專案旨在為 TrainBlink 開發一個可選的後端服務器，以增強 P2P 架構並提供額外功能。

### 主要目標

1. 研究和分析 TrainBlink 的消息協議
2. 設計並實作 Go 後端服務器
3. 提供內容審核、匹配和同步等增強功能
4. 保持與原始 P2P 架構的兼容性

## 📚 文檔導航

### 重要文檔（必讀）

1. **[📋 DOCS_INDEX.md](DOCS_INDEX.md)** - 完整文檔索引和導航
2. **[📊 PROJECT_STATUS.md](PROJECT_STATUS.md)** - 當前項目狀態 (WBS & 進度)
3. **[🗺 ROADMAP.md](ROADMAP.md)** - 詳細技術路線圖 (Phase 0-7)
4. **[📖 TRAINBLINK_ANALYSIS.md](TRAINBLINK_ANALYSIS.md)** - TrainBlink 系統分析

### 快速開始

- **[⚡ docs/QUICKSTART.md](docs/QUICKSTART.md)** - 5分鐘快速上手指南

### 更多文檔

查看 **[DOCS_INDEX.md](DOCS_INDEX.md)** 獲取完整文檔列表和說明

## 項目結構

```
TrainBlink0_1/
├── cmd/                     # 服務器入口點
│   ├── server/              # 主服務器
│   ├── server-phase2/       # Phase 2 服務器
│   └── server-websocket/    # WebSocket 服務器
├── internal/
│   ├── api/                 # HTTP/WebSocket API
│   ├── auth/                # 認證服務
│   ├── cache/               # Redis 緩存
│   ├── cleanup/             # 數據清理服務
│   ├── config/              # 配置管理
│   ├── database/            # 數據庫連接
│   ├── discovery/           # BLE 發現服務
│   ├── geofence/            # 地理圍欄服務
│   ├── matrix/              # Matrix 橋接服務
│   ├── mls/                 # MLS 加密服務
│   ├── model/               # 數據模型
│   ├── trip/                # 行程管理
│   ├── user/                # 用戶管理
│   └── websocket/           # WebSocket 管理
├── trainblink-web/          # 🆕 Web 前端應用
│   ├── src/
│   │   ├── components/      # React 組件
│   │   ├── hooks/           # React Hooks
│   │   ├── pages/           # 頁面組件
│   │   ├── services/        # API 服務
│   │   └── types/           # TypeScript 類型
│   ├── public/              # 靜態資源
│   └── package.json         # 前端依賴
├── mobile/                  # 移動端應用
│   ├── ios/                 # iOS 應用
│   └── android/             # Android 應用
├── config/                  # 配置文件
├── migrations/              # 數據庫遷移
├── docs/                    # 文檔
└── scripts/                 # 工具腳本
```

## 技術棧

### 後端 (Go)
- **語言**: Go 1.21+
- **Web 框架**: Gin
- **WebSocket**: Gorilla WebSocket
- **數據庫**: PostgreSQL + GORM
- **緩存**: Redis
- **認證**: JWT
- **日誌**: Zap
- **配置**: Viper
- **加密**: Matrix MLS

### 🆕 前端 (Web)
- **框架**: React 19.2+ + TypeScript
- **構建工具**: Vite 7.2+
- **狀態管理**: TanStack React Query
- **路由**: React Router DOM
- **動畫**: Framer Motion
- **HTTP 客戶端**: Axios
- **樣式**: CSS-in-JS + Glass Morphism
- **WebSocket**: 原生 WebSocket API

### 移動端
- **iOS**: Swift + SwiftUI
- **Android**: Kotlin + Jetpack Compose

## 快速開始

### 前置需求

#### 後端
- Go 1.21 或更高版本
- PostgreSQL 15+
- Redis 7+
- Docker (可選)

#### 🆕 前端
- Node.js 18+ 
- npm 或 yarn
- 現代瀏覽器 (Chrome, Firefox, Safari, Edge)

### 安裝依賴

#### 後端依賴
```bash
go mod download
```

#### 🆕 前端依賴
```bash
cd trainblink-web
npm install
```

### 配置

1. 複製配置範例：
```bash
cp config/config.example.yaml config/config.yaml
```

2. 編輯 `config/config.yaml` 並設置你的數據庫和 Redis 連接。

### 運行

#### 後端服務器

```bash
# 設置環境變量
export JWT_SECRET_KEY="dev-secret-key-for-testing-only-change-in-production"
export DATABASE_URL="sqlite://trainblink.db"
export REDIS_URL="redis://localhost:6379/0"

# 開發模式
make run

# 或直接運行
go run cmd/server/main.go

# 編譯並運行
make build
./bin/trainblink-server
```

#### 🆕 前端開發服務器

```bash
cd trainblink-web
npm run dev
```

**訪問地址**:
- 🌐 **Web 前端**: http://localhost:5173
- ⚡ **後端 API**: http://localhost:8080
- 💬 **WebSocket**: ws://localhost:8080/ws

#### 完整啟動流程

```bash
# 1. 啟動後端 (Terminal 1)
export JWT_SECRET_KEY="dev-secret-key-for-testing-only-change-in-production"
export DATABASE_URL="sqlite://trainblink.db"
export REDIS_URL="redis://localhost:6379/0"
make run

# 2. 啟動前端 (Terminal 2)  
cd trainblink-web
npm run dev
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

## 🆕 Web 前端功能

TrainBlink Web 應用提供了完整的瀏覽器端體驗：

### 主要功能
- 🏠 **響應式主頁** - 包含動態英雄區域和實時統計
- 💬 **即時聊天** - WebSocket 連接的實時消息系統  
- 📊 **實時狀態監控** - 服務器健康狀況和連接統計
- 🎨 **現代 UI 設計** - Glass Morphism 效果和流暢動畫
- 📱 **移動端適配** - 響應式設計支持所有設備

### 技術特色
- ⚡ **快速熱重載** - Vite 開發服務器
- 🔗 **類型安全** - 全 TypeScript 支持
- 🔄 **智能緩存** - React Query 狀態管理
- 🌊 **平滑動畫** - Framer Motion 動效
- 🔒 **API 集成** - 完整的後端服務連接

### 頁面結構
- `/` - 主頁 (英雄區域 + 統計展示)
- `/chat` - 聊天界面 (用戶名設置 + 實時聊天)

## 📊 當前狀態

**整體進度**: 58% 完成

詳細狀態和進度請查看 **[PROJECT_STATUS.md](PROJECT_STATUS.md)**

### 已完成
- ✅ 基礎架構 (HTTP Server, WebSocket, Database, JWT)
- ✅ Web 前端應用 (React + TypeScript + Vite)
- ✅ 移動端應用 (iOS + Android)
- ✅ BLE 發現服務
- ✅ Matrix 橋接服務
- ✅ MLS 端到端加密
- ✅ 地理圍欄功能
- ✅ Redis 緩存
- ✅ 自動數據清理

### 進行中
- 🚧 消息系統完善 (歷史記錄、離線隊列、已讀回執)
- 🚧 移動端 WebSocket 整合
- 🚧 車站系統 API

### 計劃中
- 📋 WebTransport 整合
- 📋 內容審核服務
- 📋 用戶安全功能 (封鎖、舉報)
- 📋 智能行程檢測
- 📋 多語言支持 (i18n)
- 📋 PWA 支持

**詳細路線圖**: 請查看 [ROADMAP.md](ROADMAP.md) 和 [PROJECT_STATUS.md](PROJECT_STATUS.md)

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

**最後更新**: 2025-12-05
**整體進度**: 🚀 58% 完成

### 🎯 系統狀態
- ✅ **後端 API** - Go 服務器 (HTTP + WebSocket)
- ✅ **Web 前端** - React 應用 (實時聊天 + 響應式設計)
- ⚠️ **移動端** - iOS/Android 應用 (BLE + Matrix 集成，WebSocket 待補完)
- ✅ **數據庫** - PostgreSQL + Redis + SQLite (自動清理機制)
- ✅ **加密** - Matrix MLS 端到端加密
- ✅ **發現** - BLE 匿名發現 + 地理圍欄

### 📚 完整文檔
查看 [DOCS_INDEX.md](DOCS_INDEX.md) 獲取所有文檔索引
