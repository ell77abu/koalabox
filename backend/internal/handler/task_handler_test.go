package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"task-backend/internal/middleware"
	"task-backend/internal/model"
	"task-backend/internal/service"
)

type taskServiceStub struct{}

func (s *taskServiceStub) List(ctx context.Context, userID uint64, userLevel uint8, page, size int, teamID *uint64, scope service.TaskScope) ([]model.Task, int64, int, int, error) {
	return nil, 0, 0, 0, service.ErrInvalidTaskScope
}
func (s *taskServiceStub) Create(ctx context.Context, userID uint64, req service.CreateTaskRequest) (*model.Task, error) {
	panic("unexpected call")
}
func (s *taskServiceStub) Update(ctx context.Context, userID uint64, taskID uint64, req service.UpdateTaskRequest) (*model.Task, error) {
	panic("unexpected call")
}
func (s *taskServiceStub) Delete(ctx context.Context, userID uint64, userLevel uint8, taskID uint64) error {
	panic("unexpected call")
}
func (s *taskServiceStub) ListPendingAssignments(ctx context.Context, userID uint64) ([]model.TaskAssignmentRequestDetail, error) {
	panic("unexpected call")
}
func (s *taskServiceStub) AcceptAssignment(ctx context.Context, userID, requestID uint64) error {
	panic("unexpected call")
}
func (s *taskServiceStub) DeclineAssignment(ctx context.Context, userID, requestID uint64) error {
	panic("unexpected call")
}

func TestTaskHandlerListRejectsInvalidScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewTaskHandler(&taskServiceStub{})
	r.GET("/tasks", func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, uint64(1))
		c.Set(middleware.ContextUserLevelKey, uint8(0))
		handler.List(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/tasks?task_scope=weird", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
