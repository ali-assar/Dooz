package routes

import (
	"dooz/internal/app/api/controllers"
	"dooz/internal/app/api/middleware"

	"github.com/gin-gonic/gin"
)

type LeaderboardRoutes struct {
	leaderboardController *controllers.LeaderboardController
	authMiddleware        *middleware.AuthMiddleware
}

func NewLeaderboardRoutes(
	leaderboardController *controllers.LeaderboardController,
	authMiddleware *middleware.AuthMiddleware,
) *LeaderboardRoutes {
	return &LeaderboardRoutes{
		leaderboardController: leaderboardController,
		authMiddleware:        authMiddleware,
	}
}

func (r *LeaderboardRoutes) SetupRoutes(router *gin.RouterGroup) {
	router.Use(r.authMiddleware.RequireAuth())
	{
		router.GET("/global", r.leaderboardController.GetGlobal)
		router.GET("/friends", r.leaderboardController.GetFriends)
	}
}
