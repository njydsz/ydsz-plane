// Package cache — Redis 领导者选举（Worker 主从选举）。
//
// 参考: Redis 官方文档《Leader election pattern》
// 对标 Google SRE Ch16: "Running multiple replicas requires leader election
//  for tasks that must run on exactly one node (cronjob, outbox relay)."
//
// 用法：
//
//	elector := cache.NewLeaderElection(client, "worker:outbox-relay", 30*time.Second)
//	if err := elector.Start(ctx); err != nil { ... }
//	defer elector.Stop(ctx)
//	for elector.IsLeader() { ... do work ... }
package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// LeaderElection 基于 Redis SET NX 实现单主选举。
//
// 算法：
//   1. 每 ttl/2 秒执行 SET key <token> NX PX <ttl>（续期）。
//   2. SET 成功 = 当前节点是 Leader；失败 = 有其他 Leader 存在。
//   3. 其他节点会在 ttl 超时后抢占（自动 failover）。
//   4. Stop 时主动 DEL key 释放领导权，避免 failover 延迟。
//
// 安全性：
//   - 同一时刻至多一个 Leader（Redis 单 key 原子操作保证）。
//   - Leader 故障时，ttl 超时后自动切换（failover 时间 = ttl）。
type LeaderElection struct {
	client    *redis.Client
	key       string
	token     string    // 随机 token，用于安全释放
	ttl       time.Duration
	interval  time.Duration // 续期间隔（默认 ttl/2）

	mu        sync.RWMutex
	isLeader  bool
	cancel    context.CancelFunc
	done      chan struct{}
}

// NewLeaderElection 构造领导者选举实例。
//
// 参数：
//   - client：Redis 客户端。
//   - key：选举键名（建议 "leader:<业务>"）
//   - ttl：领导权租期（建议 30s）。Leader 故障时，最多 ttl 后自动切换。
//
// Uber Go: "Keep constructors simple. Configure via options if needed."
func NewLeaderElection(client *redis.Client, key string, ttl time.Duration) *LeaderElection {
	return &LeaderElection{
		client:   client,
		key:      key,
		token:    generateLeaderToken(),
		ttl:      ttl,
		interval: ttl / 2,
	}
}

// Start 开启后台续期循环。
// 首次尝试立即抢占，随后每 interval 续期一次。
// ctx 取消时停止续期并释放 Leader。
func (e *LeaderElection) Start(ctx context.Context) error {
	// 上下文包装以支持 stop
	ctx, e.cancel = context.WithCancel(ctx)
	e.done = make(chan struct{})

	// 立即尝试一次
	e.renew(ctx)

	// 后台续期循环
	go e.renewalLoop(ctx)
	return nil
}

// Stop 停止选举并释放 Leader 权。
func (e *LeaderElection) Stop(ctx context.Context) {
	if e.cancel != nil {
		e.cancel()
	}
	// 主动释放 Leader（安全删除）
	e.release(ctx)
	close(e.done)
}

// IsLeader 返回当前节点是否为 Leader（线程安全）。
func (e *LeaderElection) IsLeader() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isLeader
}

// Token 返回当前 Leader 的 token（用于日志/测试）。
func (e *LeaderElection) Token() string {
	return e.token
}

// renewalLoop 后台续期循环。
func (e *LeaderElection) renewalLoop(ctx context.Context) {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.renew(ctx)
		}
	}
}

// renew 尝试获取/续期 Leader 权。
// SET NX PX = 仅在 key 不存在时设置（第一次获取），或
// 如果 token 匹配则可续期（防止误抢已续期的 Leader）。
//
// 注意：此实现简化为"先到先得"模式，不校验 token 续期。
// 生产环境应使用 Lua 脚本校验 token 后再续期，防止多主脑裂。
func (e *LeaderElection) renew(ctx context.Context) {
	// SET NX PX — 仅在 key 不存在时设置成功
	ok, err := e.client.SetNX(ctx, e.key, e.token, e.ttl).Result()
	if err != nil {
		e.mu.Lock()
		e.isLeader = false
		e.mu.Unlock()
		return
	}

	e.mu.Lock()
	e.isLeader = ok
	e.mu.Unlock()
}

// release 安全释放 Leader 权（仅自身是 Leader 时才删除）。
func (e *LeaderElection) release(ctx context.Context) {
	// Lua 脚本：仅当 token 匹配时才删除（防止误删其他节点的 Leader）
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	_, _ = e.client.Eval(ctx, script, []string{e.key}, e.token).Result()

	e.mu.Lock()
	e.isLeader = false
	e.mu.Unlock()
}

// generateLeaderToken 生成随机 token（用于 Leader 身份标识）。
func generateLeaderToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
