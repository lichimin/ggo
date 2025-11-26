package routes

import (
	"ggo/controllers"
	"ggo/database"
	"ggo/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	router := gin.Default()

	// 全局中间件
	router.Use(middleware.CORS())

	// 创建控制器实例
	userController := controllers.NewUserController(database.DB)
	skinController := controllers.NewSkinController(database.DB)
	monsterController := controllers.NewMonsterController(database.DB)
	userSkinController := controllers.NewUserSkinController(database.DB)
	bulletController := controllers.NewBulletController(database.DB)
	skillController := controllers.NewSkillController(database.DB)
	sceneController := controllers.NewSceneController(database.DB)
	treasureController := controllers.NewTreasureController(database.DB)
	myItemController := controllers.NewMyItemController(database.DB)
	homeConfigController := controllers.NewHomeConfigController(database.DB)
	equipmentController := controllers.NewEquipmentController(database.DB)
	equipmentEnhanceController := controllers.NewEquipmentEnhanceController(database.DB)
	heroController := controllers.NewHeroController(database.DB)

	// 公开路由（无需认证）
	public := router.Group("/api/v1")
	{
		public.POST("/login", userController.Login)
		public.GET("/scenes", sceneController.GetScenes)          // 场景列表设为公开接口
		public.GET("/treasures", treasureController.GetTreasures) // 宝物列表设为公开接口
		public.GET("/home-configs", homeConfigController.GetHomeConfigs)
	}

	// 受保护路由（需要认证）
	protected := router.Group("/api/v1")
	protected.Use(middleware.JWTAuth())
	{
		// 用户相关
		protected.GET("/profile", userController.GetProfile)
		protected.GET("/users", userController.GetUsers)
		protected.GET("/users/:id", userController.GetUser)
		protected.POST("/users", userController.CreateUser)
		protected.PUT("/users/:id", userController.UpdateUser)
		protected.DELETE("/users/:id", userController.DeleteUser)

		// 皮肤相关
		protected.GET("/skins", skinController.GetSkins)
		protected.GET("/skins/:id", skinController.GetSkin)
		protected.POST("/skins", skinController.CreateSkin)
		protected.PUT("/skins/:id", skinController.UpdateSkin)
		protected.DELETE("/skins/:id", skinController.DeleteSkin)

		// 怪物相关
		protected.GET("/monsters", monsterController.GetMonsters)
		protected.GET("/monsters/:id", monsterController.GetMonster)
		protected.POST("/monsters", monsterController.CreateMonster)
		protected.PUT("/monsters/:id", monsterController.UpdateMonster)
		protected.DELETE("/monsters/:id", monsterController.DeleteMonster)

		// 用户皮肤相关
		protected.POST("/user/skins/acquire", userSkinController.AcquireSkin)
		protected.GET("/user/skins", userSkinController.GetUserSkins)
		protected.PUT("/user/skins/:skin_id/activate", userSkinController.ActivateSkin)
		protected.GET("/user/skins/active", userSkinController.GetActiveSkin)
		protected.DELETE("/user/skins/:skin_id", userSkinController.DeleteUserSkin)

		// 子弹相关
		protected.GET("/bullets", bulletController.GetBullets)
		protected.GET("/bullets/:id", bulletController.GetBullet)
		protected.POST("/bullets", bulletController.CreateBullet)
		protected.PUT("/bullets/:id", bulletController.UpdateBullet)
		protected.DELETE("/bullets/:id", bulletController.DeleteBullet)

		// 技能相关
		protected.GET("/skills", skillController.GetSkills)
		protected.GET("/skills/:id", skillController.GetSkill)
		protected.POST("/skills", skillController.CreateSkill)
		protected.PUT("/skills/:id", skillController.UpdateSkill)
		protected.DELETE("/skills/:id", skillController.DeleteSkill)

		// 我的物品相关
		protected.POST("/my-items", myItemController.AddMyItem)
		protected.GET("/my-items", myItemController.GetMyItems)
		protected.POST("/my-items/sell-multiple", myItemController.SellMultipleTreasures) // 批量出售宝物

		// 装备相关
		protected.POST("/equipments/generate", equipmentController.GenerateEquipment) // 生成装备
		protected.GET("/equipments", equipmentController.GetUserEquipments)           // 获取用户装备列表
		protected.PUT("/equipments/:id/equip", equipmentController.EquipEquipment)    // 装备/取消装备
		// 装备强化相关
		protected.POST("/equipments/merge", equipmentEnhanceController.MergeEquipment)     // 融合装备
		protected.POST("/equipments/enhance", equipmentEnhanceController.EnhanceEquipment) // 强化装备

		// 英雄相关
		protected.POST("/heroes/draw", heroController.DrawHero)     // 抽取英雄（十连抽）
		protected.POST("/heroes/awaken", heroController.AwakenHero) // 觉醒英雄
		protected.GET("/heroes", heroController.GetUserHeroes)      // 获取用户英雄列表
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"message": "Service is running",
		})
	})

	return router
}
