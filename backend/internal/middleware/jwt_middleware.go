package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"task-backend/pkg/jwtx"
	"task-backend/pkg/response"
)

const (
	ContextUserIDKey    = "user_id"
	ContextUsernameKey  = "username"
	ContextUserLevelKey = "user_level"
	SessionCookieName   = "task_session"
)

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""
		if cookieToken, err := c.Cookie(SessionCookieName); err == nil {
			tokenString = cookieToken
		}

		// 优化点 1: 更优雅、高效的 Bearer Token 解析方式
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") || strings.HasPrefix(authHeader, "bearer ") {
				tokenString = authHeader[7:] // 直接截取 "Bearer " 之后的字符
			}
		}

		if tokenString == "" {
			response.Unauthorized(c, "missing auth token")
			c.Abort()
			return
		}

		claims, err := jwtx.ParseToken(tokenString, secret)
		if err != nil {
			// 优化点 2: 统一使用封装好的 response.Unauthorized 保证全站错误规范一致
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Set(ContextUserLevelKey, claims.Level)
		c.Next()
	}
}