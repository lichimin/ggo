package models

type Mail struct {
	ID        uint   `json:"id" gorm:"primarykey"`
	UserID    uint   `json:"user_id" gorm:"not null;index"`
	Title     string `json:"title" gorm:"size:100;default:''"`
	Content   string `json:"content" gorm:"type:text;not null"`
	ItemType  string `json:"item_type" gorm:"size:20;default:''"`
	ItemID    uint   `json:"item_id" gorm:"default:0"`
	Num       int    `json:"num" gorm:"default:0"`
	Lv        int    `json:"lv" gorm:"default:0"`
	Status    int    `json:"status" gorm:"not null;default:0;index"`
	CreatedAt int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

