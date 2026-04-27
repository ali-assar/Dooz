package middleware

import (
	"log/slog"
	"strings"

	"dooz/internal/app/api/response"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

type CasbinMiddleware struct {
	enforcer *casbin.Enforcer
	logger   *slog.Logger
}

func NewCasbinMiddleware(enforcer *casbin.Enforcer, logger *slog.Logger) *CasbinMiddleware {
	return &CasbinMiddleware{
		enforcer: enforcer,
		logger:   logger.With("layer", "CasbinMiddleware"),
	}
}

func (m *CasbinMiddleware) Enforce() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Unauthorized(c, response.ErrUnauthorized)
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			response.Unauthorized(c, response.ErrUnauthorized)
			c.Abort()
			return
		}

		resource := m.extractResource(c)
		action := m.extractAction(c.Request.Method)

		allowed, err := m.enforcer.Enforce(roleStr, resource, action)
		if err != nil {
			response.InternalServerError(c, response.ErrInternalServer)
			c.Abort()
			return
		}
		if !allowed {
			response.Forbidden(c, response.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *CasbinMiddleware) extractResource(c *gin.Context) string {
	path := c.FullPath()
	switch {
	case strings.Contains(path, "/logout") || strings.Contains(path, "/sessions"):
		return "sessions"
	case strings.Contains(path, "/users"):
		return "users"
	case strings.Contains(path, "/game") || strings.Contains(path, "/match"):
		return "game"
	case strings.Contains(path, "/friends"):
		return "friends"
	case strings.Contains(path, "/chat"):
		return "chat"
	case strings.Contains(path, "/leaderboard"):
		return "leaderboard"
	case strings.Contains(path, "/achievements"):
		return "achievements"
	case strings.Contains(path, "/otp"):
		return "otp"
	case strings.Contains(path, "/cron"):
		return "system"
	default:
		return "unknown"
	}
}

func (m *CasbinMiddleware) extractAction(method string) string {
	switch method {
	case "GET":
		return "read"
	case "POST":
		return "write"
	case "PUT", "PATCH":
		return "write"
	case "DELETE":
		return "delete"
	default:
		return "read"
	}
}
