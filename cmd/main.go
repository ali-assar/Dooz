package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"dooz/cmd/cron"
	"dooz/cmd/seed"
	"dooz/internal/infrastructure/godotenv"
	"dooz/internal/infrastructure/postgresql"
	"dooz/internal/infrastructure/redis"
)

//	@title			Dooz API
//	@version		1.0
//	@description	Tic-Tac-Toe Game API — auth, matchmaking, real-time, leaderboards, achievements

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

func main() {
	logger := initSlogLogger()
	env := godotenv.NewEnv()

	cronflag := flag.Bool("cron", false, "Trigger cron jobs")
	seedflag := flag.Bool("seed", false, "Seed database")
	seedCount := flag.Int("seed-count", 50, "Number of users to seed")
	flag.Parse()

	if *cronflag {
		if err := cron.Run(env, logger); err != nil {
			logger.Error("cron failed", "error", err)
			os.Exit(1)
		}
		return
	}

	pg := postgresql.NewPostgresql(env)
	if err := pg.Setup(); err != nil {
		log.Fatalf("failed to initialize PostgreSQL: %s", err)
	}

	rd := redis.NewRedis(env)
	if err := rd.Setup(); err != nil {
		log.Fatalf("failed to initialize Redis: %s", err)
	}

	if *seedflag {
		if err := seed.Run(pg.GetDB(), logger, *seedCount); err != nil {
			logger.Error("seed failed", "error", err)
			os.Exit(1)
		}
		logger.Info("seeding complete")
		return
	}

	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	boot, err := wireApp(mainCtx, logger, pg, rd)
	if err != nil {
		log.Fatalf("failed to setup app: %s", err)
	}
	boot.Boot()
}

func initSlogLogger() *slog.Logger {
	opts := &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)
	return logger
}
