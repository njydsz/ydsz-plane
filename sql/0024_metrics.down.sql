-- 0016_metrics.down.sql: Revert metrics schema (Sprint 11)
DROP TRIGGER IF EXISTS trg_metric_adjustments_updated_at ON metric_adjustments;
DROP TABLE IF EXISTS metric_adjustments;
DROP TABLE IF EXISTS deployment_events;
DROP TABLE IF EXISTS metric_snapshots;
