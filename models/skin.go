package models

type Skin struct {
	ID              uint     `json:"id" gorm:"primarykey"`
	Name            string   `json:"name" gorm:"size:100;not null"`                      // 皮肤名称
	Attack          int      `json:"attack" gorm:"default:0"`                            // 攻击力
	HP              int      `json:"hp" gorm:"default:0"`                                // 血量
	AtkType         int      `json:"atk_type" gorm:"default:0"`                          // 攻击类型 1穿透  2散射 3反弹
	AtkSpeed        int      `json:"atk_speed" gorm:"default:0"`                         // 攻击速度
	CriticalRate    float64  `json:"critical_rate" gorm:"default:0.0"`                   // 暴击率 (0-1)
	CriticalDamage  float64  `json:"critical_damage" gorm:"default:1.5"`                 // 暴击伤害 (倍数)
	BackgroundURL   string   `json:"background_url" gorm:"size:500"`                     // 背景图片地址
	IdleImageURLs   []string `json:"idle_image_urls" gorm:"type:json;serializer:json"`   // 待机图片（JSON格式多个图片）
	AttackImageURLs []string `json:"attack_image_urls" gorm:"type:json;serializer:json"` // 攻击图片（JSON格式多个图片）
	MoveImageURLs   []string `json:"move_image_urls" gorm:"type:json;serializer:json"`   // 移动图片（JSON格式多个图片）
	CreatedAt       int64    `json:"created_at" gorm:"autoCreateTime"`                   // 创建时间
	UpdatedAt       int64    `json:"updated_at" gorm:"autoUpdateTime"`                   // 更新时间
}
