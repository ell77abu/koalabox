package service

import (
	"context"
	"errors"

	"task-backend/internal/model"
	"task-backend/internal/repository"
)

func (s *teamService) requireTeamOwner(ctx context.Context, teamID, userID uint64) error {
	if _, err := s.teamRepo.GetByID(ctx, teamID); err != nil {
		return err
	}
	role, err := s.getTeamRoleOrForbidden(ctx, teamID, userID)
	if err != nil {
		return err
	}
	if role != model.TeamRoleOwner {
		return ErrTeamOwnerOnly
	}
	return nil
}

func (s *teamService) requireTeamMember(ctx context.Context, teamID, userID uint64) error {
	if _, err := s.teamRepo.GetByID(ctx, teamID); err != nil {
		return err
	}
	isMember, err := s.teamRepo.IsMember(ctx, teamID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return repository.ErrForbidden
	}
	return nil
}

func (s *teamService) getTeamRoleOrForbidden(ctx context.Context, teamID, userID uint64) (model.TeamRole, error) {
	role, err := s.teamRepo.GetMemberRole(ctx, teamID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return "", repository.ErrForbidden
	}
	if err != nil {
		return "", err
	}
	return role, nil
}

