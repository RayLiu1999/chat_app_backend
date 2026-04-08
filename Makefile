# Makefile for Chat App Backend
# 用於本地開發環境的指令集

# 預設擴展實例數
N ?= 3
# 預設測試用戶數（本地測試建議：100~200 功能驗證，500~1000 趨勢觀察）
USER_COUNT ?= 500
# 預設環境
ENV ?= development

# 版本資訊 (將用於建置參數)
VERSION ?= $(shell grep 'Version =' version/version.go | cut -d '"' -f 2 || echo "1.0.0")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# 通用環境變數組合
COMPOSE_ENV = ENV_FILE=.env.development VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) GIT_COMMIT=$(GIT_COMMIT) GIT_BRANCH=$(GIT_BRANCH)
DOCKER_COMPOSE = $(COMPOSE_ENV) docker-compose --env-file .env.development

.PHONY: help dev dev-logs dev-down dev-restart
.PHONY: test-env test-env-build test-env-logs test-env-down
.PHONY: scale scale-build scale-up scale-down scale-logs scale-status
.PHONY: db-shell db-migrate db-seed db-fresh mongo-init
.PHONY: run fmt lint vuln tidy test test-coverage test-smoke test-prepare-users test-capacity test-capacity-prepared
.PHONY: k8s-build k8s-deploy k8s-redeploy k8s-delete k8s-scale k8s-status k8s-logs k8s-pods k8s-health
.PHONY: env-check install-deps init

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
	@echo "🧪 壓測環境 (Stress Test / Production Build):"
	@echo "  make test-env         - 啟動壓測環境"
	@echo "  make test-env-build   - 建置壓測環境容器"
	@echo "  make test-env-logs    - 查看壓測環境日誌"
	@echo "  make test-env-down    - 停止壓測環境"
	@echo "  make test-prepare-users USER_COUNT=500 - 預先建立測試用戶"
	@echo "  make test-smoke       - 執行冒煙測試 (k6)"
	@echo "  make test-capacity    - 執行單體容量測試 (k6)"
	@echo "  make test-capacity-prepared USER_COUNT=500 - 先準備用戶再壓測"
	@echo ""
	@echo "🔄 水平擴展測試 (Horizontal Scaling):"
	@echo "  make scale            - 啟動擴展環境 ($(N) 個實例)"
	@echo "  make scale-build      - 建置擴展環境容器"
	@echo "  make scale-up N=5     - 擴展到 N 個實例"
	@echo "  make scale-down       - 停止擴展環境"
	@echo "  make scale-logs       - 查看擴展環境日誌"
	@echo ""
	@echo "🗄️  資料庫管理 (Database):"
	@echo "  make db-shell         - 進入 MongoDB Shell"
	@echo "  make db-migrate       - 執行資料庫初始化與遷移"
	@echo "  make db-seed          - 寫入預設/測試資料 (seed.js)"
	@echo "  make db-fresh         - 清空資料庫並重新執行遷移與 Seed"
	@echo ""
	@echo "🏗️  Go 開發:"
	@echo "  make run              - 本地執行應用"
	@echo "  make fmt              - 格式化程式碼"
	@echo "  make lint             - 執行核心的代碼風格與潛在 Bug 檢查"
	@echo "  make vuln             - 檢查漏洞 (govulncheck)"
	@echo "  make tidy             - 整理依賴"
	@echo "  make test             - 執行單元測試"
	@echo "  make test-coverage    - 生成測試覆蓋率報告"
	@echo ""

# ============================================
# 開發環境指令
# ============================================

dev:
	@echo "🚀 啟動開發環境..."
	ENV_FILE=.env.development docker-compose -f docker-compose.dev.yml --env-file .env.development up -d
	@echo "✅ 開發環境已啟動"
	@echo "📍 API: http://localhost:80"

dev-logs:
	@echo "🚀 啟動開發環境並顯示日誌..."
	ENV_FILE=.env.development docker-compose -f docker-compose.dev.yml --env-file .env.development up

dev-down:
	@echo "🛑 停止開發環境..."
	ENV_FILE=.env.development docker-compose -f docker-compose.dev.yml --env-file .env.development down

dev-restart:
	@echo "🔄 重啟開發環境..."
	docker-compose -f docker-compose.dev.yml restart
	@echo "✅ 開發環境已重啟"

# ============================================
# 壓測環境指令 (生產環境建構)
# ============================================

test-env:
	@echo "🚀 啟動壓測環境..."
	$(DOCKER_COMPOSE) -f docker-compose.yml up -d
	@echo "✅ 壓測環境已啟動"

test-env-build:
	@echo "🏗️  建置壓測環境容器..."
	$(DOCKER_COMPOSE) -f docker-compose.yml up -d --build
	docker image prune -f
	@echo "✅ 壓測環境容器已建置"

test-env-logs:
	@echo "🚀 啟動壓測環境並顯示日誌..."
	$(DOCKER_COMPOSE) -f docker-compose.yml up

test-env-down:
	@echo "🛑 停止壓測環境..."
	$(DOCKER_COMPOSE) -f docker-compose.yml down

# ============================================
# 水平擴展測試
# ============================================

scale:
	@echo "🔄 啟動水平擴展環境 ($(N) 個實例)..."
	$(DOCKER_COMPOSE) -f docker-compose.scale.yml up -d --scale app=$(N) --no-recreate
	@echo "✅ 擴展環境已啟動"
	@echo "📍 API (via nginx): http://localhost:80"

scale-build:
	@echo "🏗️  建置水平擴展環境容器..."
	$(DOCKER_COMPOSE) -f docker-compose.scale.yml up -d --build --scale app=$(N) --no-recreate
	docker image prune -f
	@echo "✅ 水平擴展環境容器已建置"

scale-up:
	@echo "📈 擴展到 $(N) 個實例..."
	$(DOCKER_COMPOSE) -f docker-compose.scale.yml up -d --scale app=$(N) --no-recreate
	@echo "✅ 已擴展到 $(N) 個實例"

scale-down:
	@echo "🛑 停止擴展環境..."
	$(DOCKER_COMPOSE) -f docker-compose.scale.yml down

scale-logs:
	$(DOCKER_COMPOSE) -f docker-compose.scale.yml logs -f

scale-status:
	$(DOCKER_COMPOSE) -f docker-compose.scale.yml ps

# ============================================
# 資料庫管理指令 (MongoDB)
# ============================================

db-shell:
	@echo "🐚 進入 MongoDB Shell..."
	docker exec --env-file .env.development -it mongodb sh -c 'mongosh -u $$MONGO_USERNAME -p $$MONGO_PASSWORD --authenticationDatabase admin $$MONGO_DB_NAME'

db-migrate:
	@echo "🚀 執行資料庫初始化/遷移 (Migrate)..."
	docker exec --env-file .env.development -i mongodb sh -c 'mongosh -u $$MONGO_USERNAME -p $$MONGO_PASSWORD --authenticationDatabase admin $$MONGO_DB_NAME' < scripts/mongo/mongo-init.js

db-seed:
	@echo "🌱 寫入初始資料 (Seed)..."
	@if [ -f scripts/mongo/seed.js ]; then \
		docker exec --env-file .env.development -i mongodb sh -c 'mongosh -u $$MONGO_USERNAME -p $$MONGO_PASSWORD --authenticationDatabase admin $$MONGO_DB_NAME' < scripts/mongo/seed.js; \
	else \
		echo "❌ 找不到 scripts/mongo/seed.js，若需要請先建立該檔案。"; \
	fi

db-fresh:
	@echo "🔥 清空並重建資料庫 (Fresh)..."
	docker exec --env-file .env.development -i mongodb sh -c 'mongosh -u $$MONGO_USERNAME -p $$MONGO_PASSWORD --authenticationDatabase admin $$MONGO_DB_NAME --eval "db.dropDatabase()"'
	$(MAKE) db-migrate
	$(MAKE) db-seed

# ============================================
# Go 開發指令
# ============================================

run:
	go run main.go

fmt:
	go fmt ./...

lint:
	@echo "🔍 執行代碼風格與潛在 Bug 檢查..."
	golangci-lint run

vuln:
	@echo "🛡️  檢查代碼漏洞..."
	govulncheck ./...

tidy:
	go mod tidy

test:
	@echo "🧪 執行測試..."
	go test ./... -v

test-coverage:
	@echo "🧪 執行測試並生成覆蓋率報告..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

test-smoke:
	@echo "🧪 執行冒煙測試..."
	cd scripts/k6 && ALLOW_RATE_LIMIT_BYPASS=1 k6 run run.js --env SCENARIO=smoke

test-prepare-users:
	@echo "🧪 預先建立測試用戶..."
	cd scripts/k6 && ALLOW_RATE_LIMIT_BYPASS=1 k6 run prepare_users.js --env USER_COUNT=$(USER_COUNT)

test-capacity:
	@echo "🧪 執行容量測試..."
	cd scripts/k6 && ALLOW_RATE_LIMIT_BYPASS=1 k6 run run.js --env SCENARIO=monolith_capacity

test-capacity-prepared:
	@echo "🧪 先準備用戶，再執行容量測試..."
	cd scripts/k6 && ALLOW_RATE_LIMIT_BYPASS=1 k6 run prepare_users.js --env USER_COUNT=$(USER_COUNT)
	cd scripts/k6 && ALLOW_RATE_LIMIT_BYPASS=1 k6 run run.js --env SCENARIO=monolith_capacity --env PREPARE_USERS=1 --env PREPARE_USER_COUNT=$(USER_COUNT)

test-broadcast:
	@echo "🧪 執行廣播測試..."
	cd scripts/k6 && ALLOW_RATE_LIMIT_BYPASS=1 k6 run run.js --env SCENARIO=ws_broadcast --env K6_VUS=100

mongo-init:
	@echo "🗄️  初始化 MongoDB (ENV=$(ENV))..."
	ADMIN_USERNAME="$(ADMIN_USERNAME)" \
	ADMIN_PASSWORD="$(ADMIN_PASSWORD)" \
	bash scripts/mongo/mongo-init.sh $(ENV)

env-check:
	@if [ ! -f .env.development ]; then \
		echo "❌ .env.development 不存在，請從 .env.example 複製"; \
		exit 1; \
	fi

# ============================================
# Kubernetes 本地部署 (OrbStack/Minikube)
# ============================================

k8s-build:
	@echo "🏗️  建置 Docker 映像 (for K8s prod target)..."
	docker build --target prod -t chat_app_backend:latest .
	@echo "✅ 映像建置完成: chat_app_backend:latest"

k8s-deploy: k8s-build
	@echo "☸️  部署到 Kubernetes (local overlay)..."
	kubectl apply -k k8s/overlays/local
	@echo "✅ K8s 部署完成"
	@echo "⏳ 等待 pods 就緒..."
	kubectl -n chat-app wait --for=condition=ready pod -l app=chat-backend --timeout=120s || true
	@make k8s-status

k8s-redeploy: k8s-build
	@echo "🔄 重新部署到 Kubernetes (chat-backend)..."
	kubectl rollout restart deployment/chat-backend -n chat-app
	@echo "✅ 重啟指令已發送"
	@make k8s-status

k8s-delete:
	@echo "🗑️  刪除 K8s 部署..."
	kubectl delete -k k8s/overlays/local --ignore-not-found
	@echo "✅ K8s 部署已刪除"

k8s-scale:
	@echo "📈 擴展到 $(N) 個 pods..."
	kubectl -n chat-app scale deployment chat-backend --replicas=$(N)
	@echo "✅ 已擴展到 $(N) 個 pods"
	kubectl -n chat-app get pods -w

k8s-status:
	@echo "📊 K8s 部署狀態 (chat-backend):"
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
	kubectl -n chat-app logs -f -l app=chat-backend --max-log-requests=10

k8s-pods:
	kubectl -n chat-app get pods -w

k8s-health:
	@echo "🏥 檢查 K8s 服務健康狀態 (via Ingress host)..."
	@curl -s -H "Host: chat.local" http://localhost/health | jq . || echo "❌ 健康檢查失敗 (請確認 Ingress 或 /etc/hosts 是否設定 chat.local)"
