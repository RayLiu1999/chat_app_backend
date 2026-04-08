# chat_app_backend

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Test Coverage](https://img.shields.io/badge/coverage-~40%25-yellow)](./docs/TEST_COVERAGE_SUMMARY.md)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](.)

**Live Demo: [https://chat-app.liu-yucheng.com/](https://chat-app.liu-yucheng.com/)**

> 🌍 **Languages:** [English](README.md) | 繁體中文

## 專案簡介

本專案為一個即時聊天室後端，模仿 Discord 架構，支援伺服器（Server/Guild）、頻道（Channel/Room）、私訊（DM）、好友系統與檔案上傳等功能。後端採用 Go 語言開發，資料儲存採用 MongoDB，並整合 Redis 進行資料快取與連線狀態管理。

## 關鍵功能

- **即時通訊**: 基於 WebSocket 的即時訊息傳輸。
- **微服務架構就緒**: 模組化的 Controller → Service → Repository 設計模式。
- **身份驗證**: 支援 JWT（Access/Refresh Token）雙 Token 驗證機制與嚴格的 CSRF 防護。
- **測試能力**: 高度重視單元測試，具備集中化 Mock 架構設計（整體覆蓋率 ~40%，Middleware 高達 94+%）。
- **部署就緒**: 完整支援 Docker 容器化、Kubernetes (K3s/OrbStack) 集群部署以及已就緒的 GitOps 配置。
- **負載測試**: 內建基於 K6 的冒煙測試 (Smoke Test) 與容量測試腳本。

## 快速啟動

### 環境需求
請確認您的機器已安裝 Docker、Docker Compose 以及 `make`。

### 本地開發流程

```bash
# 1. 初始化專案相依套件
make init

# 2. 設定環境變數
cp .env.example .env.development

# 3. 啟動開發環境 (基於 Docker Compose)
make dev

# 4. 執行全部測試
make test
```

## 常用腳本與指令

我們提供了一套完整的 `Makefile` 來簡化本地開發的流程：

- `make dev` - 啟動開發環境。
- `make logs` - 追蹤並檢視容器日誌。
- `make test` - 執行單元測試。
- `make test-smoke` - 執行輕量級的 k6 冒煙測試。
- `make k8s-deploy` - 將應用程式部署至本地的 Kubernetes 叢集 (OrbStack/Minikube)。
- `make k8s-delete` - 移除本地 Kubernetes 上的部署資源。

*若需查看所有可用的指令，請執行 `make help`。*

## 技術棧

- **程式語言:** Go 1.23+
- **Web 框架:** Gin Web Framework
- **即時通訊:** gorilla/websocket
- **資料庫:** MongoDB
- **快取系統:** Redis
- **基礎設施:** Docker & Kubernetes (Kustomize)
