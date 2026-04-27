package controllers

import (
	"context"
	"log/slog"
	"strconv"

	"dooz/internal/app/api/dto"
	"dooz/internal/app/api/response"
	"dooz/internal/constants"
	"dooz/internal/repository/tx"
	"dooz/internal/service"

	"github.com/gin-gonic/gin"
)

type GameController struct {
	gameService        service.GameService
	matchmakingService service.MatchmakingService
	logger             *slog.Logger
	t                  tx.Transaction
}

func NewGameController(
	gameService service.GameService,
	matchmakingService service.MatchmakingService,
	logger *slog.Logger,
	t tx.Transaction,
) *GameController {
	return &GameController{
		gameService:        gameService,
		matchmakingService: matchmakingService,
		logger:             logger.With("layer", "GameController"),
		t:                  t,
	}
}

// FindMatch finds or creates a game match (waits up to 20s for a real opponent, then assigns bot).
//
//	@Summary	Find match
//	@Tags		game
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response{data=dto.FindMatchResponse}
//	@Router		/game/find-match [post]
func (c *GameController) FindMatch(ctx *gin.Context) {
	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	// Use a longer context for matchmaking (up to 25s)
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.MatchmakingTimeout+5*constants.MatchmakingPollInterval)
	defer cancel()

	result, err := c.matchmakingService.FindMatch(reqCtx, userIDStr)
	if err != nil {
		c.logger.Error("matchmaking failed", "error", err)
		response.InternalServerError(ctx, "Matchmaking failed")
		return
	}

	response.SuccessWithData(ctx, 200, result)
}

// GetGameState returns the current state of a game.
//
//	@Summary	Get game state
//	@Tags		game
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Board ID"
//	@Success	200	{object}	response.Response{data=dto.GameStateResponse}
//	@Router		/game/{id} [get]
func (c *GameController) GetGameState(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	boardID := ctx.Param("id")

	state, err := c.gameService.GetGameState(reqCtx, boardID, userIDStr)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, state)
}

// MakeMove makes a move in a game.
//
//	@Summary	Make move
//	@Tags		game
//	@Security	BearerAuth
//	@Param		id		path		string				true	"Board ID"
//	@Param		body	body		dto.MakeMoveRequest	true	"Move data"
//	@Success	200		{object}	response.Response{data=dto.GameStateResponse}
//	@Router		/game/{id}/move [post]
func (c *GameController) MakeMove(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	boardID := ctx.Param("id")

	var req dto.MakeMoveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	state, err := c.gameService.MakeMove(reqCtx, boardID, userIDStr, req.Position)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, state)
}

// Resign resigns from a game.
//
//	@Summary	Resign from game
//	@Tags		game
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Board ID"
//	@Success	200	{object}	response.Response
//	@Router		/game/{id}/resign [post]
func (c *GameController) Resign(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	boardID := ctx.Param("id")

	board, err := c.gameService.Resign(reqCtx, boardID, userIDStr)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, board)
}

// GetHistory returns the user's game history.
//
//	@Summary	Game history
//	@Tags		game
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response
//	@Router		/game/history [get]
func (c *GameController) GetHistory(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	boards, err := c.gameService.GetHistory(reqCtx, userIDStr)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, boards)
}

// suppress unused import
var _ = strconv.Itoa
