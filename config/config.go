package config

import (
	"log"
	"os"
	"sync"

	"github.com/spf13/viper"
)

var (
	instance *Config
	once     sync.Once
)

type Config struct {
	Server struct {
		Port string `mapstructure:"port"`
		Mode string `mapstructure:"mode"`
	} `mapstructure:"server"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	Database       struct {
		Type    string `mapstructure:"type"` // 資料庫類型
		MongoDB struct {
			Host       string `mapstructure:"host"`
			Port       string `mapstructure:"port"`
			Username   string `mapstructure:"username"`
			Password   string `mapstructure:"password"`
			DBName     string `mapstructure:"dbname"`
			AuthSource string `mapstructure:"authSource"`
		} `mapstructure:"mongodb"`
		PostgreSQL struct {
			Host     string `mapstructure:"host"`
			Port     string `mapstructure:"port"`
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
			DBName   string `mapstructure:"dbname"`
			SSLMode  string `mapstructure:"sslmode"`
		} `mapstructure:"postgresql"`
	} `mapstructure:"database"`
	JWT struct {
		AccessToken struct {
			Secret      string  `mapstructure:"secret"`
			ExpireHours float32 `mapstructure:"expire_hours"`
		} `mapstructure:"access_token"`

		RefreshToken struct {
			Secret      string  `mapstructure:"secret"`
			ExpireHours float32 `mapstructure:"expire_hours"`
		} `mapstructure:"refresh_token"`
	} `mapstructure:"jwt"`
}

// GetConfig 使用單例模式載入配置
func GetConfig() *Config {
	once.Do(func() {
		// 獲取當前工作目錄
		workingDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error getting working directory: %v", err)
		}

		viper.SetConfigFile(workingDir + "/config.yaml")
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error loading config file: %v", err)
		}

		instance = &Config{}
		if err := viper.Unmarshal(instance); err != nil {
			log.Fatalf("Error unmarshalling config: %v", err)
		}
	})
	return instance
}
