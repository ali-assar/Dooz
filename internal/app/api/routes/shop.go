package routes

import (
	"dooz/internal/app/api/controllers"
	"dooz/internal/app/api/middleware"

	"github.com/gin-gonic/gin"
)

type ShopRoutes struct {
	shopController *controllers.ShopController
	authMiddleware *middleware.AuthMiddleware
}

func NewShopRoutes(
	shopController *controllers.ShopController,
	authMiddleware *middleware.AuthMiddleware,
) *ShopRoutes {
	return &ShopRoutes{
		shopController: shopController,
		authMiddleware: authMiddleware,
	}
}

func (r *ShopRoutes) SetupRoutes(router *gin.RouterGroup) {
	auth := router.Group("")
	auth.Use(r.authMiddleware.RequireAuth())
	{
		auth.GET("/items", r.shopController.GetItems)
		auth.GET("/inventory", r.shopController.GetMyInventory)
		auth.POST("/purchase", r.shopController.Purchase)
		auth.PATCH("/current-style", r.shopController.UpdateCurrentStyle)
		auth.POST("/wallet/add", r.shopController.AddWallet)
	}
}
