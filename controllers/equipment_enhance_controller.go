package controllers

import (
	"ggo/models"
	"ggo/utils"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EquipmentEnhanceController struct {
	db *gorm.DB
}

func NewEquipmentEnhanceController(db *gorm.DB) *EquipmentEnhanceController {
	rand.Seed(time.Now().UnixNano())
	return &EquipmentEnhanceController{db: db}
}

// MergeEquipment 融合装备
func (eec *EquipmentEnhanceController) MergeEquipment(c *gin.Context) {
	var request struct {
		UserID              uint `json:"user_id" binding:"required"`
		MainEquipmentID     uint `json:"main_equipment_id" binding:"required"`     // 主装备ID
		MaterialEquipmentID uint `json:"material_equipment_id" binding:"required"` // 材料装备ID
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 开始事务
	tx := eec.db.Begin()

	// 1. 验证装备存在且属于该用户
	var mainEquipment models.UserEquipment
	var materialEquipment models.UserEquipment

	if err := tx.Preload("EquipmentTemplate").Preload("AdditionalAttrs").
		First(&mainEquipment, request.MainEquipmentID).Error; err != nil || mainEquipment.UserID != request.UserID {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "主装备不存在或不属于该用户")
		return
	}

	if err := tx.Preload("EquipmentTemplate").Preload("AdditionalAttrs").
		First(&materialEquipment, request.MaterialEquipmentID).Error; err != nil || materialEquipment.UserID != request.UserID {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "材料装备不存在或不属于该用户")
		return
	}

	// 2. 检查附加属性数量限制（假设最多5条）
	if len(mainEquipment.AdditionalAttrs) >= 5 {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusBadRequest, "主装备附加属性已达上限")
		return
	}

	// 3. 判断是否继承材料装备的附加属性
	var newAttr *models.EquipmentAdditionalAttr

	// 优先检查材料装备是否有附加属性（30%概率继承）
	if len(materialEquipment.AdditionalAttrs) > 0 && rand.Float64() < 0.3 {
		// 随机选择一条附加属性继承
		randomIndex := rand.Intn(len(materialEquipment.AdditionalAttrs))
		materialAttr := materialEquipment.AdditionalAttrs[randomIndex]

		newAttr = &models.EquipmentAdditionalAttr{
			UserEquipmentID: mainEquipment.ID,
			AttrType:        materialAttr.AttrType,
			AttrValue:       materialAttr.AttrValue,
		}
	} else {
		// 根据材料装备品级概率新增属性
		materialLevel := materialEquipment.EquipmentTemplate.Level
		successRate := eec.getMergeSuccessRate(materialLevel)

		if rand.Float64() < successRate {
			// 生成新的附加属性
			newAttr = eec.generateAdditionalAttr(mainEquipment.EquipmentTemplate.Level)
			if newAttr != nil {
				newAttr.UserEquipmentID = mainEquipment.ID
			}
		}
	}

	// 4. 保存新的附加属性
	if newAttr != nil {
		if err := tx.Create(newAttr).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "融合失败")
			return
		}
	}

	// 5. 删除材料装备
	if err := tx.Delete(&materialEquipment).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "融合失败")
		return
	}

	// 6. 提交事务
	tx.Commit()

	// 7. 重新加载主装备信息
	eec.db.Preload("EquipmentTemplate").Preload("AdditionalAttrs").First(&mainEquipment, mainEquipment.ID)

	response := gin.H{
		"message":        "装备融合完成",
		"success":        newAttr != nil,
		"new_attribute":  newAttr,
		"main_equipment": mainEquipment,
	}

	utils.SuccessResponse(c, response)
}

// getMergeSuccessRate 获取融合成功率
func (eec *EquipmentEnhanceController) getMergeSuccessRate(level int) float64 {
	rates := map[int]float64{
		1: 0.02, // 2%
		2: 0.04, // 4%
		3: 0.08, // 8%
		4: 0.12, // 12%
		5: 0.15, // 15%
		6: 0.25, // 25%
	}
	return rates[level]
}

// generateAdditionalAttr 生成附加属性
func (eec *EquipmentEnhanceController) generateAdditionalAttr(mainLevel int) *models.EquipmentAdditionalAttr {
	// 属性权重配置（稀有属性权重为1，普通属性权重为3）
	commonAttrs := []string{"hp", "attack", "attack_speed", "bullet_speed", "drain", "critical"}
	rareAttrs := []string{"recovery", "instant_kill", "trajectory"}

	// 构建属性池
	var attrPool []string
	for _, attr := range commonAttrs {
		for i := 0; i < 3; i++ { // 普通属性权重3
			attrPool = append(attrPool, attr)
		}
	}
	for _, attr := range rareAttrs {
		attrPool = append(attrPool, attr) // 稀有属性权重1
	}

	// 随机选择属性类型
	attrType := attrPool[rand.Intn(len(attrPool))]

	// 根据主装备等级生成属性值
	attrValue := eec.generateAttrValue(attrType, mainLevel)

	return &models.EquipmentAdditionalAttr{
		AttrType:  attrType,
		AttrValue: attrValue,
	}
}

// generateAttrValue 生成属性值
func (eec *EquipmentEnhanceController) generateAttrValue(attrType string, level int) string {
	switch attrType {
	case "hp":
		ranges := map[int][2]int{
			1: {100, 200}, 2: {200, 300}, 3: {300, 400},
			4: {500, 600}, 5: {700, 800}, 6: {1000, 2000},
		}
		r := ranges[level]
		return strconv.Itoa(rand.Intn(r[1]-r[0]+1) + r[0])

	case "attack":
		ranges := map[int][2]int{
			1: {5, 10}, 2: {10, 15}, 3: {20, 25},
			4: {25, 30}, 5: {35, 40}, 6: {55, 100},
		}
		r := ranges[level]
		return strconv.Itoa(rand.Intn(r[1]-r[0]+1) + r[0])

	case "attack_speed":
		ranges := map[int][2]float64{
			1: {0.1, 0.3}, 2: {0.3, 0.5}, 3: {0.5, 0.8},
			4: {0.8, 1.0}, 5: {1.0, 2.0}, 6: {2.0, 3.0},
		}
		r := ranges[level]
		value := r[0] + rand.Float64()*(r[1]-r[0])
		return strconv.FormatFloat(value, 'f', 2, 64)

	case "bullet_speed":
		ranges := map[int][2]float64{
			1: {0.3, 0.5}, 2: {0.5, 0.8}, 3: {0.8, 1.0},
			4: {1.0, 1.2}, 5: {1.3, 1.5}, 6: {1.5, 2.5},
		}
		r := ranges[level]
		value := r[0] + rand.Float64()*(r[1]-r[0])
		return strconv.FormatFloat(value, 'f', 2, 64)

	case "drain":
		ranges := map[int][2]int{
			1: {1, 3}, 2: {2, 4}, 3: {3, 5},
			4: {3, 7}, 5: {5, 8}, 6: {8, 10},
		}
		r := ranges[level]
		return strconv.Itoa(rand.Intn(r[1]-r[0]+1) + r[0])

	case "critical":
		ranges := map[int][2]int{
			1: {1, 3}, 2: {2, 4}, 3: {3, 5},
			4: {3, 7}, 5: {5, 8}, 6: {8, 15},
		}
		r := ranges[level]
		return strconv.Itoa(rand.Intn(r[1]-r[0]+1) + r[0])

	case "recovery":
		ranges := map[int][2]int{
			2: {3, 5}, 3: {5, 10}, 4: {15, 20},
			5: {25, 30}, 6: {35, 50},
		}
		if r, exists := ranges[level]; exists {
			return strconv.Itoa(rand.Intn(r[1]-r[0]+1) + r[0])
		}

	case "instant_kill":
		ranges := map[int][2]int{
			4: {1, 2}, 5: {2, 3}, 6: {3, 4},
		}
		if r, exists := ranges[level]; exists {
			return strconv.Itoa(rand.Intn(r[1]-r[0]+1) + r[0])
		}

	case "trajectory":
		ranges := map[int]int{
			4: 1, 5: 2, 6: 3,
		}
		if value, exists := ranges[level]; exists {
			return strconv.Itoa(value)
		}
	}

	return "0"
}

// EnhanceEquipment 强化装备
func (eec *EquipmentEnhanceController) EnhanceEquipment(c *gin.Context) {
	// 从JWT获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	// 从URL获取装备ID
	idStr := c.Param("id")
	equipmentID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的装备ID")
		return
	}

	// 开始事务
	tx := eec.db.Begin()

	// 1. 验证装备存在且属于该用户
	var equipment models.UserEquipment
	if err := tx.Preload("EquipmentTemplate").First(&equipment, equipmentID).Error; err != nil || equipment.UserID != userID.(uint) {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "装备不存在或不属于该用户")
		return
	}

	// 2. 验证用户存在
	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 3. 计算强化消耗
	cost := eec.getEnhanceCost(equipment.EquipmentTemplate.Level)
	if user.Gold < cost {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusBadRequest, "金币不足")
		return
	}

	// 4. 扣除金币
	if err := tx.Model(&user).Update("gold", user.Gold-cost).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "扣除金币失败")
		return
	}

	// 5. 计算强化成功率
	successRate := eec.getEnhanceSuccessRate(equipment.EnhanceLevel)

	// 6. 强化判定
	success := rand.Float64() < successRate
	var newLevel int

	if success {
		newLevel = equipment.EnhanceLevel + 1
	} else {
		newLevel = equipment.EnhanceLevel - 1
		if newLevel < 0 {
			newLevel = 0
		}
	}

	// 7. 更新装备强化等级
	if err := tx.Model(&equipment).Update("enhance_level", newLevel).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "强化失败")
		return
	}

	// 8. 提交事务
	tx.Commit()

	// 9. 重新加载装备信息
	eec.db.Preload("EquipmentTemplate").First(&equipment, equipment.ID)

	response := gin.H{
		"message":      "强化完成",
		"success":      success,
		"cost_gold":    cost,
		"current_gold": user.Gold - cost,
		"old_level":    equipment.EnhanceLevel,
		"new_level":    newLevel,
		"equipment":    equipment,
	}

	utils.SuccessResponse(c, response)
}

// getEnhanceCost 获取强化消耗
func (eec *EquipmentEnhanceController) getEnhanceCost(level int) int {
	costs := map[int]int{
		1: 10000,  // 1万
		2: 30000,  // 3万
		3: 50000,  // 5万
		4: 80000,  // 8万
		5: 100000, // 10万
		6: 200000, // 20万
	}
	return costs[level]
}

// getEnhanceSuccessRate 获取强化成功率
func (eec *EquipmentEnhanceController) getEnhanceSuccessRate(currentLevel int) float64 {
	rates := map[int]float64{
		0: 0.90, 1: 0.90, 2: 0.90, // 1~3: 90%
		3: 0.80, 4: 0.80, // 3~5: 80%
		5: 0.70, 6: 0.70, 7: 0.70, // 5~8: 70%
		8: 0.60, 9: 0.60, // 8~10: 60%
		10: 0.50, 11: 0.50, 12: 0.50, // 10~13: 50%
		13: 0.40, 14: 0.40, // 13~15: 40%
		15: 0.30, 16: 0.30, 17: 0.30, // 15~18: 30%
	}

	// 18级以上统一20%
	if currentLevel >= 18 {
		return 0.20
	}

	return rates[currentLevel]
}
