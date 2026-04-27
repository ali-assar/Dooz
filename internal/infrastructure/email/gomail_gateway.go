package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"

	"dooz/internal/constants"

	gomail "gopkg.in/mail.v2"
)

type gomailGateway struct {
	dialer *gomail.Dialer
	sender string
	logger *slog.Logger
}

func NewGomailGateway(ctx context.Context, cfg SMTPConfig, logger *slog.Logger) Gateway {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Password)
	dialer.TLSConfig = &tls.Config{
		ServerName:         cfg.Host,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}
	if cfg.Timeout > 0 {
		dialer.Timeout = cfg.Timeout
	} else {
		dialer.Timeout = constants.SMTPConnectionTimeout
	}
	dialer.SSL = cfg.Port == 465

	return &gomailGateway{
		dialer: dialer,
		sender: cfg.Sender,
		logger: logger.With("layer", "GomailGateway"),
	}
}

func (g *gomailGateway) SendOTP(ctx context.Context, email string, code string) error {
	lg := g.logger.With("method", "SendOTP", "email", email)

	m := gomail.NewMessage()
	m.SetHeader("From", g.sender)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your Dooz verification code")

	var htmlBody string
	if OTPTemplate != nil {
		var buf bytes.Buffer
		if err := OTPTemplate.Execute(&buf, OTPTemplateData{Code: code}); err == nil {
			htmlBody = buf.String()
		}
	}
	if htmlBody == "" {
		htmlBody = fmt.Sprintf("<p>Your verification code is: <strong>%s</strong></p>", code)
	}

	m.SetBody("text/html", htmlBody)
	m.AddAlternative("text/plain", fmt.Sprintf("Your Dooz code: %s (expires in 10 minutes)", code))

	sendCtx, cancel := context.WithTimeout(ctx, constants.EmailSendTimeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- g.dialer.DialAndSend(m)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			lg.Error("failed to send OTP email", "error", err)
			return fmt.Errorf("failed to send email: %w", err)
		}
		lg.Info("OTP email sent successfully")
		return nil
	case <-sendCtx.Done():
		lg.Error("email send timeout")
		return fmt.Errorf("email send timeout: %w", sendCtx.Err())
	}
}

func (g *gomailGateway) Send(ctx context.Context, param Param) error {
	lg := g.logger.With("method", "Send", "to", param.To)

	m := gomail.NewMessage()
	m.SetHeader("From", g.sender)
	m.SetHeader("To", param.To)
	m.SetHeader("Subject", param.Subject)

	ct := param.ContentType
	if ct == "" {
		ct = "text/html"
	}
	m.SetBody(ct, param.Body)

	for _, att := range param.Attachments {
		m.AttachReader(att.Filename, bytes.NewReader(att.Data))
	}

	sendCtx, cancel := context.WithTimeout(ctx, constants.EmailSendTimeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- g.dialer.DialAndSend(m)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			lg.Error("failed to send email", "error", err)
			return fmt.Errorf("failed to send email: %w", err)
		}
		lg.Info("email sent successfully")
		return nil
	case <-sendCtx.Done():
		return fmt.Errorf("email send timeout: %w", sendCtx.Err())
	}
}
