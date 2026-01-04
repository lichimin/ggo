package main

import (
	"fmt"
	"ggo/config"
	"ggo/database"
	"ggo/routes"
	"log"
	"net/http"
	"time"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	database.InitPostgres(cfg.PostgresDSN)
	database.InitRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	// 设置路由
	router := routes.SetupRoutes(cfg)

	// 启动服务器
	go func() {
		log.Printf("服务器启动在端口 %s", cfg.ServerPort)
		if err := router.Run(cfg.ServerPort); err != nil {
			log.Fatal("服务器启动失败:", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 测试区服列表接口
	fmt.Println("\n=== 测试区服列表接口 ===")
	
	client := &http.Client{Timeout: 10 * time.Second}
	
	// 测试 GET /api/v1/areas
	resp, err := client.Get("http://localhost:8080/api/v1/areas")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ 区服列表接口调用成功")
		fmt.Printf("响应状态: %s\n", resp.Status)
		
		// 读取响应体
		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		fmt.Printf("响应内容: %s\n", string(buf[:n]))
	} else {
		fmt.Printf("❌ 接口调用失败，状态码: %d\n", resp.StatusCode)
	}

	fmt.Println("\n=== 测试完成，服务器继续运行 ===")
	fmt.Println("按 Ctrl+C 停止服务器")
	
	// 保持服务器运行
	select {}
}