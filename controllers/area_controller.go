package controllers

import (
	"ggo/models"
	"ggo/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AreaController struct {
	db *gorm.DB
}

func NewAreaController(db *gorm.DB) *AreaController {
	return &AreaController{
		db: db,
	}
}

// GetAreas 获取区服列表
func (ac *AreaController) GetAreas(c *gin.Context) {
	var areas []models.Area
	
	// 查询区服列表，只返回需要的字段
	result := ac.db.Model(&models.Area{}).
		Where("status = ?", 1). // 只返回正常状态的区服
		Order("area ASC").
		Find(&areas)
	
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询区服列表失败: "+result.Error.Error())
		return
	}

	// 转换响应格式，只返回要求的字段
	var response []gin.H
	for _, area := range areas {
		response = append(response, gin.H{
			"id":     area.ID,
			"area":   area.Area,
			"is_new": area.IsNew,
		})
	}

	utils.SuccessResponse(c, gin.H{
		"areas": response,
		"total": len(response),
	})
}