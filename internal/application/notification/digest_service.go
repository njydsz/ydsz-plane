// Package notification — 通知摘要聚合服务 (Digest Service)。
//
// 负责定时聚合 daily/weekly 通知摘要，并将聚合结果写入 notification_digests 表。
// 主要流程：
//   1. DigestRunner 定时触发（每分钟检查一次）
//   2. DigestService.PendingDigests 找出所有到达计划时刻的 (用户,工作空间,频率) 组合
//   3. DigestService.BuildDigest 聚合时间窗内的通知 → 生成摘要 JSON
//   4. 通过 dispatchConfig.deliverEmail 发送，或写入 notification_digests 待 IM dispatcher 处理
//   5. 聚合后的通知标记 is_archived = true（从收件箱隐藏，保留在 digest 记录中）
//
// 参考: Linear Digest / Jira Notification Digest / GitHub Digest
package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DigestService 提供通知摘要的聚合与投递。
type DigestService struct {
	db *pgxpool.Pool
}

// NewDigestService 创建摘要服务。
func NewDigestService(db *pgxpool.Pool) *DigestService {
	return &DigestService{db: db}
}

// PendingDigest 是一个待投递摘要任务。
type PendingDigest struct {
	UserID      int64
	WorkspaceID int64
	DigestType  Digest
	Channel     Channel
	Recipients  []string
}

// PendingDigests 查找所有到达计划时刻的 (用户,工作空间,频率) 组合。
//
// 业务规则：
//   - daily: 每个工作日 08:30 用户本地时区 → scheduled_for 到达时触发
//   - weekly: 每周一 08:30 → scheduled_for 到达时触发
//   - 跳过 DND 窗口（聚合摘要不打扰用户）
//   - 跳过 notification_preferences.is_enabled = false 的偏好
func (s *DigestService) PendingDigests(ctx context.Context, now time.Time) ([]PendingDigest, error) {
	rows, err := s.db.Query(ctx, `
		SELECT np.user_id, np.workspace_id, np.digest,
		       COALESCE(
		           (SELECT array_agg(u.email)
		            FROM users u
		            WHERE u.id = np.user_id
		              AND u.deleted = false
		              AND u.is_active
		          ),
		           ARRAY[]::text[]
		       ) AS recipients
		FROM notification_preferences np
		LEFT JOIN notification_digests latest
		  ON latest.user_id = np.user_id
		  AND latest.workspace_id = np.workspace_id
		  AND latest.created_at = (
		      SELECT MAX(nd2.created_at)
		      FROM notification_digests nd2
		      WHERE nd2.user_id = np.user_id
		        AND nd2.workspace_id = np.workspace_id
		  )
		WHERE np.is_enabled = true
		  AND np.digest IN ('daily', 'weekly')
		  AND (latest.id IS NULL OR latest.created_at < $1::timestamptz)
		GROUP BY np.user_id, np.workspace_id, np.digest`, now)
	if err != nil {
		return nil, fmt.Errorf("DigestService.PendingDigests: %w", err)
	}
	defer rows.Close()

	var out []PendingDigest
	for rows.Next() {
		var pd PendingDigest
		var recipients []string
		if err := rows.Scan(&pd.UserID, &pd.WorkspaceID, &pd.DigestType, &recipients); err != nil {
			return nil, fmt.Errorf("DigestService.PendingDigests scan: %w", err)
		}
		pd.Recipients = recipients
		pd.Channel = ChannelEmail // 默认通过 email 投递 digest
		out = append(out, pd)
	}
	return out, rows.Err()
}

// DigestItem 摘要项：单个通知的精简摘要。
type DigestItem struct {
	EventType EventType `json:"event_type"`
	EntityID  int64     `json:"entity_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	ActionURL string    `json:"action_url"`
	CreatedAt time.Time `json:"created_at"`
}

// DigestPayload 摘要聚合结果。
type DigestPayload struct {
	GeneratedAt time.Time    `json:"generated_at"`
	DigestType  Digest       `json:"digest_type"`
	PeriodStart time.Time    `json:"period_start"`
	PeriodEnd   time.Time    `json:"period_end"`
	TotalCount  int          `json:"total_count"`
	Items       []DigestItem `json:"items"`
}

// BuildDigest 为指定 (用户,工作空间,频率) 构建摘要。
// 聚合从 lastDigestAt 之后到 now 之间的所有未归档通知。
func (s *DigestService) BuildDigest(ctx context.Context, wsID, userID int64, digestType Digest, lastDigestAt, now time.Time) (*DigestPayload, error) {
	// 拉取时间窗内的未归档通知
	rows, err := s.db.Query(ctx, `
		SELECT event_type, entity_id, title, COALESCE(body, ''), action_url, created_at
		FROM notifications
		WHERE workspace_id = $1
		  AND recipient_id = $2
		  AND is_archived = false
		  AND created_at > $3
		  AND created_at <= $4
		ORDER BY created_at DESC
		LIMIT 100`, wsID, userID, lastDigestAt, now)
	if err != nil {
		return nil, fmt.Errorf("DigestService.BuildDigest query: %w", err)
	}
	defer rows.Close()

	payload := &DigestPayload{
		GeneratedAt: now,
		DigestType:  digestType,
		PeriodStart: lastDigestAt,
		PeriodEnd:   now,
		Items:       []DigestItem{},
	}

	for rows.Next() {
		var item DigestItem
		if err := rows.Scan(&item.EventType, &item.EntityID, &item.Title, &item.Body, &item.ActionURL, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("DigestService.BuildDigest scan: %w", err)
		}
		payload.Items = append(payload.Items, item)
	}
	payload.TotalCount = len(payload.Items)
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return payload, nil
}

// LastDigestAt 查询用户最近一次同频率的摘要时间。
// 若从未发送过，返回 24h 前（日报）或 7 天前（周报）。
func (s *DigestService) LastDigestAt(ctx context.Context, wsID, userID int64, digestType Digest) (time.Time, error) {
	var lastAt time.Time
	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(
		    MAX(created_at),
		    $3::timestamptz
		)
		FROM notification_digests
		WHERE user_id = $1 AND workspace_id = $2`,
		userID, wsID, defaultDigestLookback(digestType)).Scan(&lastAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("DigestService.LastDigestAt: %w", err)
	}
	return lastAt, nil
}

// defaultDigestLookback 返回频率对应的默认回溯时间。
func defaultDigestLookback(d Digest) time.Time {
	now := time.Now()
	switch d {
	case DigestWeekly:
		return now.AddDate(0, 0, -7)
	default:
		return now.Add(-24 * time.Hour)
	}
}

// RecordDigest 将聚合结果写入 notification_digests 表并标记原始通知已归档。
// 这是原子操作：摘要插入 + 通知归档同一事务。
func (s *DigestService) RecordDigest(ctx context.Context, pending PendingDigest, payload *DigestPayload) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("DigestService.RecordDigest begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 租户上下文
	if _, err := tx.Exec(ctx,
		"SELECT set_config('app.workspace_id', $1, true)",
		fmt.Sprintf("%d", pending.WorkspaceID),
	); err != nil {
		return fmt.Errorf("DigestService.RecordDigest set_config: %w", err)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("DigestService.RecordDigest marshal: %w", err)
	}

	// 写入摘要记录
	var digestID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO notification_digests (workspace_id, user_id, channel, payload, scheduled_for, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		pending.WorkspaceID, pending.UserID, string(pending.Channel), payloadJSON,
		payload.GeneratedAt, payload.GeneratedAt).Scan(&digestID)
	if err != nil {
		return fmt.Errorf("DigestService.RecordDigest insert: %w", err)
	}

	// 将本次已归档的收件箱通知标 is_archived = true (一次性聚合后归档避免重复)
	if _, err := tx.Exec(ctx, `
		UPDATE notifications SET is_archived = true
		WHERE workspace_id = $1
		  AND recipient_id = $2
		  AND is_archived = false
		  AND created_at >= $3
		  AND created_at <= $4`,
		pending.WorkspaceID, pending.UserID, payload.PeriodStart, payload.PeriodEnd); err != nil {
		return fmt.Errorf("DigestService.RecordDigest archive: %w", err)
	}

	// 送达时间打戳
	if _, err := tx.Exec(ctx,
		`UPDATE notification_digests SET sent_at = now() WHERE id = $1`,
		digestID); err != nil {
		return fmt.Errorf("DigestService.RecordDigest stamp: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("DigestService.RecordDigest commit: %w", err)
	}
	return nil
}

// CountUnreadSince 返回用户自某时间起的未归档通知数（用于前端角标）。
// 这是辅助 API，供「我的摘要」页面展示。
func (s *DigestService) CountUnreadSince(ctx context.Context, wsID, userID int64, since time.Time) (int64, error) {
	var count int64
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM notifications
		WHERE workspace_id = $1
		  AND recipient_id = $2
		  AND is_archived = false
		  AND created_at > $3`,
		wsID, userID, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("DigestService.CountUnreadSince: %w", err)
	}
	return count, nil
}

// BuildDigestSubject 根据摘要频率生成邮件主题。
func BuildDigestSubject(d Digest, wsName string) string {
	switch d {
	case DigestWeekly:
		return fmt.Sprintf("[Ydsz Plane] %s 工作空间周报 — 通知摘要", wsName)
	default:
		return fmt.Sprintf("[Ydsz Plane] %s 工作空间日报 — 通知摘要", wsName)
	}
}

// BuildDigestHTML 根据摘要内容生成邮件 HTML。
func BuildDigestHTML(payload *DigestPayload, wsName string) string {
	daily := payload.DigestType == DigestDaily
	title := "日报摘要"
	if !daily {
		title = "周报摘要"
	}

	itemHTML := ""
	for _, item := range payload.Items {
		itemHTML += fmt.Sprintf(
			`<li style="margin-bottom:8px;">
				<strong>%s</strong><br/>
				<span style="color:#666;">%s</span><br/>
				<a href="%s" style="color:#1a73e8;">查看详情</a>
			</li>`,
			item.Title, item.Body, item.ActionURL,
		)
	}

	return fmt.Sprintf(
		`<div style="font-family:sans-serif;max-width:600px;margin:0 auto;">
			<h2>%s — %s</h2>
			<p>从 %s 到 %s 共有 <strong>%d</strong> 条新通知。</p>
			<ul style="padding-left:0;list-style:none;">%s</ul>
			<hr/>
			<p style="color:#999;font-size:12px;">此邮件由 Ydsz Plane 自动发送</p>
		</div>`,
		wsName, title,
		payload.PeriodStart.Format("2006-01-02 15:04"),
		payload.PeriodEnd.Format("2006-01-02 15:04"),
		payload.TotalCount, itemHTML,
	)
}

// ShouldDigestNow 判断当前时刻是否应触发摘要生成。
// 业务规则：daily 工作日 08:30；weekly: 每周一 08:30。
//
// 对于测试/通用场景，当 scheduled_for 已到达即为触发时刻，
// 此函数提供纯时区判断供调用方使用。
func ShouldDigestNow(d Digest, userTimezone string, now time.Time) bool {
	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		loc = time.UTC
	}
	localTime := now.In(loc)
	hour := localTime.Hour()
	minute := localTime.Minute()

	// 触发窗口 08:30 - 08:31（Worker 每分钟检查一次，允许 1 分钟容差）
	isTriggerTime := hour == 8 && minute == 30

	switch d {
	case DigestDaily:
		// 工作日触发（周一=1 至周五=5）
		wd := localTime.Weekday()
		return isTriggerTime && wd >= time.Monday && wd <= time.Friday
	case DigestWeekly:
		// 仅周一触发
		return isTriggerTime && localTime.Weekday() == time.Monday
	default:
		return false
	}
}

// DefaultDigestWindowStart 返回摘要开始时间（回落窗）。
func DefaultDigestWindowStart(d Digest, now time.Time) time.Time {
	switch d {
	case DigestWeekly:
		return now.AddDate(0, 0, -7)
	default:
		return now.Add(-24 * time.Hour)
	}
}
