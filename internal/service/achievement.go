package service

import (
	"context"
	"log/slog"
	"time"

	"dooz/entity"
	"dooz/internal/app/api/dto"
	achievementRepo "dooz/internal/repository/achievement"
	friendshipRepo "dooz/internal/repository/friendship"
	userRepo "dooz/internal/repository/user"
	userAchievementRepo "dooz/internal/repository/user_achievement"
)

type AchievementService interface {
	GetAll(ctx context.Context) ([]*dto.AchievementDTO, error)
	GetUserAchievements(ctx context.Context, userID string) ([]*dto.UserAchievementWithDetailDTO, error)
	CheckAndAward(ctx context.Context, userID string) error
}

type achievementService struct {
	achievementRepo     achievementRepo.Repository
	userAchievementRepo userAchievementRepo.Repository
	userRepo            userRepo.Repository
	friendshipRepo      friendshipRepo.Repository
	logger              *slog.Logger
}

func NewAchievementService(
	achievementRepo achievementRepo.Repository,
	userAchievementRepo userAchievementRepo.Repository,
	userRepo userRepo.Repository,
	friendshipRepo friendshipRepo.Repository,
	logger *slog.Logger,
) AchievementService {
	return &achievementService{
		achievementRepo:     achievementRepo,
		userAchievementRepo: userAchievementRepo,
		userRepo:            userRepo,
		friendshipRepo:      friendshipRepo,
		logger:              logger.With("layer", "AchievementService"),
	}
}

func (s *achievementService) GetAll(ctx context.Context) ([]*dto.AchievementDTO, error) {
	all, err := s.achievementRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]*dto.AchievementDTO, len(all))
	for i, a := range all {
		dtos[i] = a.ToDTO()
	}
	return dtos, nil
}

func (s *achievementService) GetUserAchievements(ctx context.Context, userID string) ([]*dto.UserAchievementWithDetailDTO, error) {
	uas, err := s.userAchievementRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.UserAchievementWithDetailDTO, 0, len(uas))
	for _, ua := range uas {
		a, err := s.achievementRepo.GetByID(ctx, ua.AchievementID)
		if err != nil {
			continue
		}
		result = append(result, &dto.UserAchievementWithDetailDTO{
			Achievement: *a.ToDTO(),
			EarnedAt:    ua.EarnedAt,
		})
	}
	return result, nil
}

func (s *achievementService) CheckAndAward(ctx context.Context, userID string) error {
	lg := s.logger.With("method", "CheckAndAward", "userID", userID)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	friendCount, _ := s.friendshipRepo.CountAcceptedFriends(ctx, userID)

	all, err := s.achievementRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	for _, a := range all {
		earned, _ := s.userAchievementRepo.HasEarned(ctx, userID, a.ID)
		if earned {
			continue
		}

		var progress int
		switch a.RequirementType {
		case entity.RequirementWins:
			progress = user.Wins
		case entity.RequirementDraws:
			progress = user.Draws
		case entity.RequirementFriends:
			progress = int(friendCount)
		case entity.RequirementGamesPlayed:
			progress = user.Wins + user.Losses + user.Draws
		default:
			continue
		}

		if progress >= a.RequirementValue {
			ua := &entity.UserAchievement{
				UserID:        userID,
				AchievementID: a.ID,
				EarnedAt:      now,
			}
			if err := s.userAchievementRepo.Create(ctx, ua); err != nil {
				lg.Warn("failed to award achievement", "achievementID", a.ID, "error", err)
				continue
			}
			// Award coins/gems
			if a.CoinReward > 0 || a.GemReward > 0 {
				_ = s.userRepo.UpdateStats(ctx, userID, 0, 0, 0, 0, 0, a.CoinReward, a.GemReward)
			}
			lg.Info("achievement awarded", "achievement", a.Name)
		}
	}
	return nil
}
