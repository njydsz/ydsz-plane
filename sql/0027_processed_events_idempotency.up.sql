-- ============================================================================
-- 迁移 0027: processed_events 消费者幂等去重表
-- 用途: 消费者记录已成功处理的事件 ID，实现 at-least-once 投递下的幂等跳过。
-- 设计:
--   - (event_id, consumer_id) 唯一索引 → upsert 为 INSERT ON CONFLICT DO UPDATE
--   - retry_count 累计该事件被同一消费者处理的次数
--   - 30 天后自动清理，避免无限膨胀
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.processed_events (
    event_id    BIGINT NOT NULL,
    consumer_id TEXT   NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    retry_count INT    NOT NULL DEFAULT 1,
    PRIMARY KEY (event_id, consumer_id)
);

CREATE INDEX IF NOT EXISTS idx_processed_events_consumer_time
    ON public.processed_events (consumer_id, processed_at);

COMMENT ON TABLE public.processed_events IS '消费者幂等去重表（事件 at-least-once 投递下防重复处理；按 processed_at 30 天清理）';
COMMENT ON COLUMN public.processed_events.event_id IS '领域事件 ID（引用 domain_events.id）';
COMMENT ON COLUMN public.processed_events.consumer_id IS '消费者标识（如 notification-dispatcher / webhook-dispatcher）';
COMMENT ON COLUMN public.processed_events.processed_at IS '上次处理时间（用于过期清理）';
COMMENT ON COLUMN public.processed_events.retry_count IS '该事件被同一消费者累计处理次数（含重放）';
