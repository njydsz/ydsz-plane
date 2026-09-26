// Package cache — Redis 连接池指标导出（P1-2）。
//
// 对标 Google SRE Book Ch.6: "Four Golden Signals — Saturation"。
// 暴露 Redis 连接池状态和命令延迟，用于监控 Redis 健康状况。
package cache

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
)

var (
	// RedisCommandsTotal 统计按命令类型聚合的 Redis 操作总数。
	// 维度 labels: command=SET/GET/DEL/HSET/ZADD/EVAL/..., status=ok|error。
	RedisCommandsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "plane",
			Name:      "redis_commands_total",
			Help:      "Total Redis commands executed.",
		},
		[]string{"command", "status"},
	)

	// RedisCommandDuration 统计 Redis 命令延迟分布。
	RedisCommandDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "plane",
			Name:      "redis_command_duration_ms",
			Help:      "Redis command latency distribution.",
			Buckets:   []float64{0.1, 0.5, 1, 2.5, 5, 10, 25, 50, 100, 250},
		},
		[]string{"command"},
	)

	// RedisPoolStats 导出 Redis 连接池状态（pool_size / idle / active）。
	RedisPoolStats = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "plane",
			Name:      "redis_pool_stats",
			Help:      "Redis client pool statistics.",
		},
		[]string{"status"}, // "pool_size" | "idle" | "in_use"
	)

	// RedisHitRatio 缓存命中率（GET 命中 / GET 总数）。
	// 需配合 cache-aside 模式的上层埋点使用。
	RedisHitRatio = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "plane",
			Name:      "cache_hit_total",
			Help:      "Cache hit/miss counter.",
		},
		[]string{"result"}, // "hit" | "miss"
	)
)

// CollectPoolStats 采集并导出 Redis 连接池指标。
// 应在后台 Goroutine 中定期调用（如每 15 秒）。
func CollectPoolStats(client *redis.Client) {
	if client == nil {
		return
	}
	poolStats := client.PoolStats()
	RedisPoolStats.WithLabelValues("total_conns").Set(float64(poolStats.TotalConns))
	RedisPoolStats.WithLabelValues("idle").Set(float64(poolStats.IdleConns))
	RedisPoolStats.WithLabelValues("stale").Set(float64(poolStats.StaleConns))
	RedisPoolStats.WithLabelValues("in_use").Set(float64(poolStats.TotalConns - poolStats.IdleConns))
}

// InstrumentedClient 包装 *redis.Client 以自动记录 Prometheus 指标。
// 使用方式：
//
//	client := cache.NewClient(ctx, addr, passwd, db)
//	wrapped := cache.NewInstrumentedClient(client)
//	wrapped.Set(ctx, "key", "value", 0) // 自动记录延迟和计数
type InstrumentedClient struct {
	inner *redis.Client
}

// NewInstrumentedClient 包装 Redis 客户端以自动记录指标。
func NewInstrumentedClient(client *redis.Client) *InstrumentedClient {
	return &InstrumentedClient{inner: client}
}

// PoolStats 返回底层连接池统计信息。
func (c *InstrumentedClient) PoolStats() *redis.PoolStats {
	return c.inner.PoolStats()
}

// Close 关闭底层客户端。
func (c *InstrumentedClient) Close() error {
	return c.inner.Close()
}
