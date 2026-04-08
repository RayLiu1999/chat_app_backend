# K6 性能測試指南 (Load Testing Guide)

這份文件說明了 `scripts/k6/scenarios/` 目錄下的三份核心 K6 測試腳本的用途，以及如何在本地單體 (Monolith) 與水平擴展 (Horizontal Scaling) 架構下搭配 Grafana 進行效能驗證。

---

## 📂 測試腳本詳解

我們提供了三種不同維度的測試腳本，涵蓋了從基本功能驗證到極限效能壓測：

### 1. 冒煙測試 (`smoke.js`)
* **主要用途**：**系統可用性與全流程驗證**。
* **腳本行為**：模擬少量真實用戶的完整使用路徑。包含註冊、登入、取得好友列表、建立頻道、發送私訊，以及建立 WebSocket 連線並收發訊息。
* **何時使用**：
    * 每次修改核心程式碼後（如更動 Controller 或 Middleware）。
    * 部署完新架構後，確認負載均衡 (Nginx / Kubernetes Ingress) 是否有正確轉發所有 API。
* **指標觀察**：不看 RPS (每秒請求數)，只看 **HTTP 錯誤率 (Failure Rate)** 是否為 0%，以及流程是否順暢。

### 2. 容量與並發測試 (`monolith_capacity.js`)
* **主要用途**：**尋找系統的效能瓶頸與極限承載量**。
* **腳本行為**：直接大量打擊核心 API（略過冗長的註冊流程，使用預備好的 Token）。產生高頻率的 HTTP 請求與短暫的 WebSocket 建立請求。
* **何時使用**：
    * 評估當前分配的 CPU/Memory 資源最多能承受多少 Virtual Users (VU)。
    * 對比單機架構與水平擴展架構下的吞吐量差異。
* **指標觀察**：觀察 **RPS (Requests Per Second)**、**P95 延遲 (Latency)** 以及何時開始出現 `502 Bad Gateway` 或 `504 Gateway Timeout`。

### 3. WebSocket 廣播風暴測試 (`ws_broadcast.js`)
* **主要用途**：**測試即時通訊的長連線負載與訊息分發能力**。
* **腳本行為**：讓上百個或數千個 VU 同時連接到同一個 Server/Channel（模擬萬人聊天室），由少數發送者狂發訊息，其餘人作為接收者。
* **何時使用**：
    * 驗證 Redis Pub/Sub 在多實例架構下的訊息廣播是否會遺漏。
    * 測試 Nginx / Traefik Ingress 的 WebSocket 連線保持能力。
* **指標觀察**：觀察 **WebSocket P95 延遲**、**連線斷開率**，以及「發送/接收倍率」（驗證一則訊息是否有成功廣播給所有人）。

---

## 🏗️ 測試策略：單體架構 vs 水平擴展

這套腳本非常適合用來觀察「加機器」對效能的實際影響：

### 階段一：單體架構 (Monolith) 基準測試
1. 本地啟動單一 Backend 容器（或設定 K8s replicas=1）。
2. 使用 `monolith_capacity.js`，從低 VU 慢慢爬升。
3. **發現瓶頸**：您可能會發現當 VU 達到某個數字時，CPU 達到 100%，或者 API Latency 突然飆高超過 1000ms。記錄下這個閥值。

### 階段二：水平擴展 (Horizontal Scaling) 驗證
1. 擴展 Backend 實例（例如 `make k8s-scale N=3` 或 Docker Compose Nginx 擴展）。
2. 套用完全相同的 `monolith_capacity.js` 或 `ws_broadcast.js` 腳本。
3. **預期結果**：前方的 Ingress / Nginx 應該要平均分攤流量。您會發現即使達到之前的瓶頸 VU 數，系統依然穩定，且 RPS 有顯著提升。
4. **注意事項**：在多實例下跑 `ws_broadcast.js`，若 Redis 配置錯誤，會發生「A 容器的用戶發訊息，B 容器的用戶收不到」的狀況，此腳本能精準抓出這種架構缺陷。

---

## 📊 搭配 Grafana 監控指標

在執行測試指令的同時，請開啟 Grafana 面板，對照以下關鍵指標：

1. **CPU / Memory Usage (資源消耗)**
    * **看什麼**：Backend 容器的資源是否耗盡？是否發生 OOM (Out of Memory) 被強制重啟？
    * **意義**：如果 CPU 未滿但 API 卻卡住，可能瓶頸在資料庫連線池 (Connection Pool) 或 Go Runtime 的 Goroutines 鎖死。
2. **Redis Pub/Sub 流量**
    * **看什麼**：在跑 `ws_broadcast.js` 時，Redis 的 Network I/O 與 CPU 使用率。
    * **意義**：多實例聊天室高度依賴 Redis 發送廣播。如果 Redis 成為瓶頸，即使開再多 Backend 容器也沒有用。
3. **MongoDB Connection / Query Latency**
    * **看什麼**：資料庫活躍連線數是否暴增？
    * **意義**：確認是否有慢查詢拖垮整個系統，或者 Connection 不足導致 API 等待 (Pending)。
4. **Nginx / Traefik 轉發狀態**
    * **看什麼**：不同 Backend 節點的請求分布圖。
    * **意義**：確認流量的 **負載均衡是否均勻**。如果發現所有流量都只打在同一個 Pod 上，代表 Ingress 的 Sticky Session 或 Load Balancing 設定有誤。

### 如何執行測試

回到根目錄使用 Makefile 指令即可觸發這些測試：
```bash
# 執行冒煙測試
make test-smoke

# 執行容量與並發測試
make test-capacity

# 執行 WebSocket 廣播測試
# (需修改 Makefile 或直接執行 k6 命令)
k6 run --env SCENARIO=ws_broadcast scripts/k6/run.js
```
