package main

import (
	"fmt"
	"ggo/config"
	"ggo/database"
	"log"
	"time"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	database.InitPostgres(cfg.PostgresDSN)
	db := database.GetDB()

	// 获取今天的开始时间
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	todayStartTimestamp := todayStart.UnixMilli()

	fmt.Println("今天开始时间戳:", todayStartTimestamp)

	// 测试SQL查询
	type TestResult struct {
		UserID uint
		Name   string
		Value  int
		UpdatedAt int64
	}

	var results []TestResult

	// 先查询所有包含boss_last_result的记录
	query1 := `SELECT user_id, json_data->>'name' as name, CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) as value, CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) as updated_at FROM archives WHERE json_data#>>'{boss_last_result,damage}' IS NOT NULL`
	result := db.Raw(query1).Scan(&results)
	if result.Error != nil {
		log.Fatalf("查询失败: %v", result.Error)
	}

	fmt.Printf("找到 %d 条包含boss_last_result的记录\n", len(results))
	for _, r := range results {
		fmt.Printf("用户ID: %d, 名称: %s, 伤害: %d, 更新时间: %d (%s)\n", r.UserID, r.Name, r.Value, r.UpdatedAt, time.UnixMilli(r.UpdatedAt))
		if r.UpdatedAt >= todayStartTimestamp {
			fmt.Println("  -> 这条记录应该被包含在今天的查询中")
		} else {
			fmt.Println("  -> 这条记录不应该被包含在今天的查询中")
		}
	}

	// 然后测试带时间过滤的查询
	var filteredResults []TestResult
	query2 := `SELECT user_id, json_data->>'name' as name, CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) as value, CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) as updated_at FROM archives WHERE json_data#>>'{boss_last_result,damage}' IS NOT NULL AND json_data#>>'{boss_last_result,damage}' ~ '^[0-9]+$' AND CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) >= ?`
	result = db.Raw(query2, todayStartTimestamp).Scan(&filteredResults)
	if result.Error != nil {
		log.Fatalf("带过滤的查询失败: %v", result.Error)
	}

	fmt.Printf("\n带时间过滤的查询找到 %d 条记录\n", len(filteredResults))
	for _, r := range filteredResults {
		fmt.Printf("用户ID: %d, 名称: %s, 伤害: %d, 更新时间: %d (%s)\n", r.UserID, r.Name, r.Value, r.UpdatedAt, time.UnixMilli(r.UpdatedAt))
	}
}