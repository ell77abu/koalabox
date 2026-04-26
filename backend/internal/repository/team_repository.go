package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"task-backend/internal/model"
)

type TeamRepository interface {
	Create(ctx context.Context, team *model.Team) error
	ListByUserID(ctx context.Context, userID uint64, search string) ([]model.TeamWithRole, error)
	GetByID(ctx context.Context, teamID uint64) (*model.Team, error)
	AddMember(ctx context.Context, member *model.TeamMember) error
	RemoveMember(ctx context.Context, teamID, userID uint64) error
	UpdateMemberRole(ctx context.Context, teamID, userID uint64, role model.TeamRole) error
	IsMember(ctx context.Context, teamID, userID uint64) (bool, error)
	GetMemberRole(ctx context.Context, teamID, userID uint64) (model.TeamRole, error)
	ListMembers(ctx context.Context, teamID uint64) ([]model.TeamMemberDetail, error)
}

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) Create(ctx context.Context, team *model.Team) error {
	return r.db.WithContext(ctx).Create(team).Error
}

func (r *teamRepository) ListByUserID(ctx context.Context, userID uint64, search string) ([]model.TeamWithRole, error) {
	trimmedSearch := strings.TrimSpace(search)

	query := r.db.WithContext(ctx).
		Table("teams").
		Select("teams.id AS team_id, teams.name, tm.role").
		Joins("JOIN team_members tm ON tm.team_id = teams.id AND tm.deleted_at IS NULL").
		Where("teams.deleted_at IS NULL").
		Where("tm.user_id = ?", userID)

	if trimmedSearch != "" {
		likePattern := "%" + trimmedSearch + "%"
		query = query.Where("teams.name ILIKE ?", likePattern)
	}

	var list []model.TeamWithRole
	err := query.
		Order("teams.created_at DESC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *teamRepository) GetByID(ctx context.Context, teamID uint64) (*model.Team, error) {
	var team model.Team
	err := r.db.WithContext(ctx).
		Where("id = ?", teamID).
		Where("deleted_at IS NULL").
		First(&team).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) AddMember(ctx context.Context, member *model.TeamMember) error {
	err := r.db.WithContext(ctx).Create(member).Error
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return err
}

func (r *teamRepository) RemoveMember(ctx context.Context, teamID, userID uint64) error {
	res := r.db.WithContext(ctx).
		Unscoped().
		Where("team_id = ?", teamID).
		Where("user_id = ?", userID).
		Delete(&model.TeamMember{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *teamRepository) UpdateMemberRole(ctx context.Context, teamID, userID uint64, role model.TeamRole) error {
	res := r.db.WithContext(ctx).
		Model(&model.TeamMember{}).
		Where("team_id = ?", teamID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Updates(map[string]interface{}{"role": role})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *teamRepository) IsMember(ctx context.Context, teamID, userID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.TeamMember{}).
		Where("team_id = ?", teamID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *teamRepository) GetMemberRole(ctx context.Context, teamID, userID uint64) (model.TeamRole, error) {
	var member model.TeamMember
	err := r.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return member.Role, nil
}

func (r *teamRepository) ListMembers(ctx context.Context, teamID uint64) ([]model.TeamMemberDetail, error) {
	var list []model.TeamMemberDetail
	err := r.db.WithContext(ctx).
		Table("team_members tm").
		Select("tm.user_id, u.username, tm.role").
		Joins("JOIN users u ON u.id = tm.user_id AND u.deleted_at IS NULL").
		Where("tm.team_id = ?", teamID).
		Where("tm.deleted_at IS NULL").
		Order("tm.created_at ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]model.TeamMemberDetail, 0)
	}
	return list, nil
}
