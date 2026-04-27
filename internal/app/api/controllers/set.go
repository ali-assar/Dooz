package controllers

import (
	"github.com/google/wire"
)

var ControllerSet = wire.NewSet(
	NewAuthController,
	NewUserController,
	NewGameController,
	NewFriendController,
	NewChatController,
	NewLeaderboardController,
	NewAchievementController,
	NewWSController,
	NewCronController,
)
