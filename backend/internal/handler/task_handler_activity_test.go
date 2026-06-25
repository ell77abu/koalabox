package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-backend/internal/middleware"
	"task-backend/internal/model"
	"task-backend/internal/repository"
	"task-backend/internal/service"
	"task-backend/pkg/response"
)

type taskLogServiceMock struct {
	listActivityLogsFn  func(ctx context.Context, userID, taskID uint64, page, size int) ([]service.TaskActivityLogItem, int64, int, int, error)
	listTrashFn         func(ctx context.Context, userID uint64, page, size int, teamID *uint64) ([]model.Task, int64, int, int, error)
	dismissAssignmentFn func(ctx context.Context, userID, requestID uint64) (*service.AssignmentDismissResult, error)
}

func (m *taskLogServiceMock) List(ctx context.Context, userID uint64, userLevel uint8, page, size int, teamID *uint64) ([]model.Task, int64, int, int, error) {
	panic("unexpected call to List")
}

func (m *taskLogServiceMock) ListTrash(ctx context.Context, userID uint64, page, size int, teamID *uint64) ([]model.Task, int64, int, int, error) {
	if m.listTrashFn == nil {
		panic("unexpected call to ListTrash")
	}
	return m.listTrashFn(ctx, userID, page, size, teamID)
}

func (m *taskLogServiceMock) Create(ctx context.Context, userID uint64, req service.CreateTaskRequest) (*model.Task, error) {
	panic("unexpected call to Create")
}

func (m *taskLogServiceMock) Update(ctx context.Context, userID uint64, taskID uint64, req service.UpdateTaskRequest) (*model.Task, error) {
	panic("unexpected call to Update")
}

func (m *taskLogServiceMock) Delete(ctx context.Context, userID uint64, userLevel uint8, taskID uint64) error {
	panic("unexpected call to Delete")
}

func (m *taskLogServiceMock) ListPendingAssignments(ctx context.Context, userID uint64) ([]model.TaskAssignmentRequestDetail, error) {
	panic("unexpected call to ListPendingAssignments")
}

func (m *taskLogServiceMock) AcceptAssignment(ctx context.Context, userID, requestID uint64) error {
	panic("unexpected call to AcceptAssignment")
}

func (m *taskLogServiceMock) DeclineAssignment(ctx context.Context, userID, requestID uint64) error {
	panic("unexpected call to DeclineAssignment")
}

func (m *taskLogServiceMock) DismissAssignment(ctx context.Context, userID, requestID uint64) (*service.AssignmentDismissResult, error) {
	if m.dismissAssignmentFn == nil {
		panic("unexpected call to DismissAssignment")
	}
	return m.dismissAssignmentFn(ctx, userID, requestID)
}

func (m *taskLogServiceMock) AddDependency(ctx context.Context, userID, taskID, predecessorID uint64) (*model.TaskDependency, error) {
	panic("unexpected call to AddDependency")
}

func (m *taskLogServiceMock) RemoveDependency(ctx context.Context, userID, taskID, predecessorID uint64) error {
	panic("unexpected call to RemoveDependency")
}

func (m *taskLogServiceMock) GetDependencies(ctx context.Context, userID, taskID uint64) (*service.TaskDependencies, error) {
	panic("unexpected call to GetDependencies")
}

func (m *taskLogServiceMock) GetDependencyGraph(ctx context.Context, userID, taskID uint64) (*service.TaskDependencyGraph, error) {
	panic("unexpected call to GetDependencyGraph")
}

func (m *taskLogServiceMock) ListActivityLogs(ctx context.Context, userID, taskID uint64, page, size int) ([]service.TaskActivityLogItem, int64, int, int, error) {
	if m.listActivityLogsFn == nil {
		panic("unexpected call to ListActivityLogs")
	}
	return m.listActivityLogsFn(ctx, userID, taskID, page, size)
}

func TestTaskHandlerListActivityLogsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewTaskHandler(&taskLogServiceMock{
		listActivityLogsFn: func(ctx context.Context, userID, taskID uint64, page, size int) ([]service.TaskActivityLogItem, int64, int, int, error) {
			assert.Equal(t, uint64(7), userID)
			assert.Equal(t, uint64(10), taskID)
			assert.Equal(t, 1, page)
			assert.Equal(t, 10, size)
			return []service.TaskActivityLogItem{
				{
					ID:         101,
					ActorName:  "alice",
					ActionType: model.ActivityActionCreateTask,
					Content:    "创建了任务「实现登录」",
					CreatedAt:  "2026-06-01 12:00:00",
				},
			}, 1, 1, 10, nil
		},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, uint64(7))
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/10/logs?page=1&size=10", nil)

	handler.ListActivityLogs(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp response.Common
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "query success", resp.Message)
}

func TestTaskHandlerListActivityLogsUsesDefaultPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewTaskHandler(&taskLogServiceMock{
		listActivityLogsFn: func(ctx context.Context, userID, taskID uint64, page, size int) ([]service.TaskActivityLogItem, int64, int, int, error) {
			assert.Equal(t, uint64(7), userID)
			assert.Equal(t, uint64(10), taskID)
			assert.Equal(t, 1, page)
			assert.Equal(t, 10, size)
			return []service.TaskActivityLogItem{}, 0, 1, 10, nil
		},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, uint64(7))
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/10/logs", nil)

	handler.ListActivityLogs(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp response.Common
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "query success", resp.Message)
}

func TestTaskHandlerListActivityLogsInvalidTaskID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewTaskHandler(&taskLogServiceMock{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, uint64(7))
	c.Params = gin.Params{{Key: "id", Value: "bad-id"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/bad-id/logs", nil)

	handler.ListActivityLogs(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp response.Common
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "invalid task id", resp.Message)
}

func TestTaskHandlerListActivityLogsUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewTaskHandler(&taskLogServiceMock{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/10/logs", nil)

	handler.ListActivityLogs(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	var resp response.Common
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "unauthorized", resp.Message)
}

func TestTaskHandlerListActivityLogsForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewTaskHandler(&taskLogServiceMock{
		listActivityLogsFn: func(ctx context.Context, userID, taskID uint64, page, size int) ([]service.TaskActivityLogItem, int64, int, int, error) {
			return nil, 0, 0, 0, repository.ErrForbidden
		},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, uint64(7))
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/10/logs", nil)

	handler.ListActivityLogs(c)

	require.Equal(t, http.StatusForbidden, w.Code)
	var resp response.Common
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 403, resp.Code)
	assert.Equal(t, "forbidden", resp.Message)
}

func TestTaskHandlerListActivityLogsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewTaskHandler(&taskLogServiceMock{
		listActivityLogsFn: func(ctx context.Context, userID, taskID uint64, page, size int) ([]service.TaskActivityLogItem, int64, int, int, error) {
			return nil, 0, 0, 0, repository.ErrNotFound
		},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, uint64(7))
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/10/logs", nil)

	handler.ListActivityLogs(c)

	require.Equal(t, http.StatusNotFound, w.Code)
	var resp response.Common
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 404, resp.Code)
	assert.Equal(t, "task not found", resp.Message)
}

func TestTaskHandlerListTrashRejectsInvalidTeamID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewTaskHandler(&taskLogServiceMock{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, uint64(7))
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/trash?team_id=abc", nil)

	handler.ListTrash(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp response.Common
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "invalid team_id", resp.Message)
}

func TestTaskHandlerListTrashPassesTeamScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewTaskHandler(&taskLogServiceMock{
		listTrashFn: func(ctx context.Context, userID uint64, page, size int, teamID *uint64) ([]model.Task, int64, int, int, error) {
			assert.Equal(t, uint64(7), userID)
			assert.Equal(t, 2, page)
			assert.Equal(t, 20, size)
			require.NotNil(t, teamID)
			assert.Equal(t, uint64(5), *teamID)
			return []model.Task{{ID: 88, Title: "deleted"}}, 1, page, size, nil
		},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, uint64(7))
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/trash?page=2&size=20&team_id=5", nil)

	handler.ListTrash(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp response.Common
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "query success", resp.Message)
}

func TestTaskHandlerDismissAssignmentSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dismissedAt := time.Date(2026, 6, 5, 12, 30, 0, 0, time.UTC)
	handler := NewTaskHandler(&taskLogServiceMock{
		dismissAssignmentFn: func(ctx context.Context, userID, requestID uint64) (*service.AssignmentDismissResult, error) {
			assert.Equal(t, uint64(7), userID)
			assert.Equal(t, uint64(21), requestID)
			return &service.AssignmentDismissResult{RequestID: requestID, DismissAt: dismissedAt}, nil
		},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, uint64(7))
	c.Params = gin.Params{{Key: "request_id", Value: "21"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/assignment-requests/21/dismiss", nil)

	handler.DismissAssignment(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp response.Common
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "assignment notification dismissed", resp.Message)
}

func TestTaskHandlerDismissAssignmentForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewTaskHandler(&taskLogServiceMock{
		dismissAssignmentFn: func(ctx context.Context, userID, requestID uint64) (*service.AssignmentDismissResult, error) {
			return nil, repository.ErrForbidden
		},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, uint64(7))
	c.Params = gin.Params{{Key: "request_id", Value: "21"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/assignment-requests/21/dismiss", nil)

	handler.DismissAssignment(c)

	require.Equal(t, http.StatusForbidden, w.Code)
	var resp response.Common
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 403, resp.Code)
	assert.Equal(t, "forbidden", resp.Message)
}
