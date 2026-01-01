# Makefile for Chat App Backend
# 簡化 Docker Compose 和常用操作指令

.PHONY: help dev prod test clean build logs status restart stop start shell db-backup db-restore test-limit test-smoke

# 預設顯示幫助訊息
help:
	@echo "==================================================================="
	@echo "  Chat App Backend - Makefile 指令列表"
	@echo "==================================================================="
	@echo ""
	@echo "📦 開發環境 (Development):"
	@echo "  make dev              - 啟動開發環境 (detached mode)"
	@echo "  make dev-logs         - 啟動開發環境並顯示日誌"
	@echo "  make dev-down         - 停止並移除開發環境容器"
	@echo "  make dev-restart      - 重啟開發環境"
	@echo ""
	@echo "🚀 生產環境 (Production):"
	@echo "  make prod             - 啟動生產環境 (detached mode)"
	@echo "  make prod-logs        - 啟動生產環境並顯示日誌"
	@echo "  make prod-down        - 停止並移除生產環境容器"
	@echo "  make prod-restart     - 重啟生產環境"
	@echo ""
	@echo "🔧 通用操作:"
	@echo "  make build            - 重新構建 Docker 映像"
	@echo "  make logs             - 查看當前環境日誌 (dev)"
	@echo "  make status           - 查看容器狀態"
	@echo "  make ps               - 查看運行中的容器"
	@echo "  make restart          - 重啟應用服務 (dev)"
	@echo "  make stop             - 停止所有服務 (dev)"
	@echo "  make start            - 啟動已停止的服務 (dev)"
	@echo "  make clean            - 清理所有容器、映像、卷"
	@echo ""
	@echo "🐚 容器互動:"
	@echo "  make shell            - 進入應用容器 shell (dev)"
	@echo "  make shell-prod       - 進入應用容器 shell (prod)"
	@echo "  make mongo-shell      - 進入 MongoDB shell (dev)"
	@echo "  make redis-cli        - 進入 Redis CLI (dev)"
	@echo ""
	@echo "💾 資料庫操作:"
	@echo "  make db-backup        - 備份 MongoDB 資料"
	@echo "  make db-restore       - 恢復 MongoDB 資料"
	@echo "  make db-clean         - 清空 MongoDB 資料 (危險!)"
	@echo ""
	@echo "🧪 測試:"
	@echo "  make test             - 執行單元測試"
	@echo "  make test-smoke       - 執行冒煙測試 (k6)"
	@echo "  make test-limit       - 執行極限測試 (k6)"
	@echo "  make test-ws          - 執行 WebSocket 壓力測試"
	@echo "  make test-coverage    - 執行測試並生成覆蓋率報告"
	@echo ""
	@echo "🏗️  建置:"
	@echo "  make build-dev        - 建置開發環境映像"
	@echo "  make build-prod       - 建置生產環境映像"
	@echo "  make rebuild          - 強制重新建置 (無快取)"
	@echo ""
	@echo "📊 監控:"
	@echo "  make stats            - 實時顯示資源使用統計"
	@echo "  make health           - 檢查應用健康狀態"
	@echo "  make metrics          - 查看 Prometheus 指標"
	@echo "  make monitoring-up    - 啟動監控整合 (連接到 prometheus-grafana)"
	@echo "  make monitoring-down  - 停止監控整合"
	@echo "  make check-network    - 檢查 prometheus-grafana network"
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
# 生產環境指令
# ============================================

prod:
	@echo "🚀 啟動生產環境..."
	docker-compose -f docker-compose.prod.yml up -d
	@echo "✅ 生產環境已啟動"
	@echo "⚠️  請確保已設置 .env.production 文件"

prod-logs:
	@echo "🚀 啟動生產環境並顯示日誌..."
	docker-compose -f docker-compose.prod.yml up

prod-down:
	@echo "🛑 停止生產環境..."
	docker-compose -f docker-compose.prod.yml down

prod-restart:
	@echo "🔄 重啟生產環境..."
	docker-compose -f docker-compose.prod.yml restart
	@echo "✅ 生產環境已重啟"

# ============================================
# 建置指令
# ============================================

build:
	@echo "🏗️  建置 Docker 映像..."
	docker-compose -f docker-compose.dev.yml build

build-dev:
	@echo "🏗️  建置開發環境映像..."
	docker-compose -f docker-compose.dev.yml build

build-prod:
	@echo "🏗️  建置生產環境映像..."
	docker-compose -f docker-compose.prod.yml build

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

shell-prod:
	@echo "🐚 進入生產應用容器..."
	docker exec -it chat_app_backend_prod sh

mongo-shell:
	@echo "🍃 進入 MongoDB shell..."
	docker exec -it chat_mongodb_dev mongosh -u ${MONGO_INITDB_ROOT_USERNAME} -p ${MONGO_INITDB_ROOT_PASSWORD}

redis-cli:
	@echo "📮 進入 Redis CLI..."
	docker exec -it chat_redis_dev redis-cli -a ${REDIS_PASSWORD}

# ============================================
# 資料庫操作
# ============================================

db-backup:
	@echo "💾 備份 MongoDB 資料..."
	@mkdir -p backups/mongodb
	docker exec chat_mongodb_dev mongodump --username=${MONGO_INITDB_ROOT_USERNAME} --password=${MONGO_INITDB_ROOT_PASSWORD} --authenticationDatabase=admin --out=/backups/mongodb-$(shell date +%Y%m%d_%H%M%S)
	@echo "✅ 備份完成"

db-restore:
	@echo "📥 恢復 MongoDB 資料..."
	@read -p "請輸入備份目錄名稱: " backup_dir; \
	docker exec chat_mongodb_dev mongorestore --username=${MONGO_INITDB_ROOT_USERNAME} --password=${MONGO_INITDB_ROOT_PASSWORD} --authenticationDatabase=admin /backups/$$backup_dir

db-clean:
	@echo "⚠️  警告: 此操作將清空所有 MongoDB 資料!"
	@read -p "確定要繼續嗎? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		docker exec chat_mongodb_dev mongosh -u ${MONGO_INITDB_ROOT_USERNAME} -p ${MONGO_INITDB_ROOT_PASSWORD} --authenticationDatabase admin --eval "db.dropDatabase()"; \
		echo "✅ 資料庫已清空"; \
	else \
		echo "❌ 操作已取消"; \
	fi

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
# 清理
# ============================================

clean:
	@echo "🧹 清理所有容器、映像、卷..."
	@read -p "⚠️  這將刪除所有資料! 確定要繼續嗎? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		docker-compose -f docker-compose.dev.yml down -v; \
		docker-compose -f docker-compose.prod.yml down -v; \
		docker system prune -af --volumes; \
		echo "✅ 清理完成"; \
	else \
		echo "❌ 操作已取消"; \
	fi

clean-dev:
	@echo "🧹 清理開發環境容器和卷..."
	docker-compose -f docker-compose.dev.yml down -v
	@echo "✅ 開發環境已清理"

clean-prod:
	@echo "🧹 清理生產環境容器和卷..."
	docker-compose -f docker-compose.prod.yml down -v
	@echo "✅ 生產環境已清理"

# ============================================
# Go 相關指令
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
# 實用工具
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
	@make build-dev
	@echo "✅ 專案初始化完成"
	@echo "💡 使用 'make dev' 啟動開發環境"

# ============================================
# Prometheus + Grafana 監控整合
# ============================================

check-network:
	@echo "🔍 檢查 prometheus-grafana network..."
	@if docker network inspect prometheus-grafana >/dev/null 2>&1; then \
		echo "✅ prometheus-grafana network 存在"; \
		echo ""; \
		echo "📋 Network 詳細資訊:"; \
		docker network inspect prometheus-grafana --format='{{range .Containers}}  - {{.Name}} ({{.IPv4Address}}){{println}}{{end}}'; \
	else \
		echo "❌ prometheus-grafana network 不存在"; \
		echo ""; \
		echo "💡 請先確認您的 prometheus-grafana 容器 network 名稱:"; \
		echo "   docker inspect prometheus-grafana | grep NetworkMode"; \
		echo "   docker network ls | grep prometheus"; \
		echo ""; \
		echo "然後更新 docker-compose.monitoring.yml 中的 network 名稱"; \
	fi

monitoring-up:
	@echo "🚀 啟動監控整合..."
	@make check-network
	@echo ""
	@echo "📊 啟動 Promtail 和連接到監控 network..."
	docker-compose -f docker-compose.prod.yml -f docker-compose.monitoring.yml up -d
	@echo "✅ 監控整合已啟動"
	@echo ""
	@echo "📍 Metrics endpoint: http://localhost:8111/metrics"
	@echo "📍 Prometheus: http://your-prometheus:9090"
	@echo "📍 Grafana: http://your-grafana:3000"

monitoring-down:
	@echo "🛑 停止監控整合..."
	docker-compose -f docker-compose.prod.yml -f docker-compose.monitoring.yml down
	@echo "✅ 監控整合已停止"

metrics:
	@echo "📊 查看 Prometheus 指標..."
	@echo ""
	@if curl -f http://localhost:8111/metrics 2>/dev/null; then \
		echo ""; \
		echo "✅ Metrics endpoint 正常"; \
	else \
		echo "❌ 無法訪問 metrics endpoint"; \
		echo "💡 請確認應用是否正在運行: make status"; \
	fi

prometheus-config:
	@echo "📝 Prometheus 配置範例..."
	@echo ""
	@echo "# 在您的 Prometheus 配置中添加:"
	@echo "scrape_configs:"
	@echo "  - job_name: 'chat_app_backend'"
	@echo "    static_configs:"
	@echo "      - targets: ['chat_app_backend_prod:8111']"
	@echo "    metrics_path: '/metrics'"
	@echo "    scrape_interval: 10s"
	@echo ""
	@echo "# 然後重新載入 Prometheus:"
	@echo "curl -X POST http://localhost:9090/-/reload"
