-- 0026_sso_oidc.down: 回滚 SSO/OIDC 集成
DROP FUNCTION IF EXISTS fn_cleanup_sso_sessions();
DROP TABLE IF EXISTS sso_sessions;
DROP TABLE IF EXISTS sso_links;
DROP TABLE IF EXISTS sso_providers;
DROP INDEX IF EXISTS idx_users_sso_subject;
ALTER TABLE users DROP COLUMN IF EXISTS sso_subject;
ALTER TABLE users DROP COLUMN IF EXISTS sso_provider;
-- 注意: password_hash NOT NULL 不回滚（SSO 集成后已有无密码用户）
