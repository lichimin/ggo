package services

import (
	"context"
	"fmt"
	"ggo/database"
	"ggo/models"
	"time"

	"gorm.io/gorm"
)

type bossDamageRankRow struct {
	UserID uint
	Value  int
}

func StartDailyBossDamageRewardScheduler() {
	db := database.DB
	if db == nil || database.RedisClient == nil {
		return
	}

	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.Local
	}

	go func() {
		for {
			now := time.Now().In(location)
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, location)
			time.Sleep(time.Until(next))

			runDailyBossDamageRewards(database.DB, location)
		}
	}()
}

func runDailyBossDamageRewards(db *gorm.DB, location *time.Location) {
	if db == nil || database.RedisClient == nil {
		return
	}

	now := time.Now().In(location)
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	start := end.AddDate(0, 0, -1)

	startMs := start.UnixMilli()
	endMs := end.UnixMilli()
	dateKey := start.Format("20060102")

	areaValues, err := listAreasForRewards(db)
	if err != nil {
		return
	}

	ctx := context.Background()
	for _, area := range areaValues {
		lockKey := fmt.Sprintf("reward:boss_damage:%d:%s", area, dateKey)
		ok, err := database.RedisClient.SetNX(ctx, lockKey, "1", 72*time.Hour).Result()
		if err != nil || !ok {
			continue
		}

		var rows []bossDamageRankRow
		querySQL := "SELECT user_id, CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) as value FROM archives WHERE area = ? AND json_data#>>'{boss_last_result,damage}' IS NOT NULL AND json_data#>>'{boss_last_result,damage}' ~ '^[0-9]+$' AND CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) >= ? AND CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) < ? ORDER BY CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) DESC LIMIT 10"
		if err := db.Raw(querySQL, area, startMs, endMs).Scan(&rows).Error; err != nil {
			continue
		}

		if len(rows) == 0 {
			continue
		}

		mails := make([]models.Mail, 0, len(rows))
		for i, row := range rows {
			rank := i + 1
			diamond := rewardDiamondByRank(rank)
			if diamond <= 0 {
				continue
			}

			mails = append(mails, models.Mail{
				UserID:   row.UserID,
				Area:     area,
				Title:    "每日首领排行榜奖励",
				Content:  fmt.Sprintf("您在今天的首领排行榜中排行第%d名，这是您的奖励。", rank),
				ItemType: "diamond",
				ItemID:   0,
				Num:      diamond,
				Status:   0,
			})
		}

		if len(mails) == 0 {
			continue
		}

		_ = db.CreateInBatches(&mails, 1000).Error
	}
}

func listAreasForRewards(db *gorm.DB) ([]int, error) {
	var areas []models.Area
	if err := db.Select("area").Order("area asc").Find(&areas).Error; err == nil && len(areas) > 0 {
		out := make([]int, 0, len(areas))
		for _, a := range areas {
			if a.Area > 0 {
				out = append(out, a.Area)
			}
		}
		if len(out) > 0 {
			return out, nil
		}
	}

	var out []int
	if err := db.Model(&models.Archive{}).Distinct("area").Order("area asc").Pluck("area", &out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func rewardDiamondByRank(rank int) int {
	switch rank {
	case 1:
		return 1200
	case 2:
		return 1000
	case 3:
		return 800
	default:
		if rank >= 4 && rank <= 10 {
			return 500
		}
		return 0
	}
}
