package main

import (
	"encoding/json"
	"fmt"
	"ggo/database"
	"ggo/models"
	"log"
	"time"
)

func main() {
	// 连接数据库
	dsn := "host=27.154.56.154 user=postgres password=zity123456 dbname=test_zity port=10006 sslmode=disable"
	database.InitPostgres(dsn)

	// 创建测试数据
	testData := []struct {
		UserID  uint   `json:"user_id"`
		Name    string `json:"name"`
		Gold    int    `json:"gold"`
		Level   int    `json:"level"`
		Chapter int    `json:"chapter"`
	}{
		{UserID: 1, Name: "测试玩家1", Gold: 15000, Level: 25, Chapter: 8},
		{UserID: 2, Name: "测试玩家2", Gold: 25000, Level: 30, Chapter: 12},
		{UserID: 3, Name: "测试玩家3", Gold: 18000, Level: 28, Chapter: 10},
		{UserID: 4, Name: "测试玩家4", Gold: 32000, Level: 35, Chapter: 15},
		{UserID: 5, Name: "测试玩家5", Gold: 12000, Level: 22, Chapter: 6},
		{UserID: 6, Name: "测试玩家6", Gold: 28000, Level: 32, Chapter: 13},
		{UserID: 7, Name: "测试玩家7", Gold: 22000, Level: 29, Chapter: 11},
		{UserID: 8, Name: "测试玩家8", Gold: 30000, Level: 33, Chapter: 14},
		{UserID: 9, Name: "测试玩家9", Gold: 16000, Level: 26, Chapter: 9},
		{UserID: 10, Name: "测试玩家10", Gold: 26000, Level: 31, Chapter: 12},
		{UserID: 11, Name: "测试玩家11", Gold: 19000, Level: 27, Chapter: 9},
		{UserID: 12, Name: "测试玩家12", Gold: 24000, Level: 30, Chapter: 11},
	}

	fmt.Println("开始插入测试数据...")

	for _, data := range testData {
		// 构建JSON数据
		jsonData := map[string]interface{}{
			"name":    data.Name,
			"gold":    data.Gold,
			"level":   data.Level,
			"chapter": data.Chapter,
		}

		jsonBytes, err := json.Marshal(jsonData)
		if err != nil {
			log.Printf("Failed to marshal JSON for user %d: %v", data.UserID, err)
			continue
		}

		// 创建Archive记录
		archive := models.Archive{
			UserID:    data.UserID,
			JSONData:  string(jsonBytes),
			V:         1,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}

		// 插入数据库
		result := database.DB.Create(&archive)
		if result.Error != nil {
			log.Printf("Failed to insert archive for user %d: %v", data.UserID, result.Error)
		} else {
			fmt.Printf("插入成功: UserID=%d, Name=%s, Gold=%d, Level=%d, Chapter=%d\n",
				data.UserID, data.Name, data.Gold, data.Level, data.Chapter)
		}
	}

	// 验证插入结果
	var count int64
	database.DB.Model(&models.Archive{}).Count(&count)
	fmt.Printf("\n总共插入了 %d 条记录\n", count)

	fmt.Println("\n测试数据插入完成！现在可以测试排行榜接口了。")
}
