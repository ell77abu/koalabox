package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"task-backend/internal/handler"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	_ = db

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	healthHandler := handler.NewHealthHandler()
	r.GET("/ping", healthHandler.Ping)

	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/ping", healthHandler.Ping)
	}

	return r
}
