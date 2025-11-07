package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"vcs.technonext.com/carrybee/ride_engine/pkg/config"
)

type RedisDB struct {
	Client *redis.Client
}

func NewRedisDB(cfg config.RedisConfig) (*RedisDB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	fmt.Println("Redis connected successfully")
	return &RedisDB{Client: client}, nil
}

func (r *RedisDB) Close() error {
	fmt.Println("Closing Redis connection...")
	return r.Client.Close()
}

func (r *RedisDB) HealthCheck(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}
