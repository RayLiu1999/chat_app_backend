# Makefile for Chat App Backend
# 用於本地開發環境的指令集

.PHONY: help dev dev-logs dev-down dev-restart build logs status ps restart stop start
.PHONY: shell test test-smoke test-limit test-ws test-analyze
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
	@echo "🧪 壓測環境 (Stress Test):"
	@echo "  make test-env         - 啟動壓測環境 (prod build)"
	@echo "  make test-env-down    - 停止壓測環境"
	@echo "  make test-smoke       - 執行冒煙測試 (k6)"
	@echo "  make test-capacity    - 執行單體容量測試 (k6)"
	@echo ""
	@echo "🔄 水平擴展測試 (Horizontal Scaling):"
	@echo "  make scale            - 啟動 3 個實例 (nginx + 3x app)"
	@echo "  make scale-up N=5     - 擴展到 N 個實例"
	@echo "  make scale-down       - 停止擴展環境"
	@echo ""
	@echo "🔧 通用操作:"
	@echo "  make build            - 建置 Docker 映像"
	@echo "  make logs             - 查看日誌"
	@echo "  make ps               - 查看運行中的容器"
	@echo "  make stats            - 實時資源使用統計"
	@echo "  make restart          - 重啟應用服務"
	@echo "  make stop             - 停止所有服務"
	@echo "  make clean            - 清理環境"
	@echo ""
	@echo "🐚 容器互動:"
	@echo "  make shell            - 進入應用容器 shell"
	@echo ""
	@echo "🏗️  Go 開發:"
	@echo "  make run              - 本地執行應用"
	@echo "  make fmt              - 格式化程式碼"
	@echo "  make lint             - 程式碼檢查"
	@echo "  make tidy             - 整理依賴"
	@echo ""

# ============================================
# 開發環境指令
# ============================================

dev:
	@echo "🚀 啟動開發環境..."
	docker-compose up -d
	@echo "✅ 開發環境已啟動"
	@echo "📍 API: http://localhost:80"

dev-logs:
	@echo "🚀 啟動開發環境並顯示日誌..."
	docker-compose up

dev-down:
	@echo "🛑 停止開發環境..."
	docker-compose down

dev-restart:
	@echo "🔄 重啟開發環境..."
	docker-compose restart
	@echo "✅ 開發環境已重啟"

# ============================================
# 壓測環境指令
# ============================================

test-env:
	@echo "🚀 啟動壓測環境 (Prod Image)..."
	DOCKER_TARGET=prod CPU_LIMIT=4 MEMORY_LIMIT=2G GOMEMLIMIT=1800MiB docker-compose --profile test up -d
	@echo "✅ 壓測環境已啟動"

test-env-down:
	@echo "🛑 停止壓測環境..."
	docker-compose --profile test down

# ============================================
# 建置與日誌
# ============================================

build:
	@echo "🏗️  建置 Docker 映像..."
	docker-compose build

rebuild:
	@echo "🏗️  強制重新建置 (無快取)..."
	docker-compose build --no-cache


logs:
	docker-compose logs -f

ps:
	docker-compose ps

stats:
	docker stats

stop:
	docker-compose stop

start:
	docker-compose start

restart:
	docker-compose restart

clean:
	@echo "🧹 清理環境..."
	docker-compose down -v --remove-orphans
	docker-compose -f docker-compose.scale.yml down -v --remove-orphans
	@echo "✅ 環境已清理"

# ============================================
# 容器互動
# ============================================

shell:
	@echo "🐚 進入應用容器..."
	docker exec -it chat_app_backend sh

# ============================================
# 水平擴展測試
# ============================================

scale:
	@echo "🔄 啟動水平擴展環境 ($(N) 個實例)..."
	docker-compose -f docker-compose.scale.yml up -d --scale app=$(N)
	@echo "✅ 擴展環境已啟動"
	@echo "📍 API (via nginx): http://localhost:80"

scale-up:
	@echo "📈 擴展到 $(N) 個實例..."
	docker-compose -f docker-compose.scale.yml up -d --scale app=$(N) --no-recreate
	@echo "✅ 已擴展到 $(N) 個實例"

scale-down:
	@echo "🛑 停止擴展環境..."
	docker-compose -f docker-compose.scale.yml down

scale-logs:
	docker-compose -f docker-compose.scale.yml logs -f

scale-status:
	docker-compose -f docker-compose.scale.yml ps

# ============================================
# Go 開發指令
# ============================================

run:
	go run main.go

fmt:
	go fmt ./...

lint:
	golangci-lint run

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
	cd loadtest && k6 run run.js --env SCENARIO=smoke

test-capacity:
	@echo "🧪 執行容量測試..."
	cd loadtest && k6 run run.js --env SCENARIO=monolith_capacity

env-check:
	@if [ ! -f .env.development ]; then \
		echo "❌ .env.development 不存在，請從 .env.example 複製"; \
		exit 1; \
	fi

# ============================================
# Kubernetes 本地部署 (OrbStack/Minikube)
# ============================================

k8s-build:
	@echo "🏗️  建置 Docker 映像 (for K8s)..."
	docker build --target prod -t chat_app_backend:latest .
	@echo "✅ 映像建置完成: chat_app_backend:latest"

k8s-deploy: k8s-build
	@echo "☸️  部署到 Kubernetes (local overlay)..."
	kubectl apply -k k8s/overlays/local
	@echo "✅ K8s 部署完成"
	@echo "⏳ 等待 pods 就緒..."
	kubectl -n chat-app wait --for=condition=ready pod -l app=chat-app --timeout=120s || true
	@make k8s-status

k8s-redeploy: k8s-build
	@echo "🔄 重新部署到 Kubernetes..."
	kubectl rollout restart deployment/chat-app -n chat-app
	@echo "✅ 重啟指令已發送"
	@make k8s-status

k8s-delete:
	@echo "🗑️  刪除 K8s 部署..."
	kubectl delete -k k8s/overlays/local --ignore-not-found
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

k8s-health:
	@echo "🏥 檢查 K8s 服務健康狀態..."
	@curl -s http://localhost/health | jq . || echo "❌ 健康檢查失敗 (請確認 Ingress 或 Port Forwarding)"
