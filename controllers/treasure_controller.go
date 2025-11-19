package controllers

import (
	"myapp/models"
	"myapp/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TreasureController struct {
	db *gorm.DB
}

func NewTreasureController(db *gorm.DB) *TreasureController {
	return &TreasureController{db: db}
}

// GetTreasures 获取宝物列表
func (tc *TreasureController) GetTreasures(c *gin.Context) {
	var treasures []models.Treasure

	// 查询参数
	name := c.Query("name")
	level := c.Query("level")
	isActive := c.Query("is_active")

	query := tc.db.Model(&models.Treasure{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if level != "" {
		levelInt, _ := strconv.Atoi(level)
		levelInt++
		query = query.Where("level <= ?", levelInt)
	}

	if isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	// 按等级和价值排序（等级优先，价值次之）
	query = query.Order("level DESC, value DESC")

	result := query.Find(&treasures)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, treasures)
}

// CreateTreasure 创建宝物（管理员功能）
func (tc *TreasureController) CreateTreasure(c *gin.Context) {
	var treasure models.Treasure
	if err := c.ShouldBindJSON(&treasure); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	result := tc.db.Create(&treasure)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, treasure)
}

// GetTreasure 获取宝物详情
func (tc *TreasureController) GetTreasure(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var treasure models.Treasure
	result := tc.db.First(&treasure, id)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "宝物不存在")
		return
	}

	utils.SuccessResponse(c, treasure)
}

// UpdateTreasure 更新宝物
func (tc *TreasureController) UpdateTreasure(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var treasure models.Treasure
	if err := c.ShouldBindJSON(&treasure); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	treasure.ID = uint(id)

	result := tc.db.Save(&treasure)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, treasure)
}

// DeleteTreasure 删除宝物
func (tc *TreasureController) DeleteTreasure(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	result := tc.db.Delete(&models.Treasure{}, id)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "删除成功"})
}
