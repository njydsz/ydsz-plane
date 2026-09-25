// Package events 实现事务性 outbox 模式：
// 领域事件在业务事务内写入 domain_events 表，随后后台 relay 将它们发布到
// RabbitMQ（EventExchange）并标记为已发布。消费者必须按事件 ID 幂等处理。
//
// 消息栈（M1 升级后）：
//   - RabbitMQ：EventExchange（topic、at-least-once 投递 + 消费者 ack、
//     死信路由、消息 TTL）。处理必须可靠投递的领域事件（通知、webhook、
//     ES 索引、审计轨迹、自动化规则）。
//   - Redis Streams：保留用于低延迟实时推送（WebSocket 扇出）、限流、
//     分布式锁与 Asynq 任务代理。
//
// 事件 RabbitMQ 拓扑：
//   - Exchange "plane.events"（topic、持久化）
//   - 路由键：plane.events.<aggregate_type>.<event_type>
//   - 死信 exchange "plane.dlx" 路由有毒消息。
//   - 消息持久化（DeliveryMode=2），broker 重启后仍然存在。
//
// 重试与可观测性：
//   - Nack 失败的消息重新入队（受 DLX 阈值限制）。
//   - Prometheus 计数器跟踪发布/nack 数量（S3+ 加入）。
//   - 结构化日志携带 event_id、worker_id、routing_key 便于追踪。
package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// --- Prometheus Metrics (对标 Google SRE Book Ch.17: 时延分布监控) ---

var (
	// OutboxLagSeconds 记录 observed_at 与 occurred_at 之间的时延分布（含）。
	OutboxLagSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "plane",
			Name:      "outbox_lag_seconds",
			Help:      "Lag between event observed_at (relay poll time) and occurred_at.",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60},
		},
	)

	// OutboxEventsPublishedTotal 统计成功发布到 RabbitMQ 的事件总数。
	OutboxEventsPublishedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "plane",
			Name:      "outbox_events_published_total",
			Help:      "Total outbox events successfully published to RabbitMQ.",
		},
	)
)

// Event 是存储在 PostgreSQL outbox（domain_events）中的领域事件记录。
type Event struct {
	ID            int64           `json:"id"`
	WorkspaceID   int64           `json:"workspace_id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   int64           `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	OccurredAt    time.Time       `json:"occurred_at"`
}

// toEnvelope 将 outbox 记录转换为 mq.EventEnvelope。
func (e Event) toEnvelope() mq.EventEnvelope {
	return mq.EventEnvelope{
		EventID:       e.ID,
		EventType:     e.EventType,
		WorkspaceID:   e.WorkspaceID,
		AggregateType: e.AggregateType,
		AggregateID:   e.AggregateID,
		Payload:       e.Payload,
		OccurredAt:    e.OccurredAt,
		Exchange:      mq.EventExchange,
		RoutingKey:    mq.RoutingKey(e.AggregateType, e.EventType),
	}
}

// Recorder 在既有事务内将事件写入 outbox。
type Recorder struct{}

// NewRecorder 构造 Recorder。
func NewRecorder() *Recorder { return &Recorder{} }

// Record 向 domain_events 插入一条事件。必须在业务事务内调用，
// 以保证事件与状态变更原子提交。
func (r *Recorder) Record(ctx context.Context, tx pgx.Tx, e Event) error {
	const q = `INSERT INTO domain_events
		(workspace_id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4, $5)`
	if _, err := tx.Exec(ctx, q, e.WorkspaceID, e.AggregateType, e.AggregateID, e.EventType, e.Payload); err != nil {
		return fmt.Errorf("events: record %s: %w", e.EventType, err)
	}
	return nil
}

// Tx 是 relay 单批次所需的事务接口。
// 使用事务 + FOR UPDATE SKIP LOCKED 保证多个 relay 实例并发拉取时
// 不会重复消费同一条事件（行锁由同一事务持有直至 commit/rollback）。
type Tx interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Querier 是 relay 所需的最小 DB 接口（连接池满足该接口），
// 并提供 Begin 以开启事务。
type Querier interface {
	Tx
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Relay 轮询未发布事件并发布到 RabbitMQ。
//
// 生命周期：
//  1. NewRelay 初始化 RabbitMQ 客户端并声明拓扑。
//  2. Run(ctx) 启动轮询循环直至 ctx 取消。
//  3. 关闭时 ctx 取消后，进行中的批次完成后返回。
//
// 可观测性：每个批次周期记录已发布数量、错误与延迟。
// 使用 Prometheus 指标记录事件 Lag 分布和成功发布计数。
//
// 事件驱动唤醒 (对标 Google SRE Book Ch.17: 及时性 vs 效率权衡):
//  - 正常模式：500ms 周期轮询（节省 CPU）
//  - 有 NotifyNewEvent 信号时立即唤醒一次（降低端到端延迟）
//  - wakeup channel 不对外暴露，仅 NotifyNewEvent() 可触发
//
// 错误隔离：单条发布失败不阻塞整批——失败事件单独记录错误计数并
// 触发 OnPublishFailed 回调由上层处理（告警/死信标记），Relay 继续
// 投递后续事件，避免一条毒消息阻塞整个 outbox 链路。
type Relay struct {
	db     Querier
	mq     *mq.Client
	log    *zap.Logger
	batch  int
	period time.Duration

	// wakeup 用于事件驱动立即唤醒。单容量通道，NotifyNewEvent 非阻塞发送。
	wakeup chan struct{}

	// 发布失败时回调（可选，用于对账任务及告警）
	OnPublishFailed func(ctx context.Context, e Event, publishErr error)
}

// NewRelay 使用已初始化的 MQ 客户端构造 Relay。
// 调用方负责通过 Relay.Close 关闭 MQ 客户端，或共享客户端由外部关闭。
func NewRelay(db Querier, mqClient *mq.Client, log *zap.Logger) *Relay {
	r := &Relay{db: db, mq: mqClient, log: log, batch: 200, period: 500 * time.Millisecond}
	r.wakeup = make(chan struct{}, 1) // capacity=1，确保非阻塞通知又不丢失
	return r
}

// NotifyNewEvent 通知 Relay 有新事件待发布（非阻塞）。
// 该方法可被 application 层在写入 domain_events 后调用以降低端到端延迟。
// 由于 wakeup 通道容量为 1，重复通知不会造成 goroutine 泄漏。
func (r *Relay) NotifyNewEvent() {
	if r == nil {
		return
	}
	select {
	case r.wakeup <- struct{}{}:
	default:
		// 已有一个待处理唤醒信号，无需重复发送
	}
}

// Run 启动轮询循环直至 ctx 取消。
//
// 事件驱动唤醒 (对标 Google SRE Book Ch.17):
//  - 正常模式下按 period 周期轮询（节电）
//  - 当 wakeup 通道有信号时立即执行一轮 poll-to-publish
//    实现亚 period 级的端到端延迟
func (r *Relay) Run(ctx context.Context) {
	r.log.Info("outbox relay started",
		zap.Int("batch", r.batch),
		zap.Duration("period", r.period),
		zap.String("exchange", mq.EventExchange))

	ticker := time.NewTicker(r.period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.log.Info("outbox relay stopped (context cancelled)")
			return
		case <-r.wakeup:
			// 事件被写入 outbox 后立即唤醒
			r.runBatch(ctx)
		case <-ticker.C:
			r.runBatch(ctx)
		}
	}
}

// runBatch 执行单个轮询+发布批次，并自动在 sleep 前补充时间片。
func (r *Relay) runBatch(ctx context.Context) {
	if err := r.publishBatch(ctx); err != nil {
		r.log.Error("outbox relay batch failed", zap.Error(err))
	}
}

// publishBatch 轮询未发布事件并逐条发布到 RabbitMQ。
//
// 并发安全：通过 `BEGIN ... FOR UPDATE SKIP LOCKED` 在行级加锁，
// 多个 relay 实例并发运行时不会重复消费同一条事件。同一事务内
// SELECT 后逐条标记 published_at，发布语义保持 at-least-once：
// 若 PublishEvent 失败则跳过标记，解锁后仍可由本实例或另一实例重试。
//
// 可观测性：对每条成功发布的事件记录 lag 分布（observed_at - occurred_at）
// 和 published counter。
func (r *Relay) publishBatch(ctx context.Context) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("events: begin tx: %w", err)
	}
	// 无论成功失败都回滚/提交，确保锁被释放。
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now() // observed_at 近似值（用于 lag 计算）

	rows, err := tx.Query(ctx, `
		SELECT id, workspace_id, aggregate_type, aggregate_id, event_type, payload, occurred_at
		FROM domain_events
		WHERE published_at IS NULL
		ORDER BY id
		LIMIT $1
		FOR UPDATE SKIP LOCKED`, r.batch)
	if err != nil {
		return fmt.Errorf("events: poll: %w", err)
	}

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.AggregateType, &e.AggregateID, &e.EventType, &e.Payload, &e.OccurredAt); err != nil {
			rows.Close()
			return fmt.Errorf("events: scan: %w", err)
		}
		events = append(events, e)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	var published int
	for _, e := range events {
		envelope := e.toEnvelope()
		if err := r.mq.PublishEvent(ctx, envelope); err != nil {
			// 错误隔离：单条发布失败不中断整批，且留待下次重试。
			r.log.Warn("outbox publish single event failed (retry later)",
				zap.Int64("event_id", e.ID),
				zap.String("event_type", e.EventType),
				zap.Error(err))
			if r.OnPublishFailed != nil {
				r.OnPublishFailed(ctx, e, err)
			}
			continue
		}
		if _, err := tx.Exec(ctx,
			`UPDATE domain_events SET published_at = now() WHERE id = $1`,
			e.ID); err != nil {
			r.log.Warn("outbox mark published failed",
				zap.Int64("event_id", e.ID),
				zap.Error(err))
			continue
		}
		published++

		// Prometheus: lag = observed_at(now) - occurred_at
		lag := now.Sub(e.OccurredAt).Seconds()
		if lag < 0 {
			lag = 0
		}
		OutboxLagSeconds.Observe(lag)
		OutboxEventsPublishedTotal.Inc()
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("events: commit: %w", err)
	}
	r.log.Debug("outbox relay batch published",
		zap.Int("selected", len(events)),
		zap.Int("published", published))
	return nil
}

// Close 释放 relay 持有的底层 RabbitMQ 客户端。
// 若客户端是共享的且在其他地方关闭，调用方可跳过此方法。
func (r *Relay) Close() error {
	if r.mq != nil {
		return r.mq.Close()
	}
	return nil
}

// IsEventProcessed 查询某消费者是否已成功处理该事件。
func IsEventProcessed(ctx context.Context, eventID int64, consumerID string, db interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}) (bool, error) {
	const q = `SELECT 1 FROM processed_events WHERE event_id = $1 AND consumer_id = $2 LIMIT 1`
	rows, err := db.Query(ctx, q, eventID, consumerID)
	if err != nil {
		return false, fmt.Errorf("events: query processed: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, rows.Err()
}

// --- Global Notifier (对标 Google SRE Book Ch.17: 信号解耦) ---
//
// Relay 运行在 worker 进程；API 进程通过 Redis/DB 与之分离。
// 此处提供一层进程内委托：API 层调用 NotifyNewEvent() 经由全局变量
// 转发至本进程内的 Relay。若未注册 Relay（如 worker 未开启或无 relay），
// 则退化为 no-op，不影响 write 路径的稳定性。
//
// 在单体模式（relay 与 API 同一进程）下，该机制可将 outbox 端到端延迟
// 从 500ms（轮询间隔）降低到 <10ms（notify 即时唤醒）。
//
// RegisterRelay 由Relay 创建方（如 cmd/worker/main.go）调用一次。
var (
	globalRelayMu sync.RWMutex
	globalRelay   *Relay
)

// RegisterRelay 注册全局 Relay 以便 API 层 NotifyNewEvent 委托调用。
func RegisterRelay(r *Relay) {
	globalRelayMu.Lock()
	defer globalRelayMu.Unlock()
	globalRelay = r
}

// UnregisterRelay 解注册（通常在 run teardown 时）。
func UnregisterRelay() {
	globalRelayMu.Lock()
	defer globalRelayMu.Unlock()
	globalRelay = nil
}

// NotifyNewEvent 是 package-level 唤醒入口。
// API 层写入 domain_events 后调用本函数，通知本进程内的 Relay 立即 poll。
// 若 Relay 未注册（如 API 进程无 worker），则为 no-op。
func NotifyNewEvent() {
	globalRelayMu.RLock()
	r := globalRelay
	globalRelayMu.RUnlock()
	if r != nil {
		r.NotifyNewEvent()
	}
}

// EnqueueTask 向 TaskExchange 发布后台任务信封。
// 这是 fire-and-forget 异步工作的轻量路径，不需要 Asynq 完整任务队列语义
// （如自动化规则中的事件驱动反应、实时索引扇出）。
func EnqueueTask(ctx context.Context, client *mq.Client, taskType string, workspaceID int64, payload json.RawMessage) error {
	if client == nil {
		return errors.New("events: mq client is nil")
	}
	return client.Publish(ctx, mq.TaskExchange, "plane.tasks."+taskType, mq.EventEnvelope{
		EventType:   taskType,
		WorkspaceID: workspaceID,
		Payload:     payload,
		OccurredAt:  time.Now(),
		Exchange:    mq.TaskExchange,
		RoutingKey:  "plane.tasks." + taskType,
	})
}

// --------------------------------------------------------------------------
// 遗留消费者辅助（保留给 WebSocket 扇出路径）。
// 原先消费 Redis Streams 的 RealTime relay 仍与 RabbitMQ 并存：
// WS 扇出倾向 Redis P/S 的亚 100ms 延迟，RabbitMQ 负责持久、
// 可重放的事件流。
// --------------------------------------------------------------------------

var _ = sql.ErrNoRows // 保留导入，供后续新增 SQL 辅助时使用
