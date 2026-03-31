package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"task-backend/internal/model"
	"task-backend/internal/repository"
	"task-backend/pkg/jwtx"
	"task-backend/pkg/password"
)

var (
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidOldPassword = errors.New("invalid old password")
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ChangePasswordRequest 用于已登录用户修改自己的密码。
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) error
	Login(ctx context.Context, req LoginRequest) (*model.LoginResponseData, error)
	ChangePassword(ctx context.Context, userID uint64, req ChangePasswordRequest) error
}

type authService struct {
	userRepo    repository.UserRepository
	jwtSecret   string
	jwtExpireHr int
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, jwtExpireHr int) AuthService {
	return &authService{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		jwtExpireHr: jwtExpireHr,
	}
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) error {
	// 先进行基础参数校验，再访问数据库。
	if !isValidUsername(req.Username) {
		return ErrInvalidUsername
	}
	if !isValidPassword(req.Password) {
		return ErrInvalidPassword
	}

	_, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil {
		return ErrUsernameExists
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("query username: %w", err)
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Username:     req.Username,
		Level:        0,
		PasswordHash: hash,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*model.LoginResponseData, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	if err := password.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Token 中携带用户身份与角色，供后续鉴权使用。
	token, err := jwtx.GenerateToken(user.ID, user.Username, user.Level, s.jwtSecret, s.jwtExpireHr)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &model.LoginResponseData{
		Token:    token,
		Username: user.Username,
		Level:    user.Level,
	}, nil
}

func (s *authService) ChangePassword(ctx context.Context, userID uint64, req ChangePasswordRequest) error {
	// 旧密码和新密码都不能为空。
	if strings.TrimSpace(req.OldPassword) == "" || strings.TrimSpace(req.NewPassword) == "" {
		return ErrInvalidPassword
	}
	if !isValidPassword(req.NewPassword) {
		return ErrInvalidPassword
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrInvalidCredentials
	}
	if err != nil {
		return fmt.Errorf("query user by id: %w", err)
	}

	// 必须先校验旧密码正确，才允许更新。
	if err := password.Compare(user.PasswordHash, req.OldPassword); err != nil {
		return ErrInvalidOldPassword
	}

	hash, err := password.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePasswordByID(ctx, userID, hash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

func isValidUsername(username string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9_]{4,20}$`)
	return re.MatchString(username)
}

func isValidPassword(password string) bool {
	if len(password) < 6 {
		return false
	}
	hasLetter := false
	hasDigit := false
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}
