package handler

import (
	"errors"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"

	"task-backend/internal/middleware"
	"task-backend/internal/repository"
	"task-backend/internal/service"
	"task-backend/pkg/response"
)

type TaskHandler struct {
	taskService service.TaskService
}

func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

//支持排序
func (h *TaskHandler) List(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}
	userLevel, ok := getUserLevelFromContext(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	// 新增搜索 + 排序参数
	search := strings.TrimSpace(c.DefaultQuery("search", ""))
	sortField := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort", "created_at")))
	orderDir := strings.ToLower(strings.TrimSpace(c.DefaultQuery("order", "desc")))

	// 规范化 orderDir
	if orderDir != "asc" && orderDir != "desc" {
		orderDir = "desc"
	}

	list, total, fixedPage, fixedSize, err := h.taskService.List(
		c.Request.Context(),
		userID,
		userLevel,
		page,
		size,
		search,     // 新增
		sortField,
		orderDir,
	)
	if err != nil {
		response.InternalError(c, "query tasks failed")
		return
	}

	response.SuccessPage(c, "query success", list, fixedPage, fixedSize, total)
}

func (h *TaskHandler) Create(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	var req service.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	task, err := h.taskService.Create(c.Request.Context(), userID, req)
	if isTaskValidationError(err) {
		response.BadRequest(c, err.Error())
		return
	}
	if err != nil {
		response.InternalError(c, "create task failed")
		return
	}

	response.Created(c, "create success", task)
}

func (h *TaskHandler) Update(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}
	userLevel, ok := getUserLevelFromContext(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid task id")
		return
	}

	var req service.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	task, err := h.taskService.Update(c.Request.Context(), userID, userLevel, id, req)
	if isTaskValidationError(err) {
		response.BadRequest(c, err.Error())
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.Forbidden(c, "You dont have the right to modify others task or task not found")
		return
	}
	if err != nil {
		response.InternalError(c, "update task failed")
		return
	}

	response.Success(c, "update success", task)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}
	userLevel, ok := getUserLevelFromContext(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid task id")
		return
	}

	err = h.taskService.Delete(c.Request.Context(), userID, userLevel, id)
	if errors.Is(err, repository.ErrNotFound) {
		response.Forbidden(c, "cannot delete others task or task not found")
		return
	}
	if err != nil {
		response.InternalError(c, "delete task failed")
		return
	}

	response.Success(c, "delete success", nil)
}

func getUserIDFromContext(c *gin.Context) (uint64, bool) {
	v, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return 0, false
	}
	userID, ok := v.(uint64)
	if !ok {
		response.Unauthorized(c, "invalid auth context")
		return 0, false
	}
	return userID, true
}

func getUserLevelFromContext(c *gin.Context) (uint8, bool) {
	v, exists := c.Get(middleware.ContextUserLevelKey)
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return 0, false
	}
	level, ok := v.(uint8)
	if !ok {
		response.Unauthorized(c, "invalid auth context")
		return 0, false
	}
	return level, true
}

func isTaskValidationError(err error) bool {
	return errors.Is(err, service.ErrInvalidTaskTitle) ||
		errors.Is(err, service.ErrInvalidTaskStatus) ||
		errors.Is(err, service.ErrInvalidTaskPriority) ||
		errors.Is(err, service.ErrInvalidTaskDeadline)
}
