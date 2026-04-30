package move

import (
	"context"

	"dooz/entity"
	"dooz/internal/repository/tx"
)

type Repository interface {
	Create(ctx context.Context, move *entity.Move) error
	GetByBoardID(ctx context.Context, boardID string) ([]*entity.Move, error)
}

type moveRepository struct {
	t tx.Transaction
}

func New(t tx.Transaction) Repository {
	return &moveRepository{t: t}
}

func (r *moveRepository) Create(ctx context.Context, move *entity.Move) error {
	return r.t.DB(ctx).Create(move).Error
}

func (r *moveRepository) GetByBoardID(ctx context.Context, boardID string) ([]*entity.Move, error) {
	var moves []*entity.Move
	result := r.t.DB(ctx).Where("board_id = ?", boardID).Order("id ASC").Find(&moves)
	return moves, result.Error
}
