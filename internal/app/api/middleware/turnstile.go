package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"

	"dooz/internal/app/api/response"
	"dooz/internal/infrastructure/godotenv"
	"dooz/internal/service"

	"github.com/gin-gonic/gin"
)

type TurnstileMiddleware struct {
	turnstileService *service.TurnstileService
	env              *godotenv.Env
	logger           *slog.Logger
}

func NewTurnstileMiddleware(turnstileService *service.TurnstileService, env *godotenv.Env, logger *slog.Logger) *TurnstileMiddleware {
	return &TurnstileMiddleware{
		turnstileService: turnstileService,
		env:              env,
		logger:           logger.With("layer", "TurnstileMiddleware"),
	}
}

func (m *TurnstileMiddleware) Verify() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.env.CloudflareTurnstileEnabled {
			c.Next()
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response.ValidationError(c, "Invalid request format")
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var body map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			response.ValidationError(c, "Invalid request format")
			c.Abort()
			return
		}

		token, ok := body["cf-turnstile-response"].(string)
		if !ok || token == "" {
			response.ValidationError(c, "Turnstile token is required")
			c.Abort()
			return
		}

		userIP := c.ClientIP()
		ctxWithIP := service.WithClientIP(c.Request.Context(), userIP)
		c.Request = c.Request.WithContext(ctxWithIP)

		if err := m.turnstileService.Verify(ctxWithIP, token); err != nil {
			m.logger.Warn("turnstile verification failed", "error", err)
			response.ValidationError(c, "Turnstile verification failed")
			c.Abort()
			return
		}

		c.Next()
	}
}
