package cron

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dooz/internal/constants"
	"dooz/internal/infrastructure/godotenv"
)

func Run(env *godotenv.Env, logger *slog.Logger) error {
	lg := logger.With("component", "cron")

	processOTPURL := fmt.Sprintf("http://localhost:%d/api/v1/cron/process-otp-outbox", env.HTTPPort)
	deleteExpiredOTPURL := fmt.Sprintf("http://localhost:%d/api/v1/cron/delete-expired-otps", env.HTTPPort)

	processOTPTicker := time.NewTicker(constants.ProcessOTPInterval)
	deleteExpiredOTPTicker := time.NewTicker(constants.DeleteExpiredOTPInterval)
	defer processOTPTicker.Stop()
	defer deleteExpiredOTPTicker.Stop()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	lg.Info("cron service started")

	for {
		select {
		case <-processOTPTicker.C:
			callCronEndpoint(ctx, lg, processOTPURL, "process-otp-outbox")
		case <-deleteExpiredOTPTicker.C:
			callCronEndpoint(ctx, lg, deleteExpiredOTPURL, "delete-expired-otps")
		case <-ctx.Done():
			lg.Info("cron service shutting down")
			return nil
		}
	}
}

func callCronEndpoint(parentCtx context.Context, lg *slog.Logger, url, job string) {
	ctx, cancel := context.WithTimeout(parentCtx, constants.CronHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		lg.Error("failed to create request", "job", job, "error", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		lg.Error("cron call failed", "job", job, "error", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			lg.Error("failed to close response body", "job", job, "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		lg.Error("cron call returned non-200", "job", job, "status", resp.Status)
		return
	}

	lg.Info("cron executed successfully", "job", job)
}
