-- 0017_notification_dispatcher: 多渠道重试元数据 + IM Webhook 环境变量读取 (Sprint 9 — M5)
-- 目标：StartDispatchWorker 在重试时基于 next_retry_at 判断是否到点

ALTER TABLE notification_deliveries
    ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_deliveries_next_retry
    ON notification_deliveries(status, next_retry_at)
    WHERE status = 'pending';
