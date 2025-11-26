package models

type UserEquipment struct {
	ID           uint   `json:"id" gorm:"primarykey"`
	UserID       uint   `json:"user_id" gorm:"not null;index"`
	EquipmentID  uint   `json:"equipment_id" gorm:"not null"`
	IsEquipped   bool   `json:"is_equipped" gorm:"default:false"`
	Position     string `json:"position" gorm:"size:20;default:'backpack'"`
	EnhanceLevel int    `json:"enhance_level" gorm:"default:0"` // 强化等级
	CreatedAt    int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    int64  `json:"updated_at" gorm:"autoUpdateTime"`

	EquipmentTemplate EquipmentTemplate `json:"equipment_template" gorm:"foreignKey:EquipmentID"`
	// 关联附属属性
	AdditionalAttrs []EquipmentAdditionalAttr `json:"additional_attrs" gorm:"foreignKey:UserEquipmentID"`
}
