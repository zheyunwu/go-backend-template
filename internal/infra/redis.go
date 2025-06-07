package infra

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-backend-template/config"
	"github.com/redis/go-redis/v9"
)

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	slog.Info("Redis connected successfully",
		"host", cfg.Redis.Host,
		"port", cfg.Redis.Port,
		"db", cfg.Redis.DB)

	return rdb
}
