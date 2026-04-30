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

type ShopController struct {
	shopService service.ShopService
	logger      *slog.Logger
	t           tx.Transaction
}

func NewShopController(shopService service.ShopService, logger *slog.Logger, t tx.Transaction) *ShopController {
	return &ShopController{
		shopService: shopService,
		logger:      logger.With("layer", "ShopController"),
		t:           t,
	}
}

// GetItems returns all available shop items.
//
//	@Summary	Get shop items
//	@Tags		shop
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response{data=[]dto.StoreItemDTO}
//	@Router		/shop/items [get]
func (c *ShopController) GetItems(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	items, err := c.shopService.GetItems(reqCtx)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SuccessWithData(ctx, 200, items)
}

// GetMyInventory returns the authenticated user's owned items.
//
//	@Summary	Get my inventory
//	@Tags		shop
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response{data=dto.InventoryDTO}
//	@Router		/shop/inventory [get]
func (c *ShopController) GetMyInventory(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userID, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}
	inventory, err := c.shopService.GetInventory(reqCtx, userID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	response.SuccessWithData(ctx, 200, inventory)
}

// Purchase buys an item for the authenticated user.
//
//	@Summary	Purchase item
//	@Tags		shop
//	@Security	BearerAuth
//	@Param		body	body		dto.PurchaseItemRequest	true	"Item to purchase"
//	@Success	200		{object}	response.Response
//	@Router		/shop/purchase [post]
func (c *ShopController) Purchase(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userID, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}
	var req dto.PurchaseItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	if err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.shopService.Purchase(txCtx, userID, byte(req.ItemType), req.ItemValue)
	}); err != nil {
		_ = ctx.Error(err)
		return
	}
	response.Success(ctx, 200, "Item purchased successfully", nil)
}

// UpdateCurrentStyle sets current theme/xo/avatar for authenticated user.
//
//	@Summary	Update current style
//	@Tags		shop
//	@Security	BearerAuth
//	@Param		body	body		dto.UpdateCurrentStyleRequest	true	"Current style update"
//	@Success	200		{object}	response.Response
//	@Router		/shop/current-style [patch]
func (c *ShopController) UpdateCurrentStyle(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userID, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}
	var req dto.UpdateCurrentStyleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	if err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.shopService.UpdateCurrentStyle(txCtx, userID, &req)
	}); err != nil {
		_ = ctx.Error(err)
		return
	}
	response.Success(ctx, 200, "Current style updated successfully", nil)
}

// AddWallet adds coins/gems to authenticated user's wallet.
//
//	@Summary	Add wallet balance
//	@Tags		shop
//	@Security	BearerAuth
//	@Param		body	body		dto.AddWalletRequest	true	"Coins and gems to add"
//	@Success	200		{object}	response.Response
//	@Router		/shop/wallet/add [post]
func (c *ShopController) AddWallet(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userID, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}
	var req dto.AddWalletRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	if err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.shopService.AddWallet(txCtx, userID, req.Coins, req.Gems)
	}); err != nil {
		_ = ctx.Error(err)
		return
	}
	response.Success(ctx, 200, "Wallet updated successfully", nil)
}
