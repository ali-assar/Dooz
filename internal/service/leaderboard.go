package service

import (
	"context"
	"log/slog"

	"dooz/internal/app/api/dto"
	friendshipRepo "dooz/internal/repository/friendship"
	userRepo "dooz/internal/repository/user"
)

type LeaderboardService interface {
	GetGlobalLeaderboard(ctx context.Context, limit int) ([]*dto.LeaderboardEntryDTO, error)
	GetFriendsLeaderboard(ctx context.Context, userID string) ([]*dto.LeaderboardEntryDTO, error)
}

type leaderboardService struct {
	userRepo       userRepo.Repository
	friendshipRepo friendshipRepo.Repository
	logger         *slog.Logger
}

func NewLeaderboardService(
	userRepo userRepo.Repository,
	friendshipRepo friendshipRepo.Repository,
	logger *slog.Logger,
) LeaderboardService {
	return &leaderboardService{
		userRepo:       userRepo,
		friendshipRepo: friendshipRepo,
		logger:         logger.With("layer", "LeaderboardService"),
	}
}

func (s *leaderboardService) GetGlobalLeaderboard(ctx context.Context, limit int) ([]*dto.LeaderboardEntryDTO, error) {
	users, err := s.userRepo.GetAllWithCursor(ctx, nil, true, uint32(limit), nil)
	if err != nil {
		return nil, err
	}

	entries := make([]*dto.LeaderboardEntryDTO, 0, len(users))
	for i, u := range users {
		total := u.Wins + u.Losses + u.Draws
		winRate := 0.0
		if total > 0 {
			winRate = float64(u.Wins) / float64(total)
		}
		entries = append(entries, &dto.LeaderboardEntryDTO{
			Rank:       i + 1,
			User:       *u.ToDTO(),
			Wins:       u.Wins,
			Losses:     u.Losses,
			Draws:      u.Draws,
			WinRate:    winRate,
			TotalGames: total,
		})
	}
	return entries, nil
}

func (s *leaderboardService) GetFriendsLeaderboard(ctx context.Context, userID string) ([]*dto.LeaderboardEntryDTO, error) {
	friendships, err := s.friendshipRepo.GetFriendsOfUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	entries := make([]*dto.LeaderboardEntryDTO, 0, len(friendships)+1)

	// Always include self
	self, err := s.userRepo.GetByID(ctx, userID)
	if err == nil {
		total := self.Wins + self.Losses + self.Draws
		winRate := 0.0
		if total > 0 {
			winRate = float64(self.Wins) / float64(total)
		}
		entries = append(entries, &dto.LeaderboardEntryDTO{
			User:       *self.ToDTO(),
			Wins:       self.Wins,
			Losses:     self.Losses,
			Draws:      self.Draws,
			WinRate:    winRate,
			TotalGames: total,
		})
	}

	for _, f := range friendships {
		friendID := f.AddresseeID
		if f.RequesterID != userID {
			friendID = f.RequesterID
		}
		u, err := s.userRepo.GetByID(ctx, friendID)
		if err != nil {
			continue
		}
		total := u.Wins + u.Losses + u.Draws
		winRate := 0.0
		if total > 0 {
			winRate = float64(u.Wins) / float64(total)
		}
		entries = append(entries, &dto.LeaderboardEntryDTO{
			User:       *u.ToDTO(),
			Wins:       u.Wins,
			Losses:     u.Losses,
			Draws:      u.Draws,
			WinRate:    winRate,
			TotalGames: total,
		})
	}

	// Sort by wins descending and assign rank
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Wins > entries[i].Wins {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	for i := range entries {
		entries[i].Rank = i + 1
	}

	return entries, nil
}
