package response

import "github.com/gin-gonic/gin"

type Common struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PageData struct {
	List  interface{} `json:"list"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Total int64       `json:"total"`
}

func JSON(c *gin.Context, httpStatus, code int, message string, data interface{}) {
	c.JSON(httpStatus, Common{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func Success(c *gin.Context, message string, data interface{}) {
	JSON(c, 200, 0, message, data)
}

func Created(c *gin.Context, message string, data interface{}) {
	JSON(c, 201, 0, message, data)
}

func SuccessPage(c *gin.Context, message string, list interface{}, page, size int, total int64) {
	Success(c, message, PageData{
		List:  list,
		Page:  page,
		Size:  size,
		Total: total,
	})
}

func BadRequest(c *gin.Context, message string) {
	JSON(c, 400, 400, message, nil)
}

func Unauthorized(c *gin.Context, message string) {
	JSON(c, 401, 401, message, nil)
}

func Forbidden(c *gin.Context, message string) {
	JSON(c, 403, 403, message, nil)
}

func NotFound(c *gin.Context, message string) {
	JSON(c, 404, 404, message, nil)
}

func Conflict(c *gin.Context, message string) {
	JSON(c, 409, 409, message, nil)
}

func Gone(c *gin.Context, message string) {
	JSON(c, 410, 410, message, nil)
}

func InternalError(c *gin.Context, _ string) {
	JSON(c, 500, 500, "服务器内部错误，请稍后重试", nil)
}
