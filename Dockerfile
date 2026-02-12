# ---------------------------------------------------
# 1. Base Stage: 共用基礎設置
# ---------------------------------------------------
FROM golang:1.25.3-alpine AS base

# 安裝基礎工具
RUN apk update && apk upgrade && \
    apk add --no-cache git ca-certificates tzdata

# 設置時區
RUN cp /usr/share/zoneinfo/Asia/Taipei /etc/localtime && echo "Asia/Taipei" > /etc/timezone

WORKDIR /app

# 下載依賴 (利用 Docker Layer Cache)
COPY go.mod go.sum ./
RUN go mod download

# ---------------------------------------------------
# 2. Development Stage: 開發環境
# ---------------------------------------------------
FROM base AS dev

# 安裝 air 用於熱重載
RUN go install github.com/air-verse/air@latest

# 創建開發用使用者 (使用系統預設 ID)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 設置權限
RUN mkdir -p uploads && chown -R appuser:appgroup /app

USER appuser
CMD ["air"]

# ---------------------------------------------------
# 3. Builder Stage: 編譯應用
# ---------------------------------------------------
FROM base AS builder

# 版本資訊 ARGs
ARG VERSION=1.0.0
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown
ARG GIT_BRANCH=unknown

# 複製源代碼
COPY . .

# 編譯
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-X chat_app_backend/version.Version=${VERSION} \
              -X chat_app_backend/version.BuildTime=${BUILD_TIME} \
              -X chat_app_backend/version.GitCommit=${GIT_COMMIT} \
              -X chat_app_backend/version.GitBranch=${GIT_BRANCH}" \
    -o main .

# ---------------------------------------------------
# 4. Production Stage: 生產環境運行
# ---------------------------------------------------
FROM alpine:latest AS prod

RUN apk --no-cache add ca-certificates curl tzdata

# 設置時區
RUN cp /usr/share/zoneinfo/Asia/Taipei /etc/localtime && echo "Asia/Taipei" > /etc/timezone

# 定義 build args (允許在生產環境指定 UID/GID)
ARG APP_UID=1000
ARG APP_GID=1000

# 創建應用用戶
RUN addgroup -g ${APP_GID} -S appgroup && adduser -u ${APP_UID} -G appgroup -S appuser

WORKDIR /app

# 從 builder 階段複製編譯好的程式
COPY --from=builder /app/main .

# 創建必要的目錄
RUN mkdir -p uploads && chown -R appuser:appgroup /app

USER appuser
CMD ["./main"]
