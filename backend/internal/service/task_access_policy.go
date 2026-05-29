package service

import (
	"context"
	"errors"

	"task-backend/internal/model"
	"task-backend/internal/repository"
)

func (s *taskService) requireExistingTeamMember(ctx context.Context, teamID, userID uint64) error {
	if _, err := s.teamRepo.GetByIDForMemberScope(ctx, teamID, userID); errors.Is(err, repository.ErrNotFound) {
		return repository.ErrForbidden
	} else if err != nil {
		return err
	}
	return nil
}

func (s *taskService) requireActiveTeamMember(ctx context.Context, teamID, userID uint64) error {
	if _, err := s.teamRepo.GetByIDForMemberScope(ctx, teamID, userID); errors.Is(err, repository.ErrNotFound) {
		return repository.ErrForbidden
	} else if err != nil {
		return err
	}
	return nil
}

func (s *taskService) requireTeamManager(ctx context.Context, teamID, userID uint64) error {
	role, err := s.teamRepo.GetMemberRole(ctx, teamID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return repository.ErrForbidden
	}
	if err != nil {
		return err
	}
	if !canManageTeamTasks(role) {
		return ErrTeamRoleDenied
	}
	return nil
}

func (s *taskService) requireTaskReadAccess(ctx context.Context, task *model.Task, userID uint64) error {
	if task.TeamID == nil {
		if task.UserID != userID {
			return repository.ErrForbidden
		}
		return nil
	}
	return s.requireActiveTeamMember(ctx, *task.TeamID, userID)
}
