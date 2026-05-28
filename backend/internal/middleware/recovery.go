package middleware

import (
	"github.com/gin-gonic/gin"

	"task-backend/internal/apperror"
)

// RecoverWithLog 捕获 panic，记录堆栈信息并向客户端返回统一的 500 响应。
// 遵循“早抛出，晚捕获”的思想：在业务层允许 panic 向上抛出，
// 在全局中间件统一记录并屏蔽详细堆栈以免泄露给客户端。
func RecoverWithLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				apperror.RecoverPanic(c, r)
			}
		}()
		c.Next()
	}
}
