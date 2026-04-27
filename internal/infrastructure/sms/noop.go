package sms

import (
	"context"
	"log/slog"
)

type noopGateway struct {
	logger *slog.Logger
}

func NewNoOpGateway(logger *slog.Logger) Gateway {
	return &noopGateway{logger: logger.With("layer", "NoOpSMSGateway")}
}

func (g *noopGateway) SendOTP(ctx context.Context, phone string, code string) error {
	g.logger.Info("NoOp SMS: OTP would be sent", "phone", phone, "code", code)
	return nil
}
