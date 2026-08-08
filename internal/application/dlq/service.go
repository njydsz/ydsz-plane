// Package dlq — 死信队列（DLQ）监控与管理。
//
// 背景：RabbitMQ 重试耗尽的消息进入 plane.dlx 交换（worker 端），此前仅
// 停留在 MQ 内部，无持久化、无管理界面。dlq_events 表（迁移 0028）提供了
// 元数据存储，但没有任何写入方与查询 API —— 前端 DLQMonitoringView 与后端
// 能力脱节。
//
// 本包补齐闭环：
//   - worker 消费 plane.dlx.queue 并把死信元数据写入 dlq_events（生产者）；
//   - API 提供 /admin/dlq 管理端点（列表/清理/重试，消费者）；
//   - Retry 采用"出站重放"：把死信事件重新插入 domain_events（outbox），
//     由 worker 的 outbox relay 重新发布，触发订阅方重处理 —— 与事件溯源
//     重放语义一致，且 API 进程无需引入 RabbitMQ 依赖。
package dlq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// DLQEvent 死信事件元数据（对齐前端 dlq.ts DLQItem 契约）。
type DLQEvent struct {
	ID          int64           `json:"id"`
	EventID     *int64          `json:"event_id,omitempty"`
	Queue       string          `json:"queue"`
	Exchange    string          `json:"exchange"`
	RoutingKey  string          `json:"routing_key"`
	Payload     json.RawMessage `json:"payload"`
	ErrorReason string          `json:"error_reason"`
	ResolvedAt  *string         `json:"resolved_at,omitempty"`
	CreatedAt   string          `json:"created_at"`
}

// ListOptions 列表查询参数。
type ListOptions struct {
	Offset         int
	Limit          int
	UnresolvedOnly bool
}

// Service 提供 DLQ 管理应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建 DLQ 服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// List 分页查询死信列表。
func (s *Service) List(ctx context.Context, wsID int64, opts ListOptions) ([]DLQEvent, int64, error) {
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 20
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}

	where := "workspace_id = $1"
	args := []any{wsID}
	argID := 2
	if opts.UnresolvedOnly {
		where += fmt.Sprintf(" AND resolved_at IS NULL AND $%d::bool", argID)
		args = append(args, true)
		argID++
	}

	var total int64
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := s.db.QueryRow(ctx,
		fmt.Sprintf("SELECT count(*) FROM dlq_events WHERE %s", where), countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("dlq.list count: %w", err)
	}

	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT id, event_id, queue, exchange, routing_key, payload, error_reason,
		       COALESCE(to_char(resolved_at, 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"'), ''),
		       to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"')
		FROM dlq_events
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d`, where, argID, argID+1),
		append(args, opts.Limit, opts.Offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("dlq.list: %w", err)
	}
	defer rows.Close()

	var items []DLQEvent
	for rows.Next() {
		var e DLQEvent
		var resolvedAt string
		if err := rows.Scan(&e.ID, &e.EventID, &e.Queue, &e.Exchange, &e.RoutingKey,
			&e.Payload, &e.ErrorReason, &resolvedAt, &e.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("dlq.list scan: %w", err)
		}
		if resolvedAt != "" {
			e.ResolvedAt = &resolvedAt
		}
		items = append(items, e)
	}
	return items, total, rows.Err()
}

// Remove 标记单条死信为已解决（resolved）。
func (s *Service) Remove(ctx context.Context, wsID, id int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE dlq_events SET resolved_at = now(), resolved_by = 'admin'
		WHERE id = $1 AND workspace_id = $2 AND resolved_at IS NULL`, id, wsID)
	if err != nil {
		return fmt.Errorf("dlq.remove: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound.WithDetails(errs.FieldDetail{Field: "id", Reason: "死信不存在或已解决"})
	}
	return nil
}

// Cleanup 批量标记死信为已解决：按 ID 列表，或全部未解决。
func (s *Service) Cleanup(ctx context.Context, wsID int64, eventIDs []int64, resolvedAll bool) (int64, error) {
	switch {
	case resolvedAll:
		tag, err := s.db.Exec(ctx, `
			UPDATE dlq_events SET resolved_at = now(), resolved_by = 'admin'
			WHERE workspace_id = $1 AND resolved_at IS NULL`, wsID)
		if err != nil {
			return 0, fmt.Errorf("dlq.cleanup all: %w", err)
		}
		return tag.RowsAffected(), nil
	case len(eventIDs) > 0:
		tag, err := s.db.Exec(ctx, `
			UPDATE dlq_events SET resolved_at = now(), resolved_by = 'admin'
			WHERE workspace_id = $1 AND id = ANY($2) AND resolved_at IS NULL`,
			wsID, eventIDs)
		if err != nil {
			return 0, fmt.Errorf("dlq.cleanup ids: %w", err)
		}
		return tag.RowsAffected(), nil
	}
	return 0, nil
}

// Retry 重试指定死信：把原始事件重新插入 domain_events（outbox 重放）。
//
// 仅支持事件型死信（payload 可解析为 mq.EventEnvelope）；任务型（Task）死信
// 无法经 outbox 重放，返回明确错误提示走人工处理。
func (s *Service) Retry(ctx context.Context, wsID, id int64) error {
	var (
		payload    []byte
		eventID    *int64
		routingKey string
	)
	err := s.db.QueryRow(ctx, `
		SELECT payload, event_id, routing_key
		FROM dlq_events WHERE id = $1 AND workspace_id = $2 AND resolved_at IS NULL`,
		id, wsID).Scan(&payload, &eventID, &routingKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound.WithDetails(errs.FieldDetail{Field: "id", Reason: "死信不存在或已解决"})
		}
		return fmt.Errorf("dlq.retry load: %w", err)
	}

	// 解析事件信封
	var env mq.EventEnvelope
	if err := json.Unmarshal(payload, &env); err != nil || env.EventType == "" {
		return errs.Validation("DLQ.RETRY_NOT_EVENT",
			"仅支持事件型死信重试（任务型消息请人工处理）")
	}
	if env.WorkspaceID != wsID {
		return errs.Validation("DLQ.RETRY_WORKSPACE_MISMATCH", "死信工作空间与当前空间不一致")
	}
	if env.AggregateType == "" {
		env.AggregateType = parseAggregateType(env.EventType)
	}

	// 重新写入 outbox → worker relay 会重新发布 → 订阅方重处理
	if _, err := s.db.Exec(ctx, `
		INSERT INTO domain_events (workspace_id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4, $5)`,
		env.WorkspaceID, env.AggregateType, env.AggregateID, env.EventType, env.Payload); err != nil {
		return fmt.Errorf("dlq.retry reinsert: %w", err)
	}

	// 标记已解决
	if _, err := s.db.Exec(ctx, `
		UPDATE dlq_events SET resolved_at = now(), resolved_by = 'admin-retry'
		WHERE id = $1 AND workspace_id = $2`, id, wsID); err != nil {
		return fmt.Errorf("dlq.retry resolve: %w", err)
	}
	return nil
}

// parseAggregateType 从事件类型推导聚合类型（issue.created → issue）。
func parseAggregateType(eventType string) string {
	// 形如 "issue.status_changed" → "issue"
	for i := 0; i < len(eventType); i++ {
		if eventType[i] == '.' {
			return eventType[:i]
		}
	}
	return ""
}
