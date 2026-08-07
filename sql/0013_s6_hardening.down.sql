-- 0013_s6_hardening.down: 回滚 S6 加固变更

DROP TRIGGER IF EXISTS trg_versions_bump_version ON versions;
DROP FUNCTION IF EXISTS bump_version();
ALTER TABLE versions DROP CONSTRAINT IF EXISTS versions_checklist_limit;
ALTER TABLE versions DROP COLUMN IF EXISTS version;
DROP INDEX IF EXISTS idx_audit_logs_action_target;
