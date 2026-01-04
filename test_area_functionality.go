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

	fmt.Println("=== 数据库迁移测试 ===")

	// 1. 检查archives表是否包含area字段
	fmt.Println("1. 检查archives表结构:")
	var archivesCount int64
	result := database.DB.Model(&models.Archive{}).Count(&archivesCount)
	if result.Error != nil {
		log.Fatal("Failed to count archives:", result.Error)
	}
	fmt.Printf("archives表总记录数: %d\n", archivesCount)

	// 2. 创建areas表并插入测试数据
	fmt.Println("\n2. 初始化areas表数据:")

	// 先尝试创建表
	err := database.DB.AutoMigrate(&models.Area{})
	if err != nil {
		fmt.Printf("表创建可能失败: %v\n", err)
	}

	// 插入默认区服数据
	areas := []models.Area{
		{Area: 1, IsNew: false, Status: 1, Name: "一区", MaxUsers: 1000},
		{Area: 2, IsNew: true, Status: 1, Name: "二区", MaxUsers: 1000},
		{Area: 3, IsNew: true, Status: 1, Name: "三区", MaxUsers: 1000},
		{Area: 4, IsNew: false, Status: 1, Name: "四区", MaxUsers: 1000},
		{Area: 5, IsNew: false, Status: 1, Name: "五区", MaxUsers: 1000},
	}

	for _, area := range areas {
		// 使用 Where 条件来避免重复插入
		var existingArea models.Area
		result := database.DB.Where("area = ?", area.Area).First(&existingArea)
		if result.Error != nil {
			if result.Error.Error() == "record not found" {
				// 记录不存在，创建新记录
				result := database.DB.Create(&area)
				if result.Error != nil {
					fmt.Printf("插入区服 %d 失败: %v\n", area.Area, result.Error)
				} else {
					fmt.Printf("区服 %s (ID: %d) 初始化成功\n", area.Name, area.Area)
				}
			} else {
				fmt.Printf("查询区服 %d 失败: %v\n", area.Area, result.Error)
			}
		} else {
			fmt.Printf("区服 %d 已存在，跳过插入\n", area.Area)
		}
	}

	// 3. 查询并显示区服列表
	fmt.Println("\n3. 查询区服列表:")
	var areaList []models.Area
	result = database.DB.Order("area ASC").Find(&areaList)
	if result.Error != nil {
		log.Fatal("Failed to fetch areas:", result.Error)
	}

	fmt.Printf("找到 %d 个区服:\n", len(areaList))
	for _, area := range areaList {
		fmt.Printf("  ID:%d, 区服:%d, 新服:%v, 状态:%d, 名称:%s\n",
			area.ID, area.Area, area.IsNew, area.Status, area.Name)
	}

	// 4. 模拟存档的area查询
	fmt.Println("\n4. 模拟存档按area查询:")
	if archivesCount > 0 {
		var testArchive models.Archive
		result = database.DB.First(&testArchive)
		if result.Error == nil {
			fmt.Printf("示例存档 - UserID: %d, Area: %d\n", testArchive.UserID, testArchive.Area)
		}
	}

	fmt.Println("\n=== 测试完成 ===")
}
