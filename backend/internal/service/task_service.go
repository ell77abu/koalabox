package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"task-backend/internal/model"
	"task-backend/internal/repository"
)

var (
	ErrInvalidTaskTitle    = errors.New("invalid task title")
	ErrInvalidTaskStatus   = errors.New("invalid task status")
	ErrInvalidTaskPriority = errors.New("invalid task priority")
	ErrInvalidTaskDeadline = errors.New("invalid task deadline format, expected YYYY-MM-DD HH:mm:ss")
)

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    int    `json:"priority"`
	Deadline    string `json:"deadline"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *int    `json:"priority"`
	Deadline    *string `json:"deadline"`
}

type TaskService interface {
	List(ctx context.Context, userID uint64, userLevel uint8, page, size int) ([]model.Task, int64, int, int, error)
	Create(ctx context.Context, userID uint64, req CreateTaskRequest) (*model.Task, error)
	Update(ctx context.Context, userID uint64, userLevel uint8, taskID uint64, req UpdateTaskRequest) (*model.Task, error)
	Delete(ctx context.Context, userID uint64, userLevel uint8, taskID uint64) error
}

type taskService struct {
	taskRepo repository.TaskRepository
}

func NewTaskService(taskRepo repository.TaskRepository) TaskService {
	return &taskService{taskRepo: taskRepo}
}

func (s *taskService) List(ctx context.Context, userID uint64, userLevel uint8, page, size int) ([]model.Task, int64, int, int, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	var (
		list  []model.Task
		total int64
		err   error
	)
	if userLevel == 1 {
		list, total, err = s.taskRepo.ListAll(ctx, page, size)
	} else {
		list, total, err = s.taskRepo.ListByUserID(ctx, userID, page, size)
	}
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return list, total, page, size, nil
}

func (s *taskService) Create(ctx context.Context, userID uint64, req CreateTaskRequest) (*model.Task, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" || len(title) > 120 {
		return nil, ErrInvalidTaskTitle
	}

	status := model.TaskStatusTODO
	if req.Status != "" {
		status = model.TaskStatus(req.Status)
	}
	if !isValidTaskStatus(status) {
		return nil, ErrInvalidTaskStatus
	}

	priority := model.PriorityMedium
	if req.Priority != 0 {
		priority = model.Priority(req.Priority)
	}
	if !isValidPriority(priority) {
		return nil, ErrInvalidTaskPriority
	}

	deadline, err := parseDeadline(req.Deadline)
	if err != nil {
		return nil, err
	}

	task := &model.Task{
		UserID:      userID,
		Title:       title,
		Description: req.Description,
		Status:      status,
		Priority:    priority,
		Deadline:    deadline,
	}
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskService) Update(ctx context.Context, userID uint64, userLevel uint8, taskID uint64, req UpdateTaskRequest) (*model.Task, error) {
	var (
		task *model.Task
		err  error
	)
	if userLevel == 1 {
		task, err = s.taskRepo.GetByID(ctx, taskID)
	} else {
		task, err = s.taskRepo.GetByIDAndUserID(ctx, taskID, userID)
	}
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" || len(title) > 120 {
			return nil, ErrInvalidTaskTitle
		}
		task.Title = title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		status := model.TaskStatus(*req.Status)
		if !isValidTaskStatus(status) {
			return nil, ErrInvalidTaskStatus
		}
		task.Status = status
	}
	if req.Priority != nil {
		priority := model.Priority(*req.Priority)
		if !isValidPriority(priority) {
			return nil, ErrInvalidTaskPriority
		}
		task.Priority = priority
	}
	if req.Deadline != nil {
		deadline, err := parseDeadline(*req.Deadline)
		if err != nil {
			return nil, err
		}
		task.Deadline = deadline
	}

	if userLevel == 1 {
		err = s.taskRepo.UpdateByID(ctx, task)
	} else {
		err = s.taskRepo.UpdateByIDAndUserID(ctx, task)
	}
	if err != nil {
		return nil, err
	}

	if userLevel == 1 {
		return s.taskRepo.GetByID(ctx, taskID)
	}

	updated, err := s.taskRepo.GetByIDAndUserID(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *taskService) Delete(ctx context.Context, userID uint64, userLevel uint8, taskID uint64) error {
	if userLevel == 1 {
		return s.taskRepo.DeleteByID(ctx, taskID)
	}
	return s.taskRepo.DeleteByIDAndUserID(ctx, taskID, userID)
}

func isValidTaskStatus(status model.TaskStatus) bool {
	switch status {
	case model.TaskStatusTODO, model.TaskStatusDOING, model.TaskStatusDONE, model.TaskStatusERROR:
		return true
	default:
		return false
	}
}

func isValidPriority(priority model.Priority) bool {
	return priority == model.PriorityLow || priority == model.PriorityMedium || priority == model.PriorityHigh
}

func parseDeadline(input string) (*time.Time, error) {
    input = strings.TrimSpace(input)
    if input == "" {
        return nil, nil
    }

    // 优先使用 Go 内置的 RFC3339Nano（支持毫秒、微秒、Z 或 ±时区）
    if t, err := time.Parse(time.RFC3339Nano, input); err == nil {
        return &t, nil
    }

    // 兼容旧格式（可选）
    formats := []string{
        "2006-01-02 15:04:05",
        "2006-01-02T15:04:05",
        "2006-01-02T15:04:05.999999999", // 带毫秒但不带 Z
        "2006-01-02T15:04",
    }

    for _, layout := range formats {
        if t, err := time.ParseInLocation(layout, input, time.Local); err == nil {
            return &t, nil
        }
    }

    return nil, ErrInvalidTaskDeadline
}