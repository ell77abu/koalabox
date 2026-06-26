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
	// 全局使用自定义恢复中间件，记录 panic 堆栈并对外返回统一 500 错误
	r.Use(gin.Logger(), middleware.RecoverWithLog())

	healthHandler := handler.NewHealthHandler()
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	teamInviteRepo := repository.NewTeamInvitationRepository(db)
	taskAssignRepo := repository.NewTaskAssignmentRepository(db)
	activityLogRepo := repository.NewActivityLogRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	taskService := service.NewTaskService(taskRepo, teamRepo, userRepo, taskAssignRepo, activityLogRepo)
	teamService := service.NewTeamServiceWithTaskRepo(teamRepo, userRepo, teamInviteRepo, taskRepo, activityLogRepo)
	authHandler := handler.NewAuthHandler(authService, cfg.JWT.ExpireHour)
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
			authGroup.GET("/session", middleware.JWTAuth(cfg.JWT.Secret), authHandler.Session)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.PUT("/password", middleware.JWTAuth(cfg.JWT.Secret), authHandler.ChangePassword)
		}

		taskGroup := apiV1.Group("/tasks")
		taskGroup.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			taskGroup.GET("", taskHandler.List)
			taskGroup.GET("/trash", taskHandler.ListTrash)
			taskGroup.POST("", taskHandler.Create)
			taskGroup.PUT("/:id", taskHandler.Update)
			taskGroup.PATCH("/:id", taskHandler.Update)
			taskGroup.DELETE("/:id", taskHandler.Delete)
			taskGroup.GET("/:id/logs", taskHandler.ListActivityLogs)
			taskGroup.POST("/:id/dependencies", taskHandler.AddDependency)
			taskGroup.DELETE("/:id/dependencies/:predecessor_id", taskHandler.RemoveDependency)
			taskGroup.GET("/:id/dependencies", taskHandler.GetDependencies)
			taskGroup.GET("/:id/dependency-graph", taskHandler.GetDependencyGraph)
			taskGroup.GET("/assignments/pending", taskHandler.ListPendingAssignments)
			taskGroup.POST("/assignments/:request_id/accept", taskHandler.AcceptAssignment)
			taskGroup.POST("/assignments/:request_id/decline", taskHandler.DeclineAssignment)
		}

		assignmentGroup := apiV1.Group("/assignment-requests")
		assignmentGroup.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			assignmentGroup.GET("", taskHandler.ListPendingAssignments)
			assignmentGroup.POST("/:request_id/accept", taskHandler.AcceptAssignment)
			assignmentGroup.POST("/:request_id/decline", taskHandler.DeclineAssignment)
			assignmentGroup.POST("/:request_id/dismiss", taskHandler.DismissAssignment)
		}

		teamGroup := apiV1.Group("/teams")
		teamGroup.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			teamGroup.GET("", teamHandler.ListMine)
			teamGroup.GET("/mine", teamHandler.ListMine)
			teamGroup.POST("", teamHandler.Create)
			teamGroup.PUT("/:team_id", teamHandler.Update)
			teamGroup.DELETE("/:team_id", teamHandler.Delete)
			teamGroup.POST("/:team_id/leave", teamHandler.Leave)
			teamGroup.POST("/:team_id/transfer", teamHandler.TransferOwnership)
			teamGroup.POST("/:team_id/members", teamHandler.AddMember)
			teamGroup.POST("/:team_id/invitations", teamHandler.AddMember)
			teamGroup.GET("/:team_id/members", teamHandler.ListMembers)
			teamGroup.DELETE("/:team_id/members/:user_id", teamHandler.RemoveMember)
			teamGroup.PUT("/:team_id/members/:user_id", teamHandler.UpdateMemberRole)
			teamGroup.PATCH("/:team_id/members/:user_id", teamHandler.UpdateMemberRole)
			teamGroup.GET("/invitations/pending", teamHandler.ListPendingInvitations)
			teamGroup.POST("/invitations/:invitation_id/accept", teamHandler.AcceptInvitation)
			teamGroup.POST("/invitations/:invitation_id/decline", teamHandler.DeclineInvitation)
		}

		invitationGroup := apiV1.Group("/invitations")
		invitationGroup.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			invitationGroup.GET("", teamHandler.ListPendingInvitations)
			invitationGroup.POST("/:invitation_id/accept", teamHandler.AcceptInvitation)
			invitationGroup.POST("/:invitation_id/decline", teamHandler.DeclineInvitation)
		}
	}

	return r
}
