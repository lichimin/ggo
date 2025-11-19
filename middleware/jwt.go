package middleware

import (
	"ggo/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供token"})
			c.Abort()
			return
		}

		// 检查token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token格式错误"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token无效"})
			c.Abort()
			return
		}

		// 检查token是否即将过期（剩余时间小于30分钟），自动续签
		remainingTime := time.Until(claims.ExpiresAt.Time)
		if remainingTime < 30*time.Minute {
			newToken, err := utils.RefreshToken(token)
			if err == nil {
				// 设置新的token到响应头
				c.Header("X-New-Token", newToken)
			}
		}

		// 将用户信息存入context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// OptionalJWTAuth 可选的JWT认证中间件（不强制要求认证）
func OptionalJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			claims, err := utils.ParseToken(token)
			if err == nil {
				// 将用户信息存入context
				c.Set("userID", claims.UserID)
				c.Set("username", claims.Username)

				// 检查并刷新token
				remainingTime := time.Until(claims.ExpiresAt.Time)
				if remainingTime < 30*time.Minute {
					newToken, err := utils.RefreshToken(token)
					if err == nil {
						c.Header("X-New-Token", newToken)
					}
				}
			}
		}

		c.Next()
	}
}
