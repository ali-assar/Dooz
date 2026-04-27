package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"dooz/entity"
	"dooz/internal/constants"
	appErrors "dooz/internal/errors"
	"dooz/internal/infrastructure/godotenv"
	otpRepo "dooz/internal/repository/otp"
	sessionRepo "dooz/internal/repository/session"
	userRepo "dooz/internal/repository/user"
	"dooz/utils/encrypt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type AuthService interface {
	Login(ctx context.Context, email string, password string, ip string, deviceType entity.DeviceType, remember bool) (*TokenPair, error)
	Logout(ctx context.Context, userID string, deviceType entity.DeviceType) error
	Register(ctx context.Context, phone string, email string, fullname string) error
	SetPassword(ctx context.Context, phone string, password string, ip string, deviceType entity.DeviceType, remember bool) (*TokenPair, error)
	RequestPasswordReset(ctx context.Context, phone string) error
	ResetPassword(ctx context.Context, phone string, code string, newPassword string, ip string, deviceType entity.DeviceType, remember bool) (*TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	VerifyPhoneOTP(ctx context.Context, phone string, code string, purpose entity.OTPPurpose) error
	VerifyEmailOTP(ctx context.Context, email string, code string, purpose entity.OTPPurpose) error
	SendPhoneOTP(ctx context.Context, phone string, purpose entity.OTPPurpose) error
	SendEmailOTP(ctx context.Context, email string, purpose entity.OTPPurpose) error
}

type authService struct {
	env         *godotenv.Env
	logger      *slog.Logger
	userRepo    userRepo.Repository
	sessionRepo sessionRepo.Repository
	otpRepo     otpRepo.Repository
	secret      string
}

func NewAuthService(env *godotenv.Env, logger *slog.Logger, userRepo userRepo.Repository, sessionRepo sessionRepo.Repository, otpRepo otpRepo.Repository) AuthService {
	return &authService{
		env:         env,
		logger:      logger.With("layer", "AuthService"),
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		otpRepo:     otpRepo,
		secret:      env.Secret,
	}
}

func (a *authService) sendOTP(ctx context.Context, recipient string, channel entity.OTPChannel, purpose entity.OTPPurpose) error {
	lg := a.logger.With("method", "sendOTP", "recipient", recipient)

	code := generateOTP(constants.OTPLength)
	now := time.Now()

	existing, err := a.otpRepo.GetLatestByRecipient(ctx, recipient, channel, purpose)
	if err == nil && existing != nil && existing.Status == entity.OTPStatusPending {
		_ = a.otpRepo.Delete(ctx, existing.ID)
	}

	otp := &entity.OTP{
		Recipient:  recipient,
		Channel:    channel,
		Code:       code,
		Purpose:    purpose,
		ExpiresAt:  now.Add(constants.OTPExpiration).Unix(),
		CreatedAt:  now.Unix(),
		Status:     entity.OTPStatusPending,
		RetryCount: 0,
	}

	if err := a.otpRepo.Create(ctx, otp); err != nil {
		lg.Error("failed to create OTP", "error", err)
		return err
	}
	lg.Info("OTP created and queued")
	return nil
}

func (a *authService) SendPhoneOTP(ctx context.Context, phone string, purpose entity.OTPPurpose) error {
	return a.sendOTP(ctx, phone, entity.OTPChannelSMS, purpose)
}

func (a *authService) SendEmailOTP(ctx context.Context, email string, purpose entity.OTPPurpose) error {
	return a.sendOTP(ctx, email, entity.OTPChannelEmail, purpose)
}

func (a *authService) verifyOTP(ctx context.Context, recipient string, channel entity.OTPChannel, code string, purpose entity.OTPPurpose) error {
	lg := a.logger.With("method", "verifyOTP", "recipient", recipient)

	if err := a.otpRepo.VerifyAndDelete(ctx, recipient, channel, code, purpose); err != nil {
		lg.Warn("OTP verification failed", "error", err)
		return err
	}

	if purpose == entity.RegistrationPurpose {
		if channel == entity.OTPChannelSMS {
			user, err := a.userRepo.GetByPhone(ctx, recipient)
			if err != nil {
				return userRepo.ErrNotFound
			}
			if !user.IsPhoneVerified {
				user.IsPhoneVerified = true
				user.UpdatedAt = time.Now().Unix()
				_ = a.userRepo.Update(ctx, user)
			}
		} else if channel == entity.OTPChannelEmail {
			user, err := a.userRepo.GetByEmail(ctx, recipient)
			if err != nil {
				return userRepo.ErrNotFound
			}
			if !user.IsEmailVerified {
				user.IsEmailVerified = true
				user.UpdatedAt = time.Now().Unix()
				_ = a.userRepo.Update(ctx, user)
			}
		}
	}

	lg.Info("OTP verified")
	return nil
}

func (a *authService) VerifyPhoneOTP(ctx context.Context, phone string, code string, purpose entity.OTPPurpose) error {
	return a.verifyOTP(ctx, phone, entity.OTPChannelSMS, code, purpose)
}

func (a *authService) VerifyEmailOTP(ctx context.Context, email string, code string, purpose entity.OTPPurpose) error {
	return a.verifyOTP(ctx, email, entity.OTPChannelEmail, code, purpose)
}

func (a *authService) Login(ctx context.Context, email, password, ip string, deviceType entity.DeviceType, remember bool) (*TokenPair, error) {
	lg := a.logger.With("method", "Login", "email", email)

	fetchedUser, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		lg.Warn("user not found")
		return nil, appErrors.ErrUnauthorized
	}
	if fetchedUser.DeletedAt > 0 {
		return nil, appErrors.ErrUnauthorized
	}
	if fetchedUser.PasswordHash == "" {
		return nil, appErrors.NewAppError("REGISTRATION_INCOMPLETE", "Please complete registration", 400)
	}
	if fetchedUser.PasswordHash != encrypt.HashSHA256(password) {
		lg.Warn("invalid password")
		return nil, userRepo.ErrInvalidPassword
	}

	tokenPair, err := a.createTokens(ctx, fetchedUser, ip, deviceType, remember)
	if err != nil {
		lg.Error("failed to create tokens", "error", err)
		return nil, err
	}

	lg.Info("user logged in")
	return tokenPair, nil
}

func (a *authService) Register(ctx context.Context, phone, email, fullname string) error {
	lg := a.logger.With("method", "Register", "phone", phone, "email", email)

	if _, err := a.userRepo.GetByPhone(ctx, phone); err == nil {
		return userRepo.ErrDuplicatePhone
	}
	if _, err := a.userRepo.GetByEmail(ctx, email); err == nil {
		return userRepo.ErrDuplicateEmail
	}

	now := time.Now().Unix()
	newUser := &entity.User{
		Phone:     phone,
		Email:     email,
		Fullname:  fullname,
		Role:      entity.RoleUser,
		Avatar:    "",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := a.userRepo.Create(ctx, newUser); err != nil {
		lg.Error("failed to create user", "error", err)
		return err
	}

	if err := a.SendPhoneOTP(ctx, phone, entity.RegistrationPurpose); err != nil {
		lg.Error("failed to send OTP", "error", err)
		return err
	}

	lg.Info("user registered")
	return nil
}

func (a *authService) SetPassword(ctx context.Context, phone, password, ip string, deviceType entity.DeviceType, remember bool) (*TokenPair, error) {
	lg := a.logger.With("method", "SetPassword", "phone", phone)

	fetchedUser, err := a.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, userRepo.ErrNotFound
	}
	if !fetchedUser.IsPhoneVerified {
		return nil, appErrors.NewAppError("PHONE_NOT_VERIFIED", "Phone not verified", 401)
	}
	if fetchedUser.PasswordHash != "" {
		return nil, appErrors.NewAppError("PASSWORD_ALREADY_SET", "Password already set. Use login.", 400)
	}

	fetchedUser.PasswordHash = encrypt.HashSHA256(password)
	fetchedUser.UpdatedAt = time.Now().Unix()
	if err := a.userRepo.Update(ctx, fetchedUser); err != nil {
		return nil, err
	}
	_ = a.sessionRepo.DeleteByUserID(ctx, fetchedUser.ID)

	tokenPair, err := a.createTokens(ctx, fetchedUser, ip, deviceType, remember)
	if err != nil {
		return nil, err
	}

	lg.Info("password set successfully")
	return tokenPair, nil
}

func (a *authService) RequestPasswordReset(ctx context.Context, phone string) error {
	lg := a.logger.With("method", "RequestPasswordReset", "phone", phone)

	fetchedUser, err := a.userRepo.GetByPhone(ctx, phone)
	if err != nil || fetchedUser.PasswordHash == "" {
		return nil
	}

	if err := a.SendPhoneOTP(ctx, phone, entity.ForgotPasswordPurpose); err != nil {
		lg.Error("failed to send OTP", "error", err)
		return err
	}
	return nil
}

func (a *authService) ResetPassword(ctx context.Context, phone, code, newPassword, ip string, deviceType entity.DeviceType, remember bool) (*TokenPair, error) {
	lg := a.logger.With("method", "ResetPassword", "phone", phone)

	if err := a.VerifyPhoneOTP(ctx, phone, code, entity.ForgotPasswordPurpose); err != nil {
		return nil, err
	}

	fetchedUser, err := a.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, userRepo.ErrNotFound
	}

	fetchedUser.PasswordHash = encrypt.HashSHA256(newPassword)
	fetchedUser.UpdatedAt = time.Now().Unix()
	if err := a.userRepo.Update(ctx, fetchedUser); err != nil {
		return nil, err
	}
	_ = a.sessionRepo.DeleteByUserID(ctx, fetchedUser.ID)

	tokenPair, err := a.createTokens(ctx, fetchedUser, ip, deviceType, remember)
	if err != nil {
		return nil, err
	}

	lg.Info("password reset successfully")
	return tokenPair, nil
}

func (a *authService) createTokens(ctx context.Context, user *entity.User, ip string, deviceType entity.DeviceType, remember bool) (*TokenPair, error) {
	lg := a.logger.With("method", "createTokens", "userID", user.ID)

	now := time.Now()
	jwtID := uuid.New().String()
	accessExpiresAt := now.Add(constants.AccessTokenExpiration)

	accessClaims := jwt.MapClaims{
		"userID":    user.ID,
		"role":      int(user.Role),
		"jwtID":     jwtID,
		"exp":       accessExpiresAt.Unix(),
		"iat":       now.Unix(),
		"Phone":     user.Phone,
		"Email":     user.Email,
		"Fullname":  user.Fullname,
		"Avatar":    user.Avatar,
		"CreatedAt": user.CreatedAt,
		"UpdatedAt": user.UpdatedAt,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(a.secret))
	if err != nil {
		return nil, err
	}

	var refreshExpiresAt time.Time
	if remember {
		refreshExpiresAt = now.Add(constants.RefreshTokenLongExpiration)
	} else {
		refreshExpiresAt = now.Add(constants.RefreshTokenExpiration)
	}

	refreshClaims := jwt.MapClaims{
		"userID": user.ID,
		"jwtID":  jwtID,
		"exp":    refreshExpiresAt.Unix(),
		"iat":    now.Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(a.secret))
	if err != nil {
		return nil, err
	}

	nowUnix := now.Unix()
	var ipAddress *string
	if ip != "" {
		ipAddress = &ip
	}

	newSession := &entity.Session{
		UserID:           user.ID,
		DeviceType:       deviceType,
		RefreshTokenHash: encrypt.HashSHA256(refreshTokenString),
		JwtID:            jwtID,
		ExpiresAt:        refreshExpiresAt.Unix(),
		IPAddress:        ipAddress,
		LastActivityAt:   nowUnix,
		CreatedAt:        nowUnix,
	}

	if err := a.sessionRepo.Create(ctx, newSession); err != nil {
		lg.Error("failed to upsert session", "error", err)
		return nil, err
	}

	return &TokenPair{
		AccessToken:      accessTokenString,
		RefreshToken:     refreshTokenString,
		ExpiresAt:        accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (a *authService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	lg := a.logger.With("method", "RefreshToken")

	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.secret), nil
	})
	if err != nil || !token.Valid {
		return nil, appErrors.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, appErrors.ErrUnauthorized
	}
	userID, _ := claims["userID"].(string)
	jwtID, _ := claims["jwtID"].(string)

	sess, err := a.sessionRepo.GetByJwtID(ctx, jwtID)
	if err != nil {
		return nil, appErrors.ErrUnauthorized
	}
	if time.Now().Unix() > sess.ExpiresAt {
		return nil, appErrors.NewAppError("TOKEN_EXPIRED", "Token expired", 401)
	}
	if sess.RefreshTokenHash != encrypt.HashSHA256(refreshToken) {
		return nil, appErrors.ErrUnauthorized
	}

	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	nowUnix := now.Unix()
	accessExpiresAt := now.Add(constants.AccessTokenExpiration)

	accessClaims := jwt.MapClaims{
		"userID":    userID,
		"role":      int(user.Role),
		"jwtID":     jwtID,
		"exp":       accessExpiresAt.Unix(),
		"iat":       nowUnix,
		"Phone":     user.Phone,
		"Email":     user.Email,
		"Fullname":  user.Fullname,
		"Avatar":    user.Avatar,
		"CreatedAt": user.CreatedAt,
		"UpdatedAt": user.UpdatedAt,
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(a.secret))
	if err != nil {
		return nil, err
	}

	_ = a.sessionRepo.UpdateLastActivity(ctx, sess.ID)

	lg.Info("token refreshed")
	return &TokenPair{
		AccessToken:      accessTokenString,
		RefreshToken:     refreshToken,
		ExpiresAt:        accessExpiresAt,
		RefreshExpiresAt: time.Unix(sess.ExpiresAt, 0),
	}, nil
}

func (a *authService) Logout(ctx context.Context, userID string, deviceType entity.DeviceType) error {
	lg := a.logger.With("method", "Logout", "userID", userID)
	if err := a.sessionRepo.DeleteByUserAndDevice(ctx, userID, deviceType); err != nil {
		if errors.Is(err, sessionRepo.ErrNotFound) {
			return nil
		}
		return err
	}
	lg.Info("user logged out")
	return nil
}

func generateOTP(length int) string {
	const digits = "0123456789"
	b := make([]byte, length)
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		for i := range b {
			b[i] = digits[time.Now().UnixNano()%10]
		}
		return string(b)
	}
	for i := range b {
		b[i] = digits[int(randomBytes[i])%len(digits)]
	}
	return string(b)
}
