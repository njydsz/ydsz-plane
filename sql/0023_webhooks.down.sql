-- 0015_webhooks.down.sql
BEGIN;
DROP TABLE IF EXISTS webhook_logs;
DROP TABLE IF EXISTS webhooks;
COMMIT;
