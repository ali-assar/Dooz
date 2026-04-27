package controllers

import (
	"log/slog"

	"dooz/entity"
	"dooz/internal/app/api/response"

	"github.com/gin-gonic/gin"
)

func parseDeviceType(deviceTypeStr string) entity.DeviceType {
	switch deviceTypeStr {
	case "mobile":
		return entity.DeviceMobile
	case "telegram":
		return entity.DeviceTelegram
	default:
		return entity.DeviceWeb
	}
}

func getUserIDFromContext(ctx *gin.Context, logger *slog.Logger) (string, bool) {
	userID, exists := ctx.Get("userID")
	if !exists {
		logger.Warn("userID not found in context")
		response.Unauthorized(ctx, response.ErrUnauthorized)
		return "", false
	}
	userIDStr, ok := userID.(string)
	if !ok {
		response.Unauthorized(ctx, response.ErrUnauthorized)
		return "", false
	}
	return userIDStr, true
}

func getUserIDAndRoleFromContext(ctx *gin.Context, logger *slog.Logger) (userID string, role string, ok bool) {
	userIDVal, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, response.ErrUnauthorized)
		return "", "", false
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		response.Unauthorized(ctx, response.ErrUnauthorized)
		return "", "", false
	}
	roleVal, exists := ctx.Get("role")
	if !exists {
		response.Unauthorized(ctx, response.ErrUnauthorized)
		return "", "", false
	}
	roleStr, ok := roleVal.(string)
	if !ok {
		response.Unauthorized(ctx, response.ErrUnauthorized)
		return "", "", false
	}
	return userIDStr, roleStr, true
}

func isAdmin(role string) bool {
	return role == "admin" || role == "super_admin"
}
