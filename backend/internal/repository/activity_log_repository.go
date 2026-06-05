package repository

import (
	"context"

	"gorm.io/gorm"

	"task-backend/internal/model"
)

type ActivityLogRepository interface {
	Create(ctx context.Context, log *model.ActivityLog) error
	ListByTaskID(ctx context.Context, taskID uint64, page, size int) ([]model.ActivityLogListItem, int64, error)
}

type activityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) ActivityLogRepository {
	return &activityLogRepository{db: db}
}

func (r *activityLogRepository) Create(ctx context.Context, log *model.ActivityLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *activityLogRepository) ListByTaskID(ctx context.Context, taskID uint64, page, size int) ([]model.ActivityLogListItem, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size

	query := r.db.WithContext(ctx).
		Model(&model.ActivityLog{}).
		Where("target_type = ?", model.ActivityTargetTask).
		Where("target_id = ?", taskID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	list := make([]model.ActivityLogListItem, 0)
	if err := query.
		Select("id, actor_name, action_type, content, created_at").
		Order("created_at DESC").
		Order("id DESC").
		Limit(size).
		Offset(offset).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
