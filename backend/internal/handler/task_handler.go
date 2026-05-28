package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"task-backend/internal/apperror"
	"task-backend/internal/middleware"
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
			apperror.Respond(c, apperror.BadRequest("团队ID不合法"), "task list")
			return
		}
		teamID = &parsed
	}

	list, total, fixedPage, fixedSize, err := h.taskService.List(c.Request.Context(), userID, userLevel, page, size, teamID)
	if err != nil {
		apperror.Respond(c, err, "task list")
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
		apperror.Respond(c, apperror.BadRequest("请求体格式不正确"), "task create")
		return
	}

	task, err := h.taskService.Create(c.Request.Context(), userID, req)
	if err != nil {
		apperror.Respond(c, err, "task create")
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
		apperror.Respond(c, apperror.BadRequest("任务ID不合法"), "task update")
		return
	}

	var req service.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("请求体格式不正确"), "task update")
		return
	}

	task, err := h.taskService.Update(c.Request.Context(), userID, id, req)
	if err != nil {
		apperror.Respond(c, err, "task update")
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
		apperror.Respond(c, apperror.BadRequest("任务ID不合法"), "task delete")
		return
	}

	err = h.taskService.Delete(c.Request.Context(), userID, userLevel, id)
	if err != nil {
		apperror.Respond(c, err, "task delete")
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
		apperror.Respond(c, err, "assignment list pending")
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
		apperror.Respond(c, apperror.BadRequest("请求ID不合法"), "assignment accept")
		return
	}

	err = h.taskService.AcceptAssignment(c.Request.Context(), userID, requestID)
	if err != nil {
		apperror.Respond(c, err, "assignment accept")
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
		apperror.Respond(c, apperror.BadRequest("请求ID不合法"), "assignment decline")
		return
	}

	err = h.taskService.DeclineAssignment(c.Request.Context(), userID, requestID)
	if err != nil {
		apperror.Respond(c, err, "assignment decline")
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
		apperror.Respond(c, apperror.BadRequest("任务ID不合法"), "dependency add")
		return
	}

	var req addTaskDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.BadRequest("请求体格式不正确"), "dependency add")
		return
	}

	dep, err := h.taskService.AddDependency(c.Request.Context(), userID, taskID, req.PredecessorID)
	if err != nil {
		apperror.Respond(c, err, "dependency add")
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
		apperror.Respond(c, apperror.BadRequest("任务ID不合法"), "dependency remove")
		return
	}

	predecessorID, err := strconv.ParseUint(c.Param("predecessor_id"), 10, 64)
	if err != nil {
		apperror.Respond(c, apperror.BadRequest("前置任务ID不合法"), "dependency remove")
		return
	}

	err = h.taskService.RemoveDependency(c.Request.Context(), userID, taskID, predecessorID)
	if err != nil {
		apperror.Respond(c, err, "dependency remove")
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
		apperror.Respond(c, apperror.BadRequest("任务ID不合法"), "dependency list")
		return
	}

	data, err := h.taskService.GetDependencies(c.Request.Context(), userID, taskID)
	if err != nil {
		apperror.Respond(c, err, "dependency list")
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
		apperror.Respond(c, apperror.BadRequest("任务ID不合法"), "dependency graph")
		return
	}

	data, err := h.taskService.GetDependencyGraph(c.Request.Context(), userID, taskID)
	if err != nil {
		apperror.Respond(c, err, "dependency graph")
		return
	}

	response.Success(c, "query success", data)
}

func getUserIDFromContext(c *gin.Context) (uint64, bool) {
	v, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		apperror.Respond(c, apperror.Unauthorized("登录状态已失效，请重新登录"), "context user id")
		return 0, false
	}
	userID, ok := v.(uint64)
	if !ok {
		apperror.Respond(c, apperror.Unauthorized("登录状态无效，请重新登录"), "context user id")
		return 0, false
	}
	return userID, true
}

func getUserLevelFromContext(c *gin.Context) (uint8, bool) {
	v, exists := c.Get(middleware.ContextUserLevelKey)
	if !exists {
		apperror.Respond(c, apperror.Unauthorized("登录状态已失效，请重新登录"), "context user level")
		return 0, false
	}
	level, ok := v.(uint8)
	if !ok {
		apperror.Respond(c, apperror.Unauthorized("登录状态无效，请重新登录"), "context user level")
		return 0, false
	}
	return level, true
}
