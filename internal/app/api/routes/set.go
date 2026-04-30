package routes

import (
	"github.com/google/wire"
)

var RouteSet = wire.NewSet(
	NewAuthRoutes,
	NewUserRoutes,
	NewGameRoutes,
	NewWSRoutes,
	NewFriendRoutes,
	NewChatRoutes,
	NewLeaderboardRoutes,
	NewAchievementRoutes,
	NewShopRoutes,
	NewCronRoutes,
	NewRouter,
)
