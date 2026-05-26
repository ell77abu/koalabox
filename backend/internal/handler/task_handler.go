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

type addTaskDependencyRequest struct {
	PredecessorID uint64 `json:"predecessor_id"`
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
	var unresolvedErr *service.UnresolvedDependenciesError
	if errors.As(err, &unresolvedErr) {
		response.JSON(c, 409, 409, "操作失败：尚有前置任务未完成", gin.H{
			"unresolved_predecessor_ids": unresolvedErr.IDs,
			"unresolved_predecessors":    unresolvedErr.Predecessors,
		})
		return
	}
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

// AddDependency 为任务添加前置依赖。
func (h *TaskHandler) AddDependency(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid task id")
		return
	}

	var req addTaskDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	dep, err := h.taskService.AddDependency(c.Request.Context(), userID, taskID, req.PredecessorID)
	if errors.Is(err, service.ErrDependencyCycle) {
		response.BadRequest(c, "无法添加依赖：检测到依赖循环")
		return
	}
	if errors.Is(err, service.ErrCrossTaskDependency) {
		response.Forbidden(c, "跨团队或跨用户任务禁止建立依赖关系")
		return
	}
	if errors.Is(err, service.ErrDependencyOnDoneTask) {
		response.BadRequest(c, "已完成任务不能再设置前置依赖")
		return
	}
	if errors.Is(err, service.ErrTeamGone) {
		response.Gone(c, "team has been deleted")
		return
	}
	if isTaskValidationError(err) {
		response.BadRequest(c, err.Error())
		return
	}
	if isTaskForbiddenError(err) || errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "task not found")
		return
	}
	if errors.Is(err, repository.ErrConflict) {
		response.Conflict(c, "dependency already exists")
		return
	}
	if err != nil {
		response.InternalError(c, "add dependency failed")
		return
	}

	response.Created(c, "依赖添加成功", dep)
}

// RemoveDependency 移除任务的前置依赖。
func (h *TaskHandler) RemoveDependency(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid task id")
		return
	}

	predecessorID, err := strconv.ParseUint(c.Param("predecessor_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid predecessor id")
		return
	}

	err = h.taskService.RemoveDependency(c.Request.Context(), userID, taskID, predecessorID)
	if errors.Is(err, service.ErrCrossTaskDependency) {
		response.Forbidden(c, "跨团队或跨用户任务禁止建立依赖关系")
		return
	}
	if errors.Is(err, service.ErrTeamGone) {
		response.Gone(c, "team has been deleted")
		return
	}
	if isTaskValidationError(err) {
		response.BadRequest(c, err.Error())
		return
	}
	if errors.Is(err, service.ErrTeamGone) {
		response.Gone(c, "team has been deleted")
		return
	}
	if isTaskForbiddenError(err) || errors.Is(err, repository.ErrForbidden) {
		response.Forbidden(c, "forbidden")
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(c, "dependency not found")
		return
	}
	if err != nil {
		response.InternalError(c, "remove dependency failed")
		return
	}

	response.Success(c, "依赖移除成功", nil)
}

// GetDependencies 查询任务的前置与后继依赖。
func (h *TaskHandler) GetDependencies(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid task id")
		return
	}

	data, err := h.taskService.GetDependencies(c.Request.Context(), userID, taskID)
	if isTaskValidationError(err) {
		response.BadRequest(c, err.Error())
		return
	}
	if errors.Is(err, service.ErrTeamGone) {
		response.Gone(c, "team has been deleted")
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
		response.InternalError(c, "query dependencies failed")
		return
	}

	response.Success(c, "query success", data)
}

// GetDependencyGraph 查询任务依赖图的所有祖先、后代节点与依赖边。
func (h *TaskHandler) GetDependencyGraph(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid task id")
		return
	}

	data, err := h.taskService.GetDependencyGraph(c.Request.Context(), userID, taskID)
	if isTaskValidationError(err) {
		response.BadRequest(c, err.Error())
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
		response.InternalError(c, "query dependency graph failed")
		return
	}

	response.Success(c, "query success", data)
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
		errors.Is(err, service.ErrTeamIDImmutable) ||
		errors.Is(err, service.ErrInvalidAssigneeID) ||
		errors.Is(err, service.ErrAssigneeNotInTeam) ||
		errors.Is(err, service.ErrInvalidTaskDependency) ||
		errors.Is(err, service.ErrDependencyOnDoneTask) ||
		errors.Is(err, service.ErrDependenciesNotDone)
}

func isTaskForbiddenError(err error) bool {
	return errors.Is(err, service.ErrTeamRoleDenied) ||
		errors.Is(err, service.ErrMemberStatusOnly)
}
