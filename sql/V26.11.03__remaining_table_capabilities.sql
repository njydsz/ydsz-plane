-- ===========================================================================
-- V26.11.03 · 剩余数据库表能力补齐 + 性能优化
-- 依据：docs/optimization/S20-数据库表·接口测试·前后端贯通现状分析与优化建议.md §1.3.2 D4-D9
-- ===========================================================================
-- 本迁移合并落地 S20 报告中 P2 级的数据库优化项：
--   D4: version_sprint_relations 序列 + 索引（版本-迭代多对多）
--   D5: metric_adjustments 序列 + 索引（度量数据人工校准）
--   D6: project_configs 序列 + 配置字段索引
--   D8: JSONB 高频字段提取（projects.modules / defect.environment）
--   D9: PG 扩展启用（pg_trgm/pg_stat_statements，幂等）
--
-- 全部语句可重复执行（幂等）。
-- ===========================================================================

-- ═══════════════════════════════════════════════════════════════════════
-- D4: version_sprint_relations 运行时基座
-- ═══════════════════════════════════════════════════════════════════════

CREATE SEQUENCE IF NOT EXISTS version_sprint_relations_id_seq START 1 INCREMENT 1;
ALTER TABLE version_sprint_relations
    ALTER COLUMN id SET DEFAULT nextval('version_sprint_relations_id_seq');
ALTER SEQUENCE version_sprint_relations_id_seq OWNED BY version_sprint_relations.id;

-- 按版本查关联迭代
CREATE INDEX IF NOT EXISTS idx_vsr_version_id
    ON version_sprint_relations (version_id) WHERE deleted = false;
-- 按迭代查关联版本
CREATE INDEX IF NOT EXISTS idx_vsr_sprint_id
    ON version_sprint_relations (sprint_id) WHERE deleted = false;

-- ═══════════════════════════════════════════════════════════════════════
-- D5: metric_adjustments 运行时基座
-- ═══════════════════════════════════════════════════════════════════════

CREATE SEQUENCE IF NOT EXISTS metric_adjustments_id_seq START 1 INCREMENT 1;
ALTER TABLE metric_adjustments
    ALTER COLUMN id SET DEFAULT nextval('metric_adjustments_id_seq');
ALTER SEQUENCE metric_adjustments_id_seq OWNED BY metric_adjustments.id;

CREATE INDEX IF NOT EXISTS idx_metric_adjustments_project_metric
    ON metric_adjustments (project_id, metric) WHERE deleted = false;
CREATE INDEX IF NOT EXISTS idx_metric_adjustments_snapshot
    ON metric_adjustments (snapshot_id) WHERE deleted = false;

-- ═══════════════════════════════════════════════════════════════════════
-- D6: project_configs 运行时基座
-- ═══════════════════════════════════════════════════════════════════════

CREATE SEQUENCE IF NOT EXISTS project_configs_id_seq START 1 INCREMENT 1;
ALTER TABLE project_configs
    ALTER COLUMN id SET DEFAULT nextval('project_configs_id_seq');
ALTER SEQUENCE project_configs_id_seq OWNED BY project_configs.id;

-- 项目配置查询按 workspace + project 定位
CREATE INDEX IF NOT EXISTS idx_project_configs_ws_project
    ON project_configs (workspace_id, project_id) WHERE deleted = false;

-- ═══════════════════════════════════════════════════════════════════════
-- D8: JSONB 高频字段提取为列 + GIN 索引
--
-- projects.modules 是项目模块开关（JSONB），前端「项目设置」页高频读取；
-- 提取为布尔列避免每次全量解析 JSONB。
-- 兼容性：保留原 JSONB 列，新增列作为缓存；应用层双写过渡。
-- ═══════════════════════════════════════════════════════════════════════

-- 仅在列不存在时添加（幂等）。
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'projects' AND column_name = 'modules_enabled'
    ) THEN
        -- modules_enabled：项目是否启用模块功能（原 JSONB modules.enabled）
        ALTER TABLE projects ADD COLUMN modules_enabled BOOLEAN DEFAULT true;
        -- modules_config：保留原 JSONB 作为完整配置的备份
        -- （不删除，expand-contract 模式，待应用层迁移完成后再评估清理）
    END IF;
END
$$;

-- GIN 索引：加速 JSONB 内部键值查询（如 modules->>'enabled'）。
CREATE INDEX IF NOT EXISTS idx_projects_modules_gin
    ON projects USING GIN (modules) WHERE deleted = false;

-- defect.environment 同理：缺陷环境信息 JSONB 高频按环境过滤。
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'defect' AND column_name = 'environment_text'
    ) THEN
        -- environment_text：冗余存储 environment->>'value'，避免 JSONB 提取
        ALTER TABLE defect ADD COLUMN environment_text VARCHAR(50);
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_defect_environment_text
    ON defect (environment_text) WHERE deleted = false AND environment_text IS NOT NULL;

-- ═══════════════════════════════════════════════════════════════════════
-- D9: PostgreSQL 扩展启用（幂等）
--
-- pg_trgm：支撑模糊搜索（LIKE %xxx%），补充 tsvector 未覆盖的场景。
-- pg_stat_statements：SQL 执行统计，用于慢查询巡检与 P95 基线对比。
-- 注：两扩展已在 V26.09.26 迁移中启用过，此处重复声明保证独立可执行。
-- ═══════════════════════════════════════════════════════════════════════

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- 为模糊搜索场景预建 trigram 索引（若列存在）。
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'task' AND column_name = 'name'
    ) THEN
        CREATE INDEX IF NOT EXISTS idx_task_name_trgm
            ON task USING GIN (name gin_trgm_ops) WHERE deleted = false;
    END IF;
END
$$;
