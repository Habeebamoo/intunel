package configs

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg *Config, ctx context.Context) (*redis.Client, error) {
	redisOpts, err := redis.ParseURL(cfg.RedisUrl)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(redisOpts)

	// Ping Redis
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("Worker: Redis connected")
	return redisClient, nil
}