package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "dooz/docs"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"dooz/entity"
	"dooz/internal/app/api/routes"
	"dooz/internal/constants"
	"dooz/internal/infrastructure/godotenv"
	"dooz/internal/infrastructure/postgresql"
	"dooz/internal/infrastructure/redis"
	wsHub "dooz/internal/infrastructure/websocket"
	userRepo "dooz/internal/repository/user"
	"dooz/utils/encrypt"
	customValidator "dooz/utils/validator"
)

type Boot struct {
	router     *routes.Router
	postgresql *postgresql.Postgresql
	redis      *redis.Redis
	hub        *wsHub.Hub
	env        *godotenv.Env
	logger     *slog.Logger
	userRepo   userRepo.Repository
}

func NewBoot(
	router *routes.Router,
	postgresql *postgresql.Postgresql,
	redisClient *redis.Redis,
	hub *wsHub.Hub,
	env *godotenv.Env,
	logger *slog.Logger,
	userRepo userRepo.Repository,
) *Boot {
	return &Boot{
		router:     router,
		postgresql: postgresql,
		redis:      redisClient,
		hub:        hub,
		env:        env,
		logger:     logger,
		userRepo:   userRepo,
	}
}

func (b *Boot) Boot() {
	b.registerValidators()
	b.initializeSuperAdmin()

	// Start WebSocket hub goroutine
	go b.hub.Run()

	r := b.router.SetupRouter()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	port := fmt.Sprintf(":%d", b.env.HTTPPort)
	b.logger.Info("HTTP server starting", "port", port, "environment", b.env.Environment)
	b.logger.Info("Swagger available", "url", fmt.Sprintf("http://localhost%s/swagger/index.html", port))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := r.Run(port); err != nil {
			log.Fatalf("error running gin server: %s", err)
		}
	}()

	b.logger.Info("Application started successfully")
	<-sigChan
	b.logger.Info("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
	defer cancel()

	if b.postgresql != nil {
		_ = b.postgresql.Close(shutdownCtx)
	}
	if b.redis != nil {
		_ = b.redis.Close()
	}

	b.logger.Info("Shutdown complete")
}

func (b *Boot) initializeSuperAdmin() {
	if b.env.SuperAdminEmail == "" || b.env.SuperAdminPassword == "" {
		b.logger.Info("Super admin credentials not provided, skipping")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultRequestTimeout)
	defer cancel()

	lg := b.logger.With("method", "initializeSuperAdmin")

	existing, err := b.userRepo.GetByEmail(ctx, b.env.SuperAdminEmail)
	if err == nil && existing != nil {
		if existing.Role == entity.RoleSuperAdmin {
			lg.Info("Super admin already exists")
			return
		}
		existing.Role = entity.RoleSuperAdmin
		existing.PasswordHash = encrypt.HashSHA256(b.env.SuperAdminPassword)
		existing.IsEmailVerified = true
		existing.UpdatedAt = time.Now().Unix()
		if b.env.SuperAdminPhone != "" {
			existing.Phone = b.env.SuperAdminPhone
			existing.IsPhoneVerified = true
		}
		_ = b.userRepo.Update(ctx, existing)
		lg.Info("Updated existing user to super admin")
		return
	}

	now := time.Now().Unix()
	superAdmin := &entity.User{
		Email:           b.env.SuperAdminEmail,
		Fullname:        "Super Admin",
		PasswordHash:    encrypt.HashSHA256(b.env.SuperAdminPassword),
		UserCode:        900000,
		Role:            entity.RoleSuperAdmin,
		IsEmailVerified: true,
		CurrentTheme:    1,
		CurrentXOShape:  1,
		CurrentAvatar:   1,
		UpdatedAt:       now,
	}
	if b.env.SuperAdminPhone != "" {
		superAdmin.Phone = b.env.SuperAdminPhone
		superAdmin.IsPhoneVerified = true
	}

	if err := b.userRepo.Create(ctx, superAdmin); err != nil {
		lg.Error("Failed to create super admin", "error", err)
		return
	}
	lg.Info("Super admin created", "email", b.env.SuperAdminEmail)
}

func (b *Boot) registerValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := customValidator.RegisterIranianPhoneValidator(v); err != nil {
			b.logger.Error("Failed to register phone validator", "error", err)
		}
	}
}
