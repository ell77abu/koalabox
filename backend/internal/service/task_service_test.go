package service

import (
	"context"
	"errors"
	"testing"

	"task-backend/internal/model"
	"task-backend/internal/repository"
)

type taskRepoMock struct {
	listPersonalOnlyCalled     bool
	listAssignedTeamOnlyCalled bool
	listPersonalAndTeamCalled  bool
	listAllPersonalCalled      bool
	listAllTeamCalled          bool
	listAllCalled              bool
}

func (m *taskRepoMock) Create(ctx context.Context, task *model.Task) error { panic("unexpected call") }
func (m *taskRepoMock) ListByTeamID(ctx context.Context, teamID uint64, page, size int) ([]model.Task, int64, error) {
	panic("unexpected call")
}
func (m *taskRepoMock) ListByUserID(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error) {
	panic("unexpected call")
}
func (m *taskRepoMock) ListByUserIDWithOrder(ctx context.Context, userID uint64, page, size int, orderBy string) ([]model.Task, int64, error) {
	panic("unexpected call")
}
func (m *taskRepoMock) ListAllWithOrder(ctx context.Context, page, size int, orderBy string) ([]model.Task, int64, error) {
	panic("unexpected call")
}
func (m *taskRepoMock) ListByUserIDWithSearch(ctx context.Context, userID uint64, page, size int, search, orderBy string) ([]model.Task, int64, error) {
	panic("unexpected call")
}
func (m *taskRepoMock) ListAllWithSearch(ctx context.Context, page, size int, search, orderBy string) ([]model.Task, int64, error) {
	panic("unexpected call")
}
func (m *taskRepoMock) GetByIDAndUserID(ctx context.Context, id, userID uint64) (*model.Task, error) {
	panic("unexpected call")
}
func (m *taskRepoMock) GetByID(ctx context.Context, id uint64) (*model.Task, error) {
	panic("unexpected call")
}
func (m *taskRepoMock) UpdateByIDAndUserID(ctx context.Context, task *model.Task) error {
	panic("unexpected call")
}
func (m *taskRepoMock) UpdateByID(ctx context.Context, task *model.Task) error {
	panic("unexpected call")
}
func (m *taskRepoMock) DeleteByIDAndUserID(ctx context.Context, id, userID uint64) error {
	panic("unexpected call")
}
func (m *taskRepoMock) DeleteByID(ctx context.Context, id uint64) error { panic("unexpected call") }

func (m *taskRepoMock) ListPersonalAndAssignedTeam(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error) {
	m.listPersonalAndTeamCalled = true
	return []model.Task{}, 0, nil
}

func (m *taskRepoMock) ListPersonalOnly(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error) {
	m.listPersonalOnlyCalled = true
	return []model.Task{}, 0, nil
}

func (m *taskRepoMock) ListAssignedTeamOnly(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error) {
	m.listAssignedTeamOnlyCalled = true
	return []model.Task{}, 0, nil
}

func (m *taskRepoMock) ListAll(ctx context.Context, page, size int) ([]model.Task, int64, error) {
	m.listAllCalled = true
	return []model.Task{}, 0, nil
}

func (m *taskRepoMock) ListAllPersonal(ctx context.Context, page, size int) ([]model.Task, int64, error) {
	m.listAllPersonalCalled = true
	return []model.Task{}, 0, nil
}

func (m *taskRepoMock) ListAllTeam(ctx context.Context, page, size int) ([]model.Task, int64, error) {
	m.listAllTeamCalled = true
	return []model.Task{}, 0, nil
}

type teamRepoStub struct{}

func (s *teamRepoStub) Create(ctx context.Context, team *model.Team) error { panic("unexpected call") }
func (s *teamRepoStub) ListByUserID(ctx context.Context, userID uint64, search string) ([]model.TeamWithRole, error) {
	panic("unexpected call")
}
func (s *teamRepoStub) GetByID(ctx context.Context, teamID uint64) (*model.Team, error) {
	return nil, repository.ErrNotFound
}
func (s *teamRepoStub) UpdateName(ctx context.Context, teamID uint64, name string) error {
	panic("unexpected call")
}
func (s *teamRepoStub) AddMember(ctx context.Context, member *model.TeamMember) error {
	panic("unexpected call")
}
func (s *teamRepoStub) RemoveMember(ctx context.Context, teamID, userID uint64) error {
	panic("unexpected call")
}
func (s *teamRepoStub) UpdateMemberRole(ctx context.Context, teamID, userID uint64, role model.TeamRole) error {
	panic("unexpected call")
}
func (s *teamRepoStub) IsMember(ctx context.Context, teamID, userID uint64) (bool, error) {
	panic("unexpected call")
}
func (s *teamRepoStub) GetMemberRole(ctx context.Context, teamID, userID uint64) (model.TeamRole, error) {
	panic("unexpected call")
}
func (s *teamRepoStub) ListMembers(ctx context.Context, teamID uint64) ([]model.TeamMemberDetail, error) {
	panic("unexpected call")
}

type userRepoStub struct{}

func (s *userRepoStub) Create(ctx context.Context, user *model.User) error { panic("unexpected call") }
func (s *userRepoStub) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	panic("unexpected call")
}
func (s *userRepoStub) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	panic("unexpected call")
}
func (s *userRepoStub) UpdatePasswordByID(ctx context.Context, id uint64, passwordHash string) error {
	panic("unexpected call")
}

type assignmentRepoStub struct{}

func (s *assignmentRepoStub) Create(ctx context.Context, req *model.TaskAssignmentRequest) error {
	panic("unexpected call")
}
func (s *assignmentRepoStub) GetByID(ctx context.Context, id uint64) (*model.TaskAssignmentRequest, error) {
	panic("unexpected call")
}
func (s *assignmentRepoStub) GetPendingByTaskAndAssignee(ctx context.Context, taskID, assigneeID uint64) (*model.TaskAssignmentRequest, error) {
	panic("unexpected call")
}
func (s *assignmentRepoStub) ListPendingByAssignee(ctx context.Context, assigneeID uint64) ([]model.TaskAssignmentRequestDetail, error) {
	panic("unexpected call")
}
func (s *assignmentRepoStub) UpdateStatus(ctx context.Context, id uint64, status model.InvitationStatus) error {
	panic("unexpected call")
}
func (s *assignmentRepoStub) DeclinePendingByTaskID(ctx context.Context, taskID uint64) error {
	panic("unexpected call")
}

func TestTaskServiceListUsesUserScopeFilter(t *testing.T) {
	taskRepo := &taskRepoMock{}
	svc := NewTaskService(taskRepo, &teamRepoStub{}, &userRepoStub{}, &assignmentRepoStub{})

	_, _, _, _, err := svc.List(context.Background(), 7, 0, 1, 10, nil, TaskScopeTeam)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !taskRepo.listAssignedTeamOnlyCalled {
		t.Fatal("expected team-scope query for normal user")
	}
	if taskRepo.listPersonalAndTeamCalled || taskRepo.listPersonalOnlyCalled {
		t.Fatal("unexpected repository branch called")
	}
}

func TestTaskServiceListUsesAdminScopeFilter(t *testing.T) {
	taskRepo := &taskRepoMock{}
	svc := NewTaskService(taskRepo, &teamRepoStub{}, &userRepoStub{}, &assignmentRepoStub{})

	_, _, _, _, err := svc.List(context.Background(), 1, 1, 1, 10, nil, TaskScopePersonal)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !taskRepo.listAllPersonalCalled {
		t.Fatal("expected personal-scope query for admin")
	}
	if taskRepo.listAllCalled || taskRepo.listAllTeamCalled {
		t.Fatal("unexpected repository branch called")
	}
}

func TestTaskServiceListRejectsInvalidScope(t *testing.T) {
	taskRepo := &taskRepoMock{}
	svc := NewTaskService(taskRepo, &teamRepoStub{}, &userRepoStub{}, &assignmentRepoStub{})

	_, _, _, _, err := svc.List(context.Background(), 7, 0, 1, 10, nil, TaskScope("weird"))
	if !errors.Is(err, ErrInvalidTaskScope) {
		t.Fatalf("expected ErrInvalidTaskScope, got %v", err)
	}
}
