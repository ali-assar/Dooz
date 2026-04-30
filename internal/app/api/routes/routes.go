package routes

import (
	"log/slog"

	"dooz/internal/app/api/middleware"

	"github.com/gin-gonic/gin"
)

type Router struct {
	authRoutes        *AuthRoutes
	userRoutes        *UserRoutes
	gameRoutes        *GameRoutes
	wsRoutes          *WSRoutes
	friendRoutes      *FriendRoutes
	chatRoutes        *ChatRoutes
	leaderboardRoutes *LeaderboardRoutes
	achievementRoutes *AchievementRoutes
	shopRoutes        *ShopRoutes
	cronRoutes        *CronRoutes
	corsMiddleware    *middleware.CORSMiddleware
	logger            *slog.Logger
}

func NewRouter(
	authRoutes *AuthRoutes,
	userRoutes *UserRoutes,
	gameRoutes *GameRoutes,
	wsRoutes *WSRoutes,
	friendRoutes *FriendRoutes,
	chatRoutes *ChatRoutes,
	leaderboardRoutes *LeaderboardRoutes,
	achievementRoutes *AchievementRoutes,
	shopRoutes *ShopRoutes,
	cronRoutes *CronRoutes,
	corsMiddleware *middleware.CORSMiddleware,
	logger *slog.Logger,
) *Router {
	return &Router{
		authRoutes:        authRoutes,
		userRoutes:        userRoutes,
		gameRoutes:        gameRoutes,
		wsRoutes:          wsRoutes,
		friendRoutes:      friendRoutes,
		chatRoutes:        chatRoutes,
		leaderboardRoutes: leaderboardRoutes,
		achievementRoutes: achievementRoutes,
		shopRoutes:        shopRoutes,
		cronRoutes:        cronRoutes,
		corsMiddleware:    corsMiddleware,
		logger:            logger,
	}
}

func (r *Router) SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(r.corsMiddleware.Handle())
	router.Use(middleware.ErrorHandler(r.logger))

	apiV1 := router.Group("/api/v1")
	{
		r.authRoutes.SetupRoutes(apiV1.Group("/auth"))
		r.userRoutes.SetupRoutes(apiV1.Group("/users"))
		r.gameRoutes.SetupRoutes(apiV1.Group("/game"))
		r.wsRoutes.SetupRoutes(apiV1.Group("/ws"))
		r.friendRoutes.SetupRoutes(apiV1.Group("/friends"))
		r.chatRoutes.SetupRoutes(apiV1.Group("/chat"))
		r.leaderboardRoutes.SetupRoutes(apiV1.Group("/leaderboard"))
		r.achievementRoutes.SetupRoutes(apiV1.Group("/achievements"))
		r.shopRoutes.SetupRoutes(apiV1.Group("/shop"))
		r.cronRoutes.SetupRoutes(apiV1.Group("/cron"))
	}

	return router
}
