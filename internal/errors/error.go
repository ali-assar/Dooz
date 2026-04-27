package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

var (
	ErrNotFound = &AppError{
		Code:       "NOT_FOUND",
		Message:    "resource not found",
		HTTPStatus: http.StatusNotFound,
	}
	ErrUnauthorized = &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "unauthorized",
		HTTPStatus: http.StatusUnauthorized,
	}
	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "forbidden",
		HTTPStatus: http.StatusForbidden,
	}
	ErrBadRequest = &AppError{
		Code:       "BAD_REQUEST",
		Message:    "bad request",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrInternalServer = &AppError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "internal server error",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrConflict = &AppError{
		Code:       "CONFLICT",
		Message:    "conflict",
		HTTPStatus: http.StatusConflict,
	}
)

func NewAppError(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

func Wrap(err error, code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}
