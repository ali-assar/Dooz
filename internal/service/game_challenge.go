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
	friendshipRepo "dooz/internal/repository/friendship"
	gameChallengeRepo "dooz/internal/repository/game_challenge"
	userRepo "dooz/internal/repository/user"
)

const challengeExpiry = 60 * time.Second

type GameChallengeService interface {
	CreateChallenge(ctx context.Context, requesterID, addresseeID string, addresseeCode *int) (*dto.GameChallengeDTO, error)
	GetPendingChallenges(ctx context.Context, userID string) ([]*dto.PendingChallengeDTO, error)
	AcceptChallenge(ctx context.Context, challengeID, userID string) (*dto.FindMatchResponse, error)
	RejectChallenge(ctx context.Context, challengeID, userID string) error
	CancelChallenge(ctx context.Context, challengeID, userID string) error
}

type gameChallengeService struct {
	challengeRepo  gameChallengeRepo.Repository
	friendshipRepo friendshipRepo.Repository
	userRepo       userRepo.Repository
	boardRepo      boardRepo.Repository
	hub            *wsHub.Hub
	logger         *slog.Logger
}

func NewGameChallengeService(
	challengeRepo gameChallengeRepo.Repository,
	friendshipRepo friendshipRepo.Repository,
	userRepo userRepo.Repository,
	boardRepo boardRepo.Repository,
	hub *wsHub.Hub,
	logger *slog.Logger,
) GameChallengeService {
	return &gameChallengeService{
		challengeRepo:  challengeRepo,
		friendshipRepo: friendshipRepo,
		userRepo:       userRepo,
		boardRepo:      boardRepo,
		hub:            hub,
		logger:         logger.With("layer", "GameChallengeService"),
	}
}

func (s *gameChallengeService) CreateChallenge(ctx context.Context, requesterID, addresseeID string, addresseeCode *int) (*dto.GameChallengeDTO, error) {
	if addresseeID == "" && addresseeCode == nil {
		return nil, appErrors.NewAppError("INVALID_ADDRESSEE", "Addressee id or code is required", 400)
	}
	if addresseeID == "" {
		user, err := s.userRepo.GetByUserCode(ctx, *addresseeCode)
		if err != nil {
			return nil, userRepo.ErrNotFound
		}
		addresseeID = user.ID
	}
	if requesterID == addresseeID {
		return nil, appErrors.NewAppError("SELF_CHALLENGE", "Cannot challenge yourself", 400)
	}

	friendship, err := s.friendshipRepo.GetByUsers(ctx, requesterID, addresseeID)
	if err != nil || friendship.Status != entity.FriendshipAccepted {
		return nil, appErrors.NewAppError("NOT_FRIENDS", "You can challenge only accepted friends", 403)
	}

	now := time.Now().Unix()
	if existing, err := s.challengeRepo.GetActivePair(ctx, requesterID, addresseeID); err == nil {
		if existing.ExpiresAt <= now {
			existing.Status = entity.GameChallengeExpired
			if err := s.challengeRepo.Update(ctx, existing); err != nil {
				return nil, err
			}
		} else {
			return nil, appErrors.NewAppError("CHALLENGE_EXISTS", "An active challenge already exists", 409)
		}
	}

	challenge := &entity.GameChallenge{
		RequesterID: requesterID,
		AddresseeID: addresseeID,
		Status:      entity.GameChallengePending,
		ExpiresAt:   now + int64(challengeExpiry.Seconds()),
	}
	if err := s.challengeRepo.Create(ctx, challenge); err != nil {
		if err == gameChallengeRepo.ErrActiveExists {
			return nil, appErrors.NewAppError("CHALLENGE_EXISTS", "An active challenge already exists", 409)
		}
		return nil, err
	}

	s.hub.SendToUsers([]string{addresseeID}, wsHub.TypeChallengeReceived, challenge.ToDTO())
	return challenge.ToDTO(), nil
}

func (s *gameChallengeService) GetPendingChallenges(ctx context.Context, userID string) ([]*dto.PendingChallengeDTO, error) {
	now := time.Now().Unix()
	challenges, err := s.challengeRepo.GetPendingForAddressee(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.PendingChallengeDTO, 0, len(challenges))
	for _, c := range challenges {
		requester, err := s.userRepo.GetByID(ctx, c.RequesterID)
		if err != nil {
			continue
		}
		result = append(result, &dto.PendingChallengeDTO{
			Challenge: *c.ToDTO(),
			Requester: *requester.ToDTO(),
		})
	}
	return result, nil
}

func (s *gameChallengeService) AcceptChallenge(ctx context.Context, challengeID, userID string) (*dto.FindMatchResponse, error) {
	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	if challenge.AddresseeID != userID {
		return nil, appErrors.ErrForbidden
	}
	now := time.Now().Unix()
	if challenge.Status != entity.GameChallengePending || challenge.ExpiresAt <= now {
		return nil, appErrors.NewAppError("CHALLENGE_NOT_PENDING", "Challenge is no longer pending", 400)
	}

	board := &entity.Board{
		PlayerXID:   challenge.RequesterID,
		PlayerOID:   challenge.AddresseeID,
		Status:      entity.BoardStatusInProgress,
		IsBotGame:   false,
		BoardState:  "---------",
		CurrentTurn: challenge.RequesterID,
		StartedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.boardRepo.Create(ctx, board); err != nil {
		return nil, err
	}

	challenge.Status = entity.GameChallengeAccepted
	challenge.BoardID = &board.ID
	if err := s.challengeRepo.Update(ctx, challenge); err != nil {
		return nil, err
	}

	s.hub.SendToUsers([]string{challenge.RequesterID}, wsHub.TypeChallengeAccepted, challenge.ToDTO())
	s.hub.SendToUsers([]string{challenge.RequesterID}, wsHub.TypeMatchFound, dto.FindMatchResponse{
		BoardID:   board.ID,
		IsBotGame: false,
		Symbol:    "X",
	})
	s.hub.SendToUsers([]string{challenge.AddresseeID}, wsHub.TypeMatchFound, dto.FindMatchResponse{
		BoardID:   board.ID,
		IsBotGame: false,
		Symbol:    "O",
	})

	return &dto.FindMatchResponse{
		BoardID:   board.ID,
		IsBotGame: false,
		Symbol:    "O",
	}, nil
}

func (s *gameChallengeService) RejectChallenge(ctx context.Context, challengeID, userID string) error {
	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return err
	}
	if challenge.AddresseeID != userID {
		return appErrors.ErrForbidden
	}
	if challenge.Status != entity.GameChallengePending {
		return appErrors.NewAppError("CHALLENGE_NOT_PENDING", "Challenge is no longer pending", 400)
	}
	challenge.Status = entity.GameChallengeRejected
	if err := s.challengeRepo.Update(ctx, challenge); err != nil {
		return err
	}
	s.hub.SendToUsers([]string{challenge.RequesterID}, wsHub.TypeChallengeRejected, challenge.ToDTO())
	return nil
}

func (s *gameChallengeService) CancelChallenge(ctx context.Context, challengeID, userID string) error {
	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return err
	}
	if challenge.RequesterID != userID {
		return appErrors.ErrForbidden
	}
	if challenge.Status != entity.GameChallengePending {
		return appErrors.NewAppError("CHALLENGE_NOT_PENDING", "Challenge is no longer pending", 400)
	}
	challenge.Status = entity.GameChallengeCanceled
	if err := s.challengeRepo.Update(ctx, challenge); err != nil {
		return err
	}
	s.hub.SendToUsers([]string{challenge.AddresseeID}, wsHub.TypeChallengeCanceled, challenge.ToDTO())
	return nil
}
