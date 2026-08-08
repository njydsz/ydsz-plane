// Package automation — 同工作项串行化锁池单元测试。
package automation

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// TestIssueLockPool_SerializesSameIssue 验证同一工作项的锁互斥：
// 50 个 goroutine 竞争同一 issue 的锁，任意时刻持锁数必须为 1。
func TestIssueLockPool_SerializesSameIssue(t *testing.T) {
	pool := newIssueLockPool(64)
	var maxConcurrent int64
	var current int64

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unlock := pool.lockForIssue(1) // 全部竞争同一 issue
			cur := atomic.AddInt64(&current, 1)
			for {
				m := atomic.LoadInt64(&maxConcurrent)
				if cur <= m || atomic.CompareAndSwapInt64(&maxConcurrent, m, cur) {
					break
				}
			}
			time.Sleep(2 * time.Millisecond)
			atomic.AddInt64(&current, -1)
			unlock()
		}()
	}
	wg.Wait()

	if maxConcurrent != 1 {
		t.Fatalf("same issue not serialized: max concurrent = %d, want 1", maxConcurrent)
	}
}

// TestIssueLockPool_DistinctIssuesParallel 验证不同工作项可并行持有锁。
func TestIssueLockPool_DistinctIssuesParallel(t *testing.T) {
	pool := newIssueLockPool(64)
	var maxConcurrent int64
	var current int64

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			unlock := pool.lockForIssue(int64(1000 + n*4)) // 每 4 个一 shard 也可并行（不同锁对象）
			cur := atomic.AddInt64(&current, 1)
			for {
				m := atomic.LoadInt64(&maxConcurrent)
				if cur <= m || atomic.CompareAndSwapInt64(&maxConcurrent, m, cur) {
					break
				}
			}
			time.Sleep(2 * time.Millisecond)
			atomic.AddInt64(&current, -1)
			unlock()
		}(i)
	}
	wg.Wait()

	if maxConcurrent < 2 {
		t.Fatalf("distinct issues should run in parallel: max concurrent = %d, want >=2", maxConcurrent)
	}
}

// TestIssueLockPool_NoopForNonIssue 验证非工作项事件返回 no-op 解锁。
func TestIssueLockPool_NoopForNonIssue(t *testing.T) {
	pool := newIssueLockPool(64)
	unlock := pool.lockForIssue(0)
	unlock() // 不应 panic
	unlock = pool.lockForIssue(-5)
	unlock()
}

// TestEventIssueID 验证从事件 payload 提取 issue_id。
func TestEventIssueID(t *testing.T) {
	payload, _ := json.Marshal(map[string]any{"issue_id": 42, "workspace_id": 7})
	event := mq.EventEnvelope{EventType: "issue.updated", Payload: payload}
	if got := eventIssueID(event); got != 42 {
		t.Fatalf("eventIssueID = %d, want 42", got)
	}

	empty := mq.EventEnvelope{EventType: "sprint.started", Payload: json.RawMessage(`{"sprint_id":1}`)}
	if got := eventIssueID(empty); got != 0 {
		t.Fatalf("eventIssueID = %d, want 0 for non-issue event", got)
	}
}
