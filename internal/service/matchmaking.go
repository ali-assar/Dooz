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
	userRepo "dooz/internal/repository/user"

	goredis "github.com/redis/go-redis/v9"
)

const matchmakingQueueKey = "dooz:matchmaking:queue"
const botUserID = "00000000-0000-0000-0000-000000000000"
const botUserEmail = "bot@dooz.local"

type MatchmakingService interface {
	FindMatch(ctx context.Context, userID string) (*dto.FindMatchResponse, error)
}

type matchmakingService struct {
	boardRepo   boardRepo.Repository
	userRepo    userRepo.Repository
	redisClient *goredis.Client
	hub         *wsHub.Hub
	logger      *slog.Logger
}

func NewMatchmakingService(
	boardRepo boardRepo.Repository,
	userRepo userRepo.Repository,
	redisClient *goredis.Client,
	hub *wsHub.Hub,
	logger *slog.Logger,
) MatchmakingService {
	return &matchmakingService{
		boardRepo:   boardRepo,
		userRepo:    userRepo,
		redisClient: redisClient,
		hub:         hub,
		logger:      logger.With("layer", "MatchmakingService"),
	}
}

func (s *matchmakingService) FindMatch(ctx context.Context, userID string) (*dto.FindMatchResponse, error) {
	lg := s.logger.With("method", "FindMatch", "userID", userID)
	lg.Info("matchmaking started")

	deadline := time.Now().Add(constants.MatchmakingTimeout)
	ticker := time.NewTicker(constants.MatchmakingPollInterval)
	defer ticker.Stop()

	// Try to pop another waiting user from the queue.
	for {
		select {
		case <-ctx.Done():
			// Remove self from queue on context cancel
			s.redisClient.LRem(context.Background(), matchmakingQueueKey, 0, userID)
			lg.Warn("matchmaking canceled by context", "error", ctx.Err())
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
				lg.Info("matched with opponent", "opponentID", opponentID, "boardID", board.ID)
				return &dto.FindMatchResponse{
					BoardID:   board.ID,
					IsBotGame: false,
					Symbol:    "O",
				}, nil
			}

			if err == nil && opponentID == userID {
				// We popped ourselves; put back
				s.enqueueUserUnique(ctx, userID)
				lg.Debug("popped self from queue, pushed back")
			}

			if time.Now().After(deadline) {
				// Timeout — assign bot
				s.redisClient.LRem(context.Background(), matchmakingQueueKey, 0, userID)
				board, err := s.createBotBoard(ctx, userID)
				if err != nil {
					lg.Error("failed to create bot board", "error", err)
					return nil, err
				}
				lg.Info("matchmaking timeout reached, matched with bot", "boardID", board.ID)
				return &dto.FindMatchResponse{
					BoardID:   board.ID,
					IsBotGame: true,
					Symbol:    "X",
				}, nil
			}

			// Add self to queue if not already there
			if opponentID != userID {
				if err := s.enqueueUserUnique(ctx, userID); err != nil {
					lg.Error("failed to push to queue", "error", err)
				} else {
					lg.Debug("user queued for matchmaking")
				}
			}
		}
	}
}

func (s *matchmakingService) enqueueUserUnique(ctx context.Context, userID string) error {
	pipe := s.redisClient.TxPipeline()
	pipe.LRem(ctx, matchmakingQueueKey, 0, userID)
	pipe.RPush(ctx, matchmakingQueueKey, userID)
	_, err := pipe.Exec(ctx)
	return err
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
		UpdatedAt:   now,
	}
	if err := s.boardRepo.Create(ctx, board); err != nil {
		return nil, err
	}
	return board, nil
}

func (s *matchmakingService) createBotBoard(ctx context.Context, playerXID string) (*entity.Board, error) {
	if err := s.ensureBotUser(ctx); err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	board := &entity.Board{
		PlayerXID:   playerXID,
		PlayerOID:   botUserID,
		Status:      entity.BoardStatusInProgress,
		IsBotGame:   true,
		BoardState:  "---------",
		CurrentTurn: playerXID,
		StartedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.boardRepo.Create(ctx, board); err != nil {
		return nil, err
	}
	return board, nil
}

func (s *matchmakingService) ensureBotUser(ctx context.Context) error {
	if _, err := s.userRepo.GetByID(ctx, botUserID); err == nil {
		return nil
	}

	now := time.Now().Unix()
	bot := &entity.User{
		ID:              botUserID,
		Email:           botUserEmail,
		Fullname:        "Dooz Bot",
		UserCode:        999999,
		Role:            entity.RoleUser,
		IsEmailVerified: true,
		CurrentTheme:    1,
		CurrentXOShape:  1,
		CurrentAvatar:   1,
		UpdatedAt:       now,
	}

	if err := s.userRepo.Create(ctx, bot); err != nil {
		// Handle race: if another request created bot user first.
		if _, getErr := s.userRepo.GetByID(ctx, botUserID); getErr == nil {
			return nil
		}
		return err
	}
	return nil
}
