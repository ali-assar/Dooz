package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

const (
	ErrInvalidRequest = "invalid request"
	ErrInvalidFormat  = "invalid format"

	ErrUnauthorized       = "unauthorized"
	ErrMissingAuthHeader  = "missing authorization header"
	ErrInvalidAuthFormat  = "invalid authorization format"
	ErrInvalidToken       = "invalid token"
	ErrInvalidTokenClaims = "invalid token claims"

	ErrOTPNotSent = "failed to send OTP"
	ErrInvalidOTP = "invalid or expired OTP"

	ErrInternalServer  = "internal server error"
	ErrNotFound        = "resource not found"
	ErrConflict        = "resource conflict"
	ErrForbidden       = "forbidden"
	ErrTooManyRequests = "too many requests"

	ErrLogoutFailed    = "failed to logout"
	ErrSessionNotFound = "session not found"

	ErrInvalidTurnstileToken       = "invalid or expired turnstile token"
	ErrTurnstileVerificationFailed = "turnstile verification failed"
)

const (
	MsgOTPSent        = "OTP sent successfully"
	MsgOTPVerified    = "OTP verified successfully"
	MsgLoggedOut      = "Logged out successfully"
	MsgTokenRefreshed = "Token refreshed successfully"
	MsgCreated        = "Created successfully"
	MsgUpdated        = "Updated successfully"
	MsgDeleted        = "Deleted successfully"
)

func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{Success: true, Message: message, Data: data})
}

func SuccessWithData(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Response{Success: true, Data: data})
}

func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{Success: false, Error: &ErrorInfo{Message: message}})
}

func ErrorWithCode(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, Response{Success: false, Error: &ErrorInfo{Code: code, Message: message}})
}

func ValidationError(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = ErrUnauthorized
	}
	Error(c, http.StatusUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = ErrForbidden
	}
	Error(c, http.StatusForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = ErrNotFound
	}
	Error(c, http.StatusNotFound, message)
}

func Conflict(c *gin.Context, message string) {
	if message == "" {
		message = ErrConflict
	}
	Error(c, http.StatusConflict, message)
}

func InternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = ErrInternalServer
	}
	Error(c, http.StatusInternalServerError, message)
}

func TooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = ErrTooManyRequests
	}
	Error(c, http.StatusTooManyRequests, message)
}
