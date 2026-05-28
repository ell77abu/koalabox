package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"task-backend/pkg/response"
)

// RecoverWithLog 捕获 panic，记录堆栈信息并向客户端返回统一的 500 响应。
// 遵循“早抛出，晚捕获”的思想：在业务层允许 panic 向上抛出，
// 在全局中间件统一记录并屏蔽详细堆栈以免泄露给客户端。
func RecoverWithLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 记录详细堆栈到服务器日志，便于排查
				log.Printf("panic recovered: %v\n%s", r, debug.Stack())
				// 对外返回统一的错误信息，避免泄露内部实现细节
				response.JSON(c, http.StatusInternalServerError, 500, "internal server error", nil)
				c.Abort()
			}
		}()
		c.Next()
	}
}
