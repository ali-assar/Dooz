package otp

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"dooz/entity"
	appErrors "dooz/internal/errors"
	"dooz/internal/repository/tx"

	"gorm.io/gorm"
)

var (
	ErrNotFound     = appErrors.ErrNotFound
	ErrAlreadyExist = appErrors.NewAppError("OTP_ALREADY_EXISTS", "OTP already exists", 409)
	ErrExpired      = appErrors.NewAppError("OTP_EXPIRED", "OTP expired", 401)
	ErrMaxAttempts  = appErrors.NewAppError("OTP_MAX_ATTEMPTS", "Maximum attempts exceeded", 429)
)

type Repository interface {
	Create(ctx context.Context, otp *entity.OTP) error
	GetLatestByRecipient(ctx context.Context, recipient string, channel entity.OTPChannel, purpose entity.OTPPurpose) (*entity.OTP, error)
	GetPendingOTPs(ctx context.Context, limit int) ([]*entity.OTP, error)
	IncrementRetryCount(ctx context.Context, otpID string) (int, error)
	UpdateRetryInfo(ctx context.Context, otpID string, retryCount int, nextRetryAt int64) error
	SetOTPStatus(ctx context.Context, otpID string, status entity.OTPStatus) error
	VerifyAndDelete(ctx context.Context, recipient string, channel entity.OTPChannel, code string, purpose entity.OTPPurpose) error
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
}

type otpRepository struct {
	t tx.Transaction
}

func New(t tx.Transaction) Repository {
	return &otpRepository{t: t}
}

func (r *otpRepository) Create(ctx context.Context, otp *entity.OTP) error {
	result := r.t.DB(ctx).Create(otp)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrAlreadyExist
		}
		return result.Error
	}
	return nil
}

func (r *otpRepository) GetLatestByRecipient(ctx context.Context, recipient string, channel entity.OTPChannel, purpose entity.OTPPurpose) (*entity.OTP, error) {
	var otp entity.OTP
	result := r.t.DB(ctx).
		Where("recipient = ? AND channel = ? AND purpose = ?", recipient, channel, purpose).
		Order("id DESC").First(&otp)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &otp, nil
}

func (r *otpRepository) GetPendingOTPs(ctx context.Context, limit int) ([]*entity.OTP, error) {
	now := time.Now().Unix()
	var otps []*entity.OTP
	result := r.t.DB(ctx).
		Where("status = ? AND (next_retry_at IS NULL OR next_retry_at <= ?)", entity.OTPStatusPending, now).
		Order("id ASC").Limit(limit).Find(&otps)
	return otps, result.Error
}

func (r *otpRepository) SetOTPStatus(ctx context.Context, otpID string, status entity.OTPStatus) error {
	now := time.Now().Unix()
	result := r.t.DB(ctx).Model(&entity.OTP{}).Where("id = ?", otpID).
		Updates(map[string]interface{}{"status": status, "processed_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *otpRepository) IncrementRetryCount(ctx context.Context, otpID string) (int, error) {
	var retryCount int
	result := r.t.DB(ctx).
		Raw("UPDATE otp_outbox SET retry_count = retry_count + 1 WHERE id = ? RETURNING retry_count", otpID).
		Scan(&retryCount)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, ErrNotFound
	}
	return retryCount, nil
}

func (r *otpRepository) UpdateRetryInfo(ctx context.Context, otpID string, retryCount int, nextRetryAt int64) error {
	result := r.t.DB(ctx).Model(&entity.OTP{}).Where("id = ?", otpID).
		Updates(map[string]interface{}{"retry_count": retryCount, "next_retry_at": nextRetryAt})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *otpRepository) VerifyAndDelete(ctx context.Context, recipient string, channel entity.OTPChannel, code string, purpose entity.OTPPurpose) error {
	var otp entity.OTP
	result := r.t.DB(ctx).
		Where("recipient = ? AND channel = ? AND purpose = ? AND status = ?", recipient, channel, purpose, entity.OTPStatusSent).
		Order("id DESC").First(&otp)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return result.Error
	}
	if time.Now().Unix() > otp.ExpiresAt {
		return ErrExpired
	}
	if subtle.ConstantTimeCompare([]byte(otp.Code), []byte(code)) != 1 {
		return ErrNotFound
	}
	return r.Delete(ctx, otp.ID)
}

func (r *otpRepository) Delete(ctx context.Context, id string) error {
	result := r.t.DB(ctx).Delete(&entity.OTP{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *otpRepository) DeleteExpired(ctx context.Context) error {
	return r.t.DB(ctx).Delete(&entity.OTP{}, "expires_at < ?", time.Now().Unix()).Error
}
