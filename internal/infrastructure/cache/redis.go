// Package cache provides the Redis client used for caching, rate limiting,
// distributed locks and WebSocket fan-out.
package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// NewClient creates a Redis client and verifies connectivity.
func NewClient(ctx context.Context, addr, password string, db int) (*redis.Client, error) {
	cli := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	if err := cli.Ping(ctx).Err(); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("cache: ping redis: %w", err)
	}
	return cli, nil
}
