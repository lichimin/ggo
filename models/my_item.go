package models

type MyItem struct {
	ID        uint   `json:"id" gorm:"primarykey"`
	UserID    uint   `json:"user_id" gorm:"not null;index"`              // 用户ID
	ItemID    uint   `json:"item_id" gorm:"not null"`                    // 物品ID
	ItemType  string `json:"item_type" gorm:"size:20;not null"`          // 物品类型：treasure(宝物), equipment(装备)
	SellPrice int    `json:"sell_price" gorm:"default:0"`                // 出售价格
	Position  string `json:"position" gorm:"size:20;default:'backpack'"` // 物品位置：backpack(背包), warehouse(仓库), equipped(装备中)
	Quantity  int    `json:"quantity" gorm:"default:1"`                  // 数量
	IsActive  bool   `json:"is_active" gorm:"default:true"`              // 是否激活
	CreatedAt int64  `json:"created_at" gorm:"autoCreateTime"`           // 创建时间
	UpdatedAt int64  `json:"updated_at" gorm:"autoUpdateTime"`           // 更新时间
}

// TreasureInfo 用于返回物品的详细信息
type TreasureInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	ImageURL    string `json:"image_url"`
	Value       int    `json:"value"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}
