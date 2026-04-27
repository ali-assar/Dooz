package controllers

import (
	"context"
	"log/slog"

	"dooz/internal/app/api/response"
	"dooz/internal/constants"
	"dooz/internal/service"

	"github.com/gin-gonic/gin"
)

type LeaderboardController struct {
	leaderboardService service.LeaderboardService
	logger             *slog.Logger
}

func NewLeaderboardController(leaderboardService service.LeaderboardService, logger *slog.Logger) *LeaderboardController {
	return &LeaderboardController{
		leaderboardService: leaderboardService,
		logger:             logger.With("layer", "LeaderboardController"),
	}
}

// GetGlobal returns the global leaderboard.
//
//	@Summary	Global leaderboard
//	@Tags		leaderboard
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response
//	@Router		/leaderboard/global [get]
func (c *LeaderboardController) GetGlobal(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	entries, err := c.leaderboardService.GetGlobalLeaderboard(reqCtx, 50)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, entries)
}

// GetFriends returns the friends leaderboard.
//
//	@Summary	Friends leaderboard
//	@Tags		leaderboard
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response
//	@Router		/leaderboard/friends [get]
func (c *LeaderboardController) GetFriends(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	entries, err := c.leaderboardService.GetFriendsLeaderboard(reqCtx, userIDStr)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, entries)
}
