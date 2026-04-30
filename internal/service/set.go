package service

import (
	"github.com/google/wire"
)

var ServiceSet = wire.NewSet(
	NewAuthService,
	NewUserService,
	NewGameService,
	NewGameChallengeService,
	NewMatchmakingService,
	NewFriendService,
	NewChatService,
	NewLeaderboardService,
	NewAchievementService,
	NewShopService,
	NewOTPOutboxServiceWithDefaults,
	NewTurnstileService,
)
