package main

import (
	"encoding/json"
	"fmt"
	"ggo/config"
	"ggo/controllers"
	"ggo/database"
	"ggo/models"
	"net/http"
	"net/http/httptest"
	"strings"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	database.InitPostgres(cfg.PostgresDSN)

	fmt.Println("=== 测试存档和区服功能 ===")

	// 1. 测试区服列表控制器
	fmt.Println("\n1. 测试区服列表功能:")
	areaController := controllers.NewAreaController(database.DB)

	// 创建测试请求
	req := httptest.NewRequest("GET", "/api/v1/areas", nil)
	w := httptest.NewRecorder()

	areaController.GetAreas(w)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ 区服列表接口响应正常")

		var response struct {
			Code int         `json:"code"`
			Msg  string      `json:"msg"`
			Data interface{} `json:"data"`
		}

		if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
			fmt.Printf("响应内容: %+v\n", response)

			// 解析areas数据
			if areasData, ok := response.Data.(map[string]interface{}); ok {
				if areas, ok := areasData["areas"].([]interface{}); ok {
					fmt.Printf("找到 %d 个区服:\n", len(areas))
					for i, area := range areas {
						if areaMap, ok := area.(map[string]interface{}); ok {
							fmt.Printf("  %d. ID:%v, Area:%v, IsNew:%v\n",
								i+1, areaMap["id"], areaMap["area"], areaMap["is_new"])
						}
					}
				}
			}
		}
	} else {
		fmt.Printf("❌ 区服列表接口调用失败，状态码: %d\n", resp.StatusCode)
		fmt.Printf("响应内容: %s\n", w.Body.String())
	}

	// 2. 测试存档保存功能（包含area参数）
	fmt.Println("\n2. 测试存档保存功能:")
	archiveController := controllers.NewArchiveController(database.DB)

	// 模拟存档数据
	testData := map[string]interface{}{
		"name":    "测试用户",
		"gold":    1000,
		"level":   10,
		"chapter": 5,
	}

	jsonData, _ := json.Marshal(testData)

	// 创建测试请求
	saveReq := map[string]interface{}{
		"json_data": testData,
		"v":         1,
		"area":      1,
	}

	saveJSON, _ := json.Marshal(saveReq)
	req = httptest.NewRequest("POST", "/api/v1/archive",
		strings.NewReader(string(saveJSON)))

	w = httptest.NewRecorder()

	// 注意：这个测试需要JWT认证，这里只是展示结构

	// 3. 直接测试数据库查询
	fmt.Println("\n3. 直接数据库查询测试:")

	// 查询区服列表
	var areas []models.Area
	result := database.DB.Order("area ASC").Find(&areas)
	if result.Error != nil {
		fmt.Printf("❌ 查询区服列表失败: %v\n", result.Error)
	} else {
		fmt.Printf("✅ 查询到 %d 个区服:\n", len(areas))
		for _, area := range areas {
			fmt.Printf("  ID:%d, Area:%d, IsNew:%v, Name:%s\n",
				area.ID, area.Area, area.IsNew, area.Name)
		}
	}

	// 查询存档（检查是否有area字段）
	var archives []models.Archive
	result = database.DB.Limit(3).Find(&archives)
	if result.Error != nil {
		fmt.Printf("❌ 查询存档失败: %v\n", result.Error)
	} else {
		fmt.Printf("✅ 查询到 %d 条存档记录:\n", len(archives))
		for _, archive := range archives {
			fmt.Printf("  UserID:%d, Area:%d\n", archive.UserID, archive.Area)
		}
	}

	fmt.Println("\n=== 测试完成 ===")
}
