package controllers

import (
	"ggo/models"
	"ggo/services"
	"ggo/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{
		userService: services.NewUserService(db),
	}
}

// Login 用户登录/注册
func (uc *UserController) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	response, newToken, err := uc.userService.LoginOrRegister(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 使用带token的响应
	utils.SuccessResponseWithToken(c, response, newToken)
}

// GetProfile 获取用户信息
func (uc *UserController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	user, err := uc.userService.GetUserByID(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 不返回密码
	user.Password = ""
	utils.SuccessResponse(c, user)
}

// GetUsers 获取用户列表（管理员功能）
func (uc *UserController) GetUsers(c *gin.Context) {
	// 这里可以添加管理员权限检查
	// if !isAdmin(c) { ... }

	var users []models.User
	result := uc.userService.DB.Find(&users)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, result.Error.Error())
		return
	}

	// 移除密码字段
	for i := range users {
		users[i].Password = ""
	}

	utils.SuccessResponse(c, users)
}

// GetUser 获取指定用户信息
func (uc *UserController) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	user, err := uc.userService.GetUserByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 不返回密码
	user.Password = ""
	utils.SuccessResponse(c, user)
}

// CreateUser 创建用户（管理员功能）
func (uc *UserController) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "密码加密失败")
		return
	}
	user.Password = hashedPassword

	result := uc.userService.DB.Create(&user)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, result.Error.Error())
		return
	}

	// 不返回密码
	user.Password = ""
	utils.SuccessResponse(c, user)
}

// UpdateUser 更新用户信息
func (uc *UserController) UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	// 检查权限：只能更新自己的信息，除非是管理员
	userID, _ := c.Get("userID")
	if uint(id) != userID.(uint) {
		// 这里可以添加管理员权限检查
		utils.ErrorResponse(c, http.StatusForbidden, "无权修改其他用户信息")
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	user.ID = uint(id)

	// 如果提供了新密码，需要加密
	if user.Password != "" {
		hashedPassword, err := utils.HashPassword(user.Password)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "密码加密失败")
			return
		}
		user.Password = hashedPassword
	}

	result := uc.userService.DB.Save(&user)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, result.Error.Error())
		return
	}

	// 不返回密码
	user.Password = ""
	utils.SuccessResponse(c, user)
}

// DeleteUser 删除用户（管理员功能）
func (uc *UserController) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	// 这里应该检查管理员权限
	// if !isAdmin(c) { ... }

	result := uc.userService.DB.Delete(&models.User{}, id)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, result.Error.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "用户删除成功"})
}

// GetPlayerAttributes 获取玩家属性（计算已穿戴装备和皮肤属性总和）
func (uc *UserController) GetPlayerAttributes(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 定义属性汇总结构
	attributes := gin.H{
		"hp":               0,
		"attack":           0,
		"attack_speed":     1.0,
		"move_speed":       0,
		"bullet_speed":     0,
		"drain":            0,
		"critical":         0,
		"dodge":            0,
		"instant_kill":     0,
		"recovery":         0,
		"trajectory":       0,
		"critical_rate":    0.0,
		"critical_damage":  1.5,
		"atk_type":         0,
		"damage_reduction": "", // 减伤属性
	}

	// 查询用户已穿戴的装备
	var equippedItems []models.UserEquipment
	result := uc.userService.DB.Where("user_id = ? AND is_equipped = ?", userID.(uint), true).
		Preload("EquipmentTemplate").
		Preload("AdditionalAttrs").
		Find(&equippedItems)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "查询装备失败: "+result.Error.Error())
		return
	}

	// 计算装备属性总和
	for _, item := range equippedItems {
		// 基础属性
		attributes["hp"] = attributes["hp"].(int) + item.EquipmentTemplate.HP
		attributes["attack"] = attributes["attack"].(int) + item.EquipmentTemplate.Attack
		attributes["attack_speed"] = attributes["attack_speed"].(float64) + item.EquipmentTemplate.AttackSpeed
		attributes["move_speed"] = attributes["move_speed"].(int) + item.EquipmentTemplate.MoveSpeed
		attributes["bullet_speed"] = attributes["bullet_speed"].(int) + item.EquipmentTemplate.BulletSpeed
		attributes["drain"] = attributes["drain"].(int) + item.EquipmentTemplate.Drain
		attributes["critical"] = attributes["critical"].(int) + item.EquipmentTemplate.Critical
		attributes["dodge"] = attributes["dodge"].(int) + item.EquipmentTemplate.Dodge
		attributes["instant_kill"] = attributes["instant_kill"].(int) + item.EquipmentTemplate.InstantKill
		attributes["recovery"] = attributes["recovery"].(int) + item.EquipmentTemplate.Recovery
		attributes["trajectory"] = attributes["trajectory"].(int) + item.EquipmentTemplate.Trajectory

		// 附加属性处理
		for _, attr := range item.AdditionalAttrs {
			// 处理enhance类型的特殊稀有属性
			if attr.AttrType == "enhance" {
				// 清理属性值，去除百分号和其他非数字字符
				cleanValue := attr.AttrValue
				if strings.Contains(cleanValue, "%") {
					cleanValue = strings.ReplaceAll(cleanValue, "%", "")
				}
				if strings.Contains(cleanValue, "秒杀") {
					cleanValue = strings.ReplaceAll(cleanValue, "秒杀", "")
				}

				// 根据AttrName判断属性类型并累加
				switch attr.AttrName {
				case "暴食": // 增加攻速
					if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
						// 转换为百分比数值
						speedVal := attributes["attack_speed"].(float64) + (val / 100)
						attributes["attack_speed"] = speedVal
					}
				case "贪婪": // 增加秒杀几率
					if val, err := strconv.Atoi(cleanValue); err == nil {
						killVal := attributes["instant_kill"].(int) + val
						attributes["instant_kill"] = killVal
					}
				case "傲慢": // 增加最大HP
					if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
						// 转换为百分比数值，基于当前HP值增加
						hpVal := attributes["hp"].(int)
						attributes["hp"] = hpVal + int(float64(hpVal)*val/100)
					}
				case "嫉妒": // 增加暴击伤害
					if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
						// 转换为百分比数值
						damageVal := attributes["critical_damage"].(float64) + (val / 100)
						attributes["critical_damage"] = damageVal
					}
				case "色欲": // 自动回复
					if val, err := strconv.Atoi(cleanValue); err == nil {
						recoveryVal := attributes["recovery"].(int) + val
						attributes["recovery"] = recoveryVal
					}
				case "暴怒": // 增加暴击率
					if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
						// 转换为百分比数值
						rateVal := attributes["critical_rate"].(float64) + (val / 100)
						attributes["critical_rate"] = rateVal
					}
				case "怠惰": // 增加攻击力
					if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
						// 转换为百分比数值，基于当前攻击力增加
						attackVal := attributes["attack"].(int)
						attributes["attack"] = attackVal + int(float64(attackVal)*val/100)
					}
				}
			} else {
				// 处理普通附加属性
				switch attr.AttrType {
				case "hp":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						hpVal := attributes["hp"].(int) + val
						attributes["hp"] = hpVal
					}
				case "attack":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						attackVal := attributes["attack"].(int) + val
						attributes["attack"] = attackVal
					}
				case "attack_speed":
					if val, err := strconv.ParseFloat(attr.AttrValue, 64); err == nil {
						attackSpeedVal := attributes["attack_speed"].(float64) + val
						attributes["attack_speed"] = attackSpeedVal
					}
				case "move_speed":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						moveSpeedVal := attributes["move_speed"].(int) + val
						attributes["move_speed"] = moveSpeedVal
					}
				case "bullet_speed":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						bulletSpeedVal := attributes["bullet_speed"].(int) + val
						attributes["bullet_speed"] = bulletSpeedVal
					}
				case "drain":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						drainVal := attributes["drain"].(int) + val
						attributes["drain"] = drainVal
					}
				case "critical":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						criticalVal := attributes["critical"].(int) + val
						attributes["critical"] = criticalVal
					}
				case "dodge":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						dodgeVal := attributes["dodge"].(int) + val
						attributes["dodge"] = dodgeVal
					}
				case "instant_kill":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						instantKillVal := attributes["instant_kill"].(int) + val
						attributes["instant_kill"] = instantKillVal
					}
				case "recovery":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						recoveryVal := attributes["recovery"].(int) + val
						attributes["recovery"] = recoveryVal
					}
				case "trajectory":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						trajectoryVal := attributes["trajectory"].(int) + val
						attributes["trajectory"] = trajectoryVal
					}
				case "damage_reduction":
					// 减伤属性特殊处理，暂时存储为字符串
					if _, exists := attributes["damage_reduction"]; exists {
						// 如果已经存在减伤属性，将两个字符串合并
						currentDR := attributes["damage_reduction"].(string)
						attributes["damage_reduction"] = currentDR + " + " + attr.AttrValue
					} else {
						attributes["damage_reduction"] = attr.AttrValue
					}
				}
			}
		}
	}

	// 查询用户已激活的皮肤
	var activeSkin models.UserSkin
	result = uc.userService.DB.Where("user_id = ? AND is_active = ?", userID.(uint), true).
		Preload("Skin").
		First(&activeSkin)
	if result.Error == nil {
		// 计算皮肤属性
		attributes["hp"] = attributes["hp"].(int) + activeSkin.Skin.HP
		attributes["attack"] = attributes["attack"].(int) + activeSkin.Skin.Attack
		attributes["attack_speed"] = attributes["attack_speed"].(float64) + float64(activeSkin.Skin.AtkSpeed)
		attributes["critical_rate"] = attributes["critical_rate"].(float64) + activeSkin.Skin.CriticalRate
		attributes["critical_damage"] = attributes["critical_damage"].(float64) + activeSkin.Skin.CriticalDamage
		if activeSkin.Skin.AtkType > 0 {
			attributes["atk_type"] = activeSkin.Skin.AtkType
		}
	}

	utils.SuccessResponse(c, attributes)
}
