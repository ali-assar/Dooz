package email

import (
	"context"
	"log/slog"
)

type noopGateway struct {
	logger *slog.Logger
}

func NewNoOpGateway(logger *slog.Logger) Gateway {
	return &noopGateway{logger: logger.With("layer", "NoOpEmailGateway")}
}

func (g *noopGateway) SendOTP(ctx context.Context, email string, code string) error {
	g.logger.Info("NoOp Email: OTP would be sent", "email", email, "code", code)
	return nil
}

func (g *noopGateway) Send(ctx context.Context, param Param) error {
	g.logger.Info("NoOp Email: message would be sent", "to", param.To, "subject", param.Subject)
	return nil
}
