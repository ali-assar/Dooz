package controllers

import (
	"context"
	"log/slog"

	"dooz/internal/app/api/response"
	"dooz/internal/constants"
	"dooz/internal/service"

	"github.com/gin-gonic/gin"
)

type AchievementController struct {
	achievementService service.AchievementService
	logger             *slog.Logger
}

func NewAchievementController(achievementService service.AchievementService, logger *slog.Logger) *AchievementController {
	return &AchievementController{
		achievementService: achievementService,
		logger:             logger.With("layer", "AchievementController"),
	}
}

// GetAll returns all achievements.
//
//	@Summary	List achievements
//	@Tags		achievements
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response
//	@Router		/achievements [get]
func (c *AchievementController) GetAll(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	achievements, err := c.achievementService.GetAll(reqCtx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, achievements)
}

// GetMine returns achievements earned by the authenticated user.
//
//	@Summary	My achievements
//	@Tags		achievements
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response
//	@Router		/achievements/mine [get]
func (c *AchievementController) GetMine(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	achievements, err := c.achievementService.GetUserAchievements(reqCtx, userIDStr)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, achievements)
}
