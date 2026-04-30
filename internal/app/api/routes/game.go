package routes

import (
	"dooz/internal/app/api/controllers"
	"dooz/internal/app/api/middleware"

	"github.com/gin-gonic/gin"
)

type GameRoutes struct {
	gameController *controllers.GameController
	wsController   *controllers.WSController
	authMiddleware *middleware.AuthMiddleware
}

func NewGameRoutes(
	gameController *controllers.GameController,
	wsController *controllers.WSController,
	authMiddleware *middleware.AuthMiddleware,
) *GameRoutes {
	return &GameRoutes{
		gameController: gameController,
		wsController:   wsController,
		authMiddleware: authMiddleware,
	}
}

func (r *GameRoutes) SetupRoutes(router *gin.RouterGroup) {
	auth := router.Group("")
	auth.Use(r.authMiddleware.RequireAuth())
	{
		auth.POST("/find-match", r.gameController.FindMatch)
		auth.POST("/challenges", r.gameController.CreateChallenge)
		auth.GET("/challenges/pending", r.gameController.GetPendingChallenges)
		auth.PATCH("/challenges/:id/accept", r.gameController.AcceptChallenge)
		auth.PATCH("/challenges/:id/reject", r.gameController.RejectChallenge)
		auth.PATCH("/challenges/:id/cancel", r.gameController.CancelChallenge)
		auth.GET("/history", r.gameController.GetHistory)
		auth.GET("/:id", r.gameController.GetGameState)
		auth.POST("/:id/move", r.gameController.MakeMove)
		auth.POST("/:id/resign", r.gameController.Resign)
	}
}

type WSRoutes struct {
	wsController   *controllers.WSController
	authMiddleware *middleware.AuthMiddleware
}

func NewWSRoutes(wsController *controllers.WSController, authMiddleware *middleware.AuthMiddleware) *WSRoutes {
	return &WSRoutes{wsController: wsController, authMiddleware: authMiddleware}
}

func (r *WSRoutes) SetupRoutes(router *gin.RouterGroup) {
	router.Use(r.authMiddleware.RequireAuth())
	router.GET("", r.wsController.HandleWS)
}
