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
		&models.Bullet{},
		&models.Skill{},
		&models.Scene{},
		&models.Treasure{},
		&models.MyItem{},
		&models.HomeConfig{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migrated successfully")
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migrated successfully")
}
