package routes

import (
	"dooz/internal/app/api/controllers"
	"dooz/internal/app/api/middleware"

	"github.com/gin-gonic/gin"
)

type ChatRoutes struct {
	chatController *controllers.ChatController
	authMiddleware *middleware.AuthMiddleware
}

func NewChatRoutes(
	chatController *controllers.ChatController,
	authMiddleware *middleware.AuthMiddleware,
) *ChatRoutes {
	return &ChatRoutes{
		chatController: chatController,
		authMiddleware: authMiddleware,
	}
}

func (r *ChatRoutes) SetupRoutes(router *gin.RouterGroup) {
	router.Use(r.authMiddleware.RequireAuth())
	{
		router.POST("/send", r.chatController.SendDM)
		router.POST("/game", r.chatController.SendGameChat)
		router.GET("/history/:user_id", r.chatController.GetDMHistory)
		router.GET("/game/:board_id", r.chatController.GetGameChat)
	}
}
