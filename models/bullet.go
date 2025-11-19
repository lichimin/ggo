package models

type Bullet struct {
	ID          uint    `json:"id" gorm:"primarykey"`
	Name        string  `json:"name" gorm:"size:100;not null"`    // 子弹名称
	ImageURL    string  `json:"image_url" gorm:"size:500"`        // 子弹图片
	Size        float64 `json:"size" gorm:"default:1.0"`          // 子弹大小（倍数）
	Distance    float64 `json:"distance" gorm:"default:1.0"`      // 子弹飞行距离（倍数）
	Penetrate   int     `json:"penetrate" gorm:"default:1"`       // 穿透数量
	AttackCount int     `json:"attack_count" gorm:"default:1"`    // 攻击数量
	Speed       float64 `json:"speed" gorm:"default:1.0"`         // 子弹速度（倍数）
	Damage      int     `json:"damage" gorm:"default:10"`         // 基础伤害
	IsActive    bool    `json:"is_active" gorm:"default:true"`    // 是否激活
	Description string  `json:"description" gorm:"size:500"`      // 描述
	CreatedAt   int64   `json:"created_at" gorm:"autoCreateTime"` // 创建时间
	UpdatedAt   int64   `json:"updated_at" gorm:"autoUpdateTime"` // 更新时间
}
