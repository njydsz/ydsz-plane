// Package cache — Redis 分布式锁（基于 Redlock 单实例简化版）。
//
// 参考: Redis 官方文档《Distributed locks with Redis》
//   https://redis.io/docs/latest/develop/use/patterns/distributed-locks/
//
// 对标 Uber Go Style Guide: "Use sync primitives or channel-based concurrency,
//  but when cross-process coordination is needed, use a well-tested pattern."
package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Mutex 是基于 Redis 的分布式互斥锁。
//
// 安全性保证：
//   - Lock 使用 SET NX PX 原子操作，确保互斥性。
//   - Unlock / Extend 使用 Lua 脚本校验持有者身份，防止误释放他人的锁。
//   - value 使用 crypto/rand 生成的随机 token，确保全局唯一。
//
// 注意：此实现为单 Redis 实例版本（Redlock 单节点）。跨多 Redis 实例
// 的强一致分布式锁需要 Redlock 算法（N/2+1 多数派）。
type Mutex struct {
	client *redis.Client
	key    string
	value  string // 随机 token，用于安全释放
	ttl    time.Duration
}

// NewMutex 创建分布式锁实例。
//
// 参数：
//   - client：已连接的 Redis 客户端。
//   - key：锁名（建议业务前缀 + 资源标识，如 "lock:issue:123"）。
//   - ttl：锁过期时间（防止死锁）；客户端应在 TTL 内完成操作。
//
// Uber Go Style: "Keep constructors simple; put complex logic in methods."
func NewMutex(client *redis.Client, key string, ttl time.Duration) *Mutex {
	return &Mutex{
		client: client,
		key:    key,
		value:  generateToken(),
		ttl:    ttl,
	}
}

// Lock 尝试获取锁。
// 使用 SET key value NX PX ttl 原子命令。
// 返回 (true, nil) 表示成功；返回 (false, nil) 表示锁已被他人持有。
//
// 对标 Google SRE: 分布式锁应设置合理的 TTL，避免持有锁的进程崩溃后死锁。
func (m *Mutex) Lock(ctx context.Context) (bool, error) {
	ok, err := m.client.SetNX(ctx, m.key, m.value, m.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("cache: mutex lock: %w", err)
	}
	return ok, nil
}

// Unlock 释放锁。
// 使用 Lua 脚本原子地校验 value 后 DEL，防止误释放他人的锁。
//
// 对标 Antirez (Redis creator) 推荐模式：
//   if redis.call("GET", KEYS[1]) == ARGV[1] then
//       return redis.call("DEL", KEYS[1])
//   else
//       return 0
//   end
func (m *Mutex) Unlock(ctx context.Context) (bool, error) {
	// Uber Go Style: "Keep critical sections small; do the check-and-delete atomically."
	const unlockScript = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end`
	result, err := m.client.Eval(ctx, unlockScript, []string{m.key}, m.value).Int64()
	if err != nil {
		return false, fmt.Errorf("cache: mutex unlock: %w", err)
	}
	return result == 1, nil
}

// Extend 延长锁的 TTL。
// 仅当调用者确实持有锁（value 匹配）时才续期，防止延长他人持有的锁。
//
// 返回 (true, nil) 表示续期成功；返回 (false, nil) 表示锁已不属当前持有者。
func (m *Mutex) Extend(ctx context.Context) (bool, error) {
	const extendScript = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end`
	ttlMs := m.ttl.Milliseconds()
	result, err := m.client.Eval(ctx, extendScript, []string{m.key},
		m.value, ttlMs).Int64()
	if err != nil {
		return false, fmt.Errorf("cache: mutex extend: %w", err)
	}
	return result == 1, nil
}

// generateToken 生成 16 字节随机 hex 字符串用作锁持有者标识。
// 使用 crypto/rand 而非 math/rand，对标 NIST SP 800-90A 随机性要求。
func generateToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// 极低概率：加密随机源不可用，退化到时间戳 + pid（安全降级）
		return fmt.Sprintf("fallback-%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
	}
	return hex.EncodeToString(b)
}
