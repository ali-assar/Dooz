package routes

import (
	"dooz/internal/app/api/controllers"

	"github.com/gin-gonic/gin"
)

type CronRoutes struct {
	cronController *controllers.CronController
}

func NewCronRoutes(cronController *controllers.CronController) *CronRoutes {
	return &CronRoutes{cronController: cronController}
}

func (r *CronRoutes) SetupRoutes(router *gin.RouterGroup) {
	router.GET("/process-otp-outbox", r.cronController.ProcessOTPOutbox)
	router.GET("/delete-expired-otps", r.cronController.DeleteExpiredOTPs)
}
