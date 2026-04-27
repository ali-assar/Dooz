package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"dooz/internal/infrastructure/godotenv"
)

type clientIPKey struct{}

func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, ip)
}

func GetClientIP(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey{}).(string)
	return ip
}

type TurnstileService struct {
	env        *godotenv.Env
	httpClient *http.Client
	logger     *slog.Logger
}

func NewTurnstileService(env *godotenv.Env, logger *slog.Logger) *TurnstileService {
	return &TurnstileService{
		env:        env,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger.With("layer", "TurnstileService"),
	}
}

type turnstileVerifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIP string `json:"remoteip,omitempty"`
}

type turnstileVerifyResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes,omitempty"`
}

func (s *TurnstileService) Verify(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("turnstile token is required")
	}
	if s.env.CloudflareTurnstileSecretKey == "" {
		return fmt.Errorf("turnstile not configured")
	}

	body := turnstileVerifyRequest{
		Secret:   s.env.CloudflareTurnstileSecretKey,
		Response: token,
		RemoteIP: GetClientIP(ctx),
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://challenges.cloudflare.com/turnstile/v0/siteverify", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to verify: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var verifyResp turnstileVerifyResponse
	if err := json.Unmarshal(respBody, &verifyResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	if !verifyResp.Success {
		return fmt.Errorf("turnstile failed: %v", verifyResp.ErrorCodes)
	}
	return nil
}
