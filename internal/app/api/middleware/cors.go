package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

type CORSMiddleware struct {
	logger *slog.Logger
}

func NewCORSMiddleware(logger *slog.Logger) *CORSMiddleware {
	return &CORSMiddleware{logger: logger.With("layer", "CORSMiddleware")}
}

func (m *CORSMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "3600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
