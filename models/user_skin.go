package models

type UserSkin struct {
	ID        uint  `json:"id" gorm:"primarykey"`
	UserID    uint  `json:"user_id" gorm:"not null;index"`    // 用户ID
	SkinID    uint  `json:"skin_id" gorm:"not null;index"`    // 皮肤ID
	IsActive  bool  `json:"is_active" gorm:"default:false"`   // 是否启用
	CreatedAt int64 `json:"created_at" gorm:"autoCreateTime"` // 创建时间
	UpdatedAt int64 `json:"updated_at" gorm:"autoUpdateTime"` // 更新时间

	// 关联信息（用于查询时加载）
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Skin Skin `json:"skin,omitempty" gorm:"foreignKey:SkinID"`
}
