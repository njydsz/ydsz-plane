-- 0009_workbench.down.sql: Revert workbench schema changes
DROP TRIGGER IF EXISTS trg_recent_items_touch ON recent_items;
DROP FUNCTION IF EXISTS fn_touch_recent_item();
DROP TRIGGER IF EXISTS trg_workbench_configs_updated_at ON workbench_configs;
DROP TABLE IF EXISTS workbench_templates;
DROP TABLE IF EXISTS recent_items;
DROP TABLE IF EXISTS workbench_configs;
