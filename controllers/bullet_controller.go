package controllers

import (
	"ggo/models"
	"ggo/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BulletController struct {
	db *gorm.DB
}

func NewBulletController(db *gorm.DB) *BulletController {
	return &BulletController{db: db}
}

// CreateBullet 创建子弹
func (bc *BulletController) CreateBullet(c *gin.Context) {
	var bullet models.Bullet
	if err := c.ShouldBindJSON(&bullet); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	result := bc.db.Create(&bullet)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, bullet)
}

// GetBullet 获取子弹详情
func (bc *BulletController) GetBullet(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var bullet models.Bullet
	result := bc.db.First(&bullet, id)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "子弹不存在")
		return
	}

	utils.SuccessResponse(c, bullet)
}

// GetBullets 获取子弹列表
func (bc *BulletController) GetBullets(c *gin.Context) {
	var bullets []models.Bullet

	// 查询参数
	name := c.Query("name")
	isActive := c.Query("is_active")

	query := bc.db.Model(&models.Bullet{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	result := query.Find(&bullets)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, bullets)
}

// UpdateBullet 更新子弹
func (bc *BulletController) UpdateBullet(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var bullet models.Bullet
	if err := c.ShouldBindJSON(&bullet); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	bullet.ID = uint(id)

	result := bc.db.Save(&bullet)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, bullet)
}

// DeleteBullet 删除子弹
func (bc *BulletController) DeleteBullet(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	result := bc.db.Delete(&models.Bullet{}, id)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "删除成功"})
}
