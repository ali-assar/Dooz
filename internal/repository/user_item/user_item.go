package user_item

import (
	"context"
	"errors"

	"dooz/entity"
	appErrors "dooz/internal/errors"
	"dooz/internal/repository/tx"

	"gorm.io/gorm"
)

var (
	ErrAlreadyOwned = appErrors.NewAppError("ITEM_ALREADY_OWNED", "Item already owned", 409)
	ErrNotOwned     = appErrors.NewAppError("ITEM_NOT_OWNED", "Item is not owned by user", 400)
)

type Repository interface {
	Create(ctx context.Context, userItem *entity.UserItem) error
	HasItem(ctx context.Context, userID, itemID string) (bool, error)
	GetByUserID(ctx context.Context, userID string) ([]*entity.UserItem, error)
}

type userItemRepository struct {
	t tx.Transaction
}

func New(t tx.Transaction) Repository {
	return &userItemRepository{t: t}
}

func (r *userItemRepository) Create(ctx context.Context, userItem *entity.UserItem) error {
	result := r.t.DB(ctx).Create(userItem)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrAlreadyOwned
		}
		return result.Error
	}
	return nil
}

func (r *userItemRepository) HasItem(ctx context.Context, userID, itemID string) (bool, error) {
	var count int64
	result := r.t.DB(ctx).Model(&entity.UserItem{}).
		Where("user_id = ? AND item_id = ?", userID, itemID).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func (r *userItemRepository) GetByUserID(ctx context.Context, userID string) ([]*entity.UserItem, error) {
	var items []*entity.UserItem
	result := r.t.DB(ctx).Where("user_id = ?", userID).Find(&items)
	return items, result.Error
}
