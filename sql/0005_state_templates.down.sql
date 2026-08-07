-- 0005_state_templates.down: 回滚 template_set 列
ALTER TABLE projects DROP COLUMN IF EXISTS template_set;
