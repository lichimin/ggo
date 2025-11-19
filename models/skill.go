package models

type Skill struct {
	ID          uint    `json:"id" gorm:"primarykey"`
	Name        string  `json:"name" gorm:"size:100;not null"`              // 技能名称
	ImageURL    string  `json:"image_url" gorm:"size:500"`                  // 技能图片
	HealHP      int     `json:"heal_hp" gorm:"default:0"`                   // 回复生命值
	AddAttack   int     `json:"add_attack" gorm:"default:0"`                // 增加攻击力
	AddCritical float64 `json:"add_critical" gorm:"default:0.0"`            // 增加暴击率
	Damage      int     `json:"damage" gorm:"default:0"`                    // 造成伤害
	Cooldown    int     `json:"cooldown" gorm:"default:5"`                  // 冷却时间（秒）
	Duration    int     `json:"duration" gorm:"default:0"`                  // 效果持续时间（秒）
	ManaCost    int     `json:"mana_cost" gorm:"default:10"`                // 魔法消耗
	SkillType   string  `json:"skill_type" gorm:"size:50;default:'active'"` // 技能类型：active(主动), passive(被动)
	TargetType  string  `json:"target_type" gorm:"size:50;default:'enemy'"` // 目标类型：self(自己), enemy(敌人)
	IsActive    bool    `json:"is_active" gorm:"default:true"`              // 是否激活
	Description string  `json:"description" gorm:"size:500"`                // 描述
	CreatedAt   int64   `json:"created_at" gorm:"autoCreateTime"`           // 创建时间
	UpdatedAt   int64   `json:"updated_at" gorm:"autoUpdateTime"`           // 更新时间
}
