package main

import (
	"fmt"
	"log"
	"time"

	"ggo/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 数据库连接信息（从.env文件获取）
	dsn := "host=27.154.56.154 user=postgres password=zity123456 dbname=test_zity port=10006 sslmode=disable TimeZone=Asia/Shanghai"

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	fmt.Println("数据库连接成功")

	// 获取今天的开始时间戳
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	todayStartTimestamp := todayStart.UnixMilli()

	fmt.Printf("今天开始时间: %v\n", todayStart)
	fmt.Printf("今天开始时间戳: %d\n", todayStartTimestamp)

	// 查询所有存档数据，检查json_data结构
	var archives []models.Archive
	result := db.Find(&archives)
	if result.Error != nil {
		log.Fatalf("查询存档失败: %v", result.Error)
	}

	fmt.Printf("查询到 %d 条存档记录\n", result.RowsAffected)

	// 检查每条存档的json_data
	for i, archive := range archives {
		fmt.Printf("\n存档 %d, UserID: %d\n", i+1, archive.UserID)
		// 安全地显示JSONData的前100个字符
		jsonPreview := archive.JSONData
		if len(jsonPreview) > 100 {
			jsonPreview = jsonPreview[:100] + "..."
		}
		fmt.Printf("JSONData: %s\n", jsonPreview)

		// 直接执行SQL查询检查boss_last_result字段
		var hasBossResult bool
		db.Raw("SELECT json_data#>>'{boss_last_result}' IS NOT NULL AS has_boss_result FROM archives WHERE user_id = ?", archive.UserID).Scan(&hasBossResult)

		fmt.Printf("是否有boss_last_result: %t\n", hasBossResult)

		// 如果有boss_last_result，查询具体值
		if hasBossResult {
			var bossResult struct {
				Damage    int
				UpdatedAt int64
			}

			db.Raw("SELECT CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) as damage, CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) as updated_at FROM archives WHERE user_id = ?", archive.UserID).Scan(&bossResult)

			fmt.Printf("Boss伤害: %d\n", bossResult.Damage)
			fmt.Printf("更新时间戳: %d\n", bossResult.UpdatedAt)
			fmt.Printf("更新时间: %v\n", time.UnixMilli(bossResult.UpdatedAt))
			fmt.Printf("是否在今天范围内: %t\n", bossResult.UpdatedAt >= todayStartTimestamp)
		}
	}

	// 直接执行伤害排行榜的SQL查询
	fmt.Println("\n执行伤害排行榜查询:")
	var rankResults []struct {
		UserID uint
		Name   string
		Value  int
	}

	query := "SELECT user_id, json_data->>'name' as name, CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) as value FROM archives WHERE json_data#>>'{boss_last_result,damage}' IS NOT NULL AND json_data#>>'{boss_last_result,damage}' ~ '^[0-9]+$' AND CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) >= ? ORDER BY CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) DESC LIMIT 10"

	db.Raw(query, todayStartTimestamp).Scan(&rankResults)

	fmt.Printf("查询到 %d 条伤害排行榜记录\n", len(rankResults))
	for i, result := range rankResults {
		fmt.Printf("排名 %d: UserID=%d, Name=%s, Damage=%d\n", i+1, result.UserID, result.Name, result.Value)
	}

	// 测试不带时间过滤的查询
	fmt.Println("\n执行不带时间过滤的伤害查询:")
	var allDamageResults []struct {
		UserID    uint
		Name      string
		Value     int
		UpdatedAt int64
	}

	allQuery := "SELECT user_id, json_data->>'name' as name, CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) as value, CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) as updated_at FROM archives WHERE json_data#>>'{boss_last_result,damage}' IS NOT NULL AND json_data#>>'{boss_last_result,damage}' ~ '^[0-9]+$' ORDER BY CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) DESC"

	db.Raw(allQuery).Scan(&allDamageResults)

	fmt.Printf("查询到 %d 条伤害记录\n", len(allDamageResults))
	for i, result := range allDamageResults {
		fmt.Printf("记录 %d: UserID=%d, Name=%s, Damage=%d, UpdatedAt=%d (%v)\n", i+1, result.UserID, result.Name, result.Value, result.UpdatedAt, time.UnixMilli(result.UpdatedAt))
	}
}
