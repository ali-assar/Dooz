package controllers

import (
	"context"
	"log/slog"

	"dooz/entity"
	"dooz/internal/app/api/dto"
	"dooz/internal/app/api/response"
	"dooz/internal/constants"
	"dooz/internal/infrastructure/godotenv"
	"dooz/internal/repository/tx"
	"dooz/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService service.AuthService
	logger      *slog.Logger
	env         *godotenv.Env
	t           tx.Transaction
}

func NewAuthController(authService service.AuthService, logger *slog.Logger, env *godotenv.Env, t tx.Transaction) *AuthController {
	return &AuthController{
		authService: authService,
		logger:      logger.With("layer", "AuthController"),
		env:         env,
		t:           t,
	}
}

// Login authenticates user and returns JWT tokens.
//
//	@Summary	Login
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		dto.LoginRequest	true	"Login credentials"
//	@Success	200		{object}	response.Response{data=dto.LoginResponse}
//	@Router		/auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	deviceType := parseDeviceType(req.DeviceType)
	clientIP := ctx.ClientIP()

	var tokenPair *service.TokenPair
	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		var err error
		tokenPair, err = c.authService.Login(txCtx, req.Phone, req.Password, clientIP, deviceType, req.Remember)
		return err
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, c.tokenPairToResponse(tokenPair))
}

// Register initiates user registration.
//
//	@Summary	Register
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		dto.RegisterRequest	true	"Registration data"
//	@Success	200		{object}	response.Response
//	@Router		/auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.authService.Register(txCtx, req.Phone, req.Email, req.Fullname)
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.Success(ctx, 200, "Registration initiated, OTP will be sent", nil)
}

// VerifyRegistrationOTP verifies OTP for registration.
//
//	@Summary	Verify registration OTP
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		dto.VerifyRegistrationOTPRequest	true	"OTP data"
//	@Success	200		{object}	response.Response
//	@Router		/auth/register/verify-otp [post]
func (c *AuthController) VerifyRegistrationOTP(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	var req dto.VerifyRegistrationOTPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, err.Error())
		return
	}

	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.authService.VerifyPhoneOTP(txCtx, req.Phone, req.Code, entity.RegistrationPurpose)
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.Success(ctx, 200, "Phone verified", nil)
}

// SetPassword sets password after phone verification.
//
//	@Summary	Set password
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		dto.SetPasswordRequest	true	"Password data"
//	@Success	200		{object}	response.Response{data=dto.LoginResponse}
//	@Router		/auth/register/set-password [post]
func (c *AuthController) SetPassword(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	var req dto.SetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	deviceType := parseDeviceType(req.DeviceType)
	clientIP := ctx.ClientIP()

	var tokenPair *service.TokenPair
	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		var err error
		tokenPair, err = c.authService.SetPassword(txCtx, req.Phone, req.Password, clientIP, deviceType, req.Remember)
		return err
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, c.tokenPairToResponse(tokenPair))
}

// RefreshToken refreshes access token.
//
//	@Summary	Refresh token
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		dto.RefreshTokenRequest	true	"Refresh token"
//	@Success	200		{object}	response.Response{data=dto.LoginResponse}
//	@Router		/auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	var req dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	var tokenPair *service.TokenPair
	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		var err error
		tokenPair, err = c.authService.RefreshToken(txCtx, req.RefreshToken)
		return err
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, c.tokenPairToResponse(tokenPair))
}

// RequestPasswordReset sends OTP for password reset.
//
//	@Summary	Request password reset
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		dto.RequestPasswordResetRequest	true	"Phone number"
//	@Success	200		{object}	response.Response
//	@Router		/auth/password-reset/request [post]
func (c *AuthController) RequestPasswordReset(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	var req dto.RequestPasswordResetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	_ = c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.authService.RequestPasswordReset(txCtx, req.Phone)
	})

	response.Success(ctx, 200, "If the phone is registered, an OTP will be sent", nil)
}

// ResetPassword resets password using OTP.
//
//	@Summary	Reset password
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		dto.ResetPasswordRequest	true	"Reset data"
//	@Success	200		{object}	response.Response{data=dto.LoginResponse}
//	@Router		/auth/password-reset/reset [post]
func (c *AuthController) ResetPassword(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	var req dto.ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, response.ErrInvalidRequest)
		return
	}

	deviceType := parseDeviceType(req.DeviceType)
	clientIP := ctx.ClientIP()

	var tokenPair *service.TokenPair
	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		var err error
		tokenPair, err = c.authService.ResetPassword(txCtx, req.Phone, req.Code, req.NewPassword, clientIP, deviceType, req.Remember)
		return err
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.SuccessWithData(ctx, 200, c.tokenPairToResponse(tokenPair))
}

// Logout logs out user.
//
//	@Summary	Logout
//	@Tags		auth
//	@Security	BearerAuth
//	@Success	200	{object}	response.Response
//	@Router		/auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), constants.DefaultRequestTimeout)
	defer cancel()

	userIDStr, ok := getUserIDFromContext(ctx, c.logger)
	if !ok {
		return
	}

	deviceTypeStr := ctx.Query("device_type")
	deviceType := parseDeviceType(deviceTypeStr)

	err := c.t.Do(reqCtx, func(txCtx context.Context) error {
		return c.authService.Logout(txCtx, userIDStr, deviceType)
	})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	response.Success(ctx, 200, response.MsgLoggedOut, nil)
}

func (c *AuthController) tokenPairToResponse(pair *service.TokenPair) dto.LoginResponse {
	return dto.LoginResponse{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		ExpiresAt:        pair.ExpiresAt.Unix(),
		RefreshExpiresAt: pair.RefreshExpiresAt.Unix(),
	}
}
