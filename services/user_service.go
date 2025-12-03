package services

import (
	"errors"
	"ggo/models"
	"ggo/utils"
	"time"

	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{DB: db}
}

// GetDB 获取数据库实例（供控制器使用）
func (s *UserService) GetDB() *gorm.DB {
	return s.DB
}

// LoginOrRegister 登录或注册
func (s *UserService) LoginOrRegister(req *models.UserLoginRequest) (*models.UserLoginResponse, string, error) {
	var user models.User

	// 查找用户
	err := s.DB.Where("username = ?", req.Username).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 用户不存在，自动注册
		return s.register(req)
	} else if err != nil {
		return nil, "", err
	}

	// 用户存在，验证密码
	return s.login(&user, req.Password)
}

// register 注册新用户
func (s *UserService) register(req *models.UserLoginRequest) (*models.UserLoginResponse, string, error) {
	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, "", err
	}

	user := models.User{
		Username:  req.Username,
		Password:  hashedPassword,
		Gold:      100, // 新用户默认金币
		Level:     1,   // 新用户默认等级
		LastLogin: time.Now(),
	}

	result := s.DB.Create(&user)
	if result.Error != nil {
		return nil, "", result.Error
	}

	// 新用户默认绑定skinid=1的皮肤
	userSkin := models.UserSkin{
		UserID:   user.ID,
		SkinID:   1,
		IsActive: true, // 默认为激活状态
	}

	skinResult := s.DB.Create(&userSkin)
	if skinResult.Error != nil {
		return nil, "", skinResult.Error
	}

	// 生成token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", err
	}

	response := &models.UserLoginResponse{
		UserID:   user.ID,
		Username: user.Username,
		Gold:     user.Gold,
		Level:    user.Level,
		Token:    token,
	}

	return response, token, nil
}

// login 用户登录
func (s *UserService) login(user *models.User, password string) (*models.UserLoginResponse, string, error) {
	// 验证密码
	if !utils.CheckPassword(password, user.Password) {
		return nil, "", errors.New("密码错误")
	}

	// 更新最后登录时间
	s.DB.Model(user).Update("last_login", time.Now())

	// 生成新token
	newToken, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", err
	}

	response := &models.UserLoginResponse{
		Img:      user.Img,
		UserID:   user.ID,
		Username: user.Username,
		Gold:     user.Gold,
		Level:    user.Level,
		Token:    newToken,
	}

	return response, newToken, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	err := s.DB.First(&user, userID).Error
	return &user, err
}
