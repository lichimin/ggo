package controllers

import (
	"ggo/models"
	"ggo/services"
	"ggo/utils"
	"net/http"
	"strconv"

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
