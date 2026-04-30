package chat_message

import (
	"context"

	"dooz/entity"
	"dooz/internal/repository/tx"
)

type Repository interface {
	Create(ctx context.Context, msg *entity.ChatMessage) error
	GetDMHistory(ctx context.Context, userA, userB string, limit int, beforeID string) ([]*entity.ChatMessage, error)
	GetGameChatHistory(ctx context.Context, boardID string) ([]*entity.ChatMessage, error)
	MarkRead(ctx context.Context, senderID, receiverID string) error
}

type chatMessageRepository struct {
	t tx.Transaction
}

func New(t tx.Transaction) Repository {
	return &chatMessageRepository{t: t}
}

func (r *chatMessageRepository) Create(ctx context.Context, msg *entity.ChatMessage) error {
	return r.t.DB(ctx).Create(msg).Error
}

func (r *chatMessageRepository) GetDMHistory(ctx context.Context, userA, userB string, limit int, beforeID string) ([]*entity.ChatMessage, error) {
	var messages []*entity.ChatMessage
	query := r.t.DB(ctx).
		Where("((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?))",
			userA, userB, userB, userA)

	if beforeID != "" {
		query = query.Where("id < ?", beforeID)
	}

	result := query.Order("id DESC").Limit(limit).Find(&messages)
	return messages, result.Error
}

func (r *chatMessageRepository) GetGameChatHistory(ctx context.Context, boardID string) ([]*entity.ChatMessage, error) {
	var messages []*entity.ChatMessage
	result := r.t.DB(ctx).Where("board_id = ?", boardID).Order("id ASC").Find(&messages)
	return messages, result.Error
}

func (r *chatMessageRepository) MarkRead(ctx context.Context, senderID, receiverID string) error {
	return r.t.DB(ctx).Model(&entity.ChatMessage{}).
		Where("sender_id = ? AND receiver_id = ? AND read_at = 0", senderID, receiverID).
		Update("read_at", "extract(epoch from now())::bigint").Error
}
