package controllers

import (
	"encoding/json"
	"ggo/models"
	"ggo/utils"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HeroController struct {
	db *gorm.DB
}

func NewHeroController(db *gorm.DB) *HeroController {
	rand.Seed(time.Now().UnixNano())
	return &HeroController{db: db}
}

// DrawHero 抽取英雄（十连抽）
func (hc *HeroController) DrawHero(c *gin.Context) {
	var request struct {
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 开始事务
	tx := hc.db.Begin()

	// 1. 验证用户存在
	var user models.User
	if err := tx.First(&user, request.UserID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 2. 检查金币是否足够（100万金币）
	costGold := 1000000
	if user.Gold < costGold {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusBadRequest, "金币不足")
		return
	}

	// 3. 扣除金币
	if err := tx.Model(&user).Update("gold", user.Gold-costGold).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "扣除金币失败")
		return
	}

	// 4. 进行十次抽取
	var results []gin.H
	totalGoldReward := 0

	for i := 0; i < 10; i++ {
		result := hc.singleDraw(tx, request.UserID)
		results = append(results, result)

		// 如果是金币奖励，累加到总金币
		if gold, ok := result["gold_reward"].(int); ok {
			totalGoldReward += gold
		}
	}

	// 5. 如果有金币奖励，更新用户金币
	if totalGoldReward > 0 {
		if err := tx.Model(&user).Update("gold", user.Gold-costGold+totalGoldReward).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "发放金币奖励失败")
			return
		}
	}

	// 6. 提交事务
	tx.Commit()

	// 7. 返回结果
	response := gin.H{
		"message":      "抽取完成",
		"cost_gold":    costGold,
		"gold_reward":  totalGoldReward,
		"final_gold":   user.Gold - costGold + totalGoldReward,
		"draw_results": results,
	}

	utils.SuccessResponse(c, response)
}

// singleDraw 单次抽取
func (hc *HeroController) singleDraw(tx *gorm.DB, userID uint) gin.H {
	// 1. 获取所有可抽取的英雄
	var heroes []models.Hero
	if err := tx.Where("is_active = ?", true).Find(&heroes).Error; err != nil {
		return gin.H{"type": "gold", "gold_reward": hc.generateGoldReward()}
	}

	// 2. 计算总概率
	totalProbability := 0.0
	for _, hero := range heroes {
		totalProbability += hero.DrawProbability
	}

	// 3. 随机决定是否抽中英雄
	randomValue := rand.Float64() * totalProbability

	// 4. 遍历英雄列表，确定抽中的英雄
	currentProbability := 0.0
	var drawnHero models.Hero
	heroDrawn := false

	for _, hero := range heroes {
		currentProbability += hero.DrawProbability
		if randomValue <= currentProbability {
			drawnHero = hero
			heroDrawn = true
			break
		}
	}

	// 5. 如果抽中英雄，创建用户英雄记录
	if heroDrawn {
		// 检查用户是否已经拥有该英雄
		var existingUserHero models.UserHero
		if err := tx.Where("user_id = ? AND hero_id = ?", userID, drawnHero.ID).First(&existingUserHero).Error; err != nil {
			// 用户没有该英雄，创建新记录
			userHero := models.UserHero{
				UserID:   userID,
				HeroID:   drawnHero.ID,
				IsActive: false, // 新获得的英雄默认不启用
			}

			if err := tx.Create(&userHero).Error; err != nil {
				return gin.H{"type": "gold", "gold_reward": hc.generateGoldReward()}
			}

			return gin.H{
				"type":        "hero",
				"hero_id":     drawnHero.ID,
				"hero_name":   drawnHero.Name,
				"rarity":      drawnHero.Rarity,
				"attack_type": drawnHero.AttackType,
				"is_new":      true,
			}
		} else {
			// 用户已经拥有该英雄，转换为金币奖励
			return gin.H{"type": "gold", "gold_reward": hc.generateGoldReward()}
		}
	}

	// 6. 没有抽中英雄，返回金币奖励
	return gin.H{"type": "gold", "gold_reward": hc.generateGoldReward()}
}

// generateGoldReward 生成金币奖励（1000~100000，平均期望10000）
func (hc *HeroController) generateGoldReward() int {
	// 使用指数分布来生成更符合期望的随机数
	minGold := 1000
	maxGold := 100000
	expectedGold := 10000.0

	// 使用正态分布近似，但限制在范围内
	for {
		// 生成正态分布随机数
		gold := int(rand.NormFloat64()*float64(maxGold-minGold)/3 + expectedGold)

		// 确保在范围内
		if gold >= minGold && gold <= maxGold {
			return gold
		}
	}
}

// AwakenHero 觉醒英雄
func (hc *HeroController) AwakenHero(c *gin.Context) {
	var request struct {
		UserID         uint `json:"user_id" binding:"required"`
		MainHeroID     uint `json:"main_hero_id" binding:"required"`     // 主英雄ID
		MaterialHeroID uint `json:"material_hero_id" binding:"required"` // 材料英雄ID
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 开始事务
	tx := hc.db.Begin()

	// 1. 验证英雄存在且属于该用户
	var mainHero models.UserHero
	var materialHero models.UserHero

	if err := tx.Preload("Hero").First(&mainHero, request.MainHeroID).Error; err != nil || mainHero.UserID != request.UserID {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "主英雄不存在或不属于该用户")
		return
	}

	if err := tx.Preload("Hero").First(&materialHero, request.MaterialHeroID).Error; err != nil || materialHero.UserID != request.UserID {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "材料英雄不存在或不属于该用户")
		return
	}

	// 2. 验证是否为同一种英雄
	if mainHero.HeroID != materialHero.HeroID {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusBadRequest, "只能使用同种英雄进行觉醒")
		return
	}

	// 3. 检查觉醒等级是否已达上限
	if mainHero.AwakenLevel >= 5 {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusBadRequest, "觉醒等级已达上限")
		return
	}

	// 4. 删除材料英雄
	if err := tx.Delete(&materialHero).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "觉醒失败")
		return
	}

	// 5. 更新主英雄觉醒等级和属性加成
	newAwakenLevel := mainHero.AwakenLevel + 1
	bonusRate := 0.10 // 每次觉醒增加10%

	newAttackBonus := mainHero.AttackBonus + bonusRate
	newCriticalBonus := mainHero.CriticalBonus + bonusRate
	newSpeedBonus := mainHero.SpeedBonus + bonusRate

	updates := map[string]interface{}{
		"awaken_level":   newAwakenLevel,
		"attack_bonus":   newAttackBonus,
		"critical_bonus": newCriticalBonus,
		"speed_bonus":    newSpeedBonus,
	}

	if err := tx.Model(&mainHero).Updates(updates).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "觉醒失败")
		return
	}

	// 6. 提交事务
	tx.Commit()

	// 7. 重新加载主英雄信息
	hc.db.Preload("Hero").First(&mainHero, mainHero.ID)

	// 8. 计算实际属性值
	baseHero := mainHero.Hero
	actualAttack := float64(baseHero.Attack) * (1 + mainHero.AttackBonus)
	actualCritical := baseHero.CriticalRate * (1 + mainHero.CriticalBonus)
	actualSpeed := baseHero.AttackSpeed * (1 + mainHero.SpeedBonus)

	response := gin.H{
		"message":      "觉醒成功",
		"awaken_level": newAwakenLevel,
		"main_hero":    mainHero,
		"actual_stats": gin.H{
			"attack":        actualAttack,
			"critical_rate": actualCritical,
			"attack_speed":  actualSpeed,
		},
		"bonus_rates": gin.H{
			"attack_bonus":   mainHero.AttackBonus,
			"critical_bonus": mainHero.CriticalBonus,
			"speed_bonus":    mainHero.SpeedBonus,
		},
	}

	utils.SuccessResponse(c, response)
}

// GetUserHeroes 获取用户英雄列表
func (hc *HeroController) GetUserHeroes(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	var userHeroes []models.UserHero
	result := hc.db.Preload("Hero").Where("user_id = ?", userID).Find(&userHeroes)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 计算实际属性值
	var response []gin.H
	for _, userHero := range userHeroes {
		baseHero := userHero.Hero
		actualAttack := float64(baseHero.Attack) * (1 + userHero.AttackBonus)
		actualCritical := baseHero.CriticalRate * (1 + userHero.CriticalBonus)
		actualSpeed := baseHero.AttackSpeed * (1 + userHero.SpeedBonus)

		// 解析图片JSON
		var images []string
		json.Unmarshal([]byte(baseHero.Images), &images)

		response = append(response, gin.H{
			"user_hero_id": userHero.ID,
			"hero_id":      baseHero.ID,
			"name":         baseHero.Name,
			"images":       images,
			"attack_type":  baseHero.AttackType,
			"rarity":       baseHero.Rarity,
			"is_active":    userHero.IsActive,
			"awaken_level": userHero.AwakenLevel,
			"actual_stats": gin.H{
				"attack":        actualAttack,
				"critical_rate": actualCritical,
				"attack_speed":  actualSpeed,
			},
			"base_stats": gin.H{
				"attack":        baseHero.Attack,
				"critical_rate": baseHero.CriticalRate,
				"attack_speed":  baseHero.AttackSpeed,
			},
			"bonus_rates": gin.H{
				"attack_bonus":   userHero.AttackBonus,
				"critical_bonus": userHero.CriticalBonus,
				"speed_bonus":    userHero.SpeedBonus,
			},
		})
	}

	utils.SuccessResponse(c, response)
}
