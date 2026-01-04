package controllers

import (
	"encoding/json"
	"fmt"
	"ggo/models"
	"ggo/utils"
	"net/http"

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
		V        int         `json:"v" binding:"required"`
		Area     int         `json:"area" binding:"required"`
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

	// 使用事务确保数据一致性
	var saveSuccess bool
	var responseMessage string
	ac.db.Transaction(func(tx *gorm.DB) error {
		var archive models.Archive
		result := tx.Where("user_id = ? and area = ?", userID, req.Area).First(&archive)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// 创建新存档
				archive = models.Archive{
					UserID:   userID.(uint),
					JSONData: string(jsonData),
					V:        req.V,
					Area:     req.Area,
				}
				if err := tx.Create(&archive).Error; err != nil {
					return err
				}
				saveSuccess = true
				responseMessage = "新存档创建成功"
				return nil
			}
			return result.Error
		}

		// 检查版本号，如果当前版本大于数据库版本则更新
		if req.V > archive.V {
			archive.JSONData = string(jsonData)
			archive.V = req.V
			archive.Area = req.Area
			if err := tx.Save(&archive).Error; err != nil {
				return err
			}
			saveSuccess = true
			responseMessage = fmt.Sprintf("存档更新成功，版本号从 %d 升级到 %d", archive.V, req.V)
		} else {
			responseMessage = fmt.Sprintf("存档跳过，传入版本 %d 不大于当前版本 %d", req.V, archive.V)
		}

		return nil
	})

	if saveSuccess {
		utils.SuccessResponse(c, gin.H{"message": responseMessage})
	} else {
		utils.ErrorResponse(c, http.StatusOK, responseMessage)
	}
}

// LoadArchive 读取用户存档
func (ac *ArchiveController) LoadArchive(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		return
	}

	// 从请求参数获取 area
	area := 1 // 默认区服
	if areaParam := c.Query("area"); areaParam != "" {
		fmt.Sscanf(areaParam, "%d", &area)
	}

	// 直接从数据库读取存档
	var archive models.Archive
	result := ac.db.Where("user_id = ? AND area = ?", userID, area).First(&archive)
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

	utils.SuccessResponse(c, gin.H{
		"json_data": data,
		"v":         archive.V,
		"area":      archive.Area,
	})
}
