-- V26.11.03__integration.sql — 集成（Marketplace + 第三方连接器）
--
-- 对标 Plane Integrations / Linear Integrations：
--   - integrations        ：工作空间级集成实例（GitHub / GitLab / Slack / Sentry / Jira）；
--     credentials 为 AES-256-GCM 密文（工作空间派生密钥），绝不明文存储；
--     config 为明文 JSON（repo / channel / project 等）。
--   - integration_events  ：集成事件日志（入站 / 出站），用于同步记录与调试。
--
-- 应用层生成 id（雪花式 BIGINT），两表均使用 deleted 软删除。

-- ---------------------------------------------------------------------------
-- integrations
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS integrations (
    id           BIGINT       PRIMARY KEY,
    workspace_id BIGINT       NOT NULL,
    provider     VARCHAR(32)  NOT NULL,
    name         VARCHAR(200) NOT NULL,
    status       VARCHAR(20)  NOT NULL DEFAULT 'pending',
    -- AES-256-GCM 密文（base64(nonce||ciphertext||tag)）；空串表示未配置凭证。
    credentials  TEXT,
    -- 明文 JSON（repo / channel / project 等），随 provider 而异。
    config       TEXT,
    last_sync_at TIMESTAMPTZ,
    last_error   TEXT,
    created_by   BIGINT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    CONSTRAINT chk_integrations_provider CHECK (provider IN ('github', 'gitlab', 'slack', 'sentry', 'jira')),
    CONSTRAINT chk_integrations_status   CHECK (status IN ('active', 'paused', 'error', 'pending'))
);

CREATE INDEX IF NOT EXISTS idx_integrations_workspace_id ON integrations (workspace_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_integrations_workspace_provider ON integrations (workspace_id, provider) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_integrations_workspace_status ON integrations (workspace_id, status) WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- integration_events
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS integration_events (
    id             BIGINT       PRIMARY KEY,
    integration_id BIGINT       NOT NULL,
    workspace_id   BIGINT       NOT NULL,
    direction      VARCHAR(16)  NOT NULL,
    event_type     VARCHAR(64)  NOT NULL,
    status         VARCHAR(16)  NOT NULL,
    payload        TEXT,
    error          TEXT,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    CONSTRAINT chk_integration_events_direction CHECK (direction IN ('inbound', 'outbound')),
    CONSTRAINT chk_integration_events_status   CHECK (status IN ('success', 'failed', 'ignored'))
);

CREATE INDEX IF NOT EXISTS idx_integration_events_integration_id ON integration_events (integration_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_integration_events_workspace_id ON integration_events (workspace_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_integration_events_lookup
    ON integration_events (workspace_id, integration_id, created_at DESC) WHERE deleted_at IS NULL;

COMMENT ON TABLE integrations IS '集成实例（工作空间级，凭证加密存储）';
COMMENT ON TABLE integration_events IS '集成事件日志（入站 / 出站，用于同步记录与调试）';
