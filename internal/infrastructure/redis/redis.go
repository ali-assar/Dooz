package redis

import (
	"context"
	"fmt"

	"dooz/internal/infrastructure/godotenv"

	goredis "github.com/redis/go-redis/v9"
)

type Redis struct {
	env    *godotenv.Env
	Client *goredis.Client
}

func NewRedis(env *godotenv.Env) *Redis {
	return &Redis{env: env}
}

func (r *Redis) Setup() error {
	r.Client = goredis.NewClient(&goredis.Options{
		Addr: r.env.RedisHost,
	})

	if err := r.Client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return nil
}

func (r *Redis) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}

func (r *Redis) GetClient() *goredis.Client {
	return r.Client
}
