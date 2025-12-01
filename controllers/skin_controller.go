package controllers

import (
	"ggo/models"
	"ggo/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SkinController struct {
	db *gorm.DB
}

func NewSkinController(db *gorm.DB) *SkinController {
	return &SkinController{db: db}
}

// GetSkin 获取皮肤详情
func (sc *SkinController) GetSkin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var skin models.Skin
	result := sc.db.First(&skin, id)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "皮肤不存在")
		return
	}

	utils.SuccessResponse(c, skin)
}

// GetSkins 获取皮肤列表
func (sc *SkinController) GetSkins(c *gin.Context) {
	var skins []models.Skin

	// 查询参数
	name := c.Query("name")

	query := sc.db.Model(&models.Skin{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	result := query.Find(&skins)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, skins)
}
