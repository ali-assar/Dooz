package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type parameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type otpRequest struct {
	Mobile     string      `json:"mobile"`
	TemplateID int         `json:"templateId"`
	Parameters []parameter `json:"parameters"`
}

type smsirGateway struct {
	logger     *slog.Logger
	templateID int
	baseURL    string
	apiKey     string
	client     *http.Client // Added http client
}

// NewSmsirGateway creates a new instance of smsirGateway.
func NewSmsirGateway(apiKey string, templateID int, logger *slog.Logger) Gateway {
	if apiKey == "" {
		panic("SMS API key not configured: SMS_API_KEY environment variable is required")
	}

	return &smsirGateway{
		templateID: templateID,
		apiKey:     apiKey,
		baseURL:    "https://api.sms.ir/v1/send/verify",
		logger:     logger.With("layer", "sms.ir"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (g *smsirGateway) SendOTP(ctx context.Context, phone string, code string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	lg := g.logger.With("method", "SendOTP", "phone", phone, "templateID", g.templateID, "code", code)

	lg.Info("Sending OTP via SMS", "phone", phone, "otpCode", code)

	reqBody := otpRequest{
		Mobile:     phone,
		TemplateID: g.templateID,
		Parameters: []parameter{
			{Name: "CODE", Value: code},
		},
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		lg.Error("failed to marshal request body", "error", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL, bytes.NewReader(b))
	if err != nil {
		lg.Error("failed to create HTTP request", "error", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", g.apiKey)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		lg.Error("failed to send HTTP request", "error", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle the response
	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errorBody); decodeErr != nil {
			lg.Error("SMS.ir API returned non-200 status and failed to decode error body", "status", resp.StatusCode, "error", decodeErr)
			return fmt.Errorf("SMS.ir API error: status %d", resp.StatusCode)
		}
		errorMessage := "unknown error"
		if msg, ok := errorBody["message"].(string); ok {
			errorMessage = msg
		}
		lg.Error("SMS.ir API returned non-200 status", "status", resp.StatusCode, "message", errorMessage)
		return fmt.Errorf("SMS.ir API error: status %d, message: %s", resp.StatusCode, errorMessage)
	}

	lg.Info("OTP sent successfully", "status", resp.Status)
	return nil
}
