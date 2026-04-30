package entity

import "dooz/internal/app/api/dto"

type ChatMessage struct {
	ID         string  `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	SenderID   string  `gorm:"column:sender_id;type:uuid;not null" json:"sender_id"`
	ReceiverID *string `gorm:"column:receiver_id;type:uuid" json:"receiver_id"` // for DMs
	BoardID    *string `gorm:"column:board_id;type:uuid" json:"board_id"`       // for in-game chat
	Content    string  `gorm:"type:text;not null" json:"content"`
	ReadAt     int64   `gorm:"column:read_at;type:bigint" json:"read_at"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

func (m *ChatMessage) ToDTO() *dto.ChatMessageDTO {
	receiverID := ""
	if m.ReceiverID != nil {
		receiverID = *m.ReceiverID
	}

	boardID := ""
	if m.BoardID != nil {
		boardID = *m.BoardID
	}

	return &dto.ChatMessageDTO{
		ID:         m.ID,
		SenderID:   m.SenderID,
		ReceiverID: receiverID,
		BoardID:    boardID,
		Content:    m.Content,
		ReadAt:     m.ReadAt,
	}
}
