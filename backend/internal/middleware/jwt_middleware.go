package middleware

import (
	"net/http"
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

		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenString = parts[1]
			}
		}

		if tokenString == "" {
			response.Unauthorized(c, "missing auth token")
			c.Abort()
			return
		}

		claims, err := jwtx.ParseToken(tokenString, secret)
		if err != nil {
			response.JSON(c, http.StatusUnauthorized, 401, "invalid or expired token", nil)
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Set(ContextUserLevelKey, claims.Level)
		c.Next()
	}
}
