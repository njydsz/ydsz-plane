// Command worker 运行异步处理任务：
//   - Outbox Relay   —— 轮询 PostgreSQL outbox 表，将领域事件发布到 RabbitMQ EventExchange。
//   - Task Worker    —— 消费任务队列（通知、索引、webhook、自动化、积压任务）。
//
// worker 进程持有全部 RabbitMQ 消费者（outbox relay 与 task worker）。
// Redis 仅用于 API 层（缓存、限流、分布式锁、WebSocket 扇出），worker 不依赖它。
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/application/automation"
	"github.com/njydsz/ydsz-plane/internal/application/metrics"
	notifApp "github.com/njydsz/ydsz-plane/internal/application/notification"
	"github.com/njydsz/ydsz-plane/internal/application/search"
	"github.com/njydsz/ydsz-plane/internal/application/sprint"
	"github.com/njydsz/ydsz-plane/internal/application/webhook"
	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/events"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/mail"
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

	worker.Register("notifications.send", func(ctx context.Context, task mq.Task) error {
		// 兼容直接通过 TaskExchange 投递的通知任务
		var input notifApp.CreateNotificationInput
		if err := json.Unmarshal(task.Payload, &input); err != nil {
			log.Warn("task: notifications.send: bad payload",
				zap.String("id", task.ID), zap.Error(err))
			return nil // 格式错误不重试
		}
		svc := notifApp.NewService(pool.Pool)
		if _, err := svc.Create(ctx, input); err != nil {
			return fmt.Errorf("notifications.send: create: %w", err)
		}
		return nil
	})

	// ----- 通知领域事件消费者 (EventExchange → notifications 表) -----
	//
	// 消费 EventExchange 上的 issue.* 和 comment.* 事件，
	// 根据事件类型与 payload 创建对应的站内通知。
	// 独立于 Task-based 投递通道，直接订阅事件总线以获得更低延迟。
	go notifApp.RunConsumer(ctx, mqClient, pool.Pool, log)

	// ----- Webhook 事件投递链路 -----
	//
	// 1. 事件消费者（EventExchange → Dispatcher）：订阅全部领域事件，
	//    匹配 Webhook 订阅并同步投递；失败投递异步排入 webhook.retry 重试队列。
	// 2. webhook.retry 任务：消费重试队列，按指数退避（1/5/30min，≤3 次）重投。
	//
	// Dispatcher 持有真实 TaskClient（不再是 nil），使重试入队与测试 ping 均可工作。
	webhookSvc := webhook.NewService(pool.Pool)
	webhookTaskClient := mq.NewTaskClient(mqClient, log)
	webhookDispatcher := webhook.NewDispatcher(webhookSvc, webhookTaskClient, log)
	webhookConsumer := webhook.NewConsumer(webhookDispatcher, log)
	go webhook.RunConsumer(ctx, mqClient, webhookDispatcher, log)

	worker.Register("webhook.retry", func(ctx context.Context, task mq.Task) error {
		return webhookConsumer.HandleRetryTask(ctx, task)
	})

	// ----- 搜索索引填充链路 -----
	//
	// search_documents 表是 Service.Search 的数据源，既往只有 issues 有 DB 触发器自动填充，
	// sprints / versions 从未被索引且存量数据无回填。本链路:
	//   1) RunConsumer 订阅 plane.events.#，把 issue.*/sprint.*/version.* 事件增量同步进索引；
	//   2) search.index 任务：显式 upsert/delete 指定文档（可用于运维回填、
	//      迁移后修复等）。
	searchIndexer := search.NewIndexer(pool.Pool)
	searchIndexLog := log.Named("search.index")
	go search.RunConsumer(ctx, mqClient, searchIndexer, searchIndexLog)

	worker.Register("search.index", func(ctx context.Context, task mq.Task) error {
		// payload 约定: {"doc_type":"issue|sprint|version","doc_id":123,"op":"upsert"|"delete","workspace_id":456}
		var payload struct {
			DocType      string `json:"doc_type"`
			DocID        int64  `json:"doc_id"`
			Op           string `json:"op"`
			WorkspaceID  int64  `json:"workspace_id"`
		}
		if err := json.Unmarshal(task.Payload, &payload); err != nil {
			searchIndexLog.Warn("bad payload, skipping",
				zap.String("id", task.ID), zap.Error(err))
			return nil // 格式错误不重试
		}
		switch payload.Op {
		case "delete":
			if payload.WorkspaceID == 0 {
				return fmt.Errorf("search.index delete requires workspace_id")
			}
			return searchIndexer.RemoveDocument(ctx, payload.DocType, payload.WorkspaceID, payload.DocID)
		default:
			// upsert：无 workspace_id 时由 indexer 内部解析（注意 RLS 限制），有则直接走
			switch payload.DocType {
			case "sprint":
				return searchIndexer.SyncSprint(ctx, payload.DocID)
			case "version":
				return searchIndexer.SyncVersion(ctx, payload.DocID)
			default:
				return searchIndexer.SyncIssue(ctx, payload.DocID)
			}
		}
	})
	// automation.evaluate 已由 automation.RunConsumer 直连 EventExchange 驱动（cmd/worker/main.go 约 line 227），
	// 此处不再重复注册 task handler —— automation 引擎基于事件而非任务调度。

	// ----- 通知多渠道分发 Worker -----
	//
	// 每 30s 轮询 notification_deliveries 表中 pending 记录，
	// 按 channel 投递 email / wecom / dingtalk / feishu，失败指数退避重试。
	var mailSvc mail.EmailService
	if cfg.Email.Enabled {
		mailSvc = mail.NewSmtpService(mail.SmtpConfig{
			Host:     cfg.Email.Host,
			Port:     cfg.Email.Port,
			Username: cfg.Email.Username,
			Password: cfg.Email.Password,
			From:     cfg.Email.From,
			UseTLS:   cfg.Email.UseTLS,
		})
	} else {
		mailSvc = mail.NewNoopService(0)
	}
	go notifApp.StartDispatchWorker(ctx, pool.Pool, mailSvc, cfg.Email.AppBaseURL, log)

	// ----- Sprint 每日快照定时任务（幂等） -----
	//
	// 每分钟触发一次，在 00:05 UTC（北京时间 08:05）执行。
	// 使用 SnapshotAllActive，通过 ON CONFLICT (sprint_id, snapshot_date)
	// 保证幂等。容忍单个 sprint 失败 —— 一个坏掉的 sprint 不会阻塞其他 sprint。
	sprintSvc := sprint.NewService(pool.Pool)
	go runDailySnapshotCron(ctx, sprintSvc, log)

	// ----- Automation 规则引擎事件消费者 -----
	//
	// 订阅 EventExchange 的 plane.events.# 路由，对到达的领域事件
	// 执行匹配的自动化规则（触发器-条件-动作）。
	go automation.RunConsumer(ctx, mqClient, pool.Pool, log)

	// ----- Metrics 效能快照定时任务 -----
	//
	// 每天 01:30 UTC（北京时间 09:30）对所有活跃项目执行效能快照聚合：
	// velocity / wip / lead_time_p50 / lead_time_p85 / defect_density / escape_rate / DORA。
	// 经 metric_snapshots 的 ON CONFLICT (workspace_id, project_id, granularity, metric, snapshot_date) 保证幂等。
	metricsSvc := metrics.NewService(pool.Pool)
	go runMetricsSnapshotCron(ctx, metricsSvc, log)

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

// runMetricsSnapshotCron 运行 30 分钟粒度的定时器，在 01:30 UTC 触发效能快照聚合。
// 设计目标:
//   - 幂等：metric_snapshots 的 ON CONFLICT DO UPDATE
//   - 单个项目失败不阻塞其他项目（内部已容错）
//   - 便于接入 Prometheus / 告警（日志含执行数量与耗时）
func runMetricsSnapshotCron(ctx context.Context, svc *metrics.Service, log *zap.Logger) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	var lastRunDate string

	for {
		select {
		case <-ctx.Done():
			log.Info("metrics snapshot cron: shutting down")
			return
		case t := <-ticker.C:
			utc := t.UTC()
			dateKey := utc.Format("2006-01-02")
			// 01:30 UTC（北京 09:30）触发，±15 分钟容忍窗口
			if utc.Hour() == 1 && utc.Minute() >= 30 && utc.Minute() < 45 {
				if dateKey == lastRunDate {
					continue
				}
				lastRunDate = dateKey

				start := time.Now()
				count, err := svc.AggregateDailySnapshots(ctx, dateKey)
				elapsed := time.Since(start)

				if err != nil {
					log.Warn("metrics snapshot cron: failed",
						zap.String("date", dateKey),
						zap.Error(err),
						zap.Duration("elapsed", elapsed))
				} else {
					log.Info("metrics snapshot cron: completed",
						zap.String("date", dateKey),
						zap.Int("snapshots_written", count),
						zap.Duration("elapsed", elapsed))
				}
			}
		}
	}
}
