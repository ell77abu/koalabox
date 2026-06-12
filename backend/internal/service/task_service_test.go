package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-backend/internal/model"
	"task-backend/internal/repository"
)

type taskRepoMock struct {
	createFn                       func(ctx context.Context, task *model.Task) error
	listByTeamIDFn                 func(ctx context.Context, teamID uint64, page, size int) ([]model.Task, int64, error)
	listPersonalAndAssignedTeamFn  func(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error)
	getByIDFn                      func(ctx context.Context, id uint64) (*model.Task, error)
	getByIDForUserScopeFn          func(ctx context.Context, id, userID uint64) (*model.Task, error)
	updateByIDFn                   func(ctx context.Context, task *model.Task) error
	updateByIDForScopeFn           func(ctx context.Context, task *model.Task, userID uint64) error
	updateStatusByScopeFn          func(ctx context.Context, taskID, userID uint64, status model.TaskStatus) error
	deleteByIDForScopeFn           func(ctx context.Context, id, userID uint64) error
	deleteByIDFn                   func(ctx context.Context, id uint64) error
	listIncompletePredecessorIDsFn func(ctx context.Context, taskID uint64) ([]uint64, error)
	listIncompletePredecessorsFn   func(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error)
	listDoneSuccessorsFn           func(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error)
	addDependencyWithCycleCheckFn  func(ctx context.Context, predecessorID, successorID uint64) (*model.TaskDependency, error)
}

func (m *taskRepoMock) Create(ctx context.Context, task *model.Task) error {
	if m.createFn == nil {
		panic("unexpected call to Create")
	}
	return m.createFn(ctx, task)
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
func (m *taskRepoMock) GetByIDForUserScope(ctx context.Context, id, userID uint64) (*model.Task, error) {
	if m.getByIDForUserScopeFn != nil {
		return m.getByIDForUserScopeFn(ctx, id, userID)
	}
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	panic("unexpected call to GetByIDForUserScope")
}
func (m *taskRepoMock) GetByID(ctx context.Context, id uint64) (*model.Task, error) {
	if m.getByIDFn == nil {
		panic("unexpected call to GetByID")
	}
	return m.getByIDFn(ctx, id)
}
func (m *taskRepoMock) GetByIDIncludingDeleted(ctx context.Context, id uint64) (*model.Task, error) {
	if m.getByIDFn == nil {
		panic("unexpected call to GetByIDIncludingDeleted")
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
func (m *taskRepoMock) UpdateByIDForOwnerOrTeamManagerScope(ctx context.Context, task *model.Task, userID uint64) error {
	if m.updateByIDForScopeFn != nil {
		return m.updateByIDForScopeFn(ctx, task, userID)
	}
	if m.updateByIDFn != nil {
		return m.updateByIDFn(ctx, task)
	}
	panic("unexpected call to UpdateByIDForOwnerOrTeamManagerScope")
}
func (m *taskRepoMock) UpdateStatusByIDForAssignedMemberScope(ctx context.Context, taskID, userID uint64, status model.TaskStatus) error {
	if m.updateStatusByScopeFn != nil {
		return m.updateStatusByScopeFn(ctx, taskID, userID, status)
	}
	panic("unexpected call to UpdateStatusByIDForAssignedMemberScope")
}
func (m *taskRepoMock) DeleteByIDAndUserID(ctx context.Context, id, userID uint64) error {
	panic("unexpected call to DeleteByIDAndUserID")
}
func (m *taskRepoMock) DeleteByIDForOwnerOrTeamManagerScope(ctx context.Context, id, userID uint64) error {
	if m.deleteByIDForScopeFn != nil {
		return m.deleteByIDForScopeFn(ctx, id, userID)
	}
	panic("unexpected call to DeleteByIDForOwnerOrTeamManagerScope")
}
func (m *taskRepoMock) DeleteByID(ctx context.Context, id uint64) error {
	if m.deleteByIDFn == nil {
		panic("unexpected call to DeleteByID")
	}
	return m.deleteByIDFn(ctx, id)
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
func (m *taskRepoMock) AddDependencyWithCycleCheck(ctx context.Context, predecessorID, successorID uint64) (*model.TaskDependency, error) {
	if m.addDependencyWithCycleCheckFn != nil {
		return m.addDependencyWithCycleCheckFn(ctx, predecessorID, successorID)
	}
	panic("unexpected call to AddDependencyWithCycleCheck")
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
func (m *taskRepoMock) ListIncompletePredecessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error) {
	if m.listIncompletePredecessorsFn == nil {
		panic("unexpected call to ListIncompletePredecessors")
	}
	return m.listIncompletePredecessorsFn(ctx, taskID)
}
func (m *taskRepoMock) ListDoneSuccessors(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, error) {
	if m.listDoneSuccessorsFn == nil {
		panic("unexpected call to ListDoneSuccessors")
	}
	return m.listDoneSuccessorsFn(ctx, taskID)
}
func (m *taskRepoMock) HasDependencyPath(ctx context.Context, fromID, toID uint64) (bool, error) {
	panic("unexpected call to HasDependencyPath")
}
func (m *taskRepoMock) GetDependencyGraph(ctx context.Context, taskID uint64) ([]model.TaskDependencyNode, []model.TaskDependencyLink, error) {
	panic("unexpected call to GetDependencyGraph")
}

type teamRepoForTaskMock struct {
	getByIDFn               func(ctx context.Context, teamID uint64) (*model.Team, error)
	getByIDForMemberScopeFn func(ctx context.Context, teamID, userID uint64) (*model.Team, error)
	isMemberFn              func(ctx context.Context, teamID, userID uint64) (bool, error)
	getMemberRoleFn         func(ctx context.Context, teamID, userID uint64) (model.TeamRole, error)
	isDeletedFn             func(ctx context.Context, teamID uint64) (bool, error)
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
func (m *teamRepoForTaskMock) GetByIDForMemberScope(ctx context.Context, teamID, userID uint64) (*model.Team, error) {
	if m.getByIDForMemberScopeFn != nil {
		return m.getByIDForMemberScopeFn(ctx, teamID, userID)
	}
	if m.isMemberFn != nil {
		isMember, err := m.isMemberFn(ctx, teamID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, repository.ErrNotFound
		}
		return &model.Team{ID: teamID}, nil
	}
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, teamID)
	}
	panic("unexpected call to GetByIDForMemberScope")
}
func (m *teamRepoForTaskMock) GetByName(ctx context.Context, name string) (*model.Team, error) {
	panic("unexpected call to GetByName")
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
func (m *teamRepoForTaskMock) IsDeleted(ctx context.Context, teamID uint64) (bool, error) {
	if m.isDeletedFn != nil {
		return m.isDeletedFn(ctx, teamID)
	}
	return false, nil
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

type taskAssignRepoStub struct {
	declinePendingByTaskIDFn func(ctx context.Context, taskID uint64) error
}

func (s *taskAssignRepoStub) Create(ctx context.Context, req *model.TaskAssignmentRequest) error {
	panic("unexpected call")
}
func (s *taskAssignRepoStub) GetByID(ctx context.Context, id uint64) (*model.TaskAssignmentRequest, error) {
	panic("unexpected call")
}
func (s *taskAssignRepoStub) GetByIDForAssigneeScope(ctx context.Context, id, assigneeID uint64) (*model.TaskAssignmentRequest, error) {
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
func (s *taskAssignRepoStub) UpdatePendingStatusForAssigneeScope(ctx context.Context, id, assigneeID uint64, status model.InvitationStatus) error {
	panic("unexpected call")
}
func (s *taskAssignRepoStub) AcceptPendingForAssigneeScope(ctx context.Context, id, assigneeID uint64) error {
	panic("unexpected call")
}
func (s *taskAssignRepoStub) DeclinePendingByTaskID(ctx context.Context, taskID uint64) error {
	if s.declinePendingByTaskIDFn != nil {
		return s.declinePendingByTaskIDFn(ctx, taskID)
	}
	panic("unexpected call")
}

type activityLogRepoMock struct {
	createFn       func(ctx context.Context, log *model.ActivityLog) error
	listByTaskIDFn func(ctx context.Context, taskID uint64, page, size int) ([]model.ActivityLogListItem, int64, error)
}

func (m *activityLogRepoMock) Create(ctx context.Context, log *model.ActivityLog) error {
	if m.createFn == nil {
		panic("unexpected call to ActivityLog.Create")
	}
	return m.createFn(ctx, log)
}

func (m *activityLogRepoMock) ListByTaskID(ctx context.Context, taskID uint64, page, size int) ([]model.ActivityLogListItem, int64, error) {
	if m.listByTaskIDFn == nil {
		panic("unexpected call to ActivityLog.ListByTaskID")
	}
	return m.listByTaskIDFn(ctx, taskID, page, size)
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

func TestTaskServiceUpdateRejectsTeamIDChangeForPersonalTask(t *testing.T) {
	newTeamID := uint64(77)
	svc := NewTaskService(
		&taskRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.Task, error) {
				return &model.Task{
					ID:     id,
					UserID: 9,
					Status: model.TaskStatusTODO,
				}, nil
			},
		},
		&teamRepoForTaskMock{},
		&taskUserRepoStub{},
		&taskAssignRepoStub{},
	)

	_, err := svc.Update(context.Background(), 9, 101, UpdateTaskRequest{
		TeamID: &newTeamID,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTeamIDImmutable)
}

func TestTaskServiceUpdateRejectsTeamIDChangeForTeamTask(t *testing.T) {
	oldTeamID := uint64(5)
	newTeamID := uint64(8)
	svc := NewTaskService(
		&taskRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.Task, error) {
				return &model.Task{
					ID:     id,
					UserID: 2,
					TeamID: &oldTeamID,
					Status: model.TaskStatusTODO,
				}, nil
			},
		},
		&teamRepoForTaskMock{
			getMemberRoleFn: func(ctx context.Context, teamID, userID uint64) (model.TeamRole, error) {
				assert.Equal(t, oldTeamID, teamID)
				assert.Equal(t, uint64(10), userID)
				return model.TeamRoleOwner, nil
			},
		},
		&taskUserRepoStub{},
		&taskAssignRepoStub{},
	)

	_, err := svc.Update(context.Background(), 10, 102, UpdateTaskRequest{
		TeamID: &newTeamID,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTeamIDImmutable)
}

func TestTaskServiceUpdateRejectsRollbackWhenDoneSuccessorsExist(t *testing.T) {
	taskID := uint64(55)
	teamID := uint64(5)
	svc := NewTaskService(
		&taskRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.Task, error) {
				assert.Equal(t, taskID, id)
				return &model.Task{
					ID:     taskID,
					UserID: 1,
					TeamID: &teamID,
					Status: model.TaskStatusDONE,
				}, nil
			},
			listDoneSuccessorsFn: func(ctx context.Context, id uint64) ([]model.TaskDependencyNode, error) {
				assert.Equal(t, taskID, id)
				return []model.TaskDependencyNode{
					{ID: 101, Title: "done child", Status: model.TaskStatusDONE},
				}, nil
			},
		},
		&teamRepoForTaskMock{
			getMemberRoleFn: func(ctx context.Context, gotTeamID, gotUserID uint64) (model.TeamRole, error) {
				assert.Equal(t, teamID, gotTeamID)
				assert.Equal(t, uint64(7), gotUserID)
				return model.TeamRoleOwner, nil
			},
		},
		&taskUserRepoStub{},
		&taskAssignRepoStub{},
	)

	doing := "DOING"
	_, err := svc.Update(context.Background(), 7, taskID, UpdateTaskRequest{Status: &doing})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDoneSuccessorsExist)
	var rollbackErr *DoneSuccessorsError
	assert.ErrorAs(t, err, &rollbackErr)
	assert.Equal(t, []uint64{101}, rollbackErr.IDs)
}

func TestTaskServiceCreateWritesActivityLog(t *testing.T) {
	var createdLog *model.ActivityLog
	svc := NewTaskService(
		&taskRepoMock{
			createFn: func(ctx context.Context, task *model.Task) error {
				task.ID = 101
				return nil
			},
		},
		&teamRepoForTaskMock{},
		&userRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.User, error) {
				assert.Equal(t, uint64(9), id)
				return &model.User{ID: id, Username: "alice"}, nil
			},
		},
		&taskAssignRepoStub{},
		&activityLogRepoMock{
			createFn: func(ctx context.Context, log *model.ActivityLog) error {
				createdLog = log
				return nil
			},
		},
	)

	task, err := svc.Create(context.Background(), 9, CreateTaskRequest{Title: " 实现登录 "})
	require.NoError(t, err)
	require.NotNil(t, task)
	require.NotNil(t, createdLog)
	assert.Equal(t, model.ActivityActionCreateTask, createdLog.ActionType)
	assert.Equal(t, model.ActivityTargetTask, createdLog.TargetType)
	assert.Equal(t, model.ActivityScopePersonal, createdLog.Scope)
	assert.Equal(t, uint64(101), createdLog.TargetID)
	assert.Equal(t, uint64(9), createdLog.ActorID)
	assert.Equal(t, "alice", createdLog.ActorName)
	assert.Equal(t, "创建了任务「实现登录」", createdLog.Content)
}

func TestTaskServiceUpdateWritesActivityLogForActualChanges(t *testing.T) {
	stored := model.Task{
		ID:       77,
		UserID:   9,
		Title:    "旧标题",
		Status:   model.TaskStatusTODO,
		Priority: model.PriorityLow,
	}
	var createdLog *model.ActivityLog
	svc := NewTaskService(
		&taskRepoMock{
			getByIDForUserScopeFn: func(ctx context.Context, id, userID uint64) (*model.Task, error) {
				assert.Equal(t, uint64(77), id)
				assert.Equal(t, uint64(9), userID)
				cloned := cloneTask(&stored)
				return &cloned, nil
			},
			updateByIDForScopeFn: func(ctx context.Context, task *model.Task, userID uint64) error {
				assert.Equal(t, uint64(9), userID)
				stored = cloneTask(task)
				return nil
			},
		},
		&teamRepoForTaskMock{},
		&userRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.User, error) {
				return &model.User{ID: id, Username: "alice"}, nil
			},
		},
		&taskAssignRepoStub{},
		&activityLogRepoMock{
			createFn: func(ctx context.Context, log *model.ActivityLog) error {
				createdLog = log
				return nil
			},
		},
	)

	newTitle := "新标题"
	newPriority := 3
	updated, err := svc.Update(context.Background(), 9, 77, UpdateTaskRequest{
		Title:    &newTitle,
		Priority: &newPriority,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.NotNil(t, createdLog)
	assert.Equal(t, model.ActivityActionUpdateTaskInfo, createdLog.ActionType)
	assert.Equal(t, uint64(77), createdLog.TargetID)
	assert.Contains(t, createdLog.Content, "将标题从「旧标题」修改为「新标题」")
	assert.Contains(t, createdLog.Content, "将优先级从 1 修改为 3")
}

func TestTaskServiceUpdateStatusByAssignedMemberWritesActivityLog(t *testing.T) {
	teamID := uint64(5)
	stored := model.Task{
		ID:         79,
		UserID:     1,
		TeamID:     &teamID,
		AssigneeID: uint64Ptr(7),
		Title:      "团队任务",
		Status:     model.TaskStatusTODO,
		Priority:   model.PriorityMedium,
	}
	var createdLog *model.ActivityLog
	svc := NewTaskService(
		&taskRepoMock{
			getByIDForUserScopeFn: func(ctx context.Context, id, userID uint64) (*model.Task, error) {
				assert.Equal(t, uint64(79), id)
				assert.Equal(t, uint64(7), userID)
				cloned := cloneTask(&stored)
				return &cloned, nil
			},
			updateStatusByScopeFn: func(ctx context.Context, taskID, userID uint64, status model.TaskStatus) error {
				assert.Equal(t, uint64(79), taskID)
				assert.Equal(t, uint64(7), userID)
				stored.Status = status
				return nil
			},
		},
		&teamRepoForTaskMock{
			getMemberRoleFn: func(ctx context.Context, gotTeamID, gotUserID uint64) (model.TeamRole, error) {
				assert.Equal(t, teamID, gotTeamID)
				assert.Equal(t, uint64(7), gotUserID)
				return model.TeamRoleMember, nil
			},
		},
		&userRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.User, error) {
				assert.Equal(t, uint64(7), id)
				return &model.User{ID: id, Username: "member"}, nil
			},
		},
		&taskAssignRepoStub{},
		&activityLogRepoMock{
			createFn: func(ctx context.Context, log *model.ActivityLog) error {
				createdLog = log
				return nil
			},
		},
	)

	doing := "DOING"
	updated, err := svc.Update(context.Background(), 7, 79, UpdateTaskRequest{Status: &doing})
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.NotNil(t, createdLog)
	assert.Equal(t, model.ActivityActionUpdateTaskInfo, createdLog.ActionType)
	assert.Equal(t, model.ActivityScopeTeam, createdLog.Scope)
	assert.Equal(t, teamID, *createdLog.TeamID)
	assert.Contains(t, createdLog.Content, "将状态从 TODO 修改为 DOING")
}

func TestTaskServiceUpdateAssigneeWritesActivityLog(t *testing.T) {
	teamID := uint64(6)
	stored := model.Task{
		ID:       80,
		UserID:   9,
		TeamID:   &teamID,
		Title:    "分配任务",
		Status:   model.TaskStatusTODO,
		Priority: model.PriorityMedium,
	}
	var createdLog *model.ActivityLog
	declineCalled := false
	svc := NewTaskService(
		&taskRepoMock{
			getByIDForUserScopeFn: func(ctx context.Context, id, userID uint64) (*model.Task, error) {
				assert.Equal(t, uint64(80), id)
				assert.Equal(t, uint64(9), userID)
				cloned := cloneTask(&stored)
				return &cloned, nil
			},
			updateByIDForScopeFn: func(ctx context.Context, task *model.Task, userID uint64) error {
				assert.Equal(t, uint64(9), userID)
				stored = cloneTask(task)
				return nil
			},
		},
		&teamRepoForTaskMock{
			getMemberRoleFn: func(ctx context.Context, gotTeamID, gotUserID uint64) (model.TeamRole, error) {
				assert.Equal(t, teamID, gotTeamID)
				assert.Equal(t, uint64(9), gotUserID)
				return model.TeamRoleOwner, nil
			},
			isMemberFn: func(ctx context.Context, gotTeamID, gotUserID uint64) (bool, error) {
				assert.Equal(t, teamID, gotTeamID)
				assert.Equal(t, uint64(9), gotUserID)
				return true, nil
			},
		},
		&userRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.User, error) {
				switch id {
				case 9:
					return &model.User{ID: id, Username: "owner"}, nil
				default:
					return nil, repository.ErrNotFound
				}
			},
		},
		&taskAssignRepoStub{
			declinePendingByTaskIDFn: func(ctx context.Context, taskID uint64) error {
				declineCalled = true
				assert.Equal(t, uint64(80), taskID)
				return nil
			},
		},
		&activityLogRepoMock{
			createFn: func(ctx context.Context, log *model.ActivityLog) error {
				createdLog = log
				return nil
			},
		},
	)

	req := UpdateTaskRequest{AssigneeID: OptionalUint64{Set: true, Valid: true, Value: 9}}
	updated, err := svc.Update(context.Background(), 9, 80, req)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.True(t, declineCalled)
	require.NotNil(t, createdLog)
	assert.Equal(t, model.ActivityActionUpdateTaskInfo, createdLog.ActionType)
	assert.Contains(t, createdLog.Content, "将执行者从 未分配 修改为 owner")
}

func TestTaskServiceUpdateValidationFailureDoesNotWriteActivityLog(t *testing.T) {
	stored := model.Task{
		ID:       81,
		UserID:   9,
		Title:    "标题",
		Status:   model.TaskStatusTODO,
		Priority: model.PriorityMedium,
	}
	logCreated := false
	updateCalled := false
	svc := NewTaskService(
		&taskRepoMock{
			getByIDForUserScopeFn: func(ctx context.Context, id, userID uint64) (*model.Task, error) {
				cloned := cloneTask(&stored)
				return &cloned, nil
			},
			updateByIDForScopeFn: func(ctx context.Context, task *model.Task, userID uint64) error {
				updateCalled = true
				return nil
			},
		},
		&teamRepoForTaskMock{},
		&userRepoMock{},
		&taskAssignRepoStub{},
		&activityLogRepoMock{
			createFn: func(ctx context.Context, log *model.ActivityLog) error {
				logCreated = true
				return nil
			},
		},
	)

	blankTitle := "   "
	_, err := svc.Update(context.Background(), 9, 81, UpdateTaskRequest{Title: &blankTitle})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTaskTitle)
	assert.False(t, updateCalled)
	assert.False(t, logCreated)
}

func TestTaskServiceUpdateSkipsActivityLogWhenUnchanged(t *testing.T) {
	stored := model.Task{
		ID:       78,
		UserID:   9,
		Title:    "标题",
		Status:   model.TaskStatusTODO,
		Priority: model.PriorityMedium,
	}
	logCreated := false
	svc := NewTaskService(
		&taskRepoMock{
			getByIDForUserScopeFn: func(ctx context.Context, id, userID uint64) (*model.Task, error) {
				cloned := cloneTask(&stored)
				return &cloned, nil
			},
			updateByIDForScopeFn: func(ctx context.Context, task *model.Task, userID uint64) error {
				stored = cloneTask(task)
				return nil
			},
		},
		&teamRepoForTaskMock{},
		&userRepoMock{},
		&taskAssignRepoStub{},
		&activityLogRepoMock{
			createFn: func(ctx context.Context, log *model.ActivityLog) error {
				logCreated = true
				return nil
			},
		},
	)

	sameTitle := "标题"
	_, err := svc.Update(context.Background(), 9, 78, UpdateTaskRequest{Title: &sameTitle})
	require.NoError(t, err)
	assert.False(t, logCreated)
}

func TestTaskServiceDeleteWritesActivityLog(t *testing.T) {
	task := &model.Task{ID: 88, UserID: 9, Title: "修复 bug", Status: model.TaskStatusTODO}
	deleteCalled := false
	var createdLog *model.ActivityLog
	svc := NewTaskService(
		&taskRepoMock{
			getByIDForUserScopeFn: func(ctx context.Context, id, userID uint64) (*model.Task, error) {
				assert.Equal(t, task.ID, id)
				assert.Equal(t, task.UserID, userID)
				return task, nil
			},
			deleteByIDForScopeFn: func(ctx context.Context, id, userID uint64) error {
				deleteCalled = true
				assert.Equal(t, task.ID, id)
				assert.Equal(t, task.UserID, userID)
				return nil
			},
		},
		&teamRepoForTaskMock{},
		&userRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.User, error) {
				return &model.User{ID: id, Username: "alice"}, nil
			},
		},
		&taskAssignRepoStub{},
		&activityLogRepoMock{
			createFn: func(ctx context.Context, log *model.ActivityLog) error {
				createdLog = log
				return nil
			},
		},
	)

	err := svc.Delete(context.Background(), 9, 0, 88)
	require.NoError(t, err)
	assert.True(t, deleteCalled)
	require.NotNil(t, createdLog)
	assert.Equal(t, model.ActivityActionDeleteTask, createdLog.ActionType)
	assert.Equal(t, "删除了任务「修复 bug」", createdLog.Content)
}

func TestTaskServiceListActivityLogsReturnsPagedItems(t *testing.T) {
	createdAt := time.Date(2026, 6, 1, 12, 0, 0, 0, time.Local)
	svc := NewTaskService(
		&taskRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.Task, error) {
				return &model.Task{ID: id, UserID: 9, Title: "任务"}, nil
			},
		},
		&teamRepoForTaskMock{},
		&userRepoMock{},
		&taskAssignRepoStub{},
		&activityLogRepoMock{
			listByTaskIDFn: func(ctx context.Context, taskID uint64, page, size int) ([]model.ActivityLogListItem, int64, error) {
				assert.Equal(t, uint64(88), taskID)
				assert.Equal(t, 1, page)
				assert.Equal(t, 10, size)
				return []model.ActivityLogListItem{
					{
						ID:         1,
						ActorName:  "alice",
						ActionType: model.ActivityActionCreateTask,
						Content:    "创建了任务「任务」",
						CreatedAt:  createdAt,
					},
				}, 1, nil
			},
		},
	)

	list, total, page, size, err := svc.ListActivityLogs(context.Background(), 9, 88, 1, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.EqualValues(t, 1, total)
	assert.Equal(t, 1, page)
	assert.Equal(t, 10, size)
	assert.Equal(t, "2026-06-01 12:00:00", list[0].CreatedAt)
}

func TestTaskServiceListActivityLogsAllowsOwnerAfterTaskDeleted(t *testing.T) {
	deletedAt := time.Date(2026, 6, 2, 8, 0, 0, 0, time.Local)
	createdAt := time.Date(2026, 6, 1, 12, 0, 0, 0, time.Local)
	svc := NewTaskService(
		&taskRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.Task, error) {
				return &model.Task{ID: id, UserID: 9, Title: "已删除任务", DeletedAt: &deletedAt}, nil
			},
		},
		&teamRepoForTaskMock{},
		&userRepoMock{},
		&taskAssignRepoStub{},
		&activityLogRepoMock{
			listByTaskIDFn: func(ctx context.Context, taskID uint64, page, size int) ([]model.ActivityLogListItem, int64, error) {
				assert.Equal(t, uint64(88), taskID)
				return []model.ActivityLogListItem{
					{
						ID:         2,
						ActorName:  "alice",
						ActionType: model.ActivityActionDeleteTask,
						Content:    "删除了任务「已删除任务」",
						CreatedAt:  createdAt,
					},
				}, 1, nil
			},
		},
	)

	list, total, _, _, err := svc.ListActivityLogs(context.Background(), 9, 88, 1, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.EqualValues(t, 1, total)
	assert.Equal(t, model.ActivityActionDeleteTask, list[0].ActionType)
}

func TestTaskServiceListActivityLogsAllowsTeamMember(t *testing.T) {
	teamID := uint64(5)
	createdAt := time.Date(2026, 6, 3, 9, 30, 0, 0, time.Local)
	permissionChecked := false
	logListed := false

	svc := NewTaskService(
		&taskRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.Task, error) {
				assert.Equal(t, uint64(88), id)
				return &model.Task{ID: id, UserID: 1, TeamID: &teamID, Title: "团队任务"}, nil
			},
		},
		&teamRepoForTaskMock{
			getByIDForMemberScopeFn: func(ctx context.Context, gotTeamID, gotUserID uint64) (*model.Team, error) {
				permissionChecked = true
				assert.Equal(t, teamID, gotTeamID)
				assert.Equal(t, uint64(7), gotUserID)
				return &model.Team{ID: gotTeamID}, nil
			},
		},
		&userRepoMock{},
		&taskAssignRepoStub{},
		&activityLogRepoMock{
			listByTaskIDFn: func(ctx context.Context, taskID uint64, page, size int) ([]model.ActivityLogListItem, int64, error) {
				logListed = true
				assert.Equal(t, uint64(88), taskID)
				assert.Equal(t, 2, page)
				assert.Equal(t, 5, size)
				return []model.ActivityLogListItem{
					{
						ID:         3,
						ActorName:  "member",
						ActionType: model.ActivityActionUpdateTaskInfo,
						Content:    "将状态从 TODO 修改为 DOING",
						CreatedAt:  createdAt,
					},
				}, 1, nil
			},
		},
	)

	list, total, page, size, err := svc.ListActivityLogs(context.Background(), 7, 88, 2, 5)
	require.NoError(t, err)
	assert.True(t, permissionChecked)
	assert.True(t, logListed)
	require.Len(t, list, 1)
	assert.EqualValues(t, 1, total)
	assert.Equal(t, 2, page)
	assert.Equal(t, 5, size)
	assert.Equal(t, "member", list[0].ActorName)
	assert.Equal(t, "2026-06-03 09:30:00", list[0].CreatedAt)
}

func TestTaskServiceListActivityLogsRejectsForbiddenTeamTask(t *testing.T) {
	teamID := uint64(5)
	logListed := false

	svc := NewTaskService(
		&taskRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.Task, error) {
				return &model.Task{ID: id, UserID: 1, TeamID: &teamID, Title: "团队任务"}, nil
			},
		},
		&teamRepoForTaskMock{
			getByIDForMemberScopeFn: func(ctx context.Context, gotTeamID, gotUserID uint64) (*model.Team, error) {
				assert.Equal(t, teamID, gotTeamID)
				assert.Equal(t, uint64(7), gotUserID)
				return nil, repository.ErrNotFound
			},
		},
		&userRepoMock{},
		&taskAssignRepoStub{},
		&activityLogRepoMock{
			listByTaskIDFn: func(ctx context.Context, taskID uint64, page, size int) ([]model.ActivityLogListItem, int64, error) {
				logListed = true
				return nil, 0, nil
			},
		},
	)

	_, _, _, _, err := svc.ListActivityLogs(context.Background(), 7, 88, 1, 10)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrForbidden)
	assert.False(t, logListed)
}

func TestTaskServiceListActivityLogsRejectsForbiddenPersonalTask(t *testing.T) {
	svc := NewTaskService(
		&taskRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.Task, error) {
				return &model.Task{ID: id, UserID: 99, Title: "任务"}, nil
			},
		},
		&teamRepoForTaskMock{},
		&userRepoMock{},
		&taskAssignRepoStub{},
		&activityLogRepoMock{},
	)

	_, _, _, _, err := svc.ListActivityLogs(context.Background(), 9, 88, 1, 10)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrForbidden)
}

func TestTaskServiceListActivityLogsReturnsNotFoundForMissingTask(t *testing.T) {
	svc := NewTaskService(
		&taskRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.Task, error) {
				return nil, repository.ErrNotFound
			},
		},
		&teamRepoForTaskMock{},
		&userRepoMock{},
		&taskAssignRepoStub{},
		&activityLogRepoMock{},
	)

	_, _, _, _, err := svc.ListActivityLogs(context.Background(), 9, 404, 1, 10)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestTaskServiceAddDependencyMapsCycleError(t *testing.T) {
	svc := NewTaskService(
		&taskRepoMock{
			getByIDForUserScopeFn: func(ctx context.Context, id, userID uint64) (*model.Task, error) {
				assert.Equal(t, uint64(9), userID)
				switch id {
				case 20:
					return &model.Task{ID: id, UserID: userID, Status: model.TaskStatusTODO}, nil
				case 10:
					return &model.Task{ID: id, UserID: userID, Status: model.TaskStatusTODO}, nil
				default:
					return nil, repository.ErrNotFound
				}
			},
			addDependencyWithCycleCheckFn: func(ctx context.Context, predecessorID, successorID uint64) (*model.TaskDependency, error) {
				assert.Equal(t, uint64(10), predecessorID)
				assert.Equal(t, uint64(20), successorID)
				return nil, repository.ErrDependencyCycle
			},
		},
		&teamRepoForTaskMock{},
		&taskUserRepoStub{},
		&taskAssignRepoStub{},
	)

	dep, err := svc.AddDependency(context.Background(), 9, 20, 10)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDependencyCycle)
	assert.Nil(t, dep)
}

func TestTaskServiceAddDependencyRejectsCrossScope(t *testing.T) {
	teamID := uint64(5)
	svc := NewTaskService(
		&taskRepoMock{
			getByIDForUserScopeFn: func(ctx context.Context, id, userID uint64) (*model.Task, error) {
				assert.Equal(t, uint64(9), userID)
				switch id {
				case 20:
					return &model.Task{ID: id, UserID: 1, TeamID: &teamID, Status: model.TaskStatusTODO}, nil
				case 10:
					return &model.Task{ID: id, UserID: userID, Status: model.TaskStatusTODO}, nil
				default:
					return nil, repository.ErrNotFound
				}
			},
		},
		&teamRepoForTaskMock{},
		&taskUserRepoStub{},
		&taskAssignRepoStub{},
	)

	dep, err := svc.AddDependency(context.Background(), 9, 20, 10)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCrossTaskDependency)
	assert.Nil(t, dep)
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}

func strPtr(v string) *string {
	return &v
}
