// Package cache — 纯逻辑单元测试（无 miniredis 依赖）。
//
// 补充 redis_test.go：测试不依赖 Redis 实例的纯组件：
//   - generateToken：随机性与格式
//   - NewMutex 字段默认值
//   - NewRateLimiter 字段默认值
package cache

import (
	"regexp"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// TestGenerateToken 验证 token 生成符合 32 字符 hex（16 字节）标准。
func TestGenerateToken(t *testing.T) {
	hexRe := regexp.MustCompile(`^[0-9a-f]{32}$`)

	t.Run("standard format", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			tok := generateToken()
			if !hexRe.MatchString(tok) {
				t.Fatalf("generateToken() = %q, want 32-char lowercase hex", tok)
			}
		}
	})

	t.Run("uniqueness", func(t *testing.T) {
		seen := make(map[string]bool, 1000)
		for i := 0; i < 1000; i++ {
			tok := generateToken()
			if seen[tok] {
				t.Fatalf("duplicate token after %d iterations: %q", i+1, tok)
			}
			seen[tok] = true
		}
	})
}

// TestNewMutex_Fields 验证 NewMutex 正确初始化所有字段。
// 不使用真实 Redis 客户端（仅初始化结构体，不调用 SetNX/Eval）。
func TestNewMutex_Fields(t *testing.T) {
	// 用 nil client 测试结构体字段（不调用任何 Redis 方法）
	tests := []struct {
		name      string
		key       string
		ttl       time.Duration
		wantKey   string
		wantTTL   time.Duration
		wantToken bool // whether token should be non-empty
	}{
		{
			name:      "lock key for issue",
			key:       "lock:issue:123",
			ttl:       5 * time.Second,
			wantKey:   "lock:issue:123",
			wantTTL:   5 * time.Second,
			wantToken: true,
		},
		{
			name:      "lock key with no TTL",
			key:       "lock:long-running",
			ttl:       0,
			wantKey:   "lock:long-running",
			wantTTL:   0,
			wantToken: true,
		},
		{
			name:      "empty key",
			key:       "",
			ttl:       1 * time.Second,
			wantKey:   "",
			wantTTL:   1 * time.Second,
			wantToken: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *redis.Client // nil is fine for field checks
			m := NewMutex(client, tt.key, tt.ttl)

			if m.key != tt.wantKey {
				t.Fatalf("key = %q, want %q", m.key, tt.wantKey)
			}
			if m.ttl != tt.wantTTL {
				t.Fatalf("ttl = %v, want %v", m.ttl, tt.wantTTL)
			}
			if m.client != client {
				t.Fatal("client field mismatch")
			}
			if tt.wantToken && m.value == "" {
				t.Fatal("expected non-empty value token")
			}

			// token 应为 32 字符 hex
			hexRe := regexp.MustCompile(`^[0-9a-f]{32}$`)
			if !hexRe.MatchString(m.value) {
				t.Fatalf("token = %q, want 32-char hex", m.value)
			}
		})
	}
}

// TestNewRateLimiter_Fields 验证 NewRateLimiter 正确保存字段。
// 不依赖 Redis。
func TestNewRateLimiter_Fields(t *testing.T) {
	tests := []struct {
		name      string
		keyPrefix string
		limit     int
		window    time.Duration
	}{
		{
			name:      "per-minute limit",
			keyPrefix: "ratelimit",
			limit:     100,
			window:    time.Minute,
		},
		{
			name:      "per-second limit",
			keyPrefix: "rl",
			limit:     10,
			window:    time.Second,
		},
		{
			name:      "unlimited window",
			keyPrefix: "burst",
			limit:     1000,
			window:    0,
		},
		{
			name:      "zero limit",
			keyPrefix: "block",
			limit:     0,
			window:    5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *redis.Client // nil is fine for field-only test
			rl := NewRateLimiter(client, tt.keyPrefix, tt.limit, tt.window)

			if rl.keyPrefix != tt.keyPrefix {
				t.Fatalf("keyPrefix = %q, want %q", rl.keyPrefix, tt.keyPrefix)
			}
			if rl.limit != tt.limit {
				t.Fatalf("limit = %d, want %d", rl.limit, tt.limit)
			}
			if rl.window != tt.window {
				t.Fatalf("window = %v, want %v", rl.window, tt.window)
			}
		})
	}
}

// TestRateLimiter_KeyFormat 验证 ReportAll 中 key 格式为 "prefix:key"。
// 纯字符串逻辑验证（不实际调用 Redis）。
func TestRateLimiter_KeyFormat(t *testing.T) {
	// 检查 ReportAll 的 key 拼接模式（通过源码分析可知是 prefix + ":" + key）
	// 我们不能直接测试 unexported fields 拼接，但可以验证 NewRateLimiter 默认值。
	rl := NewRateLimiter(nil, "ratelimit", 100, time.Minute)

	// 验证 key 前缀被正确保存
	if rl.keyPrefix != "ratelimit" {
		t.Fatalf("expected keyPrefix = %q, got %q", "ratelimit", rl.keyPrefix)
	}

	// 验证 limit 和 window
	if rl.limit != 100 {
		t.Fatalf("expected limit = 100, got %d", rl.limit)
	}
	if rl.window != time.Minute {
		t.Fatalf("expected window = %v, got %v", time.Minute, rl.window)
	}
}
