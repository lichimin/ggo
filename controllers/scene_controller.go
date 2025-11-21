package controllers

import (
	"ggo/models"
	"ggo/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SceneController struct {
	db *gorm.DB
}

func NewSceneController(db *gorm.DB) *SceneController {
	return &SceneController{db: db}
}

// GetScenes 获取场景列表
func (sc *SceneController) GetScenes(c *gin.Context) {
	var scenes []models.Scene

	// 查询参数
	name := c.Query("name")
	region := c.Query("region")
	isActive := c.Query("is_active")
	level := c.Query("level")

	query := sc.db.Model(&models.Scene{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if level != "" {
		levelInt, _ := strconv.Atoi(level)
		query = query.Where("level <=  ?", levelInt)
	}

	if region != "" {
		query = query.Where("region = ?", region)
	}

	if isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	// 按出现概率降序排列
	query = query.Order("spawn_rate DESC")

	result := query.Find(&scenes)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, scenes)
}
