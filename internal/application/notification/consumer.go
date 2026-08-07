// Package notification — 通知领域事件消费者。
//
// 订阅 RabbitMQ EventExchange，将领域事件转换为站内通知。
// 这是"最小闭环 MVP"的核心链路：
//
//   IssueService/CommentService → Outbox (domain_events)
//     → OutboxRelay → RabbitMQ EventExchange
//     → NotificationConsumer → notifications 表
//     → 前端铃铛组件轮询 /api/v1/notifications/unread-count
//
// 设计决策：
//   - 直接在 EventExchange 上消费，不绕经 TaskExchange（减少一跳延迟）
//   - 每个事件类型对应一个处理函数，新增事件只需扩展 handleEvent
//   - 通知风暴去重：Redis 滑动窗口（默认 5 分钟）内同 issue 多次更新聚合为一条
//     避免看板拖拽/批量操作触发通知风暴
package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// dedupWindow 通知去重窗口。同一收件人 + 同一工作项 + 同一事件类型
// 在窗口期内多次触发只保留一条（若有增量信息则追加到 body）。
const dedupWindow = 5 * time.Minute

// consumer 是通知领域事件的 RabbitMQ 消费者。
type consumer struct {
	db       *pgxpool.Pool
	rdb      *redis.Client
	log      *zap.Logger
	handlers map[string]eventHandler
}

// eventHandler 将单个领域事件转换为通知写入。
type eventHandler func(ctx context.Context, event mq.EventEnvelope) error

// newConsumer 创建通知事件 consumer，注册所有已知事件处理器。
// rdb 为可选（nil 表示跳过 Redis 去重，适合无 Redis 的测试环境）。
func newConsumer(db *pgxpool.Pool, log *zap.Logger, rdb ...*redis.Client) *consumer {
	var r *redis.Client
	if len(rdb) > 0 {
		r = rdb[0]
	}
	c := &consumer{
		db:       db,
		rdb:      r,
		log:      log,
		handlers: make(map[string]eventHandler),
	}
	// 注册事件处理器（聚合类型.事件类型 → handler）
	c.handlers["issue.created"] = c.handleIssueCreated
	c.handlers["issue.assigned"] = c.handleIssueAssigned
	c.handlers["issue.status_changed"] = c.handleIssueStatusChanged
	c.handlers["comment.created"] = c.handleCommentCreated
	return c
}

// dedupKey 生成去重 key：notif:dedup:{recipient_id}:{event_type}:{entity_type}:{entity_id}
func dedupKey(recipientID int64, eventType EventType, entityType EntityType, entityID int64) string {
	return fmt.Sprintf("notif:dedup:%d:%s:%s:%d", recipientID, eventType, entityType, entityID)
}

// shouldDedup 判断是否应去重。返回 true 表示跳过（去重窗口内已发送过）。
// 使用 Redis SET NX（key 不存在时设置成功→首次，应发送；否则→去重）。
func (c *consumer) shouldDedup(ctx context.Context, recipientID int64, eventType EventType, entityType EntityType, entityID int64) bool {
	if c.rdb == nil {
		return false // 无 Redis 不去重
	}
	// 仅对 status_changed 类高频事件去重（避免信息丢失）
	if eventType != EventIssueStatusChanged {
		return false
	}
	key := dedupKey(recipientID, eventType, entityType, entityID)
	// SET NX with TTL — 首次返回 true（应发送），重复返回 false（应去重）
	set, err := c.rdb.SetNX(ctx, key, time.Now().Unix(), dedupWindow).Result()
	if err != nil {
		c.log.Warn("dedup redis error — allow notification through", zap.Error(err))
		return false // Redis 错误时放行，避免通知丢失
	}
	// set=true 表示 key 刚创建（首次），应发送；set=false 表示已存在（去重窗口内），应跳过
	return !set
}

// HandleEvent 分发领域事件到对应处理器。
// 未知事件类型静默忽略（不代表错误——其他消费者可能关心它们）。
func (c *consumer) HandleEvent(ctx context.Context, event mq.EventEnvelope) error {
	handler, ok := c.handlers[event.EventType]
	if !ok {
		// 非通知相关事件，静默 ACK
		return nil
	}
	return handler(ctx, event)
}

// eventPayload 是通知相关事件的通用 payload 结构。
type eventPayload struct {
	WorkspaceID int64  `json:"workspace_id"`
	ProjectID   int64  `json:"project_id"`
	ActorID     int64  `json:"actor_id"`
	ActorName   string `json:"actor_name"`
	IssueID     int64  `json:"issue_id"`
	Identifier  string `json:"identifier"`
	Name        string `json:"name"`
	AssigneeIDs []int64 `json:"assignee_ids"`
	FromState   string  `json:"from_state"`
	ToState     string  `json:"to_state"`
	CommentID   int64  `json:"comment_id"`
	Content     string  `json:"content"`
}

// handleIssueCreated: 工作项创建 → 通知被分配人。
func (c *consumer) handleIssueCreated(ctx context.Context, event mq.EventEnvelope) error {
	var p eventPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return fmt.Errorf("issue.created payload: %w", err)
	}

	// 获取项目信息用于 action_url
	projectName := c.getProjectName(ctx, p.ProjectID)

	// 为每个被分配人创建通知
	for _, uid := range p.AssigneeIDs {
		if uid == p.ActorID {
			continue // 自己创建的不用通知自己
		}
		title := fmt.Sprintf("%s 创建工作项", p.ActorName)
		body := fmt.Sprintf("[%s] %s", p.Identifier, p.Name)
		actionURL := fmt.Sprintf("/projects/%d/issues/%d", p.ProjectID, p.IssueID)
		if err := c.create(ctx, CreateNotificationInput{
			WorkspaceID: p.WorkspaceID,
			RecipientID: uid,
			EventType:   EventIssueCreated,
			EntityType:  EntityIssue,
			EntityID:    p.IssueID,
			Title:       title,
			Body:        body,
			ActionURL:   actionURL,
			ActorID:     &p.ActorID,
			ActorName:   p.ActorName,
			Channel:     ChannelInApp,
			Payload:     event.Payload,
		}); err != nil {
			c.log.Warn("failed to create notification for issue.assignee",
				zap.Int64("user", uid), zap.Error(err))
		}
	}
	_ = projectName
	return nil
}

// handleIssueAssigned: 工作项分配 → 通知新分配人。
func (c *consumer) handleIssueAssigned(ctx context.Context, event mq.EventEnvelope) error {
	var p eventPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return fmt.Errorf("issue.assigned payload: %w", err)
	}

	actorID := p.ActorID
	title := fmt.Sprintf("%s 将工作项分配给你", p.ActorName)
	body := fmt.Sprintf("[%s] %s", p.Identifier, p.Name)
	actionURL := fmt.Sprintf("/projects/%d/issues/%d", p.ProjectID, p.IssueID)

	for _, uid := range p.AssigneeIDs {
		if uid == p.ActorID {
			continue
		}
		if err := c.create(ctx, CreateNotificationInput{
			WorkspaceID: p.WorkspaceID,
			RecipientID: uid,
			EventType:   EventIssueAssigned,
			EntityType:  EntityIssue,
			EntityID:    p.IssueID,
			Title:       title,
			Body:        body,
			ActionURL:   actionURL,
			ActorID:     &actorID,
			ActorName:   p.ActorName,
			Channel:     ChannelInApp,
			Payload:     event.Payload,
		}); err != nil {
			c.log.Warn("failed to create notification for issue.assigned",
				zap.Int64("user", uid), zap.Error(err))
		}
	}
	return nil
}

// handleIssueStatusChanged: 状态变更 → 通知关注人（创建者 + 分配者）。
//
// 风暴去重：5 分钟窗口内同一工作项的多次 status_changed 聚合为一条，
// 避免批量操作/看板拖拽触发通知风暴压垮用户收件箱。
func (c *consumer) handleIssueStatusChanged(ctx context.Context, event mq.EventEnvelope) error {
	var p eventPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return fmt.Errorf("issue.status_changed payload: %w", err)
	}

	title := fmt.Sprintf("%s 变更了工作项状态: %s → %s", p.ActorName, p.FromState, p.ToState)
	body := fmt.Sprintf("[%s] %s", p.Identifier, p.Name)
	actionURL := fmt.Sprintf("/projects/%d/issues/%d", p.ProjectID, p.IssueID)

	// 通知所有相关人员（按 user 去重 + 发送风暴去重）
	recipients := make(map[int64]bool)
	for _, uid := range p.AssigneeIDs {
		recipients[uid] = true
	}

	var dedupSkipped int
	for uid := range recipients {
		if uid == p.ActorID {
			continue
		}
		// 风暴去重：5 分钟内同 issue 多次 status_changed 跳过
		if c.shouldDedup(ctx, uid, EventIssueStatusChanged, EntityIssue, p.IssueID) {
			dedupSkipped++
			continue
		}
		actorID := p.ActorID
		if err := c.create(ctx, CreateNotificationInput{
			WorkspaceID: p.WorkspaceID,
			RecipientID: uid,
			EventType:   EventIssueStatusChanged,
			EntityType:  EntityIssue,
			EntityID:    p.IssueID,
			Title:       title,
			Body:        body,
			ActionURL:   actionURL,
			ActorID:     &actorID,
			ActorName:   p.ActorName,
			Channel:     ChannelInApp,
			Payload:     event.Payload,
		}); err != nil {
			c.log.Warn("failed to create notification for issue.status_changed",
				zap.Int64("user", uid), zap.Error(err))
		}
	}
	if dedupSkipped > 0 {
		c.log.Info("notification dedup: skipped status_changed",
			zap.Int64("issue_id", p.IssueID), zap.Int("skipped", dedupSkipped))
	}
	return nil
}

// handleCommentCreated: 新评论 → 通知工作项关注人。
func (c *consumer) handleCommentCreated(ctx context.Context, event mq.EventEnvelope) error {
	var p eventPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return fmt.Errorf("comment.created payload: %w", err)
	}

	title := fmt.Sprintf("%s 评论了工作项", p.ActorName)
	body := p.Content
	if len(body) > 120 {
		body = body[:120] + "..."
	}
	actionURL := fmt.Sprintf("/projects/%d/issues/%d#comment-%d", p.ProjectID, p.IssueID, p.CommentID)

	recipients := make(map[int64]bool)
	for _, uid := range p.AssigneeIDs {
		recipients[uid] = true
	}

	for uid := range recipients {
		if uid == p.ActorID {
			continue
		}
		actorID := p.ActorID
		if err := c.create(ctx, CreateNotificationInput{
			WorkspaceID: p.WorkspaceID,
			RecipientID: uid,
			EventType:   EventCommentCreated,
			EntityType:  EntityIssue,
			EntityID:    p.IssueID,
			Title:       title,
			Body:        body,
			ActionURL:   actionURL,
			ActorID:     &actorID,
			ActorName:   p.ActorName,
			Channel:     ChannelInApp,
			Payload:     event.Payload,
		}); err != nil {
			c.log.Warn("failed to create notification for comment.created",
				zap.Int64("user", uid), zap.Error(err))
		}
	}
	return nil
}

// create 写入一条通知记录,返回创建的通知(含 id/recipient_id 供下游入投递记录)。
func (c *consumer) create(ctx context.Context, input CreateNotificationInput) (*Notification, error) {
	svc := c.svc
	return svc.Create(ctx, input)
}

// enqueueDeliveries 为刚创建的通知,按收件人订阅的非 in_app 渠道写入 notification_deliveries 待投递记录。
// 写入在同一事务内完成: 设置 app.workspace_id 以通过偏好表 RLS,然后读偏好、生成投递记录。
// 任何失败只打 warn 不阻塞主流程(in-app通知已落库,投递为尽力而为)。
func (c *consumer) enqueueDeliveries(ctx context.Context, wsID, recipientID, notifID int64, eventType EventType) {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		c.log.Warn("enqueue deliveries: begin tx failed", zap.Error(err))
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 租户上下文,使 preference 行级安全策略(app.workspace_id)生效。
	if _, err := tx.Exec(ctx,
		"SELECT set_config('app.workspace_id', $1, true)",
		strconv.FormatInt(wsID, 10),
	); err != nil {
		c.log.Warn("enqueue deliveries: set_config failed", zap.Error(err))
		return
	}

	pref, err := c.svc.fetchPreferenceTx(ctx, tx, wsID, recipientID)
	if err != nil && !errorsIsNoRows(err) {
		c.log.Warn("enqueue deliveries: read preference failed", zap.Error(err))
		return
	}
	if pref == nil || !pref.IsEnabled || !c.svc.isEventEnabledPref(pref, eventType) {
		_ = tx.Rollback(ctx)
		return
	}

	now := time.Now()
	for _, chStr := range pref.Channels {
		ch := Channel(chStr)
		if ch == ChannelInApp {
			continue
		}
		// DND 窗口内跳过非站内渠道(避免夜间打扰)。
		if pref.DNDEnabled && inDNDWindow(pref.DNDStart, pref.DNDEnd) {
			continue
		}
		recipient, def := c.resolveRecipient(ctx, ch, recipientID)
		if !def {
			continue
		}
		if _, err := tx.Exec(ctx,
			"INSERT INTO notification_deliveries (notification_id,channel,status,recipient,created_at) VALUES ($1,$2,'pending',$3,$4)",
			notifID, string(ch), recipient, now,
		); err != nil {
			c.log.Warn("enqueue deliveries: insert failed",
				zap.String("channel", string(ch)), zap.Error(err))
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		c.log.Warn("enqueue deliveries: commit failed", zap.Error(err))
	}
}

// resolveRecipient 解析渠道收件人地址;def=false 表示该渠道无法解析收件人(跳过)。
func (c *consumer) resolveRecipient(ctx context.Context, ch Channel, userID int64) (string, bool) {
	switch ch {
	case ChannelEmail:
		// 邮箱需查 users 表,为尽力而为:失败返回 false 跳过。
		return c.userEmail(ctx, userID)
	case ChannelWeCom, ChannelDingTalk, ChannelFeishu:
		// IM webhook 由 dispatcher 按渠道从环境变量读取,收件人填渠道标识占位。
		return string(ch), true
	default:
		return "", false
	}
}

func (c *consumer) userEmail(ctx context.Context, userID int64) (string, bool) {
	var email string
	if err := c.db.QueryRow(ctx, "SELECT email FROM users WHERE id = $1 AND deleted_at IS NULL", userID).Scan(&email); err != nil {
		return "", false
	}
	if email == "" {
		return "", false
	}
	return email, true
}

func errorsIsNoRows(err error) bool {
	return err != nil && err.Error() == "no rows in result set" || isNoRows(err)
}

// getProjectName 获取项目名称。失败时返回空字符串（非阻塞）。
func (c *consumer) getProjectName(ctx context.Context, projectID int64) string {
	var name string
	err := c.db.QueryRow(ctx, `SELECT name FROM projects WHERE id = $1`, projectID).Scan(&name)
	if err != nil {
		return ""
	}
	return name
}

