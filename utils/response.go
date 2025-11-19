package utils

import "github.com/gin-gonic/gin"

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Success: true,
		Message: "success",
		Data:    data,
	})
}

func SuccessResponseWithToken(c *gin.Context, data interface{}, token string) {
	// 如果有新token，设置到响应头
	if token != "" {
		c.Header("X-New-Token", token)
	}

	c.JSON(200, Response{
		Success: true,
		Message: "success",
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Success: false,
		Message: message,
		Data:    nil,
	})
}
