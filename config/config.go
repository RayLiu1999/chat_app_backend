package config

import (
	"log"
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
	Database struct {
		Type    string `mapstructure:"type"` // 数据库类型
		MongoDB struct {
			Host       string `mapstructure:"host"`
			Port       string `mapstructure:"port"`
			Username   string `mapstructure:"username"`
			Password   string `mapstructure:"password"`
			DBName     string `mapstructure:"dbname"`
			AuthSource string `mapstructure:"authSource"`
		} `mapstructure:"mongodb"`
		MySQL struct {
			Host     string `mapstructure:"host"`
			Port     string `mapstructure:"port"`
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
			DBName   string `mapstructure:"dbname"`
		} `mapstructure:"mysql"`

		PostgreSQL struct {
			Host     string `mapstructure:"host"`
			Port     string `mapstructure:"port"`
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
			DBName   string `mapstructure:"dbname"`
		} `mapstructure:"postgresql"`
	} `mapstructure:"database"`
	JWT struct {
		Secret      string `mapstructure:"secret"`
		ExpireHours int    `mapstructure:"expire_hours"`
	} `mapstructure:"jwt"`
}

// GetConfig 使用单例模式加载配置
func GetConfig() *Config {
	once.Do(func() {
		viper.SetConfigFile("config.yaml")
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
