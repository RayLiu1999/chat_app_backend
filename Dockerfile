# 使用官方 Go 鏡像作為構建階段
FROM golang:1.23-alpine AS builder

# 安裝必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 設置工作目錄
WORKDIR /app

# 複製 go mod 和 sum 檔案
COPY go.mod go.sum ./

# 下載依賴
RUN go mod tidy

# 複製源代碼
COPY . .

# 構建應用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 使用輕量級的 alpine 鏡像作為運行階段
FROM alpine:latest

# 安裝必要的包
RUN apk --no-cache add ca-certificates curl tzdata

# 設置時區
RUN cp /usr/share/zoneinfo/Asia/Taipei /etc/localtime && echo "Asia/Taipei" > /etc/timezone

# 創建應用用戶
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 設置工作目錄
WORKDIR /app

# 從構建階段複製二進制檔案
COPY --from=builder /app/main .

# 創建必要的目錄
RUN mkdir -p uploads logs && chown -R appuser:appgroup /app

# 切換到應用用戶
USER appuser

# 運行應用
CMD ["./main"]
