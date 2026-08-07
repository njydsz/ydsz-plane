-- 0017_notification_dispatcher.down.sql
DROP INDEX IF EXISTS idx_deliveries_next_retry;
ALTER TABLE notification_deliveries DROP COLUMN IF EXISTS next_retry_at;
