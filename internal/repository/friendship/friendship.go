package friendship

import (
	"context"
	"errors"

	"dooz/entity"
	appErrors "dooz/internal/errors"
	"dooz/internal/repository/tx"

	"gorm.io/gorm"
)

var (
	ErrNotFound     = appErrors.ErrNotFound
	ErrAlreadyExist = appErrors.NewAppError("FRIENDSHIP_EXISTS", "Friendship already exists", 409)
)

type Repository interface {
	Create(ctx context.Context, f *entity.Friendship) error
	GetByID(ctx context.Context, id string) (*entity.Friendship, error)
	GetByUsers(ctx context.Context, requesterID, addresseeID string) (*entity.Friendship, error)
	Update(ctx context.Context, f *entity.Friendship) error
	Delete(ctx context.Context, id string) error
	GetFriendsOfUser(ctx context.Context, userID string) ([]*entity.Friendship, error)
	GetPendingForUser(ctx context.Context, userID string) ([]*entity.Friendship, error)
	CountAcceptedFriends(ctx context.Context, userID string) (int64, error)
}

type friendshipRepository struct {
	t tx.Transaction
}

func New(t tx.Transaction) Repository {
	return &friendshipRepository{t: t}
}

func (r *friendshipRepository) Create(ctx context.Context, f *entity.Friendship) error {
	result := r.t.DB(ctx).Create(f)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrAlreadyExist
		}
		return result.Error
	}
	return nil
}

func (r *friendshipRepository) GetByID(ctx context.Context, id string) (*entity.Friendship, error) {
	var f entity.Friendship
	result := r.t.DB(ctx).Where("id = ?", id).First(&f)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &f, nil
}

func (r *friendshipRepository) GetByUsers(ctx context.Context, requesterID, addresseeID string) (*entity.Friendship, error) {
	var f entity.Friendship
	result := r.t.DB(ctx).
		Where("(requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)",
			requesterID, addresseeID, addresseeID, requesterID).
		First(&f)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &f, nil
}

func (r *friendshipRepository) Update(ctx context.Context, f *entity.Friendship) error {
	result := r.t.DB(ctx).Model(f).Where("id = ?", f.ID).Updates(f)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *friendshipRepository) Delete(ctx context.Context, id string) error {
	result := r.t.DB(ctx).Delete(&entity.Friendship{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *friendshipRepository) GetFriendsOfUser(ctx context.Context, userID string) ([]*entity.Friendship, error) {
	var friendships []*entity.Friendship
	result := r.t.DB(ctx).
		Where("(requester_id = ? OR addressee_id = ?) AND status = ?",
			userID, userID, entity.FriendshipAccepted).
		Order("updated_at DESC").Find(&friendships)
	return friendships, result.Error
}

func (r *friendshipRepository) GetPendingForUser(ctx context.Context, userID string) ([]*entity.Friendship, error) {
	var friendships []*entity.Friendship
	result := r.t.DB(ctx).
		Where("addressee_id = ? AND status = ?", userID, entity.FriendshipPending).
		Order("id DESC").Find(&friendships)
	return friendships, result.Error
}

func (r *friendshipRepository) CountAcceptedFriends(ctx context.Context, userID string) (int64, error) {
	var count int64
	result := r.t.DB(ctx).Model(&entity.Friendship{}).
		Where("(requester_id = ? OR addressee_id = ?) AND status = ?",
			userID, userID, entity.FriendshipAccepted).
		Count(&count)
	return count, result.Error
}
