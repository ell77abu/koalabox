package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"task-backend/pkg/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Ping(c *gin.Context) {
	response.JSON(c, http.StatusOK, 0, "ok", gin.H{
		"service": "task-backend",
	})
}
