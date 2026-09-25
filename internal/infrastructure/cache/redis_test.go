// Package cache — Redis 分布式锁与限流器黑盒测试。
//
// 使用 miniredis（github.com/alicebob/miniredis/v2）模拟真实 Redis 行为，
// 无需外部 Redis 实例即可验证分布式锁与滑动窗口限流的核心逻辑。
//
// 对标 Google SRE "Testing" philosophy:
//   "Run the same binary in test and production; use fakes over mocks for integration points."
// miniredis provides a real (in-process) Redis implementation, preferred over hand-rolled mocks.
package cache

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newTestRedis 启动一个 miniredis 实例并返回真实 *redis.Client 连接。
// Uber Go Style: "Test helpers should handle setup/teardown via t.Cleanup()."
func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run: %v", err)
	}
	t.Cleanup(mr.Close)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return mr, client
}

// --- Mutex 基础测试 ---

// TestMutex_LockUnlock 验证加锁 → 解锁基础流程。
// 对标 Redis 分布式锁官方模式：SET NX PX + Lua DEL-if-match.
func TestMutex_LockUnlock(t *testing.T) {
	_, client := newTestRedis(t)
	ctx := t.Context()

	m := NewMutex(client, "test:lock:1", 5*time.Second)

	// 首次加锁应成功
	ok, err := m.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock: %v", err)
	}
	if !ok {
		t.Fatal("expected Lock to succeed, got false")
	}

	// 第二次加锁应失败（锁已被持有）
	m2 := NewMutex(client, "test:lock:1", 5*time.Second)
	ok, err = m2.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock (second): %v", err)
	}
	if ok {
		t.Fatal("expected second Lock to fail (lock held), got true")
	}

	// 原始持有者解锁应成功
	released, err := m.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if !released {
		t.Fatal("expected Unlock to succeed for original holder")
	}

	// 其他实例尝试解锁应失败（value 不匹配）
	released, err = m2.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock (wrong holder): %v", err)
	}
	if released {
		t.Fatal("expected Unlock to fail for non-holder, got true")
	}
}

// --- RateLimiter 基础测试 ---

// TestRateLimiter_AllowWithinLimit 验证滑动窗口限流基础逻辑。
// 对标 SRE 限流策略：Window 内未超限放行，超限拒绝。
func TestRateLimiter_AllowWithinLimit(t *testing.T) {
	_, client := newTestRedis(t)
	ctx := t.Context()

	rl := NewRateLimiter(client, "test:rl", 3, time.Second)

	// 前 3 次请求应放行
	for i := 0; i < 3; i++ {
		allowed, remaining, err := rl.ReportAll(ctx, "user:1")
		if err != nil {
			t.Fatalf("Allow #%d: %v", i+1, err)
		}
		if !allowed {
			t.Fatalf("request #%d: expected allowed=true, got false", i+1)
		}
		if remaining != 2-i {
			t.Fatalf("request #%d: expected remaining=%d, got %d", i+1, 2-i, remaining)
		}
	}

	// 第 4 次请求应被拒绝（超出 limit=3）
	allowed, _, err := rl.ReportAll(ctx, "user:1")
	if err != nil {
		t.Fatalf("Allow #4: %v", err)
	}
	if allowed {
		t.Fatal("request #4: expected allowed=false (rate limited), got true")
	}

	// 不同 key 互不影响
	allowed, _, err = rl.ReportAll(ctx, "user:2")
	if err != nil {
		t.Fatalf("Allow (different key): %v", err)
	}
	if !allowed {
		t.Fatal("different key should not be rate limited")
	}
}
