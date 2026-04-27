package service

import (
	"context"
	"dooz/entity"
	"dooz/internal/constants"
	"dooz/internal/infrastructure/messaging"
	"dooz/internal/repository/otp"
	"log/slog"
	"sync"
	"time"
)

type OTPWorkerPool struct {
	otpRepo      otp.Repository
	sender       messaging.Sender
	logger       *slog.Logger
	workerCount  int
	batchSize    int
	wg           sync.WaitGroup
	jobChan      chan *entity.OTP
	quitChan     chan struct{}
	started      sync.Once          // Ensures workers are started only once
	workerCtx    context.Context    // Context for workers (child of parent context passed to Start)
	workerCancel context.CancelFunc // Cancel function for worker context
}

func NewOTPWorkerPool(
	otpRepo otp.Repository,
	sender messaging.Sender,
	logger *slog.Logger,
) *OTPWorkerPool {
	return &OTPWorkerPool{
		otpRepo:     otpRepo,
		sender:      sender,
		workerCount: constants.OTPWorkerCount,
		batchSize:   constants.OTPWorkerBatchSize,
		logger:      logger.With("Layer", "OTPWorkerPool"),
		jobChan:     make(chan *entity.OTP, constants.OTPWorkerBatchSize*2),
		quitChan:    make(chan struct{}), // TODO: user context.NewWithCancel() instead it's the same
	}
}

func (p *OTPWorkerPool) Start(ctx context.Context) {
	p.started.Do(func() {
		p.logger.Info("starting OTP worker pool", "workers", p.workerCount)

		p.workerCtx, p.workerCancel = context.WithCancel(ctx)

		for i := 0; i < p.workerCount; i++ {
			p.wg.Go(func() {
				p.worker(p.workerCtx, i)
			})
		}
	})
}

func (p *OTPWorkerPool) Stop(ctx context.Context) error {
	lg := p.logger.With("method", "stop")

	lg.Info("stopping OTP worker pool")

	if p.workerCancel != nil {
		p.workerCancel()
	}

	close(p.quitChan)

	// Wait for workers to finish, respecting the timeout context
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		lg.Info("OTP worker pool stopped")
		return nil
	case <-ctx.Done():
		lg.Warn("timeout waiting for workers to stop", "error", ctx.Err())
		return ctx.Err()
	}
}

func (p *OTPWorkerPool) ProcessBatch(ctx context.Context) error {
	lg := p.logger.With("method", "ProcessBatch")

	pendingOTPs, err := p.otpRepo.GetPendingOTPs(ctx, p.batchSize)
	if err != nil {
		lg.Error("failed to get pending OTPs", "error", err)
		return err
	}

	if len(pendingOTPs) == 0 {
		lg.Debug("no pending OTPs to process")
		return nil
	}

	lg.Info("distributing OTPs to workers",
		"count", len(pendingOTPs),
		"batchSize", p.batchSize,
	)

	queuedCount := 0
	for _, otp := range pendingOTPs {
		select {
		case <-ctx.Done():
			lg.Warn("context cancelled while queuing OTPs",
				"queued", queuedCount,
				"total", len(pendingOTPs),
			)
			return ctx.Err()
		case <-p.quitChan:
			lg.Warn("worker pool stopped while queuing OTPs",
				"queued", queuedCount,
				"total", len(pendingOTPs),
			)
			return nil
		case p.jobChan <- otp:
			queuedCount++
		}
	}

	lg.Info("all OTPs queued for processing",
		"queued", queuedCount,
		"total", len(pendingOTPs),
	)
	return nil
}

func (p *OTPWorkerPool) worker(ctx context.Context, workerID int) {

	lg := p.logger.With("worker", workerID)
	lg.Info("worker started")

	for {
		select {
		case otp, ok := <-p.jobChan:
			if !ok {
				lg.Info("worker stopping (job channel closed)")
				return
			}
			if err := p.processOTP(ctx, otp); err != nil {
				lg.Error("failed to process OTP",
					"otpID", otp.ID,
					"recipient", otp.Recipient,
					"channel", otp.Channel,
					"error", err,
				)
			} else {
				lg.Debug("OTP processed successfully",
					"otpID", otp.ID,
					"recipient", otp.Recipient,
					"channel", otp.Channel,
				)
			}

		case <-p.quitChan:
			lg.Info("worker stopping (quit signal)")
			return

		case <-ctx.Done():
			lg.Info("worker stopping (context cancelled)")
			return
		}
	}
}

func (p *OTPWorkerPool) processOTP(ctx context.Context, otp *entity.OTP) error {
	lg := p.logger.With(
		"method", "processOTP",
		"otpID", otp.ID,
		"recipient", otp.Recipient,
		"channel", otp.Channel,
		"retryCount", otp.RetryCount,
	)

	sendCtx, cancel := context.WithTimeout(ctx, constants.OTPSendTimeout)
	defer cancel()

	err := p.sender.SendOTP(sendCtx, otp.Channel, otp.Recipient, otp.Code)
	if err != nil {
		lg.Error("failed to send OTP", "error", err)
		return p.handleRetry(ctx, otp)
	}

	if err := p.otpRepo.SetOTPStatus(ctx, otp.ID, entity.OTPStatusSent); err != nil {
		lg.Error("failed to mark OTP as sent", "error", err)
		return err
	}

	lg.Info("OTP sent successfully",
		"recipient", otp.Recipient,
		"channel", otp.Channel,
		"retryCount", otp.RetryCount,
	)
	return nil
}

func (p *OTPWorkerPool) handleRetry(ctx context.Context, otp *entity.OTP) error {
	newRetryCount := otp.RetryCount + 1

	// Check if max retries exceeded
	if newRetryCount >= constants.MaxOTPRetries {
		p.logger.Error("max retries exceeded, marking OTP as failed",
			"otpID", otp.ID,
			"recipient", otp.Recipient,
			"channel", otp.Channel,
			"retryCount", newRetryCount,
			"maxRetries", constants.MaxOTPRetries,
		)
		return p.otpRepo.SetOTPStatus(ctx, otp.ID, entity.OTPStatusFailed)
	}

	nextRetryTime := p.calculateNextRetry(newRetryCount)
	nextRetryAt := nextRetryTime.Unix()

	if err := p.otpRepo.UpdateRetryInfo(ctx, otp.ID, newRetryCount, nextRetryAt); err != nil {
		p.logger.Error("failed to update retry info",
			"otpID", otp.ID,
			"error", err,
		)
		return err
	}

	retryDelay := time.Until(nextRetryTime)
	p.logger.Warn("OTP send failed, scheduled for retry",
		"otpID", otp.ID,
		"recipient", otp.Recipient,
		"channel", otp.Channel,
		"retryCount", newRetryCount,
		"nextRetryAt", nextRetryTime.Format(time.RFC3339),
		"retryDelay", retryDelay.Round(time.Second),
	)
	return nil
}

// Formula: delay = base * (2 ^ retryCount)
// Example: retryCount 0 = 5s, 1 = 10s, 2 = 20s, 3 = 40s, 4+ = 60s (capped)
func (p *OTPWorkerPool) calculateNextRetry(retryCount int) time.Time {
	// Exponential backoff: delay = base * (2 ^ retryCount)
	delay := constants.OTPRetryBaseDelay * time.Duration(1<<uint(retryCount))

	// Cap at max delay to prevent excessive wait times
	if delay > constants.OTPRetryMaxDelay {
		delay = constants.OTPRetryMaxDelay
	}

	return time.Now().Add(delay)
}
