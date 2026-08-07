// Package notification 通知域应用服务：提供站内通知的创建、
// 查询、分页与已读状态管理。
package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 提供通知领域应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建通知服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- Notification CRUD ---

// ListInput 通知列表查询参数。
type ListInput struct {
	WorkspaceID int64
	RecipientID int64
	IsRead      *bool
	IsArchived  *bool
	EventType   *string
	/** 按创建时间过滤，仅返回此时间戳之后的通知（用于 WS 断线重连补偿） */
	Since       *int64
	Limit       int
	Offset      int
}

// ListResult 通知列表结果。
type ListResult struct {
	Items []Notification `json:"items"`
	Total int64          `json:"total"`
}

// Create 创建一条通知。
func (s *Service) Create(ctx context.Context, input CreateNotificationInput) (*Notification, error) {
	if input.Channel == "" {
		input.Channel = ChannelInApp
	}
	if input.Payload == nil {
		input.Payload = json.RawMessage("{}")
	}

	var n Notification
	err := s.db.QueryRow(ctx, `
		INSERT INTO notifications
			(workspace_id, recipient_id, event_type, entity_type, entity_id,
			 title, body, action_url, actor_id, actor_name, channel, payload, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW())
		RETURNING id, workspace_id, recipient_id, event_type, entity_type, entity_id,
			title, body, action_url, actor_id, actor_name, is_read, is_archived,
			read_at, channel, payload, created_at`,
		input.WorkspaceID, input.RecipientID, input.EventType, input.EntityType, input.EntityID,
		input.Title, input.Body, input.ActionURL, input.ActorID, input.ActorName,
		input.Channel, input.Payload,
	).Scan(
		&n.ID, &n.WorkspaceID, &n.RecipientID, &n.EventType, &n.EntityType, &n.EntityID,
		&n.Title, &n.Body, &n.ActionURL, &n.ActorID, &n.ActorName,
		&n.IsRead, &n.IsArchived, &n.ReadAt, &n.Channel, &n.Payload, &n.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("notification.Create: %w", err)
	}
	return &n, nil
}

// CreateBatch 批量创建通知（一个事件 → 多个收件人）。
func (s *Service) CreateBatch(ctx context.Context, inputs []CreateNotificationInput) (int, error) {
	if len(inputs) == 0 {
		return 0, nil
	}

	// 使用 COPY 协议批量写入，比逐条 INSERT 效率高 10-100 倍
	rows := make([][]interface{}, len(inputs))
	for i, in := range inputs {
		if in.Channel == "" {
			in.Channel = ChannelInApp
		}
		if in.Payload == nil {
			in.Payload = json.RawMessage("{}")
		}
		rows[i] = []interface{}{
			in.WorkspaceID, in.RecipientID, in.EventType, in.EntityType, in.EntityID,
			in.Title, in.Body, in.ActionURL, in.ActorID, in.ActorName, in.Channel, in.Payload,
		}
	}

	_, err := s.db.CopyFrom(
		ctx,
		pgx.Identifier{"notifications"},
		[]string{"workspace_id", "recipient_id", "event_type", "entity_type", "entity_id",
			"title", "body", "action_url", "actor_id", "actor_name", "channel", "payload"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return 0, fmt.Errorf("notification.CreateBatch: %w", err)
	}
	return len(inputs), nil
}

// List 查询通知列表（按创建时间倒序）。
func (s *Service) List(ctx context.Context, input ListInput) (*ListResult, error) {
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	var args []interface{}
	argIdx := 1
	where := "recipient_id = $1 AND workspace_id = $2 AND is_archived = false"
	whereTotal := "recipient_id = $1 AND workspace_id = $2 AND is_archived = false"
	args = append(args, input.RecipientID, input.WorkspaceID)

	if input.IsRead != nil {
		cond := fmt.Sprintf(" AND is_read = $%d", argIdx+1)
		where += cond
		whereTotal += cond
		args = append(args, *input.IsRead)
		argIdx++
	}
	if input.EventType != nil {
		cond := fmt.Sprintf(" AND event_type = $%d", argIdx+1)
		where += cond
		whereTotal += cond
		args = append(args, *input.EventType)
		argIdx++
	}
	if input.Since != nil {
		cond := fmt.Sprintf(" AND created_at > to_timestamp($%d::bigint / 1000.0)", argIdx+1)
		where += cond
		whereTotal += cond
		args = append(args, *input.Since)
		argIdx++
	}

	// 计数
	var total int64
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := s.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM notifications WHERE "+whereTotal, countArgs...,
	).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("notification.List count: %w", err)
	}

	// 查询
	rows, err := s.db.Query(ctx,
		fmt.Sprintf(`SELECT id, workspace_id, recipient_id, event_type, entity_type, entity_id,
			title, body, action_url, actor_id, actor_name, is_read, is_archived,
			read_at, channel, payload, created_at
		FROM notifications WHERE %s
		ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2),
		append(args, input.Limit, input.Offset)...,
	)
	if err != nil {
		return nil, fmt.Errorf("notification.List: %w", err)
	}
	defer rows.Close()

	var items []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(
			&n.ID, &n.WorkspaceID, &n.RecipientID, &n.EventType, &n.EntityType, &n.EntityID,
			&n.Title, &n.Body, &n.ActionURL, &n.ActorID, &n.ActorName,
			&n.IsRead, &n.IsArchived, &n.ReadAt, &n.Channel, &n.Payload, &n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("notification.List scan: %w", err)
		}
		items = append(items, n)
	}

	return &ListResult{Items: items, Total: total}, nil
}

// MarkRead 标记单条通知为已读。
func (s *Service) MarkRead(ctx context.Context, notificationID, recipientID int64) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE notifications SET is_read = true, read_at = NOW()
		 WHERE id = $1 AND recipient_id = $2 AND is_read = false`,
		notificationID, recipientID)
	if err != nil {
		return fmt.Errorf("notification.MarkRead: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("NOTIFICATION.NOT_FOUND", "通知不存在或已读")
	}
	return nil
}

// MarkAllRead 将所有未读通知标记为已读。
func (s *Service) MarkAllRead(ctx context.Context, workspaceID, recipientID int64) (int64, error) {
	tag, err := s.db.Exec(ctx,
		`UPDATE notifications SET is_read = true, read_at = NOW()
		 WHERE workspace_id = $1 AND recipient_id = $2 AND is_read = false`,
		workspaceID, recipientID)
	if err != nil {
		return 0, fmt.Errorf("notification.MarkAllRead: %w", err)
	}
	return tag.RowsAffected(), nil
}

// Archive 归档通知。
func (s *Service) Archive(ctx context.Context, notificationID, recipientID int64) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE notifications SET is_archived = true
		 WHERE id = $1 AND recipient_id = $2`,
		notificationID, recipientID)
	if err != nil {
		return fmt.Errorf("notification.Archive: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("NOTIFICATION.NOT_FOUND", "通知不存在")
	}
	return nil
}

// UnreadCount 获取未读通知数量。
func (s *Service) UnreadCount(ctx context.Context, workspaceID, recipientID int64) (int64, error) {
	var count int64
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications
		 WHERE workspace_id = $1 AND recipient_id = $2 AND is_read = false AND is_archived = false`,
		workspaceID, recipientID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("notification.UnreadCount: %w", err)
	}
	return count, nil
}

// CleanupArchived 清理 90 天前的已归档通知。
func (s *Service) CleanupArchived(ctx context.Context) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -90)
	tag, err := s.db.Exec(ctx,
		`DELETE FROM notifications WHERE is_archived = true AND created_at < $1`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("notification.CleanupArchived: %w", err)
	}
	return tag.RowsAffected(), nil
}

// --- Recipient Resolution ---

// RecipientResolver 收件人解析器。
// 根据事件类型和上下文决定通知谁。
type RecipientResolver interface {
	// ResolveRecipients 返回应收到通知的用户 ID 列表。
	ResolveRecipients(ctx context.Context, eventType EventType, entityType EntityType, entityID int64) ([]int64, error)
}

// DefaultRecipientResolver 默认解析器：返回指定收件人列表。
type DefaultRecipientResolver struct {
	Recipients []int64
}

// ResolveRecipients 实现 RecipientResolver。
func (r *DefaultRecipientResolver) ResolveRecipients(ctx context.Context, eventType EventType, entityType EntityType, entityID int64) ([]int64, error) {
	return r.Recipients, nil
}
