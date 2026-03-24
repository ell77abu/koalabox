package response

import "github.com/gin-gonic/gin"

type Common struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func JSON(c *gin.Context, httpStatus, code int, message string, data interface{}) {
	c.JSON(httpStatus, Common{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
