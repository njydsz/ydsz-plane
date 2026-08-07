-- 0015_webhooks.up.sql — Webhook 注册与投递日志
-- 对标: GitHub Webhooks, GitLab Hooks, Stripe Event Webhooks
BEGIN;

-- webhook 订阅表：工作空间级配置
CREATE TABLE IF NOT EXISTS webhooks (
    id              BIGSERIAL PRIMARY KEY,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,                             -- NULL 表示工作空间级
    name            VARCHAR(100) NOT NULL,
    target_url      TEXT NOT NULL,
    secret          VARCHAR(255) NOT NULL,              -- HMAC-SHA256 签名密钥
    events          TEXT[] NOT NULL DEFAULT '{}',       -- 订阅事件列表；空=全部
    is_active       BOOLEAN NOT NULL DEFAULT true,
    -- 重试与退避状态
    last_error      TEXT,                               -- 最后一次投递失败原因
    last_triggered  TIMESTAMPTZ,                        -- 最后一次触发时间
    last_status     VARCHAR(20),                        -- 'success' | 'failed' | 'unhealthy'
    unhealthy_at    TIMESTAMPTZ,                        -- 标记 unhealthy 的时间
    created_by      BIGINT NOT NULL,                    -- 创建者 user ID
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT chk_target_url_protocol CHECK (target_url ~ '^https?://')
);

CREATE INDEX IF NOT EXISTS idx_webhooks_workspace ON webhooks (workspace_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_project ON webhooks (project_id) WHERE project_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_webhooks_active ON webhooks (workspace_id, is_active) WHERE is_active = true;

-- webhook_logs 投递日志（分区时间范围索引，30 天保留后清理）
CREATE TABLE IF NOT EXISTS webhook_logs (
    id              BIGSERIAL PRIMARY KEY,
    webhook_id      BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    delivery_id     VARCHAR(64) NOT NULL,               -- 投递唯一 ID（X-Ydsz-Delivery）
    event_type      VARCHAR(80) NOT NULL,
    event_id        BIGINT,                             -- 关联 domain_events.id
    -- 请求详情
    request_url     TEXT NOT NULL,
    request_method  VARCHAR(10) NOT NULL DEFAULT 'POST',
    request_headers JSONB,
    request_body    TEXT,
    -- 响应详情
    response_status INTEGER,                            -- HTTP 状态码（NULL 表示请求未到达）
    response_body   TEXT,
    response_headers JSONB,
    -- 投递状态
    status          VARCHAR(20) NOT NULL,               -- 'delivered' | 'failed' | 'retrying'
    attempt         SMALLINT NOT NULL DEFAULT 1,        -- 当前是第几次尝试（1=首次）
    duration_ms     INTEGER,                            -- 耗时（毫秒）
    error           TEXT,                               -- 失败原因
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_webhook FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_webhook_logs_webhook ON webhook_logs (webhook_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_workspace ON webhook_logs (workspace_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_delivery ON webhook_logs (delivery_id);
-- 用于自动清理 30 天前的日志
CREATE INDEX IF NOT EXISTS idx_webhook_logs_occurred ON webhook_logs (occurred_at);

COMMIT;
