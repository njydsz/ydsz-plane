-- V26.11.03__oauth_app.sql — OAuth 应用与授权
--
-- 对标 Plane Developer Apps / GitHub OAuth Apps：
--   - oauth_apps           ：开发者注册的 OAuth 应用（client_id / client_secret_hash + 回调白名单 + scope）；
--   - oauth_codes          ：一次性授权码（短时效、绑定 client_id 与 redirect_uri）；
--   - oauth_tokens         ：授权码换取的访问/刷新令牌（只存 hash）；
--   - oauth_authorizations ：用户对应用的授权记录（scope 收敛）。
--
-- 应用层生成 id（雪花式 BIGINT）。client_secret / access_token / refresh_token
-- 原始值仅创建/换取时返回一次，数据库只存 SHA-256 hash。

-- ---------------------------------------------------------------------------
-- oauth_apps
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS oauth_apps (
    id                BIGINT       PRIMARY KEY,
    workspace_id      BIGINT       NOT NULL,
    name              VARCHAR(200) NOT NULL,
    description       TEXT,
    client_id         VARCHAR(64)  NOT NULL,
    client_secret_hash TEXT        NOT NULL,
    redirect_uris     TEXT[]       NOT NULL DEFAULT '{}',
    scopes            TEXT[]       NOT NULL DEFAULT '{}',
    icon_url          VARCHAR(500),
    status            VARCHAR(20)  NOT NULL DEFAULT 'active',
    created_by        BIGINT,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_oauth_apps_status CHECK (status IN ('active', 'revoked'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_oauth_apps_client_id ON oauth_apps (client_id);
CREATE INDEX IF NOT EXISTS idx_oauth_apps_workspace ON oauth_apps (workspace_id) WHERE status <> 'revoked';

-- ---------------------------------------------------------------------------
-- oauth_codes（一次性授权码，10 分钟过期）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS oauth_codes (
    id           BIGINT       PRIMARY KEY,
    code         VARCHAR(128) NOT NULL,
    app_id       BIGINT       NOT NULL,
    user_id      BIGINT       NOT NULL,
    workspace_id BIGINT       NOT NULL,
    redirect_uri VARCHAR(500) NOT NULL,
    scopes       TEXT[]       NOT NULL DEFAULT '{}',
    expires_at   TIMESTAMPTZ  NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_oauth_codes_code ON oauth_codes (code);
CREATE INDEX IF NOT EXISTS idx_oauth_codes_app ON oauth_codes (app_id, expires_at);

-- ---------------------------------------------------------------------------
-- oauth_tokens（访问/刷新令牌，只存 hash）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS oauth_tokens (
    id                 BIGINT       PRIMARY KEY,
    app_id             BIGINT       NOT NULL,
    user_id            BIGINT       NOT NULL,
    workspace_id       BIGINT       NOT NULL,
    access_hash        TEXT         NOT NULL,
    refresh_hash       TEXT         NOT NULL,
    scopes             TEXT[]       NOT NULL DEFAULT '{}',
    expires_at         TIMESTAMPTZ  NOT NULL,
    refresh_expires_at TIMESTAMPTZ  NOT NULL,
    revoked_at         TIMESTAMPTZ,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_oauth_tokens_access_hash ON oauth_tokens (access_hash);
CREATE UNIQUE INDEX IF NOT EXISTS uk_oauth_tokens_refresh_hash ON oauth_tokens (refresh_hash);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_app_user ON oauth_tokens (app_id, user_id) WHERE revoked_at IS NULL;

-- ---------------------------------------------------------------------------
-- oauth_authorizations（用户授权记录）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS oauth_authorizations (
    id         BIGINT      PRIMARY KEY,
    app_id     BIGINT      NOT NULL,
    user_id    BIGINT      NOT NULL,
    scopes     TEXT[]      NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_oauth_authorizations_app_user ON oauth_authorizations (app_id, user_id);

COMMENT ON TABLE oauth_apps IS 'OAuth 应用注册（client_id/secret + 回调白名单 + scope）';
COMMENT ON TABLE oauth_codes IS '一次性 OAuth 授权码（10 分钟过期）';
COMMENT ON TABLE oauth_tokens IS 'OAuth 访问/刷新令牌（只存 hash，原始值仅返回一次）';
COMMENT ON TABLE oauth_authorizations IS '用户对 OAuth 应用的授权记录';
