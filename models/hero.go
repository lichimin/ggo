package models

type Hero struct {
	ID              uint    `json:"id" gorm:"primarykey"`
	Name            string  `json:"name" gorm:"size:100;not null"`        // 英雄名称
	Images          string  `json:"images" gorm:"type:json"`              // 图片数组（JSON格式）
	AttackType      string  `json:"attack_type" gorm:"size:20;not null"`  // 攻击类型：normal(普通型), penetrate(穿透型), explosive(爆炸型), rebound(反弹型)
	Rarity          int     `json:"rarity" gorm:"default:1"`              // 稀有度：1-6
	DrawProbability float64 `json:"draw_probability" gorm:"default:0.01"` // 抽取概率
	Attack          int     `json:"attack" gorm:"default:0"`              // 攻击力
	CriticalRate    float64 `json:"critical_rate" gorm:"default:0.0"`     // 暴击率
	AttackSpeed     float64 `json:"attack_speed" gorm:"default:1.0"`      // 攻击速度
	Description     string  `json:"description" gorm:"size:500"`          // 描述
	IsActive        bool    `json:"is_active" gorm:"default:true"`        // 是否激活
	CreatedAt       int64   `json:"created_at" gorm:"autoCreateTime"`     // 创建时间
	UpdatedAt       int64   `json:"updated_at" gorm:"autoUpdateTime"`     // 更新时间
}
