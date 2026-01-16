package main

import (
	"ggo/config"
	"ggo/database"
	"ggo/routes"
	"log"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	database.InitPostgres(cfg.PostgresDSN)
	database.InitRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	startDailyBossDamageRewardScheduler()

	// 设置路由并启动服务
	router := routes.SetupRoutes(cfg)

	log.Printf("Server starting on %s", cfg.ServerPort)
	if err := router.Run(cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
