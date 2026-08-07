-- 0010_notifications: 通知系统
-- 表: notifications, notification_preferences
-- 支持站内通知（铃铛）+ 后续邮件/IM 通道扩展

-- -----------------------------------------------------------------
-- notifications: 通知主表
-- -----------------------------------------------------------------
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

-- 注释
COMMENT ON TABLE notifications IS '通知消息表（站内铃铛+多渠道预留）';
COMMENT ON COLUMN notifications.event_type IS '事件类型: issue.created/issue.assigned/issue.status_changed/comment.created/sprint.started/sprint.completed/version.released/member.added';
COMMENT ON COLUMN notifications.entity_type IS '关联对象类型: issue/sprint/version/project/workspace/comment';
COMMENT ON COLUMN notifications.channel IS '通知渠道: in_app(站内)/email/sms/wecom/dingtalk/feishu';

-- -----------------------------------------------------------------
-- notification_preferences: 用户通知偏好（Phase B，先建表占位）
-- -----------------------------------------------------------------
CREATE TABLE IF NOT EXISTS notification_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    -- 按事件类型开关
    event_types JSONB NOT NULL DEFAULT '[]',   -- ["issue.created","issue.assigned","comment.created",...]
    -- 按渠道
    channels JSONB NOT NULL DEFAULT '["in_app"]',
    -- 摘要频率: realtime/daily/weekly/off
    digest VARCHAR(16) NOT NULL DEFAULT 'realtime',
    -- 免打扰窗口
    dnd_enabled BOOLEAN NOT NULL DEFAULT false,
    dnd_start TIME DEFAULT '22:00',
    dnd_end TIME DEFAULT '08:00',
    -- 是否启用
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(user_id, workspace_id)
);

ALTER TABLE notification_preferences ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS pref_isolation ON notification_preferences;
CREATE POLICY pref_isolation ON notification_preferences
    USING (workspace_id = current_setting('app.workspace_id')::bigint);

COMMENT ON TABLE notification_preferences IS '用户通知偏好配置';
