package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"task-backend/internal/service"
	"task-backend/pkg/response"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

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
	if errors.Is(err, service.ErrUsernameExists) {
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
