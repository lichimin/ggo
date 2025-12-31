package database

import (
	"ggo/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitPostgres(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Connected to PostgreSQL")

	// 自动迁移表结构
	err = DB.AutoMigrate(
		&models.User{},
		&models.Skin{},
		&models.Monster{},
		&models.UserSkin{},
		&models.Scene{},
		&models.Treasure{},
		&models.MyItem{},
		&models.HomeConfig{},
		&models.EquipmentTemplate{},
		&models.UserEquipment{},
		&models.EquipmentAdditionalAttr{},
		&models.Archive{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migrated successfully")

	// 为Archive表添加JSON字段的GIN索引，提升排行榜查询性能
	err = DB.Exec("CREATE INDEX IF NOT EXISTS idx_archives_json_data ON archives USING GIN ((json_data::jsonb))").Error
	if err != nil {
		log.Println("Warning: Failed to create GIN index on archives.json_data:", err)
	} else {
		log.Println("Created GIN index on archives.json_data for better performance")
	}
}
