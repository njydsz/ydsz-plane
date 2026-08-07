-- 0026_sso_oidc: OIDC/SAML 企业统一认证集成
-- 参考: 等保三级 7.1.2 身份鉴别要求 + OWASP ASVS V2.1
--
-- 核心变更:
--   1. users 表: password_hash 允许 NULL（SSO 用户无本地密码），新增 sso_provider/sso_subject
--   2. sso_providers: OIDC/SAML 身份提供商配置表
--   3. sso_links: 用户 SSO 关联表（支持一个用户绑定多个 IdP）
--   4. sso_sessions: SSO 登录会话追踪（防 CSRF state/nonce）

-- -----------------------------------------------------------------
-- users 表: SSO 支持扩展
-- -----------------------------------------------------------------
-- password_hash 允许 NULL：SSO 用户无本地密码
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- sso_provider: 用户首次登录的 SSO 提供商
ALTER TABLE users ADD COLUMN IF NOT EXISTS sso_provider TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS sso_subject  TEXT; -- IdP 中的唯一标识 (sub claim)

-- 唯一约束：同一 IdP 下的 subject 唯一
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_sso_subject
    ON users(sso_provider, sso_subject) WHERE sso_provider IS NOT NULL AND sso_subject IS NOT NULL;

-- -----------------------------------------------------------------
-- sso_providers: 身份提供商配置
-- 支持: OIDC (OpenID Connect) / SAML 2.0
-- -----------------------------------------------------------------
CREATE TABLE sso_providers (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,                          -- 显示名称（如 "企业微信 SSO"）
    provider_type   TEXT NOT NULL CHECK (provider_type IN ('oidc','saml')),

    -- OIDC 配置
    issuer_url      TEXT,                                   -- OIDC Issuer URL
    client_id       TEXT,                                   -- OAuth2 Client ID
    client_secret   TEXT,                                   -- OAuth2 Client Secret (encrypted at rest)
    redirect_uri    TEXT,                                   -- 回调 URL
    auth_url        TEXT,                                   -- Authorization Endpoint
    token_url       TEXT,                                   -- Token Endpoint
    userinfo_url    TEXT,                                   -- UserInfo Endpoint
    jwks_url        TEXT,                                   -- JWKS Endpoint
    scopes          TEXT NOT NULL DEFAULT 'openid email profile', -- 请求的 scopes

    -- SAML 配置
    idp_metadata_url TEXT,                                  -- SAML IdP Metadata URL
    idp_entity_id    TEXT,                                  -- SAML IdP Entity ID
    sp_entity_id     TEXT,                                  -- SP Entity ID
    sp_acs_url       TEXT,                                  -- Assertion Consumer Service URL
    idp_certificate  TEXT,                                  -- IdP X.509 证书 (PEM)

    -- 通用配置
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    auto_create_user BOOLEAN NOT NULL DEFAULT TRUE,        -- 是否自动创建新用户
    default_role    TEXT NOT NULL DEFAULT 'member',         -- 自动创建用户的默认角色
    attribute_mapping JSONB NOT NULL DEFAULT '{}'::jsonb,  -- 属性映射 (email/display_name/avatar)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sso_providers_ws ON sso_providers(workspace_id, enabled);

-- RLS
ALTER TABLE sso_providers ENABLE ROW LEVEL SECURITY;
ALTER TABLE sso_providers FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON sso_providers
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint);

CREATE TRIGGER trg_sso_providers_updated_at BEFORE UPDATE ON sso_providers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -----------------------------------------------------------------
-- sso_links: 用户与 SSO Provider 的关联
-- 支持一个用户通过多个 IdP 登录（如企业微信 + 钉钉）
-- -----------------------------------------------------------------
CREATE TABLE sso_links (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_id     BIGINT NOT NULL REFERENCES sso_providers(id) ON DELETE CASCADE,
    sso_subject     TEXT NOT NULL,                          -- IdP 中的唯一标识
    sso_email       TEXT,                                   -- IdP 返回的邮箱
    sso_display_name TEXT,                                  -- IdP 返回的显示名
    id_token        TEXT,                                   -- 最近一次 id_token (加密存储)
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (provider_id, sso_subject)
);

CREATE INDEX idx_sso_links_user ON sso_links(user_id);
CREATE INDEX idx_sso_links_provider ON sso_links(provider_id, sso_subject);

-- RLS
ALTER TABLE sso_links ENABLE ROW LEVEL SECURITY;
ALTER TABLE sso_links FORCE ROW LEVEL SECURITY;
-- sso_links 不直接暴露给前端 API，由服务端使用

CREATE TRIGGER trg_sso_links_updated_at BEFORE UPDATE ON sso_links
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -----------------------------------------------------------------
-- sso_sessions: SSO 登录会话（防 CSRF + state 验证）
-- 登录流程: GET /auth/oidc/login → 生成 state → 跳转 IdP → 回调验证 state
-- -----------------------------------------------------------------
CREATE TABLE sso_sessions (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    state           TEXT UNIQUE NOT NULL,                   -- CSRF state token
    nonce           TEXT,                                   -- OIDC nonce (防重放)
    provider_id     BIGINT NOT NULL REFERENCES sso_providers(id) ON DELETE CASCADE,
    redirect_to     TEXT,                                   -- 登录成功后的跳转 URL
    ip_address      TEXT,                                   -- 发起登录的客户端 IP
    user_agent      TEXT,                                   -- 发起登录的 User-Agent
    status          TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','completed','expired','failed')),
    user_id         BIGINT REFERENCES users(id),            -- 完成后的用户 ID
    error_message   TEXT,                                   -- 失败原因
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '10 minutes')
);

CREATE INDEX idx_sso_sessions_state ON sso_sessions(state);
CREATE INDEX idx_sso_sessions_expiry ON sso_sessions(expires_at) WHERE status = 'pending';

-- 自动清理过期会话
CREATE OR REPLACE FUNCTION fn_cleanup_sso_sessions()
RETURNS void AS $$
BEGIN
    DELETE FROM sso_sessions WHERE expires_at < now() AND status = 'pending';
END;
$$ LANGUAGE plpgsql;
