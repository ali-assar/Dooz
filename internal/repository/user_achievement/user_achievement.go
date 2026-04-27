package user_achievement

import (
	"context"

	"dooz/entity"
	"dooz/internal/repository/tx"
)

type Repository interface {
	Create(ctx context.Context, ua *entity.UserAchievement) error
	GetByUserID(ctx context.Context, userID string) ([]*entity.UserAchievement, error)
	HasEarned(ctx context.Context, userID, achievementID string) (bool, error)
}

type userAchievementRepository struct {
	t tx.Transaction
}

func New(t tx.Transaction) Repository {
	return &userAchievementRepository{t: t}
}

func (r *userAchievementRepository) Create(ctx context.Context, ua *entity.UserAchievement) error {
	return r.t.DB(ctx).Create(ua).Error
}

func (r *userAchievementRepository) GetByUserID(ctx context.Context, userID string) ([]*entity.UserAchievement, error) {
	var uas []*entity.UserAchievement
	result := r.t.DB(ctx).Where("user_id = ?", userID).Order("earned_at DESC").Find(&uas)
	return uas, result.Error
}

func (r *userAchievementRepository) HasEarned(ctx context.Context, userID, achievementID string) (bool, error) {
	var count int64
	result := r.t.DB(ctx).Model(&entity.UserAchievement{}).
		Where("user_id = ? AND achievement_id = ?", userID, achievementID).Count(&count)
	return count > 0, result.Error
}
