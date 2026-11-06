-- V26.11.07__permission_audit.sql — 权限审计日志
--
-- 记录权限变更：谁、何时、对哪个角色、改了什么权限、原因。
-- 用于合规审计与权限回溯。

CREATE TABLE IF NOT EXISTS permission_audit_logs (
    id            BIGINT       PRIMARY KEY,
    workspace_id  BIGINT       NOT NULL,
    -- 操作者
    actor_id      BIGINT       NOT NULL,
    actor_name    VARCHAR(200),
    -- 变更对象
    role_slug     VARCHAR(50)  NOT NULL,
    -- 变更类型：grant（授予）/ revoke（回收）/ replace（全量替换）
    action        VARCHAR(20)  NOT NULL,
    -- 权限点（单个或逗号分隔）
    permissions   TEXT[]       NOT NULL DEFAULT '{}',
    -- 变更前快照（JSON）
    before        JSONB,
    -- 变更后快照（JSON）
    after         JSONB,
    -- 原因/备注
    reason        TEXT,
    -- 客户端信息
    ip            VARCHAR(50),
    user_agent    VARCHAR(500),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_permission_audit_action CHECK (action IN ('grant','revoke','replace'))
);

CREATE INDEX IF NOT EXISTS idx_permission_audit_workspace ON permission_audit_logs (workspace_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_permission_audit_role ON permission_audit_logs (workspace_id, role_slug, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_permission_audit_actor ON permission_audit_logs (workspace_id, actor_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_permission_audit_action ON permission_audit_logs (workspace_id, action);

COMMENT ON TABLE permission_audit_logs IS '权限变更审计日志';
