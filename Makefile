# Makefile for Chat App Backend
# 用於本地開發環境的指令集
# 注意：此檔案不會部署到生產環境

.PHONY: help dev dev-logs dev-down dev-restart build logs status ps restart stop start
.PHONY: shell mongo-shell redis-cli test test-coverage test-smoke test-limit test-ws test-analyze
.PHONY: clean clean-dev fmt lint tidy run env-check install-deps init

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
	docker-compose -f docker-compose.dev.yml up -d
	@echo "✅ 開發環境已啟動"
	@echo "📍 API: http://localhost:8111"
	@echo "📍 Redis Commander: http://localhost:8081"

dev-logs:
	@echo "🚀 啟動開發環境並顯示日誌..."
	docker-compose -f docker-compose.dev.yml up

dev-down:
	@echo "🛑 停止開發環境..."
	docker-compose -f docker-compose.dev.yml down

dev-restart:
	@echo "🔄 重啟開發環境..."
	docker-compose -f docker-compose.dev.yml restart
	@echo "✅ 開發環境已重啟"

# ============================================
# 建置指令
# ============================================

build:
	@echo "🏗️  建置 Docker 映像..."
	docker-compose -f docker-compose.dev.yml build

rebuild:
	@echo "🏗️  強制重新建置 (無快取)..."
	docker-compose -f docker-compose.dev.yml build --no-cache

# ============================================
# 日誌與監控
# ============================================

logs:
	docker-compose -f docker-compose.dev.yml logs -f

logs-app:
	docker-compose -f docker-compose.dev.yml logs -f app

logs-mongodb:
	docker-compose -f docker-compose.dev.yml logs -f mongodb

logs-redis:
	docker-compose -f docker-compose.dev.yml logs -f redis

status:
	@echo "📊 容器狀態:"
	@docker-compose -f docker-compose.dev.yml ps

ps:
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

stats:
	@echo "📊 實時資源使用統計 (Ctrl+C 退出):"
	@docker stats

health:
	@echo "🏥 檢查應用健康狀態..."
	@curl -s http://localhost:8111/health | jq . || echo "❌ 健康檢查失敗"

# ============================================
# 容器操作
# ============================================

restart:
	@echo "🔄 重啟應用服務..."
	docker-compose -f docker-compose.dev.yml restart app

stop:
	@echo "🛑 停止所有服務..."
	docker-compose -f docker-compose.dev.yml stop

start:
	@echo "▶️  啟動服務..."
	docker-compose -f docker-compose.dev.yml start

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

test-limit:
	@echo "🧪 執行極限測試 (k6)..."
	cd loadtest && npm run test:limit

test-ws:
	@echo "🧪 執行 WebSocket 壓力測試 (k6)..."
	cd loadtest && npm run test:ws:stress

test-analyze:
	@echo "📊 分析最新測試結果..."
	cd loadtest && npm run analyze:limit

# ============================================
# 清理（僅限開發環境）
# ============================================

clean:
	@echo "🧹 清理開發環境容器和卷..."
	@read -p "⚠️  這將刪除所有開發環境資料! 確定要繼續嗎? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		docker-compose -f docker-compose.dev.yml down -v; \
		echo "✅ 開發環境已清理"; \
	else \
		echo "❌ 操作已取消"; \
	fi

clean-dev:
	@echo "🧹 清理開發環境容器和卷..."
	docker-compose -f docker-compose.dev.yml down -v
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
	@if [ ! -f .env ]; then \
		echo "❌ .env 文件不存在"; \
		echo "💡 請複製 .env.example 並配置:"; \
		echo "   cp .env.example .env"; \
	else \
		echo "✅ .env 文件存在"; \
	fi

env-example:
	@echo "📝 生成 .env.example..."
	@echo "# 請參考此範例配置您的 .env 文件" > .env.example
	@echo "SERVER_PORT=8111" >> .env.example
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
