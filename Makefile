# Makefile for Chat App Backend
# 用於本地開發環境的指令集
# 注意：此檔案不會部署到生產環境

.PHONY: help dev dev-logs dev-down dev-restart build logs status ps restart stop start
.PHONY: shell mongo-shell redis-cli test test-coverage test-smoke test-limit test-ws test-analyze
.PHONY: clean clean-dev fmt lint tidy run env-check install-deps init
.PHONY: scale scale-up scale-down scale-logs scale-status scale-build
.PHONY: k8s-deploy k8s-delete k8s-scale k8s-status k8s-logs k8s-pods

# 預設顯示幫助訊息
help:
	@echo "==================================================================="
	@echo "  Chat App Backend - 本地開發環境 Makefile"
	@echo "==================================================================="
	@echo ""
	@echo "📦 開發環境 (Development):"
	@echo "  make dev              - 啟動開發環境 (detached mode)"
	@echo "  make dev-logs         - 啟動開發環境並顯示日誌"
	@echo "  make dev-down         - 停止並移除開發環境容器"
	@echo "  make dev-restart      - 重啟開發環境"
	@echo ""
	@echo "🔄 水平擴展測試 (Horizontal Scaling):"
	@echo "  make scale            - 啟動 3 個實例 (nginx + 3x app)"
	@echo "  make scale-up N=5     - 擴展到 N 個實例"
	@echo "  make scale-down       - 停止擴展環境"
	@echo "  make scale-logs       - 查看擴展環境日誌"
	@echo "  make scale-status     - 查看實例狀態"
	@echo ""
	@echo "☸️  Kubernetes (OrbStack):"
	@echo "  make k8s-deploy       - 部署到本地 K8s"
	@echo "  make k8s-scale N=5    - 擴展到 N 個 pods"
	@echo "  make k8s-status       - 查看部署狀態"
	@echo "  make k8s-logs         - 查看 pods 日誌"
	@echo "  make k8s-delete       - 刪除 K8s 部署"
	@echo ""
	@echo "🔧 通用操作:"
	@echo "  make build            - 建置 Docker 映像"
	@echo "  make rebuild          - 強制重新建置 (無快取)"
	@echo "  make logs             - 查看當前環境日誌"
	@echo "  make logs-app         - 查看應用服務日誌"
	@echo "  make status           - 查看容器狀態"
	@echo "  make ps               - 查看運行中的容器"
	@echo "  make stats            - 實時資源使用統計"
	@echo "  make health           - 檢查應用健康狀態"
	@echo "  make restart          - 重啟應用服務"
	@echo "  make stop             - 停止所有服務"
	@echo "  make start            - 啟動已停止的服務"
	@echo "  make clean            - 清理開發環境容器和卷"
	@echo ""
	@echo "🐚 容器互動:"
	@echo "  make shell            - 進入應用容器 shell"
	@echo "  make mongo-shell      - 進入 MongoDB shell"
	@echo "  make redis-cli        - 進入 Redis CLI"
	@echo ""
	@echo "🧪 測試:"
	@echo "  make test             - 執行單元測試"
	@echo "  make test-smoke       - 執行冒煙測試 (k6)"
	@echo "  make test-limit       - 執行極限測試 (k6)"
	@echo "  make test-ws          - 執行 WebSocket 壓力測試"
	@echo "  make test-coverage    - 執行測試並生成覆蓋率報告"
	@echo "  make test-analyze     - 分析最新測試結果"
	@echo ""
	@echo "🏗️  Go 開發:"
	@echo "  make run              - 本地執行應用"
	@echo "  make fmt              - 格式化程式碼"
	@echo "  make lint             - 程式碼檢查"
	@echo "  make tidy             - 整理依賴"
	@echo ""
	@echo "🛠️  環境設置:"
	@echo "  make env-check        - 檢查環境變數"
	@echo "  make env-example      - 生成 .env.example"
	@echo "  make install-deps     - 安裝依賴"
	@echo "  make init             - 初始化專案"
	@echo ""
	@echo "==================================================================="

# ============================================
# 開發環境指令
# ============================================

dev:
	@echo "🚀 啟動開發環境..."
	docker-compose -f docker-compose.dev.yml --env-file .env.development up -d
	@echo "✅ 開發環境已啟動"
	@echo "📍 API: http://localhost:80"

dev-logs:
	@echo "🚀 啟動開發環境並顯示日誌..."
	docker-compose -f docker-compose.dev.yml --env-file .env.development up

dev-down:
	@echo "🛑 停止開發環境..."
	docker-compose -f docker-compose.dev.yml --env-file .env.development down

dev-restart:
	@echo "🔄 重啟開發環境..."
	docker-compose -f docker-compose.dev.yml --env-file .env.development restart
	@echo "✅ 開發環境已重啟"

# ============================================
# 建置指令
# ============================================

build:
	@echo "🏗️  建置 Docker 映像..."
	docker-compose -f docker-compose.dev.yml --env-file .env.development build

rebuild:
	@echo "🏗️  強制重新建置 (無快取)..."
	docker-compose -f docker-compose.dev.yml --env-file .env.development build --no-cache

# ============================================
# 日誌與監控
# ============================================

logs:
	docker-compose -f docker-compose.dev.yml --env-file .env.development logs -f

logs-app:
	docker-compose -f docker-compose.dev.yml --env-file .env.development logs -f app

logs-mongodb:
	docker-compose -f docker-compose.dev.yml --env-file .env.development logs -f mongodb

logs-redis:
	docker-compose -f docker-compose.dev.yml --env-file .env.development logs -f redis

status:
	@echo "📊 容器狀態:"
	@docker-compose -f docker-compose.dev.yml --env-file .env.development ps

ps:
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

stats:
	@echo "📊 實時資源使用統計 (Ctrl+C 退出):"
	@docker stats

health:
	@echo "🏥 檢查應用健康狀態..."
	@curl -s http://localhost:80/health | jq . || echo "❌ 健康檢查失敗"

# ============================================
# 容器操作
# ============================================

restart:
	@echo "🔄 重啟應用服務..."
	docker-compose -f docker-compose.dev.yml --env-file .env.development restart app

stop:
	@echo "🛑 停止所有服務..."
	docker-compose -f docker-compose.dev.yml --env-file .env.development stop

start:
	@echo "▶️  啟動服務..."
	docker-compose -f docker-compose.dev.yml --env-file .env.development start

# ============================================
# 容器 Shell
# ============================================

shell:
	@echo "🐚 進入應用容器..."
	docker exec -it chat_app_backend_dev sh

mongo-shell:
	@echo "🍃 進入 MongoDB shell..."
	docker exec -it chat_mongodb_dev mongosh -u ${MONGO_INITDB_ROOT_USERNAME} -p ${MONGO_INITDB_ROOT_PASSWORD}

redis-cli:
	@echo "📮 進入 Redis CLI..."
	docker exec -it chat_redis_dev redis-cli -a ${REDIS_PASSWORD}

# ============================================
# 測試
# ============================================

test:
	@echo "🧪 執行單元測試..."
	go test ./... -v

test-coverage:
	@echo "🧪 執行測試並生成覆蓋率報告..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆蓋率報告已生成: coverage.html"

test-smoke:
	@echo "🧪 執行冒煙測試 (k6)..."
	cd loadtest && npm run test:smoke

test-light:
	@echo "🧪 執行輕量級測試 (k6)..."
	cd loadtest && npm run test:light

test-medium:
	@echo "🧪 執行中量級測試 (k6)..."
	cd loadtest && npm run test:medium

test-heavy:
	@echo "🧪 執行極限測試 (k6)..."
	cd loadtest && npm run test:heavy

test-ws-stress-mixed:
	@echo "🧪 執行 WebSocket 壓力測試 (k6)..."
	cd loadtest && npm run test:ws:stress-mixed

test-ws-stress-connections:
	@echo "🧪 執行 WebSocket 連線測試 (k6)..."
	cd loadtest && npm run test:ws:stress-connections

test-ws-stress-messaging:
	@echo "🧪 執行 WebSocket 消息測試 (k6)..."
	cd loadtest && npm run test:ws:stress-messaging

test-ws:spike:
	@echo "🧪 執行 WebSocket 壓力測試 (k6)..."
	cd loadtest && npm run test:ws:spike

test-ws:soak:
	@echo "🧪 執行 WebSocket 浸泡測試 (k6)..."
	cd loadtest && npm run test:ws:soak

test-ws:soak:long:
	@echo "🧪 執行 WebSocket 浸泡測試 (k6)..."
	cd loadtest && npm run test:ws:soak:long

test-ws:ladder-mixed:
	@echo "🧪 執行 WebSocket 梯度測試 (k6)..."
	cd loadtest && npm run test:ws:ladder-mixed

test-ws:ladder-connections:
	@echo "🧪 執行 WebSocket 梯度測試 (k6)..."
	cd loadtest && npm run test:ws:ladder-connections

test-ws:ladder-messaging:
	@echo "🧪 執行 WebSocket 梯度測試 (k6)..."
	cd loadtest && npm run test:ws:ladder-messaging

test-ws:reconnect:
	@echo "🧪 執行 WebSocket 重連測試 (k6)..."
	cd loadtest && npm run test:ws:reconnect

test-ws:reconnect:storm:
	@echo "🧪 執行 WebSocket 重連風暴測試 (k6)..."
	cd loadtest && npm run test:ws:reconnect:storm

test-ws:reconnect:frequent:
	@echo "🧪 執行 WebSocket 頻繁重連測試 (k6)..."
	cd loadtest && npm run test:ws:reconnect:frequent

test-all:basic:
	@echo "🧪 執行 WebSocket 基本測試 (k6)..."
	cd loadtest && npm run test:all:basic

test-all:websocket:
	@echo "🧪 執行 WebSocket WebSocket 測試 (k6)..."
	cd loadtest && npm run test:all:websocket

test-quick:
	@echo "🧪 執行 WebSocket 快速測試 (k6)..."
	cd loadtest && npm run test:quick

# ============================================
# 清理（僅限開發環境）
# ============================================

clean:
	@echo "🧹 清理開發環境容器和卷..."
	@read -p "⚠️  這將刪除所有開發環境資料! 確定要繼續嗎? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		docker-compose -f docker-compose.dev.yml --env-file .env.development down -v; \
		echo "✅ 開發環境已清理"; \
	else \
		echo "❌ 操作已取消"; \
	fi

clean-dev:
	@echo "🧹 清理開發環境容器和卷..."
	docker-compose -f docker-compose.dev.yml --env-file .env.development down -v
	@echo "✅ 開發環境已清理"

# ============================================
# Go 開發指令
# ============================================

run:
	@echo "🏃 本地執行應用..."
	go run main.go

fmt:
	@echo "🎨 格式化程式碼..."
	go fmt ./...

lint:
	@echo "🔍 程式碼檢查..."
	golangci-lint run

tidy:
	@echo "📦 整理依賴..."
	go mod tidy

# ============================================
# 環境設置與初始化
# ============================================

env-check:
	@echo "🔍 檢查環境變數..."
	@if [ ! -f .env.development ]; then \
		echo "❌ .env.development 文件不存在"; \
		echo "💡 請複製 .env.example 並配置:"; \
		echo "   cp .env.example .env.development"; \
	else \
		echo "✅ .env.development 文件存在"; \
	fi

env-example:
	@echo "📝 生成 .env.example..."
	@echo "# 請參考此範例配置您的 .env.development 文件" > .env.example
	@echo "SERVER_PORT=80" >> .env.example
	@echo "✅ .env.example 已生成"

install-deps:
	@echo "📦 安裝 Go 依賴..."
	go mod download
	@echo "📦 安裝 k6 測試依賴..."
	cd loadtest && npm install
	@echo "✅ 依賴安裝完成"

init:
	@echo "🎬 初始化專案..."
	@make env-check
	@make install-deps
	@make build
	@echo "✅ 專案初始化完成"
	@echo "💡 使用 'make dev' 啟動開發環境"

# ============================================
# 水平擴展測試 (Docker Compose)
# ============================================

# 預設實例數量
N ?= 3

scale:
	@echo "🔄 啟動水平擴展環境 ($(N) 個實例)..."
	docker-compose -f docker-compose.scale.yml --env-file .env.development up -d --scale app=$(N)
	@echo "✅ 擴展環境已啟動"
	@echo "📍 API (via nginx): http://localhost:80"
	@echo "📊 查看實例狀態: make scale-status"

scale-build:
	@echo "🏗️  建置擴展環境映像..."
	docker-compose -f docker-compose.scale.yml --env-file .env.development build

scale-up:
	@echo "📈 擴展到 $(N) 個實例..."
	docker-compose -f docker-compose.scale.yml --env-file .env.development up -d --scale app=$(N) --no-recreate
	@echo "✅ 已擴展到 $(N) 個實例"

scale-down:
	@echo "🛑 停止擴展環境..."
	docker-compose -f docker-compose.scale.yml --env-file .env.development down
	@echo "✅ 擴展環境已停止"

scale-logs:
	docker-compose -f docker-compose.scale.yml --env-file .env.development logs -f

scale-status:
	@echo "📊 擴展環境狀態:"
	@docker-compose -f docker-compose.scale.yml --env-file .env.development ps
	@echo ""
	@echo "🔍 測試負載均衡 (訪問 10 次):"
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		echo -n "請求 $$i: "; \
		curl -s http://localhost:80/health 2>/dev/null | head -1 || echo "連線失敗"; \
	done

# ============================================
# Kubernetes 本地部署 (OrbStack)
# ============================================

k8s-build:
	@echo "🏗️  建置 Docker 映像 (for K8s)..."
	docker build -t chat_app_backend:latest -f Dockerfile.k8s .
	@echo "✅ 映像建置完成: chat_app_backend:latest"

k8s-deploy: k8s-build
	@echo "☸️  部署到 Kubernetes..."
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/secret.yaml
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/mongodb.yaml
	kubectl apply -f k8s/redis.yaml
	kubectl apply -f k8s/app.yaml
	kubectl apply -f k8s/service.yaml
	kubectl apply -f k8s/ingress.yaml
	kubectl apply -f k8s/hpa.yaml
	@echo "✅ K8s 部署完成"
	@echo "⏳ 等待 pods 就緒..."
	kubectl -n chat-app wait --for=condition=ready pod -l app=chat-app --timeout=120s || true
	@make k8s-status

k8s-delete:
	@echo "🗑️  刪除 K8s 部署..."
	kubectl delete -f k8s/ --ignore-not-found
	@echo "✅ K8s 部署已刪除"

k8s-scale:
	@echo "📈 擴展到 $(N) 個 pods..."
	kubectl -n chat-app scale deployment chat-app --replicas=$(N)
	@echo "✅ 已擴展到 $(N) 個 pods"
	kubectl -n chat-app get pods -w

k8s-status:
	@echo "📊 K8s 部署狀態:"
	@echo ""
	@echo "=== Pods ==="
	@kubectl -n chat-app get pods -o wide 2>/dev/null || echo "Namespace 不存在"
	@echo ""
	@echo "=== Services ==="
	@kubectl -n chat-app get svc 2>/dev/null || true
	@echo ""
	@echo "=== HPA ==="
	@kubectl -n chat-app get hpa 2>/dev/null || true
	@echo ""
	@echo "=== Ingress ==="
	@kubectl -n chat-app get ingress 2>/dev/null || true

k8s-logs:
	kubectl -n chat-app logs -f -l app=chat-app --max-log-requests=10

k8s-pods:
	kubectl -n chat-app get pods -w
