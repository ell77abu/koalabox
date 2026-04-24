package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"task-backend/internal/config"
	"task-backend/internal/handler"
	"task-backend/internal/middleware"
	"task-backend/internal/repository"
	"task-backend/internal/service"
)

func NewRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	healthHandler := handler.NewHealthHandler()
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	taskService := service.NewTaskService(taskRepo, teamRepo, userRepo)
	teamService := service.NewTeamService(teamRepo, userRepo)
	authHandler := handler.NewAuthHandler(authService)
	taskHandler := handler.NewTaskHandler(taskService)
	teamHandler := handler.NewTeamHandler(teamService)

	r.GET("/ping", healthHandler.Ping)

	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/ping", healthHandler.Ping)

		authGroup := apiV1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.PUT("/password", middleware.JWTAuth(cfg.JWT.Secret), authHandler.ChangePassword)
		}

		taskGroup := apiV1.Group("/tasks")
		taskGroup.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			taskGroup.GET("", taskHandler.List)
			taskGroup.POST("", taskHandler.Create)
			taskGroup.PUT("/:id", taskHandler.Update)
			taskGroup.PATCH("/:id", taskHandler.Update)
			taskGroup.DELETE("/:id", taskHandler.Delete)
		}

		teamGroup := apiV1.Group("/teams")
		teamGroup.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			teamGroup.GET("", teamHandler.ListMine)
			teamGroup.GET("/mine", teamHandler.ListMine)
			teamGroup.POST("", teamHandler.Create)
			teamGroup.POST("/:team_id/members", teamHandler.AddMember)
			teamGroup.GET("/:team_id/members", teamHandler.ListMembers)
			teamGroup.DELETE("/:team_id/members/:user_id", teamHandler.RemoveMember)
			teamGroup.PUT("/:team_id/members/:user_id", teamHandler.UpdateMemberRole)
			teamGroup.PATCH("/:team_id/members/:user_id", teamHandler.UpdateMemberRole)
		}
	}

	return r
}
