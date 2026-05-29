package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-backend/internal/model"
	"task-backend/internal/repository"
)

type teamRepoMock struct {
	getByIDFn          func(ctx context.Context, teamID uint64) (*model.Team, error)
	getMemberRoleFn    func(ctx context.Context, teamID, userID uint64) (model.TeamRole, error)
	isMemberFn         func(ctx context.Context, teamID, userID uint64) (bool, error)
	addMemberFn        func(ctx context.Context, member *model.TeamMember) error
	listByUserIDFn     func(ctx context.Context, userID uint64, search string) ([]model.TeamWithRole, error)
	updateNameFn       func(ctx context.Context, teamID uint64, name string) error
	removeMemberFn     func(ctx context.Context, teamID, userID uint64) error
	updateMemberRoleFn func(ctx context.Context, teamID, userID uint64, role model.TeamRole) error
	listMembersFn      func(ctx context.Context, teamID uint64) ([]model.TeamMemberDetail, error)
	isDeletedFn        func(ctx context.Context, teamID uint64) (bool, error)
}

func (m *teamRepoMock) Create(ctx context.Context, team *model.Team) error {
	panic("unexpected call to Create")
}

func (m *teamRepoMock) ListByUserID(ctx context.Context, userID uint64, search string) ([]model.TeamWithRole, error) {
	if m.listByUserIDFn == nil {
		panic("unexpected call to ListByUserID")
	}
	return m.listByUserIDFn(ctx, userID, search)
}

func (m *teamRepoMock) GetByID(ctx context.Context, teamID uint64) (*model.Team, error) {
	if m.getByIDFn == nil {
		panic("unexpected call to GetByID")
	}
	return m.getByIDFn(ctx, teamID)
}

func (m *teamRepoMock) UpdateName(ctx context.Context, teamID uint64, name string) error {
	if m.updateNameFn == nil {
		panic("unexpected call to UpdateName")
	}
	return m.updateNameFn(ctx, teamID, name)
}

func (m *teamRepoMock) AddMember(ctx context.Context, member *model.TeamMember) error {
	if m.addMemberFn == nil {
		panic("unexpected call to AddMember")
	}
	return m.addMemberFn(ctx, member)
}

func (m *teamRepoMock) RemoveMember(ctx context.Context, teamID, userID uint64) error {
	if m.removeMemberFn == nil {
		panic("unexpected call to RemoveMember")
	}
	return m.removeMemberFn(ctx, teamID, userID)
}

func (m *teamRepoMock) UpdateMemberRole(ctx context.Context, teamID, userID uint64, role model.TeamRole) error {
	if m.updateMemberRoleFn == nil {
		panic("unexpected call to UpdateMemberRole")
	}
	return m.updateMemberRoleFn(ctx, teamID, userID, role)
}

func (m *teamRepoMock) IsMember(ctx context.Context, teamID, userID uint64) (bool, error) {
	if m.isMemberFn == nil {
		panic("unexpected call to IsMember")
	}
	return m.isMemberFn(ctx, teamID, userID)
}

func (m *teamRepoMock) GetMemberRole(ctx context.Context, teamID, userID uint64) (model.TeamRole, error) {
	if m.getMemberRoleFn == nil {
		panic("unexpected call to GetMemberRole")
	}
	return m.getMemberRoleFn(ctx, teamID, userID)
}

func (m *teamRepoMock) ListMembers(ctx context.Context, teamID uint64) ([]model.TeamMemberDetail, error) {
	if m.listMembersFn == nil {
		panic("unexpected call to ListMembers")
	}
	return m.listMembersFn(ctx, teamID)
}

func (m *teamRepoMock) TransferOwnership(ctx context.Context, teamID, oldOwnerID, newOwnerID uint64) error {
	panic("unexpected call to TransferOwnership")
}
func (m *teamRepoMock) IsDeleted(ctx context.Context, teamID uint64) (bool, error) {
	if m.isDeletedFn != nil {
		return m.isDeletedFn(ctx, teamID)
	}
	return false, nil
}

type userRepoMock struct {
	getByIDFn       func(ctx context.Context, id uint64) (*model.User, error)
	getByUsernameFn func(ctx context.Context, username string) (*model.User, error)
}

func (m *userRepoMock) Create(ctx context.Context, user *model.User) error {
	panic("unexpected call to Create")
}

func (m *userRepoMock) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	if m.getByUsernameFn == nil {
		panic("unexpected call to GetByUsername")
	}
	return m.getByUsernameFn(ctx, username)
}

func (m *userRepoMock) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	if m.getByIDFn == nil {
		panic("unexpected call to GetByID")
	}
	return m.getByIDFn(ctx, id)
}

func (m *userRepoMock) UpdatePasswordByID(ctx context.Context, id uint64, passwordHash string) error {
	panic("unexpected call to UpdatePasswordByID")
}

type inviteRepoMock struct {
	getPendingByTeamAndInviteeFn func(ctx context.Context, teamID, inviteeID uint64) (*model.TeamInvitation, error)
	createFn                     func(ctx context.Context, invite *model.TeamInvitation) error
	getByIDFn                    func(ctx context.Context, id uint64) (*model.TeamInvitation, error)
	updateStatusFn               func(ctx context.Context, id uint64, status model.InvitationStatus) error
}

func (m *inviteRepoMock) Create(ctx context.Context, invite *model.TeamInvitation) error {
	if m.createFn == nil {
		panic("unexpected call to Create")
	}
	return m.createFn(ctx, invite)
}

func (m *inviteRepoMock) GetByID(ctx context.Context, id uint64) (*model.TeamInvitation, error) {
	if m.getByIDFn == nil {
		panic("unexpected call to GetByID")
	}
	return m.getByIDFn(ctx, id)
}

func (m *inviteRepoMock) GetPendingByTeamAndInvitee(ctx context.Context, teamID, inviteeID uint64) (*model.TeamInvitation, error) {
	if m.getPendingByTeamAndInviteeFn == nil {
		panic("unexpected call to GetPendingByTeamAndInvitee")
	}
	return m.getPendingByTeamAndInviteeFn(ctx, teamID, inviteeID)
}

func (m *inviteRepoMock) ListPendingByInvitee(ctx context.Context, inviteeID uint64) ([]model.TeamInvitationDetail, error) {
	panic("unexpected call to ListPendingByInvitee")
}

func (m *inviteRepoMock) UpdateStatus(ctx context.Context, id uint64, status model.InvitationStatus) error {
	if m.updateStatusFn == nil {
		panic("unexpected call to UpdateStatus")
	}
	return m.updateStatusFn(ctx, id, status)
}

func TestTeamServiceAddMemberCreatesPendingInvitationByUsername(t *testing.T) {
	const (
		operatorID = uint64(11)
		teamID     = uint64(21)
		targetID   = uint64(31)
	)

	var createdInvitation *model.TeamInvitation
	svc := NewTeamService(
		&teamRepoMock{
			getByIDFn: func(ctx context.Context, gotTeamID uint64) (*model.Team, error) {
				assert.Equal(t, teamID, gotTeamID)
				return &model.Team{ID: gotTeamID}, nil
			},
			getMemberRoleFn: func(ctx context.Context, gotTeamID, gotUserID uint64) (model.TeamRole, error) {
				assert.Equal(t, teamID, gotTeamID)
				assert.Equal(t, operatorID, gotUserID)
				return model.TeamRoleOwner, nil
			},
			isMemberFn: func(ctx context.Context, gotTeamID, gotUserID uint64) (bool, error) {
				assert.Equal(t, teamID, gotTeamID)
				assert.Equal(t, targetID, gotUserID)
				return false, nil
			},
		},
		&userRepoMock{
			getByUsernameFn: func(ctx context.Context, username string) (*model.User, error) {
				assert.Equal(t, "alice_01", username)
				return &model.User{ID: targetID, Username: username}, nil
			},
			getByIDFn: func(ctx context.Context, gotUserID uint64) (*model.User, error) {
				assert.Equal(t, targetID, gotUserID)
				return &model.User{ID: gotUserID}, nil
			},
		},
		&inviteRepoMock{
			getPendingByTeamAndInviteeFn: func(ctx context.Context, gotTeamID, gotInviteeID uint64) (*model.TeamInvitation, error) {
				assert.Equal(t, teamID, gotTeamID)
				assert.Equal(t, targetID, gotInviteeID)
				return nil, repository.ErrNotFound
			},
			createFn: func(ctx context.Context, invite *model.TeamInvitation) error {
				createdInvitation = invite
				return nil
			},
		},
	)

	member, err := svc.AddMember(context.Background(), operatorID, teamID, AddTeamMemberRequest{
		Username: "alice_01",
	})
	require.NoError(t, err)
	require.NotNil(t, member)
	require.NotNil(t, createdInvitation)

	assert.Equal(t, teamID, member.TeamID)
	assert.Equal(t, targetID, member.UserID)
	assert.Equal(t, model.TeamRoleMember, member.Role)
	assert.Equal(t, teamID, createdInvitation.TeamID)
	assert.Equal(t, operatorID, createdInvitation.InviterID)
	assert.Equal(t, targetID, createdInvitation.InviteeID)
	assert.Equal(t, model.InvitationStatusPending, createdInvitation.Status)
}

func TestTeamServiceAddMemberWrapsUsernameLookupError(t *testing.T) {
	svc := NewTeamService(
		&teamRepoMock{
			getByIDFn: func(ctx context.Context, teamID uint64) (*model.Team, error) {
				return &model.Team{ID: teamID}, nil
			},
			getMemberRoleFn: func(ctx context.Context, teamID, userID uint64) (model.TeamRole, error) {
				return model.TeamRoleOwner, nil
			},
			isMemberFn: func(ctx context.Context, teamID, userID uint64) (bool, error) {
				return false, nil
			},
		},
		&userRepoMock{
			getByUsernameFn: func(ctx context.Context, username string) (*model.User, error) {
				return nil, repository.ErrNotFound
			},
		},
		&inviteRepoMock{},
	)

	_, err := svc.AddMember(context.Background(), 1, 2, AddTeamMemberRequest{Username: "ghost_user"})
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Contains(t, err.Error(), "get user by username ghost_user")
}

func TestTeamServiceAddMemberRejectsNonOwner(t *testing.T) {
	svc := NewTeamService(
		&teamRepoMock{
			getByIDFn: func(ctx context.Context, teamID uint64) (*model.Team, error) {
				return &model.Team{ID: teamID}, nil
			},
			getMemberRoleFn: func(ctx context.Context, teamID, userID uint64) (model.TeamRole, error) {
				return model.TeamRoleMember, nil
			},
		},
		&userRepoMock{},
		&inviteRepoMock{},
	)

	_, err := svc.AddMember(context.Background(), 1, 2, AddTeamMemberRequest{UserID: 3})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTeamOwnerOnly))
}

func TestTeamServiceAcceptInvitationAddsMemberAndUpdatesStatus(t *testing.T) {
	const (
		userID       = uint64(41)
		invitationID = uint64(51)
		teamID       = uint64(61)
	)

	addMemberCalled := false
	updateStatusCalled := false

	svc := NewTeamService(
		&teamRepoMock{
			getByIDFn: func(ctx context.Context, gotTeamID uint64) (*model.Team, error) {
				assert.Equal(t, teamID, gotTeamID)
				return &model.Team{ID: gotTeamID}, nil
			},
			isMemberFn: func(ctx context.Context, gotTeamID, gotUserID uint64) (bool, error) {
				assert.Equal(t, teamID, gotTeamID)
				assert.Equal(t, userID, gotUserID)
				return false, nil
			},
			addMemberFn: func(ctx context.Context, member *model.TeamMember) error {
				addMemberCalled = true
				assert.Equal(t, teamID, member.TeamID)
				assert.Equal(t, userID, member.UserID)
				assert.Equal(t, model.TeamRoleMember, member.Role)
				return nil
			},
		},
		&userRepoMock{},
		&inviteRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.TeamInvitation, error) {
				assert.Equal(t, invitationID, id)
				return &model.TeamInvitation{
					ID:        invitationID,
					TeamID:    teamID,
					InviteeID: userID,
					Status:    model.InvitationStatusPending,
				}, nil
			},
			updateStatusFn: func(ctx context.Context, id uint64, status model.InvitationStatus) error {
				updateStatusCalled = true
				assert.Equal(t, invitationID, id)
				assert.Equal(t, model.InvitationStatusAccepted, status)
				return nil
			},
		},
	)

	err := svc.AcceptInvitation(context.Background(), userID, invitationID)
	require.NoError(t, err)
	assert.True(t, addMemberCalled)
	assert.True(t, updateStatusCalled)
}

func TestTeamServiceAcceptInvitationRejectsWrongInvitee(t *testing.T) {
	svc := NewTeamService(
		&teamRepoMock{},
		&userRepoMock{},
		&inviteRepoMock{
			getByIDFn: func(ctx context.Context, id uint64) (*model.TeamInvitation, error) {
				return &model.TeamInvitation{
					ID:        id,
					TeamID:    2,
					InviteeID: 99,
					Status:    model.InvitationStatusPending,
				}, nil
			},
		},
	)

	err := svc.AcceptInvitation(context.Background(), 1, 7)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrForbidden)
}

func TestTeamServiceLeaveRejectsOwner(t *testing.T) {
	svc := NewTeamServiceWithTaskRepo(
		&teamRepoMock{
			getByIDFn: func(ctx context.Context, teamID uint64) (*model.Team, error) {
				return &model.Team{ID: teamID}, nil
			},
			getMemberRoleFn: func(ctx context.Context, teamID, userID uint64) (model.TeamRole, error) {
				assert.Equal(t, uint64(3), teamID)
				assert.Equal(t, uint64(9), userID)
				return model.TeamRoleOwner, nil
			},
		},
		&userRepoMock{},
		&inviteRepoMock{},
		nil,
	)

	err := svc.Leave(context.Background(), 9, 3)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrOwnerCannotLeave)
}

func TestTeamServiceLeaveRejectsNonMember(t *testing.T) {
	svc := NewTeamServiceWithTaskRepo(
		&teamRepoMock{
			getByIDFn: func(ctx context.Context, teamID uint64) (*model.Team, error) {
				return &model.Team{ID: teamID}, nil
			},
			getMemberRoleFn: func(ctx context.Context, teamID, userID uint64) (model.TeamRole, error) {
				return "", repository.ErrNotFound
			},
		},
		&userRepoMock{},
		&inviteRepoMock{},
		nil,
	)

	err := svc.Leave(context.Background(), 9, 3)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrForbidden)
}

func TestTeamServiceGetTeamRoleOrForbiddenSuccess(t *testing.T) {
	svc := NewTeamServiceWithTaskRepo(
		&teamRepoMock{
			getMemberRoleFn: func(ctx context.Context, teamID, userID uint64) (model.TeamRole, error) {
				assert.Equal(t, uint64(3), teamID)
				assert.Equal(t, uint64(9), userID)
				return model.TeamRoleAdmin, nil
			},
		},
		&userRepoMock{},
		&inviteRepoMock{},
		nil,
	)

	role, err := svc.(*teamService).getTeamRoleOrForbidden(context.Background(), 3, 9)
	require.NoError(t, err)
	assert.Equal(t, model.TeamRoleAdmin, role)
}
