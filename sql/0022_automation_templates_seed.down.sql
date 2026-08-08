-- ==========================================================================
-- 回滚：移除迁移 0022 补入的 8 条模板 (slug ID 8-15)
-- ==========================================================================

DELETE FROM automation_templates WHERE slug = 'defect-assign-verifier';
DELETE FROM automation_templates WHERE slug = 'auto-set-priority';
DELETE FROM automation_templates WHERE slug = 'status-change-notify-watchers';
DELETE FROM automation_templates WHERE slug = 'sprint-complete-summary';
DELETE FROM automation_templates WHERE slug = 'sprint-auto-start-issues';
DELETE FROM automation_templates WHERE slug = 'auto-archive-old-issues';
DELETE FROM automation_templates WHERE slug = 'duplicate-issue-check';
DELETE FROM automation_templates WHERE slug = 'new-member-welcome';
