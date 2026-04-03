package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"task-backend/internal/repository"
	"task-backend/internal/service"
	"task-backend/pkg/response"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register 处理注册请求，并将 Service 错误映射到 HTTP 响应。
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	err := h.authService.Register(c.Request.Context(), req)
	if errors.Is(err, service.ErrInvalidUsername) || errors.Is(err, service.ErrInvalidPassword) {
		response.BadRequest(c, err.Error())
		return
	}
	if errors.Is(err, repository.ErrConflict) {
		response.Conflict(c, "name already exists")
		return
	}
	if err != nil {
		response.InternalError(c, "register failed")
		return
	}

	response.JSON(c, http.StatusCreated, 0, "register success", gin.H{
		"username": req.Username,
	})
}

// Login 校验登录凭证，成功后返回 JWT 相关数据。
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	data, err := h.authService.Login(c.Request.Context(), req)
	if errors.Is(err, service.ErrInvalidCredentials) {
		response.Unauthorized(c, "invalid username or password")
		return
	}
	if err != nil {
		response.InternalError(c, "login failed")
		return
	}

	response.Success(c, "login success", data)
}

// ChangePassword 允许已登录用户修改自己的密码。
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}
	userID, ok := userIDValue.(uint64)
	if !ok {
		response.Unauthorized(c, "invalid auth context")
		return
	}

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), userID, req)
	if errors.Is(err, service.ErrInvalidPassword) {
		response.BadRequest(c, "invalid new password")
		return
	}
	if errors.Is(err, service.ErrInvalidOldPassword) {
		response.BadRequest(c, "invalid old password")
		return
	}
	if errors.Is(err, service.ErrInvalidCredentials) {
		response.Unauthorized(c, "unauthorized")
		return
	}
	if err != nil {
		response.InternalError(c, "change password failed")
		return
	}

	response.Success(c, "change password success", nil)
}
