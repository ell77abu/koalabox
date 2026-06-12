package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"task-backend/internal/model"
	"task-backend/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrInvalidTeamName       = errors.New("invalid team name")
	ErrInvalidTeamRole       = errors.New("invalid team role")
	ErrInvalidTeamMember     = errors.New("invalid team member")
	ErrTeamOwnerOnly         = errors.New("only owner can manage team members")
	ErrCannotRemoveSelf      = errors.New("owner cannot remove self from team")
	ErrCannotModifyOwner     = errors.New("owner role cannot be changed")
	ErrInvitationNotPending  = errors.New("invitation is not pending")
	ErrOwnerCannotLeave      = errors.New("owner cannot leave team")
	ErrInvalidTransferTarget = errors.New("invalid transfer target")
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

type TransferTeamRequest struct {
	NewOwnerID uint64 `json:"new_owner_id"`
}

type TeamService interface {
	Create(ctx context.Context, ownerID uint64, req CreateTeamRequest) (*model.Team, error)
	ListMine(ctx context.Context, userID uint64, search string) ([]model.TeamWithRole, error)
	Update(ctx context.Context, operatorID, teamID uint64, req UpdateTeamRequest) error
	AddMember(ctx context.Context, operatorID, teamID uint64, req AddTeamMemberRequest) (*model.TeamMember, error)
	RemoveMember(ctx context.Context, operatorID, teamID, targetUserID uint64) error
	UpdateMemberRole(ctx context.Context, operatorID, teamID, targetUserID uint64, req UpdateTeamMemberRoleRequest) (*model.TeamMember, error)
	ListMembers(ctx context.Context, operatorID, teamID uint64) ([]model.TeamMemberDetail, error)
	ListPendingInvitations(ctx context.Context, userID uint64) ([]model.TeamInvitationDetail, error)
	AcceptInvitation(ctx context.Context, userID, invitationID uint64) error
	DeclineInvitation(ctx context.Context, userID, invitationID uint64) error
	Leave(ctx context.Context, userID, teamID uint64) error
	TransferOwnership(ctx context.Context, operatorID, teamID, newOwnerID uint64) error
}

type teamService struct {
	teamRepo        repository.TeamRepository
	userRepo        repository.UserRepository
	inviteRepo      repository.TeamInvitationRepository
	taskRepo        repository.TaskRepository
	activityLogRepo repository.ActivityLogRepository
}

func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository, inviteRepo repository.TeamInvitationRepository) TeamService {
	return NewTeamServiceWithTaskRepo(teamRepo, userRepo, inviteRepo, nil)
}

func NewTeamServiceWithTaskRepo(teamRepo repository.TeamRepository, userRepo repository.UserRepository, inviteRepo repository.TeamInvitationRepository, taskRepo repository.TaskRepository, activityLogRepo ...repository.ActivityLogRepository) TeamService {
	var logRepo repository.ActivityLogRepository
	if len(activityLogRepo) > 0 {
		logRepo = activityLogRepo[0]
	}
	return &teamService{
		teamRepo:        teamRepo,
		userRepo:        userRepo,
		inviteRepo:      inviteRepo,
		taskRepo:        taskRepo,
		activityLogRepo: logRepo,
	}
}

func (s *teamService) Create(ctx context.Context, ownerID uint64, req CreateTeamRequest) (*model.Team, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 100 {
		return nil, ErrInvalidTeamName
	}
	if _, err := s.teamRepo.GetByName(ctx, name); err == nil {
		return nil, repository.ErrConflict
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	db, err := getTeamRepoDB(s.teamRepo)
	if err != nil {
		return nil, err
	}

	team := &model.Team{
		Name:    name,
		OwnerID: ownerID,
	}
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		teamRepoTx := repository.NewTeamRepository(tx)
		if err := teamRepoTx.Create(ctx, team); err != nil {
			return err
		}

		owner := &model.TeamMember{
			TeamID: team.ID,
			UserID: ownerID,
			Role:   model.TeamRoleOwner,
		}
		return teamRepoTx.AddMember(ctx, owner)
	}); err != nil {
		return nil, err
	}
	return team, nil
}

func (s *teamService) ListMine(ctx context.Context, userID uint64, search string) ([]model.TeamWithRole, error) {
	list, err := s.teamRepo.ListByUserID(ctx, userID, search)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]model.TeamWithRole, 0)
	}
	return list, nil
}

func (s *teamService) Update(ctx context.Context, operatorID, teamID uint64, req UpdateTeamRequest) error {
	if err := s.requireTeamOwner(ctx, teamID, operatorID); err != nil {
		return err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 100 {
		return ErrInvalidTeamName
	}
	if existing, err := s.teamRepo.GetByName(ctx, name); err == nil {
		if existing.ID != teamID {
			return repository.ErrConflict
		}
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	return s.teamRepo.UpdateName(ctx, teamID, name)
}

func (s *teamService) AddMember(ctx context.Context, operatorID, teamID uint64, req AddTeamMemberRequest) (*model.TeamMember, error) {
	if err := s.requireTeamOwner(ctx, teamID, operatorID); err != nil {
		return nil, err
	}

	targetUserID := req.UserID
	if targetUserID == 0 {
		username := strings.TrimSpace(req.Username)
		if username == "" {
			return nil, ErrInvalidTeamMember
		}
		targetUser, err := s.userRepo.GetByUsername(ctx, username)
		if err != nil {
			// 包装下游错误，保留原始错误以便上层判断，同时提供更明确的错误上下文
			return nil, fmt.Errorf("get user by username %s: %w", username, err)
		}
		targetUserID = targetUser.ID
	}

	if _, err := s.userRepo.GetByID(ctx, targetUserID); err != nil {
		return nil, err
	}

	isMember, err := s.teamRepo.IsMember(ctx, teamID, targetUserID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, repository.ErrConflict
	}

	if _, err := s.inviteRepo.GetPendingByTeamAndInvitee(ctx, teamID, targetUserID); err == nil {
		return nil, repository.ErrConflict
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	invite := &model.TeamInvitation{
		TeamID:    teamID,
		InviterID: operatorID,
		InviteeID: targetUserID,
		Status:    model.InvitationStatusPending,
	}
	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, err
	}
	return &model.TeamMember{
		TeamID: teamID,
		UserID: targetUserID,
		Role:   model.TeamRoleMember,
	}, nil
}

func (s *teamService) RemoveMember(ctx context.Context, operatorID, teamID, targetUserID uint64) error {
	if err := s.requireTeamOwner(ctx, teamID, operatorID); err != nil {
		return err
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
	departingUsername := ""
	if s.activityLogRepo != nil {
		targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
		if err != nil {
			return err
		}
		departingUsername = targetUser.Username
	}
	if s.taskRepo == nil {
		return errors.New("task repository is not configured")
	}
	db, err := getTeamRepoDB(s.teamRepo)
	if err != nil {
		return err
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		teamRepoTx := repository.NewTeamRepository(tx)
		if err := s.cleanupDepartingMember(ctx, tx, teamID, targetUserID, operatorID, departingUsername); err != nil {
			return err
		}
		return teamRepoTx.RemoveMember(ctx, teamID, targetUserID)
	})
}

func (s *teamService) UpdateMemberRole(ctx context.Context, operatorID, teamID, targetUserID uint64, req UpdateTeamMemberRoleRequest) (*model.TeamMember, error) {
	if err := s.requireTeamOwner(ctx, teamID, operatorID); err != nil {
		return nil, err
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
	if err := s.requireTeamMember(ctx, teamID, operatorID); err != nil {
		return nil, err
	}

	return s.teamRepo.ListMembers(ctx, teamID)
}

func (s *teamService) ListPendingInvitations(ctx context.Context, userID uint64) ([]model.TeamInvitationDetail, error) {
	return s.inviteRepo.ListPendingByInvitee(ctx, userID)
}

func (s *teamService) AcceptInvitation(ctx context.Context, userID, invitationID uint64) error {
	invite, err := s.inviteRepo.GetByIDForInvitee(ctx, invitationID, userID)
	if err != nil {
		return err
	}
	if invite.Status != model.InvitationStatusPending {
		return ErrInvitationNotPending
	}

	return s.inviteRepo.AcceptPendingForInvitee(ctx, invitationID, userID)
}

func (s *teamService) DeclineInvitation(ctx context.Context, userID, invitationID uint64) error {
	invite, err := s.inviteRepo.GetByIDForInvitee(ctx, invitationID, userID)
	if err != nil {
		return err
	}
	if invite.Status != model.InvitationStatusPending {
		return ErrInvitationNotPending
	}

	return s.inviteRepo.UpdateStatus(ctx, invitationID, model.InvitationStatusDeclined)
}

func (s *teamService) Leave(ctx context.Context, userID, teamID uint64) error {
	role, err := s.getTeamRoleOrForbidden(ctx, teamID, userID)
	if err != nil {
		return err
	}
	if role == model.TeamRoleOwner {
		return ErrOwnerCannotLeave
	}
	if s.taskRepo == nil {
		return errors.New("task repository is not configured")
	}
	departingUsername := ""
	if s.activityLogRepo != nil {
		departingUser, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return err
		}
		departingUsername = departingUser.Username
	}
	db, err := getTeamRepoDB(s.teamRepo)
	if err != nil {
		return err
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		teamRepoTx := repository.NewTeamRepository(tx)
		if err := s.cleanupDepartingMember(ctx, tx, teamID, userID, userID, departingUsername); err != nil {
			return err
		}
		return teamRepoTx.RemoveMember(ctx, teamID, userID)
	})
}

func (s *teamService) cleanupDepartingMember(ctx context.Context, tx *gorm.DB, teamID, userID, actorID uint64, departingUsername string) error {
	taskRepoTx := repository.NewTaskRepository(tx)
	lister, ok := taskRepoTx.(interface {
		ListOpenAssignedTeamTasks(ctx context.Context, teamID, assigneeID uint64) ([]model.Task, error)
	})
	if !ok {
		return errors.New("task repository does not support listing open assigned tasks")
	}
	affectedTasks, err := lister.ListOpenAssignedTeamTasks(ctx, teamID, userID)
	if err != nil {
		return err
	}
	clearer, ok := taskRepoTx.(interface {
		ClearAssigneeForOpenTeamTasks(ctx context.Context, teamID, assigneeID uint64) error
	})
	if !ok {
		return errors.New("task repository does not support clearing open assigned tasks")
	}
	if err := clearer.ClearAssigneeForOpenTeamTasks(ctx, teamID, userID); err != nil {
		return err
	}
	if err := s.writeMemberLeaveActivityLogs(ctx, tx, actorID, departingUsername, affectedTasks); err != nil {
		return err
	}

	// 成员离队后，回收该成员在团队内的相关分配请求，避免残留可见性与状态歧义。
	// task_assignment_requests 有 (task_id, assignee_id, status) 唯一约束：
	// 同一成员反复加入/退出时可能已有 DECLINED 历史，直接批量更新会撞唯一键。
	activeStatuses := []model.InvitationStatus{
		model.InvitationStatusPending,
		model.InvitationStatusAccepted,
	}

	if err := tx.WithContext(ctx).
		Where("team_id = ?", teamID).
		Where("assignee_id = ?", userID).
		Where("status IN ?", activeStatuses).
		Where(`EXISTS (
			SELECT 1
			FROM task_assignment_requests declined
			WHERE declined.task_id = task_assignment_requests.task_id
				AND declined.assignee_id = task_assignment_requests.assignee_id
				AND declined.team_id = ?
				AND declined.assignee_id = ?
				AND declined.status = ?
		)`, teamID, userID, model.InvitationStatusDeclined).
		Delete(&model.TaskAssignmentRequest{}).Error; err != nil {
		return err
	}

	if err := tx.WithContext(ctx).
		Where("team_id = ?", teamID).
		Where("assignee_id = ?", userID).
		Where("status IN ?", activeStatuses).
		Where(`id NOT IN (
			SELECT keep_id
			FROM (
				SELECT MAX(id) AS keep_id
				FROM task_assignment_requests
				WHERE team_id = ?
					AND assignee_id = ?
					AND status IN ?
				GROUP BY task_id, assignee_id
			) kept
		)`, teamID, userID, activeStatuses).
		Delete(&model.TaskAssignmentRequest{}).Error; err != nil {
		return err
	}

	return tx.WithContext(ctx).
		Model(&model.TaskAssignmentRequest{}).
		Where("team_id = ?", teamID).
		Where("assignee_id = ?", userID).
		Where("status IN ?", activeStatuses).
		Updates(map[string]interface{}{
			"status": model.InvitationStatusDeclined,
		}).Error
}

func (s *teamService) writeMemberLeaveActivityLogs(ctx context.Context, tx *gorm.DB, actorID uint64, departingUsername string, tasks []model.Task) error {
	if s.activityLogRepo == nil || len(tasks) == 0 {
		return nil
	}
	actor, err := s.userRepo.GetByID(ctx, actorID)
	if err != nil {
		return err
	}
	logRepoTx := repository.NewActivityLogRepository(tx)
	content := fmt.Sprintf("成员 %s 离开团队，任务执行者已置空", departingUsername)
	for _, task := range tasks {
		if err := logRepoTx.Create(ctx, &model.ActivityLog{
			ActorID:    actorID,
			ActorName:  actor.Username,
			ActionType: model.ActivityActionAssigneeChangedByMemberLeave,
			TargetType: model.ActivityTargetTask,
			TargetID:   task.ID,
			Scope:      model.ActivityScopeTeam,
			TeamID:     cloneUint64Ptr(task.TeamID),
			Content:    content,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *teamService) TransferOwnership(ctx context.Context, operatorID, teamID, newOwnerID uint64) error {
	if newOwnerID == 0 || newOwnerID == operatorID {
		return ErrInvalidTransferTarget
	}

	if err := s.requireTeamOwner(ctx, teamID, operatorID); err != nil {
		return err
	}

	if _, err := s.teamRepo.GetMemberRole(ctx, teamID, newOwnerID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidTransferTarget
		}
		return err
	}

	return s.teamRepo.TransferOwnership(ctx, teamID, operatorID, newOwnerID)
}

func (s *teamService) DeleteTeam(ctx context.Context, operatorID, teamID uint64) error {
	if err := s.requireTeamOwner(ctx, teamID, operatorID); err != nil {
		return err
	}
	if s.taskRepo == nil {
		return errors.New("task repository is not configured")
	}
	db, err := getTeamRepoDB(s.teamRepo)
	if err != nil {
		return err
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		teamRepoTx := repository.NewTeamRepository(tx)
		taskRepoTx := repository.NewTaskRepository(tx)

		taskDeleter, ok := taskRepoTx.(interface {
			SoftDeleteByTeamID(ctx context.Context, teamID uint64) error
		})
		if !ok {
			return errors.New("task repository does not support soft delete by team")
		}
		if err := taskDeleter.SoftDeleteByTeamID(ctx, teamID); err != nil {
			return err
		}

		memberDeleter, ok := teamRepoTx.(interface {
			SoftDeleteMembersByTeamID(ctx context.Context, teamID uint64) error
		})
		if !ok {
			return errors.New("team repository does not support soft delete members")
		}
		if err := memberDeleter.SoftDeleteMembersByTeamID(ctx, teamID); err != nil {
			return err
		}

		// 团队解散后，回收仍有效的成员邀请，避免用户看到已失效邀请。
		if err := tx.WithContext(ctx).
			Model(&model.TeamInvitation{}).
			Where("team_id = ?", teamID).
			Where("status = ?", model.InvitationStatusPending).
			Updates(map[string]interface{}{
				"status": model.InvitationStatusDeclined,
			}).Error; err != nil {
			return err
		}

		// 团队解散后，回收任务分配请求，避免残留有效状态干扰审计和统计。
		if err := tx.WithContext(ctx).
			Model(&model.TaskAssignmentRequest{}).
			Where("team_id = ?", teamID).
			Where("status IN ?", []model.InvitationStatus{model.InvitationStatusPending, model.InvitationStatusAccepted}).
			Updates(map[string]interface{}{
				"status": model.InvitationStatusDeclined,
			}).Error; err != nil {
			return err
		}

		deleter, ok := teamRepoTx.(interface {
			Delete(ctx context.Context, teamID uint64) error
		})
		if !ok {
			return errors.New("team repository does not support delete")
		}
		return deleter.Delete(ctx, teamID)
	})
}

func getTeamRepoDB(repo repository.TeamRepository) (*gorm.DB, error) {
	provider, ok := repo.(interface{ GetDB() *gorm.DB })
	if !ok || provider.GetDB() == nil {
		return nil, errors.New("team repository does not expose db")
	}
	return provider.GetDB(), nil
}

func isValidTeamRole(role model.TeamRole) bool {
	switch role {
	case model.TeamRoleOwner, model.TeamRoleAdmin, model.TeamRoleMember:
		return true
	default:
		return false
	}
}
