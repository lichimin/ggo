package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"ggo/models"
	"ggo/controllers"
)

func main() {
	// 连接数据库
	dsn := "root:123456@tcp(localhost:3306)/game_go?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建控制器实例
	equipmentController := controllers.NewEquipmentController(db)

	// 创建Gin路由
	r := gin.Default()

	// 设置模拟用户ID中间件
	r.Use(func(c *gin.Context) {
		c.Set("userID", uint(1)) // 假设用户ID为1
		c.Next()
	})

	// 测试路由
	r.GET("/test-equip", func(c *gin.Context) {
		c.Param("id", "4") // 测试装备ID 4
		equipmentController.EquipItem(c)
	})

	r.GET("/test-unequip", func(c *gin.Context) {
		c.Param("id", "6") // 测试装备ID 6
		equipmentController.UnequipItem(c)
	})

	// 启动服务器
	go func() {
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	// 测试请求
	fmt.Println("Testing equipment fix...")
	
	// 测试卸下装备
	fmt.Println("1. Testing unequip item (ID: 6)...")
	unequipResp, err := http.Get("http://localhost:8080/test-unequip")
	if err != nil {
		log.Printf("Failed to test unequip: %v", err)
	} else {
		fmt.Printf("Unequip response status: %s\n", unequipResp.Status)
		unequipResp.Body.Close()
	}

	// 测试穿戴装备
	fmt.Println("\n2. Testing equip item (ID: 4)...")
	equipResp, err := http.Get("http://localhost:8080/test-equip")
	if err != nil {
		log.Printf("Failed to test equip: %v", err)
	} else {
		fmt.Printf("Equip response status: %s\n", equipResp.Status)
		equipResp.Body.Close()
	}

	fmt.Println("\nTest completed. Check the server logs for detailed response.")
}