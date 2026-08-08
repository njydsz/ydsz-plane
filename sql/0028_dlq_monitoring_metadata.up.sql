-- ============================================================================
-- 迁移 0028: DLQ 监控元数据表
-- 用途: 记录 Relay 发布失败、消费者 NACK 路由到 DLX 的消息元信息，
--       供管理界面展示与手动重放。
-- 设计:
--   - 单条死信一条记录（事件 ID + 队列 + 路由键 + 错误原因 + 创建时间）
--   - 重试 / 清理后标记 resolved_at，查询时过滤
--   - 7 天后自动清理 resolved 记录
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.dlq_events (
    id           BIGSERIAL PRIMARY KEY,
    event_id     BIGINT,
    workspace_id BIGINT,
    queue        TEXT   NOT NULL,
    exchange     TEXT   NOT NULL,
    routing_key  TEXT   NOT NULL DEFAULT '',
    payload      JSONB,
    error_reason TEXT,
    resolved_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_by  TEXT,
    CONSTRAINT uq_dlq_event_per_queue UNIQUE (event_id, queue)
);

CREATE INDEX IF NOT EXISTS idx_dlq_workspace_active   ON public.dlq_events (workspace_id, resolved_at) WHERE resolved_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dlq_event_unresolved   ON public.dlq_events (created_at DESC)        WHERE resolved_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dlq_event_resolved_at ON public.dlq_events (resolved_at)            WHERE resolved_at IS NOT NULL;

COMMENT ON TABLE public.dlq_events IS 'DLQ 死信事件元数据表（记录 Relay 发布失败 + 消费者 NACK 路由到 DLX 的消息；管理界面展示与重放）';
COMMENT ON COLUMN public.dlq_events.event_id IS '关联 domain_events.id（域事件主键）';
COMMENT ON COLUMN public.dlq_events.queue IS '死信消息所在的 RabbitMQ 队列名';
COMMENT ON COLUMN public.dlq_events.resolved_at IS '重试/清理完成后标记时间（NULL=待处理）';
COMMENT ON COLUMN public.dlq_events.resolved_by IS '处理该死信的管理员标识';
