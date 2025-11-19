package models

type Treasure struct {
	ID          uint   `json:"id" gorm:"primarykey"`
	Name        string `json:"name" gorm:"size:100;not null"`    // 宝物名称
	ImageURL    string `json:"image_url" gorm:"size:500"`        // 宝物图片
	Value       int    `json:"value" gorm:"default:0"`           // 价值（金币）
	Level       int    `json:"level" gorm:"default:1"`           // 等级
	IsActive    bool   `json:"is_active" gorm:"default:true"`    // 是否激活
	Description string `json:"description" gorm:"size:500"`      // 描述（可选）
	CreatedAt   int64  `json:"created_at" gorm:"autoCreateTime"` // 创建时间
	UpdatedAt   int64  `json:"updated_at" gorm:"autoUpdateTime"` // 更新时间
}
