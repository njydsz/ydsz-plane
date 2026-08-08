-- Ydsz Plane S14 Phase-1：通知服务独立数据库 Schema
--
-- 本文件从 ydsz-plane-init.sql 提取通知域独有表结构，
-- 用于创建独立的 notification_db（与单体 Postgres 隔离）。
--
-- 执行方式：
--   psql $DATABASE_URL -f sql/notification_schema_v1.sql
--
-- 前置条件：
--   - notification_db 数据库已创建（CREATE DATABASE notifications;）
--   - 当前连接已具有该数据库的 DDL 权限

BEGIN;

-- ============================================================
-- 1. 序列（Sequences）
-- ============================================================
CREATE SEQUENCE IF NOT EXISTS notifications_id_seq           START 1;
CREATE SEQUENCE IF NOT EXISTS notification_deliveries_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS notification_digests_id_seq    START 1;
CREATE SEQUENCE IF NOT EXISTS notification_preferences_id_seq START 1;

-- ============================================================
-- 2. 站内信主表
-- ============================================================
CREATE TABLE IF NOT EXISTS notifications (
  id           int8 NOT NULL DEFAULT nextval('notifications_id_seq'::regclass),
  workspace_id int8 NOT NULL,
  recipient_id int8 NOT NULL,
  event_type   varchar(64)   NOT NULL,
  entity_type  varchar(32)   NOT NULL,
  entity_id    int8          NOT NULL,
  title        varchar(256)  NOT NULL,
  body         text,
  action_url   varchar(512),
  actor_id     int8,
  actor_name   varchar(128),
  is_read      bool NOT NULL DEFAULT false,
  is_archived  bool NOT NULL DEFAULT false,
  read_at      timestamptz(6),
  channel      varchar(32)  NOT NULL DEFAULT 'in_app',
  payload      jsonb        DEFAULT '{}'::jsonb,
  created_at   timestamptz(6) NOT NULL DEFAULT now(),
  PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_notifications_recipient_workspace
  ON notifications (recipient_id, workspace_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_workspace_unread
  ON notifications (workspace_id, recipient_id, is_read)
  WHERE is_read = false AND is_archived = false;

-- ============================================================
-- 3. 投递渠道记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS notification_deliveries (
  id             int8 NOT NULL DEFAULT nextval('notification_deliveries_id_seq'::regclass),
  notification_id int8 NOT NULL,
  channel         varchar(32) NOT NULL,
  status          varchar(16) NOT NULL DEFAULT 'pending',
  recipient       text        NOT NULL,
  sent_at         timestamptz(6),
  error_msg       text,
  retry_count     int4 NOT NULL DEFAULT 0,
  created_at      timestamptz(6) NOT NULL DEFAULT now(),
  next_retry_at   timestamptz(6),
  PRIMARY KEY (id),
  CONSTRAINT notification_deliveries_notification_id_fkey
    FOREIGN KEY (notification_id) REFERENCES notifications (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
);

CREATE INDEX IF NOT EXISTS idx_deliveries_notification
  ON notification_deliveries (notification_id);
CREATE INDEX IF NOT EXISTS idx_deliveries_status_retry
  ON notification_deliveries (status, next_retry_at)
  WHERE status IN ('pending', 'retrying');

-- ============================================================
-- 4. 摘要聚合表（daily/weekly）
-- ============================================================
CREATE TABLE IF NOT EXISTS notification_digests (
  id               int8 NOT NULL DEFAULT nextval('notification_digests_id_seq'::regclass),
  user_id          int8 NOT NULL,
  workspace_id     int8 NOT NULL,
  digest_type      varchar(16) NOT NULL,
  notification_ids int8[] NOT NULL,
  status           varchar(16) NOT NULL DEFAULT 'pending',
  scheduled_for    timestamptz(6) NOT NULL,
  sent_at          timestamptz(6),
  created_at       timestamptz(6) NOT NULL DEFAULT now(),
  PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_digests_pending
  ON notification_digests (status, scheduled_for)
  WHERE status = 'pending';

-- ============================================================
-- 5. 用户偏好表
-- ============================================================
CREATE TABLE IF NOT EXISTS notification_preferences (
  id          int8 NOT NULL DEFAULT nextval('notification_preferences_id_seq'::regclass),
  user_id     int8 NOT NULL,
  workspace_id int8 NOT NULL,
  event_types jsonb NOT NULL DEFAULT '[]'::jsonb,
  channels    jsonb NOT NULL DEFAULT '["in_app"]'::jsonb,
  digest      varchar(16) NOT NULL DEFAULT 'realtime',
  dnd_enabled bool NOT NULL DEFAULT false,
  dnd_start   time(6) DEFAULT '22:00:00'::time,
  dnd_end     time(6) DEFAULT '08:00:00'::time,
  is_enabled  bool NOT NULL DEFAULT true,
  created_at  timestamptz(6) NOT NULL DEFAULT now(),
  updated_at  timestamptz(6) NOT NULL DEFAULT now(),
  PRIMARY KEY (id),
  CONSTRAINT notification_preferences_user_id_workspace_id_key
    UNIQUE (user_id, workspace_id)
);

-- ============================================================
-- 6. Row Level Security（如 notification_db 需要独立 RLS）
-- ============================================================
-- 由于通知服务拥有独立该 schema，暂不启用 RLS（简化部署）。
-- 如果未来需要，可参照 core_db 的 workspace_members RLS 机制添加。

COMMIT;
