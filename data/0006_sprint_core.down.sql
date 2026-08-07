-- 0006_sprint_core down
DROP INDEX IF EXISTS idx_one_active_sprint_per_project;
DROP TRIGGER IF EXISTS trg_sprints_updated_at ON sprints;
DROP TABLE IF EXISTS sprint_snapshots;
DROP TABLE IF EXISTS sprint_issues;
DROP TABLE IF EXISTS sprints;
