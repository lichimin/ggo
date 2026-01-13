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
		&models.Mail{},
		&models.MyItem{},
		&models.HomeConfig{},
		&models.EquipmentTemplate{},
		&models.UserEquipment{},
		&models.EquipmentAdditionalAttr{},
		&models.Archive{},
		&models.Area{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migrated successfully")

	// 如果json_data字段还是text类型，转换为jsonb类型
	err = DB.Exec("DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='archives' AND column_name='json_data' AND data_type='text') THEN ALTER TABLE archives ALTER COLUMN json_data TYPE jsonb USING json_data::jsonb; END IF; END $$;").Error
	if err != nil {
		log.Println("Warning: Failed to convert json_data to jsonb:", err)
	} else {
		log.Println("Successfully converted json_data to jsonb type")
	}

	// 为Archive表添加JSON字段的GIN索引，提升排行榜查询性能
	err = DB.Exec("CREATE INDEX IF NOT EXISTS idx_archives_json_data ON archives USING GIN (json_data)").Error
	if err != nil {
		log.Println("Warning: Failed to create GIN index on archives.json_data:", err)
	} else {
		log.Println("Created GIN index on archives.json_data for better performance")
	}
}
