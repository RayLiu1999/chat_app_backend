// config/config.go
package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Upload   UploadConfig
	MinIO    MinIOConfig
}
type ModeConfig string

const (
	DevelopmentMode ModeConfig = "development"
	ProductionMode  ModeConfig = "production"
	TestMode        ModeConfig = "test"
)

type ServerConfig struct {
	MainDomain     string
	Port           string
	BaseURL        string
	Mode           ModeConfig
	Timezone       string
	AllowedOrigins []string
	TrustedProxies []string
}

type DatabaseConfig struct {
	MongoURI        string
	MongoUsername   string
	MongoPassword   string
	MongoDBName     string
	MongoAuthSource string
}

type RedisConfig struct {
	Addr     string
	Password string
}

type JWTConfig struct {
	AccessSecret        string
	RefreshSecret       string
	AccessExpireMinutes int
	RefreshExpireHours  int
}

type UploadConfig struct {
	MaxSize      int64
	AllowedTypes []string
}

type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	PublicURL       string
}

var AppConfig *Config

func LoadConfig() {
	// 根據環境載入對應的 .env 檔案
	// - production: 預設 .env
	// - others:     預設 .env.development
	envFile := ".env"
	if os.Getenv("ENV") == string(DevelopmentMode) {
		envFile = ".env.development"
	}

	if err := godotenv.Load(envFile); err != nil {
		if envFile == ".env" {
			log.Printf("未找到 %s 檔案，請確保在生產環境中已正確設置環境變數", envFile)
		} else {
			log.Printf("未找到 %s 檔案，使用系統環境變數", envFile)
		}
	}

	AppConfig = &Config{
		Server: ServerConfig{
			MainDomain:     getEnv("SERVER_MAIN_DOMAIN", "localhost"),
			Port:           getEnv("SERVER_PORT", "8080"),
			BaseURL:        getEnv("SERVER_BASE_URL", "http://localhost"),
			Mode:           ModeConfig(getEnv("ENV", "development")),
			Timezone:       getEnv("TIMEZONE", "Asia/Taipei"),
			AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
			TrustedProxies: strings.Split(getEnv("TRUSTED_PROXIES", "127.0.0.1,::1,172.16.0.0/12,10.0.0.0/8,192.168.0.0/16"), ","),
		},
		Database: DatabaseConfig{
			MongoURI:        getEnv("MONGO_URI", "localhost:27017"),
			MongoUsername:   getEnv("MONGO_USERNAME", ""),
			MongoPassword:   getEnv("MONGO_PASSWORD", ""),
			MongoDBName:     getEnv("MONGO_DB_NAME", "chat_app"),
			MongoAuthSource: getEnv("MONGO_AUTH_SOURCE", "admin"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		JWT: JWTConfig{
			AccessSecret:        getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret:       getEnv("JWT_REFRESH_SECRET", ""),
			AccessExpireMinutes: getEnvAsInt("JWT_ACCESS_EXPIRE_MINUTES", 30),
			RefreshExpireHours:  getEnvAsInt("JWT_REFRESH_EXPIRE_HOURS", 168),
		},
		Upload: UploadConfig{
			MaxSize:      getEnvAsInt64("UPLOAD_MAX_SIZE", 10485760),
			AllowedTypes: strings.Split(getEnv("UPLOAD_ALLOWED_TYPES", "image/jpeg,image/png"), ","),
		},
		MinIO: MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:          getEnv("MINIO_USE_SSL", "false") == "true",
			BucketName:      getEnv("MINIO_BUCKET_NAME", "chat-app-uploads"),
			PublicURL:       getEnv("MINIO_PUBLIC_URL", "http://localhost:9000"),
		},
	}

	// 驗證必要的配置
	validateConfig()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func validateConfig() {
	if AppConfig.JWT.AccessSecret == "" {
		log.Fatal("JWT_ACCESS_SECRET is required")
	}
	if AppConfig.JWT.RefreshSecret == "" {
		log.Fatal("JWT_REFRESH_SECRET is required")
	}
	if AppConfig.Database.MongoURI == "" {
		log.Fatal("MONGO_URI is required")
	}
}

// IsProduction 是一個輔助函數，方便判斷是否為生產環境
func IsProduction() bool {
	return AppConfig.Server.Mode == ProductionMode
}
