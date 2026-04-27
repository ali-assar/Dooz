package middleware

import (
	"fmt"
	"log/slog"
	"strings"

	"dooz/internal/app/api/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	secret string
	logger *slog.Logger
}

func NewAuthMiddleware(secret string, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		secret: secret,
		logger: logger.With("layer", "AuthMiddleware"),
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := m.logger.With("path", c.Request.URL.Path)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, response.ErrMissingAuthHeader)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, response.ErrInvalidAuthFormat)
			c.Abort()
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.secret), nil
		})
		if err != nil || !token.Valid {
			logger.Warn("invalid token", "error", err)
			response.Unauthorized(c, response.ErrInvalidToken)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Unauthorized(c, response.ErrInvalidTokenClaims)
			c.Abort()
			return
		}

		userID, ok := claims["userID"].(string)
		if !ok {
			response.Unauthorized(c, response.ErrInvalidTokenClaims)
			c.Abort()
			return
		}

		var role string
		if roleInt, ok := claims["role"].(float64); ok {
			switch int(roleInt) {
			case 1:
				role = "user"
			case 2:
				role = "admin"
			case 3:
				role = "super_admin"
			default:
				role = "user"
			}
		} else if roleStr, ok := claims["role"].(string); ok {
			role = roleStr
		} else {
			response.Unauthorized(c, response.ErrInvalidTokenClaims)
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("role", role)
		c.Next()
	}
}
