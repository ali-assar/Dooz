package sms

import (
	"log/slog"

	"dooz/internal/infrastructure/godotenv"

	"github.com/google/wire"
)

var SMSGatewaySet = wire.NewSet(
	NewGatewayFromEnv,
)

func NewGatewayFromEnv(env *godotenv.Env, logger *slog.Logger) Gateway {
	providerType := ProviderType(env.SMSProvider)
	config := Config{
		APIKey:     env.SMSApiKey,
		TemplateID: env.SMSOTPTemplateID,
	}
	return NewGateway(providerType, config, logger)
}
