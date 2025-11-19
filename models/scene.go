package models

type Scene struct {
	ID          uint    `json:"id" gorm:"primarykey"`
	Name        string  `json:"name" gorm:"size:100;not null"`    // 场景名称
	ImageURL    string  `json:"image_url" gorm:"size:500"`        // 场景图片
	SpawnRate   float64 `json:"spawn_rate" gorm:"default:0.1"`    // 出现概率 (0-1)
	Region      string  `json:"region" gorm:"size:100"`           // 所属区域
	IsActive    bool    `json:"is_active" gorm:"default:true"`    // 是否激活
	Description string  `json:"description" gorm:"size:500"`      // 描述（可选）
	CreatedAt   int64   `json:"created_at" gorm:"autoCreateTime"` // 创建时间
	UpdatedAt   int64   `json:"updated_at" gorm:"autoUpdateTime"` // 更新时间
}
