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
	ErrInvalidTeamID       = errors.New("invalid team id")
	ErrInvalidAssigneeID   = errors.New("invalid assignee id")
	ErrAssigneeNotInTeam   = errors.New("assignee is not in team")
	ErrTeamRoleDenied      = errors.New("insufficient team role permission")
	ErrMemberStatusOnly    = errors.New("member can only update status for assigned task")
)

type CreateTaskRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    int     `json:"priority"`
	Deadline    string  `json:"deadline"`
	TeamID      *uint64 `json:"team_id"`
	AssigneeID  *uint64 `json:"assignee_id"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *int    `json:"priority"`
	Deadline    *string `json:"deadline"`
	TeamID      *uint64 `json:"team_id"`
	AssigneeID  *uint64 `json:"assignee_id"`
}

type TaskService interface {
	List(ctx context.Context, userID uint64, userLevel uint8, page, size int, teamID *uint64) ([]model.Task, int64, int, int, error)
	Create(ctx context.Context, userID uint64, req CreateTaskRequest) (*model.Task, error)
	Update(ctx context.Context, userID uint64, taskID uint64, req UpdateTaskRequest) (*model.Task, error)
	Delete(ctx context.Context, userID uint64, userLevel uint8, taskID uint64) error
}

type taskService struct {
	taskRepo repository.TaskRepository
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
}

func NewTaskService(taskRepo repository.TaskRepository, teamRepo repository.TeamRepository, userRepo repository.UserRepository) TaskService {
	return &taskService{taskRepo: taskRepo, teamRepo: teamRepo, userRepo: userRepo}
}

func (s *taskService) List(ctx context.Context, userID uint64, userLevel uint8, page, size int, teamID *uint64) ([]model.Task, int64, int, int, error) {
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
		if teamID != nil {
			list, total, err = s.taskRepo.ListByTeamID(ctx, *teamID, page, size)
		} else {
			list, total, err = s.taskRepo.ListAll(ctx, page, size)
		}
	} else if teamID != nil {
		if _, err = s.teamRepo.GetByID(ctx, *teamID); err != nil {
			return nil, 0, 0, 0, ErrInvalidTeamID
		}

		isMember, memberErr := s.teamRepo.IsMember(ctx, *teamID, userID)
		if memberErr != nil {
			return nil, 0, 0, 0, memberErr
		}
		if !isMember {
			return nil, 0, 0, 0, repository.ErrForbidden
		}

		list, total, err = s.taskRepo.ListByTeamID(ctx, *teamID, page, size)
	} else {
		list, total, err = s.taskRepo.ListPersonalAndAssignedTeam(ctx, userID, page, size)
	}
	if err != nil {
		return nil, 0, 0, 0, err
	}
	if list == nil {
		list = make([]model.Task, 0)
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

	if req.TeamID != nil {
		if _, err := s.teamRepo.GetByID(ctx, *req.TeamID); err != nil {
			return nil, ErrInvalidTeamID
		}

		role, err := s.teamRepo.GetMemberRole(ctx, *req.TeamID, userID)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrForbidden
		}
		if err != nil {
			return nil, err
		}
		if !canManageTeamTasks(role) {
			return nil, ErrTeamRoleDenied
		}
	}

	if req.AssigneeID != nil {
		if _, err := s.userRepo.GetByID(ctx, *req.AssigneeID); err != nil {
			return nil, ErrInvalidAssigneeID
		}
		if req.TeamID != nil {
			inTeam, err := s.teamRepo.IsMember(ctx, *req.TeamID, *req.AssigneeID)
			if err != nil {
				return nil, err
			}
			if !inTeam {
				return nil, ErrAssigneeNotInTeam
			}
		}
	}

	task := &model.Task{
		UserID:      userID,
		TeamID:      req.TeamID,
		AssigneeID:  req.AssigneeID,
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

func (s *taskService) Update(ctx context.Context, userID uint64, taskID uint64, req UpdateTaskRequest) (*model.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if task.TeamID == nil {
		if task.UserID != userID {
			return nil, repository.ErrForbidden
		}
		return s.applyTaskUpdatesAndSave(ctx, task, req)
	}

	role, err := s.teamRepo.GetMemberRole(ctx, *task.TeamID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, repository.ErrForbidden
	}
	if err != nil {
		return nil, err
	}

	if canManageTeamTasks(role) {
		return s.applyTaskUpdatesAndSave(ctx, task, req)
	}

	if role != model.TeamRoleMember {
		return nil, ErrTeamRoleDenied
	}
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		return nil, repository.ErrForbidden
	}
	if req.Status == nil || hasMemberRestrictedFields(req) {
		return nil, ErrMemberStatusOnly
	}

	status := model.TaskStatus(*req.Status)
	if !isValidTaskStatus(status) {
		return nil, ErrInvalidTaskStatus
	}
	if task.Status == status {
		return task, nil
	}
	task.Status = status

	if err := s.taskRepo.UpdateByID(ctx, task); err != nil {
		return nil, err
	}
	return s.taskRepo.GetByID(ctx, taskID)
}

func (s *taskService) Delete(ctx context.Context, userID uint64, userLevel uint8, taskID uint64) error {
	if userLevel == 1 {
		return s.taskRepo.DeleteByID(ctx, taskID)
	}

	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	if task.TeamID == nil {
		if task.UserID != userID {
			return repository.ErrForbidden
		}
		return s.taskRepo.DeleteByID(ctx, taskID)
	}

	role, err := s.teamRepo.GetMemberRole(ctx, *task.TeamID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return repository.ErrForbidden
	}
	if err != nil {
		return err
	}
	if !canManageTeamTasks(role) {
		return ErrTeamRoleDenied
	}

	return s.taskRepo.DeleteByID(ctx, taskID)
}

func (s *taskService) applyTaskUpdatesAndSave(ctx context.Context, task *model.Task, req UpdateTaskRequest) (*model.Task, error) {
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
	if req.TeamID != nil {
		if _, err := s.teamRepo.GetByID(ctx, *req.TeamID); err != nil {
			return nil, ErrInvalidTeamID
		}
		task.TeamID = req.TeamID
	}
	if req.AssigneeID != nil {
		if _, err := s.userRepo.GetByID(ctx, *req.AssigneeID); err != nil {
			return nil, ErrInvalidAssigneeID
		}
		if task.TeamID != nil {
			inTeam, err := s.teamRepo.IsMember(ctx, *task.TeamID, *req.AssigneeID)
			if err != nil {
				return nil, err
			}
			if !inTeam {
				return nil, ErrAssigneeNotInTeam
			}
		}
		task.AssigneeID = req.AssigneeID
	}

	if err := s.taskRepo.UpdateByID(ctx, task); err != nil {
		return nil, err
	}
	return s.taskRepo.GetByID(ctx, task.ID)
}

func hasMemberRestrictedFields(req UpdateTaskRequest) bool {
	return req.Title != nil || req.Description != nil || req.Priority != nil || req.Deadline != nil || req.TeamID != nil || req.AssigneeID != nil
}

func canManageTeamTasks(role model.TeamRole) bool {
	return role == model.TeamRoleOwner || role == model.TeamRoleAdmin
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

	if t, err := time.Parse(time.RFC3339Nano, input); err == nil {
		return &t, nil
	}

	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04",
	}

	for _, layout := range formats {
		if t, err := time.ParseInLocation(layout, input, time.Local); err == nil {
			return &t, nil
		}
	}

	return nil, ErrInvalidTaskDeadline
}
