-- 0015_automation.down.sql: Revert automation schema (Sprint 11)
DROP TABLE IF EXISTS automation_templates;
DROP TABLE IF EXISTS rule_executions;
DROP TRIGGER IF EXISTS trg_automation_rules_updated_at ON automation_rules;
DROP TABLE IF EXISTS automation_rules;
