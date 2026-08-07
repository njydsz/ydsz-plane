// Package notification — 用户通知偏好与多渠道投递。
//
// 参照: Plane / Jira notification settings / Linear notification preferences
package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// --- Preference API ---

// GetUserPreference 获取用户通知偏好（不存在则返回默认值）。
func (s *Service) GetUserPreference(ctx context.Context, wsID, userID int64) (*NotificationPreference, error) {
	var p NotificationPreference
	var eventTypesRaw, channelsRaw []byte
	var dndStart, dndEnd string

	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, workspace_id, event_types, channels, digest,
		       dnd_enabled, dnd_start, dnd_end, is_enabled, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1 AND workspace_id = $2`,
		userID, wsID).Scan(
		&p.ID, &p.UserID, &p.WorkspaceID, &eventTypesRaw, &channelsRaw,
		&p.Digest, &p.DNDEnabled, &dndStart, &dndEnd, &p.IsEnabled,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			// 返回默认偏好
			return defaultPreference(wsID, userID), nil
		}
		return nil, fmt.Errorf("notification.GetUserPreference: %w", err)
	}

	_ = json.Unmarshal(eventTypesRaw, &p.EventTypes)
	_ = json.Unmarshal(channelsRaw, &p.Channels)
	p.DNDStart = dndStart
	p.DNDEnd = dndEnd
	return &p, nil
}

// UpdatePreference upsert 通知偏好。
func (s *Service) UpdatePreference(ctx context.Context, wsID, userID int64, update PreferenceUpdateInput) (*NotificationPreference, error) {
	eventTypesJSON, _ := json.Marshal(update.EventTypes)
	channelsJSON, _ := json.Marshal(update.Channels)

	var p NotificationPreference
	var eventTypesRaw, channelsRaw []byte
	var dndStart, dndEnd string

	err := s.db.QueryRow(ctx, `
		INSERT INTO notification_preferences (user_id, workspace_id, event_types, channels, digest, dnd_enabled, dnd_start, dnd_end, is_enabled)
		VALUES ($1, $2, $3, $4, $5, $6,
		        COALESCE($7::time, '22:00'::time), COALESCE($8::time, '08:00'::time), $9)
		ON CONFLICT (user_id, workspace_id)
		DO UPDATE SET
			event_types = COALESCE(EXCLUDED.event_types, notification_preferences.event_types),
			channels = COALESCE(EXCLUDED.channels, notification_preferences.channels),
			digest = COALESCE(EXCLUDED.digest, notification_preferences.digest),
			dnd_enabled = COALESCE(EXCLUDED.dnd_enabled, notification_preferences.dnd_enabled),
			dnd_start = COALESCE(EXCLUDED.dnd_start, notification_preferences.dnd_start),
			dnd_end = COALESCE(EXCLUDED.dnd_end, notification_preferences.dnd_end),
			is_enabled = COALESCE(EXCLUDED.is_enabled, notification_preferences.is_enabled),
			updated_at = now()
		RETURNING id, user_id, workspace_id, event_types, channels, digest,
		          dnd_enabled, dnd_start, dnd_end, is_enabled, created_at, updated_at`,
		userID, wsID, eventTypesJSON, channelsJSON, update.Digest,
		update.DNDEnabled, update.DNDStart, update.DNDEnd, update.IsEnabled).Scan(
		&p.ID, &p.UserID, &p.WorkspaceID, &eventTypesRaw, &channelsRaw,
		&p.Digest, &p.DNDEnabled, &dndStart, &dndEnd, &p.IsEnabled,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("notification.UpdatePreference: %w", err)
	}
	_ = json.Unmarshal(eventTypesRaw, &p.EventTypes)
	_ = json.Unmarshal(channelsRaw, &p.Channels)
	p.DNDStart = dndStart
	p.DNDEnd = dndEnd
	return &p, nil
}

// IsEventEnabled 检查该事件类型在用户偏好中是否启用。
func (s *Service) IsEventEnabled(ctx context.Context, wsID, userID int64, eventType EventType) bool {
	pref, err := s.GetUserPreference(ctx, wsID, userID)
	if err != nil || !pref.IsEnabled {
		return false
	}
	// 如果 event_types 为空数组 = 全部启用
	if len(pref.EventTypes) == 0 {
		return true
	}
	for _, et := range pref.EventTypes {
		if et == string(eventType) {
			return true
		}
	}
	return false
}

// ShouldSendNow 考虑 DND 窗口和摘要频率，判断是否立即投递。
func (s *Service) ShouldSendNow(ctx context.Context, wsID, userID int64, eventType EventType) (channel Channel, immediate bool) {
	pref, err := s.GetUserPreference(ctx, wsID, userID)
	if err != nil || !pref.IsEnabled {
		return ChannelInApp, false
	}

	// DND 窗口内，非实时通道不投递
	if pref.DNDEnabled && inDNDWindow(pref.DNDStart, pref.DNDEnd) {
		return ChannelInApp, false // 仅入站允许（会聚合成 digest）
	}

	switch pref.Digest {
	case DigestRealtime:
		// per-preference channels
		if len(pref.Channels) > 0 {
			return Channel(pref.Channels[0]), true
		}
		return ChannelInApp, true
	case DigestDaily, DigestWeekly:
		return ChannelInApp, false // 会入 digest 队
	default:
		return ChannelInApp, true
	}
}

// inDNDWindow 判断当前时间是否在免打扰窗口。
func inDNDWindow(start, end string) bool {
	now := time.Now().Format("15:04")
	if start <= end {
		return now >= start && now < end
	}
	// 跨日窗口（如 22:00-08:00）
	return now >= start || now < end
}

// defaultPreference 返回默认通知偏好（全订阅 + 实时）。
func defaultPreference(wsID, userID int64) *NotificationPreference {
	return &NotificationPreference{
		UserID:      userID,
		WorkspaceID: wsID,
		EventTypes:  []string{},     // 空 = 全部
		Channels:    []string{string(ChannelInApp)},
		Digest:      DigestRealtime,
		DNDEnabled:  false,
		DNDStart:    "22:00",
		DNDEnd:      "08:00",
		IsEnabled:   true,
	}
}

// --- Multi-Channel Deliveries ---

// RecordDelivery 创建投递记录。
type DeliveryRecord struct {
	NotificationID int64
	Channel        Channel
	Recipient      string
	Status         string
}

// InsertDelivery 写入投递日志。
func (s *Service) InsertDelivery(ctx context.Context, d DeliveryRecord) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx, `
		INSERT INTO notification_deliveries (notification_id, channel, status, recipient)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		d.NotificationID, d.Channel, d.Status, d.Recipient).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// MarkDeliverySent 更新投递状态为已发送。
func (s *Service) MarkDeliverySent(ctx context.Context, deliveryID int64) error {
	_, err := s.db.Exec(ctx,
		`UPDATE notification_deliveries SET status = 'sent', sent_at = now() WHERE id = $1`,
		deliveryID)
	return err
}

// MarkDeliveryFailed 更新投递状态为失败（含重试计数）。
func (s *Service) MarkDeliveryFailed(ctx context.Context, deliveryID int64, reason string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE notification_deliveries SET status = 'failed', error_msg = $2, retry_count = retry_count + 1 WHERE id = $1`,
		deliveryID, reason)
	return err
}

// --- Digest ---

// EnqueueDigest 将通知加入用户的 digest 队列。
func (s *Service) EnqueueDigest(ctx context.Context, notifID, wsID, userID int64, digestType Digest, scheduled time.Time) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO notification_digests (user_id, workspace_id, digest_type, notification_ids, scheduled_for)
		VALUES ($1, $2, $3, ARRAY[$4], $5)
		ON CONFLICT (user_id, workspace_id, digest_type, status) WHERE status = 'pending'
		DO UPDATE SET notification_ids = array_append(notification_digests.notification_ids, $4)`,
		userID, wsID, digestType, notifID, scheduled)
	return err
}

// --- PreferenceUpdateInput ---

// PreferenceUpdateInput 更新偏好的入参。
type PreferenceUpdateInput struct {
	EventTypes []string
	Channels   []string
	Digest     string
	DNDEnabled bool
	DNDStart   *string
	DNDEnd     *string
	IsEnabled  bool
}
