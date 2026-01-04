package models

// Area 区服模型
type Area struct {
	ID       uint   `json:"id" gorm:"primarykey"`
	Area     int    `json:"area" gorm:"not null;unique;index:idx_area"` // 区服编号
	IsNew    bool   `json:"is_new" gorm:"not null;default:false"`       // 是否新服
	Status   int    `json:"status" gorm:"not null;default:1"`           // 区服状态 1:正常 2:维护 3:爆满
	Name     string `json:"name" gorm:"size:50"`                        // 区服名称
	MaxUsers int    `json:"max_users" gorm:"not null;default:1000"`     // 最大用户数
}

// TableName 指定表名
func (Area) TableName() string {
	return "areas"
}
