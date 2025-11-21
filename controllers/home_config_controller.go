package controllers

import (
	"ggo/models"
	"ggo/utils"
	"net/http"
	"strconv" // 添加这行

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HomeConfigController struct {
	db *gorm.DB
}

func NewHomeConfigController(db *gorm.DB) *HomeConfigController {
	return &HomeConfigController{db: db}
}

// GetHomeConfigs 获取首页配置
func (hcc *HomeConfigController) GetHomeConfigs(c *gin.Context) {
	var homeConfigs []models.HomeConfig

	// 查询参数
	configType := c.Query("type")    // background, button
	position := c.Query("position")  // left_sidebar, right_sidebar, bottom_tab
	isActive := c.Query("is_active") // true, false

	query := hcc.db.Model(&models.HomeConfig{})

	if configType != "" {
		query = query.Where("type = ?", configType)
	}

	if position != "" {
		query = query.Where("position = ?", position)
	}

	if isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	} else {
		// 默认只查询启用的配置
		query = query.Where("is_active = ?", true)
	}

	// 按排序顺序和创建时间排序
	query = query.Order("sort_order ASC, created_at ASC")

	result := query.Find(&homeConfigs)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, homeConfigs)
}
