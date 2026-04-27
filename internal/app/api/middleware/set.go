package middleware

import (
	"github.com/google/wire"
)

var MiddlewareSet = wire.NewSet(
	NewAuthMiddleware,
	NewCasbinMiddleware,
	NewTurnstileMiddleware,
	NewCORSMiddleware,
)
