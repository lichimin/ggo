package controllers

import (
	"ggo/models"
	"ggo/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SkillController struct {
	db *gorm.DB
}

func NewSkillController(db *gorm.DB) *SkillController {
	return &SkillController{db: db}
}

// CreateSkill 创建技能
func (sc *SkillController) CreateSkill(c *gin.Context) {
	var skill models.Skill
	if err := c.ShouldBindJSON(&skill); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	result := sc.db.Create(&skill)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "创建失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, skill)
}

// GetSkill 获取技能详情
func (sc *SkillController) GetSkill(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var skill models.Skill
	result := sc.db.First(&skill, id)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "技能不存在")
		return
	}

	utils.SuccessResponse(c, skill)
}

// GetSkills 获取技能列表
func (sc *SkillController) GetSkills(c *gin.Context) {
	var skills []models.Skill

	// 查询参数
	name := c.Query("name")
	skillType := c.Query("skill_type")
	targetType := c.Query("target_type")
	isActive := c.Query("is_active")

	query := sc.db.Model(&models.Skill{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if skillType != "" {
		query = query.Where("skill_type = ?", skillType)
	}

	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}

	if isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	result := query.Find(&skills)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, skills)
}

// UpdateSkill 更新技能
func (sc *SkillController) UpdateSkill(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var skill models.Skill
	if err := c.ShouldBindJSON(&skill); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	skill.ID = uint(id)

	result := sc.db.Save(&skill)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, skill)
}

// DeleteSkill 删除技能
func (sc *SkillController) DeleteSkill(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的ID")
		return
	}

	result := sc.db.Delete(&models.Skill{}, id)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "删除失败: "+result.Error.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "删除成功"})
}
