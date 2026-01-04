package models

// Area 区服模型
type Area struct {
	ID    uint   `json:"id" gorm:"primarykey"`
	Area  string `json:"area" gorm:"size:50"`                  // 区服编号
	IsNew bool   `json:"is_new" gorm:"not null;default:false"` // 是否新服
}

// TableName 指定表名
func (Area) TableName() string {
	return "areas"
}
