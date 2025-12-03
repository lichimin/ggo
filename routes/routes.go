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
	sceneController := controllers.NewSceneController(database.DB)
	treasureController := controllers.NewTreasureController(database.DB)
	myItemController := controllers.NewMyItemController(database.DB)
	homeConfigController := controllers.NewHomeConfigController(database.DB)
	equipmentController := controllers.NewEquipmentController(database.DB)
	equipmentEnhanceController := controllers.NewEquipmentEnhanceController(database.DB)

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
		protected.GET("/user/attributes", userController.GetPlayerAttributes) // 获取玩家属性
		protected.GET("/users", userController.GetUsers)
		protected.GET("/users/:id", userController.GetUser)
		protected.POST("/users", userController.CreateUser)
		protected.PUT("/users/:id", userController.UpdateUser)
		protected.DELETE("/users/:id", userController.DeleteUser)

		// 皮肤相关
		protected.GET("/skins", skinController.GetSkins)
		protected.GET("/skins/:id", skinController.GetSkin)

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

		// 我的物品相关
		protected.POST("/my-items", myItemController.AddMyItem)
		protected.GET("/my-items", myItemController.GetMyItems) // 获取未穿戴的装备和其他物品
		protected.GET("/my-items/equipped", myItemController.GetEquippedItems) // 获取已穿戴的装备
		protected.POST("/my-items/sell-multiple", myItemController.SellMultipleTreasures) // 批量出售宝物

		// 装备相关
		protected.POST("/equipments/generate", equipmentController.GenerateEquipment) // 生成装备
		protected.GET("/equipments", equipmentController.GetUserEquipments)           // 获取用户装备列表
		protected.PUT("/equipments/:id/equip", equipmentController.EquipItem)         // 穿戴装备
		protected.PUT("/equipments/:id/unequip", equipmentController.UnequipItem)     // 卸下装备
		// 装备强化相关
		protected.POST("/equipments/merge", equipmentEnhanceController.MergeEquipment)         // 融合装备
		protected.POST("/equipments/:id/enhance", equipmentEnhanceController.EnhanceEquipment) // 强化装备

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
