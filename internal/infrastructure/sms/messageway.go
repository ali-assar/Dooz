package sms

import (
	"context"
	"errors"
	"log/slog"

	MessageWay "github.com/MessageWay/MessageWayGolang"
)

type messagewayGateway struct {
	app        *MessageWay.App
	templateID int
	logger     *slog.Logger
}

func NewMessageWayGateway(apiKey string, templateID int, logger *slog.Logger) Gateway {
	if apiKey == "" {
		panic("SMS API key not configured: SMS_API_KEY environment variable is required")
	}

	app := MessageWay.New(MessageWay.Config{
		ApiKey:         apiKey,
		AcceptLanguage: "fa",
	})

	return &messagewayGateway{
		app:        app,
		templateID: templateID,
		logger:     logger.With("layer", "MessageWayGateway"),
	}
}

func (g *messagewayGateway) SendOTP(ctx context.Context, phone string, code string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	lg := g.logger.With("method", "SendOTP", "phone", phone, "templateID", g.templateID, "code", code)

	lg.Info("Sending OTP via SMS", "phone", phone, "otpCode", code)

	message := MessageWay.Message{
		Method:     "sms",
		Mobile:     phone,
		TemplateID: g.templateID,
		Params:     []string{code},
		Provider:   1,
		Code:       code,
		ExpireTime: 120,
	}
	res, err := g.app.Send(message)
	if err != nil {
		lg.Error("failed to send SMS", "error", err)
		return err
	}
	if res.ReferenceID == "" {
		lg.Error("SMS sent but no reference ID returned", "error", res.Error)
		return errors.New("SMS sent but no reference ID returned")
	}

	lg.Info("SMS sent successfully", "referenceID", res.ReferenceID)
	return nil
}
