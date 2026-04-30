package routes

import (
	"dooz/internal/app/api/controllers"
	"dooz/internal/app/api/middleware"

	"github.com/gin-gonic/gin"
)

type UserRoutes struct {
	userController   *controllers.UserController
	authMiddleware   *middleware.AuthMiddleware
	casbinMiddleware *middleware.CasbinMiddleware
}

func NewUserRoutes(
	userController *controllers.UserController,
	authMiddleware *middleware.AuthMiddleware,
	casbinMiddleware *middleware.CasbinMiddleware,
) *UserRoutes {
	return &UserRoutes{
		userController:   userController,
		authMiddleware:   authMiddleware,
		casbinMiddleware: casbinMiddleware,
	}
}

func (r *UserRoutes) SetupRoutes(router *gin.RouterGroup) {
	admin := router.Group("")
	admin.Use(r.authMiddleware.RequireAuth())
	admin.Use(r.casbinMiddleware.Enforce())
	{
		admin.GET("", r.userController.GetAllUsers)
	}

	auth := router.Group("")
	auth.Use(r.authMiddleware.RequireAuth())
	{
		auth.GET("/me", r.userController.GetMe)
		auth.PATCH("/me", r.userController.UpdateUser)
		auth.PUT("/me/change-password", r.userController.ChangePassword)
		auth.GET("/by-code/:code", r.userController.GetUserByCode)
		auth.GET("/:id", r.userController.GetUserByID)
	}
}
