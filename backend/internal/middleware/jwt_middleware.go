package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"task-backend/internal/apperror"
	"task-backend/pkg/jwtx"
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

		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenString = parts[1]
			}
		}

		if tokenString == "" {
			apperror.Respond(c, apperror.Unauthorized("未登录或登录状态已失效"), "jwt auth")
			c.Abort()
			return
		}

		claims, err := jwtx.ParseToken(tokenString, secret)
		if err != nil {
			apperror.Respond(c, apperror.Unauthorized("登录状态无效或已过期，请重新登录"), "jwt auth")
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Set(ContextUserLevelKey, claims.Level)
		c.Next()
	}
}
