package models

type Monster struct {
	ID             uint    `json:"id" gorm:"primarykey"`
	Name           string  `json:"name" gorm:"size:100;not null"`      // 怪物名称
	Attack         int     `json:"attack" gorm:"default:0"`            // 攻击力
	Speed          int     `json:"speed" gorm:"default:0"`             // 移动速度
	AtkSpeed       int     `json:"atk_speed" gorm:"default:0"`         // 攻击速度
	HP             int     `json:"hp" gorm:"default:0"`                // 血量
	CriticalRate   float64 `json:"critical_rate" gorm:"default:0.0"`   // 暴击率 (0-1)
	CriticalDamage float64 `json:"critical_damage" gorm:"default:1.5"` // 暴击伤害 (倍数)
	Drain          float64 `json:"drain" gorm:"default:0.0"`           // 吸血 (0-1)
	SpawnRate      float64 `json:"spawn_rate" gorm:"default:0.1"`      // 出现概率 (0-1)
	SpawnLocation  string  `json:"spawn_location" gorm:"size:200"`     // 出现地点
	DodgeRate      float64 `json:"dodge_rate" gorm:"default:0.0"`      // 闪避率 (0-1)
	InstantKill    float64 `json:"instant_kill" gorm:"default:0.0"`    // 秒杀率 (0-1)
	HealEffect     float64 `json:"heal_effect" gorm:"default:1.0"`     // 回复效果 (倍数)
	ImageURL       string  `json:"image_url" gorm:"size:500"`          // 图片地址
	BackgroundURL  string  `json:"background_url" gorm:"size:500"`     // 背景图片地址
	SkillID        int     `json:"skill_id" gorm:"default:0"`          // 技能ID
	BulletID       int     `json:"bullet_id" gorm:"default:0"`         // 子弹ID
	Level          int     `json:"level" gorm:"default:1"`             // 怪物等级
	ExpReward      int     `json:"exp_reward" gorm:"default:10"`       // 经验奖励
	GoldReward     int     `json:"gold_reward" gorm:"default:5"`       // 金币奖励
	IsActive       bool    `json:"is_active" gorm:"default:true"`      // 是否激活
	CreatedAt      int64   `json:"created_at" gorm:"autoCreateTime"`   // 创建时间
	UpdatedAt      int64   `json:"updated_at" gorm:"autoUpdateTime"`   // 更新时间
}
