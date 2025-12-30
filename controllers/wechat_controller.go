package controllers

import (
	"encoding/json"
	"fmt"
	"ggo/config"
	"ggo/utils"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WeChatController 微信小程序控制器
type WeChatController struct {
	cfg *config.Config
}

// NewWeChatController 创建微信控制器实例
func NewWeChatController(cfg *config.Config) *WeChatController {
	return &WeChatController{
		cfg: cfg,
	}
}

// GetOpenID 通过临时登录凭证code获取微信小游戏的openid
func (wc *WeChatController) GetOpenID(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 调用微信API获取openid和session_key
	// 微信API文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/login/auth.code2Session.html
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		wc.cfg.WeChatAppID, wc.cfg.WeChatAppSecret, req.Code)

	resp, err := http.Get(url)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "调用微信API失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "读取响应失败: "+err.Error())
		return
	}

	// 解析微信API返回的数据
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "解析响应数据失败: "+err.Error())
		return
	}

	// 检查是否有错误
	if errCode, ok := response["errcode"]; ok && errCode.(float64) != 0 {
		errMsg := "获取openid失败"
		if msg, ok := response["errmsg"]; ok {
			errMsg = fmt.Sprintf("获取openid失败: %s", msg)
		}
		utils.ErrorResponse(c, http.StatusBadRequest, errMsg)
		return
	}

	// 提取openid和session_key
	openid, ok := response["openid"]
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "响应中缺少openid")
		return
	}

	sessionKey, ok := response["session_key"]
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "响应中缺少session_key")
		return
	}

	// 返回openid和session_key
	utils.SuccessResponse(c, gin.H{
		"openid":      openid,
		"session_key": sessionKey,
	})
}
