package service

import (
	"context"
	"errors"
	"strings"

	"task-backend/internal/model"
	"task-backend/internal/repository"
)

var (
	ErrInvalidTeamName   = errors.New("invalid team name")
	ErrInvalidTeamRole   = errors.New("invalid team role")
	ErrInvalidTeamMember = errors.New("invalid team member")
	ErrTeamOwnerOnly     = errors.New("only owner can manage team members")
	ErrCannotRemoveSelf  = errors.New("owner cannot remove self from team")
	ErrCannotModifyOwner = errors.New("owner role cannot be changed")
)

type CreateTeamRequest struct {
	Name string `json:"name"`
}

type AddTeamMemberRequest struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
}

type UpdateTeamMemberRoleRequest struct {
	Role string `json:"role"`
}

type UpdateTeamRequest struct {
	Name string `json:"name"`
}

type TeamService interface {
	Create(ctx context.Context, ownerID uint64, req CreateTeamRequest) (*model.Team, error)
	ListMine(ctx context.Context, userID uint64, search string) ([]model.TeamWithRole, error)
	Update(ctx context.Context, operatorID, teamID uint64, req UpdateTeamRequest) error
	AddMember(ctx context.Context, operatorID, teamID uint64, req AddTeamMemberRequest) (*model.TeamMember, error)
	RemoveMember(ctx context.Context, operatorID, teamID, targetUserID uint64) error
	UpdateMemberRole(ctx context.Context, operatorID, teamID, targetUserID uint64, req UpdateTeamMemberRoleRequest) (*model.TeamMember, error)
	ListMembers(ctx context.Context, operatorID, teamID uint64) ([]model.TeamMemberDetail, error)
}

type teamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
}

func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository) TeamService {
	return &teamService{teamRepo: teamRepo, userRepo: userRepo}
}

func (s *teamService) Create(ctx context.Context, ownerID uint64, req CreateTeamRequest) (*model.Team, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 100 {
		return nil, ErrInvalidTeamName
	}

	team := &model.Team{
		Name:    name,
		OwnerID: ownerID,
	}
	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	owner := &model.TeamMember{
		TeamID: team.ID,
		UserID: ownerID,
		Role:   model.TeamRoleOwner,
	}
	if err := s.teamRepo.AddMember(ctx, owner); err != nil {
		return nil, err
	}
	return team, nil
}

func (s *teamService) ListMine(ctx context.Context, userID uint64, search string) ([]model.TeamWithRole, error) {
	return s.teamRepo.ListByUserID(ctx, userID, search)
}

func (s *teamService) Update(ctx context.Context, operatorID, teamID uint64, req UpdateTeamRequest) error {
	if _, err := s.teamRepo.GetByID(ctx, teamID); err != nil {
		return err
	}

	operatorRole, err := s.teamRepo.GetMemberRole(ctx, teamID, operatorID)
	if errors.Is(err, repository.ErrNotFound) {
		return repository.ErrForbidden
	}
	if err != nil {
		return err
	}
	if operatorRole != model.TeamRoleOwner {
		return ErrTeamOwnerOnly
	}

	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 100 {
		return ErrInvalidTeamName
	}

	return s.teamRepo.UpdateName(ctx, teamID, name)
}

func (s *teamService) AddMember(ctx context.Context, operatorID, teamID uint64, req AddTeamMemberRequest) (*model.TeamMember, error) {
	if _, err := s.teamRepo.GetByID(ctx, teamID); err != nil {
		return nil, err
	}

	operatorRole, err := s.teamRepo.GetMemberRole(ctx, teamID, operatorID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, repository.ErrForbidden
	}
	if err != nil {
		return nil, err
	}
	if operatorRole != model.TeamRoleOwner {
		return nil, ErrTeamOwnerOnly
	}

	targetUserID := req.UserID
	if targetUserID == 0 {
		username := strings.TrimSpace(req.Username)
		if username == "" {
			return nil, ErrInvalidTeamMember
		}
		targetUser, err := s.userRepo.GetByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		targetUserID = targetUser.ID
	}

	if _, err := s.userRepo.GetByID(ctx, targetUserID); err != nil {
		return nil, err
	}

	member := &model.TeamMember{
		TeamID: teamID,
		UserID: targetUserID,
		Role:   model.TeamRoleMember,
	}
	if err := s.teamRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *teamService) RemoveMember(ctx context.Context, operatorID, teamID, targetUserID uint64) error {
	if _, err := s.teamRepo.GetByID(ctx, teamID); err != nil {
		return err
	}

	operatorRole, err := s.teamRepo.GetMemberRole(ctx, teamID, operatorID)
	if errors.Is(err, repository.ErrNotFound) {
		return repository.ErrForbidden
	}
	if err != nil {
		return err
	}
	if operatorRole != model.TeamRoleOwner {
		return ErrTeamOwnerOnly
	}
	if operatorID == targetUserID {
		return ErrCannotRemoveSelf
	}

	targetRole, err := s.teamRepo.GetMemberRole(ctx, teamID, targetUserID)
	if err != nil {
		return err
	}
	if targetRole == model.TeamRoleOwner {
		return ErrCannotModifyOwner
	}

	return s.teamRepo.RemoveMember(ctx, teamID, targetUserID)
}

func (s *teamService) UpdateMemberRole(ctx context.Context, operatorID, teamID, targetUserID uint64, req UpdateTeamMemberRoleRequest) (*model.TeamMember, error) {
	if _, err := s.teamRepo.GetByID(ctx, teamID); err != nil {
		return nil, err
	}

	operatorRole, err := s.teamRepo.GetMemberRole(ctx, teamID, operatorID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, repository.ErrForbidden
	}
	if err != nil {
		return nil, err
	}
	if operatorRole != model.TeamRoleOwner {
		return nil, ErrTeamOwnerOnly
	}

	targetRole, err := s.teamRepo.GetMemberRole(ctx, teamID, targetUserID)
	if err != nil {
		return nil, err
	}
	if targetRole == model.TeamRoleOwner {
		return nil, ErrCannotModifyOwner
	}

	role := model.TeamRole(strings.TrimSpace(req.Role))
	if role != model.TeamRoleAdmin && role != model.TeamRoleMember {
		return nil, ErrInvalidTeamRole
	}

	if err := s.teamRepo.UpdateMemberRole(ctx, teamID, targetUserID, role); err != nil {
		return nil, err
	}

	return &model.TeamMember{
		TeamID: teamID,
		UserID: targetUserID,
		Role:   role,
	}, nil
}

func (s *teamService) ListMembers(ctx context.Context, operatorID, teamID uint64) ([]model.TeamMemberDetail, error) {
	if _, err := s.teamRepo.GetByID(ctx, teamID); err != nil {
		return nil, err
	}

	isMember, err := s.teamRepo.IsMember(ctx, teamID, operatorID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, repository.ErrForbidden
	}

	return s.teamRepo.ListMembers(ctx, teamID)
}

func isValidTeamRole(role model.TeamRole) bool {
	switch role {
	case model.TeamRoleOwner, model.TeamRoleAdmin, model.TeamRoleMember:
		return true
	default:
		return false
	}
}
