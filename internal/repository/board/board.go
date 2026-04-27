package board

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
	Create(ctx context.Context, board *entity.Board) error
	GetByID(ctx context.Context, id string) (*entity.Board, error)
	Update(ctx context.Context, board *entity.Board) error
	GetByUserID(ctx context.Context, userID string, limit int) ([]*entity.Board, error)
	// GetWaitingBoard finds any waiting board that is not the requesting user's own board.
	GetWaitingBoard(ctx context.Context, excludeUserID string) (*entity.Board, error)
	CountCompletedByUser(ctx context.Context, userID string) (int64, error)
}

type boardRepository struct {
	t tx.Transaction
}

func New(t tx.Transaction) Repository {
	return &boardRepository{t: t}
}

func (r *boardRepository) Create(ctx context.Context, board *entity.Board) error {
	return r.t.DB(ctx).Create(board).Error
}

func (r *boardRepository) GetByID(ctx context.Context, id string) (*entity.Board, error) {
	var board entity.Board
	result := r.t.DB(ctx).Where("id = ?", id).First(&board)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &board, nil
}

func (r *boardRepository) Update(ctx context.Context, board *entity.Board) error {
	result := r.t.DB(ctx).Model(board).Where("id = ?", board.ID).Updates(board)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *boardRepository) GetByUserID(ctx context.Context, userID string, limit int) ([]*entity.Board, error) {
	var boards []*entity.Board
	result := r.t.DB(ctx).
		Where("player_x_id = ? OR player_o_id = ?", userID, userID).
		Order("created_at DESC").
		Limit(limit).Find(&boards)
	return boards, result.Error
}

func (r *boardRepository) GetWaitingBoard(ctx context.Context, excludeUserID string) (*entity.Board, error) {
	var board entity.Board
	result := r.t.DB(ctx).
		Where("status = ? AND player_x_id != ?", entity.BoardStatusWaiting, excludeUserID).
		Order("created_at ASC").
		First(&board)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &board, nil
}

func (r *boardRepository) CountCompletedByUser(ctx context.Context, userID string) (int64, error) {
	var count int64
	result := r.t.DB(ctx).Model(&entity.Board{}).
		Where("(player_x_id = ? OR player_o_id = ?) AND status = ?", userID, userID, entity.BoardStatusCompleted).
		Count(&count)
	return count, result.Error
}
