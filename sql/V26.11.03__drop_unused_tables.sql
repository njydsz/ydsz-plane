-- ===========================================================================
-- V26.11.03 · 清理未使用的死表（前后端均无 CRUD 的设计残留表）
-- 依据：docs/optimization/S20-数据库表·接口测试·前后端贯通现状分析与优化建议.md §1.3.1
-- ===========================================================================
-- 背景：
--   以下 9 张表在 ydsz-plane-init.sql 中定义，但全量 Go 代码（internal/）与
--   Vue 代码（web/src/）中零引用，且从未有数据写入。它们是早期设计残留，
--   已被后续演进中的其他表替代：
--
--   menus / role_menus        → RBAC 改用 roles + role_permissions 表
--   user_roles                → 改用 workspace_members.role + users.role_scope
--   tenant_members            → 改用 workspace_members
--   user_preferences          → 改用 view_preferences / notification_preferences
--   idempotency_keys          → Outbox 改用 processed_events
--   notification_subscriptions → 改用 notification_preferences
--   share_links               → 改用 page_shares
--   documents                 → 改用 pages 表
--
-- 安全前置：DROP TABLE 会自动删除依附于该表的 RLS 策略、索引、约束。
--           本迁移使用 IF EXISTS 幂等，可重复执行。
--           已合入环境的表确认无外键引用（init.sql 中无 REFERENCES 指向它们）。
-- ===========================================================================

DROP TABLE IF EXISTS role_menus CASCADE;
DROP TABLE IF EXISTS menus CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS tenant_members CASCADE;
DROP TABLE IF EXISTS user_preferences CASCADE;
DROP TABLE IF EXISTS idempotency_keys CASCADE;
DROP TABLE IF EXISTS notification_subscriptions CASCADE;
DROP TABLE IF EXISTS share_links CASCADE;
DROP TABLE IF EXISTS documents CASCADE;
