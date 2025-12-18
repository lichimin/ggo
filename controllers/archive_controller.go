package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"ggo/database"
	"ggo/models"
	"ggo/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArchiveController struct {
	db *gorm.DB
}

func NewArchiveController(db *gorm.DB) *ArchiveController {
	return &ArchiveController{
		db: db,
	}
}

// SaveArchive 保存或更新用户存档
func (ac *ArchiveController) SaveArchive(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}

	var req struct {
		JSONData interface{} `json:"json_data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 将JSON数据转换为字符串
	jsonData, err := json.Marshal(req.JSONData)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "JSON序列化失败: "+err.Error())
		return
	}

	// 生成Redis键
	redisKey := fmt.Sprintf("archive:%d", userID.(uint))
	ctx := context.Background()

	// 使用Redis分布式锁确保并发安全
	lockKey := fmt.Sprintf("lock:archive:%d", userID.(uint))
	lock := database.RedisClient.SetNX(ctx, lockKey, "locked", 5*time.Second)
	if lock.Val() == false {
		utils.ErrorResponse(c, http.StatusConflict, "存档正在保存中，请稍后重试")
		return
	}
	defer database.RedisClient.Del(ctx, lockKey)

	// 先保存到Redis缓存
	database.RedisClient.Set(ctx, redisKey, string(jsonData), 24*time.Hour)

	// 保存到数据库（异步处理）
	go func() {
		var archive models.Archive
		// 使用事务确保数据一致性
		ac.db.Transaction(func(tx *gorm.DB) error {
			// 查找是否存在存档
			result := tx.Where("user_id = ?", userID).First(&archive)
			if result.Error != nil {
				if result.Error == gorm.ErrRecordNotFound {
					// 创建新存档
					archive = models.Archive{
						UserID:   userID.(uint),
						JSONData: string(jsonData),
					}
					return tx.Create(&archive).Error
				}
				return result.Error
			}

			// 更新现有存档
			archive.JSONData = string(jsonData)
			return tx.Save(&archive).Error
		})
	}()

	utils.SuccessResponse(c, gin.H{"message": "存档保存成功"})
}

// LoadArchive 读取用户存档
func (ac *ArchiveController) LoadArchive(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		return
	}

	// 生成Redis键
	redisKey := fmt.Sprintf("archive:%d", userID.(uint))
	ctx := context.Background()

	// 先从Redis缓存读取
	jsonData, err := database.RedisClient.Get(ctx, redisKey).Result()
	if err == nil {
		// 缓存命中，解析并返回
		var data interface{}
		if err := json.Unmarshal([]byte(jsonData), &data); err == nil {
			utils.SuccessResponse(c, gin.H{"json_data": data})
			return
		}
	}

	// 缓存未命中或解析失败，从数据库读取
	var archive models.Archive
	result := ac.db.Where("user_id = ?", userID).First(&archive)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "存档不存在")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "读取存档失败: "+result.Error.Error())
		}
		return
	}

	// 解析JSON数据
	var data interface{}
	if err := json.Unmarshal([]byte(archive.JSONData), &data); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "解析存档数据失败: "+err.Error())
		return
	}

	// 将数据存入Redis缓存
	database.RedisClient.Set(ctx, redisKey, archive.JSONData, 24*time.Hour)

	utils.SuccessResponse(c, gin.H{"json_data": data})
}
