package service

import (
	"context"
	"errors"
	"strings"

	"task-backend/internal/model"
	"task-backend/internal/repository"
)

var (
	ErrInvalidTeamName = errors.New("invalid team name")
	ErrInvalidTeamRole = errors.New("invalid team role")
)

type CreateTeamRequest struct {
	Name string `json:"name"`
}

type AddTeamMemberRequest struct {
	UserID uint64 `json:"user_id"`
	Role   string `json:"role"`
}

type TeamService interface {
	Create(ctx context.Context, ownerID uint64, req CreateTeamRequest) (*model.Team, error)
	ListMine(ctx context.Context, userID uint64) ([]model.Team, error)
	AddMember(ctx context.Context, operatorID, teamID uint64, req AddTeamMemberRequest) (*model.TeamMember, error)
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

func (s *teamService) ListMine(ctx context.Context, userID uint64) ([]model.Team, error) {
	return s.teamRepo.ListByUserID(ctx, userID)
}

func (s *teamService) AddMember(ctx context.Context, operatorID, teamID uint64, req AddTeamMemberRequest) (*model.TeamMember, error) {
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

	if _, err := s.userRepo.GetByID(ctx, req.UserID); err != nil {
		return nil, err
	}

	role := model.TeamRole(req.Role)
	if !isValidTeamRole(role) {
		return nil, ErrInvalidTeamRole
	}

	member := &model.TeamMember{
		TeamID: teamID,
		UserID: req.UserID,
		Role:   role,
	}
	if err := s.teamRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}
	return member, nil
}

func isValidTeamRole(role model.TeamRole) bool {
	switch role {
	case model.TeamRoleOwner, model.TeamRoleAdmin, model.TeamRoleMember:
		return true
	default:
		return false
	}
}
