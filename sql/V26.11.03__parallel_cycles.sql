-- V26.11.03__parallel_cycles.sql — 并行周期（Parallel Cycles）
--
-- 对标 Plane Enterprise Grid 的并行周期能力：
-- 允许同一项目同时存在多个 active 迭代。
--
-- 变更：
--   1. 删除 uq_sprints_active 唯一索引（该索引强制每项目仅一个 active）；
--   2. sprints 新增 allow_parallel 标记（项目级配置，默认 false 保持向后兼容）；
--   3. 新增活跃迭代数索引以支持 dashboard 查询。

-- 1. 删除单活跃约束（如存在）
DROP INDEX IF EXISTS uq_sprints_active;
DROP INDEX IF EXISTS idx_one_active_sprint_per_project;

-- 2. 项目级并行开关
ALTER TABLE projects ADD COLUMN IF NOT EXISTS allow_parallel_cycles BOOLEAN NOT NULL DEFAULT FALSE;

-- 3. 活跃迭代索引（非唯一）
CREATE INDEX IF NOT EXISTS idx_sprints_active_lookup
    ON sprints (tenant_id, project_id, status)
    WHERE status = 'active' AND NOT deleted;

COMMENT ON COLUMN projects.allow_parallel_cycles IS '是否允许并行周期（同项目多个 active 迭代）';
