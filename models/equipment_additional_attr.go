package models

type EquipmentAdditionalAttr struct {
	ID              uint   `json:"id" gorm:"primarykey"`
	UserEquipmentID uint   `json:"user_equipment_id" gorm:"not null;index"` // 玩家装备ID
	AttrType        string `json:"attr_type" gorm:"size:20;not null"`       // 属性类型：hp, attack, attack_speed, move_speed, bullet_speed, drain, critical, dodge, instant_kill, recovery, trajectory, enhance
	AttrValue       string `json:"attr_value" gorm:"size:50;not null"`      // 属性值（字符串存储，可能是数字或描述）
	CreatedAt       int64  `json:"created_at" gorm:"autoCreateTime"`        // 创建时间
	UpdatedAt       int64  `json:"updated_at" gorm:"autoUpdateTime"`        // 更新时间
}
