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

type UserController struct {
	userService service.UserService
	logger      *slog.Logger
	t           tx.Transaction
}

func NewUserController(userService service.UserService, logger *slog.Logger, t tx.Transaction) *UserController {
	return &UserController{
		userService: userService,
		logger:      logger.With("layer", "UserController"),
		t:           t,
	}
}

// GetMe returns the authenticated user's profile.
//
//	@Summary	Get my profile
//	@Tags		users
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response{data=dto.UserDTO}
//	@Router		/users/me [get]
func (c *UserController) GetMe(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	user, err := c.userService.GetUserByID(reqCtx, userIDStr)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, user)
}

// GetUserByID returns a user by ID.
//
//	@Summary	Get user by ID
//	@Tags		users
//	@Security	BearerAuth
//	@Param		id	path		string	true	"User ID"
//	@Success	200	{object}	response.Response{data=dto.UserDTO}
//	@Router		/users/{id} [get]
func (c *UserController) GetUserByID(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	id := ctx.Param("id")
	if id == "" {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	user, err := c.userService.GetUserByID(reqCtx, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, user)
}

// GetAllUsers returns paginated list of users (admin only).
//
//	@Summary	List all users
//	@Tags		users
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response
//	@Router		/users [get]
func (c *UserController) GetAllUsers(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	var paginationReq dto.CursorPaginationRequest
	if err := ctx.ShouldBindQuery(&paginationReq); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	req := paginationReq.ToPaginationRequest()

	result, err := c.userService.GetAllUsers(reqCtx, req)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, gin.H{
		"users":      result.Data,
		"pagination": result.Pagination,
	})
}

// UpdateUser updates user profile.
//
//	@Summary	Update profile
//	@Tags		users
//	@Security	BearerAuth
//	@Param		body	body		dto.UpdateUserRequest	true	"Update data"
//	@Success	200		{object}	response.Response{data=dto.UserDTO}
//	@Router		/users/me [patch]
func (c *UserController) UpdateUser(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	authenticatedUserID, role, ok := getUserIDAndRoleFromContext(ctx, c.logger)
	if !ok {
		return
	}

	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	targetUserID := authenticatedUserID
	if isAdmin(role) && req.UserID != "" {
		targetUserID = req.UserID
	}

	if !isAdmin(role) {
		if req.UserID != "" || req.Role != "" || req.Phone != "" || req.Email != "" {
			response.Forbidden(ctx, "Cannot change restricted fields")
			return
		}
	}

	var user *dto.UserDTO
	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		var err error
		user, err = c.userService.UpdateUser(txCtx, targetUserID, &req, isAdmin(role))
		return err
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, user)
}

// ChangePassword changes the authenticated user's password.
//
//	@Summary	Change password
//	@Tags		users
//	@Security	BearerAuth
//	@Param		body	body		dto.ChangePasswordRequest	true	"Password data"
//	@Success	200		{object}	response.Response
//	@Router		/users/me/change-password [put]
func (c *UserController) ChangePassword(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	var req dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		response.ValidationError(ctx, "Passwords do not match")
		return
	}

	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.userService.ChangePassword(txCtx, userIDStr, req.NewPassword)
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.Success(ctx, 200, "Password changed successfully", nil)
}
