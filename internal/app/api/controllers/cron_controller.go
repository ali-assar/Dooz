package controllers

import (
	"context"
	"log/slog"

	"dooz/internal/app/api/response"
	"dooz/internal/constants"
	"dooz/internal/repository/tx"
	"dooz/internal/service"

	"github.com/gin-gonic/gin"
)

type CronController struct {
	otpOutboxService service.OTPOutboxService
	logger           *slog.Logger
	t                tx.Transaction
}

func NewCronController(otpOutboxService service.OTPOutboxService, logger *slog.Logger, t tx.Transaction) *CronController {
	return &CronController{
		otpOutboxService: otpOutboxService,
		logger:           logger.With("layer", "CronController"),
		t:                t,
	}
}

// ProcessOTPOutbox processes pending OTPs.
//
//	@Summary	Process OTP outbox
//	@Tags		cron
//	@Success	200	{object}	response.Response
//	@Router		/cron/process-otp-outbox [get]
func (c *CronController) ProcessOTPOutbox(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.CronRequestTimeout)
	defer cancel()

	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.otpOutboxService.ProcessPendingOTPs(txCtx)
	})
	if err != nil {
		c.logger.Error("failed to process OTPs", "error", err)
		response.InternalServerError(ctx, response.ErrInternalServer)
		return
	}

	response.Success(ctx, 200, "OTP outbox processed", nil)
}

// DeleteExpiredOTPs deletes expired OTPs.
//
//	@Summary	Delete expired OTPs
//	@Tags		cron
//	@Success	200	{object}	response.Response
//	@Router		/cron/delete-expired-otps [get]
func (c *CronController) DeleteExpiredOTPs(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.CronDeleteTimeout)
	defer cancel()

	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.otpOutboxService.DeleteExpiredOTPs(txCtx)
	})
	if err != nil {
		response.InternalServerError(ctx, response.ErrInternalServer)
		return
	}

	response.Success(ctx, 200, "Expired OTPs deleted", nil)
}
