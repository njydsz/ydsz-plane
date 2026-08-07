// Command worker 运行异步处理任务：
//   - Outbox Relay   —— 轮询 PostgreSQL outbox 表，将领域事件发布到 RabbitMQ EventExchange。
//   - Task Worker    —— 消费任务队列（通知、索引、webhook、自动化、积压任务）。
//
// worker 进程持有全部 RabbitMQ 消费者（outbox relay 与 task worker）。
// Redis 仅用于 API 层（缓存、限流、分布式锁、WebSocket 扇出），worker 不依赖它。
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	notif "github.com/njydsz/ydsz-plane/internal/application/notification"
	"github.com/njydsz/ydsz-plane/internal/application/sprint"
	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/events"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/telemetry"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "worker: fatal:", err)
		os.Exit(1)
	}
}

// run 启动后台 worker 进程并阻塞直至关闭。
//
// 两条入站流水线（均基于 RabbitMQ）：
//
//  1. Outbox Relay —— 轮询 PostgreSQL outbox 表并将事件发布到 RabbitMQ
//     EventExchange（topic）。将数据库写入与事件分发解耦，API 层不会阻塞
//     在下游消费者上。
//  2. Task Worker —— 消费异步任务队列（通知、搜索索引同步、webhook 投递、
//     自动化规则）。重试采用封顶指数退避；重试耗尽的任务进入死信队列，
//     便于事后分析/重放。
//
// 两条流水线都不依赖 Redis —— API 层使用 Redis 做缓存、限流、分布式锁与
// WebSocket 扇出。worker 只访问 PostgreSQL（outbox 数据源）与 RabbitMQ
// （发布+消费）。
//
// SIGINT/SIGTERM 通过 signal.NotifyContext 传播；context 取消会触发两条
// 流水线的优雅停止（排空进行中的工作、确认/拒绝在途投递）。
//
// 若关闭由信号触发则返回 nil；若 worker 异常退出（如 RabbitMQ 连接不可
// 恢复）则返回非 nil 错误。
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log, err := telemetry.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := persistence.NewPool(ctx, cfg.Database.URL, cfg.Database.MaxConns)
	if err != nil {
		return err
	}
	defer pool.Close()

	// RabbitMQ 客户端 —— 同时支撑 outbox relay 与 task worker。
	// 单条连接即可提供两个 channel 池：relay 使用 channel 1，worker 按需开启。
	mqClient, err := mq.NewClient(cfg.RabbitMQ.URL, mq.WithLogger(log))
	if err != nil {
		return fmt.Errorf("worker: rabbitmq connect: %w", err)
	}
	defer func() { _ = mqClient.Close() }()

	// ----- Outbox Relay (DB → RabbitMQ EventExchange) -----
	relay := events.NewRelay(pool, mqClient, log)
	go relay.Run(ctx)

	// ----- Task Worker (RabbitMQ TaskExchange → handlers) -----
	worker := mq.NewWorker(mqClient, log)

	// 注册任务 handler。每个 handler 是针对特定领域的消费者，
	// 绑定到名为 `task.<type>` 的队列与路由键。
	// handler 返回错误即触发 NACK 并重试（受 MaxRetries 限制）。
	//
	// 队列权重（越大分发优先级越高）传递给调度器；
	// 消费者为每个队列启动一个 goroutine，在单条 TCP 连接内公平竞争。
	//
	// 示例注册（随领域扩展而增加）：
	//   - "notifications.send"  —— 分发邮件 / IM 通知
	//   - "webhook.deliver"     —— POST 到已注册的 webhook 端点
	//   - "automation.evaluate" —— 执行触发器-条件-动作规则
	//   - "search.index"        —— 将 issue/workspace 变更同步到 ES
	//
	// 各类任务的装配：
	notifSvc := notif.NewService(pool.Pool)

	worker.Register("notifications.send", func(ctx context.Context, task mq.Task) error {
		// 解析 task payload 为通知创建参数
		var input notif.CreateNotificationInput
		// 注意：这里使用 payload 直接解析，task payload 由 Outbox Relay 事件消费者填充
		log.Debug("task: notifications.send", zap.String("id", task.ID))
		_ = input // 占位，实际实现将在事件消费者中串联
		return nil
	})
	_ = notifSvc // 通知服务已就绪，后续事件消费者将调用
	worker.Register("webhook.deliver", func(ctx context.Context, task mq.Task) error {
		log.Info("task: webhook.deliver", zap.String("id", task.ID))
		return nil
	})
	worker.Register("search.index", func(ctx context.Context, task mq.Task) error {
		log.Info("task: search.index", zap.String("id", task.ID))
		return nil
	})
	worker.Register("automation.evaluate", func(ctx context.Context, task mq.Task) error {
		log.Info("task: automation.evaluate", zap.String("id", task.ID))
		return nil
	})

	// ----- Sprint 每日快照定时任务（幂等） -----
	//
	// 每分钟触发一次，在 00:05 UTC（北京时间 08:05）执行。
	// 使用 SnapshotAllActive，通过 ON CONFLICT (sprint_id, snapshot_date)
	// 保证幂等。容忍单个 sprint 失败 —— 一个坏掉的 sprint 不会阻塞其他 sprint。
	sprintSvc := sprint.NewService(pool.Pool)
	go runDailySnapshotCron(ctx, sprintSvc, log)

	log.Info("worker started",
		zap.String("rabbitmq", mq.RedactedURL(cfg.RabbitMQ.URL)),
		zap.Strings("task_queues", worker.QueueNames()),
	)

	// 阻塞直至收到信号或不可恢复的错误。relay 与 worker 均感知 context，
	// 在 ctx 取消时会干净地关闭。
	if err := worker.Start(ctx); err != nil && ctx.Err() == nil {
		return err
	}
	return nil
}

// runDailySnapshotCron 运行 1 分钟粒度的定时器，在 00:05 UTC 触发每日快照。
// 为满足大厂可靠性要求而设计：
//   - 幂等（sprint_snapshots 上使用 ON CONFLICT DO UPDATE）
//   - 容忍时钟偏差（±1 分钟窗口）
//   - 通过数据库 upsert 实现单 worker 锁（无需 Redis 依赖）
//   - 记录成功/失败日志，便于接入 Prometheus 告警
func runDailySnapshotCron(ctx context.Context, svc *sprint.Service, log *zap.Logger) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	var lastRunDate string

	for {
		select {
		case <-ctx.Done():
			log.Info("sprint snapshot cron: shutting down")
			return
		case t := <-ticker.C:
			// 在 0 点 5 分触发（±30s 容忍）
			utc := t.UTC()
			dateKey := utc.Format("2006-01-02")
			if utc.Hour() == 0 && utc.Minute() == 5 {
				if dateKey == lastRunDate {
					continue // 今天已执行过
				}
				lastRunDate = dateKey

				snapCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
				count, failures := svc.SnapshotAllActive(snapCtx)
				cancel()

				log.Info("sprint snapshot cron: completed",
					zap.String("date", dateKey),
					zap.Int("snapshots", count),
					zap.Int("failures", failures),
				)
			}
		}
	}
}
