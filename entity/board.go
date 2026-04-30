package entity

import "dooz/internal/app/api/dto"

type BoardStatus byte

const (
	BoardStatusWaiting    BoardStatus = 1
	BoardStatusInProgress BoardStatus = 2
	BoardStatusCompleted  BoardStatus = 3
	BoardStatusAbandoned  BoardStatus = 4
)

func (s BoardStatus) String() string {
	switch s {
	case BoardStatusWaiting:
		return "waiting"
	case BoardStatusInProgress:
		return "in_progress"
	case BoardStatusCompleted:
		return "completed"
	case BoardStatusAbandoned:
		return "abandoned"
	default:
		return "unknown"
	}
}

// Board represents a tic-tac-toe game.
// BoardState is a 9-char string: 'X', 'O', or '-' for each cell (positions 0-8).
type Board struct {
	ID          string      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	PlayerXID   string      `gorm:"column:player_x_id;type:uuid;not null" json:"player_x_id"`
	PlayerOID   string      `gorm:"column:player_o_id;type:uuid" json:"player_o_id"`
	WinnerID    *string     `gorm:"column:winner_id;type:uuid" json:"winner_id"`
	Status      BoardStatus `gorm:"type:smallint;not null;default:1" json:"status"`
	IsBotGame   bool        `gorm:"column:is_bot_game;type:boolean;not null;default:false" json:"is_bot_game"`
	BoardState  string      `gorm:"column:board_state;type:char(9);not null;default:'---------'" json:"board_state"`
	CurrentTurn string      `gorm:"column:current_turn;type:uuid" json:"current_turn"`
	StartedAt   int64       `gorm:"column:started_at;type:bigint" json:"started_at"`
	EndedAt     int64       `gorm:"column:ended_at;type:bigint" json:"ended_at"`
	UpdatedAt   int64       `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (Board) TableName() string {
	return "boards"
}

func (b *Board) ToDTO() *dto.BoardDTO {
	winnerID := ""
	if b.WinnerID != nil {
		winnerID = *b.WinnerID
	}

	return &dto.BoardDTO{
		ID:          b.ID,
		PlayerXID:   b.PlayerXID,
		PlayerOID:   b.PlayerOID,
		WinnerID:    winnerID,
		Status:      b.Status.String(),
		IsBotGame:   b.IsBotGame,
		BoardState:  b.BoardState,
		CurrentTurn: b.CurrentTurn,
		StartedAt:   b.StartedAt,
		EndedAt:     b.EndedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}
