-- ==========================================================================
-- 自动化内置模板补齐 (automation_templates)
--
-- 现有 ydsz-plane-init.sql 已包含 7 条模板 (ID 1-7):
--   auto-complete-parent, overdue-reminder, version-release-transition,
--   epic-points-rollup, auto-start-date, auto-assign-least-loaded,
--   defect-notify-tech-lead。
--
-- 本迁移补齐 BuiltInTemplates() 中定义的剩余 8 条模板 (ID 8-15)，
-- 覆盖质量/效率/通知/管理四类场景：
--   - defect-assign-verifier  缺陷修复后自动通知创建者验证
--   - auto-set-priority       含"紧急"关键词自动设为高优
--   - status-change-notify-watchers  状态变更通知关注人
--   - sprint-complete-summary 迭代完成通知团队
--   - sprint-auto-start-issues 迭代启动后自动流转待办项
--   - auto-archive-old-issues 长期未更新已完成项自动归档
--   - duplicate-issue-check   新建工作项时检测重复提醒
--   - new-member-welcome      新成员加入通知
--
-- 设计决策：
--   - 使用 ON CONFLICT (slug) DO UPDATE 幂等，重复执行不会重复插入
--   - 比 SQL init 文件更便于增量维护；seed 脚本不再重复写模板数据
-- ==========================================================================

-- 唯一约束检查（若创建表时未定义）
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uk_automation_templates_slug') THEN
    ALTER TABLE automation_templates ADD CONSTRAINT uk_automation_templates_slug UNIQUE (slug);
  END IF;
END$$;

INSERT INTO automation_templates (slug, name, description, category, dsl_template, icon, sort_order, is_recommended, created_at)
VALUES (
  'defect-assign-verifier',
  '缺陷修复后自动指派验证人',
  '缺陷修复后自动将验证任务指派给创建者',
  'quality',
  '{"conditions": [], "trigger": {"type": "issue.status_changed", "filter": {"type_code": "defect", "to_group": "completed"}}, "actions": [{"type": "notify", "target": "${issue.created_by}", "channel": "in_app", "template": "缺陷 {{issue.identifier}} 已修复，请验证"}]}',
  'check-circle', 8, false, now()
)
ON CONFLICT (slug) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  category = EXCLUDED.category,
  dsl_template = EXCLUDED.dsl_template,
  icon = EXCLUDED.icon,
  sort_order = EXCLUDED.sort_order,
  is_recommended = EXCLUDED.is_recommended;

INSERT INTO automation_templates (slug, name, description, category, dsl_template, icon, sort_order, is_recommended, created_at)
VALUES (
  'auto-set-priority',
  '高优需求自动标记',
  '根据关键词自动设置工作项优先级',
  'efficiency',
  '{"conditions": [{"op": "contains", "field": "issue.name", "value": "紧急"}], "trigger": {"type": "issue.created"}, "actions": [{"type": "update_field", "field": "priority", "value": "urgent"}]}',
  'zap', 9, false, now()
)
ON CONFLICT (slug) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  category = EXCLUDED.category,
  dsl_template = EXCLUDED.dsl_template,
  icon = EXCLUDED.icon,
  sort_order = EXCLUDED.sort_order,
  is_recommended = EXCLUDED.is_recommended;

INSERT INTO automation_templates (slug, name, description, category, dsl_template, icon, sort_order, is_recommended, created_at)
VALUES (
  'status-change-notify-watchers',
  '状态变更通知关注人',
  '工作项状态变更时通知所有关注人',
  'notification',
  '{"conditions": [], "trigger": {"type": "issue.status_changed"}, "actions": [{"type": "notify", "target": "${issue.watchers}", "channel": "in_app", "template": "{{issue.identifier}} 状态变更为 {{issue.state_name}}"}]}',
  'bell', 10, false, now()
)
ON CONFLICT (slug) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  category = EXCLUDED.category,
  dsl_template = EXCLUDED.dsl_template,
  icon = EXCLUDED.icon,
  sort_order = EXCLUDED.sort_order,
  is_recommended = EXCLUDED.is_recommended;

INSERT INTO automation_templates (slug, name, description, category, dsl_template, icon, sort_order, is_recommended, created_at)
VALUES (
  'sprint-complete-summary',
  '迭代完成自动通知团队',
  '迭代完成时自动通知所有成员并发送总结',
  'notification',
  '{"conditions": [], "trigger": {"type": "sprint.completed"}, "actions": [{"type": "notify", "target": "${project.members}", "channel": "in_app", "template": "迭代 {{sprint.name}} 已完成"}]}',
  'flag', 11, false, now()
)
ON CONFLICT (slug) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  category = EXCLUDED.category,
  dsl_template = EXCLUDED.dsl_template,
  icon = EXCLUDED.icon,
  sort_order = EXCLUDED.sort_order,
  is_recommended = EXCLUDED.is_recommended;

INSERT INTO automation_templates (slug, name, description, category, dsl_template, icon, sort_order, is_recommended, created_at)
VALUES (
  'sprint-auto-start-issues',
  '迭代启动后自动开始工作项',
  '迭代启动后，自动将所有待办工作项流转到进行中',
  'management',
  '{"conditions": [{"op": "eq", "field": "state.group", "value": "todo"}], "trigger": {"type": "sprint.started"}, "actions": [{"type": "transition", "field": "state", "value": "started"}]}',
  'play-circle', 12, false, now()
)
ON CONFLICT (slug) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  category = EXCLUDED.category,
  dsl_template = EXCLUDED.dsl_template,
  icon = EXCLUDED.icon,
  sort_order = EXCLUDED.sort_order,
  is_recommended = EXCLUDED.is_recommended;

INSERT INTO automation_templates (slug, name, description, category, dsl_template, icon, sort_order, is_recommended, created_at)
VALUES (
  'auto-archive-old-issues',
  '长期未更新工作项自动归档',
  '超过 30 天未更新的已完成工作项自动归档',
  'management',
  '{"conditions": [{"op": "eq", "field": "state.group", "value": "completed"}, {"op": "lt", "field": "issue.updated_at", "value": "now-30d"}], "trigger": {"type": "scheduled", "cron": "0 2 * * *"}, "actions": [{"type": "update_field", "field": "is_archived", "value": "true"}]}',
  'archive', 13, false, now()
)
ON CONFLICT (slug) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  category = EXCLUDED.category,
  dsl_template = EXCLUDED.dsl_template,
  icon = EXCLUDED.icon,
  sort_order = EXCLUDED.sort_order,
  is_recommended = EXCLUDED.is_recommended;

INSERT INTO automation_templates (slug, name, description, category, dsl_template, icon, sort_order, is_recommended, created_at)
VALUES (
  'duplicate-issue-check',
  '重复工作项提醒',
  '新建工作项时检测可能的重复项并提醒',
  'management',
  '{"conditions": [], "trigger": {"type": "issue.created"}, "actions": [{"type": "notify", "target": "${issue.created_by}", "channel": "in_app", "template": "⚠️ 检测到可能的重复工作项请确认"}]}',
  'copy', 14, false, now()
)
ON CONFLICT (slug) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  category = EXCLUDED.category,
  dsl_template = EXCLUDED.dsl_template,
  icon = EXCLUDED.icon,
  sort_order = EXCLUDED.sort_order,
  is_recommended = EXCLUDED.is_recommended;

INSERT INTO automation_templates (slug, name, description, category, dsl_template, icon, sort_order, is_recommended, created_at)
VALUES (
  'new-member-welcome',
  '新成员加入通知',
  '工作空间有新成员加入时通知所有成员',
  'management',
  '{"conditions": [], "trigger": {"type": "member.added"}, "actions": [{"type": "notify", "target": "${workspace.members}", "channel": "in_app", "template": "欢迎 {{actor.user_name}} 加入工作空间"}]}',
  'user-plus', 15, false, now()
)
ON CONFLICT (slug) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  category = EXCLUDED.category,
  dsl_template = EXCLUDED.dsl_template,
  icon = EXCLUDED.icon,
  sort_order = EXCLUDED.sort_order,
  is_recommended = EXCLUDED.is_recommended;
