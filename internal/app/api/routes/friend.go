package routes

import (
	"dooz/internal/app/api/controllers"
	"dooz/internal/app/api/middleware"

	"github.com/gin-gonic/gin"
)

type FriendRoutes struct {
	friendController *controllers.FriendController
	authMiddleware   *middleware.AuthMiddleware
}

func NewFriendRoutes(
	friendController *controllers.FriendController,
	authMiddleware *middleware.AuthMiddleware,
) *FriendRoutes {
	return &FriendRoutes{
		friendController: friendController,
		authMiddleware:   authMiddleware,
	}
}

func (r *FriendRoutes) SetupRoutes(router *gin.RouterGroup) {
	router.Use(r.authMiddleware.RequireAuth())
	{
		router.GET("", r.friendController.GetFriends)
		router.GET("/pending", r.friendController.GetPendingRequests)
		router.POST("/request", r.friendController.SendRequest)
		router.PATCH("/:id/accept", r.friendController.AcceptRequest)
		router.PATCH("/:id/reject", r.friendController.RejectRequest)
		router.DELETE("/:id", r.friendController.RemoveFriend)
	}
}
