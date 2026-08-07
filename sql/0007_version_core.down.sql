-- 0007_version_core 回滚
DROP TRIGGER IF EXISTS trg_versions_updated_at ON versions;
DROP TABLE IF EXISTS version_sprints;
DROP TABLE IF EXISTS versions;
ALTER TABLE issues DROP COLUMN IF EXISTS found_version_id;
ALTER TABLE issues DROP COLUMN IF EXISTS fix_version_id;
ALTER TABLE issues DROP COLUMN IF EXISTS release_version_id;
