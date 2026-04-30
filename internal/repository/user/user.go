package user

import (
	"context"
	"errors"

	"dooz/entity"
	appErrors "dooz/internal/errors"
	"dooz/internal/repository/tx"

	"gorm.io/gorm"
)

var (
	ErrNotFound        = appErrors.ErrNotFound
	ErrAlreadyExist    = appErrors.NewAppError("USER_ALREADY_EXISTS", "User already exists", 409)
	ErrDuplicatePhone  = appErrors.NewAppError("DUPLICATE_PHONE", "Phone number already registered", 409)
	ErrDuplicateEmail  = appErrors.NewAppError("DUPLICATE_EMAIL", "Email already registered", 409)
	ErrInvalidPassword = appErrors.NewAppError("INVALID_PASSWORD", "Invalid password", 401)
	ErrInvalidRole     = appErrors.NewAppError("INVALID_ROLE", "Invalid role", 400)
)

type Repository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByUserCode(ctx context.Context, userCode int) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
	GetByPhone(ctx context.Context, phone string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetAllWithCursor(ctx context.Context, cursorID *string, sortDesc bool, limit uint32, filters map[string]interface{}) ([]*entity.User, error)
	MarkPhoneVerified(ctx context.Context, id string) error
	MarkEmailVerified(ctx context.Context, id string) error
	UpdateStats(ctx context.Context, userID string, wins, losses, draws, xCount, oCount, coins, gems int) error
	SetOnline(ctx context.Context, userID string, online bool) error
	UpdateCurrentStyle(ctx context.Context, userID string, theme, xoShape, avatar *int) error
	AddBalance(ctx context.Context, userID string, coins, gems int) error
	DeductBalance(ctx context.Context, userID string, currency byte, amount int) error
}

type userRepository struct {
	t tx.Transaction
}

func New(t tx.Transaction) Repository {
	return &userRepository{t: t}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	result := r.t.DB(ctx).Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			if user.Phone != "" {
				return ErrDuplicatePhone
			}
			return ErrDuplicateEmail
		}
		return result.Error
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	result := r.t.DB(ctx).Where("id = ? AND deleted_at = 0", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) GetByUserCode(ctx context.Context, userCode int) (*entity.User, error) {
	var user entity.User
	result := r.t.DB(ctx).Where("user_code = ? AND deleted_at = 0", userCode).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*entity.User, error) {
	var user entity.User
	result := r.t.DB(ctx).Where("phone = ? AND deleted_at = 0", phone).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	result := r.t.DB(ctx).Where("email = ? AND deleted_at = 0", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) GetAllWithCursor(ctx context.Context, cursorID *string, sortDesc bool, limit uint32, filters map[string]interface{}) ([]*entity.User, error) {
	var users []*entity.User
	query := r.t.DB(ctx).Where("deleted_at = 0")

	if filters != nil {
		if email, ok := filters["email"].(string); ok && email != "" {
			query = query.Where("email = ?", email)
		}
		if phone, ok := filters["phone"].(string); ok && phone != "" {
			query = query.Where("phone = ?", phone)
		}
	}

	if cursorID != nil && *cursorID != "" {
		if sortDesc {
			query = query.Where("id < ?", *cursorID)
		} else {
			query = query.Where("id > ?", *cursorID)
		}
	}

	if sortDesc {
		query = query.Order("id DESC")
	} else {
		query = query.Order("id ASC")
	}

	result := query.Limit(int(limit)).Find(&users)
	return users, result.Error
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	result := r.t.DB(ctx).Model(user).Where("id = ? AND deleted_at = 0", user.ID).Updates(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	result := r.t.DB(ctx).Model(&entity.User{}).
		Where("id = ? AND deleted_at = 0", id).
		Update("deleted_at", gorm.Expr("extract(epoch from now())::bigint"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepository) MarkPhoneVerified(ctx context.Context, id string) error {
	result := r.t.DB(ctx).Model(&entity.User{}).
		Where("id = ? AND deleted_at = 0", id).
		Update("is_phone_verified", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepository) MarkEmailVerified(ctx context.Context, id string) error {
	result := r.t.DB(ctx).Model(&entity.User{}).
		Where("id = ? AND deleted_at = 0", id).
		Update("is_email_verified", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepository) UpdateStats(ctx context.Context, userID string, wins, losses, draws, xCount, oCount, coins, gems int) error {
	result := r.t.DB(ctx).Model(&entity.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"wins":    gorm.Expr("wins + ?", wins),
			"losses":  gorm.Expr("losses + ?", losses),
			"draws":   gorm.Expr("draws + ?", draws),
			"x_count": gorm.Expr("x_count + ?", xCount),
			"o_count": gorm.Expr("o_count + ?", oCount),
			"coins":   gorm.Expr("coins + ?", coins),
			"gems":    gorm.Expr("gems + ?", gems),
		})
	return result.Error
}

func (r *userRepository) SetOnline(ctx context.Context, userID string, online bool) error {
	updates := map[string]interface{}{"is_online": online}
	if !online {
		updates["last_seen_at"] = gorm.Expr("extract(epoch from now())::bigint")
	}
	result := r.t.DB(ctx).Model(&entity.User{}).Where("id = ?", userID).Updates(updates)
	return result.Error
}

func (r *userRepository) UpdateCurrentStyle(ctx context.Context, userID string, theme, xoShape, avatar *int) error {
	updates := map[string]interface{}{}
	if theme != nil {
		updates["current_theme"] = *theme
	}
	if xoShape != nil {
		updates["current_xo_shape"] = *xoShape
	}
	if avatar != nil {
		updates["current_avatar"] = *avatar
	}
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = gorm.Expr("extract(epoch from now())::bigint")
	result := r.t.DB(ctx).Model(&entity.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepository) AddBalance(ctx context.Context, userID string, coins, gems int) error {
	result := r.t.DB(ctx).Model(&entity.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"coins":      gorm.Expr("coins + ?", coins),
			"gems":       gorm.Expr("gems + ?", gems),
			"updated_at": gorm.Expr("extract(epoch from now())::bigint"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepository) DeductBalance(ctx context.Context, userID string, currency byte, amount int) error {
	var result *gorm.DB
	if currency == 1 {
		result = r.t.DB(ctx).Model(&entity.User{}).
			Where("id = ? AND coins >= ?", userID, amount).
			Updates(map[string]interface{}{
				"coins":      gorm.Expr("coins - ?", amount),
				"updated_at": gorm.Expr("extract(epoch from now())::bigint"),
			})
	} else {
		result = r.t.DB(ctx).Model(&entity.User{}).
			Where("id = ? AND gems >= ?", userID, amount).
			Updates(map[string]interface{}{
				"gems":       gorm.Expr("gems - ?", amount),
				"updated_at": gorm.Expr("extract(epoch from now())::bigint"),
			})
	}
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appErrors.NewAppError("INSUFFICIENT_BALANCE", "Not enough balance", 400)
	}
	return nil
}
