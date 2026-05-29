package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"task-backend/internal/model"
)

type TeamInvitationRepository interface {
	Create(ctx context.Context, invite *model.TeamInvitation) error
	GetByID(ctx context.Context, id uint64) (*model.TeamInvitation, error)
	GetByIDForInvitee(ctx context.Context, id, inviteeID uint64) (*model.TeamInvitation, error)
	GetPendingByTeamAndInvitee(ctx context.Context, teamID, inviteeID uint64) (*model.TeamInvitation, error)
	ListPendingByInvitee(ctx context.Context, inviteeID uint64) ([]model.TeamInvitationDetail, error)
	UpdateStatus(ctx context.Context, id uint64, status model.InvitationStatus) error
}

type teamInvitationRepository struct {
	db *gorm.DB
}

func NewTeamInvitationRepository(db *gorm.DB) TeamInvitationRepository {
	return &teamInvitationRepository{db: db}
}

func (r *teamInvitationRepository) Create(ctx context.Context, invite *model.TeamInvitation) error {
	err := r.db.WithContext(ctx).Create(invite).Error
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return err
}

func (r *teamInvitationRepository) GetByID(ctx context.Context, id uint64) (*model.TeamInvitation, error) {
	var invite model.TeamInvitation
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&invite).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *teamInvitationRepository) GetByIDForInvitee(ctx context.Context, id, inviteeID uint64) (*model.TeamInvitation, error) {
	var invite model.TeamInvitation
	err := r.db.WithContext(ctx).
		Table("team_invitations ti").
		Select("ti.*").
		Joins("JOIN teams t ON t.id = ti.team_id AND t.deleted_at IS NULL").
		Where("ti.id = ?", id).
		Where("ti.invitee_id = ?", inviteeID).
		First(&invite).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *teamInvitationRepository) GetPendingByTeamAndInvitee(ctx context.Context, teamID, inviteeID uint64) (*model.TeamInvitation, error) {
	var invite model.TeamInvitation
	err := r.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Where("invitee_id = ?", inviteeID).
		Where("status = ?", model.InvitationStatusPending).
		First(&invite).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *teamInvitationRepository) ListPendingByInvitee(ctx context.Context, inviteeID uint64) ([]model.TeamInvitationDetail, error) {
	var list []model.TeamInvitationDetail
	err := r.db.WithContext(ctx).
		Table("team_invitations ti").
		Select("ti.id, ti.team_id, t.name AS team_name, ti.inviter_id, u.username AS inviter_username, ti.status, ti.created_at").
		Joins("JOIN teams t ON t.id = ti.team_id AND t.deleted_at IS NULL").
		Joins("JOIN users u ON u.id = ti.inviter_id AND u.deleted_at IS NULL").
		Where("ti.invitee_id = ?", inviteeID).
		Where("ti.status = ?", model.InvitationStatusPending).
		Order("ti.created_at DESC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]model.TeamInvitationDetail, 0)
	}
	return list, nil
}

func (r *teamInvitationRepository) UpdateStatus(ctx context.Context, id uint64, status model.InvitationStatus) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var invite model.TeamInvitation
		err := tx.Select("id", "team_id", "invitee_id").
			Where("id = ?", id).
			First(&invite).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}

		if status != model.InvitationStatusPending {
			err = tx.
				Where("team_id = ?", invite.TeamID).
				Where("invitee_id = ?", invite.InviteeID).
				Where("status = ?", status).
				Where("id <> ?", invite.ID).
				Delete(&model.TeamInvitation{}).Error
			if err != nil {
				return err
			}
		}

		res := tx.Model(&model.TeamInvitation{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{"status": status})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}
