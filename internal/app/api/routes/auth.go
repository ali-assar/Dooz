package routes

import (
	"dooz/internal/app/api/controllers"
	"dooz/internal/app/api/middleware"

	"github.com/gin-gonic/gin"
)

type AuthRoutes struct {
	authController      *controllers.AuthController
	authMiddleware      *middleware.AuthMiddleware
	casbinMiddleware    *middleware.CasbinMiddleware
	turnstileMiddleware *middleware.TurnstileMiddleware
}

func NewAuthRoutes(
	authController *controllers.AuthController,
	authMiddleware *middleware.AuthMiddleware,
	casbinMiddleware *middleware.CasbinMiddleware,
	turnstileMiddleware *middleware.TurnstileMiddleware,
) *AuthRoutes {
	return &AuthRoutes{
		authController:      authController,
		authMiddleware:      authMiddleware,
		casbinMiddleware:    casbinMiddleware,
		turnstileMiddleware: turnstileMiddleware,
	}
}

func (r *AuthRoutes) SetupRoutes(router *gin.RouterGroup) {
	turnstileProtected := router.Group("")
	turnstileProtected.Use(r.turnstileMiddleware.Verify())
	{
		turnstileProtected.POST("/login", r.authController.Login)
		turnstileProtected.POST("/register", r.authController.Register)
		turnstileProtected.POST("/password-reset/request", r.authController.RequestPasswordReset)
		turnstileProtected.POST("/password-reset/reset", r.authController.ResetPassword)
	}

	router.POST("/register/verify-otp", r.authController.VerifyRegistrationOTP)
	router.POST("/register/set-password", r.authController.SetPassword)
	router.POST("/refresh", r.authController.RefreshToken)

	protected := router.Group("")
	protected.Use(r.authMiddleware.RequireAuth())
	protected.Use(r.casbinMiddleware.Enforce())
	{
		protected.POST("/logout", r.authController.Logout)
	}
}
