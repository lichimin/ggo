package models

type EquipmentAdditionalAttr struct {
	ID              uint   `json:"id" gorm:"primarykey"`
	UserEquipmentID uint   `json:"user_equipment_id" gorm:"not null;index"` // 玩家装备ID
	AttrType        string `json:"attr_type" gorm:"size:20;not null"`       // 属性类型：
	// hp - 生命值，增加角色的最大生命值
	// attack - 攻击力，增加角色的基础攻击力
	// attack_speed - 攻击速度，增加角色的攻击频率
	// move_speed - 移动速度，增加角色的移动速度
	// bullet_speed - 子弹速度，增加角色发射子弹的飞行速度
	// drain - 吸血，攻击时吸取敌方生命值的百分比
	// critical - 暴击率，增加攻击暴击的概率
	// dodge - 闪避率，减少被敌方攻击命中的概率
	// instant_kill - 秒杀概率，攻击时有概率直接秒杀敌人
	// recovery - 生命恢复，每秒自动恢复的生命值
	// trajectory - 弹道数，增加角色一次发射的子弹数量
	// enhance - 强化加成，增加装备的总体属性效果
	AttrName  string `json:"attr_name" gorm:"size:20"`           // 属性名称（用于显示稀有属性的名称，如七宗罪）
	AttrValue string `json:"attr_value" gorm:"size:50;not null"` // 属性值（字符串存储，可能是数字或描述）
	CreatedAt int64  `json:"created_at" gorm:"autoCreateTime"`   // 创建时间
	UpdatedAt int64  `json:"updated_at" gorm:"autoUpdateTime"`   // 更新时间
}
