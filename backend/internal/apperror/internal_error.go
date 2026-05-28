package apperror

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"task-backend/internal/repository"
	"task-backend/internal/service"
	"task-backend/pkg/response"
)

const PublicInternalErrorMessage = "服务器内部错误，请稍后重试"

// HTTPError 用于在边界层统一表达可直接返回给客户端的业务错误。
type HTTPError struct {
	HTTPStatus int
	Code       int
	Message    string
	Data       interface{}
}

func (e *HTTPError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func newHTTPError(httpStatus, code int, message string, data interface{}) *HTTPError {
	return &HTTPError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
		Data:       data,
	}
}

func BadRequest(message string) error {
	return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, message, nil)
}

func Unauthorized(message string) error {
	return newHTTPError(http.StatusUnauthorized, http.StatusUnauthorized, message, nil)
}

func Forbidden(message string) error {
	return newHTTPError(http.StatusForbidden, http.StatusForbidden, message, nil)
}

// Respond 统一错误出口：优先处理可映射业务错误，无法识别时降级为内部错误。
func Respond(c *gin.Context, err error, operation string) {
	if err == nil {
		return
	}

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		response.JSON(c, httpErr.HTTPStatus, httpErr.Code, httpErr.Message, httpErr.Data)
		return
	}

	if mapped, ok := mapDomainError(err); ok {
		response.JSON(c, mapped.HTTPStatus, mapped.Code, mapped.Message, mapped.Data)
		return
	}

	RespondInternal(c, err, operation)
}

func mapDomainError(err error) (*HTTPError, bool) {
	var unresolvedErr *service.UnresolvedDependenciesError
	if errors.As(err, &unresolvedErr) {
		return newHTTPError(http.StatusConflict, http.StatusConflict, "尚有前置任务未完成", gin.H{
			"unresolved_predecessor_ids": unresolvedErr.IDs,
			"unresolved_predecessors":    unresolvedErr.Predecessors,
		}), true
	}

	switch {
	case errors.Is(err, service.ErrInvalidUsername):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "用户名格式不正确", nil), true
	case errors.Is(err, service.ErrInvalidPassword):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "密码格式不正确", nil), true
	case errors.Is(err, service.ErrInvalidOldPassword):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "旧密码错误", nil), true
	case errors.Is(err, service.ErrInvalidCredentials):
		return newHTTPError(http.StatusUnauthorized, http.StatusUnauthorized, "用户名或密码错误", nil), true

	case errors.Is(err, service.ErrInvalidTaskTitle):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "任务标题不合法", nil), true
	case errors.Is(err, service.ErrInvalidTaskStatus):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "任务状态不合法", nil), true
	case errors.Is(err, service.ErrInvalidTaskPriority):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "任务优先级不合法", nil), true
	case errors.Is(err, service.ErrInvalidTaskDeadline):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "截止时间格式不正确", nil), true
	case errors.Is(err, service.ErrInvalidTeamID):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "团队不存在或团队ID不合法", nil), true
	case errors.Is(err, service.ErrInvalidAssigneeID):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "执行者不存在或执行者ID不合法", nil), true
	case errors.Is(err, service.ErrAssigneeNotInTeam):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "执行者不在该团队中", nil), true
	case errors.Is(err, service.ErrInvalidTaskDependency):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "任务依赖参数不合法", nil), true
	case errors.Is(err, service.ErrDependenciesNotDone):
		return newHTTPError(http.StatusConflict, http.StatusConflict, "尚有前置任务未完成", nil), true
	case errors.Is(err, service.ErrDependencyCycle):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "无法添加依赖：检测到依赖循环", nil), true
	case errors.Is(err, service.ErrDependencyOnDoneTask):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "已完成任务不能再设置前置依赖", nil), true
	case errors.Is(err, service.ErrCrossTaskDependency):
		return newHTTPError(http.StatusForbidden, http.StatusForbidden, "跨团队或跨个人任务禁止建立依赖关系", nil), true
	case errors.Is(err, service.ErrAssignmentNotPending):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "任务指派已处理，不能重复操作", nil), true
	case errors.Is(err, service.ErrTeamRoleDenied):
		return newHTTPError(http.StatusForbidden, http.StatusForbidden, "权限不足，无法执行该操作", nil), true
	case errors.Is(err, service.ErrMemberStatusOnly):
		return newHTTPError(http.StatusForbidden, http.StatusForbidden, "成员仅可修改分配给自己的任务状态", nil), true
	case errors.Is(err, service.ErrTeamIDImmutable):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "任务不允许修改所属团队", nil), true
	case errors.Is(err, service.ErrTeamGone):
		return newHTTPError(http.StatusGone, http.StatusGone, "团队已删除", nil), true

	case errors.Is(err, service.ErrInvalidTeamName):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "团队名称不合法", nil), true
	case errors.Is(err, service.ErrInvalidTeamRole):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "团队角色不合法", nil), true
	case errors.Is(err, service.ErrInvalidTeamMember):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "团队成员参数不合法", nil), true
	case errors.Is(err, service.ErrInvitationNotPending):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "邀请已处理，不能重复操作", nil), true
	case errors.Is(err, service.ErrInvalidTransferTarget):
		return newHTTPError(http.StatusBadRequest, http.StatusBadRequest, "转让目标不合法", nil), true
	case errors.Is(err, service.ErrTeamOwnerOnly),
		errors.Is(err, service.ErrOwnerCannotLeave),
		errors.Is(err, service.ErrCannotRemoveSelf),
		errors.Is(err, service.ErrCannotModifyOwner):
		return newHTTPError(http.StatusForbidden, http.StatusForbidden, "权限不足，无法执行该操作", nil), true

	case errors.Is(err, repository.ErrForbidden):
		return newHTTPError(http.StatusForbidden, http.StatusForbidden, "权限不足，无法执行该操作", nil), true
	case errors.Is(err, repository.ErrNotFound):
		return newHTTPError(http.StatusNotFound, http.StatusNotFound, "请求的资源不存在或已被删除", nil), true
	case errors.Is(err, repository.ErrConflict):
		return newHTTPError(http.StatusConflict, http.StatusConflict, "资源冲突，请勿重复操作", nil), true
	}

	return nil, false
}

// RespondInternal 统一记录内部异常并返回安全的 500 响应。
func RespondInternal(c *gin.Context, err error, operation string) {
	if err == nil {
		err = fmt.Errorf("nil internal error")
	}

	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}

	userID, hasUserID := c.Get("user_id")
	if hasUserID {
		slog.Error(
			"internal_error",
			slog.String("operation", operation),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Any("user_id", userID),
			slog.String("error", err.Error()),
			slog.String("stack", string(debug.Stack())),
		)
	} else {
		slog.Error(
			"internal_error",
			slog.String("operation", operation),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("error", err.Error()),
			slog.String("stack", string(debug.Stack())),
		)
	}

	response.InternalError(c, PublicInternalErrorMessage)
}

// RecoverPanic 统一记录 panic 堆栈并返回安全的 500 响应。
func RecoverPanic(c *gin.Context, recovered interface{}) {
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}

	slog.Error(
		"panic_recovered",
		slog.String("method", c.Request.Method),
		slog.String("path", path),
		slog.Any("panic", recovered),
		slog.String("stack", string(debug.Stack())),
	)

	response.InternalError(c, PublicInternalErrorMessage)
	c.Abort()
}
