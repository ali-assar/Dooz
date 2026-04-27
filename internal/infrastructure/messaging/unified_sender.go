package messaging

import (
	"context"
	"fmt"
	"log/slog"

	"dooz/entity"
	"dooz/internal/infrastructure/email"
	"dooz/internal/infrastructure/sms"
)

type unifiedSender struct {
	smsGateway   sms.Gateway
	emailGateway email.Gateway
	logger       *slog.Logger
}

func NewUnifiedSender(smsGateway sms.Gateway, emailGateway email.Gateway, logger *slog.Logger) Sender {
	return &unifiedSender{
		smsGateway:   smsGateway,
		emailGateway: emailGateway,
		logger:       logger.With("layer", "UnifiedSender"),
	}
}

func (s *unifiedSender) SendOTP(ctx context.Context, channel entity.OTPChannel, recipient string, code string) error {
	switch channel {
	case entity.OTPChannelSMS:
		return s.smsGateway.SendOTP(ctx, recipient, code)
	case entity.OTPChannelEmail:
		return s.emailGateway.SendOTP(ctx, recipient, code)
	default:
		return fmt.Errorf("unknown OTP channel: %d", channel)
	}
}
