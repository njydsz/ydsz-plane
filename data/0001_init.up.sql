-- 0001_init: base identity + tenant + outbox (Sprint 1 基座)
-- 参考 docs/architecture/04-数据模型设计.md

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id     UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name  TEXT NOT NULL,
    avatar_url    TEXT,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    timezone      TEXT NOT NULL DEFAULT 'Asia/Shanghai',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE TABLE workspaces (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL,
    logo_url    TEXT,
    timezone    TEXT NOT NULL DEFAULT 'Asia/Shanghai',
    language    TEXT NOT NULL DEFAULT 'zh-CN',
    owner_id    BIGINT NOT NULL REFERENCES users(id),
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','archived')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX uq_workspaces_slug ON workspaces(slug) WHERE status <> 'archived';

CREATE TABLE workspace_members (
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role         TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('owner','admin','member','guest')),
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, user_id)
);

-- 事务型 Outbox（见 docs/architecture/01 ADR-4）
CREATE TABLE domain_events (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id   BIGINT NOT NULL,
    aggregate_type TEXT NOT NULL,
    aggregate_id   BIGINT NOT NULL,
    event_type     TEXT NOT NULL,
    payload        JSONB NOT NULL,
    occurred_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at   TIMESTAMPTZ
);
CREATE INDEX idx_events_unpublished ON domain_events(id) WHERE published_at IS NULL;

-- API 写幂等
CREATE TABLE idempotency_keys (
    key        TEXT PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id),
    response   JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 审计日志（只增不改）
CREATE TABLE audit_logs (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT,
    actor_id     BIGINT REFERENCES users(id),
    action       TEXT NOT NULL,
    target       TEXT,
    detail       JSONB,
    ip           INET,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_logs_ws_time ON audit_logs(workspace_id, created_at DESC);

-- updated_at 自动维护
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_workspaces_updated_at BEFORE UPDATE ON workspaces
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- RLS 模板（首批表启用；后续业务表沿用同一模式，见 04 文档 §3.2）
ALTER TABLE workspace_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE workspace_members FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON workspace_members
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- 应用低权账号（部署时由 DBA/初始化脚本执行；本地开发注释保留）
-- CREATE ROLE ydsz_app LOGIN PASSWORD '***';
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ydsz_app;
