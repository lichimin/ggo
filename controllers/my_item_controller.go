package controllers

import (
	"ggo/models"
	"ggo/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MyItemController struct {
	db *gorm.DB
}

func NewMyItemController(db *gorm.DB) *MyItemController {
	return &MyItemController{db: db}
}

// AddMyItem 添加我的物品
func (mic *MyItemController) AddMyItem(c *gin.Context) {
	// 从context中获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 支持物品数组的请求参数
	type ItemRequest struct {
		ItemID   uint `json:"item_id" binding:"required"`
		Quantity int  `json:"quantity" binding:"min=1"`
	}
	var requests []ItemRequest

	if err := c.ShouldBindJSON(&requests); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if len(requests) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "物品列表不能为空")
		return
	}

	// 验证用户是否存在
	var user models.User
	if err := mic.db.First(&user, userID.(uint)).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 设置默认值
	itemType := "treasure" // 默认物品类型为宝物
	position := "backpack" // 默认位置为背包
	sellPrice := 0         // 默认售价为0

	// 创建响应结构，包含物品详细信息
	type MyItemResponse struct {
		models.MyItem
		ItemName string `json:"item_name"`
	}

	var responses []MyItemResponse

	// 批量处理物品
	for _, req := range requests {
		// 验证宝物是否存在
		var treasure models.Treasure
		if err := mic.db.First(&treasure, req.ItemID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "宝物ID "+strconv.Itoa(int(req.ItemID))+" 不存在")
			return
		}
		itemName := treasure.Name

		// 创建我的物品记录
		myItem := models.MyItem{
			UserID:    userID.(uint),
			ItemID:    req.ItemID,
			ItemType:  itemType,
			SellPrice: sellPrice,
			Position:  position,
			Quantity:  req.Quantity,
		}

		if err := mic.db.Create(&myItem).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "添加物品失败: "+err.Error())
			return
		}

		// 添加到响应列表
		response := MyItemResponse{
			MyItem:   myItem,
			ItemName: itemName,
		}
		responses = append(responses, response)
	}

	utils.SuccessResponse(c, responses)
}

// GetMyItems 获取我的物品列表（装备、皮肤、宝物）
func (mic *MyItemController) GetMyItems(c *gin.Context) {
	// 从context中获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 获取type参数，默认为空字符串（查询所有类型）
	itemType := c.DefaultQuery("type", "")

	// 创建统一的响应结构
	type ItemResponse struct {
		ID        uint                  `json:"id"`
		Type      string                `json:"type"` // treasure, equipment
		ItemID    uint                  `json:"item_id"`
		Treasure  *models.Treasure      `json:"treasure,omitempty"`
		Equipment *models.UserEquipment `json:"equipment,omitempty"`
		Position  string                `json:"position,omitempty"`
		Quantity  int                   `json:"quantity,omitempty"`
		Slot      string                `json:"slot,omitempty"` // 装备部位
	}

	var responseItems []ItemResponse

	// 查询宝物
	if itemType == "" || itemType == "treasure" {
		var treasures []models.MyItem
		query := mic.db.Where("user_id = ? AND item_type = ?", userID.(uint), "treasure")
		if err := query.Find(&treasures).Error; err == nil && len(treasures) > 0 {
			// 批量查询宝物详细信息
			var treasureIDs []uint
			for _, t := range treasures {
				treasureIDs = append(treasureIDs, t.ItemID)
			}

			var treasureDetails []models.Treasure
			if err := mic.db.Where("id IN (?)", treasureIDs).Find(&treasureDetails).Error; err == nil {
				treasureMap := make(map[uint]models.Treasure)
				for _, td := range treasureDetails {
					treasureMap[td.ID] = td
				}

				for _, t := range treasures {
					responseItem := ItemResponse{
						ID:       t.ID,
						Type:     "treasure",
						ItemID:   t.ItemID,
						Position: t.Position,
						Quantity: t.Quantity,
					}
					if td, exists := treasureMap[t.ItemID]; exists {
						responseItem.Treasure = &td
					}
					responseItems = append(responseItems, responseItem)
				}
			}
		}
	}

	// 查询装备（只查询未穿戴的装备）
	if itemType == "" || itemType == "equipment" {
		var equipments []models.UserEquipment
		query := mic.db.Where("user_id = ? AND is_equipped = ?", userID.(uint), false).Preload("EquipmentTemplate").Preload("AdditionalAttrs")
		if err := query.Find(&equipments).Error; err == nil {
			for _, eq := range equipments {
				responseItem := ItemResponse{
					ID:        eq.ID,
					Type:      "equipment",
					ItemID:    eq.EquipmentID,
					Position:  eq.Position,
					Equipment: &eq,
					Slot:      eq.EquipmentTemplate.Slot,
				}
				responseItems = append(responseItems, responseItem)
			}
		}
	}

	utils.SuccessResponse(c, responseItems)
}

// GetEquippedItems 获取已穿戴的装备列表
func (mic *MyItemController) GetEquippedItems(c *gin.Context) {
	// 从context中获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 创建统一的响应结构
	type ItemResponse struct {
		ID        uint                  `json:"id"`
		Type      string                `json:"type"` // treasure, equipment
		ItemID    uint                  `json:"item_id"`
		Treasure  *models.Treasure      `json:"treasure,omitempty"`
		Equipment *models.UserEquipment `json:"equipment,omitempty"`
		Position  string                `json:"position,omitempty"`
		Quantity  int                   `json:"quantity,omitempty"`
		Slot      string                `json:"slot,omitempty"` // 装备部位
	}

	var responseItems []ItemResponse

	// 查询已穿戴的装备
	var equipments []models.UserEquipment
	query := mic.db.Where("user_id = ? AND is_equipped = ?", userID.(uint), true).Preload("EquipmentTemplate").Preload("AdditionalAttrs")
	if err := query.Find(&equipments).Error; err == nil {
		for _, eq := range equipments {
			responseItem := ItemResponse{
				ID:        eq.ID,
				Type:      "equipment",
				ItemID:    eq.EquipmentID,
				Position:  eq.Position,
				Equipment: &eq,
				Slot:      eq.EquipmentTemplate.Slot,
			}
			responseItems = append(responseItems, responseItem)
		}
	}

	utils.SuccessResponse(c, responseItems)
}

// GetMyTreasures 获取我的宝物列表
func (mic *MyItemController) GetMyTreasures(c *gin.Context) {
	// 从context中获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 查询用户拥有的宝物
	var myItems []models.MyItem
	if err := mic.db.Where("user_id = ? AND item_type = ?", userID.(uint), "treasure").Find(&myItems).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询宝物失败: "+err.Error())
		return
	}

	// 定义宝物响应结构
	type TreasureResponse struct {
		ID          uint   `json:"id"`
		MyItemID    uint   `json:"my_item_id"` // 我的物品ID
		Name        string `json:"name"`
		ImageURL    string `json:"image_url"`
		Level       int    `json:"level"`
		Value       int    `json:"value"`
		Quantity    int    `json:"quantity"`
		Description string `json:"description,omitempty"`
	}

	// 如果没有宝物，直接返回空列表
	if len(myItems) == 0 {
		utils.SuccessResponse(c, []TreasureResponse{})
		return
	}

	// 收集所有宝物ID
	var treasureIDs []uint
	for _, item := range myItems {
		treasureIDs = append(treasureIDs, item.ItemID)
	}

	// 查询宝物详细信息
	var treasures []models.Treasure
	if err := mic.db.Where("id IN ?", treasureIDs).Find(&treasures).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询宝物信息失败: "+err.Error())
		return
	}

	// 构建宝物ID到宝物信息的映射
	treasureMap := make(map[uint]models.Treasure)
	for _, treasure := range treasures {
		treasureMap[treasure.ID] = treasure
	}

	// 构建响应列表
	var response []TreasureResponse
	for _, item := range myItems {
		if treasure, exists := treasureMap[item.ItemID]; exists {
			response = append(response, TreasureResponse{
				ID:          treasure.ID,
				MyItemID:    item.ID,
				Name:        treasure.Name,
				ImageURL:    treasure.ImageURL,
				Level:       treasure.Level,
				Value:       treasure.Value,
				Quantity:    item.Quantity,
				Description: treasure.Description,
			})
		}
	}

	utils.SuccessResponse(c, response)
}

// SellMultipleTreasures 批量出售宝物
func (mic *MyItemController) SellMultipleTreasures(c *gin.Context) {
	// 从context中获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 请求参数支持每个物品的数量
	type SellItemRequest struct {
		MyItemID uint `json:"my_item_id" binding:"required"`
		Quantity int  `json:"quantity" binding:"min=1"`
	}

	var request struct {
		Items []SellItemRequest `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if len(request.Items) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "物品列表不能为空")
		return
	}

	// 开始事务
	tx := mic.db.Begin()

	// 1. 提取所有物品ID
	var allMyItemIDs []uint
	for _, item := range request.Items {
		allMyItemIDs = append(allMyItemIDs, item.MyItemID)
	}

	// 2. 查询所有要出售的物品
	var myItems []models.MyItem
	if err := tx.Where("id IN (?) AND user_id = ? AND item_type = ?", allMyItemIDs, userID.(uint), "treasure").Find(&myItems).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询物品失败: "+err.Error())
		return
	}

	if len(myItems) == 0 {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "未找到要出售的宝物")
		return
	}

	// 3. 创建物品ID到物品的映射
	myItemMap := make(map[uint]models.MyItem)
	for _, item := range myItems {
		myItemMap[item.ID] = item
	}

	// 4. 收集宝物ID
	var treasureIDs []uint
	for _, item := range myItems {
		treasureIDs = append(treasureIDs, item.ItemID)
	}

	// 5. 批量查询宝物信息
	var treasures []models.Treasure
	if err := tx.Where("id IN (?)", treasureIDs).Find(&treasures).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询宝物信息失败: "+err.Error())
		return
	}

	// 6. 创建宝物映射表
	treasureMap := make(map[uint]models.Treasure)
	for _, treasure := range treasures {
		treasureMap[treasure.ID] = treasure
	}

	// 7. 处理每个出售请求
	totalPrice := 0
	var soldItems []gin.H

	for _, reqItem := range request.Items {
		// 检查物品是否存在
		myItem, exists := myItemMap[reqItem.MyItemID]
		if !exists {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusNotFound, "未找到ID为 "+strconv.Itoa(int(reqItem.MyItemID))+" 的宝物")
			return
		}

		// 检查宝物信息是否存在
		treasure, exists := treasureMap[myItem.ItemID]
		if !exists {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusNotFound, "未找到宝物信息")
			return
		}

		// 确定出售数量（默认1）
		sellQuantity := reqItem.Quantity
		if sellQuantity <= 0 {
			sellQuantity = 1
		}

		// 检查是否有足够的数量
		if sellQuantity > myItem.Quantity {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusBadRequest, "ID为 "+strconv.Itoa(int(reqItem.MyItemID))+" 的宝物数量不足")
			return
		}

		// 计算该物品的出售价格
		itemSellPrice := myItem.SellPrice
		if itemSellPrice == 0 {
			itemSellPrice = treasure.Value
		}
		itemTotalPrice := itemSellPrice * sellQuantity
		totalPrice += itemTotalPrice

		// 记录出售的物品信息
		soldItems = append(soldItems, gin.H{
			"id":            myItem.ID,
			"item_name":     treasure.Name,
			"item_value":    treasure.Value,
			"sell_quantity": sellQuantity,
			"sold_price":    itemTotalPrice,
		})

		// 更新或删除物品
		if sellQuantity == myItem.Quantity {
			// 数量相等，删除物品
			if err := tx.Delete(&models.MyItem{}, myItem.ID).Error; err != nil {
				tx.Rollback()
				utils.ErrorResponse(c, http.StatusInternalServerError, "出售物品失败: "+err.Error())
				return
			}
		} else {
			// 数量不等，更新数量
			newQuantity := myItem.Quantity - sellQuantity
			if err := tx.Model(&models.MyItem{}).Where("id = ?", myItem.ID).Update("quantity", newQuantity).Error; err != nil {
				tx.Rollback()
				utils.ErrorResponse(c, http.StatusInternalServerError, "更新物品数量失败: "+err.Error())
				return
			}
		}
	}

	// 8. 更新用户金币
	var user models.User
	if err := tx.First(&user, userID.(uint)).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}

	newGold := user.Gold + totalPrice
	if err := tx.Model(&user).Update("gold", newGold).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新金币失败: "+err.Error())
		return
	}

	// 提交事务
	tx.Commit()

	// 返回结果
	response := gin.H{
		"message":      "批量出售成功",
		"total_price":  totalPrice,
		"current_gold": newGold,
		"sold_count":   len(myItems),
		"sold_items":   soldItems,
	}

	utils.SuccessResponse(c, response)
}
