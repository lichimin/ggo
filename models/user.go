package models

import (
	"time"
)

type User struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	Img       string    `json:"img" gorm:"size:255;not null"`                 // 账号
	Username  string    `json:"username" gorm:"size:50;uniqueIndex;not null"` // 账号
	Password  string    `json:"-" gorm:"size:255;not null"`                   // 密码（不序列化到JSON）
	Gold      int       `json:"gold" gorm:"default:0"`                        // 金币
	Diamond   int       `json:"diamond" gorm:"default:0"`                     // 钻石
	Level     int       `json:"level" gorm:"default:1"`                       // 等级
	LastLogin time.Time `json:"last_login"`                                   // 最后登录时间
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserLoginRequest 登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required,min=1,max=50"`
	Password string `json:"password" binding:"required,min=3,max=50"`
	IsToken  int    `json:"is_token" binding:"omitempty,min=0,max=1"` // 0: 返回完整信息，1: 只返回token
}

// UserLoginResponse 登录响应
type UserLoginResponse struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Img      string `json:"img"`
	Gold     int    `json:"gold"`
	Diamond  int    `json:"diamond"`
	Level    int    `json:"level"`
	Token    string `json:"token"`
}
