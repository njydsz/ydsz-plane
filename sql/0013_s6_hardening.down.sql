-- 0013_s6_hardening.down: 回滚 S6 加固变更

DROP POLICY IF EXISTS tenant_isolation ON version_delivery_snapshots;
DROP TABLE IF EXISTS version_delivery_snapshots;
DROP INDEX IF EXISTS idx_audit_logs_action_target;
DROP TRIGGER IF EXISTS trg_versions_bump_version ON versions;
DROP FUNCTION IF EXISTS bump_version();
ALTER TABLE versions DROP CONSTRAINT IF EXISTS versions_checklist_limit;
ALTER TABLE versions DROP COLUMN IF EXISTS version;
