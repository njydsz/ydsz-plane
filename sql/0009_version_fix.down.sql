-- 0009_version_fix 回滚: 恢复 version_sprints M2M 关联表，移除 sprints.version_id

-- -----------------------------------------------------------------
-- Step 1: 重新创建 version_sprints 关联表
-- -----------------------------------------------------------------
CREATE TABLE version_sprints (
    version_id      BIGINT NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    sprint_id       BIGINT NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    added_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    added_by        BIGINT REFERENCES users(id),
    PRIMARY KEY (version_id, sprint_id)
);
CREATE INDEX idx_version_sprints_sprint ON version_sprints(sprint_id);

ALTER TABLE version_sprints ENABLE ROW LEVEL SECURITY;
ALTER TABLE version_sprints FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON version_sprints
    USING (version_id IN (SELECT id FROM versions WHERE workspace_id = current_setting('app.workspace_id', true)::bigint));

-- -----------------------------------------------------------------
-- Step 2: 恢复数据 — 将 sprints.version_id 写入 version_sprints
-- -----------------------------------------------------------------
INSERT INTO version_sprints (version_id, sprint_id)
SELECT version_id, id FROM sprints WHERE version_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------
-- Step 3: 移除 sprints.version_id FK + 索引
-- -----------------------------------------------------------------
DROP INDEX IF EXISTS idx_sprints_version;
ALTER TABLE sprints DROP CONSTRAINT IF EXISTS sprints_version_id_fkey;
ALTER TABLE sprints DROP COLUMN IF EXISTS version_id;

-- -----------------------------------------------------------------
-- Step 4: 移除 versions 表的 start_date / end_date + 约束
-- -----------------------------------------------------------------
ALTER TABLE versions DROP CONSTRAINT IF EXISTS versions_date_range;
ALTER TABLE versions DROP COLUMN IF EXISTS start_date;
ALTER TABLE versions DROP COLUMN IF EXISTS end_date;
