package game_challenge

import (
	"context"
	"errors"

	"dooz/entity"
	appErrors "dooz/internal/errors"
	"dooz/internal/repository/tx"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	ErrNotFound     = appErrors.ErrNotFound
	ErrActiveExists = appErrors.NewAppError("CHALLENGE_EXISTS", "An active challenge already exists", 409)
)

type Repository interface {
	Create(ctx context.Context, challenge *entity.GameChallenge) error
	GetByID(ctx context.Context, id string) (*entity.GameChallenge, error)
	Update(ctx context.Context, challenge *entity.GameChallenge) error
	GetPendingForAddressee(ctx context.Context, addresseeID string, now int64) ([]*entity.GameChallenge, error)
	GetActivePair(ctx context.Context, userA, userB string) (*entity.GameChallenge, error)
}

type gameChallengeRepository struct {
	t tx.Transaction
}

func New(t tx.Transaction) Repository {
	return &gameChallengeRepository{t: t}
}

func (r *gameChallengeRepository) Create(ctx context.Context, challenge *entity.GameChallenge) error {
	err := r.t.DB(ctx).Create(challenge).Error
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.ConstraintName == "uq_active_challenge_pair" {
		return ErrActiveExists
	}
	return err
}

func (r *gameChallengeRepository) GetByID(ctx context.Context, id string) (*entity.GameChallenge, error) {
	var challenge entity.GameChallenge
	result := r.t.DB(ctx).Where("id = ?", id).First(&challenge)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &challenge, nil
}

func (r *gameChallengeRepository) Update(ctx context.Context, challenge *entity.GameChallenge) error {
	result := r.t.DB(ctx).Model(challenge).Where("id = ?", challenge.ID).Updates(challenge)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *gameChallengeRepository) GetPendingForAddressee(ctx context.Context, addresseeID string, now int64) ([]*entity.GameChallenge, error) {
	var challenges []*entity.GameChallenge
	result := r.t.DB(ctx).
		Where("addressee_id = ? AND status = ? AND expires_at > ?", addresseeID, entity.GameChallengePending, now).
		Order("id DESC").
		Find(&challenges)
	return challenges, result.Error
}

func (r *gameChallengeRepository) GetActivePair(ctx context.Context, userA, userB string) (*entity.GameChallenge, error) {
	var challenge entity.GameChallenge
	result := r.t.DB(ctx).
		Where("((requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)) AND status = ?",
			userA, userB, userB, userA, entity.GameChallengePending).
		First(&challenge)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &challenge, nil
}
