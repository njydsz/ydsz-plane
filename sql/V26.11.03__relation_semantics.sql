-- V26.11.03__relation_semantics.sql — 工作项关系语义增强
--
-- 对标 Plane 的关系类型：relates / duplicate / blocked / blocking
-- 并补充 dependency（依赖）与 parent/child（父子）。
--
-- 说明：task_relations / requirement_relations / defect_relations 三表结构一致，
-- 本迁移统一为 relation_type 增加 CHECK 约束（幂等，先删后建）。

DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY['task_relations', 'requirement_relations', 'defect_relations'];
BEGIN
    FOREACH tbl IN ARRAY tables LOOP
        -- 删除旧约束（若存在）
        EXECUTE format(
            'ALTER TABLE %I DROP CONSTRAINT IF EXISTS chk_%s_relation_type',
            tbl, tbl
        );
        -- 建立新约束
        EXECUTE format(
            'ALTER TABLE %I ADD CONSTRAINT chk_%s_relation_type CHECK (relation_type IN (
                ''relates'', ''duplicate'', ''blocked'', ''blocking'',
                ''depends_on'', ''parent'', ''child'', ''relates_to''
            ))',
            tbl, tbl
        );
    END LOOP;
END $$;

-- 关系反向查询索引（blocked/blocking 需要双向解析）
CREATE INDEX IF NOT EXISTS idx_task_relations_target
    ON task_relations (target_task_id, relation_type) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_relations_target
    ON requirement_relations (target_task_id, relation_type) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_relations_target
    ON defect_relations (target_task_id, relation_type) WHERE NOT deleted;

COMMENT ON COLUMN task_relations.relation_type IS
    '关系类型：relates/duplicate/blocked/blocking/depends_on/parent/child';
