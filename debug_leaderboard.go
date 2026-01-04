package main

import (
	"fmt"
	"ggo/database"
	"ggo/models"
	"log"
)

func main() {
	// 连接数据库
	dsn := "host=27.154.56.154 user=postgres password=zity123456 dbname=test_zity port=10006 sslmode=disable"
	database.InitPostgres(dsn)

	fmt.Println("=== 排行榜调试查询 ===")

	// 1. 查看总记录数
	var totalCount int64
	result := database.DB.Model(&models.Archive{}).Count(&totalCount)
	if result.Error != nil {
		log.Fatal("Failed to count archives:", result.Error)
	}
	fmt.Printf("总记录数: %d\n", totalCount)

	// 2. 查看前3条记录的原始JSON
	var archives []models.Archive
	result = database.DB.Limit(3).Find(&archives)
	if result.Error != nil {
		log.Fatal("Failed to fetch archives:", result.Error)
	}

	fmt.Println("\n前3条记录的JSON数据:")
	for i, archive := range archives {
		fmt.Printf("记录 %d: UserID=%d\n", i+1, archive.UserID)
		fmt.Printf("JSONData: %s\n\n", archive.JSONData)
	}

	// 3. 直接测试JSON提取
	fmt.Println("3. JSON字段提取测试:")
	for _, archive := range archives {
		var goldValue, nameValue string
		result := database.DB.Raw("SELECT json_data->>'gold', json_data->>'name' FROM archives WHERE user_id = ?", archive.UserID).Row().Scan(&goldValue, &nameValue)
		if result != nil {
			fmt.Printf("UserID %d: 提取失败 - %v\n", archive.UserID, result)
			continue
		}
		fmt.Printf("UserID %d: gold='%s', name='%s'\n", archive.UserID, goldValue, nameValue)
	}

	// 4. 简化版本的排行榜查询
	fmt.Println("\n4. 简化排行榜查询:")
	var simpleResults []struct {
		UserID uint   `json:"user_id"`
		Name   string `json:"name"`
		Value  int    `json:"value"`
	}

	querySQL := "SELECT user_id, json_data->>'name' as name, CAST(json_data->>'gold' AS INTEGER) as value FROM archives WHERE json_data->>'gold' IS NOT NULL ORDER BY CAST(json_data->>'gold' AS INTEGER) DESC LIMIT 10"
	result = database.DB.Raw(querySQL).Scan(&simpleResults)
	if result.Error != nil {
		fmt.Printf("简化查询失败: %v\n", result.Error)
	} else {
		fmt.Printf("简化查询返回 %d 条记录\n", len(simpleResults))
		for _, r := range simpleResults {
			fmt.Printf("  %s#%d: %d\n", r.Name, r.UserID, r.Value)
		}
	}

	// 5. 检查json_data字段类型
	fmt.Println("\n5. 检查json_data字段类型:")
	var fieldInfo []struct {
		ColumnName string
		DataType   string
	}
	result = database.DB.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'archives' AND column_name = 'json_data'").Scan(&fieldInfo)
	if result.Error != nil {
		fmt.Printf("字段类型查询失败: %v\n", result.Error)
	} else {
		for _, f := range fieldInfo {
			fmt.Printf("字段: %s, 类型: %s\n", f.ColumnName, f.DataType)
		}
	}
}
