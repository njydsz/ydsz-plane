-- ===========================================================================
-- V26.09.27 · public_id UUID 安全引用 ID 补齐
-- 依据：docs/architecture/04-数据模型设计.md §1 主键约定
--       S18 分析报告 P1-D3
-- ===========================================================================
-- 设计原则：
--   1. 外部 API 展示的 ID 使用 public_id UUID（不可遍历），
--      内部关联与查询仍用 id BIGINT 雪花主键（性能最优）
--   2. 已含 public_id 且带 UNIQUE 约束的表 42 张，本次补齐剩余 ~15 张核心表
--   3. 历史库已有数据自动填充 UUID，新库 gen_random_uuid() 默认即生成
-- ===========================================================================

-- 幂等：仅当列不存在时添加（安全重入）
DO $$
DECLARE
    t TEXT;
    tables TEXT[] := ARRAY[
        'roles',
        'project_sequences',
        'states', 'state_transitions',
        'labels', 'modules', 'estimate_points',
        'sprints',
        'invitations',
        'workbench_configs',
        'audit_logs'
    ];
BEGIN
    FOREACH t IN ARRAY tables LOOP
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = t AND table_schema = 'public')
           AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = t AND column_name = 'public_id' AND table_schema = 'public') THEN
            EXECUTE format('ALTER TABLE %I ADD COLUMN public_id UUID DEFAULT gen_random_uuid() NOT NULL', t);
            EXECUTE format('ALTER TABLE %I ADD CONSTRAINT %I UNIQUE (public_id)', t, t || '_public_id_key');

            -- 为 public_id 建索引（按 UUID 查询场景：外部 API lookup）
            EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_public_id ON %I (public_id)', t, t);
        END IF;
    END LOOP;
END;
$$;
