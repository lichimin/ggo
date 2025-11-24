package controllers

import (
	"ggo/models"
	"ggo/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserSkinController struct {
	db *gorm.DB
}

func NewUserSkinController(db *gorm.DB) *UserSkinController {
	return &UserSkinController{db: db}
}

// AcquireSkin 用户获得皮肤
func (usc *UserSkinController) AcquireSkin(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	var request struct {
		SkinID uint `json:"skin_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 检查皮肤是否存在
	var skin models.Skin
	if err := usc.db.First(&skin, request.SkinID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "皮肤不存在")
		return
	}

	// 检查用户是否已经拥有该皮肤
	var existingUserSkin models.UserSkin
	result := usc.db.Where("user_id = ? AND skin_id = ?", userID, request.SkinID).First(&existingUserSkin)

	if result.Error == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "已经拥有该皮肤")
		return
	}

	// 创建用户皮肤记录
	userSkin := models.UserSkin{
		UserID:   userID.(uint),
		SkinID:   request.SkinID,
		IsActive: false, // 新获得的皮肤默认不启用
	}

	if err := usc.db.Create(&userSkin).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取皮肤失败: "+err.Error())
		return
	}

	// 加载关联数据返回完整信息
	usc.db.Preload("Skin").First(&userSkin, userSkin.ID)

	utils.SuccessResponse(c, userSkin)
}

// GetUserSkins 获取用户拥有的皮肤列表
func (usc *UserSkinController) GetUserSkins(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	var userSkins []models.UserSkin

	// 查询用户的所有皮肤，并预加载皮肤信息
	result := usc.db.Preload("Skin").Where("user_id = ?", userID).Find(&userSkins)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, userSkins)
}

// ActivateSkin 启用皮肤
func (usc *UserSkinController) ActivateSkin(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	skinID, err := strconv.Atoi(c.Param("skin_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的皮肤ID")
		return
	}

	// 开始事务
	tx := usc.db.Begin()

	// 先禁用用户的所有皮肤
	if err := tx.Model(&models.UserSkin{}).Where("user_id = ?", userID).Update("is_active", false).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "操作失败: "+err.Error())
		return
	}

	// 启用指定的皮肤
	result := tx.Model(&models.UserSkin{}).Where("user_id = ? AND skin_id = ?", userID, skinID).Update("is_active", true)
	if result.Error != nil || result.RowsAffected == 0 {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "皮肤不存在或不属于该用户")
		return
	}

	tx.Commit()

	utils.SuccessResponse(c, gin.H{"message": "皮肤启用成功"})
}

// GetActiveSkin 获取用户当前启用的皮肤
func (usc *UserSkinController) GetActiveSkin(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	var userSkin models.UserSkin

	// 查询用户当前启用的皮肤
	result := usc.db.Preload("Skin").Where("user_id = ?", userID).First(&userSkin)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "未启用任何皮肤")
		return
	}

	utils.SuccessResponse(c, userSkin)
}

// DeleteUserSkin 删除用户皮肤（退还等操作）
func (usc *UserSkinController) DeleteUserSkin(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	skinID, err := strconv.Atoi(c.Param("skin_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的皮肤ID")
		return
	}

	result := usc.db.Where("user_id = ? AND skin_id = ?", userID, skinID).Delete(&models.UserSkin{})
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除失败: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "皮肤不存在或不属于该用户")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "皮肤删除成功"})
}
