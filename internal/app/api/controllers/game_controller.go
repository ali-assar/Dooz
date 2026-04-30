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
	challengeService   service.GameChallengeService
	matchmakingService service.MatchmakingService
	logger             *slog.Logger
	t                  tx.Transaction
}

func NewGameController(
	gameService service.GameService,
	challengeService service.GameChallengeService,
	matchmakingService service.MatchmakingService,
	logger *slog.Logger,
	t tx.Transaction,
) *GameController {
	return &GameController{
		gameService:        gameService,
		challengeService:   challengeService,
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

	lg := c.logger.With(
		"method", "FindMatch",
		"userID", userIDStr,
		"path", ctx.FullPath(),
	)
	lg.Info("find-match request received")

	// Use a longer context for matchmaking (up to 25s)
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.MatchmakingTimeout+5*constants.MatchmakingPollInterval)
	defer cancel()

	result, err := c.matchmakingService.FindMatch(reqCtx, userIDStr)
	if err != nil {
		lg.Error("find-match failed", "error", err)
		response.InternalServerError(ctx, "Matchmaking failed")
		return
	}

	lg.Info(
		"find-match response ready",
		"boardID", result.BoardID,
		"isBotGame", result.IsBotGame,
		"symbol", result.Symbol,
	)

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

	var uri dto.MakeMoveURI
	if err := ctx.ShouldBindUri(&uri); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	var req dto.MakeMoveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}
	if req.Position == nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	state, err := c.gameService.MakeMove(reqCtx, uri.BoardID, userIDStr, *req.Position)
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

// CreateChallenge sends a game challenge to a friend.
//
//	@Summary	Create game challenge
//	@Tags		game
//	@Security	BearerAuth
//	@Param		body	body		dto.CreateChallengeRequest	true	"Challenge target (id or user_code)"
//	@Success	201		{object}	response.Response{data=dto.GameChallengeDTO}
//	@Router		/game/challenges [post]
func (c *GameController) CreateChallenge(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()
	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	var req dto.CreateChallengeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	var challenge *dto.GameChallengeDTO
	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		var err error
		challenge, err = c.challengeService.CreateChallenge(txCtx, userIDStr, req.AddresseeID, req.AddresseeCode)
		return err
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SuccessWithData(ctx, 201, challenge)
}

// GetPendingChallenges lists incoming pending challenges.
//
//	@Summary	List pending challenges
//	@Tags		game
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response{data=[]dto.PendingChallengeDTO}
//	@Router		/game/challenges/pending [get]
func (c *GameController) GetPendingChallenges(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()
	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}
	challenges, err := c.challengeService.GetPendingChallenges(reqCtx, userIDStr)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SuccessWithData(ctx, 200, challenges)
}

// AcceptChallenge accepts a challenge and creates a board.
//
//	@Summary	Accept challenge
//	@Tags		game
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Challenge ID"
//	@Success	200	{object}	response.Response{data=dto.FindMatchResponse}
//	@Router		/game/challenges/{id}/accept [patch]
func (c *GameController) AcceptChallenge(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()
	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}
	challengeID := ctx.Param("id")
	var result *dto.FindMatchResponse
	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		var err error
		result, err = c.challengeService.AcceptChallenge(txCtx, challengeID, userIDStr)
		return err
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SuccessWithData(ctx, 200, result)
}

// RejectChallenge rejects a pending challenge.
//
//	@Summary	Reject challenge
//	@Tags		game
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Challenge ID"
//	@Success	200	{object}	response.Response
//	@Router		/game/challenges/{id}/reject [patch]
func (c *GameController) RejectChallenge(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()
	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}
	challengeID := ctx.Param("id")
	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.challengeService.RejectChallenge(txCtx, challengeID, userIDStr)
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.Success(ctx, 200, "Challenge rejected", nil)
}

// CancelChallenge cancels a pending challenge by requester.
//
//	@Summary	Cancel challenge
//	@Tags		game
//	@Security	BearerAuth
//	@Param		id	path		string	true	"Challenge ID"
//	@Success	200	{object}	response.Response
//	@Router		/game/challenges/{id}/cancel [patch]
func (c *GameController) CancelChallenge(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()
	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}
	challengeID := ctx.Param("id")
	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.challengeService.CancelChallenge(txCtx, challengeID, userIDStr)
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.Success(ctx, 200, "Challenge canceled", nil)
}

// suppress unused import
var _ = strconv.Itoa
