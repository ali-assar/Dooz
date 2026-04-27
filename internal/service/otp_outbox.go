package service

import (
	"context"
	"log/slog"
	"time"

	"dooz/internal/constants"
	"dooz/internal/infrastructure/messaging"
	otpRepo "dooz/internal/repository/otp"
)

type OTPOutboxService interface {
	ProcessPendingOTPs(ctx context.Context) error
	DeleteExpiredOTPs(ctx context.Context) error
	StartWorkerPool(ctx context.Context) error
	StopWorkerPool() error
}

type otpOutboxService struct {
	otpRepo    otpRepo.Repository
	sender     messaging.Sender
	workerPool *OTPWorkerPool
	logger     *slog.Logger
}

func NewOTPOutboxService(otpRepo otpRepo.Repository, sender messaging.Sender, logger *slog.Logger) OTPOutboxService {
	svc := &otpOutboxService{
		otpRepo: otpRepo,
		sender:  sender,
		logger:  logger.With("layer", "OTPOutboxService"),
	}
	svc.workerPool = NewOTPWorkerPool(otpRepo, sender, logger)
	return svc
}

func NewOTPOutboxServiceWithDefaults(otpRepo otpRepo.Repository, sender messaging.Sender, logger *slog.Logger) OTPOutboxService {
	return NewOTPOutboxService(otpRepo, sender, logger)
}

func (s *otpOutboxService) ProcessPendingOTPs(ctx context.Context) error {
	s.workerPool.Start(ctx)
	return s.workerPool.ProcessBatch(ctx)
}

func (s *otpOutboxService) StartWorkerPool(ctx context.Context) error {
	lg := s.logger.With("method", "StartWorkerPool")
	s.workerPool.Start(ctx)

	ticker := time.NewTicker(constants.ProcessOTPInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.workerPool.ProcessBatch(ctx); err != nil {
				lg.Error("failed to process batch", "error", err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *otpOutboxService) StopWorkerPool() error {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
	defer cancel()
	return s.workerPool.Stop(ctx)
}

func (s *otpOutboxService) DeleteExpiredOTPs(ctx context.Context) error {
	if err := s.otpRepo.DeleteExpired(ctx); err != nil {
		s.logger.Error("failed to delete expired OTPs", "error", err)
		return err
	}
	return nil
}
