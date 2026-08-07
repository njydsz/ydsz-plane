DROP TRIGGER IF EXISTS trg_workspaces_updated_at ON workspaces;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at();
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS domain_events;
DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS users;
