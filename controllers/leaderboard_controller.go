package controllers

import (
	"ggo/utils"
	"net/http"
	"strconv"
	"strings"

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
		"level":   true,
		"chapter": true,
	}
	if !validTypes[rankType] {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的type参数，支持: gold, level, chapter")
		return
	}

	// 使用PostgreSQL的JSON操作函数直接查询和排序 - 优化版本
	var rankQuery []RankQuery

	// 根据排行榜类型构建查询语句，使用高效的jsonb操作符
	var querySQL string
	switch rankType {
	case "gold":
		querySQL = "SELECT user_id, json_data->>'name' as name, CAST(json_data->>'gold' AS INTEGER) as value FROM archives WHERE json_data->>'gold' IS NOT NULL AND json_data->>'gold' ~ '^[0-9]+$' ORDER BY CAST(json_data->>'gold' AS INTEGER) DESC LIMIT 10"
	case "level":
		querySQL = "SELECT user_id, json_data->>'name' as name, CAST(json_data->>'level' AS INTEGER) as value FROM archives WHERE json_data->>'level' IS NOT NULL AND json_data->>'level' ~ '^[0-9]+$' ORDER BY CAST(json_data->>'level' AS INTEGER) DESC LIMIT 10"
	case "chapter":
		querySQL = "SELECT user_id, json_data->>'name' as name, CAST(json_data->>'chapter' AS INTEGER) as value FROM archives WHERE json_data->>'chapter' IS NOT NULL AND json_data->>'chapter' ~ '^[0-9]+$' ORDER BY CAST(json_data->>'chapter' AS INTEGER) DESC LIMIT 10"
	}

	// 执行原生SQL查询
	result := lc.db.Raw(querySQL).Scan(&rankQuery)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取排行榜数据失败: "+result.Error.Error())
		return
	}

	// 构建返回结果
	var playerRanks []PlayerRank
	for i, rank := range rankQuery {
		// 拼接name和user_id为name#ID格式
		displayName := rank.Name + "#" + strconv.Itoa(int(rank.UserID))
		playerRanks = append(playerRanks, PlayerRank{
			Name:  displayName,
			Value: rank.Value,
			Rank:  i + 1,
		})
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

	// 验证type参数
	validTypes := map[string]bool{
		"gold":    true,
		"level":   true,
		"chapter": true,
	}
	if !validTypes[rankType] {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的type参数，支持: gold, level, chapter")
		return
	}

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
		querySQL = "SELECT CAST(json_data->>'gold' AS INTEGER) FROM archives WHERE user_id = ? AND json_data->>'gold' IS NOT NULL AND json_data->>'gold' ~ '^[0-9]+$' ORDER BY CAST(json_data->>'gold' AS INTEGER) DESC LIMIT 1"
		countSQL = "SELECT COUNT(*) FROM archives WHERE CAST(json_data->>'gold' AS INTEGER) > (SELECT COALESCE(CAST(json_data->>'gold' AS INTEGER), 0) FROM archives WHERE user_id = ? AND json_data->>'gold' ~ '^[0-9]+$') AND json_data->>'gold' ~ '^[0-9]+$'"
	case "level":
		querySQL = "SELECT CAST(json_data->>'level' AS INTEGER) FROM archives WHERE user_id = ? AND json_data->>'level' IS NOT NULL AND json_data->>'level' ~ '^[0-9]+$' ORDER BY CAST(json_data->>'level' AS INTEGER) DESC LIMIT 1"
		countSQL = "SELECT COUNT(*) FROM archives WHERE CAST(json_data->>'level' AS INTEGER) > (SELECT COALESCE(CAST(json_data->>'level' AS INTEGER), 0) FROM archives WHERE user_id = ? AND json_data->>'level' ~ '^[0-9]+$') AND json_data->>'level' ~ '^[0-9]+$'"
	case "chapter":
		querySQL = "SELECT CAST(json_data->>'chapter' AS INTEGER) FROM archives WHERE user_id = ? AND json_data->>'chapter' IS NOT NULL AND json_data->>'chapter' ~ '^[0-9]+$' ORDER BY CAST(json_data->>'chapter' AS INTEGER) DESC LIMIT 1"
		countSQL = "SELECT COUNT(*) FROM archives WHERE CAST(json_data->>'chapter' AS INTEGER) > (SELECT COALESCE(CAST(json_data->>'chapter' AS INTEGER), 0) FROM archives WHERE user_id = ? AND json_data->>'chapter' ~ '^[0-9]+$') AND json_data->>'chapter' ~ '^[0-9]+$'"
	}

	// 获取玩家数值
	result := lc.db.Raw(querySQL, userID).Scan(&playerValue)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取玩家数据失败: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "未找到该玩家")
		return
	}

	// 获取排名（比玩家数值高的记录数 + 1）
	result = lc.db.Raw(countSQL, userID).Scan(&totalCount)
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
