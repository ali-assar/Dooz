package email

import (
	"context"
	"log/slog"
	"time"
)

type ProviderType byte

const (
	ProviderGomail ProviderType = 1
	ProviderNoOp   ProviderType = 2
)

type SMTPConfig struct {
	Host               string
	Port               int
	User               string
	Password           string
	Sender             string
	RequireTLS         bool
	InsecureSkipVerify bool
	Timeout            time.Duration
}

func NewGateway(ctx context.Context, providerType ProviderType, config SMTPConfig, logger *slog.Logger) Gateway {
	switch providerType {
	case ProviderGomail:
		return NewGomailGateway(ctx, config, logger)
	case ProviderNoOp:
		return NewNoOpGateway(logger)
	default:
		return NewNoOpGateway(logger)
	}
}
