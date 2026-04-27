package achievement

import (
	"context"
	"errors"

	"dooz/entity"
	appErrors "dooz/internal/errors"
	"dooz/internal/repository/tx"

	"gorm.io/gorm"
)

var ErrNotFound = appErrors.ErrNotFound

type Repository interface {
	Create(ctx context.Context, a *entity.Achievement) error
	GetAll(ctx context.Context) ([]*entity.Achievement, error)
	GetByID(ctx context.Context, id string) (*entity.Achievement, error)
	GetByRequirementType(ctx context.Context, reqType entity.RequirementType) ([]*entity.Achievement, error)
}

type achievementRepository struct {
	t tx.Transaction
}

func New(t tx.Transaction) Repository {
	return &achievementRepository{t: t}
}

func (r *achievementRepository) Create(ctx context.Context, a *entity.Achievement) error {
	return r.t.DB(ctx).Create(a).Error
}

func (r *achievementRepository) GetAll(ctx context.Context) ([]*entity.Achievement, error) {
	var achievements []*entity.Achievement
	result := r.t.DB(ctx).Order("requirement_value ASC").Find(&achievements)
	return achievements, result.Error
}

func (r *achievementRepository) GetByID(ctx context.Context, id string) (*entity.Achievement, error) {
	var a entity.Achievement
	result := r.t.DB(ctx).Where("id = ?", id).First(&a)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &a, nil
}

func (r *achievementRepository) GetByRequirementType(ctx context.Context, reqType entity.RequirementType) ([]*entity.Achievement, error) {
	var achievements []*entity.Achievement
	result := r.t.DB(ctx).Where("requirement_type = ?", reqType).Order("requirement_value ASC").Find(&achievements)
	return achievements, result.Error
}
