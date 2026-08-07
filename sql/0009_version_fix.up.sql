-- 0009_version_fix: 版本聚合根修正
-- 变更:
--   1. versions 表加 start_date / end_date 属性（版本起止时间）
--   2. sprints 表加 version_id FK（迭代 → 版本的 N:1 关联，替换原 version_sprints M2M）
--   3. 迁移 version_sprints 已有数据到 sprints.version_id
--   4. 删除 version_sprints 关联表
-- 业务规则: 一个版本聚合多个迭代，一个迭代只属于一个版本

-- -----------------------------------------------------------------
-- Step 1: versions 表增加 start_date / end_date
-- -----------------------------------------------------------------
ALTER TABLE versions ADD COLUMN IF NOT EXISTS start_date DATE;
ALTER TABLE versions ADD COLUMN IF NOT EXISTS end_date   DATE;

-- 校验: start_date <= end_date
ALTER TABLE versions ADD CONSTRAINT versions_date_range CHECK (
    start_date IS NULL OR end_date IS NULL OR start_date <= end_date
);

-- -----------------------------------------------------------------
-- Step 2: sprints 表增加 version_id FK（N:1 关联）
-- -----------------------------------------------------------------
ALTER TABLE sprints ADD COLUMN IF NOT EXISTS version_id BIGINT REFERENCES versions(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_sprints_version ON sprints(version_id) WHERE deleted_at IS NULL;

-- RLS 策略不需要调整（sprints 已有 workspace_id 隔离）

-- -----------------------------------------------------------------
-- Step 3: 数据迁移 — 将 version_sprints 中每个迭代关联的"最早版本"写入 sprints.version_id
-- 保证约束: 一个迭代只属于一个版本
-- -----------------------------------------------------------------
DO $$
DECLARE
    sprint_record RECORD;
    first_version_id BIGINT;
BEGIN
    -- 遍历所有在 version_sprints 中有记录的迭代
    FOR sprint_record IN
        SELECT DISTINCT sprint_id FROM version_sprints
    LOOP
        -- 取该 sprint 关联的最早版本作为归属版本
        SELECT version_id INTO first_version_id
        FROM version_sprints
        WHERE sprint_id = sprint_record.sprint_id
        ORDER BY added_at ASC
        LIMIT 1;

        -- 写入 sprints.version_id
        UPDATE sprints SET version_id = first_version_id
        WHERE id = sprint_record.sprint_id;
    END LOOP;
END $$;

-- -----------------------------------------------------------------
-- Step 4: 删除 version_sprints 关联表（已无用途）
-- -----------------------------------------------------------------
-- 先清理依赖对象
DROP POLICY IF EXISTS tenant_isolation ON version_sprints;
DROP INDEX IF EXISTS idx_version_sprints_sprint;
DROP TABLE IF EXISTS version_sprints;
