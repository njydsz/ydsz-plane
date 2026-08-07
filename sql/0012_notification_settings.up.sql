-- 0012_notification_settings: 多渠道投递记录 + 偏好设置 API (Sprint 9 — M5)
-- notification_preferences 已在 0010 创建，此处仅建投递日志表

-- notification_deliveries: 多通道投递记录（审计 + 重试依据）
CREATE TABLE IF NOT EXISTS notification_deliveries (
    id              BIGSERIAL PRIMARY KEY,
    notification_id BIGINT NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    channel         VARCHAR(32) NOT NULL,         -- in_app/email/wecom/dingtalk/feishu
    status          VARCHAR(16) NOT NULL DEFAULT 'pending',  -- pending/sent/failed/bounced
    recipient       TEXT NOT NULL,                -- 目标地址（邮箱/UID/webhook URL）
    sent_at         TIMESTAMPTZ,
    error_msg       TEXT,
    retry_count     INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_deliveries_notification ON notification_deliveries(notification_id);
CREATE INDEX idx_deliveries_status ON notification_deliveries(status, created_at) WHERE status = 'pending';

ALTER TABLE notification_deliveries ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON notification_deliveries
    USING (notification_id IN (SELECT id FROM notifications WHERE workspace_id = current_setting('app.workspace_id')::bigint));

-- notification_digests: 摘要邮件待发送队列（由 worker 定期聚合）
CREATE TABLE IF NOT EXISTS notification_digests (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    digest_type     VARCHAR(16) NOT NULL,        -- daily/weekly
    notification_ids BIGINT[] NOT NULL,          -- 待聚合的通知 id 列表
    status          VARCHAR(16) NOT NULL DEFAULT 'pending',  -- pending/sent/skipped
    scheduled_for   TIMESTAMPTZ NOT NULL,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, workspace_id, digest_type, status) WHERE status = 'pending'
);
CREATE INDEX idx_digests_pending ON notification_digests(status, scheduled_for) WHERE status = 'pending';

ALTER TABLE notification_digests ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON notification_digests
    USING (workspace_id = current_setting('app.workspace_id')::bigint);
