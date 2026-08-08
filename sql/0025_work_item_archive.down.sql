-- 大表归档功能回滚
DROP INDEX IF EXISTS idx_defect_archived;
ALTER TABLE defect DROP COLUMN IF EXISTS archived_at;

DROP INDEX IF EXISTS idx_requirement_archived;
ALTER TABLE requirement DROP COLUMN IF EXISTS archived_at;

DROP INDEX IF EXISTS idx_task_archived;
ALTER TABLE task DROP COLUMN IF EXISTS archived_at;
