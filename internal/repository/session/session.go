package session

import (
	"context"
	"errors"
	"time"

	"dooz/entity"
	appErrors "dooz/internal/errors"
	"dooz/internal/repository/tx"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNotFound     = appErrors.ErrNotFound
	ErrAlreadyExist = appErrors.NewAppError("SESSION_ALREADY_EXISTS", "Session already exists", 409)
)

type Repository interface {
	Create(ctx context.Context, session *entity.Session) error
	GetByUserAndDevice(ctx context.Context, userID string, deviceType entity.DeviceType) (*entity.Session, error)
	GetByJwtID(ctx context.Context, jwtID string) (*entity.Session, error)
	Update(ctx context.Context, session *entity.Session) error
	Delete(ctx context.Context, id string) error
	DeleteByUserAndDevice(ctx context.Context, userID string, deviceType entity.DeviceType) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
	UpdateLastActivity(ctx context.Context, sessionID string) error
}

type sessionRepository struct {
	t tx.Transaction
}

func New(t tx.Transaction) Repository {
	return &sessionRepository{t: t}
}

func (r *sessionRepository) Create(ctx context.Context, session *entity.Session) error {
	result := r.t.DB(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "device_type"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"refresh_token_hash", "jwt_id", "expires_at", "ip_address", "last_activity_at",
			}),
		}).
		Create(session)
	return result.Error
}

func (r *sessionRepository) GetByUserAndDevice(ctx context.Context, userID string, deviceType entity.DeviceType) (*entity.Session, error) {
	var session entity.Session
	result := r.t.DB(ctx).Where("user_id = ? AND device_type = ?", userID, deviceType).First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &session, nil
}

func (r *sessionRepository) GetByJwtID(ctx context.Context, jwtID string) (*entity.Session, error) {
	var session entity.Session
	result := r.t.DB(ctx).Where("jwt_id = ?", jwtID).First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &session, nil
}

func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	return r.t.DB(ctx).Delete(&entity.Session{}, "id = ?", id).Error
}

func (r *sessionRepository) DeleteByUserAndDevice(ctx context.Context, userID string, deviceType entity.DeviceType) error {
	return r.t.DB(ctx).Delete(&entity.Session{}, "user_id = ? AND device_type = ?", userID, deviceType).Error
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.t.DB(ctx).Delete(&entity.Session{}, "user_id = ?", userID).Error
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	return r.t.DB(ctx).Delete(&entity.Session{}, "expires_at < ?", time.Now().Unix()).Error
}

func (r *sessionRepository) Update(ctx context.Context, session *entity.Session) error {
	result := r.t.DB(ctx).Model(session).Where("id = ?", session.ID).Updates(session)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sessionRepository) UpdateLastActivity(ctx context.Context, sessionID string) error {
	return r.t.DB(ctx).Model(&entity.Session{}).
		Where("id = ?", sessionID).
		Update("last_activity_at", time.Now().Unix()).Error
}
