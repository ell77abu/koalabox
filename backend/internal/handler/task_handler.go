package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"task-backend/internal/middleware"
	"task-backend/internal/repository"
	"task-backend/internal/service"
	"task-backend/pkg/response"
)

type TaskHandler struct {
	taskService service.TaskService
}

// NewTaskHandler 创建任务模块的 HTTP 处理器。
func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// List 查询当前用户可见的个人任务或团队任务列表。
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

	var teamID *uint64
	teamIDRaw := c.Query("team_id")
	if teamIDRaw != "" {
		parsed, err := strconv.ParseUint(teamIDRaw, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid team_id")
			return
		}
		teamID = &parsed
	}

	list, total, fixedPage, fixedSize, err := h.taskService.List(c.Request.Context(), userID, userLevel, page, size, teamID)
	if errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, service.ErrInvalidTeamID) {
		response.BadRequest(c, err.Error())
		return
	}
	if err != nil {
		response.InternalError(c, "query tasks failed")
		return
	}

	response.SuccessPage(c, "query success", list, fixedPage, fixedSize, total)
}

// Create 创建个人任务或团队任务，并在需要时生成任务分配通知。
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
	if isTaskForbiddenError(err) || errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrConflict) {
		response.Conflict(c, "assignment already pending")
		return
	}
	if err != nil {
		response.InternalError(c, "create task failed")
		return
	}

	response.Created(c, "create success", task)
}

// Update 更新任务内容，并根据用户角色限制团队成员可修改的字段范围。
func (h *TaskHandler) Update(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
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

	task, err := h.taskService.Update(c.Request.Context(), userID, id, req)
	if isTaskValidationError(err) {
		response.BadRequest(c, err.Error())
		return
	}
	if isTaskForbiddenError(err) {
		response.Forbidden(c, err.Error())
		return
	}
	if errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "task not found")
		return
	}
	if errors.Is(err, repository.ErrConflict) {
		response.Conflict(c, "assignment already pending")
		return
	}
	if err != nil {
		response.InternalError(c, "update task failed")
		return
	}

	response.Success(c, "update success", task)
}

// Delete 删除任务，并对普通用户执行任务归属或团队角色校验。
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
	if isTaskForbiddenError(err) {
		response.Forbidden(c, err.Error())
		return
	}
	if errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "task not found")
		return
	}
	if err != nil {
		response.InternalError(c, "delete task failed")
		return
	}

	response.Success(c, "delete success", nil)
}

// ListPendingAssignments 返回当前用户待处理的团队任务分配请求。
func (h *TaskHandler) ListPendingAssignments(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	list, err := h.taskService.ListPendingAssignments(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "query assignment requests failed")
		return
	}

	response.Success(c, "query success", list)
}

// AcceptAssignment 接受一个团队任务分配请求。
func (h *TaskHandler) AcceptAssignment(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	requestID, err := strconv.ParseUint(c.Param("request_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid request id")
		return
	}

	err = h.taskService.AcceptAssignment(c.Request.Context(), userID, requestID)
	if errors.Is(err, service.ErrAssignmentNotPending) {
		response.BadRequest(c, "assignment is not pending")
		return
	}
	if errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "assignment or task not found")
		return
	}
	if err != nil {
		response.InternalError(c, "accept assignment failed")
		return
	}

	response.Success(c, "assignment accepted", nil)
}

// DeclineAssignment 拒绝一个团队任务分配请求。
func (h *TaskHandler) DeclineAssignment(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	requestID, err := strconv.ParseUint(c.Param("request_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid request id")
		return
	}

	err = h.taskService.DeclineAssignment(c.Request.Context(), userID, requestID)
	if errors.Is(err, service.ErrAssignmentNotPending) {
		response.BadRequest(c, "assignment is not pending")
		return
	}
	if errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "assignment not found")
		return
	}
	if err != nil {
		response.InternalError(c, "decline assignment failed")
		return
	}

	response.Success(c, "assignment declined", nil)
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
		errors.Is(err, service.ErrInvalidTaskDeadline) ||
		errors.Is(err, service.ErrInvalidTeamID) ||
		errors.Is(err, service.ErrInvalidAssigneeID) ||
		errors.Is(err, service.ErrAssigneeNotInTeam)
}

func isTaskForbiddenError(err error) bool {
	return errors.Is(err, service.ErrTeamRoleDenied) ||
		errors.Is(err, service.ErrMemberStatusOnly)
}
