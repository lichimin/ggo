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

// CreateSkin 创建皮肤
func (sc *SkinController) CreateSkin(c *gin.Context) {
	var skin models.Skin
	if err := c.ShouldBindJSON(&skin); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	result := sc.db.Create(&skin)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, skin)
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
	isActive := c.Query("is_active")

	query := sc.db.Model(&models.Skin{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	result := query.Find(&skins)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, skins)
}

// UpdateSkin 更新皮肤
func (sc *SkinController) UpdateSkin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var skin models.Skin
	if err := c.ShouldBindJSON(&skin); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	skin.ID = uint(id)

	result := sc.db.Save(&skin)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, skin)
}

// DeleteSkin 删除皮肤
func (sc *SkinController) DeleteSkin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	result := sc.db.Delete(&models.Skin{}, id)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "删除成功"})
}
