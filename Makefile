# Chat App Backend - Makefile
# 快速開發與測試工具集

.PHONY: help

# 預設目標：顯示幫助資訊
help:
	@echo "Chat App Backend - 可用指令："
	@echo ""
	@echo "🚀 開發相關："
	@echo "  make run              - 在本地運行應用程式"
	@echo "  make dev              - 使用 air 熱重載運行（開發模式）"
	@echo "  make build            - 編譯應用程式"
	@echo "  make clean            - 清理編譯產物"
	@echo ""
	@echo "🧪 測試相關："
	@echo "  make test             - 執行所有測試"
	@echo "  make test-verbose     - 執行測試（詳細輸出）"
	@echo "  make test-coverage    - 執行測試並產生覆蓋率報告"
	@echo "  make test-watch       - 監控檔案變更並自動測試"
	@echo "  make test-unit        - 只執行單元測試"
	@echo "  make test-service     - 只測試 services 層"
	@echo "  make test-controller  - 只測試 controllers 層"
	@echo "  make test-middleware  - 只測試 middlewares 層"
	@echo "  make test-utils       - 只測試 utils 層"
	@echo ""
	@echo "📊 覆蓋率相關："
	@echo "  make coverage         - 查看覆蓋率摘要"
	@echo "  make coverage-html    - 開啟 HTML 覆蓋率報告"
	@echo "  make coverage-func    - 顯示函數級覆蓋率"
	@echo ""
	@echo "🐳 Docker 相關："
	@echo "  make docker-build     - 建置 Docker 映像"
	@echo "  make docker-up        - 啟動所有服務"
	@echo "  make docker-down      - 停止所有服務"
	@echo "  make docker-restart   - 重啟所有服務"
	@echo "  make docker-logs      - 查看應用日誌"
	@echo "  make docker-clean     - 清理 Docker 資源"
	@echo ""
	@echo "🛠️ 工具相關："
	@echo "  make fmt              - 格式化程式碼"
	@echo "  make lint             - 執行 linter 檢查"
	@echo "  make vet              - 執行 go vet 檢查"
	@echo "  make mod-tidy         - 整理依賴"
	@echo "  make mod-download     - 下載依賴"
	@echo ""
	@echo "🗄️ 資料庫相關："
	@echo "  make db-up            - 啟動資料庫服務"
	@echo "  make db-down          - 停止資料庫服務"
	@echo "  make db-logs          - 查看資料庫日誌"
	@echo ""

# ==================== 開發相關 ====================

# 編譯應用程式
build:
	@echo "📦 編譯應用程式..."
	go build -o bin/chat_app_backend main.go
	@echo "✅ 編譯完成：bin/chat_app_backend"

# 運行應用程式
run: build
	@echo "🚀 啟動應用程式..."
	./bin/chat_app_backend

# 開發模式（使用 air 熱重載）
dev:
	@echo "🔥 啟動開發模式（熱重載）..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "❌ air 未安裝，請執行：go install github.com/air-verse/air@latest"; \
		exit 1; \
	fi

# 清理編譯產物
clean:
	@echo "🧹 清理編譯產物..."
	rm -rf bin/
	rm -rf tmp/
	rm -f coverage*.out coverage*.html
	find . -name "*.test" -type f -delete
	@echo "✅ 清理完成"

# ==================== 測試相關 ====================

# 執行所有測試
test:
	@echo "🧪 執行所有測試..."
	go test -v -race ./...

# 詳細測試輸出
test-verbose:
	@echo "🔍 執行詳細測試..."
	go test -v -race -count=1 ./...

# 測試並產生覆蓋率報告
test-coverage:
	@echo "📊 執行測試並產生覆蓋率報告..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "=== 總覆蓋率 ==="
	@go tool cover -func=coverage.out | grep total
	@echo ""
	@echo "=== 各層級覆蓋率 ==="
	@echo "Controller 層:"
	@go tool cover -func=coverage.out | grep "app/http/controllers" | tail -1 || echo "  無資料"
	@echo "Service 層:"
	@go tool cover -func=coverage.out | grep "app/services" | tail -1 || echo "  無資料"
	@echo "Middleware 層:"
	@go tool cover -func=coverage.out | grep "app/http/middlewares" | tail -1 || echo "  無資料"
	@echo "Utils 層:"
	@go tool cover -func=coverage.out | grep "utils" | tail -1 || echo "  無資料"
	@echo ""
	@echo "產生 HTML 報告..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆蓋率報告已產生：coverage.html"

# 監控檔案變更並自動測試
test-watch:
	@echo "👀 監控測試檔案變更..."
	@if command -v watchexec > /dev/null; then \
		watchexec -e go -c -r "make test"; \
	else \
		echo "❌ watchexec 未安裝"; \
		echo "macOS: brew install watchexec"; \
		echo "Linux: cargo install watchexec-cli"; \
		exit 1; \
	fi

# 只執行單元測試（排除整合測試）
test-unit:
	@echo "🧪 執行單元測試..."
	go test -v -race -short ./...

# 測試 services 層
test-service:
	@echo "🔧 測試 services 層..."
	go test -v -race -coverprofile=coverage_services.out ./app/services
	@go tool cover -func=coverage_services.out | grep total

# 測試 controllers 層
test-controller:
	@echo "🎮 測試 controllers 層..."
	go test -v -race -coverprofile=coverage_controllers.out ./app/http/controllers
	@go tool cover -func=coverage_controllers.out | grep total

# 測試 middlewares 層
test-middleware:
	@echo "🛡️ 測試 middlewares 層..."
	go test -v -race -coverprofile=coverage_middlewares.out ./app/http/middlewares
	@go tool cover -func=coverage_middlewares.out | grep total

# 測試 utils 層
test-utils:
	@echo "🔨 測試 utils 層..."
	go test -v -race -coverprofile=coverage_utils.out ./utils
	@go tool cover -func=coverage_utils.out | grep total

# ==================== 覆蓋率相關 ====================

# 查看覆蓋率摘要
coverage:
	@if [ -f coverage.out ]; then \
		echo "📊 覆蓋率摘要："; \
		go tool cover -func=coverage.out | grep total; \
	else \
		echo "❌ 找不到 coverage.out，請先執行 make test-coverage"; \
		exit 1; \
	fi

# 開啟 HTML 覆蓋率報告
coverage-html:
	@if [ -f coverage.html ]; then \
		echo "🌐 開啟 HTML 覆蓋率報告..."; \
		open coverage.html || xdg-open coverage.html 2>/dev/null || echo "請手動開啟 coverage.html"; \
	else \
		echo "❌ 找不到 coverage.html，請先執行 make test-coverage"; \
		exit 1; \
	fi

# 顯示函數級覆蓋率
coverage-func:
	@if [ -f coverage.out ]; then \
		echo "📊 函數級覆蓋率："; \
		go tool cover -func=coverage.out; \
	else \
		echo "❌ 找不到 coverage.out，請先執行 make test-coverage"; \
		exit 1; \
	fi

# ==================== Docker 相關 ====================

# 建置 Docker 映像
docker-build:
	@echo "🐳 建置 Docker 映像..."
	docker-compose build --no-cache

# 啟動所有服務
docker-up:
	@echo "🚀 啟動所有服務..."
	docker-compose up -d
	@echo "✅ 服務已啟動"
	@docker-compose ps

# 停止所有服務
docker-down:
	@echo "🛑 停止所有服務..."
	docker-compose down
	@echo "✅ 服務已停止"

# 重啟所有服務
docker-restart:
	@echo "🔄 重啟所有服務..."
	docker-compose down
	docker-compose up -d --build
	@echo "✅ 服務已重啟"
	@docker-compose ps

# 查看應用日誌
docker-logs:
	@echo "📋 查看應用日誌..."
	docker-compose logs -f chat_app_backend

# 清理 Docker 資源
docker-clean:
	@echo "🧹 清理 Docker 資源..."
	docker-compose down -v
	docker system prune -f
	@echo "✅ Docker 資源已清理"

# ==================== 工具相關 ====================

# 格式化程式碼
fmt:
	@echo "✨ 格式化程式碼..."
	go fmt ./...
	@echo "✅ 格式化完成"

# 執行 linter 檢查
lint:
	@echo "🔍 執行 linter 檢查..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint 未安裝，使用 go vet 代替"; \
		go vet ./...; \
	fi

# 執行 go vet 檢查
vet:
	@echo "🔍 執行 go vet 檢查..."
	go vet ./...
	@echo "✅ 檢查完成"

# 整理依賴
mod-tidy:
	@echo "📦 整理依賴..."
	go mod tidy
	@echo "✅ 依賴已整理"

# 下載依賴
mod-download:
	@echo "⬇️  下載依賴..."
	go mod download
	@echo "✅ 依賴已下載"

# ==================== 資料庫相關 ====================

# 啟動資料庫服務（MongoDB + Redis）
db-up:
	@echo "🗄️ 啟動資料庫服務..."
	docker-compose up -d mongodb redis
	@echo "✅ 資料庫服務已啟動"
	@docker-compose ps mongodb redis

# 停止資料庫服務
db-down:
	@echo "🛑 停止資料庫服務..."
	docker-compose stop mongodb redis
	@echo "✅ 資料庫服務已停止"

# 查看資料庫日誌
db-logs:
	@echo "📋 查看資料庫日誌..."
	docker-compose logs -f mongodb redis

# ==================== 快捷組合指令 ====================

# 完整檢查（格式化 + 測試 + 覆蓋率）
check: fmt vet test-coverage
	@echo "✅ 完整檢查完成"

# 快速測試（不產生覆蓋率）
quick-test:
	@echo "⚡ 快速測試..."
	go test -short ./...

# 準備提交（格式化 + 測試）
pre-commit: fmt vet test
	@echo "✅ 準備提交完成"

# 全新安裝（下載依賴 + 建置）
install: mod-download build
	@echo "✅ 安裝完成"
