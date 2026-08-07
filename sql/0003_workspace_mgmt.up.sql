-- 0003_workspace_mgmt: invitations + api_tokens + password_reset_tokens (Sprint 2 — M1)
-- 参考 docs/architecture/04-数据模型设计.md + 06-权限与安全设计.md

-- ---------------------------------------------------------------
-- invitations: 工作空间成员邀请（邮箱链接、7 天有效、可撤销、批量）
-- ---------------------------------------------------------------
CREATE TABLE invitations (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    inviter_id      BIGINT NOT NULL REFERENCES users(id),
    email           TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('admin','member','guest')),
    token_hash      TEXT NOT NULL UNIQUE,         -- SHA-256(token)；原始 token 只存在于邮件中
    message         TEXT,                         -- 附言（可选）
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','accepted','revoked','expired')),
    expires_at      TIMESTAMPTZ NOT NULL DEFAULT now() + interval '7 days',
    accepted_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_invitations_workspace ON invitations(workspace_id, status) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_invitations_email ON invitations(email) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_invitations_token ON invitations(token_hash);

-- RLS（与 workspace_members 同一模式）
ALTER TABLE invitations ENABLE ROW LEVEL SECURITY;
ALTER TABLE invitations FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON invitations;
CREATE POLICY tenant_isolation ON invitations
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- ---------------------------------------------------------------
-- api_tokens: 个人 API Token（对标 GitHub Personal Access Token）
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS api_tokens (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,                 -- 用户自定义名称（如"CI Deploy Key"）
    token_hash      TEXT NOT NULL UNIQUE,         -- SHA-256 hash；ydz_ 前缀原始值仅创建时返回一次
    scopes          JSONB NOT NULL DEFAULT '["read:workspace"]',
    last_used_at    TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,                  -- NULL = 永不过期
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_api_tokens_user ON api_tokens(user_id) WHERE revoked_at IS NULL;

-- password_reset_tokens 已存在于 0002 迁移，跳过重复创建

-- ---------------------------------------------------------------
-- projects 表（M1 最低可用：支持工作空间下建项目）
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS projects (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    public_id       UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL,
    identifier      TEXT NOT NULL,                 -- 项目前缀（如 YD、INFRA），用于生成工作项编号 YD-123
    description     TEXT,
    network         TEXT NOT NULL DEFAULT 'public' CHECK (network IN ('public','private')),
    icon            TEXT,
    color           TEXT,
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','archived')),
    sort_order      DOUBLE PRECISION NOT NULL DEFAULT 65535,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON projects;
CREATE POLICY tenant_isolation ON projects
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

CREATE INDEX IF NOT EXISTS idx_projects_workspace ON projects(workspace_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_workspace_slug ON projects(workspace_id, slug) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_workspace_identifier ON projects(workspace_id, identifier) WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------
-- updated_at 触发器
-- ---------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_invitations_updated_at ON invitations;
CREATE TRIGGER trg_invitations_updated_at BEFORE UPDATE ON invitations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
DROP TRIGGER IF EXISTS trg_api_tokens_updated_at ON api_tokens;
CREATE TRIGGER trg_api_tokens_updated_at BEFORE UPDATE ON api_tokens
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
DROP TRIGGER IF EXISTS trg_projects_updated_at ON projects;
CREATE TRIGGER trg_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

