package middleware

import (
	"context"
	"fmt"
	"ggo/database"
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

		if database.RedisClient == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证服务未就绪"})
			c.Abort()
			return
		}

		currentToken, err := database.RedisClient.Get(context.Background(), fmt.Sprintf("auth:token:%d", claims.UserID)).Result()
		if err != nil || currentToken != token {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token已作废"})
			c.Abort()
			return
		}

		// 每次请求都续签token
		newToken, err := utils.RefreshToken(token)
		if err == nil {
			// 设置新的token到响应头
			c.Header("X-New-Token", newToken)

			ttl, ttlErr := utils.GetRemainingTime(newToken)
			if ttlErr != nil {
				ttl = 7 * 24 * time.Hour
			}
			database.RedisClient.Set(context.Background(), fmt.Sprintf("auth:token:%d", claims.UserID), newToken, ttl)
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
				if database.RedisClient != nil {
					currentToken, err := database.RedisClient.Get(context.Background(), fmt.Sprintf("auth:token:%d", claims.UserID)).Result()
					if err != nil || currentToken != token {
						c.Next()
						return
					}
				}

				// 将用户信息存入context
				c.Set("userID", claims.UserID)
				c.Set("username", claims.Username)

				// 每次请求都续签token
				newToken, err := utils.RefreshToken(token)
				if err == nil {
					c.Header("X-New-Token", newToken)

					if database.RedisClient != nil {
						ttl, ttlErr := utils.GetRemainingTime(newToken)
						if ttlErr != nil {
							ttl = 7 * 24 * time.Hour
						}
						database.RedisClient.Set(context.Background(), fmt.Sprintf("auth:token:%d", claims.UserID), newToken, ttl)
					}
				}
			}
		}

		c.Next()
	}
}
