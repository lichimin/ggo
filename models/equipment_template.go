package models

type EquipmentTemplate struct {
	ID          uint    `json:"id" gorm:"primarykey"`
	Name        string  `json:"name" gorm:"size:100;not null"`    // 装备名称
	Level       int     `json:"level" gorm:"default:1"`           // 品级：1-普通, 2-稀有, 3-史诗, 4-传说, 5-神话, 6-创世
	Slot        string  `json:"slot" gorm:"size:20;not null"`     // 部位：weapon(武器), helmet(防具-头), chest(防具-胸), gloves(防具-护手), pants(防具-护腿), boots(防具-鞋子)
	HP          int     `json:"hp" gorm:"default:0"`              // 生命值
	Attack      int     `json:"attack" gorm:"default:0"`          // 攻击力
	AttackSpeed float64 `json:"attack_speed" gorm:"default:1.0"`  // 攻速
	MoveSpeed   int     `json:"move_speed" gorm:"default:0"`      // 移速
	BulletSpeed int     `json:"bullet_speed" gorm:"default:0"`    // 弹速
	Drain       int     `json:"drain" gorm:"default:0"`           // 吸血
	Critical    int     `json:"critical" gorm:"default:0"`        // 暴击
	Dodge       int     `json:"dodge" gorm:"default:0"`           // 闪避
	InstantKill int     `json:"instant_kill" gorm:"default:0"`    // 秒杀
	Recovery    int     `json:"recovery" gorm:"default:0"`        // 恢复
	Trajectory  int     `json:"trajectory" gorm:"default:0"`      // 弹道
	ImageURL    string  `json:"image_url" gorm:"size:500"`        // 装备图片
	Description string  `json:"description" gorm:"size:500"`      // 描述
	IsActive    bool    `json:"is_active" gorm:"default:true"`    // 是否激活
	CreatedAt   int64   `json:"created_at" gorm:"autoCreateTime"` // 创建时间
	UpdatedAt   int64   `json:"updated_at" gorm:"autoUpdateTime"` // 更新时间
}
