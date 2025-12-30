package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort      string
	PostgresDSN     string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	WeChatAppID     string
	WeChatAppSecret string
}

func LoadConfig() *Config {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	} else {
		log.Println(".env file loaded successfully")
	}

	return &Config{
		ServerPort:      getEnv("SERVER_PORT", ":8080"),
		PostgresDSN:     getEnv("POSTGRES_DSN", "host=localhost user=postgres password=postgres dbname=ggo port=5432 sslmode=disable"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         getEnvAsInt("REDIS_DB", 0),
		WeChatAppID:     getEnv("WECHAT_APP_ID", ""),
		WeChatAppSecret: getEnv("WECHAT_APP_SECRET", ""),
	}
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
