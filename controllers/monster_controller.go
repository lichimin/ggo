package controllers

import (
	"ggo/models"
	"ggo/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MonsterController struct {
	db *gorm.DB
}

func NewMonsterController(db *gorm.DB) *MonsterController {
	return &MonsterController{db: db}
}

// CreateMonster 创建怪物
func (mc *MonsterController) CreateMonster(c *gin.Context) {
	var monster models.Monster
	if err := c.ShouldBindJSON(&monster); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	result := mc.db.Create(&monster)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, monster)
}

// GetMonster 获取怪物详情
func (mc *MonsterController) GetMonster(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var monster models.Monster
	result := mc.db.First(&monster, id)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "怪物不存在")
		return
	}

	utils.SuccessResponse(c, monster)
}

// GetMonsters 获取怪物列表
func (mc *MonsterController) GetMonsters(c *gin.Context) {
	var monsters []models.Monster

	// 查询参数
	name := c.Query("name")
	level := c.Query("level")
	location := c.Query("location")
	isActive := c.Query("is_active")

	query := mc.db.Model(&models.Monster{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if level != "" {
		levelInt, _ := strconv.Atoi(level)
		query = query.Where("level <=  ?", levelInt)
	}

	if location != "" {
		query = query.Where("spawn_location LIKE ?", "%"+location+"%")
	}

	if isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	result := query.Find(&monsters)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, monsters)
}

// UpdateMonster 更新怪物
func (mc *MonsterController) UpdateMonster(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var monster models.Monster
	if err := c.ShouldBindJSON(&monster); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	monster.ID = uint(id)

	result := mc.db.Save(&monster)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, monster)
}

// DeleteMonster 删除怪物
func (mc *MonsterController) DeleteMonster(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	result := mc.db.Delete(&models.Monster{}, id)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "删除成功"})
}
