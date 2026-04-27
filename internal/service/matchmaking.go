package service

import (
	"context"
	"log/slog"
	"time"

	"dooz/entity"
	"dooz/internal/app/api/dto"
	"dooz/internal/constants"
	wsHub "dooz/internal/infrastructure/websocket"
	boardRepo "dooz/internal/repository/board"

	goredis "github.com/redis/go-redis/v9"
)

const matchmakingQueueKey = "dooz:matchmaking:queue"
const botUserID = "00000000-0000-0000-0000-000000000000"

type MatchmakingService interface {
	FindMatch(ctx context.Context, userID string) (*dto.FindMatchResponse, error)
}

type matchmakingService struct {
	boardRepo   boardRepo.Repository
	redisClient *goredis.Client
	hub         *wsHub.Hub
	logger      *slog.Logger
}

func NewMatchmakingService(
	boardRepo boardRepo.Repository,
	redisClient *goredis.Client,
	hub *wsHub.Hub,
	logger *slog.Logger,
) MatchmakingService {
	return &matchmakingService{
		boardRepo:   boardRepo,
		redisClient: redisClient,
		hub:         hub,
		logger:      logger.With("layer", "MatchmakingService"),
	}
}

func (s *matchmakingService) FindMatch(ctx context.Context, userID string) (*dto.FindMatchResponse, error) {
	lg := s.logger.With("method", "FindMatch", "userID", userID)

	deadline := time.Now().Add(constants.MatchmakingTimeout)
	ticker := time.NewTicker(constants.MatchmakingPollInterval)
	defer ticker.Stop()

	// Try to pop another waiting user from the queue.
	for {
		select {
		case <-ctx.Done():
			// Remove self from queue on context cancel
			s.redisClient.LRem(context.Background(), matchmakingQueueKey, 0, userID)
			return nil, ctx.Err()

		case <-ticker.C:
			opponentID, err := s.redisClient.LPop(ctx, matchmakingQueueKey).Result()
			if err == nil && opponentID != userID {
				// Matched with a real opponent
				board, err := s.createBoard(ctx, opponentID, userID)
				if err != nil {
					lg.Error("failed to create board", "error", err)
					return nil, err
				}
				// Notify opponent via WS
				s.hub.SendToUsers([]string{opponentID}, wsHub.TypeMatchFound, dto.FindMatchResponse{
					BoardID:   board.ID,
					IsBotGame: false,
					Symbol:    "X",
				})
				lg.Info("matched with opponent", "opponent", opponentID)
				return &dto.FindMatchResponse{
					BoardID:   board.ID,
					IsBotGame: false,
					Symbol:    "O",
				}, nil
			}

			if err == nil && opponentID == userID {
				// We popped ourselves; put back
				s.redisClient.RPush(ctx, matchmakingQueueKey, userID)
			}

			if time.Now().After(deadline) {
				// Timeout — assign bot
				s.redisClient.LRem(context.Background(), matchmakingQueueKey, 0, userID)
				board, err := s.createBotBoard(ctx, userID)
				if err != nil {
					lg.Error("failed to create bot board", "error", err)
					return nil, err
				}
				lg.Info("matched with bot")
				return &dto.FindMatchResponse{
					BoardID:   board.ID,
					IsBotGame: true,
					Symbol:    "X",
				}, nil
			}

			// Add self to queue if not already there
			if opponentID != userID {
				if err := s.redisClient.RPush(ctx, matchmakingQueueKey, userID).Err(); err != nil {
					lg.Error("failed to push to queue", "error", err)
				}
			}
		}
	}
}

func (s *matchmakingService) createBoard(ctx context.Context, playerXID, playerOID string) (*entity.Board, error) {
	now := time.Now().Unix()
	board := &entity.Board{
		PlayerXID:   playerXID,
		PlayerOID:   playerOID,
		Status:      entity.BoardStatusInProgress,
		IsBotGame:   false,
		BoardState:  "---------",
		CurrentTurn: playerXID,
		StartedAt:   now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.boardRepo.Create(ctx, board); err != nil {
		return nil, err
	}
	return board, nil
}

func (s *matchmakingService) createBotBoard(ctx context.Context, playerXID string) (*entity.Board, error) {
	now := time.Now().Unix()
	board := &entity.Board{
		PlayerXID:   playerXID,
		PlayerOID:   botUserID,
		Status:      entity.BoardStatusInProgress,
		IsBotGame:   true,
		BoardState:  "---------",
		CurrentTurn: playerXID,
		StartedAt:   now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.boardRepo.Create(ctx, board); err != nil {
		return nil, err
	}
	return board, nil
}
