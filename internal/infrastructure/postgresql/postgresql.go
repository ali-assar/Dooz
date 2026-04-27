package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dooz/internal/constants"
	"dooz/internal/infrastructure/godotenv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Postgresql struct {
	env *godotenv.Env
	DB  *gorm.DB
}

func NewPostgresql(env *godotenv.Env) *Postgresql {
	return &Postgresql{env: env}
}

func (p *Postgresql) Setup() error {
	db, err := gorm.Open(postgres.Open(p.env.DatabaseHost), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create GORM connection: %w", err)
	}
	p.DB = db

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(constants.MaxOpenConns)
	sqlDB.SetMaxIdleConns(constants.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(constants.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(constants.ConnMaxIdleTime)

	return nil
}

func (p *Postgresql) HealthCheck(ctx context.Context) error {
	if p.DB == nil {
		return errors.New("database is not initialized")
	}
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (p *Postgresql) Close(ctx context.Context) error {
	if p.DB == nil {
		return nil
	}
	sqlDB, err := p.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL DB for closing: %w", err)
	}
	sqlDB.SetMaxIdleConns(0)

	done := make(chan error, 1)
	go func() {
		done <- sqlDB.Close()
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
		p.DB = nil
		return nil
	case <-ctx.Done():
		_ = sqlDB.Close()
		p.DB = nil
		return fmt.Errorf("database close timeout: %w", ctx.Err())
	}
}

func (p *Postgresql) GetDB() *gorm.DB {
	return p.DB
}
