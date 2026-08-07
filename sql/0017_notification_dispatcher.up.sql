-- 0017_notification_dispatcher: 通知多渠道投递 Worker 支撑
-- 为 notification_deliveries 增加下次重试时间戳 + 部分索引，
-- 方便 Dispatcher 轮询 pending 记录。

BEGIN;

ALTER TABLE notification_deliveries
    ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_deliveries_next_retry
    ON notification_deliveries (status, next_retry_at)
    WHERE status = 'pending';

COMMIT;
