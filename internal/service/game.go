package service

import (
	"context"
	"log/slog"
	"time"

	"dooz/entity"
	"dooz/internal/app/api/dto"
	appErrors "dooz/internal/errors"
	wsHub "dooz/internal/infrastructure/websocket"
	boardRepo "dooz/internal/repository/board"
	moveRepo "dooz/internal/repository/move"
	userRepo "dooz/internal/repository/user"
)

type GameService interface {
	GetGameState(ctx context.Context, boardID, userID string) (*dto.GameStateResponse, error)
	MakeMove(ctx context.Context, boardID, userID string, position int) (*dto.GameStateResponse, error)
	Resign(ctx context.Context, boardID, userID string) (*dto.BoardDTO, error)
	GetHistory(ctx context.Context, userID string) ([]*dto.BoardDTO, error)
}

type gameService struct {
	boardRepo boardRepo.Repository
	moveRepo  moveRepo.Repository
	userRepo  userRepo.Repository
	hub       *wsHub.Hub
	logger    *slog.Logger
}

func NewGameService(
	boardRepo boardRepo.Repository,
	moveRepo moveRepo.Repository,
	userRepo userRepo.Repository,
	hub *wsHub.Hub,
	logger *slog.Logger,
) GameService {
	return &gameService{
		boardRepo: boardRepo,
		moveRepo:  moveRepo,
		userRepo:  userRepo,
		hub:       hub,
		logger:    logger.With("layer", "GameService"),
	}
}

func (s *gameService) GetGameState(ctx context.Context, boardID, userID string) (*dto.GameStateResponse, error) {
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if board.PlayerXID != userID && board.PlayerOID != userID {
		return nil, appErrors.ErrForbidden
	}

	moves, err := s.moveRepo.GetByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	moveDTOs := make([]*dto.MoveDTO, len(moves))
	for i, m := range moves {
		moveDTOs[i] = m.ToDTO()
	}

	return &dto.GameStateResponse{
		Board: board.ToDTO(),
		Moves: moveDTOs,
	}, nil
}

func (s *gameService) MakeMove(ctx context.Context, boardID, userID string, position int) (*dto.GameStateResponse, error) {
	lg := s.logger.With("method", "MakeMove", "boardID", boardID, "userID", userID, "position", position)

	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	if board.Status != entity.BoardStatusInProgress {
		return nil, appErrors.NewAppError("GAME_NOT_ACTIVE", "Game is not in progress", 400)
	}
	if board.CurrentTurn != userID {
		return nil, appErrors.NewAppError("NOT_YOUR_TURN", "It is not your turn", 400)
	}
	if board.BoardState[position] != '-' {
		return nil, appErrors.NewAppError("CELL_TAKEN", "Cell already occupied", 400)
	}

	symbol := "X"
	if board.PlayerOID == userID {
		symbol = "O"
	}

	newState := []byte(board.BoardState)
	newState[position] = symbol[0]
	board.BoardState = string(newState)

	now := time.Now().Unix()

	move := &entity.Move{
		BoardID:   boardID,
		UserID:    userID,
		Position:  position,
		Symbol:    symbol,
		CreatedAt: now,
	}
	if err := s.moveRepo.Create(ctx, move); err != nil {
		return nil, err
	}

	winner := checkWinner(board.BoardState)
	isDraw := winner == "" && isBoardFull(board.BoardState)

	if winner != "" || isDraw {
		board.Status = entity.BoardStatusCompleted
		board.EndedAt = now
		if winner == "X" {
			board.WinnerID = board.PlayerXID
		} else if winner == "O" {
			board.WinnerID = board.PlayerOID
		}
		s.updatePlayerStats(ctx, board, lg)
	} else {
		if board.CurrentTurn == board.PlayerXID {
			board.CurrentTurn = board.PlayerOID
		} else {
			board.CurrentTurn = board.PlayerXID
		}
	}

	board.UpdatedAt = now
	if err := s.boardRepo.Update(ctx, board); err != nil {
		return nil, err
	}

	moveDTOs := []*dto.MoveDTO{move.ToDTO()}

	if board.Status == entity.BoardStatusCompleted {
		s.hub.SendToUsers(
			[]string{board.PlayerXID, board.PlayerOID},
			wsHub.TypeGameEnd,
			board.ToDTO(),
		)
	} else {
		s.hub.SendToUsers(
			[]string{board.PlayerXID, board.PlayerOID},
			wsHub.TypeMove,
			move.ToDTO(),
		)
	}

	if board.IsBotGame && board.Status == entity.BoardStatusInProgress {
		go s.makeBotMove(context.Background(), board)
	}

	return &dto.GameStateResponse{
		Board: board.ToDTO(),
		Moves: moveDTOs,
	}, nil
}

func (s *gameService) makeBotMove(ctx context.Context, board *entity.Board) {
	time.Sleep(800 * time.Millisecond)

	pos := bestMove(board.BoardState)
	if pos == -1 {
		return
	}

	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, _ = s.MakeMove(ctx2, board.ID, board.PlayerOID, pos)
}

func (s *gameService) updatePlayerStats(ctx context.Context, board *entity.Board, lg *slog.Logger) {
	const coinWin = 25
	const coinDraw = 10

	playerX := board.PlayerXID
	playerO := board.PlayerOID

	if board.WinnerID == playerX {
		_ = s.userRepo.UpdateStats(ctx, playerX, 1, 0, 0, 1, 0, coinWin, 0)
		if playerO != "" && !board.IsBotGame {
			_ = s.userRepo.UpdateStats(ctx, playerO, 0, 1, 0, 0, 1, 0, 0)
		}
	} else if board.WinnerID == playerO && !board.IsBotGame {
		_ = s.userRepo.UpdateStats(ctx, playerO, 1, 0, 0, 0, 1, coinWin, 0)
		_ = s.userRepo.UpdateStats(ctx, playerX, 0, 1, 0, 1, 0, 0, 0)
	} else {
		_ = s.userRepo.UpdateStats(ctx, playerX, 0, 0, 1, 1, 0, coinDraw, 0)
		if playerO != "" && !board.IsBotGame {
			_ = s.userRepo.UpdateStats(ctx, playerO, 0, 0, 1, 0, 1, coinDraw, 0)
		}
	}
	lg.Info("player stats updated")
}

func (s *gameService) Resign(ctx context.Context, boardID, userID string) (*dto.BoardDTO, error) {
	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if board.Status != entity.BoardStatusInProgress {
		return nil, appErrors.NewAppError("GAME_NOT_ACTIVE", "Game is not in progress", 400)
	}
	if board.PlayerXID != userID && board.PlayerOID != userID {
		return nil, appErrors.ErrForbidden
	}

	now := time.Now().Unix()
	board.Status = entity.BoardStatusCompleted
	board.EndedAt = now
	board.UpdatedAt = now
	if board.PlayerXID == userID {
		board.WinnerID = board.PlayerOID
	} else {
		board.WinnerID = board.PlayerXID
	}

	if err := s.boardRepo.Update(ctx, board); err != nil {
		return nil, err
	}

	s.hub.SendToUsers([]string{board.PlayerXID, board.PlayerOID}, wsHub.TypeGameEnd, board.ToDTO())

	return board.ToDTO(), nil
}

func (s *gameService) GetHistory(ctx context.Context, userID string) ([]*dto.BoardDTO, error) {
	boards, err := s.boardRepo.GetByUserID(ctx, userID, 20)
	if err != nil {
		return nil, err
	}
	dtos := make([]*dto.BoardDTO, len(boards))
	for i, b := range boards {
		dtos[i] = b.ToDTO()
	}
	return dtos, nil
}

// checkWinner checks the board state string for a winner. Returns "X", "O", or "".
func checkWinner(state string) string {
	lines := [][3]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // cols
		{0, 4, 8}, {2, 4, 6}, // diagonals
	}
	for _, line := range lines {
		a, b, c := state[line[0]], state[line[1]], state[line[2]]
		if a != '-' && a == b && b == c {
			return string(a)
		}
	}
	return ""
}

func isBoardFull(state string) bool {
	for _, ch := range state {
		if ch == '-' {
			return false
		}
	}
	return true
}

// bestMove uses minimax to find the optimal move for the bot ('O').
func bestMove(state string) int {
	best := -1000
	bestPos := -1
	for i := 0; i < 9; i++ {
		if state[i] == '-' {
			newState := []byte(state)
			newState[i] = 'O'
			score := minimax(string(newState), 0, false)
			if score > best {
				best = score
				bestPos = i
			}
		}
	}
	return bestPos
}

func minimax(state string, depth int, isMaximizing bool) int {
	winner := checkWinner(state)
	if winner == "O" {
		return 10 - depth
	}
	if winner == "X" {
		return depth - 10
	}
	if isBoardFull(state) {
		return 0
	}

	if isMaximizing {
		best := -1000
		for i := 0; i < 9; i++ {
			if state[i] == '-' {
				newState := []byte(state)
				newState[i] = 'O'
				score := minimax(string(newState), depth+1, false)
				if score > best {
					best = score
				}
			}
		}
		return best
	}

	best := 1000
	for i := 0; i < 9; i++ {
		if state[i] == '-' {
			newState := []byte(state)
			newState[i] = 'X'
			score := minimax(string(newState), depth+1, true)
			if score < best {
				best = score
			}
		}
	}
	return best
}
