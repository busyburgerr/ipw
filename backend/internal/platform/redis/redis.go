// Package redis opens the application's Redis client.
package redis

import (
	"context"
	"fmt"
	"time"

	"ipw/internal/config"

	"github.com/redis/go-redis/v9"
)

// Open connects to Redis and verifies the connection with a ping.
func Open(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}
