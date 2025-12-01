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

	// 简化的请求参数
	var request struct {
		ItemID   uint `json:"item_id" binding:"required"`
		Quantity int  `json:"quantity" binding:"min=1"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
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

	// 验证宝物是否存在
	var treasure models.Treasure
	if err := mic.db.First(&treasure, request.ItemID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "宝物不存在")
		return
	}
	itemName := treasure.Name

	// 创建我的物品记录
	myItem := models.MyItem{
		UserID:    userID.(uint),
		ItemID:    request.ItemID,
		ItemType:  itemType,
		SellPrice: sellPrice,
		Position:  position,
		Quantity:  request.Quantity,
	}

	if err := mic.db.Create(&myItem).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "添加物品失败: "+err.Error())
		return
	}

	// 创建响应结构，包含物品详细信息
	type MyItemResponse struct {
		models.MyItem
		ItemName string `json:"item_name"`
	}

	response := MyItemResponse{
		MyItem:   myItem,
		ItemName: itemName,
	}

	utils.SuccessResponse(c, response)
}

// 修复 GetMyItems 方法中的 loadItemDetails
func (mic *MyItemController) GetMyItems(c *gin.Context) {
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	position := c.Query("position")

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

	// 创建响应结构，包含物品详细信息
	type MyItemResponse struct {
		models.MyItem
		Treasure *models.Treasure `json:"treasure,omitempty"`
	}

	var responseItems []MyItemResponse

	// 批量查询宝物信息
	if len(myItems) > 0 {
		var treasureIDs []uint
		for _, item := range myItems {
			if item.ItemType == "treasure" {
				treasureIDs = append(treasureIDs, item.ItemID)
			}
		}

		var treasures []models.Treasure
		if err := mic.db.Where("id IN (?)", treasureIDs).Find(&treasures).Error; err == nil {
			treasureMap := make(map[uint]models.Treasure)
			for _, treasure := range treasures {
				treasureMap[treasure.ID] = treasure
			}

			for _, item := range myItems {
				responseItem := MyItemResponse{MyItem: item}
				if treasure, exists := treasureMap[item.ItemID]; exists {
					responseItem.Treasure = &treasure
				}
				responseItems = append(responseItems, responseItem)
			}
		}
	}

	utils.SuccessResponse(c, responseItems)
}

// SellMultipleTreasures 批量出售宝物
func (mic *MyItemController) SellMultipleTreasures(c *gin.Context) {
	var request struct {
		UserID    uint   `json:"user_id" binding:"required"`
		MyItemIDs []uint `json:"my_item_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 开始事务
	tx := mic.db.Begin()

	// 1. 查询所有要出售的物品
	var myItems []models.MyItem
	if err := tx.Where("id IN (?) AND user_id = ? AND item_type = ?", request.MyItemIDs, request.UserID, "treasure").Find(&myItems).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询物品失败: "+err.Error())
		return
	}

	if len(myItems) == 0 {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "未找到要出售的宝物")
		return
	}

	// 2. 计算总出售价格和收集宝物ID
	totalPrice := 0
	var soldItems []gin.H
	var treasureIDs []uint

	for _, item := range myItems {
		treasureIDs = append(treasureIDs, item.ItemID)
	}

	// 3. 批量查询宝物信息
	var treasures []models.Treasure
	if err := tx.Where("id IN (?)", treasureIDs).Find(&treasures).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询宝物信息失败: "+err.Error())
		return
	}

	// 4. 创建宝物映射表
	treasureMap := make(map[uint]models.Treasure)
	for _, treasure := range treasures {
		treasureMap[treasure.ID] = treasure
	}

	// 5. 计算总价格
	for _, item := range myItems {
		treasure, exists := treasureMap[item.ItemID]
		if !exists {
			continue
		}

		sellPrice := item.SellPrice
		if sellPrice == 0 {
			sellPrice = treasure.Value
		}
		totalPrice += sellPrice

		soldItems = append(soldItems, gin.H{
			"id":         item.ID,
			"item_name":  treasure.Name,
			"item_value": treasure.Value,
			"sold_price": sellPrice,
		})
	}

	// 6. 更新用户金币
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

	// 7. 删除所有出售的物品
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
