package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"task-backend/internal/model"
)

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	ListByUserID(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error)
	ListAll(ctx context.Context, page, size int) ([]model.Task, int64, error)
	GetByIDAndUserID(ctx context.Context, id, userID uint64) (*model.Task, error)
	GetByID(ctx context.Context, id uint64) (*model.Task, error)
	UpdateByIDAndUserID(ctx context.Context, task *model.Task) error
	UpdateByID(ctx context.Context, task *model.Task) error
	DeleteByIDAndUserID(ctx context.Context, id, userID uint64) error
	DeleteByID(ctx context.Context, id uint64) error
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) ListByUserID(ctx context.Context, userID uint64, page, size int) ([]model.Task, int64, error) {
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
		Model(&model.Task{}).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []model.Task
	err := query.
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (r *taskRepository) ListAll(ctx context.Context, page, size int) ([]model.Task, int64, error) {
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
		Model(&model.Task{}).
		Where("deleted_at IS NULL")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []model.Task
	err := query.
		Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (r *taskRepository) GetByIDAndUserID(ctx context.Context, id, userID uint64) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) GetByID(ctx context.Context, id uint64) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) UpdateByIDAndUserID(ctx context.Context, task *model.Task) error {
	updates := map[string]interface{}{
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
		"priority":    task.Priority,
		"deadline":    task.Deadline,
		"updated_at":  time.Now(),
	}

	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", task.ID).
		Where("user_id = ?", task.UserID).
		Where("deleted_at IS NULL").
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskRepository) UpdateByID(ctx context.Context, task *model.Task) error {
	updates := map[string]interface{}{
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
		"priority":    task.Priority,
		"deadline":    task.Deadline,
		"updated_at":  time.Now(),
	}

	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", task.ID).
		Where("deleted_at IS NULL").
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskRepository) DeleteByIDAndUserID(ctx context.Context, id, userID uint64) error {
	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", id).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *taskRepository) DeleteByID(ctx context.Context, id uint64) error {
	res := r.db.WithContext(ctx).
		Model(&model.Task{}).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
