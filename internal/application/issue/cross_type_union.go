// Package issue — 三表联动 SQL 片段与常量（聚合根拆分后的跨表查询基础设施）。
package issue

// CrossTypeWorkitemUnion 是 requirement/task/defect 三表 UNION ALL 的标准列投影。
// 用于替换原来基于统一 issues 表的任何只读聚合查询（dashboard / metrics / AI / workbench 等）。
// 列顺序/类型必须与三个独立表严格对齐（利用 NULL::xxx 补齐差异列），保证 UNION ALL 兼容。
const CrossTypeWorkitemUnion = `(SELECT
  id, public_id, workspace_id, project_id, sequence_id,
  'requirement'::text AS type_code,
  parent_id, depth, name, description_json, description_html, description_stripped,
  state_id, priority,
  NULL::smallint AS severity, NULL::text AS found_phase,
  NULL::text AS root_cause_category, NULL::bigint AS verifier_id,
  NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps,
  NULL::text AS category, NULL::numeric AS actual_effort,
  NULL::numeric AS remaining_effort, NULL::text AS delay_reason,
  source, point, sprint_id, progress,
  start_date, target_date, completed_at,
  is_draft, sort_order, version, version_id,
  NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id,
  created_by, created_at, updated_at, deleted
FROM requirement WHERE deleted = false
UNION ALL
SELECT
  id, public_id, workspace_id, project_id, sequence_id,
  'task'::text,
  parent_id, depth, name, description_json, description_html, description_stripped,
  state_id, priority,
  NULL::smallint, NULL::text, NULL::text, NULL::bigint,
  NULL::jsonb, NULL::jsonb,
  category, actual_effort, remaining_effort, delay_reason,
  NULL::text, point, sprint_id, progress,
  start_date, target_date, completed_at,
  is_draft, sort_order, version, version_id,
  NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id,
  created_by, created_at, updated_at, deleted
FROM task WHERE deleted = false
UNION ALL
SELECT
  id, public_id, workspace_id, project_id, sequence_id,
  'defect'::text,
  parent_id, depth, name, description_json, description_html, description_stripped,
  state_id, priority,
  severity, found_phase, root_cause_category, verifier_id,
  environment, reproduce_steps,
  NULL::text, NULL::numeric, NULL::numeric, NULL::text,
  NULL::text, point, sprint_id, progress,
  start_date, target_date, completed_at,
  is_draft, sort_order, version, version_id,
  found_version_id, fix_version_id,
  created_by, created_at, updated_at, deleted
FROM defect WHERE deleted = false)`

// WorkitemUnionAlias 是 UNION ALL 子查询的标准别名（SQL 中与 CrossTypeWorkitemUnion 配套使用）。
const WorkitemUnionAlias = "w"
