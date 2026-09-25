-- ============================================================================
-- S16 P1-3: 搜索能力补齐 — PG FTS + JQL op 扩展
-- 日期: 2026-09-25
-- 对标: Jira JQL 子集、云效搜索、Linear full-text
-- ============================================================================

-- --- 1. 三表 tsvector 列 + GIN 索引（通过 to_tsvector 中文+英文自动分词） ---
-- 注意：中文场景依赖 pg_trgm（已内建或需额外安装 extension）。生产建议用 zhparser。
-- 信创版本（达梦/金仓）中此 migration 可跳过，PG FTS 由 ES 替代。

-- 为 task 表添加搜索向量列
ALTER TABLE task ADD COLUMN IF NOT EXISTS search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(description_stripped, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(category, '')), 'C')
    ) STORED;

-- 为 requirement 表添加搜索向量列
ALTER TABLE requirement ADD COLUMN IF NOT EXISTS search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(description_stripped, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(source, '')), 'C')
    ) STORED;

-- 为 defect 表添加搜索向量列
ALTER TABLE defect ADD COLUMN IF NOT EXISTS search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(description_stripped, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(found_phase, '')), 'C') ||
        setweight(to_tsvector('simple', coalesce(root_cause_category, '')), 'D')
    ) STORED;

-- GIN 索引加速全文搜索
CREATE INDEX IF NOT EXISTS idx_task_search_vector ON task USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_requirement_search_vector ON requirement USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_defect_search_vector ON defect USING GIN (search_vector);

-- trigram 索引（补充中文模糊搜索，无需分词）
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_task_name_trgm ON task USING GIN (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_requirement_name_trgm ON requirement USING GIN (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_defect_name_trgm ON defect USING GIN (name gin_trgm_ops);

-- --- 2. issue_activities 搜索辅助（"was in"/"changed during" JQL 算子） ---
-- 为状态变更历史添加复合索引，支持 "status changed DURING (-7d, now)" 类查询
CREATE INDEX IF NOT EXISTS idx_activities_issue_field_time
    ON issue_activities (issue_id, field, created_at DESC);

-- --- 3. 跨类型搜索期望字段对齐视图 ---
-- 聚合三表，统一字段名，供应用层 UNION ALL 使用
CREATE OR REPLACE VIEW v_issues_search AS
SELECT
    id,
    workspace_id,
    project_id,
    'task'::text AS type_code,
    name,
    state_id,
    priority,
    severity,
    COALESCE(point, 0) AS point,
    assignee_ids,
    search_vector
FROM task
WHERE deleted = false
UNION ALL
SELECT
    id,
    workspace_id,
    project_id,
    'requirement'::text,
    name,
    state_id,
    priority,
    NULL::int,
    COALESCE(point, 0),
    assignee_ids,
    search_vector
FROM requirement
WHERE deleted = false
UNION ALL
SELECT
    id,
    workspace_id,
    project_id,
    'defect'::text,
    name,
    state_id,
    priority,
    severity,
    COALESCE(point, 0),
    assignee_ids,
    search_vector
FROM defect
WHERE deleted = false;

COMMENT ON VIEW v_issues_search IS 'S16: 跨类型搜索对齐视图，统一字段供 FTS + JQL 使用';
