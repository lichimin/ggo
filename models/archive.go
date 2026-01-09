package models

import (
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm"
)

// JSONB 自定义JSONB类型，用于处理PostgreSQL的jsonb字段
type JSONB map[string]interface{}

// Value 实现 driver.Valuer 接口，将JSONB转换为数据库值
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口，将数据库值扫描到JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}
	*j = result
	return nil
}

// Archive 存档模型
type Archive struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	UserID    uint           `json:"user_id" gorm:"not null;index:idx_user_id,unique"` // 用户ID，唯一索引
	JSONData  JSONB          `json:"json_data" gorm:"type:jsonb;not null"`             // 存档数据，直接存储JSON对象
	V         int            `json:"v" gorm:"not null;default:0"`                      // 版本号，用于控制存档的更新顺序
	Area      int            `json:"area" gorm:"not null;default:1"`                   // 区服ID
	CreatedAt int64          `json:"created_at" gorm:"autoCreateTime"`                 // 创建时间
	UpdatedAt int64          `json:"updated_at" gorm:"autoUpdateTime"`                 // 更新时间
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`                                   // 软删除
}

// TableName 指定表名
func (Archive) TableName() string {
	return "archives"
}
