# K6 負載測試套件

這是一個完整的 k6 負載測試套件，用於測試 Chat App Backend 的性能和穩定性。

## 📁 目錄結構

```
loadtest/
├── README.md                    # 本文件
├── SMOKE_TEST_GUIDE.md         # Smoke Test 詳細指南
├── config.js                    # 測試配置（端點、閾值等）
├── run.js                       # 主執行文件
├── scenarios/                   # 測試場景
│   ├── smoke.js                # 冒煙測試
│   ├── light.js                # 輕量負載測試
│   ├── medium.js               # 中等負載測試
│   └── heavy.js                # 重量負載測試
├── scripts/                     # 測試腳本模組
│   ├── common/                 # 通用工具
│   │   ├── auth.js            # 認證輔助函數
│   │   ├── csrf.js            # CSRF Token 處理
│   │   ├── logger.js          # 日誌工具
│   │   └── utils.js           # 工具函數
│   ├── api/                    # API 測試
│   │   ├── auth.js            # 認證 API 測試
│   │   ├── servers.js         # 伺服器管理測試
│   │   ├── friends.js         # 好友系統測試
│   │   ├── chat.js            # 聊天功能測試
│   │   └── upload.js          # 檔案上傳測試
│   └── websocket/              # WebSocket 測試
│       ├── connection.js      # WebSocket 連線
│       ├── rooms.js           # 房間管理
│       └── messaging.js       # 訊息發送
├── data/                        # 測試資料
│   ├── users.json              # 測試用戶
│   ├── messages.json           # 測試訊息
│   └── scenarios.json          # 測試場景資料
└── test_results/               # 測試結果輸出
    └── load_tests/             # 負載測試結果
```

## 🚀 快速開始

### 1. 安裝 k6

```bash
# macOS
brew install k6

# Linux
sudo apt-get install k6

# 或從官網下載
https://k6.io/docs/get-started/installation/
```

### 2. 啟動應用程式

確保應用程式正在運行：

```bash
# 檢查健康狀態
curl http://localhost:8111/health

# 或使用 docker-compose 啟動
cd ..
docker-compose up -d
```

### 3. 執行測試

```bash
cd loadtest

# 執行 Smoke Test（冒煙測試）
k6 run --env SCENARIO=smoke run.js

# 執行 Light Test（輕量負載）
k6 run --env SCENARIO=light run.js

# 執行 Medium Test（中等負載）
k6 run --env SCENARIO=medium run.js

# 執行 Heavy Test（重量負載）
k6 run --env SCENARIO=heavy run.js
```

## 📊 測試場景說明

### Smoke Test（冒煙測試）✅ 已完成

**目的**：驗證所有核心功能是否正常運作

- **虛擬用戶（VU）**：1 個
- **持續時間**：15 秒
- **測試內容**：
  - ✅ 認證系統（註冊、登入）
  - ✅ 伺服器管理（創建、查詢、更新）
  - ✅ 頻道管理（創建、查詢）
  - ✅ 好友系統（列表查詢）
  - ✅ 聊天功能（房間列表）

**執行命令**：
```bash
k6 run --env SCENARIO=smoke run.js
```

**預期結果**：
- HTTP 平均響應時間 < 20ms
- HTTP p(95) 響應時間 < 100ms
- Check 成功率 > 50%（某些測試使用假 ID 預期失敗）

### Light Test（輕量負載測試）

**目的**：測試系統在小規模負載下的表現

- **虛擬用戶（VU）**：5-10 個
- **持續時間**：2 分鐘
- **測試內容**：同 Smoke Test，但並發執行

### Medium Test（中等負載測試）

**目的**：測試系統在正常負載下的性能

- **虛擬用戶（VU）**：20-50 個
- **持續時間**：5 分鐘
- **測試內容**：完整的 API 和 WebSocket 測試

### Heavy Test（重量負載測試）

**目的**：測試系統的極限和瓶頸

- **虛擬用戶（VU）**：50-200 個
- **持續時間**：10 分鐘
- **測試內容**：高併發情況下的系統穩定性

## 🔧 配置選項

### 環境變數

```bash
# 指定測試場景
--env SCENARIO=smoke|light|medium|heavy

# 指定 API 基礎 URL
--env BASE_URL=http://localhost:8111

# 指定 WebSocket URL
--env WS_URL=ws://localhost:8111/ws

# 啟用詳細日誌
--env VERBOSE=1
```

### 修改配置

編輯 `config.js` 來調整：
- 測試端點
- 性能閾值
- 測試階段（stages）
- 測試用戶

## 📈 查看測試結果

### 即時查看

測試執行過程中會顯示即時日誌：
```
✅ POST /register | Status: 200 | Duration: 97.66ms
✅ POST /login | Status: 200 | Duration: 71.22ms
✅ GET /servers | Status: 200 | Duration: 3.45ms
```

### 測試報告

測試完成後，報告保存在 `test_results/load_tests/` 目錄：

```bash
# 查看最新的 Markdown 報告
cat test_results/load_tests/*_smoke_summary.md

# 查看 JSON 詳細結果
cat test_results/load_tests/*_smoke_results.json
```

### k6 Cloud（可選）

如果需要更詳細的視覺化報告，可以使用 k6 Cloud：

```bash
# 註冊並登入 k6 Cloud
k6 login cloud

# 上傳測試結果
k6 run --out cloud run.js --env SCENARIO=smoke
```

## ⚠️ 注意事項

### 1. CSRF Token 處理

測試已經實現 CSRF Token 自動處理，無需手動設置。

### 2. 測試用戶

測試使用 `data/users.json` 中的用戶：
- testuser1@example.com
- testuser2@example.com
- testuser3@example.com
- testuser4@example.com
- testuser5@example.com

密碼統一為：`Password123!`

### 3. 測試資料清理

測試會創建真實資料（伺服器、頻道等），建議：
- 使用專門的測試環境
- 定期清理測試資料
- 不要在生產環境執行

### 4. WebSocket 協議

後端使用 `action` 欄位而不是 `type`，測試已更新以符合：
```javascript
{
  "action": "join_room",  // 不是 "type"
  "data": { ... }
}
```

## 🐛 故障排除

### 問題：連線被拒絕（ECONNREFUSED）

**解決方案**：
```bash
# 檢查應用程式是否運行
curl http://localhost:8111/health

# 啟動應用程式
docker-compose up -d
# 或
go run main.go
```

### 問題：401 Unauthorized

**解決方案**：
- 檢查 CSRF Token 設置
- 檢查 Access Token 是否過期
- 確認用戶已正確註冊和登入

### 問題：Check 失敗率高

**說明**：這是正常的，因為某些測試使用假 ID 來測試錯誤處理：
- 創建私聊時使用假用戶 ID（預期 403/400）
- 查詢不存在的房間（預期 404）
- 更新不存在的資源（預期 400）

這些是有意為之的測試，用於驗證錯誤處理機制。

### 問題：測試結果檔案未生成

**解決方案**：
```bash
# 確保目錄存在
mkdir -p test_results/load_tests

# 檢查寫入權限
ls -la test_results/
```

## 📝 最近更新

### 2025-10-07
- ✅ 完成 Smoke Test 場景
- ✅ 修復 WebSocket 訊息格式（action vs type）
- ✅ 修正 API 端點路徑
- ✅ 改善日誌輸出
- ✅ 修復測試結果目錄問題
- ✅ 調整性能閾值以適應冒煙測試
- ✅ 創建完整的測試文檔

### 待完成
- ⏳ Light Test 場景優化
- ⏳ Medium Test 場景優化
- ⏳ Heavy Test 場景優化
- ⏳ WebSocket 連線測試整合
- ⏳ 完整的端到端測試流程

## 📚 參考資料

- [k6 官方文檔](https://k6.io/docs/)
- [k6 測試指南](https://k6.io/docs/test-types/)
- [SMOKE_TEST_GUIDE.md](./SMOKE_TEST_GUIDE.md) - 詳細的 Smoke Test 指南
- [../API.md](../API.md) - 後端 API 文檔

## 🤝 貢獻

如果你想改進測試套件：

1. 確保所有測試通過
2. 添加新的測試場景到 `scenarios/` 目錄
3. 更新相關文檔
4. 提交 Pull Request

## 📧 聯絡

如有問題或建議，請聯絡專案維護者。
