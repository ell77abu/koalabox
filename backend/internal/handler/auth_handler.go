package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"task-backend/internal/apperror"
	"task-backend/internal/middleware"
	"task-backend/internal/service"
	"task-backend/pkg/response"
)

const sessionCookieName = "task_session"

type AuthHandler struct {
	authService       service.AuthService
	sessionExpireHour int
}

func NewAuthHandler(authService service.AuthService, sessionExpireHour int) *AuthHandler {
	return &AuthHandler{authService: authService, sessionExpireHour: sessionExpireHour}
}

// Register 处理注册请求，并将 Service 错误映射到 HTTP 响应。
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("请求体格式不正确"), "auth register")
		return
	}

	err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		apperror.Respond(c, err, "auth register")
		return
	}

	response.JSON(c, http.StatusCreated, 0, "register success", gin.H{
		"username": req.Username,
	})
}

// Login 校验登录凭证，成功后通过 HttpOnly Cookie 维护会话。
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("请求体格式不正确"), "auth login")
		return
	}

	data, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		apperror.Respond(c, err, "auth login")
		return
	}

	maxAge := int((time.Duration(h.sessionExpireHour) * time.Hour).Seconds())
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, data.Token, maxAge, "/", "", false, true)

	response.Success(c, "login success", gin.H{
		"username": data.Username,
		"level":    data.Level,
	})
}

// Session 返回当前登录会话信息。
func (h *AuthHandler) Session(c *gin.Context) {
	usernameValue, exists := c.Get(middleware.ContextUsernameKey)
	if !exists {
		apperror.Respond(c, apperror.Unauthorized("登录状态已失效，请重新登录"), "auth session")
		return
	}
	username, ok := usernameValue.(string)
	if !ok || username == "" {
		apperror.Respond(c, apperror.Unauthorized("登录状态无效，请重新登录"), "auth session")
		return
	}

	levelValue, exists := c.Get(middleware.ContextUserLevelKey)
	if !exists {
		apperror.Respond(c, apperror.Unauthorized("登录状态已失效，请重新登录"), "auth session")
		return
	}
	level, ok := levelValue.(uint8)
	if !ok {
		apperror.Respond(c, apperror.Unauthorized("登录状态无效，请重新登录"), "auth session")
		return
	}

	response.Success(c, "session active", gin.H{
		"username": username,
		"level":    level,
	})
}

// Logout 清理会话 Cookie。
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, "", -1, "/", "", false, true)
	response.Success(c, "logout success", nil)
}

// ChangePassword 允许已登录用户修改自己的密码。
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		apperror.Respond(c, apperror.Unauthorized("登录状态已失效，请重新登录"), "auth change password")
		return
	}
	userID, ok := userIDValue.(uint64)
	if !ok {
		apperror.Respond(c, apperror.Unauthorized("登录状态无效，请重新登录"), "auth change password")
		return
	}

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("请求体格式不正确"), "auth change password")
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), userID, req)
	if err != nil {
		apperror.Respond(c, err, "auth change password")
		return
	}

	response.Success(c, "change password success", nil)
}
