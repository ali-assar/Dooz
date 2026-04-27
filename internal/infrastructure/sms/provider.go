package sms

import "log/slog"

type ProviderType byte

const (
	ProviderMessageWay ProviderType = 1
	ProviderNoOp       ProviderType = 2
	ProviderSMSir      ProviderType = 3
)

type Config struct {
	APIKey     string
	TemplateID int
}

func NewGateway(providerType ProviderType, config Config, logger *slog.Logger) Gateway {
	switch providerType {
	case ProviderMessageWay:
		return NewMessageWayGateway(config.APIKey, config.TemplateID, logger)
	case ProviderNoOp:
		return NewNoOpGateway(logger)
	case ProviderSMSir:
		return NewSmsirGateway(config.APIKey, config.TemplateID, logger)
	default:
		return NewNoOpGateway(logger)
	}
}
