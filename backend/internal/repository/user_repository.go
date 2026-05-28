package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"task-backend/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id uint64) (*model.User, error)
	// UpdatePasswordByID 仅更新目标用户的密码哈希。
	UpdatePasswordByID(ctx context.Context, id uint64, passwordHash string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	// 创建用户记录，并保留调试日志便于部署排查。
	result := r.db.WithContext(ctx).Create(user)

	// 写入失败时输出原因，例如唯一键冲突或表结构不一致。
	if result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		slog.Error("create_user_failed", slog.String("error", result.Error.Error()))
		return result.Error
	}
	slog.Info("user_created", slog.Uint64("user_id", user.ID), slog.Int64("rows_affected", result.RowsAffected))
	return nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		Where("deleted_at IS NULL").
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
		Where("deleted_at IS NULL").
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
