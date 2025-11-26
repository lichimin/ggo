package models

type UserHero struct {
	ID            uint    `json:"id" gorm:"primarykey"`
	UserID        uint    `json:"user_id" gorm:"not null;index"`     // 用户ID
	HeroID        uint    `json:"hero_id" gorm:"not null;index"`     // 英雄ID
	IsActive      bool    `json:"is_active" gorm:"default:false"`    // 是否启用
	AwakenLevel   int     `json:"awaken_level" gorm:"default:0"`     // 觉醒等级（0-5）
	AttackBonus   float64 `json:"attack_bonus" gorm:"default:0.0"`   // 攻击力加成（百分比）
	CriticalBonus float64 `json:"critical_bonus" gorm:"default:0.0"` // 暴击率加成（百分比）
	SpeedBonus    float64 `json:"speed_bonus" gorm:"default:0.0"`    // 攻击速度加成（百分比）
	CreatedAt     int64   `json:"created_at" gorm:"autoCreateTime"`  // 创建时间
	UpdatedAt     int64   `json:"updated_at" gorm:"autoUpdateTime"`  // 更新时间

	// 关联信息
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Hero Hero `json:"hero,omitempty" gorm:"foreignKey:HeroID"`
}
