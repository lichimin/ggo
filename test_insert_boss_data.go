package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Archive 模型定义
type Archive struct {
	ID       uint   `json:"id" gorm:"primarykey"`
	UserID   uint   `json:"user_id" gorm:"not null;index:idx_user_id,unique"`
	JSONData string `json:"json_data" gorm:"type:jsonb;not null"`
}

func main() {
	// 从环境变量获取数据库连接信息
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		// 如果环境变量不存在，使用默认值
		dsn = "host=27.154.56.154 user=postgres password=zity123456 dbname=test_zity port=10006 sslmode=disable"
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("数据库连接成功")

	// 准备测试数据 - 使用用户提供的完整JSON数据
	jsonData := `{
		"openid": "",
		"name": "lcm",
		"open_id": "oofyC18yIM2sdokf7KAiy80dRz6Q",
		"token": "",
		"img": "/res/head.jpg",
		"title": "/res/game/scene/title/fbz.png",
		"gold": 5133908,
		"diamond": 115000,
		"level": 25,
		"exp": 3421,
		"chapter": 27,
		"touchchapter": 26,
		"gametime": 31910,
		"take_time": 1767834760486,
		"cdkey": ["lcm1108"],
		"kill": 0,
		"guide_list": {
			"1": {"level": 1, "submitted": 0},
			"2": {"level": 3, "submitted": 0},
			"3": {"level": 1, "submitted": 0},
			"7": {"level": 2, "submitted": 1},
			"8": {"level": 2, "submitted": 0},
			"9": {"level": 6, "submitted": 5},
			"10": {"level": 1, "submitted": 0},
			"11": {"level": 1, "submitted": 0},
			"12": {"level": 1, "submitted": 0},
			"13": {"level": 1, "submitted": 0}
		},
		"user_treasures": [
			{"treasure_id": 7, "quantity": 79},
			{"treasure_id": 1, "quantity": 92},
			{"treasure_id": 8, "quantity": 95},
			{"treasure_id": 13, "quantity": 90},
			{"treasure_id": 11, "quantity": 85},
			{"treasure_id": 2, "quantity": 87},
			{"treasure_id": 12, "quantity": 87},
			{"treasure_id": 10, "quantity": 76},
			{"treasure_id": 9, "quantity": 40},
			{"treasure_id": 3, "quantity": 63}
		],
		"user_skins": {
			"skin_id": [3, 2, {"id": 4, "star": 3}],
			"use_skin_id": [2]
		},
		"user_equipments": {
			"equipments": [
				{
					"equipment_id": 61,
					"level": 0,
					"attr": [
						{"attr_name": "move_speed", "attr_value": 13, "is_rare": false, "desc": "移动速度"},
						{"attr_name": "poison_damage", "attr_value": 22, "is_rare": false, "desc": "毒系伤害"},
						{"attr_name": "frost_damage", "attr_value": 18, "is_rare": false, "desc": "冰霜伤害"}
					],
					"identify_count": 5,
					"reinforceLevel": 6
				},
				{
					"equipment_id": 62,
					"level": 0,
					"attr": [
						{"attr_name": "critical_rate", "attr_value": 0.057, "is_rare": false, "desc": "暴击率"},
						{"attr_name": "flame_damage", "attr_value": 19, "is_rare": false, "desc": "火焰伤害"},
						{"attr_name": "frost_damage", "attr_value": 8, "is_rare": false, "desc": "冰霜伤害"},
						{"attr_name": "poison_damage", "attr_value": 6, "is_rare": false, "desc": "毒系伤害"}
					],
					"identify_count": 5
				}
			],
			"unequipments": [
				{"equipment_id": 38, "level": 0, "attr": [], "identify_count": 0},
				{"equipment_id": 63, "level": 0, "attr": [], "identify_count": 0},
				{"equipment_id": 28, "level": 0, "attr": [{"attr_name": "hp", "attr_value": 31, "is_rare": false, "desc": "生命值"}], "identify_count": 2},
				{"equipment_id": 22, "level": 0, "attr": [], "identify_count": 0},
				{"equipment_id": 47, "attr": [], "reinforceLevel": 0, "identify_count": 0},
				{"equipment_id": 15, "level": 1, "attr": [], "identify_count": 0},
				{"equipment_id": 16, "level": 1, "attr": [], "identify_count": 0},
				{"equipment_id": 23, "level": 2, "attr": [], "identify_count": 0}
			]
		},
		"Area": 3,
		"starChart": {"unlockedNodes": []},
		"base_attributes": {
			"hp": 250,
			"attack": 50,
			"attack_speed": 1.3,
			"move_speed": 100,
			"drain": 0,
			"critical_rate": 0,
			"critical_damage": 1.5,
			"dodge": 0,
			"instant_kill": 0,
			"recovery": 0,
			"flame_damage": 0,
			"frost_damage": 0,
			"poison_damage": 0,
			"damage_reduction": 0,
			"atk_type": 0,
			"attack_mode": 1
		},
		"autoDecompose": false,
		"autoDecomposeLevel": 1,
		"recycle_data": {"date": "2026/1/8", "count": 2},
		"mystery_shop": {
			"last_refresh_time": 1767768424209,
			"items": [
				{
					"type": "treasure",
					"id": 10,
					"level": 2,
					"price": 20000,
					"isPurchased": true,
					"hasExtraStats": false,
					"extraStats": null,
					"name": "小兔玩偶",
					"image_url": "https://czrimg.godqb.com/game/bw/2-3.jpg"
				},
				{
					"type": "equipment",
					"id": 17,
					"level": 1,
					"price": 20000,
					"isPurchased": false,
					"hasExtraStats": false,
					"extraStats": null,
					"name": "生锈铁戒",
					"image_url": "/res/game/zb/pt2.png"
				}
			]
		},
		"diamonds": 809,
		"pending_idle_rewards": null,
		"skins_time": 1767766749286,
		"boss_last_result": {
			"time": 60,
			"damage": 4002,
			"updated_at": 1767854464876
		}
	}`

	// 插入测试数据
	archive := Archive{
		UserID:   13, // 使用新的UserID以避免冲突
		JSONData: jsonData,
	}

	result := db.Create(&archive)
	if result.Error != nil {
		log.Fatalf("插入数据失败: %v", result.Error)
	}
	fmt.Printf("成功插入存档记录，ID: %d, UserID: %d\n", archive.ID, archive.UserID)

	// 验证插入的数据
	var insertedArchive Archive
	result = db.Where("user_id = ?", 13).First(&insertedArchive)
	if result.Error != nil {
		log.Fatalf("查询插入的数据失败: %v", result.Error)
	}
	fmt.Printf("验证插入成功，JSONData长度: %d\n", len(insertedArchive.JSONData))

	// 测试查询boss_last_result字段
	var count int64
	db.Model(&Archive{}).Where("json_data#>>'{boss_last_result,damage}' IS NOT NULL").Count(&count)
	fmt.Printf("当前数据库中有 %d 条记录包含boss_last_result.damage字段\n", count)
}
