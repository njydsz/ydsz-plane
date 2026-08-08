-- S13 SSO 迁移脚本（安全幂等：可重复执行）
-- 适用场景：已有 ydsz-plane 数据库，升级为支持 OIDC/SAML SSO
-- 执行方式: psql -h <host> -U <user> -d <db> -f scripts/sso_migrate.sql

-- 1) users 表补充 SSO 关联字段
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS sso_provider TEXT;
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS sso_subject TEXT;
CREATE INDEX IF NOT EXISTS idx_users_sso ON public.users (sso_provider, sso_subject) WHERE sso_provider IS NOT NULL;

-- 2) sso_providers：工作空间级 SSO/OIDC Provider 配置
CREATE TABLE IF NOT EXISTS public.sso_providers (
  id int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
    INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 START 1 CACHE 1
  ),
  workspace_id int8 NOT NULL,
  name text NOT NULL,
  protocol text NOT NULL DEFAULT 'oidc',
  issuer_url text,
  client_id text NOT NULL,
  client_secret text NOT NULL,
  redirect_uri text NOT NULL,
  auth_url text,
  token_url text,
  userinfo_url text,
  jwks_url text,
  scopes text NOT NULL DEFAULT 'openid email profile',
  auto_create_user bool NOT NULL DEFAULT true,
  default_role text NOT NULL DEFAULT 'member',
  attribute_mapping jsonb NOT NULL DEFAULT '{}'::jsonb,
  enabled bool NOT NULL DEFAULT true,
  created_at timestamptz(6) NOT NULL DEFAULT now(),
  updated_at timestamptz(6) NOT NULL DEFAULT now(),
  CONSTRAINT sso_providers_pkey PRIMARY KEY (id),
  CONSTRAINT sso_providers_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_sso_providers_workspace ON public.sso_providers (workspace_id, enabled);

-- 3) sso_sessions：OIDC 登录会话（含 PKCE code_verifier）
CREATE TABLE IF NOT EXISTS public.sso_sessions (
  id int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
    INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 START 1 CACHE 1
  ),
  state text NOT NULL,
  nonce text NOT NULL,
  code_verifier text,
  provider_id int8 NOT NULL,
  redirect_to text,
  ip_address inet,
  user_agent text,
  user_id int8,
  status text NOT NULL DEFAULT 'pending',
  error_message text,
  expires_at timestamptz(6) NOT NULL DEFAULT now() + interval '10 minutes',
  completed_at timestamptz(6),
  created_at timestamptz(6) NOT NULL DEFAULT now(),
  CONSTRAINT sso_sessions_pkey PRIMARY KEY (id),
  CONSTRAINT sso_sessions_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.sso_providers (id) ON DELETE CASCADE,
  CONSTRAINT sso_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users (id) ON DELETE SET NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sso_sessions_state ON public.sso_sessions (state);
CREATE INDEX IF NOT EXISTS idx_sso_sessions_status_expires ON public.sso_sessions (status, expires_at);

-- 4) sso_links：用户-SSO 身份绑定
CREATE TABLE IF NOT EXISTS public.sso_links (
  id int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
    INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 START 1 CACHE 1
  ),
  user_id int8 NOT NULL,
  provider_id int8 NOT NULL,
  sso_subject text NOT NULL,
  sso_email text,
  sso_display_name text,
  last_login_at timestamptz(6),
  created_at timestamptz(6) NOT NULL DEFAULT now(),
  updated_at timestamptz(6) NOT NULL DEFAULT now(),
  CONSTRAINT sso_links_pkey PRIMARY KEY (id),
  CONSTRAINT sso_links_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users (id) ON DELETE CASCADE,
  CONSTRAINT sso_links_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.sso_providers (id) ON DELETE CASCADE,
  CONSTRAINT sso_links_provider_subject_unique UNIQUE (provider_id, sso_subject)
);
CREATE INDEX IF NOT EXISTS idx_sso_links_user ON public.sso_links (user_id);

-- Comments
COMMENT ON TABLE public.sso_providers IS 'S13: Workspace-level SSO/OIDC Provider configuration';
COMMENT ON COLUMN public.sso_providers.protocol IS 'Protocol: oidc | saml';
COMMENT ON TABLE public.sso_sessions IS 'S13: OIDC login session state with PKCE code_verifier';
COMMENT ON COLUMN public.sso_sessions.status IS 'Session status: pending | completed | failed | expired';
COMMENT ON TABLE public.sso_links IS 'S13: User-SSO identity binding (one provider+subject maps to one user)';

\echo 'S13 SSO migration complete'
