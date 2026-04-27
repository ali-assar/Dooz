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

type FriendController struct {
	friendService service.FriendService
	logger        *slog.Logger
	t             tx.Transaction
}

func NewFriendController(friendService service.FriendService, logger *slog.Logger, t tx.Transaction) *FriendController {
	return &FriendController{
		friendService: friendService,
		logger:        logger.With("layer", "FriendController"),
		t:             t,
	}
}

// GetFriends returns accepted friends list.
//
//	@Summary	Get friends
//	@Tags		friends
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response
//	@Router		/friends [get]
func (c *FriendController) GetFriends(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	friends, err := c.friendService.GetFriends(reqCtx, userIDStr)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, friends)
}

// GetPendingRequests returns pending friend requests for the user.
//
//	@Summary	Get pending friend requests
//	@Tags		friends
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response
//	@Router		/friends/pending [get]
func (c *FriendController) GetPendingRequests(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	pending, err := c.friendService.GetPendingRequests(reqCtx, userIDStr)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, pending)
}

// SendRequest sends a friend request.
//
//	@Summary	Send friend request
//	@Tags		friends
//	@Security	BearerAuth
//	@Router		/friends/request [post]
func (c *FriendController) SendRequest(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	var body struct {
		AddresseeID string `json:"addressee_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	var f interface{}
	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		var err error
		f, err = c.friendService.SendRequest(txCtx, userIDStr, body.AddresseeID)
		return err
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 201, f)
}

// AcceptRequest accepts a friend request.
//
//	@Summary	Accept friend request
//	@Tags		friends
//	@Security	BearerAuth
//	@Param		id	path	string	true	"Friendship ID"
//	@Router		/friends/{id}/accept [patch]
func (c *FriendController) AcceptRequest(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	friendshipID := ctx.Param("id")

	var f interface{}
	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		var err error
		f, err = c.friendService.AcceptRequest(txCtx, friendshipID, userIDStr)
		return err
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, f)
}

// RejectRequest rejects a friend request.
//
//	@Summary	Reject friend request
//	@Tags		friends
//	@Security	BearerAuth
//	@Param		id	path	string	true	"Friendship ID"
//	@Router		/friends/{id}/reject [patch]
func (c *FriendController) RejectRequest(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	friendshipID := ctx.Param("id")

	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.friendService.RejectRequest(txCtx, friendshipID, userIDStr)
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.Success(ctx, 200, "Request rejected", nil)
}

// RemoveFriend removes a friend.
//
//	@Summary	Remove friend
//	@Tags		friends
//	@Security	BearerAuth
//	@Param		id	path	string	true	"Friendship ID"
//	@Router		/friends/{id} [delete]
func (c *FriendController) RemoveFriend(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	friendshipID := ctx.Param("id")

	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.friendService.RemoveFriend(txCtx, friendshipID, userIDStr)
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.Success(ctx, 200, response.MsgDeleted, nil)
}
