package routes

import (
	"dooz/internal/app/api/controllers"
	"dooz/internal/app/api/middleware"

	"github.com/gin-gonic/gin"
)

type AchievementRoutes struct {
	achievementController *controllers.AchievementController
	authMiddleware        *middleware.AuthMiddleware
}

func NewAchievementRoutes(
	achievementController *controllers.AchievementController,
	authMiddleware *middleware.AuthMiddleware,
) *AchievementRoutes {
	return &AchievementRoutes{
		achievementController: achievementController,
		authMiddleware:        authMiddleware,
	}
}

func (r *AchievementRoutes) SetupRoutes(router *gin.RouterGroup) {
	router.Use(r.authMiddleware.RequireAuth())
	{
		router.GET("", r.achievementController.GetAll)
		router.GET("/mine", r.achievementController.GetMine)
	}
}
