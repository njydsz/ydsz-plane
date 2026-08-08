-- 工作项数据迁移脚本：将旧issues表的数据按type_code拆分到新表
-- 遵循项目迁移只增不改规则，保留旧表数据不变，仅做数据同步
-- 所有新表字段已包含旧表所有公共字段，特殊字段已做对应映射

-- 1. 迁移task类型工作项
INSERT INTO task (
    public_id, workspace_id, project_id, sequence_id, parent_id, depth,
    name, description_json, description_html, state_id, priority,
    category, actual_effort, remaining_effort, delay_reason,
    point, estimate_point_id, sprint_id, version_id, progress,
    start_date, target_date, completed_at, is_draft, sort_order,
    created_by, created_at, updated_at, deleted_at
)
SELECT 
    public_id, workspace_id, project_id, sequence_id, parent_id, depth,
    name, description_json, description_html, state_id, priority,
    category, actual_effort, remaining_effort, delay_reason,
    point, estimate_point_id, sprint_id, version_id, progress,
    start_date, target_date, completed_at, is_draft, sort_order,
    created_by, created_at, updated_at, deleted_at
FROM issues 
WHERE type_code = 'task'
ON CONFLICT (project_id, sequence_id) DO NOTHING; -- 避免重复插入，跳过已存在的记录


-- 2. 迁移requirement类型工作项
INSERT INTO requirement (
    public_id, workspace_id, project_id, sequence_id, parent_id, depth,
    name, description_json, description_html, state_id, priority,
    source, point, estimate_point_id, sprint_id, version_id, progress,
    start_date, target_date, completed_at, is_draft, sort_order,
    created_by, created_at, updated_at, deleted_at
)
SELECT 
    public_id, workspace_id, project_id, sequence_id, parent_id, depth,
    name, description_json, description_html, state_id, priority,
    source, point, estimate_point_id, sprint_id, version_id, progress,
    start_date, target_date, completed_at, is_draft, sort_order,
    created_by, created_at, updated_at, deleted_at
FROM issues 
WHERE type_code = 'requirement'
ON CONFLICT (project_id, sequence_id) DO NOTHING;


-- 3. 迁移defect类型工作项
INSERT INTO defect (
    public_id, workspace_id, project_id, sequence_id, parent_id, depth,
    name, description_json, description_html, state_id, priority,
    severity, found_phase, found_version_id, fix_version_id, root_cause_category,
    verifier_id, environment, reproduce_steps,
    point, estimate_point_id, sprint_id, version_id, progress,
    start_date, target_date, completed_at, is_draft, sort_order,
    created_by, created_at, updated_at, deleted_at
)
SELECT 
    public_id, workspace_id, project_id, sequence_id, parent_id, depth,
    name, description_json, description_html, state_id, priority,
    severity, found_phase, found_version_id, fix_version_id, root_cause_category,
    verifier_id, environment, reproduce_steps,
    point, estimate_point_id, sprint_id, version_id, progress,
    start_date, target_date, completed_at, is_draft, sort_order,
    created_by, created_at, updated_at, deleted_at
FROM issues 
WHERE type_code = 'defect'
ON CONFLICT (project_id, sequence_id) DO NOTHING;


-- 4. 迁移关联关系数据
INSERT INTO biz_entity_relation (
    workspace_id, project_id, source_type, source_id, target_type, target_id,
    relation_type, created_by, created_at
)
SELECT 
    ir.workspace_id, ir.project_id,
    CASE WHEN i1.type_code = 'task' THEN 'task' WHEN i1.type_code = 'requirement' THEN 'requirement' ELSE 'defect' END as source_type,
    ir.source_id as source_id,
    CASE WHEN i2.type_code = 'task' THEN 'task' WHEN i2.type_code = 'requirement' THEN 'requirement' ELSE 'defect' END as target_type,
    ir.target_id as target_id,
    ir.relation_type, ir.created_by, ir.created_at
FROM issue_relations ir
JOIN issues i1 ON ir.source_id = i1.id
JOIN issues i2 ON ir.target_id = i2.id
ON CONFLICT (source_type, source_id, target_type, target_id, relation_type) DO NOTHING;


-- 5. 注意：该迁移脚本只做初始数据同步，后续双写逻辑由应用层代码实现，完成后可执行0027_work_item_migration_complete.sql清理旧表读写逻辑
