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

	// 检查表是否存在
	var count int64
	result := database.DB.Model(&models.Archive{}).Count(&count)
	if result.Error != nil {
		log.Fatal("Failed to count archives:", result.Error)
	}

	fmt.Printf("Total records in archives: %d\n", count)

	if count > 0 {
		// 查看前几条记录
		var archives []models.Archive
		result = database.DB.Limit(3).Find(&archives)
		if result.Error != nil {
			log.Fatal("Failed to fetch archives:", result.Error)
		}

		fmt.Println("\nFirst 3 records:")
		for i, archive := range archives {
			fmt.Printf("Record %d: UserID=%d, JSONData=%s\n", i+1, archive.UserID, archive.JSONData)
		}
	} else {
		fmt.Println("No data found in archives table")
	}

	// 检查数据库版本
	var version string
	result = database.DB.Raw("SELECT version()").Scan(&version)
	if result.Error == nil {
		fmt.Printf("\nPostgreSQL version: %s\n", version)
	}
}
