-- 0010_dashboard.down.sql: Revert dashboard schema
DROP TABLE IF EXISTS dashboard_templates;
DROP TRIGGER IF EXISTS trg_risk_rules_updated_at ON risk_rules;
DROP TABLE IF EXISTS risk_alerts;
DROP TABLE IF EXISTS risk_rules;
DROP TRIGGER IF EXISTS trg_dashboard_widgets_updated_at ON dashboard_widgets;
DROP TABLE IF EXISTS dashboard_snapshots;
DROP TABLE IF EXISTS dashboard_widgets;
