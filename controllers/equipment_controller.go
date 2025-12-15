package controllers

import (
	"fmt"
	"ggo/models"
	"ggo/utils"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
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
	// 从context获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	var request struct {
		ItemIDs []uint `json:"itemids" binding:"required,min=3,max=3"` // 固定3个宝物ID，可以重复
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 开始事务
	tx := ec.db.Begin()

	// 1. 验证用户存在
	var user models.User
	if err := tx.First(&user, userID.(uint)).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 2. 验证宝物存在且属于该用户，并且数量大于0
	var treasures []models.Treasure
	var myItems []models.MyItem

	for _, myItemID := range request.ItemIDs {
		// 先根据itemID找到MyItem记录
		var myItem models.MyItem
		if err := tx.Where("id = ? AND user_id = ? AND item_type = ?", myItemID, userID.(uint), "treasure").First(&myItem).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusNotFound, "宝物不存在或不属于该用户")
			return
		}

		// 检查宝物数量是否大于0
		if myItem.Quantity <= 0 {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusBadRequest, "宝物数量不足")
			return
		}

		// 然后使用MyItem中的item_id作为treasureID搜索宝物信息
		var treasure models.Treasure
		if err := tx.First(&treasure, myItem.ItemID).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusNotFound, "宝物信息不存在")
			return
		}

		treasures = append(treasures, treasure)
		myItems = append(myItems, myItem)
	}

	// 3. 固定消耗5万金币
	costGold := 50000
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

	// 5. 减少使用的宝物数量
	for _, myItem := range myItems {
		// 减少数量
		myItem.Quantity--
		if myItem.Quantity <= 0 {
			// 数量为0时删除记录
			if err := tx.Delete(&myItem).Error; err != nil {
				tx.Rollback()
				utils.ErrorResponse(c, http.StatusInternalServerError, "删除宝物失败")
				return
			}
		} else {
			// 数量不为0时更新数量
			if err := tx.Save(&myItem).Error; err != nil {
				tx.Rollback()
				utils.ErrorResponse(c, http.StatusInternalServerError, "更新宝物数量失败")
				return
			}
		}
	}

	// 6. 从提供的宝物中随机选择一件的品级
	randomIndex := rand.Intn(len(treasures))
	equipmentLevel := treasures[randomIndex].Level

	// 7. 随机选择对应等级的装备模板
	var equipmentTemplate models.EquipmentTemplate
	if err := tx.Where("level = ? AND is_active = ?", equipmentLevel, true).Order("RANDOM()").First(&equipmentTemplate).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "没有找到合适的装备模板")
		return
	}

	// 8. 创建玩家装备记录
	userEquipment := models.UserEquipment{
		UserID:      userID.(uint),
		EquipmentID: equipmentTemplate.ID,
		Position:    "backpack",
	}

	if err := tx.Create(&userEquipment).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建装备失败")
		return
	}

	// 9. 处理附加属性
	var newAttr *models.EquipmentAdditionalAttr
	randomValue := rand.Float64()

	if randomValue < 0.5 {
		// 1% 概率添加稀有属性（七宗罪）
		newAttr = ec.generateRareAttr(equipmentLevel)
	} else if randomValue < 0.9 {
		// 10% 概率添加普通属性
		newAttr = ec.generateCommonAttr(equipmentLevel)
	}

	// 10. 保存附加属性
	if newAttr != nil {
		newAttr.UserEquipmentID = userEquipment.ID
		if err := tx.Create(newAttr).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "创建附加属性失败")
			return
		}
	}

	// 11. 提交事务
	tx.Commit()

	// 12. 加载装备完整信息
	ec.db.Preload("EquipmentTemplate").Preload("AdditionalAttrs").First(&userEquipment, userEquipment.ID)

	// 13. 返回结果
	response := gin.H{
		"message":                 "装备打造成功",
		"cost_gold":               costGold,
		"current_gold":            user.Gold - costGold,
		"equipment_level":         equipmentLevel,
		"user_equipment":          userEquipment,
		"used_treasures":          treasures,
		"selected_treasure_index": randomIndex,
	}

	utils.SuccessResponse(c, response)
}

// generateCommonAttr 生成普通附加属性
func (ec *EquipmentController) generateCommonAttr(level int) *models.EquipmentAdditionalAttr {
	// 普通属性类型和范围
	commonAttrs := []struct {
		Type      string
		Name      string
		Min, Max  float64
		IsPercent bool
	}{
		{"attack_bonus", "攻击力加成", 1, 3, true},  // 攻击力加成1~3%
		{"critical_rate", "暴击率", 1, 3, true},   // 暴击率1~3%
		{"drain", "吸血", 1, 3, true},            // 吸血1~3%
		{"damage_reduction", "减伤", 1, 3, true}, // 减伤1~3%
		{"recovery", "生命恢复", 20, 100, false},   // 自动回复20-100
		{"attack_fixed", "攻击力", 10, 30, false}, // 攻击力+10~30
		{"hp_bonus", "血量加成", 1, 5, true},       // 血量加成1%~5%
	}

	// 随机选择一个属性
	attr := commonAttrs[rand.Intn(len(commonAttrs))]
	value := attr.Min + rand.Float64()*(attr.Max-attr.Min)

	// 格式化属性值
	var attrValue string
	if attr.IsPercent {
		attrValue = strconv.FormatFloat(value, 'f', 1, 64) + "%"
	} else {
		attrValue = strconv.Itoa(int(value))
	}

	return &models.EquipmentAdditionalAttr{
		AttrType:  attr.Type,
		AttrName:  attr.Name, // 设置普通属性名称
		AttrValue: attrValue,
	}
}

// generateRareAttr 生成稀有附加属性（七宗罪）
func (ec *EquipmentController) generateRareAttr(level int) *models.EquipmentAdditionalAttr {
	// 稀有属性类型和范围
	rareAttrs := []struct {
		Type      string
		Name      string
		Min, Max  float64
		IsPercent bool
	}{
		{"sloth", "傲慢", 5, 20, true},     // 最大HP提升5-20%
		{"envy", "嫉妒", 10, 20, true},     // 暴击伤害提升10-20%
		{"gluttony", "暴食", 10, 20, true}, // 攻速提升10-20%
		{"greed", "贪婪", 1, 1, true},      // 攻击时1%几率秒杀
		{"lust", "色欲", 100, 200, false},  // 自动回复100~200
		{"wrath", "暴怒", 8, 10, true},     // 暴击率8~10%
		{"pride", "怠惰", 5, 20, true},     // 攻击力提升5-20%
	}

	// 随机选择一个属性
	attr := rareAttrs[rand.Intn(len(rareAttrs))]
	value := attr.Min + rand.Float64()*(attr.Max-attr.Min)

	// 格式化属性值
	var attrValue string
	if attr.IsPercent {
		if attr.Type == "greed" {
			attrValue = "秒杀1%"
		} else {
			attrValue = strconv.FormatFloat(value, 'f', 1, 64) + "%"
		}
	} else {
		attrValue = strconv.Itoa(int(value))
	}

	return &models.EquipmentAdditionalAttr{
		AttrType:  attr.Type,
		AttrName:  attr.Name, // 设置稀有属性名称
		AttrValue: attrValue,
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

// EquipItem 穿戴装备
func (ec *EquipmentController) EquipItem(c *gin.Context) {
	// 从context获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 获取装备ID参数
	equipmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的装备ID")
		return
	}

	// 开始事务
	tx := ec.db.Begin()

	// 1. 查询装备信息并预加载装备模板
	var userEquipment models.UserEquipment
	if err := tx.Preload("EquipmentTemplate").First(&userEquipment, equipmentID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "装备不存在")
		return
	}

	// 2. 验证装备归属
	if userEquipment.UserID != userID.(uint) {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusForbidden, "无权操作该装备")
		return
	}

	// 3. 如果装备已经穿戴，返回错误
	if userEquipment.IsEquipped {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusBadRequest, "该装备已经穿戴")
		return
	}

	// 4. 获取装备的Slot信息
	slot := userEquipment.EquipmentTemplate.Slot

	// 5. 卸下用户在同一Slot上已穿戴的其他装备
	// 即使没有找到需要卸下的装备，也不应报错，继续执行
	// 使用子查询方式确保正确的JOIN操作
	subQuery := tx.Table("equipment_templates").Select("id").Where("slot = ?", slot)
	result := tx.Model(&models.UserEquipment{}).
		Where("user_id = ? AND is_equipped = ? AND equipment_id IN (?)", userID, true, subQuery).
		Update("is_equipped", false)

	// 只有在数据库操作发生错误时才回滚，没有匹配记录不是错误
	if result.Error != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "卸下同部位装备失败")
		return
	}

	// 6. 更新当前装备为穿戴状态，并更新位置为"equipped"
	if err := tx.Model(&userEquipment).Updates(map[string]interface{}{
		"is_equipped": true,
		"position":    "equipped",
	}).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "穿戴装备失败")
		return
	}

	// 提交事务
	tx.Commit()

	// 重新加载装备信息
	ec.db.Preload("EquipmentTemplate").Preload("AdditionalAttrs").First(&userEquipment, equipmentID)

	// 返回成功响应
	utils.SuccessResponse(c, gin.H{
		"message":     "装备穿戴成功",
		"equipment":   userEquipment,
		"slot":        slot,
		"is_equipped": true,
	})
}

// UnequipItem 卸下装备
func (ec *EquipmentController) UnequipItem(c *gin.Context) {
	// 从context获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 获取装备ID参数
	equipmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的装备ID")
		return
	}

	// 开始事务
	tx := ec.db.Begin()

	// 1. 查询装备信息
	var userEquipment models.UserEquipment
	if err := tx.Preload("EquipmentTemplate").First(&userEquipment, equipmentID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "装备不存在")
		return
	}

	// 2. 验证装备归属
	if userEquipment.UserID != userID.(uint) {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusForbidden, "无权操作该装备")
		return
	}

	// 3. 检查装备是否已穿戴
	if !userEquipment.IsEquipped {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusBadRequest, "该装备未穿戴")
		return
	}

	// 4. 更新装备为未穿戴状态，并更新位置为"backpack"
	if err := tx.Model(&userEquipment).Updates(map[string]interface{}{
		"is_equipped": false,
		"position":    "backpack",
	}).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "卸下装备失败")
		return
	}

	// 提交事务
	tx.Commit()

	// 重新加载装备信息
	ec.db.Preload("EquipmentTemplate").Preload("AdditionalAttrs").First(&userEquipment, equipmentID)

	// 返回成功响应，包含slot字段
	utils.SuccessResponse(c, gin.H{
		"message":     "装备卸下成功",
		"equipment":   userEquipment,
		"slot":        userEquipment.EquipmentTemplate.Slot,
		"is_equipped": false,
	})
}

// GetMyEquipment 获取我的装备，返回穿戴中和未穿戴的装备数组
func (ec *EquipmentController) GetMyEquipment(c *gin.Context) {
	// 从context获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 查询用户的所有装备
	var allEquipments []models.UserEquipment
	if err := ec.db.Where("user_id = ?", userID.(uint)).
		Preload("EquipmentTemplate").
		Preload("AdditionalAttrs").
		Find(&allEquipments).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询装备失败: "+err.Error())
		return
	}

	// 定义装备响应结构
	type EquipmentResponse struct {
		ID              uint    `json:"id"`
		Name            string  `json:"name"`
		ImageURL        string  `json:"image_url"`
		Rarity          string  `json:"rarity"`
		Slot            string  `json:"slot"`
		EnhanceLevel    int     `json:"enhance_level"` // 强化等级
		BaseAttributes  gin.H   `json:"base_attributes"`
		AdditionalAttrs []gin.H `json:"additional_attrs"`
		RareAttrs       []gin.H `json:"rare_attrs"`
	}

	// 稀有度映射
	rarityMap := map[int]string{
		1: "普通",
		2: "稀有",
		3: "史诗",
		4: "传说",
		5: "神话",
		6: "创世",
	}

	// 部位映射
	slotMap := map[string]string{
		"weapon": "武器",
		"helmet": "头盔",
		"chest":  "胸甲",
		"gloves": "手套",
		"pants":  "护腿",
		"boots":  "鞋子",
	}

	// 附加属性映射
	additionalAttrMap := map[string]string{
		// 普通属性映射
		"attack_bonus":     "攻击力加成",
		"critical_rate":    "暴击率",
		"drain":            "吸血",
		"damage_reduction": "减伤",
		"recovery":         "生命恢复",
		"attack_fixed":     "攻击力",
		"hp_bonus":         "血量加成",
		// 稀有属性映射
		"sloth":    "最大HP提升",
		"envy":     "暴击伤害提升",
		"gluttony": "攻速提升",
		"greed":    "秒杀几率",
		"lust":     "生命恢复",
		"wrath":    "暴击率",
		"pride":    "攻击力提升",
	}

	// 分类装备
	var equipped []EquipmentResponse
	var unequipped []EquipmentResponse

	for _, eq := range allEquipments {
		// 转换基础属性为中文
		baseAttrs := gin.H{}
		if eq.EquipmentTemplate.HP > 0 {
			baseAttrs["生命值"] = eq.EquipmentTemplate.HP
		}
		if eq.EquipmentTemplate.Attack > 0 {
			baseAttrs["攻击力"] = eq.EquipmentTemplate.Attack
		}
		if eq.EquipmentTemplate.AttackSpeed != 1.0 {
			baseAttrs["攻击速度"] = eq.EquipmentTemplate.AttackSpeed
		}
		if eq.EquipmentTemplate.MoveSpeed > 0 {
			baseAttrs["移动速度"] = fmt.Sprintf("%d%%", eq.EquipmentTemplate.MoveSpeed)
		}
		if eq.EquipmentTemplate.BulletSpeed > 0 {
			baseAttrs["子弹速度"] = fmt.Sprintf("%d%%", eq.EquipmentTemplate.BulletSpeed)
		}
		if eq.EquipmentTemplate.Drain > 0 {
			baseAttrs["吸血"] = fmt.Sprintf("%d%%", eq.EquipmentTemplate.Drain)
		}
		if eq.EquipmentTemplate.Critical > 0 {
			baseAttrs["暴击"] = eq.EquipmentTemplate.Critical
		}
		if eq.EquipmentTemplate.Dodge > 0 {
			baseAttrs["闪避"] = fmt.Sprintf("%d%%", eq.EquipmentTemplate.Dodge)
		}
		if eq.EquipmentTemplate.InstantKill > 0 {
			baseAttrs["秒杀"] = fmt.Sprintf("%d%%", eq.EquipmentTemplate.InstantKill)
		}
		if eq.EquipmentTemplate.Recovery > 0 {
			baseAttrs["生命恢复"] = eq.EquipmentTemplate.Recovery
		}
		if eq.EquipmentTemplate.Trajectory > 0 {
			baseAttrs["弹道"] = eq.EquipmentTemplate.Trajectory
		}

		// 转换附加属性为中文
		addAttrs := []gin.H{}
		rareAttrs := []gin.H{}

		for _, attr := range eq.AdditionalAttrs {
			// 清理属性值，去除百分号和其他非数字字符
			cleanValue := attr.AttrValue
			if strings.Contains(cleanValue, "%") {
				cleanValue = strings.ReplaceAll(cleanValue, "%", "")
			}
			if strings.Contains(cleanValue, "秒杀") {
				cleanValue = strings.ReplaceAll(cleanValue, "秒杀", "")
			}

			// 将清理后的字符串转换为float64
			attrValue, err := strconv.ParseFloat(cleanValue, 64)
			if err != nil {
				attrValue = 0.0
			}

			// 获取属性名称
			attrName := additionalAttrMap[attr.AttrType]

			// 格式化属性值
			var formattedValue string
			if attr.AttrType == "greed" { // 特殊处理秒杀属性
				formattedValue = "秒杀1%"
			} else if attr.AttrType == "lust" || attr.AttrType == "attack_fixed" { // 非百分比属性
				formattedValue = strconv.Itoa(int(attrValue))
			} else { // 百分比属性
				formattedValue = fmt.Sprintf("%.1f%%", attrValue)
			}

			if attr.AttrName != "" {
				// 稀有属性：格式化为"傲慢·最大HP提升15%"这样
				fullName := fmt.Sprintf("%s·%s %s", attr.AttrName, attrName, formattedValue)
				rareAttrs = append(rareAttrs, gin.H{
					"name": fullName,
				})
			} else {
				// 普通附加属性：只显示"最大HP提升15%"
				addAttrs = append(addAttrs, gin.H{
					"name":  attrName,
					"value": formattedValue,
				})
			}
		}

		// 构建装备响应
		equipmentResp := EquipmentResponse{
			ID:              eq.ID,
			Name:            eq.EquipmentTemplate.Name,
			ImageURL:        eq.EquipmentTemplate.ImageURL,
			Rarity:          rarityMap[eq.EquipmentTemplate.Level],
			Slot:            slotMap[eq.EquipmentTemplate.Slot],
			EnhanceLevel:    eq.EnhanceLevel, // 设置强化等级
			BaseAttributes:  baseAttrs,
			AdditionalAttrs: addAttrs,
			RareAttrs:       rareAttrs,
		}

		// 分类到穿戴或未穿戴
		if eq.IsEquipped {
			equipped = append(equipped, equipmentResp)
		} else {
			unequipped = append(unequipped, equipmentResp)
		}
	}

	// 返回结果
	response := gin.H{
		"equipped":   equipped,
		"unequipped": unequipped,
	}

	utils.SuccessResponse(c, response)
}
