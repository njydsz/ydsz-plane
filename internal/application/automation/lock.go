// Package automation — 同工作项串行化锁池。
//
// 文档（models.go）声称"同一工作项的规则执行经 Redis 锁串行化"，但此前
// 引擎并未实现任何锁：事件消费者与 scheduled 调度器分属不同 goroutine，
// 同一 issue 的事件与定时求值可能并发执行，导致状态流转/字段更新的竞态。
//
// 本文件实现进程内分片互斥锁池（按 issue_id 哈希到 64 个 shard）：
//   - 覆盖同一进程内事件驱动（RunConsumer）与定时（RunScheduledCron）
//     两条执行路径的并发互斥；
//   - 无锁表膨胀（固定 shard 数），锁粒度 = 单工作项；
//
// 跨实例（多 worker HA）的串行化需要分布式锁（Redis SETNX），列为 Phase 3；
// 当前 RabbitMQ 单队列由单实例消费，进程内锁已满足正确性要求。
package automation

import (
	"encoding/json"
	"sync"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// issueLockShards 锁池分片数（2 的幂便于哈希取模）。
const issueLockShards = 64

// issueLockPool 是包级共享的按工作项分片互斥锁池。
// 事件消费者与定时调度器共用同一池，保证同 issue 求值互斥。
var issueLockPool = newIssueLockPool(issueLockShards)

// issueLockShardsPool 分片互斥锁池。
type issueLockShardsPool struct {
	shards []*sync.Mutex
}

// newIssueLockPool 创建分片锁池。
func newIssueLockPool(n int) *issueLockShardsPool {
	p := &issueLockShardsPool{shards: make([]*sync.Mutex, n)}
	for i := range p.shards {
		p.shards[i] = &sync.Mutex{}
	}
	return p
}

// lockForIssue 锁定指定工作项，返回解锁函数。
// issueID <= 0 时返回 no-op（非工作项事件无需串行化）。
func (p *issueLockShardsPool) lockForIssue(issueID int64) (unlock func()) {
	if issueID <= 0 {
		return func() {}
	}
	m := p.shards[uint64(issueID)%uint64(len(p.shards))]
	m.Lock()
	return m.Unlock
}

// eventIssueID 从事件 payload 中提取 issue_id（无则返回 0）。
func eventIssueID(event mq.EventEnvelope) int64 {
	var payload struct {
		IssueID int64 `json:"issue_id"`
	}
	_ = json.Unmarshal(event.Payload, &payload)
	return payload.IssueID
}
