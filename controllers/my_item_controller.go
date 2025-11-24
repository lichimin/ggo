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
	var request struct {
		UserID    uint   `json:"user_id" binding:"required"`
		ItemID    uint   `json:"item_id" binding:"required"`
		ItemType  string `json:"item_type" binding:"required,oneof=treasure equipment"`
		SellPrice int    `json:"sell_price" binding:"min=0"`
		Position  string `json:"position" binding:"required,oneof=backpack warehouse equipped"`
		Quantity  int    `json:"quantity" binding:"min=1"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 验证用户是否存在
	var user models.User
	if err := mic.db.First(&user, request.UserID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 根据物品类型验证物品是否存在
	switch request.ItemType {
	case "treasure":
		var treasure models.Treasure
		if err := mic.db.First(&treasure, request.ItemID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "宝物不存在")
			return
		}
	case "equipment":
		// 这里可以添加装备的验证逻辑
		// var equipment models.Equipment
		// if err := mic.db.First(&equipment, request.ItemID).Error; err != nil {
		// 	utils.ErrorResponse(c, http.StatusNotFound, "装备不存在")
		// 	return
		// }
	default:
		utils.ErrorResponse(c, http.StatusBadRequest, "不支持的物品类型")
		return
	}

	// 创建我的物品记录
	myItem := models.MyItem{
		UserID:    request.UserID,
		ItemID:    request.ItemID,
		ItemType:  request.ItemType,
		SellPrice: request.SellPrice,
		Position:  request.Position,
		Quantity:  request.Quantity,
	}

	result := mic.db.Create(&myItem)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "添加物品失败: "+result.Error.Error())
		return
	}

	// 加载关联的物品信息
	mic.loadItemDetails(&myItem)

	utils.SuccessResponse(c, myItem)
}

// GetMyItems 查询我的物品
func (mic *MyItemController) GetMyItems(c *gin.Context) {
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	position := c.Query("position") // backpack, warehouse, equipped

	// 构建查询条件
	query := mic.db.Where("user_id = ?", userID)
	if position != "" {
		query = query.Where("position = ?", position)
	}

	var myItems []models.MyItem
	result := query.Find(&myItems)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+result.Error.Error())
		return
	}

	// 加载每个物品的详细信息
	for i := range myItems {
		mic.loadItemDetails(&myItems[i])
	}

	utils.SuccessResponse(c, myItems)
}

// loadItemDetails 加载物品详细信息
func (mic *MyItemController) loadItemDetails(myItem *models.MyItem) {
	switch myItem.ItemType {
	case "treasure":
		var treasure models.Treasure
		if err := mic.db.First(&treasure, myItem.ItemID).Error; err == nil {
			myItem.Treasure = &treasure
		}
	case "equipment":
		// 这里可以加载装备信息
		// var equipment models.Equipment
		// if err := mic.db.First(&equipment, myItem.ItemID).Error; err == nil {
		// 	myItem.Equipment = &equipment
		// }
	}
}

// UpdateMyItem 更新我的物品
func (mic *MyItemController) UpdateMyItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var request struct {
		SellPrice int    `json:"sell_price" binding:"min=0"`
		Position  string `json:"position" binding:"oneof=backpack warehouse equipped"`
		Quantity  int    `json:"quantity" binding:"min=1"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	var myItem models.MyItem
	if err := mic.db.First(&myItem, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "物品不存在")
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if request.SellPrice >= 0 {
		updates["sell_price"] = request.SellPrice
	}
	if request.Position != "" {
		updates["position"] = request.Position
	}
	if request.Quantity > 0 {
		updates["quantity"] = request.Quantity
	}

	result := mic.db.Model(&myItem).Updates(updates)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新失败: "+result.Error.Error())
		return
	}

	// 重新加载数据
	mic.db.First(&myItem, id)
	mic.loadItemDetails(&myItem)

	utils.SuccessResponse(c, myItem)
}

// DeleteMyItem 删除我的物品
func (mic *MyItemController) DeleteMyItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	result := mic.db.Delete(&models.MyItem{}, id)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除失败: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "物品不存在")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "删除成功"})
}

// SellMultipleTreasures 批量出售宝物
func (mic *MyItemController) SellMultipleTreasures(c *gin.Context) {
	var request struct {
		UserID    uint   `json:"user_id" binding:"required"`
		MyItemIDs []uint `json:"my_item_ids" binding:"required"` // 我的物品ID列表
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 开始事务
	tx := mic.db.Begin()

	// 1. 查询所有要出售的物品
	var myItems []models.MyItem
	if err := tx.Preload("Treasure").Where("id IN (?) AND user_id = ? AND item_type = ?", request.MyItemIDs, request.UserID, "treasure").Find(&myItems).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询物品失败: "+err.Error())
		return
	}

	if len(myItems) == 0 {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "未找到要出售的宝物")
		return
	}

	// 2. 计算总出售价格
	totalPrice := 0
	var soldItems []gin.H

	for _, item := range myItems {
		sellPrice := item.SellPrice
		if sellPrice == 0 && item.Treasure != nil {
			sellPrice = item.Treasure.Value
		}
		totalPrice += sellPrice

		soldItems = append(soldItems, gin.H{
			"id":         item.ID,
			"item_name":  item.Treasure.Name,
			"item_value": item.Treasure.Value,
			"sold_price": sellPrice,
		})
	}

	// 3. 更新用户金币
	var user models.User
	if err := tx.First(&user, request.UserID).Error; err != nil {
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

	// 4. 删除所有出售的物品
	if err := tx.Where("id IN (?)", request.MyItemIDs).Delete(&models.MyItem{}).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "出售失败: "+err.Error())
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
