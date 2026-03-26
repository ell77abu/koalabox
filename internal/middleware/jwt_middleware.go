package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"task-backend/pkg/jwtx"
	"task-backend/pkg/response"
)

const (
	ContextUserIDKey   = "user_id"
	ContextUsernameKey = "username"
)

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing Authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "invalid Authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtx.ParseToken(parts[1], secret)
		if err != nil {
			response.JSON(c, http.StatusUnauthorized, 401, "invalid or expired token", nil)
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Next()
	}
}
