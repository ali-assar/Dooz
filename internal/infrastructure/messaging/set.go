package messaging

import (
	"log/slog"

	"dooz/internal/infrastructure/email"
	"dooz/internal/infrastructure/sms"

	"github.com/google/wire"
)

var MessagingSet = wire.NewSet(
	NewUnifiedSenderFromGateways,
)

func NewUnifiedSenderFromGateways(smsGateway sms.Gateway, emailGateway email.Gateway, logger *slog.Logger) Sender {
	return NewUnifiedSender(smsGateway, emailGateway, logger)
}
