// Package webhook — Webhook 应用服务：订阅管理、投递编排与查询。
//
// 投递流水线：
//  1. EventExchange 事件到达 → WebhookDispatcher 过滤出匹配的订阅。
//  2. 每订阅同步投递（HTTP POST + HMAC-SHA256 签名）。
//  3. 失败走 TaskExchange（webhook.retry 延迟队列）— 退避 1min/5min/30min。
//  4. 连续失败 3 次标记 unhealthy + 通知创建者。
//  5. 日志写入 webhook_logs，30 天后自动清理。
package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 提供 Webhook 订阅管理、投递与查询。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建 Webhook 服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- 订阅 CRUD ---

// CreateInput 创建 Webhook 的入参。
type CreateInput struct {
	WorkspaceID int64
	ProjectID   *int64
	Name        string
	TargetURL   string
	Secret      string
	Events      []string
	CreatedBy   int64
}

// UpdateInput 更新 Webhook 的入参。
type UpdateInput struct {
	Name      *string
	TargetURL *string
	Events    []string
	IsActive  *bool
}

// Create 创建新的 Webhook 订阅。
// 注意：secret 由调用方生成（crypto/rand 32 字节 hex），服务端不保管明文衍化。
func (s *Service) Create(ctx context.Context, input CreateInput) (*Webhook, error) {
	if input.Name == "" {
		return nil, errs.Validation("WEBHOOK.NAME_REQUIRED", "Webhook 名称不能为空")
	}
	if input.TargetURL == "" {
		return nil, errs.Validation("WEBHOOK.URL_REQUIRED", "目标 URL 不能为空")
	}
	if input.Secret == "" {
		return nil, errs.Validation("WEBHOOK.SECRET_REQUIRED", "签名密钥不能为空")
	}
	if len(input.Events) == 0 {
		// 默认订阅全部事件
		input.Events = AllEvents()
	}

	var w Webhook
	var projectID *int64
	if input.ProjectID != nil {
		projectID = input.ProjectID
	}

	err := s.db.QueryRow(ctx, `
		INSERT INTO webhooks (workspace_id, project_id, name, target_url, secret, events, is_active, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, $7, NOW(), NOW())
		RETURNING id, workspace_id, project_id, name, target_url, secret, events, is_active,
			last_error, last_triggered, last_status, unhealthy_at, created_by, created_at, updated_at`,
		input.WorkspaceID, projectID, input.Name, input.TargetURL, input.Secret,
		input.Events, input.CreatedBy,
	).Scan(
		&w.ID, &w.WorkspaceID, &w.ProjectID, &w.Name, &w.TargetURL, &w.Secret,
		&w.Events, &w.IsActive, &w.LastError, &w.LastTriggered, &w.LastStatus,
		&w.UnhealthyAt, &w.CreatedBy, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("webhook.Create: %w", err)
	}
	return &w, nil
}

// GetByID 查询单条 Webhook 订阅。
func (s *Service) GetByID(ctx context.Context, workspaceID, webhookID int64) (*Webhook, error) {
	var w Webhook
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, name, target_url, '' AS secret, events, is_active,
			last_error, last_triggered, last_status, unhealthy_at, created_by, created_at, updated_at
		FROM webhooks WHERE id = $1 AND workspace_id = $2`,
		webhookID, workspaceID,
	).Scan(
		&w.ID, &w.WorkspaceID, &w.ProjectID, &w.Name, &w.TargetURL, &w.Secret,
		&w.Events, &w.IsActive, &w.LastError, &w.LastTriggered, &w.LastStatus,
		&w.UnhealthyAt, &w.CreatedBy, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("WEBHOOK.NOT_FOUND", "Webhook 不存在")
		}
		return nil, fmt.Errorf("webhook.GetByID: %w", err)
	}
	return &w, nil
}

// ListInput 查询参数。
type ListInput struct {
	WorkspaceID int64
	ProjectID   *int64
	Limit       int
	Offset      int
}

// List 列出工作空间内（或某项目内）的 Webhook 订阅。
func (s *Service) List(ctx context.Context, input ListInput) (*ListResult, error) {
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	// 构造动态 WHERE
	var where string
	var args []interface{}
	argIdx := 1
	if input.ProjectID != nil {
		where = fmt.Sprintf("WHERE workspace_id = $%d AND project_id = $%d", argIdx, argIdx+1)
		args = append(args, input.WorkspaceID, *input.ProjectID)
		argIdx += 2
	} else {
		where = fmt.Sprintf("WHERE workspace_id = $%d", argIdx)
		args = append(args, input.WorkspaceID)
		argIdx++
	}

	// 计数
	var total int64
	countArgs := make([]interface{}, argIdx-1)
	copy(countArgs, args)
	if err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM webhooks "+where, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("webhook.List count: %w", err)
	}

	// 查询
	query := fmt.Sprintf(`
		SELECT id, workspace_id, project_id, name, target_url, '' AS secret, events, is_active,
			last_error, last_triggered, last_status, unhealthy_at, created_by, created_at, updated_at
		FROM webhooks %s ORDER BY id DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	rows, err := s.db.Query(ctx, query, append(args, input.Limit, input.Offset)...)
	if err != nil {
		return nil, fmt.Errorf("webhook.List: %w", err)
	}
	defer rows.Close()

	var items []Webhook
	for rows.Next() {
		var w Webhook
		if err := rows.Scan(
			&w.ID, &w.WorkspaceID, &w.ProjectID, &w.Name, &w.TargetURL, &w.Secret,
			&w.Events, &w.IsActive, &w.LastError, &w.LastTriggered, &w.LastStatus,
			&w.UnhealthyAt, &w.CreatedBy, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("webhook.List scan: %w", err)
		}
		items = append(items, w)
	}
	return &ListResult{Items: items, Total: total}, nil
}

// ListResult 列表结果。
type ListResult struct {
	Items []Webhook `json:"items"`
	Total int64     `json:"total"`
}

// Update 更新 Webhook 配置。
func (s *Service) Update(ctx context.Context, workspaceID, webhookID int64, input UpdateInput) (*Webhook, error) {
	// 动态构建 SET 子句
	sets := make([]string, 0, 4)
	args := make([]interface{}, 0, 5)
	argIdx := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *input.Name)
		argIdx++
	}
	if input.TargetURL != nil {
		sets = append(sets, fmt.Sprintf("target_url = $%d", argIdx))
		args = append(args, *input.TargetURL)
		argIdx++
	}
	if input.Events != nil {
		sets = append(sets, fmt.Sprintf("events = $%d", argIdx))
		args = append(args, input.Events)
		argIdx++
	}
	if input.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *input.IsActive)
		argIdx++
	}

	if len(sets) == 0 {
		return s.GetByID(ctx, workspaceID, webhookID)
	}

	sets = append(sets, fmt.Sprintf("updated_at = NOW()"))
	query := fmt.Sprintf("UPDATE webhooks SET %s WHERE id = $%d AND workspace_id = $%d", joinSets(sets), argIdx, argIdx+1)
	args = append(args, webhookID, workspaceID)

	tag, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("webhook.Update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, errs.NotFound("WEBHOOK.NOT_FOUND", "Webhook 不存在")
	}
	return s.GetByID(ctx, workspaceID, webhookID)
}

// Delete 删除 Webhook 订阅。
func (s *Service) Delete(ctx context.Context, workspaceID, webhookID int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM webhooks WHERE id = $1 AND workspace_id = $2`, webhookID, workspaceID)
	if err != nil {
		return fmt.Errorf("webhook.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("WEBHOOK.NOT_FOUND", "Webhook 不存在")
	}
	return nil
}

// --- 匹配查询 ---

// ListActiveForEvent 列出订阅某事件的所有活跃 Webhook。
// dispatcher 使用此方法查找需要投递的订阅。
func (s *Service) ListActiveForEvent(ctx context.Context, workspaceID int64, eventType string) ([]*Webhook, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, name, target_url, secret, events, is_active,
			last_error, last_triggered, last_status, unhealthy_at, created_by, created_at, updated_at
		FROM webhooks
		WHERE workspace_id = $1 AND is_active = true AND $2 = ANY(events)`,
		workspaceID, eventType)
	if err != nil {
		return nil, fmt.Errorf("webhook.ListActiveForEvent: %w", err)
	}
	defer rows.Close()

	var items []*Webhook
	for rows.Next() {
		var w Webhook
		if err := rows.Scan(
			&w.ID, &w.WorkspaceID, &w.ProjectID, &w.Name, &w.TargetURL, &w.Secret,
			&w.Events, &w.IsActive, &w.LastError, &w.LastTriggered, &w.LastStatus,
			&w.UnhealthyAt, &w.CreatedBy, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("webhook.ListActiveForEvent scan: %w", err)
		}
		items = append(items, &w)
	}
	return items, nil
}

// --- 日志查询 ---

// ListLogsInput 日志查询参数。
type ListLogsInput struct {
	WorkspaceID int64
	WebhookID   *int64
	Status      *string
	EventType   *string
	Limit       int
	Offset      int
}

// ListLogsResult 日志查询结果。
type ListLogsResult struct {
	Items []WebhookLog `json:"items"`
	Total int64        `json:"total"`
}

// ListLogs 查询投递日志。
func (s *Service) ListLogs(ctx context.Context, input ListLogsInput) (*ListLogsResult, error) {
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	var args []interface{}
	var where string
	argIdx := 1

	conjunction := func(cond string, val interface{}) {
		prefix := "WHERE"
		if where != "" {
			prefix = "AND"
		}
		where = fmt.Sprintf("%s %s %s $%d", where, prefix, cond, argIdx)
		args = append(args, val)
		argIdx++
	}

	conjunction("workspace_id =", input.WorkspaceID)
	if input.WebhookID != nil {
		conjunction("webhook_id =", *input.WebhookID)
	}
	if input.Status != nil {
		conjunction("status =", *input.Status)
	}
	if input.EventType != nil {
		conjunction("event_type =", *input.EventType)
	}

	var total int64
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	if err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM webhook_logs "+where, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("webhook.ListLogs count: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, webhook_id, workspace_id, delivery_id, event_type, event_id,
			request_url, request_method, request_headers, request_body,
			response_status, response_body, response_headers,
			status, attempt, duration_ms, error, occurred_at
		FROM webhook_logs %s ORDER BY id DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	rows, err := s.db.Query(ctx, query, append(args, input.Limit, input.Offset)...)
	if err != nil {
		return nil, fmt.Errorf("webhook.ListLogs: %w", err)
	}
	defer rows.Close()

	var items []WebhookLog
	for rows.Next() {
		var l WebhookLog
		if err := rows.Scan(
			&l.ID, &l.WebhookID, &l.WorkspaceID, &l.DeliveryID, &l.EventType, &l.EventID,
			&l.RequestURL, &l.RequestMethod, &l.RequestHeaders, &l.RequestBody,
			&l.ResponseStatus, &l.ResponseBody, &l.ResponseHeaders,
			&l.Status, &l.Attempt, &l.DurationMs, &l.Error, &l.OccurredAt,
		); err != nil {
			return nil, fmt.Errorf("webhook.ListLogs scan: %w", err)
		}
		items = append(items, l)
	}
	return &ListLogsResult{Items: items, Total: total}, nil
}

// --- 状态维护 ---

// RecordResult 更新投递结果标记。
func (s *Service) RecordResult(ctx context.Context, webhookID int64, lastStatus string, lastErr string) error {
	lastStatusEnum := lastStatus
	var unhealthyAt interface{}

	if lastStatus == WebhookStatusFailed {
		unhealthyAt = time.Now()
	}

	_, err := s.db.Exec(ctx, `
		UPDATE webhooks SET last_status = $1, last_error = $2, last_triggered = NOW(),
			unhealthy_at = COALESCE($3, unhealthy_at), updated_at = NOW()
		WHERE id = $4`,
		lastStatusEnum, lastErr, unhealthyAt, webhookID)
	if err != nil {
		return fmt.Errorf("webhook.RecordResult: %w", err)
	}
	return nil
}

// SaveLog 写入单条投递日志。
func (s *Service) SaveLog(ctx context.Context, log *WebhookLog) error {
	reqHeaders := json.RawMessage("null")
	if len(log.RequestHeaders) > 0 {
		reqHeaders = log.RequestHeaders
	}
	respBytes := json.RawMessage("null")
	if len(log.ResponseHeaders) > 0 {
		respBytes = log.ResponseHeaders
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO webhook_logs
			(webhook_id, workspace_id, delivery_id, event_type, event_id,
			 request_url, request_method, request_headers, request_body,
			 response_status, response_body, response_headers,
			 status, attempt, duration_ms, error, occurred_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		log.WebhookID, log.WorkspaceID, log.DeliveryID, log.EventType, log.EventID,
		log.RequestURL, log.RequestMethod, reqHeaders, log.RequestBody,
		log.ResponseStatus, log.ResponseBody, respBytes,
		log.Status, log.Attempt, log.DurationMs, log.Error, log.OccurredAt,
	)
	if err != nil {
		return fmt.Errorf("webhook.SaveLog: %w", err)
	}
	return nil
}

// CleanupLogs 删除 30 天前的日志。由后台 Job 调用。
func (s *Service) CleanupLogs(ctx context.Context) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -30)
	tag, err := s.db.Exec(ctx, `DELETE FROM webhook_logs WHERE occurred_at < $1`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("webhook.CleanupLogs: %w", err)
	}
	return tag.RowsAffected(), nil
}

// --- 测试模式 ---

// SendTestPing 触发合成 ping 事件，用于管理页"测试推送"按钮。
// 构造模拟 payload 投递给指定 webhook。
func (s *Service) SendTestPing(ctx context.Context, webhookID, workspaceID int64) (*WebhookLog, error) {
	var w Webhook
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, name, target_url, secret, events, is_active,
			last_error, last_triggered, last_status, unhealthy_at, created_by, created_at, updated_at
		FROM webhooks WHERE id = $1 AND workspace_id = $2`,
		webhookID, workspaceID,
	).Scan(
		&w.ID, &w.WorkspaceID, &w.ProjectID, &w.Name, &w.TargetUR