-- 0012_notification_settings: 多渠道投递记录 + 偏好设置 API (Sprint 9 — M5)
-- notification_preferences 已在 0010 创建，此处仅建投递日志表
-- notifications 表（前置依赖，后续迁移 0021 也会创建）

CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    recipient_id BIGINT NOT NULL REFERENCES users(id),
    event_type VARCHAR(64) NOT NULL,
    -- 关联对象（多态）
    entity_type VARCHAR(32) NOT NULL,   -- 'issue' | 'sprint' | 'version' | 'project' | 'workspace' | 'comment'
    entity_id BIGINT NOT NULL,
    -- 显示信息
    title VARCHAR(256) NOT NULL,
    body TEXT,
    action_url VARCHAR(512),            -- 点击跳转链接
    -- 参与人信息
    actor_id BIGINT REFERENCES users(id),
    actor_name VARCHAR(128),
    -- 状态
    is_read BOOLEAN NOT NULL DEFAULT false,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    -- 通道（预留: in_app/email/wecom/dingtalk/feishu）
    channel VARCHAR(32) NOT NULL DEFAULT 'in_app',
    -- 元数据
    payload JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引：收件人未读查询（铃铛主查询）
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_unread
    ON notifications(recipient_id, workspace_id, is_read, created_at DESC)
    WHERE is_archived = false;

-- 索引：按事件类型回溯
CREATE INDEX IF NOT EXISTS idx_notifications_entity
    ON notifications(entity_type, entity_id);

-- 索引：清理已归档通知
CREATE INDEX IF NOT EXISTS idx_notifications_archived
    ON notifications(created_at)
    WHERE is_archived = true;

-- RLS
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS notifications_isolation ON notifications;
CREATE POLICY notifications_isolation ON notifications
    USING (workspace_id = current_setting('app.workspace_id')::bigint);

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
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_notification_digests_pending ON notification_digests(user_id, workspace_id, digest_type, status) WHERE status = 'pending';
CREATE INDEX idx_digests_pending ON notification_digests(status, scheduled_for) WHERE status = 'pending';

ALTER TABLE notification_digests ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON notification_digests
    USING (workspace_id = current_setting('app.workspace_id')::bigint);
