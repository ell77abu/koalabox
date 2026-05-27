package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-backend/internal/model"
	"task-backend/internal/repository"
)

type authUserRepoMock struct {
	createFn           func(ctx context.Context, user *model.User) error
	getByUsernameFn    func(ctx context.Context, username string) (*model.User, error)
	getByIDFn          func(ctx context.Context, id uint64) (*model.User, error)
	updatePasswordByID func(ctx context.Context, id uint64, passwordHash string) error
}

func (m *authUserRepoMock) Create(ctx context.Context, user *model.User) error {
	if m.createFn == nil {
		panic("unexpected call to Create")
	}
	return m.createFn(ctx, user)
}

func (m *authUserRepoMock) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	if m.getByUsernameFn == nil {
		panic("unexpected call to GetByUsername")
	}
	return m.getByUsernameFn(ctx, username)
}

func (m *authUserRepoMock) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	if m.getByIDFn == nil {
		panic("unexpected call to GetByID")
	}
	return m.getByIDFn(ctx, id)
}

func (m *authUserRepoMock) UpdatePasswordByID(ctx context.Context, id uint64, passwordHash string) error {
	if m.updatePasswordByID == nil {
		panic("unexpected call to UpdatePasswordByID")
	}
	return m.updatePasswordByID(ctx, id, passwordHash)
}

func TestAuthServiceLoginReturnsInvalidCredentialsWhenUserNotFound(t *testing.T) {
	svc := NewAuthService(&authUserRepoMock{
		getByUsernameFn: func(ctx context.Context, username string) (*model.User, error) {
			assert.Equal(t, "deleted_user", username)
			// 仓储层应将“已删除用户不可见”与“用户不存在”统一映射为 not found。
			return nil, repository.ErrNotFound
		},
	}, "test-secret", 24)

	result, err := svc.Login(context.Background(), LoginRequest{
		Username: "deleted_user",
		Password: "abc12345",
	})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthServiceChangePasswordRejectsMissingUser(t *testing.T) {
	svc := NewAuthService(&authUserRepoMock{
		getByIDFn: func(ctx context.Context, id uint64) (*model.User, error) {
			assert.Equal(t, uint64(42), id)
			return nil, repository.ErrNotFound
		},
	}, "test-secret", 24)

	err := svc.ChangePassword(context.Background(), 42, ChangePasswordRequest{
		OldPassword: "abc12345",
		NewPassword: "newpass123",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}
