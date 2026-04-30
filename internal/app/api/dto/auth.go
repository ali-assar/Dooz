package dto

type LoginRequest struct {
	Phone      string `json:"phone" binding:"required,iranian_phone"`
	Password   string `json:"password" binding:"required"`
	DeviceType string `json:"device_type" binding:"required,oneof=web mobile telegram"`
	Remember   bool   `json:"remember"`
}

type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required,iranian_phone"`
	Email    string `json:"email" binding:"required,email"`
	Fullname string `json:"fullname" binding:"required,min=3,max=100"`
}

type VerifyRegistrationOTPRequest struct {
	Phone string `json:"phone" binding:"required,iranian_phone"`
	Code  string `json:"code" binding:"required,len=5,numeric"`
}

type VerifyEmailOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=5,numeric"`
}

type SendEmailOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type SetPasswordRequest struct {
	Phone      string `json:"phone" binding:"required,iranian_phone"`
	Password   string `json:"password" binding:"required,min=8,max=72"`
	DeviceType string `json:"device_type" binding:"required,oneof=web mobile telegram"`
	Remember   bool   `json:"remember"`
}

type RequestPasswordResetRequest struct {
	Phone string `json:"phone" binding:"required,iranian_phone"`
}

type ResetPasswordRequest struct {
	Phone       string `json:"phone" binding:"required,iranian_phone"`
	Code        string `json:"code" binding:"required,len=5,numeric"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
	DeviceType  string `json:"device_type" binding:"required,oneof=web mobile telegram"`
	Remember    bool   `json:"remember"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LoginResponse struct {
	AccessToken      string  `json:"access_token"`
	RefreshToken     string  `json:"refresh_token"`
	ExpiresAt        int64   `json:"expires_at"`
	RefreshExpiresAt int64   `json:"refresh_expires_at"`
	User             UserDTO `json:"user"`
}
