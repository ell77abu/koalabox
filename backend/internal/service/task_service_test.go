package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-backend/internal/model"
	"task-backend/internal/repository"
)

type taskRepoMock struct {
	listByTeamIDFn                 func(ctx context.Context, teamID uint64, page, size int) ([]model.Task, int64, error)
	listPersonalAndAssignedTeamFn  func(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error)
	getByIDFn                      func(ctx context.Context, id uint64) (*model.Task, error)
	updateByIDFn                   func(ctx context.Context, task *model.Task) error
	listIncompletePredecessorIDsFn func(ctx context.Context, taskID uint64) ([]uint64, error)
}

func (m *taskRepoMock) Create(ctx context.Context, task *model.Task) error {
	panic("unexpected call to Create")
}
func (m *taskRepoMock) ListPersonalAndAssignedTeam(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error) {
	if m.listPersonalAndAssignedTeamFn == nil {
		panic("unexpected call to ListPersonalAndAssignedTeam")
	}
	return m.listPersonalAndAssignedTeamFn(ctx, userID, page, size)
}
func (m *taskRepoMock) ListByTeamID(ctx context.Context, teamID uint64, page, size int) ([]model.Task, int64, error) {
	if m.listByTeamIDFn == nil {
		panic("unexpected call to ListByTeamID")
	}
	return m.listByTeamIDFn(ctx, teamID, page, size)
}
func (m *taskRepoMock) ListByUserID(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error) {
	panic("unexpected call to ListByUserID")
}
func (m *taskRepoMock) ListAll(ctx context.Context, page, size int) ([]model.Task, int64, error) {
	panic("unexpected call to ListAll")
}
func (m *taskRepoMock) GetByIDAndUserID(ctx context.Context, id, userID uint64) (*model.Task, error) {
	panic("unexpected call to GetByIDAndUserID")
}
func (m *taskRepoMock) GetByID(ctx context.Context, id uint64) (*model.Task, error) {
	if m.getByIDFn == nil {
		panic("unexpected call to GetByID")
	}
	return m.getByIDFn(ctx, id)
}
func (m *taskRepoMock) UpdateByIDAndUserID(ctx context.Context, task *model.Task) error {
	panic("unexpected call to UpdateByIDAndUserID")
}
func (m *taskRepoMock) UpdateByID(ctx context.Context, task *model.Task) error {
	if m.updateByIDFn == nil {
		panic("unexpected call to UpdateByID")
	}
	return m.updateByIDFn(ctx, task)
}
func (m *taskRepoMock) DeleteByIDAndUserID(ctx context.Context, id, userID uint64) error {
	panic("unexpected call to DeleteByIDAndUserID")
}
func (m *taskRepoMock) DeleteByID(ctx context.Context, id uint64) error {
	panic("unexpected call to DeleteByID")
}
func (m *taskRepoMock) ListByUserIDWithOrder(ctx context.Context, userID uint64, page, size int, orderBy string) ([]model.Task, int64, error) {
	panic("unexpected call to ListByUserIDWithOrder")
}
func (m *taskRepoMock) ListAllWithOrder(ctx context.Context, page, size int, orderBy string) ([]model.Task, int64, error) {
	panic("unexpected call to ListAllWithOrder")
}
func (m *taskRepoMock) ListByUserIDWithSearch(ctx context.Context, userID uint64, page, size int, search, orderBy string) ([]model.Task, int64, error) {
	panic("unexpected call to ListByUserIDWithSearch")
}
func (m *taskRepoMock) ListAllWithSearch(ctx context.Context, page, size int, search, orderBy string) ([]model.Task, int64, error) {
	panic("unexpected call to ListAllWithSearch")
}
func (m *taskRepoMock) ClearAssigneeForTeamTasks(ctx context.Context, teamID, assigneeID uint64) error {
	panic("unexpected call to ClearAssigneeForTeamTasks")
}
func (m *taskRepoMock) AddDependency(ctx context.Context, predecessorID, successorID uint64) (*model.TaskDependency, error) {
	panic("unexpected call to AddDependency")
}
func (m *taskRepoMock) RemoveDependency(ctx context.Context, predecessorID, successorID uint64) error {
	panic("unexpected call to RemoveDependency")
}
func (m *taskRepoMock) ListPredecessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error) {
	panic("unexpected call to ListPredecessors")
}
func (m *taskRepoMock) ListSuccessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error) {
	panic("unexpected call to ListSuccessors")
}
func (m *taskRepoMock) CountIncompletePredecessors(ctx context.Context, taskID uint64) (int64, error) {
	panic("unexpected call to CountIncompletePredecessors")
}
func (m *taskRepoMock) ListIncompletePredecessorIDs(ctx context.Context, taskID uint64) ([]uint64, error) {
	if m.listIncompletePredecessorIDsFn == nil {
		panic("unexpected call to ListIncompletePredecessorIDs")
	}
	return m.listIncompletePredecessorIDsFn(ctx, taskID)
}
func (m *taskRepoMock) HasDependencyPath(ctx context.Context, fromID, toID uint64) (bool, error) {
	panic("unexpected call to HasDependencyPath")
}
func (m *taskRepoMock) GetDependencyGraph(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, []model.TaskDependencyLink, error) {
	panic("unexpected call to GetDependencyGraph")
}

type teamRepoForTaskMock struct {
	getByIDFn       func(ctx context.Context, teamID uint64) (*model.Team, error)
	isMemberFn      func(ctx context.Context, teamID, userID uint64) (bool, error)
	getMemberRoleFn func(ctx context.Context, teamID, userID uint64) (model.TeamRole, error)
}

func (m *teamRepoForTaskMock) Create(ctx context.Context, team *model.Team) error {
	panic("unexpected call to Create")
}
func (m *teamRepoForTaskMock) ListByUserID(ctx context.Context, userID uint64, search string) ([]model.TeamWithRole, error) {
	panic("unexpected call to ListByUserID")
}
func (m *teamRepoForTaskMock) GetByID(ctx context.Context, teamID uint64) (*model.Team, error) {
	if m.getByIDFn == nil {
		panic("unexpected call to GetByID")
	}
	return m.getByIDFn(ctx, teamID)
}
func (m *teamRepoForTaskMock) UpdateName(ctx context.Context, teamID uint64, name string) error {
	panic("unexpected call to UpdateName")
}
func (m *teamRepoForTaskMock) AddMember(ctx context.Context, member *model.TeamMember) error {
	panic("unexpected call to AddMember")
}
func (m *teamRepoForTaskMock) RemoveMember(ctx context.Context, teamID, userID uint64) error {
	panic("unexpected call to RemoveMember")
}
func (m *teamRepoForTaskMock) UpdateMemberRole(ctx context.Context, teamID, userID uint64, role model.TeamRole) error {
	panic("unexpected call to UpdateMemberRole")
}
func (m *teamRepoForTaskMock) IsMember(ctx context.Context, teamID, userID uint64) (bool, error) {
	if m.isMemberFn == nil {
		panic("unexpected call to IsMember")
	}
	return m.isMemberFn(ctx, teamID, userID)
}
func (m *teamRepoForTaskMock) GetMemberRole(ctx context.Context, teamID, userID uint64) (model.TeamRole, error) {
	if m.getMemberRoleFn == nil {
		panic("unexpected call to GetMemberRole")
	}
	return m.getMemberRoleFn(ctx, teamID, userID)
}
func (m *teamRepoForTaskMock) ListMembers(ctx context.Context, teamID uint64) ([]model.TeamMemberDetail, error) {
	panic("unexpected call to ListMembers")
}
func (m *teamRepoForTaskMock) TransferOwnership(ctx context.Context, teamID, oldOwnerID, newOwnerID uint64) error {
	panic("unexpected call to TransferOwnership")
}

type taskUserRepoStub struct{}

func (s *taskUserRepoStub) Create(ctx context.Context, user *model.User) error {
	panic("unexpected call")
}
func (s *taskUserRepoStub) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	panic("unexpected call")
}
func (s *taskUserRepoStub) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	panic("unexpected call")
}
func (s *taskUserRepoStub) UpdatePasswordByID(ctx context.Context, id uint64, passwordHash string) error {
	panic("unexpected call")
}

type taskAssignRepoStub struct{}

func (s *taskAssignRepoStub) Create(ctx context.Context, req *model.TaskAssignmentRequest) error {
	panic("unexpected call")
}
func (s *taskAssignRepoStub) GetByID(ctx context.Context, id uint64) (*model.TaskAssignmentRequest, error) {
	panic("unexpected call")
}
func (s *taskAssignRepoStub) GetPendingByTaskAndAssignee(ctx context.Context, taskID, assigneeID uint64) (*model.TaskAssignmentRequest, error) {
	panic("unexpected call")
}
func (s *taskAssignRepoStub) ListPendingByAssignee(ctx context.Context, assigneeID uint64) ([]model.TaskAssignmentRequestDetail, error) {
	panic("unexpected call")
}
func (s *taskAssignRepoStub) UpdateStatus(ctx context.Context, id uint64, status model.InvitationStatus) error {
	panic("unexpected call")
}
func (s *taskAssignRepoStub) DeclinePendingByTaskID(ctx context.Context, taskID uint64) error {
	panic("unexpected call")
}

func TestTaskServiceListRejectsNonMemberTeamView(t *testing.T) {
	svc := NewTaskService(
		&taskRepoMock{},
		&teamRepoForTaskMock{
			getByIDFn: func(ctx context.Context, teamID uint64) (*model.Team, error) {
				return &model.Team{ID: teamID}, nil
			},
			isMemberFn: func(ctx context.Context, teamID, userID uint64) (bool, error) {
				return false, nil
			},
		},
		&taskUserRepoStub{},
		&taskAssignRepoStub{},
	)

	teamID := uint64(7)
	_, _, _, _, err := svc.List(context.Background(), 3, 0, 1, 10, &teamID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrForbidden)
}

func TestTaskServiceListUsesTeamIDFilterWhenProvided(t *testing.T) {
	const teamID = uint64(8)
	called := false

	svc := NewTaskService(
		&taskRepoMock{
			listByTeamIDFn: func(ctx context.Context, gotTeamID uint64, page, size int) ([]model.Task, int64, error) {
				called = true
				assert.Equal(t, teamID, gotTeamID)
				return []model.Task{{ID: 101, TeamID: uint64Ptr(teamID)}}, 1, nil
			},
		},
		&teamRepoForTaskMock{
			getByIDFn: func(ctx context.Context, gotTeamID uint64) (*model.Team, error) {
				assert.Equal(t, teamID, gotTeamID)
				return &model.Team{ID: gotTeamID}, nil
			},
			isMemberFn: func(ctx context.Context, gotTeamID, gotUserID uint64) (bool, error) {
				assert.Equal(t, teamID, gotTeamID)
				assert.Equal(t, uint64(2), gotUserID)
				return true, nil
			},
		},
		&taskUserRepoStub{},
		&taskAssignRepoStub{},
	)

	list, total, _, _, err := svc.List(context.Background(), 2, 0, 1, 10, uint64Ptr(teamID))
	require.NoError(t, err)
	require.True(t, called)
	require.Len(t, list, 1)
	assert.EqualValues(t, 1, total)
}

func TestTaskServiceUpdateRejectsMemberEditingNonStatusFields(t *testing.T) {
	taskTeamID := uint64(5)
	task := &model.Task{
		ID:         99,
		UserID:     1,
		TeamID:     &taskTeamID,
		AssigneeID: uint64Ptr(7),
		Status:     model.TaskStatusTODO,
	}

	svc := NewTaskService(
		&taskRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.Task, error) {
				assert.Equal(t, uint64(99), id)
				return task, nil
			},
		},
		&teamRepoForTaskMock{
			getMemberRoleFn: func(ctx context.Context, teamID, userID uint64) (model.TeamRole, error) {
				assert.Equal(t, taskTeamID, teamID)
				assert.Equal(t, uint64(7), userID)
				return model.TeamRoleMember, nil
			},
		},
		&taskUserRepoStub{},
		&taskAssignRepoStub{},
	)

	newTitle := "new title"
	_, err := svc.Update(context.Background(), 7, 99, UpdateTaskRequest{
		Status: strPtr("DOING"),
		Title:  &newTitle,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrMemberStatusOnly)
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}

func strPtr(v string) *string {
	return &v
}
