package repository

import (
	achievementRepo "dooz/internal/repository/achievement"
	boardRepo "dooz/internal/repository/board"
	chatRepo "dooz/internal/repository/chat_message"
	friendshipRepo "dooz/internal/repository/friendship"
	gameChallengeRepo "dooz/internal/repository/game_challenge"
	moveRepo "dooz/internal/repository/move"
	otpRepo "dooz/internal/repository/otp"
	sessionRepo "dooz/internal/repository/session"
	storeItemRepo "dooz/internal/repository/store_item"
	"dooz/internal/repository/tx"
	userRepo "dooz/internal/repository/user"
	userAchievementRepo "dooz/internal/repository/user_achievement"
	userItemRepo "dooz/internal/repository/user_item"

	"github.com/google/wire"
)

var RepositorySet = wire.NewSet(
	tx.NewTransaction,
	userRepo.New,
	sessionRepo.New,
	otpRepo.New,
	boardRepo.New,
	moveRepo.New,
	gameChallengeRepo.New,
	friendshipRepo.New,
	chatRepo.New,
	achievementRepo.New,
	userAchievementRepo.New,
	storeItemRepo.New,
	userItemRepo.New,
)
