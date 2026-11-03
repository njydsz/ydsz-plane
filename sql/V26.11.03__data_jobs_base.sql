-- ===========================================================================
-- V26.11.03 · 数据任务（异步作业）基座
-- 依据：docs/optimization/S20-数据库表·接口测试·前后端贯通现状分析与优化建议.md §1.3.2 D3
-- ===========================================================================
-- 背景：
--   data_jobs 表用于跟踪异步数据任务（导入/导出/同步/报表生成），
--   此前无任何前后端代码引用。本迁移补齐运行时基座：
--     1. id 列序列默认值
--     2. job_type / status 索引
--     3. 状态枚举扩展：原表复用 work_item_status（draft/active/completed/cancelled），
--        但异步任务语义需要 pending/running/failed，故新增专用枚举类型。
--
-- 兼容性策略：
--   - 新增 data_job_status 枚举（pending/running/success/failed/cancelled）；
--   - 将 data_jobs.status 列类型切换为新枚举，旧值映射：
--       draft → pending、active → running、completed → success、cancelled → cancelled；
--   - 使用 USING 子句在 ALTER 时完成值转换，保证存量数据安全。
-- ===========================================================================

-- 1. 创建专用状态枚举（幂等）。
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'data_job_status') THEN
        CREATE TYPE data_job_status AS ENUM ('pending', 'running', 'success', 'failed', 'cancelled');
    END IF;
END
$$;

-- 2. id 序列。
CREATE SEQUENCE IF NOT EXISTS data_jobs_id_seq START 1 INCREMENT 1;

ALTER TABLE data_jobs
    ALTER COLUMN id SET DEFAULT nextval('data_jobs_id_seq');

ALTER SEQUENCE data_jobs_id_seq OWNED BY data_jobs.id;

-- 3. 状态列类型切换（仅在尚未是新枚举时执行，避免重复迁移报错）。
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'data_jobs'
          AND column_name = 'status'
          AND udt_name = 'work_item_status'
    ) THEN
        ALTER TABLE data_jobs
            ALTER COLUMN status DROP DEFAULT,
            ALTER COLUMN status TYPE data_job_status
                USING CASE status::text
                    WHEN 'draft' THEN 'pending'
                    WHEN 'active' THEN 'running'
                    WHEN 'completed' THEN 'success'
                    WHEN 'cancelled' THEN 'cancelled'
                    ELSE 'pending'
                END::data_job_status,
            ALTER COLUMN status SET DEFAULT 'pending';
    END IF;
END
$$;

-- 4. 索引：job_type + status 组合加速「按类型查运行中任务」。
CREATE INDEX IF NOT EXISTS idx_data_jobs_type_status
    ON data_jobs (job_type, status)
    WHERE deleted = false;

-- 5. 索引：scheduled_at 加速调度扫描（ worker 轮询待执行任务）。
CREATE INDEX IF NOT EXISTS idx_data_jobs_scheduled_at
    ON data_jobs (scheduled_at)
    WHERE deleted = false AND status = 'pending';

-- 6. 部分唯一索引：同一工作空间内同名任务不重复创建（软删除行不参与约束）。
CREATE UNIQUE INDEX IF NOT EXISTS uk_data_jobs_workspace_name
    ON data_jobs (workspace_id, name)
    WHERE deleted = false;
