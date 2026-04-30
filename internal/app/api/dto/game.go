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
	UpdatedAt   int64  `json:"updated_at"`
}

type MoveDTO struct {
	ID       string `json:"id"`
	BoardID  string `json:"board_id"`
	UserID   string `json:"user_id"`
	Position int    `json:"position"`
	Symbol   string `json:"symbol"`
}

type MakeMoveURI struct {
	BoardID string `uri:"id" binding:"required"`
}

type MakeMoveRequest struct {
	Position *int `json:"position" binding:"required,min=0,max=8"`
}

type FindMatchResponse struct {
	BoardID   string `json:"board_id"`
	IsBotGame bool   `json:"is_bot_game"`
	Symbol    string `json:"symbol"`
}

type CreateChallengeRequest struct {
	AddresseeID   string `json:"addressee_id,omitempty"`
	AddresseeCode *int   `json:"addressee_code,omitempty" binding:"omitempty,min=100000,max=999999"`
}

type GameChallengeDTO struct {
	ID          string `json:"id"`
	RequesterID string `json:"requester_id"`
	AddresseeID string `json:"addressee_id"`
	BoardID     string `json:"board_id,omitempty"`
	Status      string `json:"status"`
	ExpiresAt   int64  `json:"expires_at"`
}

type PendingChallengeDTO struct {
	Challenge GameChallengeDTO `json:"challenge"`
	Requester UserDTO          `json:"requester"`
}

type GameStateResponse struct {
	Board *BoardDTO  `json:"board"`
	Moves []*MoveDTO `json:"moves"`
}
