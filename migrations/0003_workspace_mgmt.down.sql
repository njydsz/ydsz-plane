-- 0003_workspace_mgmt down: drop tables + triggers + policies

DROP TRIGGER IF EXISTS trg_projects_updated_at ON projects;
DROP TRIGGER IF EXISTS trg_api_tokens_updated_at ON api_tokens;
DROP TRIGGER IF EXISTS trg_invitations_updated_at ON invitations;

DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS api_tokens;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS invitations;
