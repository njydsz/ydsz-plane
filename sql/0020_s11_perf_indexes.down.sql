-- S11 性能优化索引回滚
DROP INDEX CONCURRENTLY IF EXISTS "idx_automation_rules_trigger_active";
DROP INDEX CONCURRENTLY IF EXISTS "idx_rule_executions_rule_created";
DROP INDEX CONCURRENTLY IF EXISTS "idx_rule_executions_idempotent";
DROP INDEX CONCURRENTLY IF EXISTS "idx_metric_snap_trend";
DROP INDEX CONCURRENTLY IF EXISTS "idx_deployment_dora";
