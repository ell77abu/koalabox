package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"task-backend/internal/model"
)

type TaskAssignmentRepository interface {
	Create(ctx context.Context, req *model.TaskAssignmentRequest) error
	GetByID(ctx context.Context, id uint64) (*model.TaskAssignmentRequest, error)
	GetByIDForAssigneeScope(ctx context.Context, id, assigneeID uint64) (*model.TaskAssignmentRequest, error)
	GetPendingByTaskAndAssignee(ctx context.Context, taskID, assigneeID uint64) (*model.TaskAssignmentRequest, error)
	ListPendingByAssignee(ctx context.Context, assigneeID uint64) ([]model.TaskAssignmentRequestDetail, error)
	UpdateStatus(ctx context.Context, id uint64, status model.InvitationStatus) error
	UpdatePendingStatusForAssigneeScope(ctx context.Context, id, assigneeID uint64, status model.InvitationStatus) error
	AcceptPendingForAssigneeScope(ctx context.Context, id, assigneeID uint64) error
	DeclinePendingByTaskID(ctx context.Context, taskID uint64) error
}

type taskAssignmentRepository struct {
	db *gorm.DB
}

func NewTaskAssignmentRepository(db *gorm.DB) TaskAssignmentRepository {
	return &taskAssignmentRepository{db: db}
}

func (r *taskAssignmentRepository) Create(ctx context.Context, req *model.TaskAssignmentRequest) error {
	err := r.db.WithContext(ctx).Create(req).Error
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return err
}

func (r *taskAssignmentRepository) GetByID(ctx context.Context, id uint64) (*model.TaskAssignmentRequest, error) {
	var req model.TaskAssignmentRequest
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&req).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *taskAssignmentRepository) GetByIDForAssigneeScope(ctx context.Context, id, assigneeID uint64) (*model.TaskAssignmentRequest, error) {
	var req model.TaskAssignmentRequest
	err := r.db.WithContext(ctx).
		Table("task_assignment_requests tar").
		Select("tar.*").
		Joins("JOIN tasks t ON t.id = tar.task_id AND t.deleted_at IS NULL").
		Joins("JOIN teams tm ON tm.id = tar.team_id AND tm.deleted_at IS NULL").
		Joins("JOIN team_members mem ON mem.team_id = tar.team_id AND mem.user_id = tar.assignee_id AND mem.deleted_at IS NULL").
		Where("tar.id = ?", id).
		Where("tar.assignee_id = ?", assigneeID).
		First(&req).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *taskAssignmentRepository) GetPendingByTaskAndAssignee(ctx context.Context, taskID, assigneeID uint64) (*model.TaskAssignmentRequest, error) {
	var req model.TaskAssignmentRequest
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Where("assignee_id = ?", assigneeID).
		Where("status = ?", model.InvitationStatusPending).
		First(&req).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *taskAssignmentRepository) ListPendingByAssignee(ctx context.Context, assigneeID uint64) ([]model.TaskAssignmentRequestDetail, error) {
	var list []model.TaskAssignmentRequestDetail
	err := r.db.WithContext(ctx).
		Table("task_assignment_requests tar").
		Select("tar.id, tar.task_id, t.title AS task_title, tar.team_id, tm.name AS team_name, tar.requester_id, u.username AS requester_username, tar.status, tar.created_at").
		Joins("JOIN tasks t ON t.id = tar.task_id AND t.deleted_at IS NULL").
		Joins("JOIN teams tm ON tm.id = tar.team_id AND tm.deleted_at IS NULL").
		Joins("JOIN team_members mem ON mem.team_id = tar.team_id AND mem.user_id = tar.assignee_id AND mem.deleted_at IS NULL").
		Joins("JOIN users u ON u.id = tar.requester_id AND u.deleted_at IS NULL").
		Where("tar.assignee_id = ?", assigneeID).
		Where("tar.status IN ?", []model.InvitationStatus{model.InvitationStatusPending, model.InvitationStatusAccepted}).
		Order("tar.created_at DESC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]model.TaskAssignmentRequestDetail, 0)
	}
	return list, nil
}

func (r *taskAssignmentRepository) UpdateStatus(ctx context.Context, id uint64, status model.InvitationStatus) error {
	res := r.db.WithContext(ctx).
		Model(&model.TaskAssignmentRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": status})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskAssignmentRepository) UpdatePendingStatusForAssigneeScope(ctx context.Context, id, assigneeID uint64, status model.InvitationStatus) error {
	res := r.db.WithContext(ctx).
		Model(&model.TaskAssignmentRequest{}).
		Where("id = ?", id).
		Where("assignee_id = ?", assigneeID).
		Where("status = ?", model.InvitationStatusPending).
		Where(`EXISTS (
			SELECT 1
			FROM teams t
			JOIN team_members tm ON tm.team_id = t.id AND tm.deleted_at IS NULL
			WHERE t.id = task_assignment_requests.team_id
				AND t.deleted_at IS NULL
				AND tm.user_id = task_assignment_requests.assignee_id
		)`).
		Updates(map[string]interface{}{"status": status})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskAssignmentRepository) AcceptPendingForAssigneeScope(ctx context.Context, id, assigneeID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var req model.TaskAssignmentRequest
		err := tx.
			Table("task_assignment_requests tar").
			Select("tar.*").
			Joins("JOIN tasks task_scope ON task_scope.id = tar.task_id AND task_scope.deleted_at IS NULL").
			Joins("JOIN teams team_scope ON team_scope.id = tar.team_id AND team_scope.deleted_at IS NULL").
			Joins("JOIN team_members mem ON mem.team_id = tar.team_id AND mem.user_id = tar.assignee_id AND mem.deleted_at IS NULL").
			Where("tar.id = ?", id).
			Where("tar.assignee_id = ?", assigneeID).
			Where("tar.status = ?", model.InvitationStatusPending).
			First(&req).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}

		taskRes := tx.Model(&model.Task{}).
			Where("id = ?", req.TaskID).
			Where("team_id = ?", req.TeamID).
			Where("deleted_at IS NULL").
			Where(`EXISTS (
				SELECT 1
				FROM teams t
				JOIN team_members tm ON tm.team_id = t.id AND tm.deleted_at IS NULL
				WHERE t.id = tasks.team_id
					AND t.deleted_at IS NULL
					AND tm.user_id = ?
			)`, assigneeID).
			Updates(map[string]interface{}{"assignee_id": assigneeID})
		if taskRes.Error != nil {
			return taskRes.Error
		}
		if taskRes.RowsAffected == 0 {
			return ErrNotFound
		}

		reqRes := tx.Model(&model.TaskAssignmentRequest{}).
			Where("id = ?", id).
			Where("assignee_id = ?", assigneeID).
			Where("status = ?", model.InvitationStatusPending).
			Updates(map[string]interface{}{"status": model.InvitationStatusAccepted})
		if reqRes.Error != nil {
			return reqRes.Error
		}
		if reqRes.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (r *taskAssignmentRepository) DeclinePendingByTaskID(ctx context.Context, taskID uint64) error {
	res := r.db.WithContext(ctx).
		Model(&model.TaskAssignmentRequest{}).
		Where("task_id = ?", taskID).
		Where("status = ?", model.InvitationStatusPending).
		Updates(map[string]interface{}{"status": model.InvitationStatusDeclined})
	return res.Error
}
