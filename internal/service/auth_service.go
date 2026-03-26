package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
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
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) error
	Login(ctx context.Context, req LoginRequest) (*model.LoginResponseData, error)
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

	token, err := jwtx.GenerateToken(user.ID, user.Username, s.jwtSecret, s.jwtExpireHr)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &model.LoginResponseData{
		Token:    token,
		Username: user.Username,
	}, nil
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
