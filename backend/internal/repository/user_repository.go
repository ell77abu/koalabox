package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"task-backend/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id uint64) (*model.User, error)
	UpdatePasswordByID(ctx context.Context, id uint64, passwordHash string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	// 执行创建操作
	result := r.db.WithContext(ctx).Create(user)
	
	// 调试日志：如果写入失败，这里会打印原因（如唯一键冲突、表不存在等）
	if result.Error != nil {
		fmt.Printf("[GORM ERROR] Create User Failed: %v\n", result.Error)
		return result.Error
	}
	
	// 调试日志：确认受影响行数和生成的 ID
	fmt.Printf("[GORM INFO] User Created Successfully! ID: %d, RowsAffected: %d\n", user.ID, result.RowsAffected)
	return nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 这里直接使用同包下 errors.go 定义好的 ErrNotFound
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdatePasswordByID(ctx context.Context, id uint64, passwordHash string) error {
	res := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Updates(map[string]interface{}{
			"password_hash": passwordHash,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
