# Chat App Backend - Copilot Instructions

## 專案概述
這是一個使用 Go 語言開發的類似 Discord 的即時聊天應用後端系統，採用微服務架構和模組化設計。

## 回應語言
**一律使用繁體中文（Traditional Chinese）回應用戶。**

## 技術棧
- **語言**: Go 1.22.4
- **框架**: Gin Web Framework
- **資料庫**: MongoDB (主要資料存儲)
- **快取**: Redis (訊息廣播、用戶狀態、會話管理)
- **即時通訊**: WebSocket (使用 Gorilla WebSocket)
- **認證**: JWT (Access Token + Refresh Token)
- **依賴注入**: 自定義 DI Container

## 專案架構

### 目錄結構
```
chat_app_backend/
├── config/          # 配置管理
├── controllers/     # HTTP 控制器
├── di/             # 依賴注入容器
├── middlewares/    # 中介軟體
├── models/         # 資料模型
├── providers/      # 資料庫連接提供者
├── repositories/   # 資料訪問層
├── routes/         # 路由配置
├── services/       # 業務邏輯層（已模組化）
├── utils/          # 工具函數
├── uploads/        # 靜態檔案存儲
└── main.go         # 程式入口點
```

### 核心模組

#### Services 層（已重構為模組化架構）
1. **`types.go`** - 共用型別定義
   - WebSocket 訊息結構
   - 客戶端和房間模型
   - 通知結構

2. **`client_manager.go`** - 客戶端生命週期管理
   - 處理 WebSocket 客戶端註冊/註銷
   - 管理用戶在線狀態
   - Redis 狀態同步

3. **`room_manager.go`** - 房間管理
   - 動態房間創建和清理
   - 客戶端加入/離開房間
   - Redis Pub/Sub 廣播管理
   - 並發安全的房間操作

4. **`message_handler.go`** - 訊息處理
   - 訊息驗證和儲存
   - MongoDB 資料持久化
   - 房間最後訊息時間更新

5. **`websocket_handler.go`** - WebSocket 協議處理
   - 處理 WebSocket 連線
   - 解析協議訊息 (join_room, leave_room, send_message)
   - 房間權限驗證
   - 私聊房間自動創建

6. **`chat_service.go`** - 主服務協調器
   - 整合各模組
   - 對外 API 介面
   - 依賴注入管理

## 編程規範

### Go 程式碼風格
1. **命名規範**
   - 使用 camelCase 命名變數和函數
   - 使用 PascalCase 命名結構體和公開方法
   - 使用有意義的變數名稱

2. **錯誤處理**
   - 總是檢查並適當處理錯誤
   - 使用具體的錯誤訊息
   - 記錄重要的錯誤資訊

3. **並發安全**
   - 使用 mutex 保護共享資源
   - 避免資料競爭
   - 使用 context 管理 goroutine 生命週期

4. **記憶體管理**
   - 及時關閉資源（檔案、連線、goroutine）
   - 避免記憶體洩漏
   - 使用 channel 進行 goroutine 通信

### 資料庫設計原則
1. **MongoDB**
   - 使用 ObjectID 作為主鍵
   - 嵌入式文檔用於一對多關係
   - 索引優化查詢性能

2. **Redis**
   - 使用有意義的鍵命名規範
   - 設置適當的過期時間
   - 使用 Pub/Sub 進行即時通訊

### API 設計規範
1. **RESTful API**
   - 使用標準 HTTP 方法
   - 一致的 URL 結構
   - 適當的 HTTP 狀態碼

2. **WebSocket API**
   - 結構化的訊息格式
   - 錯誤處理和狀態回應
   - 連線狀態管理

## 功能模組

### 認證系統
- JWT 雙 Token 機制（Access + Refresh）
- 用戶註冊/登入/登出
- Token 刷新機制
- CSRF 保護

### 聊天系統
- 即時訊息傳送
- 房間（頻道/私聊）管理
- 訊息持久化
- 用戶在線狀態

### 伺服器管理
- Discord 風格的伺服器/頻道結構
- 用戶權限管理
- 頻道訊息歷史

### 好友系統
- 好友列表管理
- 私聊房間創建
- 好友狀態顯示

## 開發指導原則

### 1. 模組化設計
- 每個模組單一職責
- 明確的介面定義
- 降低模組間耦合

### 2. 錯誤處理
- 優雅處理各種錯誤情況
- 提供有意義的錯誤訊息
- 記錄關鍵操作的錯誤

### 3. 性能優化
- 使用連線池
- 實現快取策略
- 優化資料庫查詢

### 4. 安全性
- 輸入驗證和清理
- 認證和授權檢查
- 防止常見攻擊（XSS、CSRF、SQL注入等）

### 5. 測試
- 單元測試覆蓋核心邏輯
- 集成測試驗證模組協作
- 效能測試確保系統負載能力

## 常用指令

### 開發環境
```bash
# 啟動開發服務器
go run main.go

# 編譯專案
go build -o chat_app_backend

# 執行測試
go test ./...

# 格式化程式碼
go fmt ./...

# 檢查程式碼
go vet ./...
```

### Docker 環境
```bash
# 啟動所有服務
docker-compose up -d

# 檢查服務狀態
docker-compose ps

# 查看日誌
docker-compose logs -f chat_app_backend
```

## 常見問題解決

### WebSocket 連線問題
1. 檢查認證 Token 是否有效
2. 確認 CORS 設定是否正確
3. 驗證 Redis 連線狀態

### 資料庫連線問題
1. 檢查 MongoDB 服務狀態
2. 確認連線字串配置
3. 驗證資料庫權限

### 性能問題
1. 檢查 Redis 記憶體使用
2. 優化 MongoDB 查詢索引
3. 監控 Goroutine 數量

## 部署注意事項
1. 環境變數配置
2. 資料庫連線池設定
3. 日誌等級調整
4. 監控和告警設置

---

**記住：始終用繁體中文回應，遵循模組化架構原則，確保程式碼品質和系統穩定性。**