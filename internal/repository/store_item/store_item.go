package store_item

import (
	"context"
	"errors"

	"dooz/entity"
	appErrors "dooz/internal/errors"
	"dooz/internal/repository/tx"

	"gorm.io/gorm"
)

var (
	ErrNotFound = appErrors.ErrNotFound
)

type Repository interface {
	GetActive(ctx context.Context) ([]*entity.StoreItem, error)
	GetByKey(ctx context.Context, itemKey string) (*entity.StoreItem, error)
	GetByTypeAndValue(ctx context.Context, itemType byte, itemValue int) (*entity.StoreItem, error)
}

type storeItemRepository struct {
	t tx.Transaction
}

func New(t tx.Transaction) Repository {
	return &storeItemRepository{t: t}
}

func (r *storeItemRepository) GetActive(ctx context.Context) ([]*entity.StoreItem, error) {
	var items []*entity.StoreItem
	result := r.t.DB(ctx).Where("is_active = ?", true).Order("item_type ASC, item_key ASC").Find(&items)
	return items, result.Error
}

func (r *storeItemRepository) GetByKey(ctx context.Context, itemKey string) (*entity.StoreItem, error) {
	var item entity.StoreItem
	result := r.t.DB(ctx).Where("item_key = ? AND is_active = ?", itemKey, true).First(&item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &item, nil
}

func (r *storeItemRepository) GetByTypeAndValue(ctx context.Context, itemType byte, itemValue int) (*entity.StoreItem, error) {
	var item entity.StoreItem
	result := r.t.DB(ctx).Where("item_type = ? AND item_value = ? AND is_active = ?", itemType, itemValue, true).First(&item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &item, nil
}
