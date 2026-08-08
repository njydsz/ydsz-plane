-- 工作项分表拆分回滚脚本
-- 删除分表相关的所有表、索引、RLS策略，恢复states和state_transitions表原有结构

-- 1. 删除关联关系表
DROP INDEX IF EXISTS idx_biz_entity_relation_target;
DROP INDEX IF EXISTS idx_biz_entity_relation_source;
DROP POLICY IF EXISTS tenant_isolation ON biz_entity_relation;
ALTER TABLE biz_entity_relation DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS biz_entity_relation;


-- 2. 删除扩展属性表
DROP POLICY IF EXISTS tenant_isolation ON task_ext;
ALTER TABLE task_ext DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS task_ext;

DROP POLICY IF EXISTS tenant_isolation ON requirement_ext;
ALTER TABLE requirement_ext DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS requirement_ext;

DROP POLICY IF EXISTS tenant_isolation ON defect_ext;
ALTER TABLE defect_ext DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS defect_ext;


-- 3. 删除task表
DROP INDEX IF EXISTS idx_task_assignee;
DROP INDEX IF EXISTS idx_task_sort;
DROP INDEX IF EXISTS idx_task_fts;
DROP INDEX IF EXISTS idx_task_updated;
DROP INDEX IF EXISTS idx_task_target_date;
DROP INDEX IF EXISTS idx_task_parent;
DROP INDEX IF EXISTS idx_task_project_sprint;
DROP INDEX IF EXISTS idx_task_project_state;
DROP POLICY IF EXISTS tenant_isolation ON task;
ALTER TABLE task DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS task;


-- 4. 删除requirement表
DROP INDEX IF EXISTS idx_requirement_sort;
DROP INDEX IF EXISTS idx_requirement_fts;
DROP INDEX IF EXISTS idx_requirement_updated;
DROP INDEX IF EXISTS idx_requirement_target_date;
DROP INDEX IF EXISTS idx_requirement_parent;
DROP INDEX IF EXISTS idx_requirement_project_sprint;
DROP INDEX IF EXISTS idx_requirement_project_state;
DROP POLICY IF EXISTS tenant_isolation ON requirement;
ALTER TABLE requirement DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS requirement;


-- 5. 删除defect表
DROP INDEX IF EXISTS idx_defect_root_cause;
DROP INDEX IF EXISTS idx_defect_severity;
DROP INDEX IF EXISTS idx_defect_sort;
DROP INDEX IF EXISTS idx_defect_fts;
DROP INDEX IF EXISTS idx_defect_updated;
DROP INDEX IF EXISTS idx_defect_target_date;
DROP INDEX IF EXISTS idx_defect_parent;
DROP INDEX IF EXISTS idx_defect_project_sprint;
DROP INDEX IF EXISTS idx_defect_project_state;
DROP POLICY IF EXISTS tenant_isolation ON defect;
ALTER TABLE defect DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS defect;


-- 6. 恢复states表
DROP INDEX IF EXISTS idx_states_applicable_types;
ALTER TABLE states DROP COLUMN IF EXISTS applicable_types;


-- 7. 恢复state_transitions表
DROP INDEX IF EXISTS idx_state_transitions_type;
ALTER TABLE state_transitions ALTER COLUMN type_code SET DEFAULT '';
