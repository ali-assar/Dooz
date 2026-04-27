package service

import (
	"github.com/google/wire"
)

var ServiceSet = wire.NewSet(
	NewAuthService,
	NewUserService,
	NewGameService,
	NewMatchmakingService,
	NewFriendService,
	NewChatService,
	NewLeaderboardService,
	NewAchievementService,
	NewOTPOutboxServiceWithDefaults,
	NewTurnstileService,
)
