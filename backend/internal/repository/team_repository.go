package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"task-backend/internal/model"
)

type TeamRepository interface {
	Create(ctx context.Context, team *model.Team) error
	ListByUserID(ctx context.Context, userID uint64, search string) ([]model.TeamWithRole, error)
	ListByUserIDPage(ctx context.Context, userID uint64, search string, page, size int) ([]model.TeamWithRole, int64, error)
	GetByID(ctx context.Context, teamID uint64) (*model.Team, error)
	GetByIDForMemberScope(ctx context.Context, teamID, userID uint64) (*model.Team, error)
	GetByName(ctx context.Context, name string) (*model.Team, error)
	UpdateName(ctx context.Context, teamID uint64, name string) error
	AddMember(ctx context.Context, member *model.TeamMember) error
	RemoveMember(ctx context.Context, teamID, userID uint64) error
	UpdateMemberRole(ctx context.Context, teamID, userID uint64, role model.TeamRole) error
	IsMember(ctx context.Context, teamID, userID uint64) (bool, error)
	GetMemberRole(ctx context.Context, teamID, userID uint64) (model.TeamRole, error)
	ListMembers(ctx context.Context, teamID uint64) ([]model.TeamMemberDetail, error)
	TransferOwnership(ctx context.Context, teamID, oldOwnerID, newOwnerID uint64) error
	// IsDeleted reports whether a team record exists and is soft-deleted.
	IsDeleted(ctx context.Context, teamID uint64) (bool, error)
}

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *teamRepository) Create(ctx context.Context, team *model.Team) error {
	err := r.db.WithContext(ctx).Create(team).Error
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return err
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

func (r *teamRepository) ListByUserIDPage(ctx context.Context, userID uint64, search string, page, size int) ([]model.TeamWithRole, int64, error) {
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
		Table("teams").
		Select("teams.id AS team_id, teams.name, tm.role").
		Joins("JOIN team_members tm ON tm.team_id = teams.id AND tm.deleted_at IS NULL").
		Where("teams.deleted_at IS NULL").
		Where("tm.user_id = ?", userID)

	if trimmedSearch := strings.TrimSpace(search); trimmedSearch != "" {
		query = query.Where("teams.name ILIKE ?", "%"+trimmedSearch+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	list := make([]model.TeamWithRole, 0)
	err := query.
		Order("teams.created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
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

func (r *teamRepository) GetByIDForMemberScope(ctx context.Context, teamID, userID uint64) (*model.Team, error) {
	var team model.Team
	err := r.db.WithContext(ctx).
		Table("teams t").
		Select("t.*").
		Joins("JOIN team_members tm ON tm.team_id = t.id AND tm.deleted_at IS NULL").
		Where("t.id = ?", teamID).
		Where("tm.user_id = ?", userID).
		Where("t.deleted_at IS NULL").
		First(&team).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) GetByName(ctx context.Context, name string) (*model.Team, error) {
	var team model.Team
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
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

func (r *teamRepository) UpdateName(ctx context.Context, teamID uint64, name string) error {
	res := r.db.WithContext(ctx).
		Model(&model.Team{}).
		Where("id = ?", teamID).
		Where("deleted_at IS NULL").
		Updates(map[string]interface{}{"name": name})
	if res.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(res.Error, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
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
		Table("team_members tm").
		Joins("JOIN teams t ON t.id = tm.team_id AND t.deleted_at IS NULL").
		Where("tm.team_id = ?", teamID).
		Where("tm.user_id = ?", userID).
		Where("tm.deleted_at IS NULL").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *teamRepository) GetMemberRole(ctx context.Context, teamID, userID uint64) (model.TeamRole, error) {
	var member model.TeamMember
	err := r.db.WithContext(ctx).
		Table("team_members tm").
		Select("tm.*").
		Joins("JOIN teams t ON t.id = tm.team_id AND t.deleted_at IS NULL").
		Where("tm.team_id = ?", teamID).
		Where("tm.user_id = ?", userID).
		Where("tm.deleted_at IS NULL").
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

func (r *teamRepository) TransferOwnership(ctx context.Context, teamID, oldOwnerID, newOwnerID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Team{}).
			Where("id = ?", teamID).
			Where("deleted_at IS NULL").
			Updates(map[string]interface{}{
				"owner_id":   newOwnerID,
				"updated_at": time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}

		res = tx.Model(&model.TeamMember{}).
			Where("team_id = ?", teamID).
			Where("user_id = ?", newOwnerID).
			Where("deleted_at IS NULL").
			Updates(map[string]interface{}{
				"role":       model.TeamRoleOwner,
				"updated_at": time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}

		res = tx.Model(&model.TeamMember{}).
			Where("team_id = ?", teamID).
			Where("user_id = ?", oldOwnerID).
			Where("deleted_at IS NULL").
			Updates(map[string]interface{}{
				"role":       model.TeamRoleAdmin,
				"updated_at": time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}

		return nil
	})
}

func (r *teamRepository) Delete(ctx context.Context, teamID uint64) error {
	res := r.db.WithContext(ctx).
		Model(&model.Team{}).
		Where("id = ?", teamID).
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

func (r *teamRepository) SoftDeleteMembersByTeamID(ctx context.Context, teamID uint64) error {
	res := r.db.WithContext(ctx).
		Model(&model.TeamMember{}).
		Where("team_id = ?", teamID).
		Where("deleted_at IS NULL").
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *teamRepository) IsDeleted(ctx context.Context, teamID uint64) (bool, error) {
	var deletedAt *time.Time
	err := r.db.WithContext(ctx).
		Model(&model.Team{}).
		Select("deleted_at").
		Where("id = ?", teamID).
		Scan(&deletedAt).Error
	if err != nil {
		return false, err
	}
	return deletedAt != nil, nil
}
