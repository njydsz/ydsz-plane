// Package cache — Redis 滑动窗口限流器。
//
// 参考: Redis 官方文档《Rate limiting pattern — sliding window》。
//
// 对标 Google SRE Book 第 21 章《Overload》：
//  "Protecting against overload requires both client-side throttling and server-side rate limiting."
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter 基于 Redis ZSET 的滑动窗口限流器。
//
// 算法：
//   1. ZADD key <current_timestamp_micro> <unique_member> — 记录本次请求。
//   2. ZREMRANGEBYSCORE key 0 <window_start> — 清理过期条目。
//   3. ZCARD key — 统计当前窗口请求数。
//   4. EXPIRE key <window> — 设置 ZSET 过期时间作为兜底清理。
//
// 使用 Lua 脚本保证三步操作的原子性（对标官方 Redis 文档推荐模式）。
type RateLimiter struct {
	client    *redis.Client
	keyPrefix string
	limit     int
	window    time.Duration
}

// NewRateLimiter 构造滑动窗口限流器。
//
// 参数：
//   - client：已连接的 Redis 客户端。
//   - keyPrefix：限流键前缀（如 "ratelimit"）。
//   - limit：窗口内最大请求次数。
//   - window：窗口大小（如 time.Minute）。
//
// Uber Go Style: "Keep constructors simple; put complex logic in methods."
func NewRateLimiter(client *redis.Client, keyPrefix string, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client:    client,
		keyPrefix: keyPrefix,
		limit:     limit,
		window:    window,
	}
}

// ReportAll 判断指定 key 是否允许本次请求，并返回剩余配额。
//
// 返回值：
//   - allowed：true 表示允许；false 表示触发限流。
//   - remaining：当前窗口剩余可请求次数（仅 allowed=true 时有效）。
//   - err：Redis 错误。
//
// 对标 Uber Go Style: "Return values in a predictable order (result, error)"
// 这里为了可读性调整为 (allowed, remaining, error)，因为 remaining 是辅助信息。
func (r *RateLimiter) ReportAll(ctx context.Context, key string) (bool, int, error) {
	fullKey := r.keyPrefix + ":" + key
	now := time.Now().UnixMicro()
	windowStart := now - r.window.Microseconds()

	const slideLua = `
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local ttl_sec = tonumber(ARGV[4])
		local member = ARGV[5]

		-- 1. 清除窗口外旧记录
		redis.call("ZREMRANGEBYSCORE", key, "-inf", window_start)

		-- 2. 统计当前窗口内请求数
		local count = redis.call("ZCARD", key)

		if count >= limit then
			return {0, 0}
		end

		-- 3. 计入本次请求
		redis.call("ZADD", key, now, member)

		-- 4. 刷新键过期时间（兜底清理）
		redis.call("EXPIRE", key, ttl_sec)

		local remaining = limit - count - 1
		return {1, remaining}`

	ttlSec := int(r.window.Seconds())
	if ttlSec < 1 {
		ttlSec = 1
	}
	// 唯一 member 确保同微秒内的不同请求也被独立计数
	member := fmt.Sprintf("%d-%d", now, now%1000000)

	result, err := r.client.Eval(ctx, slideLua, []string{fullKey},
		now, windowStart, r.limit, ttlSec, member).Result()
	if err != nil {
		return false, 0, fmt.Errorf("cache: rate limiter eval: %w", err)
	}

	vals, ok := result.([]interface{})
	if !ok || len(vals) != 2 {
		return false, 0, fmt.Errorf("cache: rate limiter unexpected result type: %T", result)
	}

	allowed := vals[0].(int64) == 1
	remaining := int(vals[1].(int64))
	return allowed, remaining, nil
}
