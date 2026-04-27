package email

import (
	"context"
	"log/slog"

	"dooz/internal/constants"
	"dooz/internal/infrastructure/godotenv"

	"github.com/google/wire"
)

var EmailGatewaySet = wire.NewSet(
	NewGatewayFromEnv,
)

func NewGatewayFromEnv(ctx context.Context, env *godotenv.Env, logger *slog.Logger) Gateway {
	LoadTemplates(logger)

	providerType := ProviderType(env.EmailProvider)

	config := SMTPConfig{
		Host:               env.SMTPHost,
		Port:               env.SMTPPort,
		User:               env.SMTPUser,
		Password:           env.SMTPPassword,
		Sender:             env.SMTPSender,
		RequireTLS:         true,
		InsecureSkipVerify: false,
		Timeout:            constants.EmailSendTimeout,
	}

	return NewGateway(ctx, providerType, config, logger)
}
