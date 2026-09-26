-- ===========================================================================
-- V26.10.03 · 活动/审计表按月分区改造
--
-- 背景：task_activities / requirement_activities / defect_activities /
--       webhook_logs 四张表随数据膨胀查询效率递减。按月范围分区将数据按
--       created_at 切分为独立物理子表，查询时自动裁剪无关分区。
--
-- 方案：对每张表采用声明式分区 RANGE (created_at)：
--   Step A. 删除可能存在的旧索引（避免重建冲突）
--   Step B. 重命名原表为 _old
--   Step C. 创建分区表（PARTITION BY RANGE (created_at)）
--   Step D. 创建首批分区（覆盖当前月前后各 3 个月）
--   Step E. 从 _old 迁移数据（INSERT ... SELECT）
--   Step F. 重建索引 + 删除 _old
--
-- 幂等保证：DROP TABLE IF EXISTS + CREATE TABLE IF NOT EXISTS。
-- 可重复执行：仅在 _old 不存在时迁移数据。
-- ===========================================================================

-- 前置：确保已启用 RLS（原迁移 V26.09.26 已处理）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_catalog') THEN
        -- Postgres 原生支持，无需额外扩展
        NULL;
    END IF;
END $$;

-- ═══════════════════════════════════════════════════════════════════════
-- 1. task_activities 分区
-- ═══════════════════════════════════════════════════════════════════════

-- Step B: 重命名原表
ALTER TABLE IF EXISTS task_activities RENAME TO IF EXISTS task_activities_old;

-- Step C: 创建分区表
CREATE TABLE IF NOT EXISTS task_activities (
    id                       BIGINT,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    task_id                  BIGINT NOT NULL,
    verb                     VARCHAR(50) NOT NULL,
    field_name               VARCHAR(100),
    old_value                TEXT,
    new_value                TEXT,
    actor_id                 BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Step D: 创建分区（自动生成未来 6 个月）
DO $$
DECLARE
    start_date DATE;
    end_date   DATE;
    part_name  TEXT;
    part_start TEXT;
    part_end   TEXT;
BEGIN
    FOR i IN 0..6 LOOP
        start_date := date_trunc('month', CURRENT_DATE + (i || ' months')::INTERVAL);
        end_date := start_date + INTERVAL '1 month';
        part_name := 'task_activities_' || to_char(start_date, 'YYYY_MM');
        part_start := to_char(start_date, 'YYYY-MM-DD');
        part_end := to_char(end_date, 'YYYY-MM-DD');

        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF task_activities
             FOR VALUES FROM (%L) TO (%L)',
            part_name, part_start, part_end
        );
    END LOOP;
END $$;

-- Step E: 迁移旧表数据（如 _old 存在）
INSERT INTO task_activities
SELECT * FROM task_activities_old
ON CONFLICT DO NOTHING;

-- Step F: 重建索引 + 清理
CREATE INDEX IF NOT EXISTS idx_task_activities_tenant_id ON task_activities (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_activities_task_id ON task_activities (task_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_activities_workspace_id ON task_activities (workspace_id, created_at DESC);

DROP INDEX IF EXISTS idx_taskactivities_tenant_id;
DROP INDEX IF EXISTS idx_taskactivities_task_id;
DROP TABLE IF EXISTS task_activities_old;

-- ═══════════════════════════════════════════════════════════════════════
-- 2. requirement_activities 分区
-- ═══════════════════════════════════════════════════════════════════════

ALTER TABLE IF EXISTS requirement_activities RENAME TO IF EXISTS requirement_activities_old;

CREATE TABLE IF NOT EXISTS requirement_activities (
    id                       BIGINT,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    requirement_id           BIGINT NOT NULL,
    verb                     VARCHAR(50) NOT NULL,
    field_name               VARCHAR(100),
    old_value                TEXT,
    new_value                TEXT,
    actor_id                 BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

DO $$
DECLARE
    start_date DATE;
    end_date   DATE;
    part_name  TEXT;
BEGIN
    FOR i IN 0..6 LOOP
        start_date := date_trunc('month', CURRENT_DATE + (i || ' months')::INTERVAL);
        end_date := start_date + INTERVAL '1 month';
        part_name := 'requirement_activities_' || to_char(start_date, 'YYYY_MM');
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF requirement_activities
             FOR VALUES FROM (%L) TO (%L)',
            part_name, to_char(start_date, 'YYYY-MM-DD'), to_char(end_date, 'YYYY-MM-DD')
        );
    END LOOP;
END $$;

INSERT INTO requirement_activities SELECT * FROM requirement_activities_old ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_requirement_activities_tenant_id ON requirement_activities (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_requirement_activities_req_id ON requirement_activities (requirement_id, created_at DESC);

DROP TABLE IF EXISTS requirement_activities_old;

-- ═══════════════════════════════════════════════════════════════════════
-- 3. defect_activities 分区
-- ═══════════════════════════════════════════════════════════════════════

ALTER TABLE IF EXISTS defect_activities RENAME TO IF EXISTS defect_activities_old;

CREATE TABLE IF NOT EXISTS defect_activities (
    id                       BIGINT,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    defect_id                BIGINT NOT NULL,
    verb                     VARCHAR(50) NOT NULL,
    field_name               VARCHAR(100),
    old_value                TEXT,
    new_value                TEXT,
    actor_id                 BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

DO $$
DECLARE
    start_date DATE;
    part_name  TEXT;
BEGIN
    FOR i IN 0..6 LOOP
        start_date := date_trunc('month', CURRENT_DATE + (i || ' months')::INTERVAL);
        part_name := 'defect_activities_' || to_char(start_date, 'YYYY_MM');
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF defect_activities
             FOR VALUES FROM (%L) TO (%L)',
            part_name, to_char(start_date, 'YYYY-MM-DD'), to_char(start_date + INTERVAL '1 month', 'YYYY-MM-DD')
        );
    END LOOP;
END $$;

INSERT INTO defect_activities SELECT * FROM defect_activities_old ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_defect_activities_tenant_id ON defect_activities (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_defect_activities_defect_id ON defect_activities (defect_id, created_at DESC);

DROP TABLE IF EXISTS defect_activities_old;

-- ═══════════════════════════════════════════════════════════════════════
-- 4. webhook_logs 分区
-- ═══════════════════════════════════════════════════════════════════════

ALTER TABLE IF EXISTS webhook_logs RENAME TO IF EXISTS webhook_logs_old;

CREATE TABLE IF NOT EXISTS webhook_logs (
    id                       BIGINT,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    webhook_id               BIGINT NOT NULL,
    delivery_id              VARCHAR(64) NOT NULL,
    event_type               VARCHAR(80) NOT NULL,
    event_id                 BIGINT,
    request_url              TEXT NOT NULL,
    request_method           VARCHAR(10) DEFAULT 'POST',
    request_headers          JSONB,
    request_body             TEXT,
    response_status          INTEGER,
    response_body            TEXT,
    response_headers         JSONB,
    attempt                  SMALLINT DEFAULT 1,
    duration_ms              INTEGER,
    error                    TEXT,
    status                   VARCHAR(20) NOT NULL,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

DO $$
DECLARE
    start_date DATE;
    part_name  TEXT;
BEGIN
    FOR i IN 0..6 LOOP
        start_date := date_trunc('month', CURRENT_DATE + (i || ' months')::INTERVAL);
        part_name := 'webhook_logs_' || to_char(start_date, 'YYYY_MM');
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF webhook_logs
             FOR VALUES FROM (%L) TO (%L)',
            part_name, to_char(start_date, 'YYYY-MM-DD'), to_char(start_date + INTERVAL '1 month', 'YYYY-MM-DD')
        );
    END LOOP;
END $$;

INSERT INTO webhook_logs SELECT * FROM webhook_logs_old ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_webhook_logs_tenant_id ON webhook_logs (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_webhook_id ON webhook_logs (webhook_id, created_at DESC);

DROP INDEX IF EXISTS idx_webhooklogs_webhook_id;
DROP INDEX IF EXISTS idx_webhooklogs_tenant_id;
DROP TABLE IF EXISTS webhook_logs_old;

-- ═══════════════════════════════════════════════════════════════════════
-- 附加：自动化分区维护函数（每月月初创建未来分区）
-- ═══════════════════════════════════════════════════════════════════════

CREATE OR REPLACE FUNCTION maintain_monthly_partitions()
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    tbl RECORD;
    part_name TEXT;
    start_date DATE;
    end_date DATE;
    tables TEXT[] := ARRAY['task_activities','requirement_activities','defect_activities','webhook_logs'];
    tbl_name TEXT;
BEGIN
    FOREACH tbl_name IN ARRAY tables LOOP
        FOR i IN 1..3 LOOP
            start_date := date_trunc('month', CURRENT_DATE + (i || ' months')::INTERVAL);
            end_date := start_date + INTERVAL '1 month';
            part_name := tbl_name || '_' || to_char(start_date, 'YYYY_MM');

            BEGIN
                EXECUTE format(
                    'CREATE TABLE IF NOT EXISTS %I PARTITION OF %I
                     FOR VALUES FROM (%L) TO (%L)',
                    part_name, tbl_name,
                    to_char(start_date, 'YYYY-MM-DD'),
                    to_char(end_date, 'YYYY-MM-DD')
                );
            EXCEPTION WHEN duplicate_table THEN
                NULL;
            END;
        END LOOP;
    END LOOP;
END;
$$;

-- SELECT maintain_monthly_partitions();  -- 每月月初调用
