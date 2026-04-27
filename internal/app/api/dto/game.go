package dto

type BoardDTO struct {
	ID          string `json:"id"`
	PlayerXID   string `json:"player_x_id"`
	PlayerOID   string `json:"player_o_id"`
	WinnerID    string `json:"winner_id,omitempty"`
	Status      string `json:"status"`
	IsBotGame   bool   `json:"is_bot_game"`
	BoardState  string `json:"board_state"`
	CurrentTurn string `json:"current_turn,omitempty"`
	StartedAt   int64  `json:"started_at,omitempty"`
	EndedAt     int64  `json:"ended_at,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type MoveDTO struct {
	ID        string `json:"id"`
	BoardID   string `json:"board_id"`
	UserID    string `json:"user_id"`
	Position  int    `json:"position"`
	Symbol    string `json:"symbol"`
	CreatedAt int64  `json:"created_at"`
}

type MakeMoveRequest struct {
	BoardID  string `uri:"id" binding:"required"`
	Position int    `json:"position" binding:"required,min=0,max=8"`
}

type FindMatchResponse struct {
	BoardID   string `json:"board_id"`
	IsBotGame bool   `json:"is_bot_game"`
	Symbol    string `json:"symbol"`
}

type GameStateResponse struct {
	Board *BoardDTO  `json:"board"`
	Moves []*MoveDTO `json:"moves"`
}
