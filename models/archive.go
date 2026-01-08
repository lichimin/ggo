package models

import (
	"gorm.io/gorm"
)

// Archive 存档模型
type Archive struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	UserID    uint           `json:"user_id" gorm:"not null;index:idx_user_id,unique"` // 用户ID，唯一索引
	JSONData  interface{}    `json:"json_data" gorm:"type:jsonb;not null"`             // 存档数据，直接存储JSON对象
	V         int            `json:"v" gorm:"not null;default:0"`                      // 版本号，用于控制存档的更新顺序
	Area      int            `json:"area" gorm:"not null;default:1"`                   // 区服ID
	CreatedAt int64          `json:"created_at" gorm:"autoCreateTime"`                 // 创建时间
	UpdatedAt int64          `json:"updated_at" gorm:"autoUpdateTime"`                 // 更新时间
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`                                   // 软删除
}

// TableName 指定表名
func (Archive) TableName() string {
	return "archives"
}
