package controllers

import (
	"ggo/models"
	"ggo/utils"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EquipmentController struct {
	db *gorm.DB
}

func NewEquipmentController(db *gorm.DB) *EquipmentController {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())
	return &EquipmentController{db: db}
}

// GenerateEquipment 生成装备
func (ec *EquipmentController) GenerateEquipment(c *gin.Context) {
	var request struct {
		UserID      uint   `json:"user_id" binding:"required"`
		TreasureIDs []uint `json:"treasure_ids" binding:"required,len=3"` // 3个宝物ID
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 开始事务
	tx := ec.db.Begin()

	// 1. 验证用户存在
	var user models.User
	if err := tx.First(&user, request.UserID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 2. 验证宝物存在且属于该用户
	var treasures []models.Treasure
	var myItems []models.MyItem
	var totalLevel int

	for _, treasureID := range request.TreasureIDs {
		var myItem models.MyItem
		if err := tx.Where("user_id = ? AND item_id = ? AND item_type = ?", request.UserID, treasureID, "treasure").First(&myItem).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusNotFound, "宝物不存在或不属于该用户")
			return
		}

		var treasure models.Treasure
		if err := tx.First(&treasure, treasureID).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusNotFound, "宝物信息不存在")
			return
		}

		treasures = append(treasures, treasure)
		myItems = append(myItems, myItem)
		totalLevel += treasure.Level
	}

	// 3. 计算消耗金币
	costGold := totalLevel * 10000
	if user.Gold < costGold {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusBadRequest, "金币不足")
		return
	}

	// 4. 扣除金币
	if err := tx.Model(&user).Update("gold", user.Gold-costGold).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "扣除金币失败")
		return
	}

	// 5. 删除使用的宝物
	for _, myItem := range myItems {
		if err := tx.Delete(&myItem).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "删除宝物失败")
			return
		}
	}

	// 6. 确定生成的装备等级
	equipmentLevel := ec.determineEquipmentLevel(treasures)

	// 7. 随机选择对应等级的装备模板
	var equipmentTemplate models.EquipmentTemplate
	if err := tx.Where("level = ? AND is_active = ?", equipmentLevel, true).Order("RANDOM()").First(&equipmentTemplate).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "没有找到合适的装备模板")
		return
	}

	// 8. 创建玩家装备记录
	userEquipment := models.UserEquipment{
		UserID:      request.UserID,
		EquipmentID: equipmentTemplate.ID,
		Position:    "backpack",
	}

	if err := tx.Create(&userEquipment).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建装备失败")
		return
	}

	// 9. 提交事务
	tx.Commit()

	// 10. 加载装备模板信息
	ec.db.Preload("EquipmentTemplate").First(&userEquipment, userEquipment.ID)

	// 11. 返回结果
	response := gin.H{
		"message":         "装备生成成功",
		"cost_gold":       costGold,
		"current_gold":    user.Gold - costGold,
		"equipment_level": equipmentLevel,
		"user_equipment":  userEquipment,
		"used_treasures":  treasures,
	}

	utils.SuccessResponse(c, response)
}

// determineEquipmentLevel 根据概率确定生成的装备等级
func (ec *EquipmentController) determineEquipmentLevel(treasures []models.Treasure) int {
	// 提取宝物等级
	levels := make([]int, len(treasures))
	for i, treasure := range treasures {
		levels[i] = treasure.Level
	}

	// 排序等级
	sort.Ints(levels)
	minLevel := levels[0]
	maxLevel := levels[2]
	midLevel := levels[1]

	// 根据概率随机选择
	randomValue := rand.Float64()

	if randomValue < 0.2 {
		// 20% 概率生成最高等级装备
		return maxLevel
	} else if randomValue < 0.7 {
		// 50% 概率生成最低等级装备
		return minLevel
	} else {
		// 30% 概率生成中间等级装备
		return midLevel
	}
}

// GetUserEquipments 获取用户装备列表
func (ec *EquipmentController) GetUserEquipments(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	position := c.Query("position") // backpack, warehouse
	isEquipped := c.Query("is_equipped")

	query := ec.db.Preload("EquipmentTemplate").Where("user_id = ?", userID)

	if position != "" {
		query = query.Where("position = ?", position)
	}

	if isEquipped != "" {
		equipped, _ := strconv.ParseBool(isEquipped)
		query = query.Where("is_equipped = ?", equipped)
	}

	var userEquipments []models.UserEquipment
	result := query.Find(&userEquipments)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	utils.SuccessResponse(c, userEquipments)
}

// EquipEquipment 装备装备
func (ec *EquipmentController) EquipEquipment(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	equipmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的装备ID")
		return
	}

	// 开始事务
	tx := ec.db.Begin()

	// 获取装备信息
	var userEquipment models.UserEquipment
	if err := tx.Preload("EquipmentTemplate").First(&userEquipment, equipmentID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "装备不存在")
		return
	}

	// 验证装备属于该用户
	if userEquipment.UserID != userID.(uint) {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusForbidden, "无权操作该装备")
		return
	}

	// 如果已经装备，则取消装备
	if userEquipment.IsEquipped {
		if err := tx.Model(&userEquipment).Update("is_equipped", false).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "取消装备失败")
			return
		}
	} else {
		// 先取消同部位的其他装备
		var equipmentTemplate models.EquipmentTemplate
		tx.First(&equipmentTemplate, userEquipment.EquipmentID)

		if err := tx.Model(&models.UserEquipment{}).
			Where("user_id = ? AND is_equipped = ?", userID, true).
			Joins("JOIN equipment_templates ON user_equipments.equipment_id = equipment_templates.id").
			Where("equipment_templates.slot = ?", equipmentTemplate.Slot).
			Update("is_equipped", false).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "装备失败")
			return
		}

		// 装备当前装备
		if err := tx.Model(&userEquipment).Update("is_equipped", true).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "装备失败")
			return
		}
	}

	tx.Commit()

	// 重新加载数据
	ec.db.Preload("EquipmentTemplate").First(&userEquipment, equipmentID)

	utils.SuccessResponse(c, userEquipment)
}
