package main

import (
	"fmt"
	"ggo/database"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 从环境变量获取数据库连接信息
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=ggo port=5432 sslmode=disable"
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	fmt.Println("Connected to database successfully")

	// 设置全局DB变量
	database.DB = db

	// 测试查询gold排行榜
	fmt.Println("\n=== Testing Gold Leaderboard ===")
	var goldResults []struct {
		UserID uint
		Name   string
		Value  int
	}
	goldSQL := "SELECT user_id, json_data->>'name' as name, CAST(json_data->>'gold' AS INTEGER) as value FROM archives WHERE json_data->>'gold' IS NOT NULL AND json_data->>'gold' ~ '^[0-9]+$' ORDER BY CAST(json_data->>'gold' AS INTEGER) DESC LIMIT 10"
	err = db.Raw(goldSQL).Scan(&goldResults).Error
	if err != nil {
		fmt.Println("Gold query error:", err)
	} else {
		fmt.Printf("Found %d gold records\n", len(goldResults))
		for _, r := range goldResults {
			fmt.Printf("  UserID: %d, Name: %s, Gold: %d\n", r.UserID, r.Name, r.Value)
		}
	}

	// 测试查询chapter排行榜
	fmt.Println("\n=== Testing Chapter Leaderboard ===")
	var chapterResults []struct {
		UserID uint
		Name   string
		Value  int
	}
	chapterSQL := "SELECT user_id, json_data->>'name' as name, CAST(json_data->>'chapter' AS INTEGER) as value FROM archives WHERE json_data->>'chapter' IS NOT NULL AND json_data->>'chapter' ~ '^[0-9]+$' ORDER BY CAST(json_data->>'chapter' AS INTEGER) DESC LIMIT 10"
	err = db.Raw(chapterSQL).Scan(&chapterResults).Error
	if err != nil {
		fmt.Println("Chapter query error:", err)
	} else {
		fmt.Printf("Found %d chapter records\n", len(chapterResults))
		for _, r := range chapterResults {
			fmt.Printf("  UserID: %d, Name: %s, Chapter: %d\n", r.UserID, r.Name, r.Value)
		}
	}

	// 测试查询damage排行榜
	fmt.Println("\n=== Testing Damage Leaderboard ===")
	// 获取今天的开始时间
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	todayStartTimestamp := todayStart.UnixMilli()
	fmt.Printf("Today: %s, TodayStartTimestamp: %d\n", today.Format("2006-01-02 15:04:05"), todayStartTimestamp)

	var damageResults []struct {
		UserID uint
		Name   string
		Value  int
	}
	damageSQL := "SELECT user_id, json_data->>'name' as name, CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) as value FROM archives WHERE json_data#>>'{boss_last_result,damage}' IS NOT NULL AND json_data#>>'{boss_last_result,damage}' ~ '^[0-9]+$' AND CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) >= ? ORDER BY CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) DESC LIMIT 10"
	err = db.Raw(damageSQL, todayStartTimestamp).Scan(&damageResults).Error
	if err != nil {
		fmt.Println("Damage query error:", err)
	} else {
		fmt.Printf("Found %d damage records\n", len(damageResults))
		for _, r := range damageResults {
			fmt.Printf("  UserID: %d, Name: %s, Damage: %d\n", r.UserID, r.Name, r.Value)
		}
	}

	// 测试不添加时间过滤的damage查询
	fmt.Println("\n=== Testing Damage Leaderboard (Without Time Filter) ===")
	var damageResultsNoTime []struct {
		UserID    uint
		Name      string
		Value     int
		UpdatedAt int64
	}
	damageSQLNoTime := "SELECT user_id, json_data->>'name' as name, CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) as value, CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) as updated_at FROM archives WHERE json_data#>>'{boss_last_result,damage}' IS NOT NULL AND json_data#>>'{boss_last_result,damage}' ~ '^[0-9]+$' ORDER BY CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) DESC LIMIT 10"
	err = db.Raw(damageSQLNoTime).Scan(&damageResultsNoTime).Error
	if err != nil {
		fmt.Println("Damage query (no time) error:", err)
	} else {
		fmt.Printf("Found %d damage records (no time filter)\n", len(damageResultsNoTime))
		for _, r := range damageResultsNoTime {
			fmt.Printf("  UserID: %d, Name: %s, Damage: %d, UpdatedAt: %d (%s)\n",
				r.UserID, r.Name, r.Value, r.UpdatedAt,
				time.UnixMilli(r.UpdatedAt).Format("2006-01-02 15:04:05"))
		}
	}

	// 检查Archive表的结构
	fmt.Println("\n=== Checking Archive Table Structure ===")
	var columns []struct {
		ColumnName string
		DataType   string
	}
	db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name='archives'").Scan(&columns)
	for _, col := range columns {
		fmt.Printf("  %s: %s\n", col.ColumnName, col.DataType)
	}

	// 查询所有Archive记录的数量
	fmt.Println("\n=== Checking Archive Records ===")
	var count int64
	db.Table("archives").Count(&count)
	fmt.Printf("Total archive records: %d\n", count)

	// 查询一条Archive记录的json_data内容
	fmt.Println("\n=== Checking Single Archive json_data ===")
	var jsonContent string
	db.Raw("SELECT json_data FROM archives LIMIT 1").Scan(&jsonContent)
	fmt.Printf("First archive json_data: %s\n", jsonContent)
	if jsonContent != "" {
		fmt.Printf("json_data length: %d\n", len(jsonContent))
	}
}
