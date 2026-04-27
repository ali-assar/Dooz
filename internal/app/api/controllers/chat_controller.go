package controllers

import (
	"context"
	"log/slog"

	"dooz/internal/app/api/dto"
	"dooz/internal/app/api/response"
	"dooz/internal/constants"
	"dooz/internal/repository/tx"
	"dooz/internal/service"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	chatService service.ChatService
	logger      *slog.Logger
	t           tx.Transaction
}

func NewChatController(chatService service.ChatService, logger *slog.Logger, t tx.Transaction) *ChatController {
	return &ChatController{
		chatService: chatService,
		logger:      logger.With("layer", "ChatController"),
		t:           t,
	}
}

// SendDM sends a direct message.
//
//	@Summary	Send DM
//	@Tags		chat
//	@Security	BearerAuth
//	@Param		body	body		dto.SendDMRequest	true	"Message"
//	@Success	201		{object}	response.Response
//	@Router		/chat/send [post]
func (c *ChatController) SendDM(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	var req dto.SendDMRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	msg, err := c.chatService.SendDM(reqCtx, userIDStr, req.ReceiverID, req.Content)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 201, msg)
}

// SendGameChat sends a message in a game.
//
//	@Summary	Send game chat
//	@Tags		chat
//	@Security	BearerAuth
//	@Param		body	body		dto.SendGameChatRequest	true	"Message"
//	@Success	201		{object}	response.Response
//	@Router		/chat/game [post]
func (c *ChatController) SendGameChat(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	var req dto.SendGameChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	msg, err := c.chatService.SendGameChat(reqCtx, userIDStr, req.BoardID, req.Content)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 201, msg)
}

// GetDMHistory returns DM history with another user.
//
//	@Summary	DM history
//	@Tags		chat
//	@Security	BearerAuth
//	@Param		user_id	path		string	true	"Other user ID"
//	@Success	200		{object}	response.Response
//	@Router		/chat/history/{user_id} [get]
func (c *ChatController) GetDMHistory(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	var req dto.ChatHistoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}
	if req.Limit == 0 {
		req.Limit = 50
	}

	messages, err := c.chatService.GetDMHistory(reqCtx, userIDStr, req.UserID, req.Limit, req.Before)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, messages)
}

// GetGameChat returns in-game chat history.
//
//	@Summary	Game chat history
//	@Tags		chat
//	@Security	BearerAuth
//	@Param		board_id	path		string	true	"Board ID"
//	@Success	200			{object}	response.Response
//	@Router		/chat/game/{board_id} [get]
func (c *ChatController) GetGameChat(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	boardID := ctx.Param("board_id")
	if boardID == "" {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	messages, err := c.chatService.GetGameChatHistory(reqCtx, boardID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, messages)
}
