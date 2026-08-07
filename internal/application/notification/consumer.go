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
//   - 幂等性由 notifications 表的 (entity_type, entity_id, event_type, recipient_id)
//     组合隐式保证——同一事件重复投递会创建多条记录（可通过去重逻辑优化，
//     MVP 阶段暂不处理）
package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// consumer 是通知领域事件的 RabbitMQ 消费者。
type consumer struct {
	db       *pgxpool.Pool
	log      *zap.Logger
	handlers map[string]eventHandler
}

// eventHandler 将单个领域事件转换为通知写入。
type eventHandler func(ctx context.Context, event mq.EventEnvelope) error

// newConsumer 创建通知事件 consumer，注册所有已知事件处理器。
func newConsumer(db *pgxpool.Pool, log *zap.Logger) *consumer {
	c := &consumer{
		db:       db,
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
func (c *consumer) handleIssueStatusChanged(ctx context.Context, event mq.EventEnvelope) error {
	var p eventPayload
	if err := json.Unmarshal(event.Payload, &p); err != nil {
		return fmt.Errorf("issue.status_changed payload: %w", err)
	}

	title := fmt.Sprintf("%s 变更了工作项状态: %s → %s", p.ActorName, p.FromState, p.ToState)
	body := fmt.Sprintf("[%s] %s", p.Identifier, p.Name)
	actionURL := fmt.Sprintf("/projects/%d/issues/%d", p.ProjectID, p.IssueID)

	// 通知所有相关人员（去重）
	recipients := make(map[int64]bool)
	for _, uid := range p.AssigneeIDs {
		recipients[uid] = true
	}
	// 如果有 issue creator 信息在 payload 中也可以加入，MVP 阶段仅通知 assignees

	for uid := range recipients {
		if uid == p.ActorID {
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

// create 写入一条通知记录。
func (c *consumer) create(ctx context.Context, input CreateNotificationInput) error {
	svc := NewService(c.db)
	_, err := svc.Create(ctx, input)
	return err
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

