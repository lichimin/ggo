package controllers

import (
	"context"
	"encoding/json"
	"ggo/database"
	"ggo/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LeaderboardController 排行榜控制器
type LeaderboardController struct {
	db *gorm.DB
}

// NewLeaderboardController 创建排行榜控制器实例
func NewLeaderboardController(db *gorm.DB) *LeaderboardController {
	return &LeaderboardController{
		db: db,
	}
}

// PlayerRank 玩家排行榜数据结构体
type PlayerRank struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
	Rank  int    `json:"rank"`
}

// RankQuery 排行榜查询结果结构体
type RankQuery struct {
	UserID uint   `json:"user_id" gorm:"column:user_id"`
	Name   string `json:"name" gorm:"column:name"`
	Value  int    `json:"value" gorm:"column:value"`
}

// GetLeaderboard 获取排行榜 - 优化版本，使用数据库层面的查询和排序
func (lc *LeaderboardController) GetLeaderboard(c *gin.Context) {
	// 获取排行榜类型参数
	rankType := c.Query("type")
	if rankType == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "缺少type参数")
		return
	}

	// 验证type参数
	validTypes := map[string]bool{
		"gold":    true,
		"chapter": true,
		"damage":  true,
	}
	if !validTypes[rankType] {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的type参数，支持: gold, chapter, damage")
		return
	}

	// 获取区服参数，默认为1
	area := 1
	if areaParam := c.Query("area"); areaParam != "" {
		if parsedArea, err := strconv.Atoi(areaParam); err == nil && parsedArea > 0 {
			area = parsedArea
		}
	}

	// 获取今天的开始时间（0点0分0秒）
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	todayStartTimestamp := todayStart.UnixMilli()

	// 生成缓存键
	cacheKey := "leaderboard:" + rankType + ":" + strconv.Itoa(area)
	if rankType == "damage" {
		// 伤害排行榜缓存键包含日期，确保每天的数据独立缓存
		cacheKey += ":" + today.Format("2006-01-02")
	}
	ctx := context.Background()

	// 先查询Redis缓存
	cachedData, err := database.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// 缓存存在，解析并返回
		var playerRanks []PlayerRank
		if err := json.Unmarshal([]byte(cachedData), &playerRanks); err == nil {
			utils.SuccessResponse(c, playerRanks)
			return
		}
	}

	// 缓存不存在或解析失败，从数据库查询
	// 使用PostgreSQL的JSON操作函数直接查询和排序 - 优化版本
	var rankQuery []RankQuery

	// 根据排行榜类型构建查询语句，使用高效的jsonb操作符
	var querySQL string
	var queryParams []interface{}
	switch rankType {
	case "gold":
		querySQL = "SELECT user_id, json_data->>'name' as name, CAST(json_data->>'gold' AS INTEGER) as value FROM archives WHERE area = ? AND json_data->>'gold' IS NOT NULL AND json_data->>'gold' ~ '^[0-9]+$' ORDER BY CAST(json_data->>'gold' AS INTEGER) DESC LIMIT 10"
		queryParams = []interface{}{area}
	case "chapter":
		querySQL = "SELECT user_id, json_data->>'name' as name, CAST(json_data->>'chapter' AS INTEGER) as value FROM archives WHERE area = ? AND json_data->>'chapter' IS NOT NULL AND json_data->>'chapter' ~ '^[0-9]+$' ORDER BY CAST(json_data->>'chapter' AS INTEGER) DESC LIMIT 10"
		queryParams = []interface{}{area}
	case "damage":
		// 只查询今天的伤害数据
		querySQL = "SELECT user_id, json_data->>'name' as name, CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) as value FROM archives WHERE area = ? AND json_data#>>'{boss_last_result,damage}' IS NOT NULL AND json_data#>>'{boss_last_result,damage}' ~ '^[0-9]+$' AND CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) >= ? ORDER BY CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) DESC LIMIT 10"
		queryParams = []interface{}{area, todayStartTimestamp}
	}

	// 执行原生SQL查询
	var result *gorm.DB
	if rankType == "damage" {
		result = lc.db.Raw(querySQL, queryParams...).Scan(&rankQuery)
	} else {
		result = lc.db.Raw(querySQL, queryParams...).Scan(&rankQuery)
	}
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取排行榜数据失败: "+result.Error.Error())
		return
	}

	// 构建返回结果
	playerRanks := []PlayerRank{}
	for i, rank := range rankQuery {
		// 拼接name和user_id为name#ID格式
		displayName := rank.Name + "#" + strconv.Itoa(int(rank.UserID))
		playerRanks = append(playerRanks, PlayerRank{
			Name:  displayName,
			Value: rank.Value,
			Rank:  i + 1,
		})
	}

	// 将结果存入Redis，设置10分钟过期时间
	if len(playerRanks) > 0 {
		if jsonData, err := json.Marshal(playerRanks); err == nil {
			database.RedisClient.Set(ctx, cacheKey, jsonData, 5*time.Minute)
		}
	}

	utils.SuccessResponse(c, playerRanks)
}

// GetPlayerRank 获取单个玩家的排名 - 优化版本，使用数据库层面的查询
func (lc *LeaderboardController) GetPlayerRank(c *gin.Context) {
	// 获取排行榜类型参数
	rankType := c.Query("type")
	if rankType == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "缺少type参数")
		return
	}

	// 获取玩家名称参数
	playerName := c.Query("name")
	if playerName == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "缺少name参数")
		return
	}

	// 获取区服参数，默认为1
	area := 1
	if areaParam := c.Query("area"); areaParam != "" {
		if parsedArea, err := strconv.Atoi(areaParam); err == nil && parsedArea > 0 {
			area = parsedArea
		}
	}

	// 验证type参数
	validTypes := map[string]bool{
		"gold":    true,
		"chapter": true,
		"damage":  true,
	}
	if !validTypes[rankType] {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的type参数，支持: gold, chapter, damage")
		return
	}

	// 获取今天的开始时间（0点0分0秒）- 用于伤害排行榜过滤
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	todayStartTimestamp := todayStart.UnixMilli()

	// 获取玩家在指定排行榜中的排名 - 优化版本
	var playerValue int
	var totalCount int
	var querySQL, countSQL string

	// 解析name参数，提取原始name和user_id
	// name格式: "name#ID"
	var userID uint
	if strings.Contains(playerName, "#") {
		// 分割name和ID
		parts := strings.Split(playerName, "#")
		if len(parts) == 2 {
			if parsedID, err := strconv.ParseUint(parts[1], 10, 32); err == nil {
				userID = uint(parsedID)
			} else {
				utils.ErrorResponse(c, http.StatusBadRequest, "无效的玩家名称格式，应为 name#ID")
				return
			}
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "无效的玩家名称格式，应为 name#ID")
			return
		}
	} else {
		utils.ErrorResponse(c, http.StatusBadRequest, "玩家名称格式错误，应为 name#ID")
		return
	}

	// 构建查询SQL，使用高效的jsonb操作符
	switch rankType {
	case "gold":
		querySQL = "SELECT CAST(json_data->>'gold' AS INTEGER) FROM archives WHERE user_id = ? AND area = ? AND json_data->>'gold' IS NOT NULL AND json_data->>'gold' ~ '^[0-9]+$' ORDER BY CAST(json_data->>'gold' AS INTEGER) DESC LIMIT 1"
		countSQL = "SELECT COUNT(*) FROM archives WHERE area = ? AND CAST(json_data->>'gold' AS INTEGER) > (SELECT COALESCE(CAST(json_data->>'gold' AS INTEGER), 0) FROM archives WHERE user_id = ? AND area = ? AND json_data->>'gold' ~ '^[0-9]+$') AND json_data->>'gold' ~ '^[0-9]+$'"
	case "chapter":
		querySQL = "SELECT CAST(json_data->>'chapter' AS INTEGER) FROM archives WHERE user_id = ? AND area = ? AND json_data->>'chapter' IS NOT NULL AND json_data->>'chapter' ~ '^[0-9]+$' ORDER BY CAST(json_data->>'chapter' AS INTEGER) DESC LIMIT 1"
		countSQL = "SELECT COUNT(*) FROM archives WHERE area = ? AND CAST(json_data->>'chapter' AS INTEGER) > (SELECT COALESCE(CAST(json_data->>'chapter' AS INTEGER), 0) FROM archives WHERE user_id = ? AND area = ? AND json_data->>'chapter' ~ '^[0-9]+$') AND json_data->>'chapter' ~ '^[0-9]+$'"
	case "damage":
		querySQL = "SELECT CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) FROM archives WHERE user_id = ? AND area = ? AND json_data#>>'{boss_last_result,damage}' IS NOT NULL AND json_data#>>'{boss_last_result,damage}' ~ '^[0-9]+$' AND CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) >= ? ORDER BY CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) DESC LIMIT 1"
		countSQL = "SELECT COUNT(*) FROM archives WHERE area = ? AND CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER) > (SELECT COALESCE(CAST(json_data#>>'{boss_last_result,damage}' AS INTEGER), 0) FROM archives WHERE user_id = ? AND area = ? AND json_data#>>'{boss_last_result,damage}' ~ '^[0-9]+$' AND CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) >= ?) AND json_data#>>'{boss_last_result,damage}' ~ '^[0-9]+$' AND CAST(json_data#>>'{boss_last_result,updated_at}' AS BIGINT) >= ?"
	}

	// 获取玩家数值
	var result *gorm.DB
	if rankType == "damage" {
		result = lc.db.Raw(querySQL, userID, area, todayStartTimestamp).Scan(&playerValue)
	} else {
		result = lc.db.Raw(querySQL, userID, area).Scan(&playerValue)
	}
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取玩家数据失败: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "未找到该玩家")
		return
	}

	// 获取排名（比玩家数值高的记录数 + 1）
	if rankType == "damage" {
		result = lc.db.Raw(countSQL, area, userID, area, todayStartTimestamp, todayStartTimestamp).Scan(&totalCount)
	} else {
		result = lc.db.Raw(countSQL, area, userID, area).Scan(&totalCount)
	}
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "计算玩家排名失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"name":  playerName, // 返回完整的name#ID格式
		"value": playerValue,
		"rank":  totalCount + 1,
	})
}
