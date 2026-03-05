package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chat_app_backend/app/providers"
	"chat_app_backend/app/services"
	"chat_app_backend/config"
	"chat_app_backend/di"
	"chat_app_backend/routes"
	"chat_app_backend/utils"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func init() {
	// 載入環境變數配置
	config.LoadConfig()

	// 初始化日誌系統
	utils.InitLogger(string(config.AppConfig.Server.Mode))

	// 設置時區
	location, err := time.LoadLocation(config.AppConfig.Server.Timezone)
	if err != nil {
		log.Fatalf("Failed to load location: %v", err)
	}
	time.Local = location
}

func main() {
	// 初始化資料庫連接
	mongodb, err := providers.DBConnect[*providers.MongoWrapper]("mongodb")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongodb.Close()

	// 初始化 Redis 連接
	redis, err := providers.NewRedisClient(config.AppConfig)
	if err != nil {
		log.Fatalf("Redis 初始化失敗: %v", err)
	}

	if err := redis.Ping(); err != nil {
		log.Fatalf("Redis 連線測試失敗: %v", err)
	}
	defer redis.Close()

	// 初始化 Gin
	// 使用 New() 而不是 Default() 以便自定義中間件 (例如使用 JSON Logger)
	r := gin.New()
	r.Use(gin.Recovery())

	// 設置信任的代理
	// 在生產環境中，使用配置中的可信代理設置
	if config.AppConfig.Server.Mode == config.ProductionMode {
		r.SetTrustedProxies(config.AppConfig.Server.TrustedProxies)
	} else {
		// 開發環境不信任任何代理
		r.SetTrustedProxies(nil)
	}

	// 建立全域 Context 以支援優雅停機
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 構建依賴
	deps := di.BuildDependencies(config.AppConfig, mongodb, redis)

	// 啟動 ClientManager 健康檢查器
	go deps.Services.ClientManager.StartHealthChecker(ctx)

	// 使用依賴容器中的 UserService 來啟動後台任務
	backgroundTasks := services.NewBackgroundTasks(deps.Services.UserService)
	go backgroundTasks.StartAllBackgroundTasks(ctx)

	// 註冊 pprof（僅限非生產環境，避免暴露敏感效能資訊）
	if config.AppConfig.Server.Mode != config.ProductionMode {
		pprof.Register(r)
	}

	// 設置路由
	routes.SetupRoutes(r, config.AppConfig, redis, deps.Controllers)

	// 確保上傳目錄存在
	err = os.MkdirAll("uploads", 0755)
	if err != nil {
		log.Fatalf("無法創建上傳目錄: %v", err)
	}

	// 啟動服務器
	srv := &http.Server{
		Addr:    ":" + config.AppConfig.Server.Port,
		Handler: r,
	}

	// 在背景啟動 Server
	go func() {
		log.Println("服務器已啟動，Port: " + config.AppConfig.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("監聽失敗: %v", err)
		}
	}()

	// 阻塞等待終止訊號 (使用者按 Ctrl+C，或 Docker/K8s 發送 SIGTERM)
	<-ctx.Done()
	log.Println("收到關閉訊號，開始優雅停機...")

	// 給伺服器 5 秒鐘處理尚未完成的請求
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("伺服器強制關閉: %v", err)
	}

	log.Println("伺服器已安全退出")
}
