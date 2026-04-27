package middleware

import (
	"errors"
	"log/slog"

	"dooz/internal/app/api/response"
	appErrors "dooz/internal/errors"

	"github.com/gin-gonic/gin"
)

func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			var appErr *appErrors.AppError
			if errors.As(err, &appErr) {
				response.ErrorWithCode(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
				return
			}

			logger.Error("unhandled error", "error", err)
			response.InternalServerError(c, "An unexpected error occurred")
		}
	}
}
