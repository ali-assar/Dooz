package entity

import "dooz/internal/app/api/dto"

type Move struct {
	ID        string `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	BoardID   string `gorm:"column:board_id;type:uuid;not null" json:"board_id"`
	UserID    string `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	Position  int    `gorm:"column:position;type:smallint;not null" json:"position"` // 0-8
	Symbol    string `gorm:"column:symbol;type:char(1);not null" json:"symbol"`      // 'X' or 'O'
	CreatedAt int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
}

func (Move) TableName() string {
	return "moves"
}

func (m *Move) ToDTO() *dto.MoveDTO {
	return &dto.MoveDTO{
		ID:        m.ID,
		BoardID:   m.BoardID,
		UserID:    m.UserID,
		Position:  m.Position,
		Symbol:    m.Symbol,
		CreatedAt: m.CreatedAt,
	}
}
