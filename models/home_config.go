package models

type HomeConfig struct {
	ID         uint    `json:"id" gorm:"primarykey"`
	Type       string  `json:"type" gorm:"size:20;not null"`     // 类型：background(首页背景), button(按钮)
	ImageURL   string  `json:"image_url" gorm:"size:500"`        // 图片地址
	Scale      float64 `json:"scale" gorm:"default:1.0"`         // 缩放大小
	ButtonName string  `json:"button_name" gorm:"size:50"`       // 按钮名称
	IsActive   bool    `json:"is_active" gorm:"default:true"`    // 是否启用
	Position   string  `json:"position" gorm:"size:20"`          // 位置：left_sidebar(左侧按钮栏), right_sidebar(右侧按钮栏), bottom_tab(下方tab栏)
	SortOrder  int     `json:"sort_order" gorm:"default:0"`      // 排序顺序
	CreatedAt  int64   `json:"created_at" gorm:"autoCreateTime"` // 创建时间
	UpdatedAt  int64   `json:"updated_at" gorm:"autoUpdateTime"` // 更新时间
}
