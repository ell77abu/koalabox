package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"task-backend/internal/model"
	"task-backend/internal/repository"
)

var (
	ErrInvalidTaskTitle      = errors.New("invalid task title")
	ErrInvalidTaskStatus     = errors.New("invalid task status")
	ErrInvalidTaskPriority   = errors.New("invalid task priority")
	ErrInvalidTaskDeadline   = errors.New("invalid task deadline format, expected YYYY-MM-DD HH:mm:ss")
	ErrInvalidTeamID         = errors.New("invalid team id")
	ErrInvalidAssigneeID     = errors.New("invalid assignee id")
	ErrAssigneeNotInTeam     = errors.New("assignee is not in team")
	ErrTeamRoleDenied        = errors.New("insufficient team role permission")
	ErrMemberStatusOnly      = errors.New("member can only update status for assigned task")
	ErrAssignmentNotPending  = errors.New("assignment request is not pending")
	ErrInvalidTaskDependency = errors.New("invalid task dependency")
	ErrDependencyOnDoneTask  = errors.New("cannot add dependency to a completed task")
	ErrCrossTaskDependency   = errors.New("cross team or cross user task dependency is forbidden")
	ErrDependenciesNotDone   = errors.New("predecessor tasks are not completed")
	ErrDependencyCycle       = errors.New("dependency cycle detected")
	ErrTeamIDImmutable       = errors.New("cannot change team id for existing task")
	ErrTeamGone              = errors.New("team is gone")
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

// OptionalUint64 用于区分字段未传、传 null、传具体数值三种状态。
type OptionalUint64 struct {
	Set   bool
	Valid bool
	Value uint64
}

func (o *OptionalUint64) UnmarshalJSON(data []byte) error {
	o.Set = true
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "null" {
		o.Valid = false
		o.Value = 0
		return nil
	}

	var v uint64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	o.Valid = true
	o.Value = v
	return nil
}

type UpdateTaskRequest struct {
	Title       *string        `json:"title"`
	Description *string        `json:"description"`
	Status      *string        `json:"status"`
	Priority    *int           `json:"priority"`
	Deadline    *string        `json:"deadline"`
	TeamID      *uint64        `json:"team_id"`
	AssigneeID  OptionalUint64 `json:"assignee_id"`
}

type AddTaskDependencyRequest struct {
	PredecessorID uint64 `json:"predecessor_id"`
}

type TaskDependencies struct {
	Predecessors []model.TaskDependencyNode `json:"predecessors"`
	Successors   []model.TaskDependencyNode `json:"successors"`
}

type TaskDependencyGraph struct {
	Nodes []model.TaskDependencyNode `json:"nodes"`
	Links []model.TaskDependencyLink `json:"links"`
}

type UnresolvedDependenciesError struct {
	IDs          []uint64
	Predecessors []model.TaskDependencyNode
}

func (e *UnresolvedDependenciesError) Error() string {
	return fmt.Sprintf("predecessor tasks are not completed: %v", e.IDs)
}

func (e *UnresolvedDependenciesError) Is(target error) bool {
	return target == ErrDependenciesNotDone
}

type TaskService interface {
	List(ctx context.Context, userID uint64, userLevel uint8, page, size int, teamID *uint64) ([]model.Task, int64, int, int, error)
	Create(ctx context.Context, userID uint64, req CreateTaskRequest) (*model.Task, error)
	Update(ctx context.Context, userID uint64, taskID uint64, req UpdateTaskRequest) (*model.Task, error)
	Delete(ctx context.Context, userID uint64, userLevel uint8, taskID uint64) error
	ListPendingAssignments(ctx context.Context, userID uint64) ([]model.TaskAssignmentRequestDetail, error)
	AcceptAssignment(ctx context.Context, userID, requestID uint64) error
	DeclineAssignment(ctx context.Context, userID, requestID uint64) error
	AddDependency(ctx context.Context, userID, taskID, predecessorID uint64) (*model.TaskDependency, error)
	RemoveDependency(ctx context.Context, userID, taskID, predecessorID uint64) error
	GetDependencies(ctx context.Context, userID, taskID uint64) (*TaskDependencies, error)
	GetDependencyGraph(ctx context.Context, userID, taskID uint64) (*TaskDependencyGraph, error)
}

type taskService struct {
	taskRepo   repository.TaskRepository
	teamRepo   repository.TeamRepository
	userRepo   repository.UserRepository
	assignRepo repository.TaskAssignmentRepository
}

func NewTaskService(taskRepo repository.TaskRepository, teamRepo repository.TeamRepository, userRepo repository.UserRepository, assignRepo repository.TaskAssignmentRepository) TaskService {
	return &taskService{taskRepo: taskRepo, teamRepo: teamRepo, userRepo: userRepo, assignRepo: assignRepo}
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
		if err := s.requireExistingTeamMember(ctx, *teamID, userID); err != nil {
			return nil, 0, 0, 0, err
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
		if err := s.requireTeamManager(ctx, *req.TeamID, userID); err != nil {
			return nil, err
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
	if err := s.createAssignmentNotification(ctx, task, userID, nil, nil); err != nil {
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
		return s.applyTaskUpdatesAndSave(ctx, task, req, userID)
	}

	role, err := s.teamRepo.GetMemberRole(ctx, *task.TeamID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, repository.ErrForbidden
	}
	if err != nil {
		return nil, err
	}

	if canManageTeamTasks(role) {
		return s.applyTaskUpdatesAndSave(ctx, task, req, userID)
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
	if status == model.TaskStatusDONE {
		if err := s.ensureDependenciesCompleted(ctx, task.ID); err != nil {
			return nil, err
		}
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

	if err := s.requireTeamManager(ctx, *task.TeamID, userID); err != nil {
		return err
	}

	return s.taskRepo.DeleteByID(ctx, taskID)
}

func (s *taskService) applyTaskUpdatesAndSave(ctx context.Context, task *model.Task, req UpdateTaskRequest, operatorID uint64) (*model.Task, error) {
	originalAssigneeID := task.AssigneeID
	originalTeamID := task.TeamID
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
		if status == model.TaskStatusDONE && task.Status != model.TaskStatusDONE {
			if err := s.ensureDependenciesCompleted(ctx, task.ID); err != nil {
				return nil, err
			}
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
		return nil, ErrTeamIDImmutable
	}
	if req.AssigneeID.Set {
		if !req.AssigneeID.Valid {
			task.AssigneeID = nil
		} else {
			assigneeID := req.AssigneeID.Value
			if _, err := s.userRepo.GetByID(ctx, assigneeID); err != nil {
				return nil, ErrInvalidAssigneeID
			}
			if task.TeamID != nil {
				inTeam, err := s.teamRepo.IsMember(ctx, *task.TeamID, assigneeID)
				if err != nil {
					return nil, err
				}
				if !inTeam {
					return nil, ErrAssigneeNotInTeam
				}
			}
			task.AssigneeID = &assigneeID
		}
	}

	assigneeChanged := !isSameUint64Ptr(originalAssigneeID, task.AssigneeID)

	if err := s.taskRepo.UpdateByID(ctx, task); err != nil {
		return nil, err
	}
	if assigneeChanged && task.TeamID != nil {
		if err := s.assignRepo.DeclinePendingByTaskID(ctx, task.ID); err != nil {
			return nil, err
		}
	}
	updatedTask, err := s.taskRepo.GetByID(ctx, task.ID)
	if err != nil {
		return nil, err
	}
	if err := s.createAssignmentNotification(ctx, updatedTask, operatorID, originalAssigneeID, originalTeamID); err != nil {
		return nil, err
	}
	return updatedTask, nil
}

func (s *taskService) ListPendingAssignments(ctx context.Context, userID uint64) ([]model.TaskAssignmentRequestDetail, error) {
	return s.assignRepo.ListPendingByAssignee(ctx, userID)
}

func (s *taskService) AcceptAssignment(ctx context.Context, userID, requestID uint64) error {
	req, err := s.assignRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.AssigneeID != userID {
		return repository.ErrForbidden
	}
	if req.Status != model.InvitationStatusPending {
		return ErrAssignmentNotPending
	}

	if _, err := s.teamRepo.GetByID(ctx, req.TeamID); err != nil {
		return err
	}
	if inTeam, err := s.teamRepo.IsMember(ctx, req.TeamID, userID); err != nil {
		return err
	} else if !inTeam {
		return repository.ErrForbidden
	}

	task, err := s.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		return err
	}

	task.AssigneeID = &userID
	if err := s.taskRepo.UpdateByID(ctx, task); err != nil {
		return err
	}

	return s.assignRepo.UpdateStatus(ctx, requestID, model.InvitationStatusAccepted)
}

func (s *taskService) DeclineAssignment(ctx context.Context, userID, requestID uint64) error {
	req, err := s.assignRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.AssigneeID != userID {
		return repository.ErrForbidden
	}
	if req.Status != model.InvitationStatusPending {
		return ErrAssignmentNotPending
	}

	return s.assignRepo.UpdateStatus(ctx, requestID, model.InvitationStatusDeclined)
}

func (s *taskService) AddDependency(ctx context.Context, userID, taskID, predecessorID uint64) (*model.TaskDependency, error) {
	if taskID == 0 || predecessorID == 0 || taskID == predecessorID {
		return nil, ErrInvalidTaskDependency
	}

	successor, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if successor.Status == model.TaskStatusDONE {
		return nil, ErrDependencyOnDoneTask
	}

	predecessor, err := s.taskRepo.GetByID(ctx, predecessorID)
	if err != nil {
		return nil, err
	}

	if successor.TeamID == nil && predecessor.TeamID == nil {
		if successor.UserID != userID || predecessor.UserID != userID {
			return nil, ErrCrossTaskDependency
		}
		dep, err := s.taskRepo.AddDependencyWithCycleCheck(ctx, predecessorID, taskID)
		if errors.Is(err, repository.ErrDependencyCycle) {
			return nil, ErrDependencyCycle
		}
		return dep, err
	}

	if successor.TeamID == nil || predecessor.TeamID == nil {
		return nil, ErrCrossTaskDependency
	}
	if *successor.TeamID != *predecessor.TeamID {
		return nil, ErrCrossTaskDependency
	}
	if err := s.ensureTeamActive(ctx, *successor.TeamID); err != nil {
		return nil, err
	}

	if err := s.requireTeamManager(ctx, *successor.TeamID, userID); err != nil {
		return nil, err
	}

	dep, err := s.taskRepo.AddDependencyWithCycleCheck(ctx, predecessorID, taskID)
	if errors.Is(err, repository.ErrDependencyCycle) {
		return nil, ErrDependencyCycle
	}
	return dep, err
}

func (s *taskService) RemoveDependency(ctx context.Context, userID, taskID, predecessorID uint64) error {
	if taskID == 0 || predecessorID == 0 {
		return ErrInvalidTaskDependency
	}

	successor, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	predecessor, err := s.taskRepo.GetByID(ctx, predecessorID)
	if err != nil {
		return err
	}

	if successor.TeamID == nil && predecessor.TeamID == nil {
		if successor.UserID != userID || predecessor.UserID != userID {
			return ErrCrossTaskDependency
		}
		return s.taskRepo.RemoveDependency(ctx, predecessorID, taskID)
	}

	if successor.TeamID == nil || predecessor.TeamID == nil {
		return ErrCrossTaskDependency
	}
	if *successor.TeamID != *predecessor.TeamID {
		return ErrCrossTaskDependency
	}
	if err := s.ensureTeamActive(ctx, *successor.TeamID); err != nil {
		return err
	}

	if err := s.requireTeamManager(ctx, *successor.TeamID, userID); err != nil {
		return err
	}

	return s.taskRepo.RemoveDependency(ctx, predecessorID, taskID)
}

func (s *taskService) GetDependencies(ctx context.Context, userID, taskID uint64) (*TaskDependencies, error) {
	if taskID == 0 {
		return nil, ErrInvalidTaskDependency
	}

	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if err := s.requireTaskReadAccess(ctx, task, userID); err != nil {
		return nil, err
	}

	predecessors, err := s.taskRepo.ListPredecessors(ctx, taskID)
	if err != nil {
		return nil, err
	}
	successors, err := s.taskRepo.ListSuccessors(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return &TaskDependencies{
		Predecessors: predecessors,
		Successors:   successors,
	}, nil
}

func (s *taskService) GetDependencyGraph(ctx context.Context, userID, taskID uint64) (*TaskDependencyGraph, error) {
	if taskID == 0 {
		return nil, ErrInvalidTaskDependency
	}

	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if err := s.requireTaskReadAccess(ctx, task, userID); err != nil {
		return nil, err
	}

	nodes, links, err := s.taskRepo.GetDependencyGraph(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return &TaskDependencyGraph{
		Nodes: nodes,
		Links: links,
	}, nil
}

func (s *taskService) createAssignmentNotification(ctx context.Context, task *model.Task, requesterID uint64, originalAssigneeID, originalTeamID *uint64) error {
	if task.TeamID == nil || task.AssigneeID == nil {
		return nil
	}
	if *task.AssigneeID == requesterID {
		return nil
	}

	assigneeChanged := true
	teamChanged := true
	if originalAssigneeID != nil && *originalAssigneeID == *task.AssigneeID {
		assigneeChanged = false
	}
	if originalTeamID != nil && *originalTeamID == *task.TeamID {
		teamChanged = false
	}
	if originalAssigneeID != nil || originalTeamID != nil {
		if !assigneeChanged && !teamChanged {
			return nil
		}
	}

	req := &model.TaskAssignmentRequest{
		TaskID:      task.ID,
		TeamID:      *task.TeamID,
		RequesterID: requesterID,
		AssigneeID:  *task.AssigneeID,
		Status:      model.InvitationStatusAccepted,
	}
	if err := s.assignRepo.Create(ctx, req); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil
		}
		return err
	}
	return nil
}

func (s *taskService) ensureDependenciesCompleted(ctx context.Context, taskID uint64) error {
	predecessors, err := s.taskRepo.ListIncompletePredecessors(ctx, taskID)
	if err != nil {
		return err
	}
	if len(predecessors) > 0 {
		ids := make([]uint64, 0, len(predecessors))
		for _, predecessor := range predecessors {
			ids = append(ids, predecessor.ID)
		}
		return &UnresolvedDependenciesError{IDs: ids, Predecessors: predecessors}
	}
	return nil
}

func (s *taskService) ensureTeamActive(ctx context.Context, teamID uint64) error {
	if _, err := s.teamRepo.GetByID(ctx, teamID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			deleted, derr := s.teamRepo.IsDeleted(ctx, teamID)
			if derr != nil {
				return derr
			}
			if deleted {
				return ErrTeamGone
			}
		}
		return err
	}
	return nil
}

func hasMemberRestrictedFields(req UpdateTaskRequest) bool {
	return req.Title != nil || req.Description != nil || req.Priority != nil || req.Deadline != nil || req.TeamID != nil || req.AssigneeID.Set
}

func isSameUint64Ptr(left, right *uint64) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
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
