package service

import (
	"context"
	"log/slog"
	"time"

	"dooz/entity"
	"dooz/internal/app/api/dto"
	appErrors "dooz/internal/errors"
	friendshipRepo "dooz/internal/repository/friendship"
	userRepo "dooz/internal/repository/user"
)

type FriendService interface {
	SendRequest(ctx context.Context, requesterID, addresseeID string, addresseeCode *int) (*dto.FriendshipDTO, error)
	AcceptRequest(ctx context.Context, friendshipID, userID string) (*dto.FriendshipDTO, error)
	RejectRequest(ctx context.Context, friendshipID, userID string) error
	RemoveFriend(ctx context.Context, friendshipID, userID string) error
	GetFriends(ctx context.Context, userID string) ([]*dto.FriendWithUserDTO, error)
	GetPendingRequests(ctx context.Context, userID string) ([]*dto.PendingFriendRequestDTO, error)
}

type friendService struct {
	friendshipRepo     friendshipRepo.Repository
	userRepo           userRepo.Repository
	achievementService AchievementService
	logger             *slog.Logger
}

func NewFriendService(
	friendshipRepo friendshipRepo.Repository,
	userRepo userRepo.Repository,
	achievementService AchievementService,
	logger *slog.Logger,
) FriendService {
	return &friendService{
		friendshipRepo:     friendshipRepo,
		userRepo:           userRepo,
		achievementService: achievementService,
		logger:             logger.With("layer", "FriendService"),
	}
}

func (s *friendService) SendRequest(ctx context.Context, requesterID, addresseeID string, addresseeCode *int) (*dto.FriendshipDTO, error) {
	lg := s.logger.With("method", "SendRequest", "requesterID", requesterID, "addresseeID", addresseeID, "addresseeCode", addresseeCode)

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
		return nil, appErrors.NewAppError("SELF_FRIEND", "Cannot send friend request to yourself", 400)
	}

	if _, err := s.userRepo.GetByID(ctx, addresseeID); err != nil {
		return nil, userRepo.ErrNotFound
	}

	existing, err := s.friendshipRepo.GetByUsers(ctx, requesterID, addresseeID)
	if err == nil && existing != nil {
		if existing.Status == entity.FriendshipAccepted {
			return nil, appErrors.NewAppError("ALREADY_FRIENDS", "Already friends", 409)
		}
		if existing.Status == entity.FriendshipPending {
			return nil, appErrors.NewAppError("REQUEST_EXISTS", "Friend request already sent", 409)
		}
	}

	now := time.Now().Unix()
	f := &entity.Friendship{
		RequesterID: requesterID,
		AddresseeID: addresseeID,
		Status:      entity.FriendshipPending,
		UpdatedAt:   now,
	}

	if err := s.friendshipRepo.Create(ctx, f); err != nil {
		lg.Error("failed to create friendship", "error", err)
		return nil, err
	}

	return f.ToDTO(), nil
}

func (s *friendService) AcceptRequest(ctx context.Context, friendshipID, userID string) (*dto.FriendshipDTO, error) {
	f, err := s.friendshipRepo.GetByID(ctx, friendshipID)
	if err != nil {
		return nil, friendshipRepo.ErrNotFound
	}
	if f.AddresseeID != userID {
		return nil, appErrors.ErrForbidden
	}
	if f.Status != entity.FriendshipPending {
		return nil, appErrors.NewAppError("NOT_PENDING", "Request is not pending", 400)
	}

	f.Status = entity.FriendshipAccepted
	f.UpdatedAt = time.Now().Unix()

	if err := s.friendshipRepo.Update(ctx, f); err != nil {
		return nil, err
	}
	for _, uid := range []string{f.RequesterID, f.AddresseeID} {
		if err := s.achievementService.CheckAndAward(ctx, uid); err != nil {
			s.logger.Warn("achievement check failed", "method", "AcceptRequest", "userID", uid, "error", err)
		}
	}
	return f.ToDTO(), nil
}

func (s *friendService) RejectRequest(ctx context.Context, friendshipID, userID string) error {
	f, err := s.friendshipRepo.GetByID(ctx, friendshipID)
	if err != nil {
		return friendshipRepo.ErrNotFound
	}
	if f.AddresseeID != userID {
		return appErrors.ErrForbidden
	}

	f.Status = entity.FriendshipRejected
	f.UpdatedAt = time.Now().Unix()
	return s.friendshipRepo.Update(ctx, f)
}

func (s *friendService) RemoveFriend(ctx context.Context, friendshipID, userID string) error {
	f, err := s.friendshipRepo.GetByID(ctx, friendshipID)
	if err != nil {
		return friendshipRepo.ErrNotFound
	}
	if f.RequesterID != userID && f.AddresseeID != userID {
		return appErrors.ErrForbidden
	}
	return s.friendshipRepo.Delete(ctx, friendshipID)
}

func (s *friendService) GetFriends(ctx context.Context, userID string) ([]*dto.FriendWithUserDTO, error) {
	friendships, err := s.friendshipRepo.GetFriendsOfUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.FriendWithUserDTO, 0, len(friendships))
	for _, f := range friendships {
		friendID := f.AddresseeID
		if f.RequesterID != userID {
			friendID = f.RequesterID
		}
		friend, err := s.userRepo.GetByID(ctx, friendID)
		if err != nil {
			continue
		}
		result = append(result, &dto.FriendWithUserDTO{
			FriendshipID: f.ID,
			Friend:       *friend.ToDTO(),
			Status:       f.Status.String(),
		})
	}
	return result, nil
}

func (s *friendService) GetPendingRequests(ctx context.Context, userID string) ([]*dto.PendingFriendRequestDTO, error) {
	friendships, err := s.friendshipRepo.GetPendingForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.PendingFriendRequestDTO, 0, len(friendships))
	for _, f := range friendships {
		requester, err := s.userRepo.GetByID(ctx, f.RequesterID)
		if err != nil {
			continue
		}
		result = append(result, &dto.PendingFriendRequestDTO{
			ID:          f.ID,
			Requester:   *requester.ToDTO(),
			AddresseeID: f.AddresseeID,
			Status:      f.Status.String(),
			UpdatedAt:   f.UpdatedAt,
		})
	}
	return result, nil
}
