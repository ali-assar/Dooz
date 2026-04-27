//go:build wireinject
// +build wireinject

package main

import (
	"context"
	"log/slog"

	"github.com/casbin/casbin/v2"
	"github.com/google/wire"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"dooz/internal/app/api/controllers"
	"dooz/internal/app/api/middleware"
	"dooz/internal/app/api/routes"
	"dooz/internal/infrastructure"
	"dooz/internal/infrastructure/godotenv"
	"dooz/internal/infrastructure/postgresql"
	"dooz/internal/infrastructure/redis"
	wsHub "dooz/internal/infrastructure/websocket"
	"dooz/internal/repository"
	"dooz/internal/service"
)

func wireApp(
	ctx context.Context,
	logger *slog.Logger,
	pg *postgresql.Postgresql,
	rd *redis.Redis,
) (*Boot, error) {
	panic(wire.Build(
		infrastructure.InfrastructureSet,
		getGormDB,
		getRedisClient,
		initializeCasbinEnforcer,
		getSecret,
		wire.NewSet(wsHub.NewHub),

		repository.RepositorySet,
		service.ServiceSet,
		controllers.ControllerSet,
		middleware.MiddlewareSet,
		routes.RouteSet,
		wire.NewSet(NewBoot),
	))
}

func getGormDB(pg *postgresql.Postgresql, ctx context.Context) (*gorm.DB, error) {
	if err := pg.Setup(); err != nil {
		return nil, err
	}
	return pg.GetDB(), nil
}

func getRedisClient(rd *redis.Redis) *goredis.Client {
	return rd.GetClient()
}

func initializeCasbinEnforcer(env *godotenv.Env) (*casbin.Enforcer, error) {
	return casbin.NewEnforcer("config/rbac_model.conf", "config/rbac_policy.csv")
}

func getSecret(env *godotenv.Env) string {
	return env.Secret
}
