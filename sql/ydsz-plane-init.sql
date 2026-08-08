/*
 Navicat Premium Dump SQL

 Source Server         : 127.0.0.1
 Source Server Type    : PostgreSQL
 Source Server Version : 180004 (180004)
 Source Host           : 127.0.0.1:5432
 Source Catalog        : ydsz-plane
 Source Schema         : public

 Target Server Type    : PostgreSQL
 Target Server Version : 180004 (180004)
 File Encoding         : 65001

 Date: 08/08/2026 00:31:56
*/


-- ----------------------------
-- Sequence structure for api_tokens_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."api_tokens_id_seq";
CREATE SEQUENCE "public"."api_tokens_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for attachments_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."attachments_id_seq";
CREATE SEQUENCE "public"."attachments_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for audit_logs_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."audit_logs_id_seq";
CREATE SEQUENCE "public"."audit_logs_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for automation_rules_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."automation_rules_id_seq";
CREATE SEQUENCE "public"."automation_rules_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for automation_templates_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."automation_templates_id_seq";
CREATE SEQUENCE "public"."automation_templates_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for dashboard_snapshots_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."dashboard_snapshots_id_seq";
CREATE SEQUENCE "public"."dashboard_snapshots_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for dashboard_templates_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."dashboard_templates_id_seq";
CREATE SEQUENCE "public"."dashboard_templates_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for dashboard_widgets_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."dashboard_widgets_id_seq";
CREATE SEQUENCE "public"."dashboard_widgets_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for deployment_events_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."deployment_events_id_seq";
CREATE SEQUENCE "public"."deployment_events_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for domain_events_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."domain_events_id_seq";
CREATE SEQUENCE "public"."domain_events_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for intake_channels_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."intake_channels_id_seq";
CREATE SEQUENCE "public"."intake_channels_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for intake_issues_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."intake_issues_id_seq";
CREATE SEQUENCE "public"."intake_issues_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for invitations_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."invitations_id_seq";
CREATE SEQUENCE "public"."invitations_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for issue_activities_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."issue_activities_id_seq";
CREATE SEQUENCE "public"."issue_activities_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for issue_comments_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."issue_comments_id_seq";
CREATE SEQUENCE "public"."issue_comments_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for issue_dependencies_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."issue_dependencies_id_seq";
CREATE SEQUENCE "public"."issue_dependencies_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for issue_relations_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."issue_relations_id_seq";
CREATE SEQUENCE "public"."issue_relations_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for issues_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."issues_id_seq";
CREATE SEQUENCE "public"."issues_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for labels_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."labels_id_seq";
CREATE SEQUENCE "public"."labels_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for metric_adjustments_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."metric_adjustments_id_seq";
CREATE SEQUENCE "public"."metric_adjustments_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for metric_snapshots_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."metric_snapshots_id_seq";
CREATE SEQUENCE "public"."metric_snapshots_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for modules_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."modules_id_seq";
CREATE SEQUENCE "public"."modules_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for notification_deliveries_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."notification_deliveries_id_seq";
CREATE SEQUENCE "public"."notification_deliveries_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for notification_digests_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."notification_digests_id_seq";
CREATE SEQUENCE "public"."notification_digests_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for notification_preferences_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."notification_preferences_id_seq";
CREATE SEQUENCE "public"."notification_preferences_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for notifications_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."notifications_id_seq";
CREATE SEQUENCE "public"."notifications_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for password_reset_tokens_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."password_reset_tokens_id_seq";
CREATE SEQUENCE "public"."password_reset_tokens_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for projects_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."projects_id_seq";
CREATE SEQUENCE "public"."projects_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for recent_items_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."recent_items_id_seq";
CREATE SEQUENCE "public"."recent_items_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for risk_alerts_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."risk_alerts_id_seq";
CREATE SEQUENCE "public"."risk_alerts_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for risk_rules_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."risk_rules_id_seq";
CREATE SEQUENCE "public"."risk_rules_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for rule_executions_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."rule_executions_id_seq";
CREATE SEQUENCE "public"."rule_executions_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for search_bookmarks_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."search_bookmarks_id_seq";
CREATE SEQUENCE "public"."search_bookmarks_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for search_documents_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."search_documents_id_seq";
CREATE SEQUENCE "public"."search_documents_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for search_history_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."search_history_id_seq";
CREATE SEQUENCE "public"."search_history_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for sprint_snapshots_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."sprint_snapshots_id_seq";
CREATE SEQUENCE "public"."sprint_snapshots_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for sprints_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."sprints_id_seq";
CREATE SEQUENCE "public"."sprints_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for state_transitions_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."state_transitions_id_seq";
CREATE SEQUENCE "public"."state_transitions_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for states_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."states_id_seq";
CREATE SEQUENCE "public"."states_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for time_logs_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."time_logs_id_seq";
CREATE SEQUENCE "public"."time_logs_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for users_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."users_id_seq";
CREATE SEQUENCE "public"."users_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for version_delivery_snapshots_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."version_delivery_snapshots_id_seq";
CREATE SEQUENCE "public"."version_delivery_snapshots_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for versions_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."versions_id_seq";
CREATE SEQUENCE "public"."versions_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for view_preferences_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."view_preferences_id_seq";
CREATE SEQUENCE "public"."view_preferences_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for webhook_logs_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."webhook_logs_id_seq";
CREATE SEQUENCE "public"."webhook_logs_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for webhooks_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."webhooks_id_seq";
CREATE SEQUENCE "public"."webhooks_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for workbench_configs_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."workbench_configs_id_seq";
CREATE SEQUENCE "public"."workbench_configs_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for workbench_templates_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."workbench_templates_id_seq";
CREATE SEQUENCE "public"."workbench_templates_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for workspaces_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."workspaces_id_seq";
CREATE SEQUENCE "public"."workspaces_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Table structure for api_tokens
-- ----------------------------
DROP TABLE IF EXISTS "public"."api_tokens";
CREATE TABLE "public"."api_tokens" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "user_id" int8 NOT NULL,
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "token_hash" text COLLATE "pg_catalog"."default" NOT NULL,
  "scopes" jsonb NOT NULL DEFAULT '["read:workspace"]'::jsonb,
  "last_used_at" timestamptz(6),
  "expires_at" timestamptz(6),
  "revoked_at" timestamptz(6),
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.api_tokens IS 'API Token 表（scopes 白名单；仅存 SHA-256 hash；支持过期 + last_used_at 审计）';
COMMENT ON COLUMN public.api_tokens.id IS '主键 ID';
COMMENT ON COLUMN public.api_tokens.user_id IS '用户 FK';
COMMENT ON COLUMN public.api_tokens.name IS 'Token 自定义名称（方便管理/撤销；如"CI Pipeline"）';
COMMENT ON COLUMN public.api_tokens.token_hash IS 'SHA-256 hash 值（ydz_ 前缀明文仅展示一次；存入 hash 校验）';
COMMENT ON COLUMN public.api_tokens.scopes IS '权限范围 JSONB ["read:issues", "write:issues", "admin:*"]；白名单';
COMMENT ON COLUMN public.api_tokens.expires_at IS '过期时间 TIMESTAMPTZ（NULL=不过期；推荐设置 ≤90d）';
COMMENT ON COLUMN public.api_tokens.last_used_at IS '最后使用时间（活跃度审计；长期未用可告警/建议吊销）';
COMMENT ON COLUMN public.api_tokens.created_by IS '创建人 FK；管理与撤销监听';
COMMENT ON COLUMN public.api_tokens.created_at IS '创建时间';
COMMENT ON COLUMN public.api_tokens.revoked_at IS '吊销时间；NULL=有效；revoked_at 非空校验时拒绝';

-- 补齐旧 dump 中 COMMENT 引用但实际缺失的列
ALTER TABLE public.api_tokens
    ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES public.users(id),
    ADD COLUMN IF NOT EXISTS deleted_at  TIMESTAMPTZ;

COMMENT ON COLUMN public.api_tokens.created_by IS '创建人 FK；管理与撤销监听';
COMMENT ON COLUMN public.api_tokens.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Table structure for attachments
-- ----------------------------
DROP TABLE IF EXISTS "public"."attachments";
CREATE TABLE "public"."attachments" (
  "id" int8 NOT NULL DEFAULT nextval('attachments_id_seq'::regclass),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "entity_type" varchar(20) COLLATE "pg_catalog"."default" NOT NULL,
  "entity_id" int8 NOT NULL,
  "file_name" varchar(512) COLLATE "pg_catalog"."default" NOT NULL,
  "file_size" int8 NOT NULL,
  "content_type" varchar(128) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'application/octet-stream'::character varying,
  "storage_key" varchar(512) COLLATE "pg_catalog"."default" NOT NULL,
  "storage_url" varchar(2048) COLLATE "pg_catalog"."default",
  "thumb_key" varchar(512) COLLATE "pg_catalog"."default",
  "uploaded_by" int8 NOT NULL,
  "deleted_at" timestamptz(6),
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.attachments IS '附件表（多态关联: issue/comment/workspace/project/user；元数据 JSONB）';
COMMENT ON COLUMN public.attachments.id IS '主键 ID';
COMMENT ON COLUMN public.attachments.attachable_type IS '多态类型: issue / comment / workspace / project / user / intake_issue';
COMMENT ON COLUMN public.attachments.attachable_id IS '多态关联 ID（bigint 统一；配合 attachable_type 唯一）';
COMMENT ON COLUMN public.attachments.workspace_id IS '工作空间 FK（RLS 依据）';
COMMENT ON COLUMN public.attachments.file_name IS '原始文件名（用户上传时显示名）';
COMMENT ON COLUMN public.attachments.file_size IS '文件大小（字节；最大 10MB）';
COMMENT ON COLUMN public.attachments.content_type IS 'MIME content-type（类型白名单校验）';
COMMENT ON COLUMN public.attachments.storage_key IS 'MinIO 对象存储 key（UUID，文件名已重命名）';
COMMENT ON COLUMN public.attachments.metadata IS '附件元数据 JSONB（图片尺寸/时长/EXIF/CRC32 等）';
COMMENT ON COLUMN public.attachments.uploaded_by IS '上传人 FK';
COMMENT ON COLUMN public.attachments.created_at IS '上传时间';
COMMENT ON COLUMN public.attachments.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of attachments
-- ----------------------------

-- ----------------------------
-- Table structure for audit_logs
-- ----------------------------
DROP TABLE IF EXISTS "public"."audit_logs";
CREATE TABLE "public"."audit_logs" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8,
  "actor_id" int8,
  "action" text COLLATE "pg_catalog"."default" NOT NULL,
  "target" text COLLATE "pg_catalog"."default",
  "detail" jsonb,
  "ip" inet,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.audit_logs IS '审计日志表（只增不改；登录/权限变更/删除/Token/Webhook 等安全操作；在线 12 个月）';
COMMENT ON COLUMN public.audit_logs.id IS '主键 ID';
COMMENT ON COLUMN public.audit_logs.workspace_id IS '工作空间 FK（RLS 依据）';
COMMENT ON COLUMN public.audit_logs.actor_id IS '操作人 FK（users.id）；系统动作为 NULL';
COMMENT ON COLUMN public.audit_logs.action IS '操作名称（固定枚举: login / logout / permission_change / member_add / member_remove / token_revoke / webhook.create / data_export / setting.update / issue.delete）';
COMMENT ON COLUMN public.audit_logs.target_type IS '目标类型: workspace / project / issue / user / member / token / webhook';
COMMENT ON COLUMN public.audit_logs.target_id IS '目标 ID（BIGINT）';
-- FIXED: COMMENT ON COLUMN public.audit_logs.detail IS '变更前→后 diff JSONB（字段级 audit；如 {role:'member'→'admin'}）';
COMMENT ON COLUMN public.audit_logs.ip_address IS '客户端 IP（IPv4/IPv6；安全监控/异常登录识别）';
COMMENT ON COLUMN public.audit_logs.user_agent IS '客户端 UA 字符串（浏览器/设备识别）';
COMMENT ON COLUMN public.audit_logs.request_id IS '关联请求 ID（trace_id；日志/错误/链路追踪关联）';
COMMENT ON COLUMN public.audit_logs.created_at IS '操作时间 TIMESTAMPTZ（在线 12 个月 + 归档 3 年；只增不改）';

-- ----------------------------
-- Records of audit_logs
-- ----------------------------
INSERT INTO "public"."audit_logs" OVERRIDING SYSTEM VALUE VALUES (1, NULL, 1, 'seed.demo_event_0', 'demo-target-0', '{"note": "seeded for demo", "index": 0}', NULL, '2026-08-08 00:01:22.176717+08');
INSERT INTO "public"."audit_logs" OVERRIDING SYSTEM VALUE VALUES (2, NULL, 1, 'seed.demo_event_1', 'demo-target-1', '{"note": "seeded for demo", "index": 1}', NULL, '2026-08-08 00:01:22.181096+08');
INSERT INTO "public"."audit_logs" OVERRIDING SYSTEM VALUE VALUES (3, NULL, 1, 'seed.demo_event_2', 'demo-target-2', '{"note": "seeded for demo", "index": 2}', NULL, '2026-08-08 00:01:22.183852+08');
INSERT INTO "public"."audit_logs" OVERRIDING SYSTEM VALUE VALUES (4, NULL, 1, 'seed.demo_event_3', 'demo-target-3', '{"note": "seeded for demo", "index": 3}', NULL, '2026-08-08 00:01:22.186463+08');
INSERT INTO "public"."audit_logs" OVERRIDING SYSTEM VALUE VALUES (5, NULL, 1, 'seed.demo_event_4', 'demo-target-4', '{"note": "seeded for demo", "index": 4}', NULL, '2026-08-08 00:01:22.18972+08');

-- ----------------------------
-- Table structure for automation_rules
-- ----------------------------
DROP TABLE IF EXISTS "public"."automation_rules";
CREATE TABLE "public"."automation_rules" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8,
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "dsl" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "trigger_type" text COLLATE "pg_catalog"."default" NOT NULL,
  "action_count" int4 NOT NULL DEFAULT 0,
  "status" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'draft'::text,
  "created_by" int8 NOT NULL,
  "last_run_at" timestamptz(6),
  "last_error" text COLLATE "pg_catalog"."default",
  "consecutive_failures" int4 NOT NULL DEFAULT 0,
  "execution_count" int8 NOT NULL DEFAULT 0,
  "sort_order" int4 NOT NULL DEFAULT 0,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.automation_rules IS '自动化规则表（JSON DSL: trigger/conditions/actions；支持试运行 dry-run）';
COMMENT ON COLUMN public.automation_rules.id IS '主键 ID';
COMMENT ON COLUMN public.automation_rules.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.automation_rules.project_id IS '项目 FK（NULL=工作空间级通用规则）';
COMMENT ON COLUMN public.automation_rules.name IS '规则名称（如"缺陷修复自动指派验证人"）';
COMMENT ON COLUMN public.automation_rules.enabled IS '启用开关: true=生效, false=暂停（连续失败 3 次自动禁用）';
COMMENT ON COLUMN public.automation_rules.trigger IS '触发器 JSONB {type: issue.status_changed, filter:{type_code, to_group}}';
COMMENT ON COLUMN public.automation_rules.conditions IS '条件矩阵 JSONB {all/any: [{field, op, value}]}；纯函数无 IO';
COMMENT ON COLUMN public.automation_rules.actions IS '动作列表 JSONB [{type: assign/transition/notify/create_issue, ...}]；最多 10 个顺序执行';
COMMENT ON COLUMN public.automation_rules.last_executed_at IS '最后执行时间（监控规则活跃度）';
COMMENT ON COLUMN public.automation_rules.failure_count IS '连续失败计数；>=3 触发熔断 + 通知 admin';
COMMENT ON COLUMN public.automation_rules.execution_count IS '累计执行成功次数（效能统计）';
COMMENT ON COLUMN public.automation_rules.created_by IS '创建人 FK（automation.failed 通知的默认接收人 + 项目 admin）';
COMMENT ON COLUMN public.automation_rules.created_at IS '创建时间';
COMMENT ON COLUMN public.automation_rules.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.automation_rules.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of automation_rules
-- ----------------------------

-- ----------------------------
-- Table structure for automation_templates
-- ----------------------------
DROP TABLE IF EXISTS "public"."automation_templates";
CREATE TABLE "public"."automation_templates" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "slug" text COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "category" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'efficiency'::text,
  "dsl_template" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "icon" text COLLATE "pg_catalog"."default",
  "sort_order" int4 NOT NULL DEFAULT 0,
  "is_recommended" bool NOT NULL DEFAULT false,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.automation_templates IS '自动化内置模板表（15 条开箱即用模板；创建项目时可批量复制）';
COMMENT ON COLUMN public.automation_templates.id IS '主键 ID';
COMMENT ON COLUMN public.automation_templates.name IS '内置模板名称（中文/英文双语文档链接）';
COMMENT ON COLUMN public.automation_templates.description IS '模板功能说明（如"子项全完成 → 父项自动完成"）';
COMMENT ON COLUMN public.automation_templates.category IS '模板分类: quality/issue_management/sprint/version/intake/assignment';
COMMENT ON COLUMN public.automation_templates.dsl_template IS '模板 rule JSON（与 automation_rules 同结构；创建项目时批量复制）';
COMMENT ON COLUMN public.automation_templates.sort_order IS '排序权重（预设模板安装时按顺序展示）';
COMMENT ON COLUMN public.automation_templates.is_active IS '模板是否启用；false=不在可用列表';
COMMENT ON COLUMN public.automation_templates.created_at IS '创建时间';
COMMENT ON COLUMN public.automation_templates.updated_at IS '修改时间';

-- ----------------------------
-- Records of automation_templates
-- ----------------------------
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (1, '子项全部完成后自动完成父项', 'auto-complete-parent', '当所有子工作项都标记为完成时，自动将父工作项状态更新为已完成', 'quality', '{"actions": [{"type": "transition", "field": "state", "value": "completed"}], "trigger": {"type": "issue.status_changed", "filter": {"to_group": "completed"}}, "conditions": {"all": [{"op": "is_not_empty", "field": "parent"}]}}', 'git-merge', 1, 't', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (2, '逾期提醒', 'overdue-reminder', '工作项到期前 1 天自动提醒负责人', 'notification', '{"actions": [{"type": "notify", "target": "${issue.assignees}", "channel": "in_app", "template": "工作项 {{issue.identifier}} 即将到期"}], "trigger": {"cron": "0 9 * * *", "type": "scheduled", "filter": {"due_within_hours": 24}}, "conditions": {"all": [{"op": "ne", "field": "state.group", "value": "completed"}]}}', 'clock', 2, 't', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (3, '版本发布后自动流转工作项', 'version-release-transition', '版本发布时，自动将该版本下的工作项状态更新为已完成', 'efficiency', '{"actions": [{"type": "transition", "field": "state", "value": "completed"}], "trigger": {"type": "version.released"}, "conditions": {"all": [{"op": "eq", "field": "issue.fix_version", "value": "${version.id}"}]}}', 'rocket', 3, 'f', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (4, 'Epic 点数自动汇总', 'epic-points-rollup', '当子工作项点数变更时，自动汇总到 Epic 的聚合点数字段', 'efficiency', '{"actions": [{"type": "copy_field", "source": "${issue.sum_children_points}", "target": "${parent.estimate_points"}], "trigger": {"type": "issue.updated", "filter": {"field_changes": ["estimate_points"]}}, "conditions": {"all": [{"op": "ne", "field": "issue.type_code", "value": "epic"}]}}', 'layers', 4, 'f', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (5, '进入"进行中"时自动填写开始日期', 'auto-start-date', '工作项首次进入进行中状态时，自动记录开始时间', 'efficiency', '{"actions": [{"type": "update_field", "field": "started_at", "value": "${now}"}], "trigger": {"type": "issue.status_changed", "filter": {"to_group": "started"}}, "conditions": {"all": [{"op": "is_empty", "field": "started_at"}]}}', 'play', 5, 't', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (6, '最闲人自动指派', 'auto-assign-least-loaded', '新建工作项时自动分配给当前负载最轻的成员', 'efficiency', '{"actions": [{"role": "member", "type": "assign", "scope": "project", "strategy": "least_loaded"}], "trigger": {"type": "issue.created"}, "conditions": {"all": [{"op": "is_empty", "field": "assignees"}]}}', 'user-plus', 6, 'f', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (7, '新缺陷通知技术负责人', 'defect-notify-tech-lead', '项目里新建高优缺陷时，自动通知项目技术负责人', 'notification', '{"actions": [{"type": "notify", "target": "${project.tech_lead}", "channel": "in_app", "template": "🚨 新建紧急缺陷: [{{issue.identifier}}] {{issue.name}}"}], "trigger": {"type": "issue.created", "filter": {"priority": "urgent", "type_code": "defect"}}, "conditions": []}', 'alert-triangle', 7, 't', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (8, '缺陷修复后自动指派验证人', 'defect-assign-verifier', '缺陷修复后自动将验证任务指派给创建者', 'quality', '{"conditions": [], "trigger": {"type": "issue.status_changed", "filter": {"type_code": "defect", "to_group": "completed"}}, "actions": [{"type": "notify", "target": "${issue.created_by}", "channel": "in_app", "template": "缺陷 {{issue.identifier}} 已修复，请验证"}]}', 'check-circle', 8, 'f', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (9, '高优需求自动标记', 'auto-set-priority', '根据关键词自动设置工作项优先级', 'efficiency', '{"conditions": [{"op": "contains", "field": "issue.name", "value": "紧急"}], "trigger": {"type": "issue.created"}, "actions": [{"type": "update_field", "field": "priority", "value": "urgent"}]}', 'zap', 9, 'f', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (10, '状态变更通知关注人', 'status-change-notify-watchers', '工作项状态变更时通知所有关注人', 'notification', '{"conditions": [], "trigger": {"type": "issue.status_changed"}, "actions": [{"type": "notify", "target": "${issue.watchers}", "channel": "in_app", "template": "{{issue.identifier}} 状态变更为 {{issue.state_name}}"}]}', 'bell', 10, 'f', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (11, '迭代完成自动通知团队', 'sprint-complete-summary', '迭代完成时自动通知所有成员并发送总结', 'notification', '{"conditions": [], "trigger": {"type": "sprint.completed"}, "actions": [{"type": "notify", "target": "${project.members}", "channel": "in_app", "template": "迭代 {{sprint.name}} 已完成"}]}', 'flag', 11, 'f', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (12, '迭代启动后自动开始工作项', 'sprint-auto-start-issues', '迭代启动后，自动将所有待办工作项流转到进行中', 'management', '{"conditions": [{"op": "eq", "field": "state.group", "value": "todo"}], "trigger": {"type": "sprint.started"}, "actions": [{"type": "transition", "field": "state", "value": "started"}]}', 'play-circle', 12, 'f', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (13, '长期未更新工作项自动归档', 'auto-archive-old-issues', '超过 30 天未更新的已完成工作项自动归档', 'management', '{"conditions": [{"op": "eq", "field": "state.group", "value": "completed"}, {"op": "lt", "field": "issue.updated_at", "value": "now-30d"}], "trigger": {"type": "scheduled", "cron": "0 2 * * *"}, "actions": [{"type": "update_field", "field": "is_archived", "value": "true"}]}', 'archive', 13, 'f', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (14, '重复工作项提醒', 'duplicate-issue-check', '新建工作项时检测可能的重复项并提醒', 'management', '{"conditions": [], "trigger": {"type": "issue.created"}, "actions": [{"type": "notify", "target": "${issue.created_by}", "channel": "in_app", "template": "⚠️ 检测到可能的重复工作项请确认"}]}', 'copy', 14, 'f', '2026-08-08 00:01:05.002175+08');
INSERT INTO "public"."automation_templates" OVERRIDING SYSTEM VALUE VALUES (15, '新成员加入通知', 'new-member-welcome', '工作空间有新成员加入时通知所有成员', 'management', '{"conditions": [], "trigger": {"type": "member.added"}, "actions": [{"type": "notify", "target": "${workspace.members}", "channel": "in_app", "template": "欢迎 {{actor.user_name}} 加入工作空间"}]}', 'user-plus', 15, 'f', '2026-08-08 00:01:05.002175+08');

-- ----------------------------
-- Table structure for dashboard_snapshots
-- ----------------------------
DROP TABLE IF EXISTS "public"."dashboard_snapshots";
CREATE TABLE "public"."dashboard_snapshots" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "project_id" int8 NOT NULL,
  "widget_type" text COLLATE "pg_catalog"."default" NOT NULL,
  "data" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "refreshed_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.dashboard_snapshots IS '仪表盘快照表（定时导出/分享大屏只读链接数据）';
COMMENT ON COLUMN public.dashboard_snapshots.id IS '主键 ID';
COMMENT ON COLUMN public.dashboard_snapshots.dashboard_id IS '仪表盘 FK';
COMMENT ON COLUMN public.dashboard_snapshots.snapshot_time IS '快照时间（大屏分享链接/定时导出触发）';
COMMENT ON COLUMN public.dashboard_snapshots.data IS '完整快照 JSONB（所有卡片数据缓存；大屏只读展示）';
COMMENT ON COLUMN public.dashboard_snapshots.share_token IS '分享令牌 SHA-256（NULL=非分享；过期/吊销后清空）';
COMMENT ON COLUMN public.dashboard_snapshots.expires_at IS '分享有效期 TIMESTAMPTZ；到期后快照URL 404';
COMMENT ON COLUMN public.dashboard_snapshots.created_by IS '创建人 FK';
COMMENT ON COLUMN public.dashboard_snapshots.created_at IS '创建时间';
COMMENT ON COLUMN public.dashboard_snapshots.workspace_id IS '工作空间 FK（RLS 依据）';

-- ----------------------------
-- Records of dashboard_snapshots
-- ----------------------------

-- ----------------------------
-- Table structure for dashboard_templates
-- ----------------------------
DROP TABLE IF EXISTS "public"."dashboard_templates";
CREATE TABLE "public"."dashboard_templates" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "slug" text COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "layout" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "icon" text COLLATE "pg_catalog"."default",
  "category" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'agile'::text,
  "is_default" bool NOT NULL DEFAULT false,
  "sort_order" int4 NOT NULL DEFAULT 0,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.dashboard_templates IS '仪表盘预设模板表（敏捷项目/版本交付/PMO 多项目；布局+卡片组合 JSON）';
COMMENT ON COLUMN public.dashboard_templates.id IS '主键 ID';
COMMENT ON COLUMN public.dashboard_templates.name IS '模板名称（如"工程效能"/"QA 质量"/"PMO 战略多项目"）';
COMMENT ON COLUMN public.dashboard_templates.description IS '模板适用场景说明';
COMMENT ON COLUMN public.dashboard_templates.scope IS '作用域: project(单项目) / workspace(多项目聚合)';
COMMENT ON COLUMN public.dashboard_templates.layout IS '预设布局 JSONB（卡片数组 [{type, x, y, w, h, config}]；gridstack.js 格式）';
COMMENT ON COLUMN public.dashboard_templates.is_system IS '是否内置模板: true=系统预设, false=用户自定义（可分享）';
COMMENT ON COLUMN public.dashboard_templates.created_by IS '创建人 FK（is_system=true 时为 NULL）';
COMMENT ON COLUMN public.dashboard_templates.created_at IS '创建时间';
COMMENT ON COLUMN public.dashboard_templates.updated_at IS '修改时间';

-- ----------------------------
-- Records of dashboard_templates
-- ----------------------------
INSERT INTO "public"."dashboard_templates" OVERRIDING SYSTEM VALUE VALUES (1, '项目概览', 'project-overview', '项目级核心指标：进度、趋势、风险点', '{"widgets": [{"h": 2, "w": 12, "x": 0, "y": 0, "type": "progress_overview"}, {"h": 4, "w": 6, "x": 0, "y": 2, "type": "burndown"}, {"h": 4, "w": 3, "x": 6, "y": 2, "type": "risk_alert"}, {"h": 4, "w": 3, "x": 9, "y": 2, "type": "overdue_list"}, {"h": 4, "w": 6, "x": 0, "y": 6, "type": "recent_activity"}, {"h": 4, "w": 6, "x": 6, "y": 6, "type": "team_workload"}]}', 'chart', 'agile', 't', 1, '2026-08-08 00:01:04.927962+08');
INSERT INTO "public"."dashboard_templates" OVERRIDING SYSTEM VALUE VALUES (2, '项目管理', 'pmo-dashboard', 'PMO 视角：多维度统计 + 风险清单', '{"widgets": [{"h": 2, "w": 12, "x": 0, "y": 0, "type": "progress_overview"}, {"h": 4, "w": 4, "x": 0, "y": 2, "type": "state_distribution"}, {"h": 4, "w": 4, "x": 4, "y": 2, "type": "priority_split"}, {"h": 4, "w": 4, "x": 8, "y": 2, "type": "velocity"}, {"h": 3, "w": 6, "x": 0, "y": 6, "type": "risk_alert"}, {"h": 3, "w": 6, "x": 6, "y": 6, "type": "overdue_list"}]}', 'monitor', 'pmo', 'f', 2, '2026-08-08 00:01:04.927962+08');
INSERT INTO "public"."dashboard_templates" OVERRIDING SYSTEM VALUE VALUES (3, '质量看板', 'quality-dashboard', '质量指标：缺陷趋势 + 阻塞分析', '{"widgets": [{"h": 3, "w": 6, "x": 0, "y": 0, "type": "priority_split"}, {"h": 3, "w": 6, "x": 6, "y": 0, "type": "blocked_list"}, {"h": 4, "w": 12, "x": 0, "y": 3, "type": "burndown"}, {"h": 4, "w": 12, "x": 0, "y": 7, "type": "recent_activity"}]}', 'bug', 'quality', 'f', 3, '2026-08-08 00:01:04.927962+08');

-- ----------------------------
-- Table structure for dashboard_widgets
-- ----------------------------
DROP TABLE IF EXISTS "public"."dashboard_widgets";
CREATE TABLE "public"."dashboard_widgets" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "project_id" int8 NOT NULL,
  "widget_type" text COLLATE "pg_catalog"."default" NOT NULL,
  "title" text COLLATE "pg_catalog"."default" NOT NULL,
  "grid_x" int4 NOT NULL DEFAULT 0,
  "grid_y" int4 NOT NULL DEFAULT 0,
  "grid_w" int4 NOT NULL DEFAULT 4,
  "grid_h" int4 NOT NULL DEFAULT 3,
  "config" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "is_visible" bool NOT NULL DEFAULT true,
  "sort_order" int4 NOT NULL DEFAULT 0,
  "user_id" int8,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.dashboard_widgets IS '仪表盘卡片实例表（type→渲染器/数据接口映射；config JSONB 个性化配置）';
COMMENT ON COLUMN public.dashboard_widgets.id IS '主键 ID';
COMMENT ON COLUMN public.dashboard_widgets.dashboard_id IS '仪表盘 FK（dashboard_configs.id；可选，可存个人临时配置）';
COMMENT ON COLUMN public.dashboard_widgets.type IS '卡片类型: project_overview / sprint_burndown / version_progress / module_distribution / quality_indicators / risk_alerts / resource_load / velocity_trend / dora_summary';
COMMENT ON COLUMN public.dashboard_widgets.config IS '卡片个性化配置 JSONB（时间范围/项目/迭代/对比维度）';
COMMENT ON COLUMN public.dashboard_widgets.position_x IS '布局横坐标（grid 列；0 起始）';
COMMENT ON COLUMN public.dashboard_widgets.position_y IS '布局纵坐标（grid 行；0 起始）';
COMMENT ON COLUMN public.dashboard_widgets.width IS '卡片宽度（grid 单元数 1-4）';
COMMENT ON COLUMN public.dashboard_widgets.height IS '卡片高度（grid 单元数 1-4）';
COMMENT ON COLUMN public.dashboard_widgets.refresh_interval_s IS '刷新间隔秒数（实时 30s；缓存 dash:{dashboard}:widget:{id} TTL）';
COMMENT ON COLUMN public.dashboard_widgets.data_source IS '数据源标识符（前端渲染器 ↔ 后端接口 1:1 映射）';
COMMENT ON COLUMN public.dashboard_widgets.workspace_id IS '工作空间 FK（RLS 依据）';
COMMENT ON COLUMN public.dashboard_widgets.created_at IS '创建时间';
COMMENT ON COLUMN public.dashboard_widgets.updated_at IS '修改时间';

-- ----------------------------
-- Records of dashboard_widgets
-- ----------------------------

-- ----------------------------
-- Table structure for deployment_events
-- ----------------------------
DROP TABLE IF EXISTS "public"."deployment_events";
CREATE TABLE "public"."deployment_events" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8,
  "deployment_id" text COLLATE "pg_catalog"."default",
  "env" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'production'::text,
  "status" text COLLATE "pg_catalog"."default" NOT NULL,
  "commit_sha" text COLLATE "pg_catalog"."default",
  "started_at" timestamptz(6),
  "deployed_at" timestamptz(6),
  "source" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'webhook'::text,
  "metadata" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.deployment_events IS '部署事件表（DORA 数据源；CI/CD Webhook 推送；HMAC 验签；env+status+commit_sha）';
COMMENT ON COLUMN public.deployment_events.id IS '主键 ID';
COMMENT ON COLUMN public.deployment_events.workspace_id IS '工作空间 FK（RLS 依据）';
COMMENT ON COLUMN public.deployment_events.project_id IS '项目 FK';
COMMENT ON COLUMN public.deployment_events.environment IS '部署环境: dev / staging / production';
COMMENT ON COLUMN public.deployment_events.status IS '部署状态: success / failed';
COMMENT ON COLUMN public.deployment_events.commit_sha IS '部署 commit SHA（精确到 commit；用于关联工作项/PR）';
COMMENT ON COLUMN public.deployment_events.started_at IS '部署开始时间 TIMESTAMPTZ；started_at→deployed_at = lead time';
COMMENT ON COLUMN public.deployment_events.deployed_at IS '部署成功时间 TIMESTAMPTZ；DORA 计算数据源';
COMMENT ON COLUMN public.deployment_events.source IS '触发来源: ci_cd / webhook / manual；标识 CI 流水线';
COMMENT ON COLUMN public.deployment_events.meta IS '元数据 JSONB（workflow_id / branch / tags / runner 等）';
COMMENT ON COLUMN public.deployment_events.created_by IS '操作人 FK';
COMMENT ON COLUMN public.deployment_events.created_at IS '注册时间（Webhook POST /hooks/deployments 入口）';

-- ----------------------------
-- Records of deployment_events
-- ----------------------------

-- ----------------------------
-- Table structure for domain_events
-- ----------------------------
DROP TABLE IF EXISTS "public"."domain_events";
CREATE TABLE "public"."domain_events" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "aggregate_type" text COLLATE "pg_catalog"."default" NOT NULL,
  "aggregate_id" int8 NOT NULL,
  "event_type" text COLLATE "pg_catalog"."default" NOT NULL,
  "payload" jsonb NOT NULL,
  "occurred_at" timestamptz(6) NOT NULL DEFAULT now(),
  "published_at" timestamptz(6)
)
;
COMMENT ON TABLE public.domain_events IS '领域事件表（Transactional Outbox 模式；未发布事件 id 排序投递 + 7 天清理）';
COMMENT ON COLUMN public.domain_events.id IS '主键 ID';
COMMENT ON COLUMN public.domain_events.workspace_id IS '工作空间 FK（RLS 依据）';
COMMENT ON COLUMN public.domain_events.aggregate_type IS '聚合类型: issue / sprint / version / automation / deployment';
COMMENT ON COLUMN public.domain_events.aggregate_id IS '聚合 ID（BIGINT 统一）';
COMMENT ON COLUMN public.domain_events.event_type IS '事件类型: issue.status_changed；sprint.started；automation.executed ...';
COMMENT ON COLUMN public.domain_events.payload IS '事件载荷 JSONB（当时完整状态快照；消费者解析依据）';
COMMENT ON COLUMN public.domain_events.occurred_at IS '事件发生时间（与事务提交时间对齐；Outbox 写入时间）';
COMMENT ON COLUMN public.domain_events.published_at IS '投递时间（worker 处理后置；NULL=待投递；WHERE published_at IS NULL 发布）';
COMMENT ON COLUMN public.domain_events.created_at IS '写入时间（事务内与业务表同事务；唯一事实源）';

-- ----------------------------
-- Records of domain_events
-- ----------------------------

-- ----------------------------
-- Table structure for idempotency_keys
-- ----------------------------
DROP TABLE IF EXISTS "public"."idempotency_keys";
CREATE TABLE "public"."idempotency_keys" (
  "key" text COLLATE "pg_catalog"."default" NOT NULL,
  "user_id" int8 NOT NULL,
  "response" jsonb,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.idempotency_keys IS 'API 幂等键表（写操作去重窗口；复用 response 缓存）';
COMMENT ON COLUMN public.idempotency_keys.id IS '主键 ID';
COMMENT ON COLUMN public.idempotency_keys.key IS '幂等键（客户端生成 UUID；API 请求头 X-Idempotency-Key；unique）';
COMMENT ON COLUMN public.idempotency_keys.user_id IS '用户 FK（校验请求者身份）';
COMMENT ON COLUMN public.idempotency_keys.response IS '原响应缓存 JSONB（同 key 重复请求直接返回原 response；省 DB 调用）';
COMMENT ON COLUMN public.idempotency_keys.request_hash IS '请求体 SHA-256 摘要（可选；防 request body 修改后复用旧响应）';
COMMENT ON COLUMN public.idempotency_keys.created_at IS '首次请求时间（过期清理窗口 ≥24h）';
COMMENT ON COLUMN public.idempotency_keys.expires_at IS '过期时间 TIMESTAMPTZ；过期后同 key 可重放';

-- ----------------------------
-- Records of idempotency_keys
-- ----------------------------

-- ----------------------------
-- Table structure for intake_channels
-- ----------------------------
DROP TABLE IF EXISTS "public"."intake_channels";
CREATE TABLE "public"."intake_channels" (
  "id" int8 NOT NULL DEFAULT nextval('intake_channels_id_seq'::regclass),
  "workspace_id" int8 NOT NULL,
  "project_id" int8,
  "slug" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "name" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "is_public" bool NOT NULL DEFAULT true,
  "default_issue_type" varchar(20) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'requirement'::character varying,
  "default_priority" int2 NOT NULL DEFAULT 0,
  "auto_assign_rules" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "rate_limit_per_min" int2 NOT NULL DEFAULT 20,
  "require_captcha" bool NOT NULL DEFAULT true,
  "custom_fields" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "branding" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "notify_on_submit" bool NOT NULL DEFAULT true,
  "notify_users" int8[] NOT NULL DEFAULT '{}'::bigint[],
  "is_active" bool NOT NULL DEFAULT true,
  "created_by" int8 NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.intake_channels IS '收件箱入口通道表（公开门户 slug；限流 + 行为验证码 + 自动分配规则）';
COMMENT ON COLUMN public.intake_channels.id IS '主键 ID';
COMMENT ON COLUMN public.intake_channels.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.intake_channels.project_id IS '项目 FK；关联默认工作项类型与指派规则';
COMMENT ON COLUMN public.intake_channels.name IS '通道名称（客户门户标题）';
COMMENT ON COLUMN public.intake_channels.slug IS '公开 URL 路径段 slug（/intake/{slug}；限流 20/min/IP）';
COMMENT ON COLUMN public.intake_channels.is_public IS '门户公开开关: true=免登录可提交, false=需登录验证';
COMMENT ON COLUMN public.intake_channels.default_issue_type IS '默认转正类型: requirement / defect（管理员可配置）';
COMMENT ON COLUMN public.intake_channels.auto_assign_rules IS '自动分配规则 JSONB [{match:{keyword,tags}, assign_to:user_id}]；顺序匹配';
COMMENT ON COLUMN public.intake_channels.created_by IS '创建人 FK';
COMMENT ON COLUMN public.intake_channels.created_at IS '创建时间';
COMMENT ON COLUMN public.intake_channels.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.intake_channels.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of intake_channels
-- ----------------------------

-- ----------------------------
-- Table structure for intake_issues
-- ----------------------------
DROP TABLE IF EXISTS "public"."intake_issues";
CREATE TABLE "public"."intake_issues" (
  "id" int8 NOT NULL DEFAULT nextval('intake_issues_id_seq'::regclass),
  "channel_id" int8 NOT NULL,
  "workspace_id" int8 NOT NULL,
  "project_id" int8,
  "tracking_id" varchar(32) COLLATE "pg_catalog"."default" NOT NULL,
  "submitter_name" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "submitter_email" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "submitter_user_id" int8,
  "title" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT ''::text,
  "issue_type" varchar(20) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'requirement'::character varying,
  "priority" int2 NOT NULL DEFAULT 0,
  "custom_fields" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "attachment_ids" int8[] NOT NULL DEFAULT '{}'::bigint[],
  "status" varchar(20) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'open'::character varying,
  "status_reason" text COLLATE "pg_catalog"."default",
  "converted_issue_id" int8,
  "assigned_to" int8,
  "reviewed_by" int8,
  "reviewed_at" timestamptz(6),
  "notify_on_status" bool NOT NULL DEFAULT true,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.intake_issues IS '收件箱提交记录表（tracking_id YD-IN-XXXX 提交回执；status open/accepted/rejected/archived）';
COMMENT ON COLUMN public.intake_issues.id IS '主键 ID';
COMMENT ON COLUMN public.intake_issues.channel_id IS '收件箱通道 FK';
COMMENT ON COLUMN public.intake_issues.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.intake_issues.tracking_id IS '提交回执编号（YD-IN-XXXX；提交后邮件通知；/track/{id} 跟踪）';
COMMENT ON COLUMN public.intake_issues.status IS '处理状态: open(待审核) / accepted(已转正) / rejected(已拒绝) / archived(暂存)';
COMMENT ON COLUMN public.intake_issues.submitter_name IS '提交者姓名（脱敏显示）';
COMMENT ON COLUMN public.intake_issues.submitter_email IS '提交者邮箱（邮件校验码找回进展）';
COMMENT ON COLUMN public.intake_issues.subject IS '提交标题';
COMMENT ON COLUMN public.intake_issues.description IS '提交描述（纯文本；不支持富文本/附件）';
COMMENT ON COLUMN public.intake_issues.converted_issue_id IS '转正后工作项 FK（issues.id；创建时复制标题/描述/附件）';
COMMENT ON COLUMN public.intake_issues.converted_at IS '转正时间（管理员审核通过时）';
COMMENT ON COLUMN public.intake_issues.rejection_reason IS '拒绝原因（填写后通知提交者）';
COMMENT ON COLUMN public.intake_issues.created_by IS '提交人 FK（可 NULL；匿名提交）';
COMMENT ON COLUMN public.intake_issues.created_at IS '提交时间';
COMMENT ON COLUMN public.intake_issues.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.intake_issues.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of intake_issues
-- ----------------------------

-- ----------------------------
-- Table structure for invitations
-- ----------------------------
DROP TABLE IF EXISTS "public"."invitations";
CREATE TABLE "public"."invitations" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "inviter_id" int8 NOT NULL,
  "email" text COLLATE "pg_catalog"."default" NOT NULL,
  "role" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'member'::text,
  "token_hash" text COLLATE "pg_catalog"."default" NOT NULL,
  "message" text COLLATE "pg_catalog"."default",
  "status" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'pending'::text,
  "expires_at" timestamptz(6) NOT NULL DEFAULT (now() + '7 days'::interval),
  "accepted_at" timestamptz(6),
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.invitations IS '工作空间邀请表（邮件邀请码 + 角色 + 过期 + 幂等）';
COMMENT ON COLUMN public.invitations.id IS '主键 ID';
COMMENT ON COLUMN public.invitations.workspace_id IS '目标工作空间 FK';
COMMENT ON COLUMN public.invitations.email IS '被邀請人邮箱（小写；workspace+email 唯一排除软删除）';
COMMENT ON COLUMN public.invitations.role IS '邀请角色: owner / admin / member / guest';
COMMENT ON COLUMN public.invitations.status IS '邀请状态: pending(待确认) / accepted / rejected / expired';
COMMENT ON COLUMN public.invitations.token IS '邀请校验令牌（UUID；SHA-256 hash 存储；邮件内链接使用）';
COMMENT ON COLUMN public.invitations.expires_at IS '邀请有效期 TIMESTAMPTZ；过期自动标记 status=expired';
COMMENT ON COLUMN public.invitations.invited_by IS '邀请人 FK';
COMMENT ON COLUMN public.invitations.accepted_at IS '接受时间（跳转注册/登录后）';
COMMENT ON COLUMN public.invitations.created_at IS '创建时间';
COMMENT ON COLUMN public.invitations.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of invitations
-- ----------------------------

-- ----------------------------
-- Table structure for issue_activities
-- ----------------------------
DROP TABLE IF EXISTS "public"."issue_activities";
CREATE TABLE "public"."issue_activities" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "issue_id" int8 NOT NULL,
  "verb" text COLLATE "pg_catalog"."default" NOT NULL,
  "field" text COLLATE "pg_catalog"."default",
  "old_value" text COLLATE "pg_catalog"."default",
  "new_value" text COLLATE "pg_catalog"."default",
  "old_ref" jsonb,
  "new_ref" jsonb,
  "actor_id" int8,
  "actor_email" text COLLATE "pg_catalog"."default",
  "actor_name" text COLLATE "pg_catalog"."default",
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.issue_activities IS '工作项活动历史表（按月 RANGE 分区；字段 diff / 流转 / 附件 / 关联等全量审计）';
COMMENT ON COLUMN public.issue_activities.id IS '主键 ID';
COMMENT ON COLUMN public.issue_activities.issue_id IS '关联工作项 FK';
COMMENT ON COLUMN public.issue_activities.actor_id IS '操作人 FK（users.id）；系统动作为 NULL';
COMMENT ON COLUMN public.issue_activities.verb IS '操作动词: created/updated/transitioned/attached/linked/detached/deleted/restored/commented/mentioned';
COMMENT ON COLUMN public.issue_activities.field IS '变更字段名（verb=updated 时使用；如 state, priority, assignees）';
COMMENT ON COLUMN public.issue_activities.old_value IS '变更前值（TEXT 或 JSONB 序列化）';
COMMENT ON COLUMN public.issue_activities.new_value IS '变更后值（TEXT 或 JSONB 序列化）';
COMMENT ON COLUMN public.issue_activities.metadata IS '附加信息 JSONB（如流转 from→to、字段 diff 详情）';
COMMENT ON COLUMN public.issue_activities.created_at IS '记录时间 TIMESTAMPTZ';
COMMENT ON COLUMN public.issue_activities.workspace_id IS '工作空间 FK（支撑 RLS 与按月分区键）';

-- ----------------------------
-- Records of issue_activities
-- ----------------------------

-- ----------------------------
-- Table structure for issue_assignees
-- ----------------------------
DROP TABLE IF EXISTS "public"."issue_assignees";
CREATE TABLE "public"."issue_assignees" (
  "issue_id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "assigned_at" timestamptz(6) NOT NULL DEFAULT now(),
  "assigned_by" int8
)
;
COMMENT ON TABLE public.issue_assignees IS '工作项-指派人物多对一关联表（含主负责人标记）';
COMMENT ON COLUMN public.issue_assignees.issue_id IS '工作项 FK';
COMMENT ON COLUMN public.issue_assignees.user_id IS '被指派人 FK';
COMMENT ON COLUMN public.issue_assignees.is_primary IS '是否主负责人: true=主, false=辅助；每个工作项仅一个主负责人';
COMMENT ON COLUMN public.issue_assignees.created_at IS '指派时间';

-- ----------------------------
-- Records of issue_assignees
-- ----------------------------

-- ----------------------------
-- Table structure for issue_comments
-- ----------------------------
DROP TABLE IF EXISTS "public"."issue_comments";
CREATE TABLE "public"."issue_comments" (
  "id" int8 NOT NULL DEFAULT nextval('issue_comments_id_seq'::regclass),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "issue_id" int8 NOT NULL,
  "content_json" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "content_html" text COLLATE "pg_catalog"."default",
  "content_stripped" text COLLATE "pg_catalog"."default",
  "created_by" int8 NOT NULL,
  "mentions" int8[] DEFAULT '{}'::bigint[],
  "parent_id" int8,
  "is_edited" bool NOT NULL DEFAULT false,
  "edited_at" timestamptz(6),
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.issue_comments IS '工作项评论表（TipTap JSON 富文本，支持 @提及、回复、反应）';
COMMENT ON COLUMN "public"."issue_comments"."content_json" IS 'TipTap 编辑器的 JSON 输出';
COMMENT ON COLUMN "public"."issue_comments"."mentions" IS '@提及的用户 ID 数组';
COMMENT ON COLUMN "public"."issue_comments"."parent_id" IS '父评论 ID（嵌套回复）';
COMMENT ON TABLE "public"."issue_comments" IS '工作项评论表（支持富文本 + @提及 + 嵌套回复）';

-- ----------------------------
-- Records of issue_comments
-- ----------------------------

-- ----------------------------
-- Table structure for issue_dependencies
-- ----------------------------
DROP TABLE IF EXISTS "public"."issue_dependencies";
CREATE TABLE "public"."issue_dependencies" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "predecessor_id" int8 NOT NULL,
  "successor_id" int8 NOT NULL,
  "dependency_type" text COLLATE "pg_catalog"."default" NOT NULL,
  "lag_days" int4 NOT NULL DEFAULT 0,
  "created_by" int8 NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.issue_dependencies IS '工作项依赖关系表（FS/SS/FF/SF + lag_days，DFS 检测环）';
COMMENT ON COLUMN public.issue_dependencies.id IS '主键 ID';
COMMENT ON COLUMN public.issue_dependencies.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.issue_dependencies.issue_id IS '前置工作项 FK（dependencies 的"from"）';
COMMENT ON COLUMN public.issue_dependencies.dependent_id IS '后续工作项 FK（dependencies 的"to"）';
COMMENT ON COLUMN public.issue_dependencies.dependency_type IS '依赖类型: FS(完成→开始) / SS(开始→开始) / FF(完成→完成) / SF(开始→完成)';
COMMENT ON COLUMN public.issue_dependencies.lag_days IS '延迟天数（正=延迟等待，负=提前开始）；0 表示无延迟';
COMMENT ON COLUMN public.issue_dependencies.created_by IS '创建人 FK';
COMMENT ON COLUMN public.issue_dependencies.created_at IS '创建时间';
COMMENT ON COLUMN public.issue_dependencies.deleted_at IS '软删除时间戳（唯一约束 WHERE deleted_at IS NULL）';

-- ----------------------------
-- Records of issue_dependencies
-- ----------------------------

-- ----------------------------
-- Table structure for issue_labels
-- ----------------------------
DROP TABLE IF EXISTS "public"."issue_labels";
CREATE TABLE "public"."issue_labels" (
  "issue_id" int8 NOT NULL,
  "label_id" int8 NOT NULL
)
;
COMMENT ON TABLE public.issue_labels IS '工作项-标签多对一关联表';
COMMENT ON COLUMN public.issue_labels.issue_id IS '工作项 FK';
COMMENT ON COLUMN public.issue_labels.label_id IS '标签 FK';
COMMENT ON COLUMN public.issue_labels.created_at IS '关联创建时间';

-- ----------------------------
-- Records of issue_labels
-- ----------------------------

-- ----------------------------
-- Table structure for issue_modules
-- ----------------------------
DROP TABLE IF EXISTS "public"."issue_modules";
CREATE TABLE "public"."issue_modules" (
  "issue_id" int8 NOT NULL,
  "module_id" int8 NOT NULL
)
;
COMMENT ON TABLE public.issue_modules IS '工作项-模块多对一关联表';
COMMENT ON COLUMN public.issue_modules.issue_id IS '工作项 FK';
COMMENT ON COLUMN public.issue_modules.module_id IS '模块 FK';
COMMENT ON COLUMN public.issue_modules.created_at IS '关联创建时间';

-- ----------------------------
-- Records of issue_modules
-- ----------------------------

-- ----------------------------
-- Table structure for issue_relations
-- ----------------------------
DROP TABLE IF EXISTS "public"."issue_relations";
CREATE TABLE "public"."issue_relations" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "source_issue_id" int8 NOT NULL,
  "target_issue_id" int8 NOT NULL,
  "relation_type" text COLLATE "pg_catalog"."default" NOT NULL,
  "created_by" int8 NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.issue_relations IS '工作项语义关联表（duplicate/relates_to/blocked_by/start_before/finish_before/implemented_by）';
COMMENT ON COLUMN public.issue_relations.id IS '主键 ID';
COMMENT ON COLUMN public.issue_relations.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.issue_relations.issue_a_id IS '工作项 A FK';
COMMENT ON COLUMN public.issue_relations.issue_b_id IS '工作项 B FK';
COMMENT ON COLUMN public.issue_relations.relation_type IS '语义关联类型: duplicate(重复) / relates_to(关联) / blocked_by(被阻塞) / start_before(先于开始) / finish_before(先于完成) / implemented_by(由…实现)';
COMMENT ON COLUMN public.issue_relations.created_by IS '创建人 FK';
COMMENT ON COLUMN public.issue_relations.created_at IS '创建时间';

-- ----------------------------
-- Records of issue_relations
-- ----------------------------

-- ----------------------------
-- Table structure for issue_watchers
-- ----------------------------
DROP TABLE IF EXISTS "public"."issue_watchers";
CREATE TABLE "public"."issue_watchers" (
  "issue_id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.issue_watchers IS '工作项-关注人多对一关联表';
COMMENT ON COLUMN public.issue_watchers.issue_id IS '工作项 FK';
COMMENT ON COLUMN public.issue_watchers.user_id IS '关注人 FK；事件驱动通知（issue.updated/issue.commented 等）';
COMMENT ON COLUMN public.issue_watchers.created_at IS '关注时间';

-- ----------------------------
-- Records of issue_watchers
-- ----------------------------

-- ----------------------------
-- Table structure for issue_reactions
-- ----------------------------
DROP TABLE IF EXISTS "public"."issue_reactions";
CREATE TABLE "public"."issue_reactions" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "issue_id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "reaction_type" text COLLATE "pg_catalog"."default" NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.issue_reactions IS '工作项表情反应表（emoji 轻量反馈，参考 Linear/Plane Reaction）';
COMMENT ON COLUMN public.issue_reactions.id IS '主键 ID';
COMMENT ON COLUMN public.issue_reactions.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.issue_reactions.project_id IS '项目 FK';
COMMENT ON COLUMN public.issue_reactions.issue_id IS '工作项 FK';
COMMENT ON COLUMN public.issue_reactions.user_id IS '反应用户 FK';
COMMENT ON COLUMN public.issue_reactions.reaction_type IS '表情类型（emoji 字符串，如 👍 👀 🎉 ❤️ 😄）';
COMMENT ON COLUMN public.issue_reactions.created_at IS '反应时间（唯一约束: 同人同工作项同表情仅一条）';

-- ----------------------------
-- Records of issue_reactions
-- ----------------------------

-- ----------------------------
-- Table structure for issue_votes
-- ----------------------------
DROP TABLE IF EXISTS "public"."issue_votes";
CREATE TABLE "public"."issue_votes" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "issue_id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "vote" int2 NOT NULL DEFAULT 1,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.issue_votes IS '工作项投票表（支持赞成/反对，同人同工作项仅一票）';
COMMENT ON COLUMN public.issue_votes.id IS '主键 ID';
COMMENT ON COLUMN public.issue_votes.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.issue_votes.project_id IS '项目 FK';
COMMENT ON COLUMN public.issue_votes.issue_id IS '工作项 FK';
COMMENT ON COLUMN public.issue_votes.user_id IS '投票用户 FK';
COMMENT ON COLUMN public.issue_votes.vote IS '投票值: 1=赞成(upvote) / -1=反对(downvote)；0 表示撤销';
COMMENT ON COLUMN public.issue_votes.created_at IS '首次投票时间';
COMMENT ON COLUMN public.issue_votes.updated_at IS '最近更新（改票时刷新）';

-- ----------------------------
-- Records of issue_votes
-- ----------------------------

-- ----------------------------
-- Table structure for issues
-- ----------------------------
DROP TABLE IF EXISTS "public"."issues";
CREATE TABLE "public"."issues" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "public_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "sequence_id" int8 NOT NULL,
  "type_code" text COLLATE "pg_catalog"."default" NOT NULL,
  "parent_id" int8,
  "depth" int2 NOT NULL DEFAULT 1,
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "description_json" jsonb,
  "description_html" text COLLATE "pg_catalog"."default",
  "description_stripped" text COLLATE "pg_catalog"."default",
  "state_id" int8 NOT NULL,
  "priority" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'none'::text,
  "severity" int2,
  "found_phase" text COLLATE "pg_catalog"."default",
  "root_cause_category" text COLLATE "pg_catalog"."default",
  "verifier_id" int8,
  "environment" jsonb,
  "reproduce_steps" jsonb,
  "category" text COLLATE "pg_catalog"."default",
  "actual_effort" numeric(8,2),
  "remaining_effort" numeric(8,2),
  "delay_reason" text COLLATE "pg_catalog"."default",
  "source" text COLLATE "pg_catalog"."default",
  "point" int2,
  "sprint_id" int8,
  "progress" int2 NOT NULL DEFAULT 0,
  "start_date" date,
  "target_date" date,
  "completed_at" timestamptz(6),
  "is_draft" bool NOT NULL DEFAULT false,
  "sort_order" float8 NOT NULL DEFAULT 65535,
  "created_by" int8 NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now(),
  "deleted_at" timestamptz(6),
  "version" int4 NOT NULL DEFAULT 1,
  "found_version_id" int8,
  "fix_version_id" int8,
  "release_version_id" int8,
  "search_tsv" tsvector GENERATED ALWAYS AS (
((setweight(to_tsvector('simple'::regconfig, COALESCE(name, ''::text)), 'A'::"char") || setweight(to_tsvector('simple'::regconfig, COALESCE(description_stripped, ''::text)), 'B'::"char")) || setweight(to_tsvector('simple'::regconfig, COALESCE(type_code, ''::text)), 'C'::"char"))
) STORED
)
;
COMMENT ON TABLE public.issues IS '工作项主表（需求/任务/缺陷统一存储，支撑看板、迭代、搜索、关联等核心能力）';
COMMENT ON COLUMN public.issues.id IS '主键 ID（GENERATED ALWAYS AS IDENTITY，不外泄）';
COMMENT ON COLUMN public.issues.public_id IS '对外暴露的唯一标识（UUID），用于 API 与外部引用';
COMMENT ON COLUMN public.issues.workspace_id IS '工作空间/租户 FK，RLS 依据，复合索引首列';
COMMENT ON COLUMN public.issues.project_id IS '所属项目 FK（projects.id），聚合根容器';
COMMENT ON COLUMN public.issues.sequence_id IS '项目内自增序号，配合 project.identifier 展示为 YD-123';
COMMENT ON COLUMN public.issues.type_code IS '工作项类型: epic(史诗) / requirement(需求) / task(任务) / defect(缺陷)';
COMMENT ON COLUMN public.issues.parent_id IS 'WBS 父工作项 FK（issues.id），NULL=顶级，limit depth ≤3';
COMMENT ON COLUMN public.issues.depth IS 'WBS 冗余层级（1..3），父项 depth+1 自动填充，>3 拒绝';
COMMENT ON COLUMN public.issues.name IS '工作项标题（短文本，索引全文检索命中）';
COMMENT ON COLUMN public.issues.description_json IS 'TipTap 编辑器结构化 JSON（Node 数组，ProseMirror 格式）';
COMMENT ON COLUMN public.issues.description_html IS '从 description_json 渲染的 HTML（展示层 + 通知邮件富文本）';
COMMENT ON COLUMN public.issues.description_stripped IS '纯文本（HTML strip 后），tsvector 全文字段索引源';
COMMENT ON COLUMN public.issues.state_id IS '状态 FK（states.id）；group ∈ backlog/unstarted/started/completed/cancelled/triage';
COMMENT ON COLUMN public.issues.priority IS '优先级: urgent(紧急) / high(高) / medium(中) / low(低) / none(无)';
COMMENT ON COLUMN public.issues.severity IS '缺陷严重程度（1=致命..5=轻微）；仅 type_code=defect 使用，必填';
COMMENT ON COLUMN public.issues.found_phase IS '缺陷发现阶段: unit(单元测试) / integration(集成) / uat(验收) / production(生产) / customer(客户反馈)';
COMMENT ON COLUMN public.issues.root_cause_category IS '缺陷根因分类: requirement(需求) / technical(技术) / environment(环境) / data(数据)；流转至 completed 时必填';
COMMENT ON COLUMN public.issues.verifier_id IS '验证人 FK（users.id）；缺陷待验证阶段可指派';
COMMENT ON COLUMN public.issues.environment IS '缺陷发现环境 JSONB（OS / Browser / Device / AppVersion）';
COMMENT ON COLUMN public.issues.reproduce_steps IS '缺陷复现步骤 JSONB {steps:[], expected:'', actual:''}';
COMMENT ON COLUMN public.issues.category IS '工作项分类标签: frontend/backend/qa/devops/design/doc 等（自由填写）';
COMMENT ON COLUMN public.issues.actual_effort IS '实际已用工时 NUMERIC(8,2) 单位: 小时；time_logs sum 回写';
COMMENT ON COLUMN public.issues.remaining_effort IS '剩余预估工时 NUMERIC(8,2) 单位: 小时；0 表示完成';
COMMENT ON COLUMN public.issues.delay_reason IS '延期原因: requirement_change(需求变更) / resource(资源) / blocked(阻塞) / other(其他)';
COMMENT ON COLUMN public.issues.source IS '需求来源: customer(客户) / internal(内部) / competitor(竞品) / other';
COMMENT ON COLUMN public.issues.point IS '故事点估算 SMALLINT 0-12；斐波那契数列 0,1,2,3,5,8,13';
COMMENT ON COLUMN public.issues.sprint_id IS '归属迭代 FK（sprints.id），同一项目内一个活跃迭代（可配置）';
COMMENT ON COLUMN public.issues.progress IS '完成百分比 0-100（冗余字段；子项 state.group=completed 时事件触发的回写）';
COMMENT ON COLUMN public.issues.start_date IS '计划开始日期（用户指定；逾期触发 risk_rule 告警）';
COMMENT ON COLUMN public.issues.target_date IS '目标完成日期（用户指定；逾期触发 risk_rule 告警）';
COMMENT ON COLUMN public.issues.completed_at IS '实际完成时间 TIMESTAMPTZ；进入 completed 状态时自动赋值';
COMMENT ON COLUMN public.issues.is_draft IS '草稿标记: true=草稿(仅草稿流可见)，false=正式发布；默认 false';
COMMENT ON COLUMN public.issues.sort_order IS '看板列内排序权重 DOUBLE PRECISION（默认 65535 末尾追加；中值插入；碎片化触发重排）';
COMMENT ON COLUMN public.issues.created_by IS '创建人 FK（users.id）；通知默认接收人';
COMMENT ON COLUMN public.issues.created_at IS '创建时间（迁移写入后不可变）';
COMMENT ON COLUMN public.issues.updated_at IS '最后修改时间（触发器 trg_xxx_updated_at 自动维护 now()）';
COMMENT ON COLUMN public.issues.deleted_at IS '软删除时间戳；NULL=有效；部分索引 WHERE deleted_at IS NULL 排除软删除';
COMMENT ON COLUMN public.issues.version IS '乐观锁版本号（默认 1）；UPDATE 条件带 version，冲突返回 409';
COMMENT ON COLUMN public.issues.found_version_id IS '缺陷发现版本 FK（versions.id）；type_code=defect 时关联';
COMMENT ON COLUMN public.issues.fix_version_id IS '缺陷修复版本 FK（versions.id）；流转至待验证/已修复时必填';
COMMENT ON COLUMN public.issues.release_version_id IS '首次发布版本 FK（versions.id）；需求/任务在发布时回填';
COMMENT ON COLUMN public.issues.search_tsv IS 'tsvector 全文索引（simple 配置，中文降级兜底；ES 为主）';

-- ----------------------------
-- Records of issues
-- ----------------------------

-- ----------------------------
-- Table structure for labels
-- ----------------------------
DROP TABLE IF EXISTS "public"."labels";
CREATE TABLE "public"."labels" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "color" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT '#8DA2C2'::text,
  "description" text COLLATE "pg_catalog"."default",
  "created_by" int8 NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now(),
  "deleted_at" timestamptz(6)
)
;
COMMENT ON TABLE public.labels IS '标签表（工作项分类标记，支持颜色、描述、软删除）';
COMMENT ON COLUMN public.labels.id IS '主键 ID';
COMMENT ON COLUMN public.labels.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.labels.name IS '标签名称（工作空间+项目内唯一，部分索引排除软删除）';
COMMENT ON COLUMN public.labels.color IS '标签颜色 HEX 值（如 #FF5733）';
COMMENT ON COLUMN public.labels.description IS '标签用途说明（可选）';
COMMENT ON COLUMN public.labels.project_id IS '所属项目 FK（NULL=工作空间级全局标签）';
COMMENT ON COLUMN public.labels.created_by IS '创建人 FK';
COMMENT ON COLUMN public.labels.created_at IS '创建时间';
COMMENT ON COLUMN public.labels.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.labels.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of labels
-- ----------------------------

-- ----------------------------
-- Table structure for metric_adjustments
-- ----------------------------
DROP TABLE IF EXISTS "public"."metric_adjustments";
CREATE TABLE "public"."metric_adjustments" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8,
  "snapshot_id" int8,
  "metric" text COLLATE "pg_catalog"."default" NOT NULL,
  "snapshot_date" date NOT NULL,
  "original_value" numeric(12,4),
  "adjusted_value" numeric(12,4) NOT NULL,
  "reason" text COLLATE "pg_catalog"."default" NOT NULL,
  "adjusted_by" int8 NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.metric_adjustments IS '指标修正记录表（admin 校准用；不覆盖原值，查询时叠加；全程审计）';
COMMENT ON COLUMN public.metric_adjustments.id IS '主键 ID';
COMMENT ON COLUMN public.metric_adjustments.snapshot_id IS '原始快照 FK（metric_snapshots.id）';
COMMENT ON COLUMN public.metric_adjustments.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.metric_adjustments.adjusted_by IS '修正人 FK（admin 角色才可操作）';
COMMENT ON COLUMN public.metric_adjustments.original_value IS '修正前原值（不可变；审计依据）';
COMMENT ON COLUMN public.metric_adjustments.adjusted_value IS '修正后值（查询时叠加；不覆盖原值）';
COMMENT ON COLUMN public.metric_adjustments.reason IS '修正原因（必填；如"剔除异常测试数据"）';
COMMENT ON COLUMN public.metric_adjustments.created_at IS '修正时间';

-- ----------------------------
-- Records of metric_adjustments
-- ----------------------------

-- ----------------------------
-- Table structure for metric_snapshots
-- ----------------------------
DROP TABLE IF EXISTS "public"."metric_snapshots";
CREATE TABLE "public"."metric_snapshots" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8,
  "granularity" text COLLATE "pg_catalog"."default" NOT NULL,
  "ref_id" int8,
  "metric" text COLLATE "pg_catalog"."default" NOT NULL,
  "value" numeric(12,4),
  "dimensions" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "snapshot_date" date NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.metric_snapshots IS '效能指标快照表（每日 01:30 聚合 job；granularity daily/sprint/version；幂等 upsert）';
COMMENT ON COLUMN public.metric_snapshots.id IS '主键 ID';
COMMENT ON COLUMN public.metric_snapshots.workspace_id IS '工作空间 FK（分区键 + RLS 依据）';
COMMENT ON COLUMN public.metric_snapshots.project_id IS '项目 FK（NULL=跨项目工作空间级聚合）';
COMMENT ON COLUMN public.metric_snapshots.granularity IS '聚合粒度: daily(日) / sprint(迭代) / version(版本)';
COMMENT ON COLUMN public.metric_snapshots.ref_id IS '粒度引用 ID（granularity=sprint→sprint_id；version→version_id；daily→NULL）';
COMMENT ON COLUMN public.metric_snapshots.metric IS '指标名: lead_time / throughput / wip_count / bug_density / escape_rate / velocity / dora_df / dora_lt / dora_cfr / dora_mttr / flow_efficiency';
COMMENT ON COLUMN public.metric_snapshots.value IS '指标值 NUMERIC；存储精确计算结果（查询时直接使用，避免重复聚合）';
COMMENT ON COLUMN public.metric_snapshots.dimensions IS '维度 JSONB {type_code, state_group, module_id, assignee_id}；钻取过滤';
COMMENT ON COLUMN public.metric_snapshots.snapshot_date IS '快照日期（幂等 upsert key: (granularity, ref_id, metric, snapshot_date)）';
COMMENT ON COLUMN public.metric_snapshots.created_at IS '写入时间（Cron 每日 01:30 聚合 Job）';
COMMENT ON COLUMN public.metric_snapshots.updated_at IS '修改时间（触发器自动维护）';

-- ----------------------------
-- Records of metric_snapshots
-- ----------------------------

-- ----------------------------
-- Table structure for modules
-- ----------------------------
DROP TABLE IF EXISTS "public"."modules";
CREATE TABLE "public"."modules" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "lead_id" int8,
  "status" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'active'::text,
  "start_date" date,
  "target_date" date,
  "sort_order" float8 NOT NULL DEFAULT 65535,
  "created_by" int8 NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now(),
  "deleted_at" timestamptz(6)
)
;
COMMENT ON TABLE public.modules IS '模块/组件表（项目内功能域划分，层级结构 ≤3 层，用于工作项分类）';
COMMENT ON COLUMN public.modules.id IS '主键 ID';
COMMENT ON COLUMN public.modules.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.modules.project_id IS '所属项目 FK';
COMMENT ON COLUMN public.modules.name IS '模块/组件名称';
COMMENT ON COLUMN public.modules.description IS '模块功能说明';
COMMENT ON COLUMN public.modules.parent_id IS '父模块 FK（modules.id）；层级 ≤3，顶层为 NULL';
COMMENT ON COLUMN public.modules.lead_id IS '模块负责人 FK（users.id）';
COMMENT ON COLUMN public.modules.created_by IS '创建人 FK';
COMMENT ON COLUMN public.modules.created_at IS '创建时间';
COMMENT ON COLUMN public.modules.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.modules.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of modules
-- ----------------------------

-- ----------------------------
-- Table structure for notification_deliveries
-- ----------------------------
DROP TABLE IF EXISTS "public"."notification_deliveries";
CREATE TABLE "public"."notification_deliveries" (
  "id" int8 NOT NULL DEFAULT nextval('notification_deliveries_id_seq'::regclass),
  "notification_id" int8 NOT NULL,
  "channel" varchar(32) COLLATE "pg_catalog"."default" NOT NULL,
  "status" varchar(16) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'pending'::character varying,
  "recipient" text COLLATE "pg_catalog"."default" NOT NULL,
  "sent_at" timestamptz(6),
  "error_msg" text COLLATE "pg_catalog"."default",
  "retry_count" int4 NOT NULL DEFAULT 0,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "next_retry_at" timestamptz(6)
)
;
COMMENT ON TABLE public.notification_deliveries IS '通知渠道投递记录表（in_app/email/wecom/dingtalk/feishu；含重试/失败状态；分区）';
COMMENT ON COLUMN public.notification_deliveries.id IS '主键 ID';
COMMENT ON COLUMN public.notification_deliveries.notification_id IS '站内信 FK（notifications.id）';
COMMENT ON COLUMN public.notification_deliveries.channel IS '发货渠道: email / wecom / dingtalk / feishu';
COMMENT ON COLUMN public.notification_deliveries.status IS '投递状态: pending / sent / failed / retrying';
COMMENT ON COLUMN public.notification_deliveries.attempt_count IS '重试次数（最大 3 次；指数退避 1min/5min/30min）';
COMMENT ON COLUMN public.notification_deliveries.error_message IS '最后错误信息（用于排障；不存敏感内容）';
COMMENT ON COLUMN public.notification_deliveries.sent_at IS '最终投递成功时间';
COMMENT ON COLUMN public.notification_deliveries.created_at IS '创建时间（分区键；按月 RANGE 分区）';
COMMENT ON COLUMN public.notification_deliveries.workspace_id IS '工作空间 FK（RLS 依据）';

-- ----------------------------
-- Records of notification_deliveries
-- ----------------------------

-- ----------------------------
-- Table structure for notification_digests
-- ----------------------------
DROP TABLE IF EXISTS "public"."notification_digests";
CREATE TABLE "public"."notification_digests" (
  "id" int8 NOT NULL DEFAULT nextval('notification_digests_id_seq'::regclass),
  "user_id" int8 NOT NULL,
  "workspace_id" int8 NOT NULL,
  "digest_type" varchar(16) COLLATE "pg_catalog"."default" NOT NULL,
  "notification_ids" int8[] NOT NULL,
  "status" varchar(16) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'pending'::character varying,
  "scheduled_for" timestamptz(6) NOT NULL,
  "sent_at" timestamptz(6),
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.notification_digests IS '通知摘要暂存表（daily/weekly 模式按用户+渠道聚合；定时合并发送）';
COMMENT ON COLUMN public.notification_digests.id IS '主键 ID';
COMMENT ON COLUMN public.notification_digests.user_id IS '用户 FK';
COMMENT ON COLUMN public.notification_digests.channel IS '聚合渠道: email / wecom / dingtalk / feishu';
COMMENT ON COLUMN public.notification_digests.payload IS '聚合数据 JSONB [{event_type, issue_id, title, body}]；按项目分组';
COMMENT ON COLUMN public.notification_digests.scheduled_for IS '计划发送时间 TIMESTAMPTZ（Cron 触发；daily=08:30 用户时区）';
COMMENT ON COLUMN public.notification_digests.sent_at IS '实际发送时间；NULL=待发送';
COMMENT ON COLUMN public.notification_digests.created_at IS '创建时间（分区键）';
COMMENT ON COLUMN public.notification_digests.workspace_id IS '工作空间 FK（RLS 依据）';

-- ----------------------------
-- Records of notification_digests
-- ----------------------------

-- ----------------------------
-- Table structure for notification_preferences
-- ----------------------------
DROP TABLE IF EXISTS "public"."notification_preferences";
CREATE TABLE "public"."notification_preferences" (
  "id" int8 NOT NULL DEFAULT nextval('notification_preferences_id_seq'::regclass),
  "user_id" int8 NOT NULL,
  "workspace_id" int8 NOT NULL,
  "event_types" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "channels" jsonb NOT NULL DEFAULT '["in_app"]'::jsonb,
  "digest" varchar(16) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'realtime'::character varying,
  "dnd_enabled" bool NOT NULL DEFAULT false,
  "dnd_start" time(6) DEFAULT '22:00:00'::time without time zone,
  "dnd_end" time(6) DEFAULT '08:00:00'::time without time zone,
  "is_enabled" bool NOT NULL DEFAULT true,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.notification_preferences IS '用户通知偏好订阅表（按 项目×事件类型×渠道三维开关；支持免打扰 digest）';
COMMENT ON COLUMN public.notification_preferences.id IS '主键 ID';
COMMENT ON COLUMN public.notification_preferences.user_id IS '用户 FK';
COMMENT ON COLUMN public.notification_preferences.scope IS '订阅范围: workspace / project；决定 ref_id 语义';
COMMENT ON COLUMN public.notification_preferences.ref_id IS '范围 ID（scope=workspace → workspace_id；scope=project → project_id）';
COMMENT ON COLUMN public.notification_preferences.event_type IS '事件类型（通配: issue.* / sprint.* / version.* / automation.*）';
COMMENT ON COLUMN public.notification_preferences.channel IS '通知渠道: in_app / email / wecom / dingtalk / feishu';
COMMENT ON COLUMN public.notification_preferences.is_enabled IS '渠道开关: true=启用, false=禁用（覆盖默认矩阵）';
COMMENT ON COLUMN public.notification_preferences.digest IS '投递模式: realtime(实时) / daily(每日 08:30 聚合) / weekly(每周一聚合)';
COMMENT ON COLUMN public.notification_preferences.dnd_start IS '免打扰开始时间 (HH:MM 用户时区 如 22:00)；期间 realtime 降级为 digest';
COMMENT ON COLUMN public.notification_preferences.dnd_end IS '免打扰结束时间 (HH:MM 如 08:00)；高优事件(mention/automation.failed)可豁免';
COMMENT ON COLUMN public.notification_preferences.created_at IS '创建时间';
COMMENT ON COLUMN public.notification_preferences.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON TABLE "public"."notification_preferences" IS '用户通知偏好配置';

-- ----------------------------
-- Records of notification_preferences
-- ----------------------------

-- ----------------------------
-- Table structure for notifications
-- ----------------------------
DROP TABLE IF EXISTS "public"."notifications";
CREATE TABLE "public"."notifications" (
  "id" int8 NOT NULL DEFAULT nextval('notifications_id_seq'::regclass),
  "workspace_id" int8 NOT NULL,
  "recipient_id" int8 NOT NULL,
  "event_type" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "entity_type" varchar(32) COLLATE "pg_catalog"."default" NOT NULL,
  "entity_id" int8 NOT NULL,
  "title" varchar(256) COLLATE "pg_catalog"."default" NOT NULL,
  "body" text COLLATE "pg_catalog"."default",
  "action_url" varchar(512) COLLATE "pg_catalog"."default",
  "actor_id" int8,
  "actor_name" varchar(128) COLLATE "pg_catalog"."default",
  "is_read" bool NOT NULL DEFAULT false,
  "is_archived" bool NOT NULL DEFAULT false,
  "read_at" timestamptz(6),
  "channel" varchar(32) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'in_app'::character varying,
  "payload" jsonb DEFAULT '{}'::jsonb,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.notifications IS '站内信表（按月分区；recipient/actor/issue/event_type/data JSONB；WS 实时推送 + 未读数缓存）';
COMMENT ON COLUMN public.notifications.id IS '主键 ID';
COMMENT ON COLUMN public.notifications.workspace_id IS '工作空间 FK（RLS + 按月分区键）';
COMMENT ON COLUMN public.notifications.recipient_id IS '接收人 FK（users.id；通知投递主目标）';
COMMENT ON COLUMN public.notifications.actor_id IS '触发人 FK（users.id）；自我豁免: 操作者==接收人不落库';
COMMENT ON COLUMN public.notifications.issue_id IS '关联工作项 FK（可 NULL；点击跳转定位）';
COMMENT ON COLUMN public.notifications.event_type IS '事件类型: issue.created/updated/status_changed/assigned/commented/mentioned；sprint.* / version.* / automation.*';
COMMENT ON COLUMN public.notifications.title IS '通知标题（多语言模板渲染后；纯文本）';
COMMENT ON COLUMN public.notifications.body IS '通知正文（纯文本摘要；邮件使用 HTML 模板二次渲染）';
COMMENT ON COLUMN public.notifications.data IS '跳转上下文 JSONB {url, issue_identifier, project_identifier}；前端点击定位';
COMMENT ON COLUMN public.notifications.is_read IS '已读状态: true=已读(已点击), false=未读；未读数缓存 unread:{uid}';
COMMENT ON COLUMN public.notifications.read_at IS '首次阅读时间 TIMESTAMPTZ（列表"全部已读"触发）';
COMMENT ON COLUMN public.notifications.created_at IS '创建时间 TIMESTAMPTZ（列表/游标分页倒序 + 分区裁剪）';
COMMENT ON COLUMN "public"."notifications"."event_type" IS '事件类型: issue.created/issue.assigned/issue.status_changed/comment.created/sprint.started/sprint.completed/version.released/member.added';
COMMENT ON COLUMN "public"."notifications"."entity_type" IS '关联对象类型: issue/sprint/version/project/workspace/comment';
COMMENT ON COLUMN "public"."notifications"."channel" IS '通知渠道: in_app(站内)/email/sms/wecom/dingtalk/feishu';
COMMENT ON TABLE "public"."notifications" IS '通知消息表（站内铃铛+多渠道预留）';

-- ----------------------------
-- Records of notifications
-- ----------------------------

-- ----------------------------
-- Table structure for password_reset_tokens
-- ----------------------------
DROP TABLE IF EXISTS "public"."password_reset_tokens";
CREATE TABLE "public"."password_reset_tokens" (
  "id" int8 NOT NULL DEFAULT nextval('password_reset_tokens_id_seq'::regclass),
  "user_id" int8 NOT NULL,
  "token_hash" text COLLATE "pg_catalog"."default" NOT NULL,
  "expires_at" timestamptz(6) NOT NULL,
  "used_at" timestamptz(6),
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.password_reset_tokens IS '密码重置令牌表（UUID 一次性令牌 + 过期时间 + used_at）';
COMMENT ON COLUMN public.password_reset_tokens.id IS '主键 ID';
COMMENT ON COLUMN public.password_reset_tokens.user_id IS '用户 FK';
COMMENT ON COLUMN public.password_reset_tokens.token_hash IS '一次性令牌 SHA-256 hash（邮件链接中使用 token 本身）';
COMMENT ON COLUMN public.password_reset_tokens.expires_at IS '有效期 TIMESTAMPTZ（默认 1 小时；过期即失效）';
COMMENT ON COLUMN public.password_reset_tokens.used_at IS '首个使用时间（标志已消费；一次性原则）';
COMMENT ON COLUMN public.password_reset_tokens.created_at IS '创建时间';

-- ----------------------------
-- Records of password_reset_tokens
-- ----------------------------

-- ----------------------------
-- Table structure for project_sequences
-- ----------------------------
DROP TABLE IF EXISTS "public"."project_sequences";
CREATE TABLE "public"."project_sequences" (
  "project_id" int8 NOT NULL,
  "next_value" int8 NOT NULL DEFAULT 1
)
;
COMMENT ON TABLE public.project_sequences IS '项目发号器表（project_id → next_value 原子自增；发工作项编号 YD-123；允许跳号）';
COMMENT ON COLUMN public.project_sequences.project_id IS '项目 FK（PRIMARY KEY）';
COMMENT ON COLUMN public.project_sequences.next_value IS '当前已发号 + 1（原子自增: SET next_value = next_value + 1 RETURNING next_value - 1）';
COMMENT ON COLUMN public.project_sequences.created_at IS '创建时间';
COMMENT ON COLUMN public.project_sequences.updated_at IS '修改时间（触发器自动维护）';

-- ----------------------------
-- Records of project_sequences
-- ----------------------------

-- ----------------------------
-- Table structure for projects
-- ----------------------------
DROP TABLE IF EXISTS "public"."projects";
CREATE TABLE "public"."projects" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "public_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "slug" text COLLATE "pg_catalog"."default" NOT NULL,
  "identifier" text COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "network" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'public'::text,
  "icon" text COLLATE "pg_catalog"."default",
  "color" text COLLATE "pg_catalog"."default",
  "status" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'active'::text,
  "sort_order" float8 NOT NULL DEFAULT 65535,
  "created_by" int8 NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now(),
  "deleted_at" timestamptz(6),
  "template" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'generic'::text,
  "modules" jsonb NOT NULL DEFAULT '{"intake": true, "sprint": true, "version": true, "estimate": true}'::jsonb
)
;
COMMENT ON TABLE public.projects IS '项目表（工作项聚合根容器，含状态模板、工作流配置、封面、默认外观）';
COMMENT ON COLUMN public.projects.id IS '主键 ID';
COMMENT ON COLUMN public.projects.workspace_id IS '所属工作空间 FK';
COMMENT ON COLUMN public.projects.name IS '项目名称';
COMMENT ON COLUMN public.projects.identifier IS '项目标识符（大写 2-10 字符，工作空间内唯一；用于 YD-123 编号）';
COMMENT ON COLUMN public.projects.description IS '项目描述（富文本/Markdown）';
COMMENT ON COLUMN public.projects.state IS '项目状态: active(活跃) / archived(归档) / deleted(软删除)';
COMMENT ON COLUMN public.projects.default_view IS '默认视图偏好: list/board/calendar/gantt';
COMMENT ON COLUMN public.projects.cover_image IS '封面图片附件 ID';
COMMENT ON COLUMN public.projects.start_date IS '项目开始日期';
COMMENT ON COLUMN public.projects.target_date IS '项目目标日期';
COMMENT ON COLUMN public.projects.created_by IS '创建人 FK';
COMMENT ON COLUMN public.projects.created_at IS '创建时间';
COMMENT ON COLUMN public.projects.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.projects.deleted_at IS '软删除时间戳';
COMMENT ON COLUMN public.projects.version IS '乐观锁版本号（默认 1）';
COMMENT ON COLUMN public.projects.is_default IS '是否工作空间默认项目: true=默认（用于项目选择器/快捷入口）';
COMMENT ON COLUMN public.projects.icon IS '项目图标（Emoji / Lucide 图标名）';
COMMENT ON COLUMN public.projects.modules IS '功能模块开关 JSON: {intake, sprint, version, estimate}';

-- ----------------------------
-- Records of projects
-- ----------------------------

-- ----------------------------
-- Table structure for recent_items
-- ----------------------------
DROP TABLE IF EXISTS "public"."recent_items";
CREATE TABLE "public"."recent_items" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "item_type" text COLLATE "pg_catalog"."default" NOT NULL,
  "item_id" int8 NOT NULL,
  "project_id" int8,
  "title" text COLLATE "pg_catalog"."default",
  "identifier" text COLLATE "pg_catalog"."default",
  "accessed_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.recent_items IS '最近访问记录表（用户最近查看/操作的工作项；touch 触发 updated_at 更新）';
COMMENT ON COLUMN public.recent_items.id IS '主键 ID';
COMMENT ON COLUMN public.recent_items.user_id IS '用户 FK';
COMMENT ON COLUMN public.recent_items.workspace_id IS '工作空间 FK（RLS 依据）';
COMMENT ON COLUMN public.recent_items.project_id IS '所属项目 FK';
COMMENT ON COLUMN public.recent_items.item_type IS '最近访问类型: issue / sprint / version / project / page';
COMMENT ON COLUMN public.recent_items.item_id IS '关联 ID（BIGINT 统一）';
COMMENT ON COLUMN public.recent_items.access_count IS '访问次数；首页"最近"列表排序依据（加权访问时间 + 频次）';
COMMENT ON COLUMN public.recent_items.created_at IS '首次访问时间';
COMMENT ON COLUMN public.recent_items.updated_at IS '最后访问时间（触发器 trg_recent_items_touch 更新）';

-- ----------------------------
-- Records of recent_items
-- ----------------------------

-- ----------------------------
-- Table structure for risk_alerts
-- ----------------------------
DROP TABLE IF EXISTS "public"."risk_alerts";
CREATE TABLE "public"."risk_alerts" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8,
  "rule_id" int8 NOT NULL,
  "severity" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'medium'::text,
  "title" text COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "metadata" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "is_resolved" bool NOT NULL DEFAULT false,
  "resolved_at" timestamptz(6),
  "resolved_by" int8,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.risk_alerts IS '风险告警记录表（越限事件含快照、关联问题、处理状态）';
COMMENT ON COLUMN public.risk_alerts.id IS '主键 ID';
COMMENT ON COLUMN public.risk_alerts.rule_id IS '触发规则 FK';
COMMENT ON COLUMN public.risk_alerts.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.risk_alerts.project_id IS '项目 FK';
COMMENT ON COLUMN public.risk_alerts.issue_id IS '关联工作项 FK（可 NULL；如 sprint 级别告警）';
COMMENT ON COLUMN public.risk_alerts.severity IS '告警严重度: info / warning / critical';
COMMENT ON COLUMN public.risk_alerts.message IS '告警描述（含触发值/阈值对比）';
COMMENT ON COLUMN public.risk_alerts.snapshot IS '告警快照 JSONB（当时状态: 迭代进度/关键路径/剩余工时等）';
COMMENT ON COLUMN public.risk_alerts.status IS '处理状态: open(待处理) / acknowledged(已确认) / resolved(已解决) / dismissed(已忽略)';
COMMENT ON COLUMN public.risk_alerts.resolved_by IS '处理人 FK（status=resolved 时必填）';
COMMENT ON COLUMN public.risk_alerts.resolved_at IS '处理时间';
COMMENT ON COLUMN public.risk_alerts.created_at IS '触发时间 TIMESTAMPTZ';
COMMENT ON COLUMN public.risk_alerts.workspace_id2 IS 'RLS 依据（若有）';

-- ----------------------------
-- Records of risk_alerts
-- ----------------------------

-- ----------------------------
-- Table structure for risk_rules
-- ----------------------------
DROP TABLE IF EXISTS "public"."risk_rules";
CREATE TABLE "public"."risk_rules" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8,
  "rule_name" text COLLATE "pg_catalog"."default" NOT NULL,
  "rule_type" text COLLATE "pg_catalog"."default" NOT NULL,
  "condition_json" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "notify_channels" text[] COLLATE "pg_catalog"."default" NOT NULL DEFAULT '{}'::text[],
  "is_active" bool NOT NULL DEFAULT true,
  "last_triggered" timestamptz(6),
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.risk_rules IS '风险预警规则表（监控阈值: 逾期/阻塞/依赖链/燃尽偏差；越限→通知）';
COMMENT ON COLUMN public.risk_rules.id IS '主键 ID';
COMMENT ON COLUMN public.risk_rules.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.risk_rules.project_id IS '项目 FK（NULL=工作空间级通用规则）';
COMMENT ON COLUMN public.risk_rules.name IS '规则名称';
COMMENT ON COLUMN public.risk_rules.metric IS '监控指标: overdue_days / blocked_days / dependency_chain / sprint_deviation';
COMMENT ON COLUMN public.risk_rules.operator IS '比较运算符: > / >= / < / <= / ==';
COMMENT ON COLUMN public.risk_rules.threshold IS '阈值（数值；如逾期 3 天、燃尽偏差 20%）';
COMMENT ON COLUMN public.risk_rules.severity IS '告警严重度: info / warning / critical';
COMMENT ON COLUMN public.risk_rules.is_active IS '启用开关: true=生效, false=暂停';
COMMENT ON COLUMN public.risk_rules.notification_channels IS '通知渠道 JSONB [{channel, target}]；默认项目 admin + 规则创建者';
COMMENT ON COLUMN public.risk_rules.created_by IS '创建人 FK';
COMMENT ON COLUMN public.risk_rules.created_at IS '创建时间';
COMMENT ON COLUMN public.risk_rules.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.risk_rules.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of risk_rules
-- ----------------------------

-- ----------------------------
-- Table structure for rule_executions
-- ----------------------------
DROP TABLE IF EXISTS "public"."rule_executions";
CREATE TABLE "public"."rule_executions" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8,
  "rule_id" int8 NOT NULL,
  "trigger_event_id" int8,
  "status" text COLLATE "pg_catalog"."default" NOT NULL,
  "duration_ms" int4,
  "error_message" text COLLATE "pg_catalog"."default",
  "context_json" jsonb,
  "trigger_depth" int2 NOT NULL DEFAULT 0,
  "via_automation" bool NOT NULL DEFAULT false,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.rule_executions IS '规则执行审计表（按月分区；status/duration/results；连续失败 3 次触发熔断）';
COMMENT ON COLUMN public.rule_executions.id IS '主键 ID';
COMMENT ON COLUMN public.rule_executions.rule_id IS '规则 FK（automation_rules.id）';
COMMENT ON COLUMN public.rule_executions.issue_id IS '关联工作项 FK（可 NULL；非 issue 触发器则为 NULL）';
COMMENT ON COLUMN public.rule_executions.status IS '执行状态: success / failure / skipped / timeout';
COMMENT ON COLUMN public.rule_executions.duration_ms IS '执行耗时毫秒（含条件求值 + 所有 Action 总时长）';
COMMENT ON COLUMN public.rule_executions.results IS '执行明细 JSONB [{action_type, status, error?}]；success 时记录变更后值';
COMMENT ON COLUMN public.rule_executions.error_message IS '失败错误信息（failure 时必填；用于排障）';
COMMENT ON COLUMN public.rule_executions.created_at IS '执行时间 TIMESTAMPTZ（按月分区键；30 天 TTL drop partition）';
COMMENT ON COLUMN public.rule_executions.workspace_id IS '工作空间 FK（RLS 依据）';

-- ----------------------------
-- Records of rule_executions
-- ----------------------------

-- ----------------------------
-- Table structure for schema_migrations
-- ----------------------------
DROP TABLE IF EXISTS "public"."schema_migrations";
CREATE TABLE "public"."schema_migrations" (
  "version" int8 NOT NULL,
  "dirty" bool NOT NULL
)
;
COMMENT ON TABLE public.schema_migrations IS 'Schema 迁移记录表（Flyway 风格版本号；CI 校验连续性）';
COMMENT ON COLUMN public.schema_migrations.version IS '迁移版本编号（NNNN；递增；CI 校验连续性）';
COMMENT ON COLUMN public.schema_migrations.applied_at IS '应用时间（迁移框架自动写入；幂等检测依据）';
COMMENT ON COLUMN public.schema_migrations.description IS '迁移描述（便于排查 + CHANGELOG 对齐）';
COMMENT ON COLUMN public.schema_migrations.execution_ms IS '执行耗时毫秒（用于慢迁移告警）';
COMMENT ON COLUMN public.schema_migrations.checksum IS '迁移内容 SHA-256（防篡改；运行时校验与声明的 checksum 匹配）';

-- ----------------------------
-- Records of schema_migrations
-- ----------------------------
INSERT INTO "public"."schema_migrations" VALUES (24, 'f');

-- ----------------------------
-- Table structure for search_bookmarks
-- ----------------------------
DROP TABLE IF EXISTS "public"."search_bookmarks";
CREATE TABLE "public"."search_bookmarks" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8,
  "user_id" int8 NOT NULL,
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "query" text COLLATE "pg_catalog"."default",
  "filters" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "is_shared" bool NOT NULL DEFAULT false,
  "sort_order" float8 NOT NULL DEFAULT 65535,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now(),
  "deleted_at" timestamptz(6)
)
;
COMMENT ON TABLE public.search_bookmarks IS '搜索收藏查询表（用户保存的常用 JQL 查询；可分享）';
COMMENT ON COLUMN public.search_bookmarks.id IS '主键 ID';
COMMENT ON COLUMN public.search_bookmarks.user_id IS '用户 FK';
COMMENT ON COLUMN public.search_bookmarks.workspace_id IS '工作空间 FK（RLS 依据）';
COMMENT ON COLUMN public.search_bookmarks.name IS '收藏查询名称（如"我的待办"）';
COMMENT ON COLUMN public.search_bookmarks.query IS 'JQL 查询字符串（如"assignee:me status:todo"）；完整保存';
COMMENT ON COLUMN public.search_bookmarks.description IS '收藏说明（可选）';
COMMENT ON COLUMN public.search_bookmarks.is_shared IS '是否分享: true=同工作空间成员可见, false=仅自己';
COMMENT ON COLUMN public.search_bookmarks.sort_order IS '展示排序权重';
COMMENT ON COLUMN public.search_bookmarks.created_by IS '创建人 FK';
COMMENT ON COLUMN public.search_bookmarks.created_at IS '创建时间';
COMMENT ON COLUMN public.search_bookmarks.updated_at IS '修改时间';

-- ----------------------------
-- Records of search_bookmarks
-- ----------------------------

-- ----------------------------
-- Table structure for search_documents
-- ----------------------------
DROP TABLE IF EXISTS "public"."search_documents";
CREATE TABLE "public"."search_documents" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "doc_type" text COLLATE "pg_catalog"."default" NOT NULL,
  "doc_id" int8 NOT NULL,
  "title" text COLLATE "pg_catalog"."default" NOT NULL,
  "identifier" text COLLATE "pg_catalog"."default",
  "content" text COLLATE "pg_catalog"."default",
  "search_tsv" tsvector,
  "metadata" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.search_documents IS '搜索文档表（ES 索引失败的降级兜底；PG to_tsvector 源）';
COMMENT ON COLUMN public.search_documents.id IS '主键 ID';
COMMENT ON COLUMN public.search_documents.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.search_documents.project_id IS '项目 FK';
COMMENT ON COLUMN public.search_documents.item_type IS '索引对象类型: issue / page / doc';
COMMENT ON COLUMN public.search_documents.item_id IS '关联 ID（BIGINT）';
COMMENT ON COLUMN public.search_documents.title IS '索引标题';
COMMENT ON COLUMN public.search_documents.content IS '索引内容（纯文本；tsvector 计算源）';
COMMENT ON COLUMN public.search_documents.metadata IS '索引元数据 JSONB（attribution/comments/sprint/labels 等）';
COMMENT ON COLUMN public.search_documents.indexed_at IS '索引时间；对账 Job 检测 updated_at > indexed_at 重做';
COMMENT ON COLUMN public.search_documents.created_at IS '创建时间';
COMMENT ON COLUMN public.search_documents.updated_at IS '修改时间';

-- ----------------------------
-- Records of search_documents
-- ----------------------------

-- ----------------------------
-- Table structure for search_history
-- ----------------------------
DROP TABLE IF EXISTS "public"."search_history";
CREATE TABLE "public"."search_history" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "query" text COLLATE "pg_catalog"."default" NOT NULL,
  "filters" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "result_count" int4 NOT NULL DEFAULT 0,
  "searched_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.search_history IS '搜索历史表（localStorage 20 条 + 服务端同步；自动补全数据源）';
COMMENT ON COLUMN public.search_history.id IS '主键 ID';
COMMENT ON COLUMN public.search_history.user_id IS '用户 FK';
COMMENT ON COLUMN public.search_history.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.search_history.query IS '完整搜索字符串（含 JQL 语法）';
COMMENT ON COLUMN public.search_history.result_count IS '返回结果数（缓存用于效果分析）';
COMMENT ON COLUMN public.search_history.clicked_item_id IS '点击第一条结果（优化搜索算法依据）';
COMMENT ON COLUMN public.search_history.searched_at IS '搜索时间 TIMESTAMPTZ；自动补全历史数据源';

-- ----------------------------
-- Records of search_history
-- ----------------------------

-- ----------------------------
-- Table structure for sprint_issues
-- ----------------------------
DROP TABLE IF EXISTS "public"."sprint_issues";
CREATE TABLE "public"."sprint_issues" (
  "sprint_id" int8 NOT NULL,
  "issue_id" int8 NOT NULL,
  "added_midway" bool NOT NULL DEFAULT false,
  "sort_order" float8 NOT NULL DEFAULT 65535,
  "added_at" timestamptz(6) NOT NULL DEFAULT now(),
  "added_by" int8
)
;
COMMENT ON TABLE public.sprint_issues IS '迭代-工作项关联表（含中途加项标记 added_midway，复盘报告使用）';
COMMENT ON COLUMN public.sprint_issues.id IS '主键 ID';
COMMENT ON COLUMN public.sprint_issues.sprint_id IS '迭代 FK';
COMMENT ON COLUMN public.sprint_issues.issue_id IS '工作项 FK';
COMMENT ON COLUMN public.sprint_issues.added_midway IS '是否中途加入: true=迭代启动后新增（复盘报告单独统计对速率影响）';
COMMENT ON COLUMN public.sprint_issues.created_at IS '关联创建时间（即工作项加入迭代的时间点）';
COMMENT ON COLUMN public.sprint_issues.workspace_id IS '工作空间 FK（RLS 依据 + 复合索引）';

-- ----------------------------
-- Records of sprint_issues
-- ----------------------------

-- ----------------------------
-- Table structure for sprint_snapshots
-- ----------------------------
DROP TABLE IF EXISTS "public"."sprint_snapshots";
CREATE TABLE "public"."sprint_snapshots" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "sprint_id" int8 NOT NULL,
  "snapshot_date" date NOT NULL,
  "data" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "deleted_at" timestamptz(6)
)
;
COMMENT ON TABLE public.sprint_snapshots IS '迭代燃尽快照表（Cron 每日 00:05 写入；by_state_group 字段支撑燃起图/CFD）';
COMMENT ON COLUMN public.sprint_snapshots.id IS '主键 ID';
COMMENT ON COLUMN public.sprint_snapshots.sprint_id IS '迭代 FK';
COMMENT ON COLUMN public.sprint_snapshots.snapshot_date IS '快照日期（每 sprint+date 唯一；Cron 每日 00:05 写入）';
COMMENT ON COLUMN public.sprint_snapshots.total_points IS '当日总计划故事点';
COMMENT ON COLUMN public.sprint_snapshots.done_points IS '当日已完成故事点';
COMMENT ON COLUMN public.sprint_snapshots.by_state_group IS '各状态组故事点分布 JSONB {backlog:N, unstarted:N, started:N, completed:N}';
COMMENT ON COLUMN public.sprint_snapshots.added_points IS '启动后新增故事点（复盘中期加入影响计算）';
COMMENT ON COLUMN public.sprint_snapshots.removed_points IS '启动后移除故事点';
COMMENT ON COLUMN public.sprint_snapshots.created_at IS '写入时间';
COMMENT ON COLUMN public.sprint_snapshots.workspace_id IS '工作空间 FK（RLS 依据）';

-- ----------------------------
-- Records of sprint_snapshots
-- ----------------------------

-- ----------------------------
-- Table structure for sprints
-- ----------------------------
DROP TABLE IF EXISTS "public"."sprints";
CREATE TABLE "public"."sprints" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "goal" text COLLATE "pg_catalog"."default",
  "status" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'planned'::text,
  "start_date" date,
  "end_date" date,
  "capacity" numeric(10,2),
  "owner_id" int8,
  "viewport" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "review_snapshot" jsonb,
  "started_at" timestamptz(6),
  "completed_at" timestamptz(6),
  "created_by" int8 NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now(),
  "deleted_at" timestamptz(6),
  "version_id" int8
)
;
COMMENT ON TABLE public.sprints IS '敏捷迭代表（生命周期 planned→active→completed；容量/目标/速率/乐观锁）';
COMMENT ON COLUMN public.sprints.id IS '主键 ID';
COMMENT ON COLUMN public.sprints.workspace_id IS '所属工作空间 FK';
COMMENT ON COLUMN public.sprints.project_id IS '所属项目 FK';
COMMENT ON COLUMN public.sprints.name IS '迭代名称（如"Sprint 42 - 用户画像"）';
COMMENT ON COLUMN public.sprints.goal IS '迭代目标（简短描述；燃尽图上方展示）';
COMMENT ON COLUMN public.sprints.status IS '迭代状态: planned(计划中) / active(进行中) / completed(已结束)';
COMMENT ON COLUMN public.sprints.start_date IS '迭代开始日期（active 时必填；同一项目同一时间仅一个 active）';
COMMENT ON COLUMN public.sprints.end_date IS '迭代结束日期（active 时必填；结束日触发 sprint.ending_soon 提醒）';
COMMENT ON COLUMN public.sprints.capacity IS '团队容量（人天）；与故事点总和对比计算饱和度';
COMMENT ON COLUMN public.sprints.total_points IS '当前迭代总故事点（redundant，SprintIssue 事件回写更新）';
COMMENT ON COLUMN public.sprints.done_points IS '已完成故事点（redundant；SprintIssue 完成事件回写）';
COMMENT ON COLUMN public.sprints.created_by IS '创建人 FK';
COMMENT ON COLUMN public.sprints.created_at IS '创建时间';
COMMENT ON COLUMN public.sprints.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.sprints.deleted_at IS '软删除时间戳';
COMMENT ON COLUMN public.sprints.version IS '乐观锁版本号（默认 1）；UPDATE 条件带 version 防并发冲突';

-- ----------------------------
-- Records of sprints
-- ----------------------------

-- ----------------------------
-- Table structure for state_transitions
-- ----------------------------
DROP TABLE IF EXISTS "public"."state_transitions";
CREATE TABLE "public"."state_transitions" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "type_code" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'all'::text,
  "from_state_id" int8 NOT NULL,
  "to_state_id" int8 NOT NULL,
  "required_fields" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.state_transitions IS '状态流转规则表（按 project×type 维度定义合法流转；含必填字段、允许角色约束）';
COMMENT ON COLUMN public.state_transitions.id IS '主键 ID';
COMMENT ON COLUMN public.state_transitions.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.state_transitions.project_id IS '所属项目 FK';
COMMENT ON COLUMN public.state_transitions.type_code IS '工作项类型: requirement / task / defect；独立流转集';
COMMENT ON COLUMN public.state_transitions.from_state_id IS '起始状态 FK（states.id）';
COMMENT ON COLUMN public.state_transitions.to_state_id IS '目标状态 FK（states.id）';
COMMENT ON COLUMN public.state_transitions.required_fields IS '流转必填字段 JSONB [{field, condition}]（如缺陷→已完成要求 root_cause_category 非空）';
-- FIXED: COMMENT ON COLUMN public.state_transitions.allowed_roles IS '允许执行的角色列表 JSONB ['owner','admin',...]；空数组=继承项目角色默认';
COMMENT ON COLUMN public.state_transitions.created_by IS '创建人 FK';
COMMENT ON COLUMN public.state_transitions.created_at IS '创建时间';
COMMENT ON COLUMN public.state_transitions.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.state_transitions.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of state_transitions
-- ----------------------------

-- ----------------------------
-- Table structure for states
-- ----------------------------
DROP TABLE IF EXISTS "public"."states";
CREATE TABLE "public"."states" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "group" text COLLATE "pg_catalog"."default" NOT NULL,
  "color" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT '#8DA2C2'::text,
  "sequence" float8 NOT NULL DEFAULT 65535,
  "is_default" bool NOT NULL DEFAULT false,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now(),
  "deleted_at" timestamptz(6),
  "template_set" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'custom'::text
)
;
COMMENT ON TABLE public.states IS '状态表（项目级自定义状态集；group ∈ backlog/unstarted/started/completed/cancelled/triage）';
COMMENT ON COLUMN public.states.id IS '主键 ID';
COMMENT ON COLUMN public.states.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.states.project_id IS '所属项目 FK';
COMMENT ON COLUMN public.states.name IS '状态显示名（如"ToDo"/"In Progress"/"Done"）';
COMMENT ON COLUMN public.states.group IS '状态组: backlog / unstarted / started / completed / cancelled / triage（前端据此渲染卡片颜色 + 看板列）';
COMMENT ON COLUMN public.states.color IS '状态颜色 HEX 值';
COMMENT ON COLUMN public.states.sequence IS '状态在组内排序权重（升序）';
COMMENT ON COLUMN public.states.description IS '状态含义说明（可选）';
COMMENT ON COLUMN public.states.is_default IS '是否项目默认状态: true=新工作项初始状态（每个 type_code 一个默认）';
COMMENT ON COLUMN public.states.type_code IS '工作项类型: requirement / task / defect；决定状态集隔离';
COMMENT ON COLUMN public.states.created_by IS '创建人 FK';
COMMENT ON COLUMN public.states.created_at IS '创建时间';
COMMENT ON COLUMN public.states.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.states.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of states
-- ----------------------------

-- ----------------------------
-- Table structure for time_logs
-- ----------------------------
DROP TABLE IF EXISTS "public"."time_logs";
CREATE TABLE "public"."time_logs" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "issue_id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "spent_date" date NOT NULL DEFAULT CURRENT_DATE,
  "duration_minutes" int4 NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now(),
  "deleted_at" timestamptz(6)
)
;
COMMENT ON TABLE public.time_logs IS '工时记录表（单位: 分钟；关联工作项 + 用户，差值回写 actual_effort）';
COMMENT ON COLUMN public.time_logs.id IS '主键 ID';
COMMENT ON COLUMN public.time_logs.issue_id IS '关联工作项 FK';
COMMENT ON COLUMN public.time_logs.user_id IS '登记人 FK';
COMMENT ON COLUMN public.time_logs.minutes IS '工时分钟数（正整数；写入/编辑/删除差值回写 actual_effort）';
COMMENT ON COLUMN public.time_logs.log_date IS '工时日期（用户可指定历史日期；默认当天）';
COMMENT ON COLUMN public.time_logs.description IS '工时说明（可选: 做了什么）';
COMMENT ON COLUMN public.time_logs.billable IS '是否计费工时: true=计费, false=不计费';
COMMENT ON COLUMN public.time_logs.created_by IS '登记人/修改人 FK';
COMMENT ON COLUMN public.time_logs.created_at IS '创建时间';
COMMENT ON COLUMN public.time_logs.updated_at IS '修改时间';
COMMENT ON COLUMN public.time_logs.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of time_logs
-- ----------------------------

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS "public"."users";
CREATE TABLE "public"."users" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "public_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" text COLLATE "pg_catalog"."default" NOT NULL,
  "password_hash" text COLLATE "pg_catalog"."default" NOT NULL,
  "display_name" text COLLATE "pg_catalog"."default" NOT NULL,
  "avatar_url" text COLLATE "pg_catalog"."default",
  "is_active" bool NOT NULL DEFAULT true,
  "timezone" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'Asia/Shanghai'::text,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now(),
  "deleted_at" timestamptz(6)
)
;
COMMENT ON TABLE public.users IS '平台用户表（跨工作空间；认证信息、MFA、时区、偏好设置）';
COMMENT ON COLUMN public.users.id IS '主键 ID';
COMMENT ON COLUMN public.users.email IS '邮箱（小写唯一；登录主凭证；CI 强制唯一）';
COMMENT ON COLUMN public.users.password_hash IS 'bcrypt(cost=12) 密码哈希值';
COMMENT ON COLUMN public.users.first_name IS '名字';
COMMENT ON COLUMN public.users.last_name IS '姓氏';
COMMENT ON COLUMN public.users.display_name IS '显示名（默认 first + last；可自定义）';
COMMENT ON COLUMN public.users.avatar IS '头像附件 ID（attachments.id）';
COMMENT ON COLUMN public.users.is_active IS '账号激活状态: true=激活, false=禁用（软锁定）';
COMMENT ON COLUMN public.users.is_email_verified IS '邮箱是否已验证: true=已验证, false=未验证（发送验证邮件）';
COMMENT ON COLUMN public.users.mfa_secret IS 'TOTP 双因子密钥（AES-GCM 加密存储；NULL=未启用）';
COMMENT ON COLUMN public.users.last_login_at IS '最后登录时间（登录成功后更新；判断账号活跃度）';
COMMENT ON COLUMN public.users.timezone IS '用户时区（IANA 名称如 Asia/Shanghai；用于 digest/免打扰计算）';
COMMENT ON COLUMN public.users.locale IS '偏好语言 (en/zh-CN/zh-TW/ja/ko)';
COMMENT ON COLUMN public.users.preferences IS '用户偏好 JSONB（主题/快捷手势/通知默认开关等）';
COMMENT ON COLUMN public.users.created_at IS '注册时间';
COMMENT ON COLUMN public.users.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.users.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of users
-- ----------------------------
INSERT INTO "public"."users" OVERRIDING SYSTEM VALUE VALUES (1, 'ac16c343-bc4e-4350-a91c-51a5c13d9ec8', 'admin@ydsz.dev', '$2a$10$y6TLegzzFJaHATj.Chh8vuGECtjCv4mdPe9iiyjMujnIUDbjZYPj2', '系统管理员', NULL, 't', 'Asia/Shanghai', '2026-08-08 00:01:21.898145+08', '2026-08-08 00:01:21.898145+08', NULL);
INSERT INTO "public"."users" OVERRIDING SYSTEM VALUE VALUES (2, '173e9870-49f3-4005-9e3d-b8f282b89c4d', 'pm@ydsz.dev', '$2a$10$/0jRHCuSnbZR8eLMsF9fCuVl7ATuqxcOv/1HupSY7yuwUy5i5f1tS', '李产品', NULL, 't', 'Asia/Shanghai', '2026-08-08 00:01:21.952153+08', '2026-08-08 00:01:21.952153+08', NULL);
INSERT INTO "public"."users" OVERRIDING SYSTEM VALUE VALUES (3, '96366e29-eaba-4fbb-b878-4feeec48921c', 'dev@ydsz.dev', '$2a$10$UInxW0QAqJOhaZwTY5DP/eDwxwB5AseWyfjQZnnvHdBYzwygczdUu', '王工程', NULL, 't', 'Asia/Shanghai', '2026-08-08 00:01:22.011685+08', '2026-08-08 00:01:22.011685+08', NULL);
INSERT INTO "public"."users" OVERRIDING SYSTEM VALUE VALUES (4, 'e202d5ab-db83-4aa3-ac97-8bf686eb5fdf', 'designer@ydsz.dev', '$2a$10$w9.G/KEkRbGKFJOpWNYqKuSL1YuUTjQ9wspxRpBia0jkEOWUjFLFC', '张设计', NULL, 't', 'Asia/Shanghai', '2026-08-08 00:01:22.071616+08', '2026-08-08 00:01:22.071616+08', NULL);
INSERT INTO "public"."users" OVERRIDING SYSTEM VALUE VALUES (5, 'dd0a9ebd-a996-4545-b04f-7b31865d5170', 'viewer@ydsz.dev', '$2a$10$mdDoBEbopCVd6HFWAZGpTOfMoQNhzKNqx6o6ZSC9e.KDxG2BKA.wi', '访客小赵', NULL, 't', 'Asia/Shanghai', '2026-08-08 00:01:22.133897+08', '2026-08-08 00:01:22.133897+08', NULL);

-- ----------------------------
-- Table structure for version_delivery_snapshots
-- ----------------------------
DROP TABLE IF EXISTS "public"."version_delivery_snapshots";
CREATE TABLE "public"."version_delivery_snapshots" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "version_id" int8 NOT NULL,
  "workspace_id" int8 NOT NULL,
  "progress" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "quality" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "release_notes" text COLLATE "pg_catalog"."default",
  "snapshot_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.version_delivery_snapshots IS '版本交付快照表（发布时生成的交付报告数据: 缺陷数/通过率/准出率等）';
COMMENT ON COLUMN public.version_delivery_snapshots.id IS '主键 ID';
COMMENT ON COLUMN public.version_delivery_snapshots.version_id IS '版本 FK';
COMMENT ON COLUMN public.version_delivery_snapshots.snapshot_date IS '快照时间（版本维度每日/里程碑聚合）';
COMMENT ON COLUMN public.version_delivery_snapshots.total_points IS '当日版本总计划故事点';
COMMENT ON COLUMN public.version_delivery_snapshots.done_points IS '当日版本已完成故事点';
COMMENT ON COLUMN public.version_delivery_snapshots.bug_count IS '缺陷总数（含未关闭）';
COMMENT ON COLUMN public.version_delivery_snapshots.open_bug_count IS '未关闭缺陷数';
COMMENT ON COLUMN public.version_delivery_snapshots.pass_rate IS '测试通过率百分比';
COMMENT ON COLUMN public.version_delivery_snapshots.deployment_count IS '累计部署次数（计数；与 DORA-DF 对齐）';
COMMENT ON COLUMN public.version_delivery_snapshots.metrics IS '详细效能指标 JSONB（逃逸率/返工率/各状态分布等）';
COMMENT ON COLUMN public.version_delivery_snapshots.created_at IS '写入时间';
COMMENT ON COLUMN public.version_delivery_snapshots.workspace_id IS '工作空间 FK（RLS 依据）';

-- ----------------------------
-- Records of version_delivery_snapshots
-- ----------------------------

-- ----------------------------
-- Table structure for versions
-- ----------------------------
DROP TABLE IF EXISTS "public"."versions";
CREATE TABLE "public"."versions" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "semver" text COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "status" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'planning'::text,
  "checklist" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "release_notes" text COLLATE "pg_catalog"."default",
  "delivered_at" timestamptz(6),
  "target_date" date,
  "archived_at" timestamptz(6),
  "created_by" int8 NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now(),
  "deleted_at" timestamptz(6),
  "start_date" date,
  "end_date" date,
  "version" int4 NOT NULL DEFAULT 1
)
;
COMMENT ON TABLE public.versions IS '版本发布表（SemVer 语义化版本；生命周期 planning→active→released→archived；聚合跨迭代进度）';
COMMENT ON COLUMN public.versions.id IS '主键 ID';
COMMENT ON COLUMN public.versions.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.versions.project_id IS '所属项目 FK';
COMMENT ON COLUMN public.versions.name IS '版本展示名（如"用户画像一期"）';
COMMENT ON COLUMN public.versions.semver IS '语义化版本号（如 1.2.3；项目内唯一；发布后只读）';
COMMENT ON COLUMN public.versions.description IS '版本目标与范围说明';
COMMENT ON COLUMN public.versions.status IS '版本状态: planning(规划中) / active(进行中) / released(已发布) / archived(已归档)';
COMMENT ON COLUMN public.versions.start_date IS '计划开始时间（可选）';
COMMENT ON COLUMN public.versions.end_date IS '计划结束时间（可选）';
COMMENT ON COLUMN public.versions.target_date IS '计划发布日（版本日；触发风险预警: 到期未发布/剩余缺陷）';
COMMENT ON COLUMN public.versions.checklist IS '发布检查清单 JSONB [{id,label,required,checked}]；全部勾选才可 release';
COMMENT ON COLUMN public.versions.release_notes IS 'Release Notes（发布时按模板三段式生成: 需求/缺陷修复/已知问题；可编辑）';
COMMENT ON COLUMN public.versions.delivered_at IS '实际发布 TIMESTAMPTZ（发布动作时写入）';
COMMENT ON COLUMN public.versions.archived_at IS '归档时间 TIMESTAMPTZ';
COMMENT ON COLUMN public.versions.delivery_report IS '交付报告 JSONB（缺陷数/通过率/准出率/迭代完成度明细；发布时生成）';
COMMENT ON COLUMN public.versions.progress IS '聚合进度 0-100（读时计算；缓存版本失效键 version:{id}:progress）';
COMMENT ON COLUMN public.versions.created_by IS '创建人 FK';
COMMENT ON COLUMN public.versions.created_at IS '创建时间';
COMMENT ON COLUMN public.versions.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.versions.deleted_at IS '软删除时间戳';
COMMENT ON COLUMN public.versions.version IS '乐观锁版本号（默认 1）';
COMMENT ON COLUMN "public"."versions"."version" IS '乐观锁版本号，每次 UPDATE 自增';

-- ----------------------------
-- Records of versions
-- ----------------------------

-- ----------------------------
-- Table structure for view_preferences
-- ----------------------------
DROP TABLE IF EXISTS "public"."view_preferences";
CREATE TABLE "public"."view_preferences" (
  "id" int8 NOT NULL DEFAULT nextval('view_preferences_id_seq'::regclass),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "view_type" varchar(20) COLLATE "pg_catalog"."default" NOT NULL,
  "layout" varchar(20) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'list'::character varying,
  "columns" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "filters" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "sort" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "extra" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.view_preferences IS '视图偏好表（按项目隔离；排序/过滤/分组/字段裁剪配置 upsert）';
COMMENT ON COLUMN public.view_preferences.id IS '主键 ID';
COMMENT ON COLUMN public.view_preferences.user_id IS '用户 FK';
COMMENT ON COLUMN public.view_preferences.workspace_id IS '工作空间 FK（RLS 依据）';
COMMENT ON COLUMN public.view_preferences.project_id IS '项目 FK（NULL=工作空间级默认视图偏好）';
COMMENT ON COLUMN public.view_preferences.view_type IS '视图类型: list / board / calendar / gantt';
COMMENT ON COLUMN public.view_preferences.preferences IS '偏好 JSONB {sort_by, sort_order, filters, group_by, field_visibility}；upsert key (user, project, view)';
COMMENT ON COLUMN public.view_preferences.created_at IS '创建时间';
COMMENT ON COLUMN public.view_preferences.updated_at IS '修改时间（触发器自动维护）';

-- ----------------------------
-- Records of view_preferences
-- ----------------------------

-- ----------------------------
-- Table structure for webhook_logs
-- ----------------------------
DROP TABLE IF EXISTS "public"."webhook_logs";
CREATE TABLE "public"."webhook_logs" (
  "id" int8 NOT NULL DEFAULT nextval('webhook_logs_id_seq'::regclass),
  "webhook_id" int8 NOT NULL,
  "workspace_id" int8 NOT NULL,
  "delivery_id" varchar(64) COLLATE "pg_catalog"."default" NOT NULL,
  "event_type" varchar(80) COLLATE "pg_catalog"."default" NOT NULL,
  "event_id" int8,
  "request_url" text COLLATE "pg_catalog"."default" NOT NULL,
  "request_method" varchar(10) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'POST'::character varying,
  "request_headers" jsonb,
  "request_body" text COLLATE "pg_catalog"."default",
  "response_status" int4,
  "response_body" text COLLATE "pg_catalog"."default",
  "response_headers" jsonb,
  "status" varchar(20) COLLATE "pg_catalog"."default" NOT NULL,
  "attempt" int2 NOT NULL DEFAULT 1,
  "duration_ms" int4,
  "error" text COLLATE "pg_catalog"."default",
  "occurred_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.webhook_logs IS 'Webhook 投递日志表（按月分区 + 30 天 TTL；request/response/耗时；手动重投）';
COMMENT ON COLUMN public.webhook_logs.id IS '主键 ID';
COMMENT ON COLUMN public.webhook_logs.webhook_id IS 'Webhook 配置 FK';
COMMENT ON COLUMN public.webhook_logs.event_type IS '事件类型（与 domain_events.event_type 对齐）';
COMMENT ON COLUMN public.webhook_logs.delivery_id IS '投递唯一 UUID（X-Ydsz-Delivery 头；接收方幂等）';
COMMENT ON COLUMN public.webhook_logs.target_url IS '投递目标（当时快照；配置修改后仍展示原值）';
COMMENT ON COLUMN public.webhook_logs.request_body IS 'POST body（JSON；截断 >10KB 时存 attachment）';
COMMENT ON COLUMN public.webhook_logs.response_status IS '接收方 HTTP status code；5xx/429/超时(10s) 触发重试';
COMMENT ON COLUMN public.webhook_logs.response_body IS '接收方响应体（截断 >1KB）；便于排障';
COMMENT ON COLUMN public.webhook_logs.duration_ms IS '投递耗时毫秒';
COMMENT ON COLUMN public.webhook_logs.attempt_number IS '本次重试次数（1=首次；最大 3 次）';
COMMENT ON COLUMN public.webhook_logs.status IS '投递状态: success / failed / retrying';
COMMENT ON COLUMN public.webhook_logs.error_message IS '错误信息（失败时）';
COMMENT ON COLUMN public.webhook_logs.created_at IS '投递时间 TIMESTAMPTZ（按月分区 + 30 天 TTL）';
COMMENT ON COLUMN public.webhook_logs.workspace_id IS '工作空间 FK（RLS 依据）';

-- ----------------------------
-- Records of webhook_logs
-- ----------------------------

-- ----------------------------
-- Table structure for webhooks
-- ----------------------------
DROP TABLE IF EXISTS "public"."webhooks";
CREATE TABLE "public"."webhooks" (
  "id" int8 NOT NULL DEFAULT nextval('webhooks_id_seq'::regclass),
  "workspace_id" int8 NOT NULL,
  "project_id" int8,
  "name" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "target_url" text COLLATE "pg_catalog"."default" NOT NULL,
  "secret" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "events" text[] COLLATE "pg_catalog"."default" NOT NULL DEFAULT '{}'::text[],
  "is_active" bool NOT NULL DEFAULT true,
  "last_error" text COLLATE "pg_catalog"."default",
  "last_triggered" timestamptz(6),
  "last_status" varchar(20) COLLATE "pg_catalog"."default",
  "unhealthy_at" timestamptz(6),
  "created_by" int8 NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.webhooks IS 'Webhook 配置表（target_url + secret HMAC + 事件白名单 + SSRF 防护）';
COMMENT ON COLUMN public.webhooks.id IS '主键 ID';
COMMENT ON COLUMN public.webhooks.workspace_id IS '工作空间 FK（RLS 依据）';
COMMENT ON COLUMN public.webhooks.project_id IS '项目 FK（决定事件触发的运行上下文）';
COMMENT ON COLUMN public.webhooks.name IS 'Webhook 配置名称';
COMMENT ON COLUMN public.webhooks.target_url IS '投递目标 URL（SSRF 防护: 协议白名单 https 优先；解析后 IP 非内网）';
COMMENT ON COLUMN public.webhooks.secret IS 'HMAC-SHA256 密钥（X-Ydsz-Signature-256 签名头）；仅存 SHA-256 hash';
COMMENT ON COLUMN public.webhooks.events IS '事件白名单 JSONB ["issue.created", "issue.status_changed", ...]；空数组=全部禁用';
COMMENT ON COLUMN public.webhooks.is_active IS '启用开关: true=生效, false=暂停';
COMMENT ON COLUMN public.webhooks.ssrf_whitelist IS '出站 IP 白名单（空=使用默认安全防护；显式声明例外 IP/CIDR）';
COMMENT ON COLUMN public.webhooks.last_triggered_at IS '最后触发时间（监控活跃度；不活跃可告警）';
COMMENT ON COLUMN public.webhooks.last_error IS '最后错误信息（用于排障；连续失败通知创建者）';
COMMENT ON COLUMN public.webhooks.failure_count IS '连续失败次数；>=5 自动 unhealthy + 通知 admin';
COMMENT ON COLUMN public.webhooks.created_by IS '创建人 FK（Webhook 失败默认通知人）';
COMMENT ON COLUMN public.webhooks.created_at IS '创建时间';
COMMENT ON COLUMN public.webhooks.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.webhooks.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of webhooks
-- ----------------------------

-- ----------------------------
-- Table structure for workbench_configs
-- ----------------------------
DROP TABLE IF EXISTS "public"."workbench_configs";
CREATE TABLE "public"."workbench_configs" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8,
  "user_id" int8 NOT NULL,
  "layout" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "widget_states" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "focus_enabled" bool NOT NULL DEFAULT false,
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.workbench_configs IS '个人工作台配置表（置顶项目/收藏视图/自定义组件布局；per-user）';
COMMENT ON COLUMN public.workbench_configs.id IS '主键 ID';
COMMENT ON COLUMN public.workbench_configs.user_id IS '用户 FK（per-user 工作台；每个用户一行）';
COMMENT ON COLUMN public.workbench_configs.workspace_id IS '工作空间 FK（RLS 依据）';
COMMENT ON COLUMN public.workbench_configs.layout IS '工作台布局 JSONB [{widget_type, x, y, w, h, config, is_pinned}]';
COMMENT ON COLUMN public.workbench_configs.pinned_projects IS '置顶项目 ID 数组 BIGINT[]；快速访问入口';
COMMENT ON COLUMN public.workbench_configs.recent_views IS '最近使用视图 ID 数组（用于视图选择器默认值）';
COMMENT ON COLUMN public.workbench_configs.preferences IS '工作台偏好 JSONB（主题/快捷手势/默认看板）';
COMMENT ON COLUMN public.workbench_configs.created_at IS '创建时间';
COMMENT ON COLUMN public.workbench_configs.updated_at IS '修改时间（触发器自动维护）';

-- ----------------------------
-- Records of workbench_configs
-- ----------------------------

-- ----------------------------
-- Table structure for workbench_templates
-- ----------------------------
DROP TABLE IF EXISTS "public"."workbench_templates";
CREATE TABLE "public"."workbench_templates" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "slug" text COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "layout" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "icon" text COLLATE "pg_catalog"."default",
  "is_default" bool NOT NULL DEFAULT false,
  "sort_order" int4 NOT NULL DEFAULT 0,
  "created_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.workbench_templates IS '工作台预设模板表（供创建账号时初始化工作台）';
COMMENT ON COLUMN public.workbench_templates.id IS '主键 ID';
COMMENT ON COLUMN public.workbench_templates.name IS '模板名称（如"PM 视角"/"开发者视角"/"QA 视角"）';
COMMENT ON COLUMN public.workbench_templates.description IS '模板适用角色说明';
COMMENT ON COLUMN public.workbench_templates.layout IS '预设布局 JSONB（widgets 数组；新账号注册时复制为默认配置）';
COMMENT ON COLUMN public.workbench_templates.is_system IS '是否内置模板';
COMMENT ON COLUMN public.workbench_templates.sort_order IS '排序权重';
COMMENT ON COLUMN public.workbench_templates.created_at IS '创建时间';
COMMENT ON COLUMN public.workbench_templates.updated_at IS '修改时间';

-- ----------------------------
-- Records of workbench_templates
-- ----------------------------
INSERT INTO "public"."workbench_templates" OVERRIDING SYSTEM VALUE VALUES (1, '敏捷开发', 'agile', 'Scrum/Kanban 团队的默认工作台，含迭代概览、待办事项、燃尽任务、阻塞工作项', '{"widgets": [{"h": 4, "w": 6, "type": "my_issues"}, {"h": 3, "w": 6, "type": "sprint_overview"}, {"h": 2, "w": 4, "type": "overdue"}, {"h": 3, "w": 4, "type": "recent"}, {"h": 2, "w": 4, "type": "quick_actions"}]}', NULL, 't', 1, '2026-08-08 00:01:05.060295+08');
INSERT INTO "public"."workbench_templates" OVERRIDING SYSTEM VALUE VALUES (2, '项目监控', 'pmo', 'PMO/管理者视角，关注项目进度、逾期预警、团队速率', '{"widgets": [{"h": 2, "w": 12, "type": "project_overview"}, {"h": 3, "w": 6, "type": "overdue"}, {"h": 3, "w": 6, "type": "risk_alert"}, {"h": 3, "w": 6, "type": "recent"}, {"h": 2, "w": 6, "type": "quick_actions"}]}', NULL, 'f', 2, '2026-08-08 00:01:05.060295+08');
INSERT INTO "public"."workbench_templates" OVERRIDING SYSTEM VALUE VALUES (3, '个人聚焦', 'focus', '专注模式，只有待办和专注计时器', '{"widgets": [{"h": 6, "w": 12, "type": "my_issues"}, {"h": 2, "w": 6, "type": "focus_timer"}, {"h": 2, "w": 6, "type": "overdue"}]}', NULL, 'f', 3, '2026-08-08 00:01:05.060295+08');

-- ----------------------------
-- Table structure for workspace_members
-- ----------------------------
DROP TABLE IF EXISTS "public"."workspace_members";
CREATE TABLE "public"."workspace_members" (
  "workspace_id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "role" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'member'::text,
  "joined_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.workspace_members IS '工作空间成员表（user↔workspace 多对一，含 owner/admin/member/guest 角色）';
COMMENT ON COLUMN public.workspace_members.id IS '主键 ID';
COMMENT ON COLUMN public.workspace_members.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.workspace_members.user_id IS '用户 FK';
COMMENT ON COLUMN public.workspace_members.role IS '工作空间级角色: owner / admin / member / guest；用于权限点收敛';
COMMENT ON COLUMN public.workspace_members.is_active IS '成员状态: true=激活, false=暂停（不立即离职，恢复使用）';
COMMENT ON COLUMN public.workspace_members.joined_at IS '加入时间（接受邀请/被添加时）';
COMMENT ON COLUMN public.workspace_members.created_by IS '添加人 FK';
COMMENT ON COLUMN public.workspace_members.created_at IS '创建时间';
COMMENT ON COLUMN public.workspace_members.updated_at IS '修改时间（触发器自动维护）';

-- ----------------------------
-- Records of workspace_members
-- ----------------------------
INSERT INTO "public"."workspace_members" VALUES (1, 3, 'member', '2026-07-09 00:01:22.102629+08');
INSERT INTO "public"."workspace_members" VALUES (1, 4, 'member', '2026-07-09 00:01:22.106294+08');
INSERT INTO "public"."workspace_members" VALUES (1, 5, 'guest', '2026-07-09 00:01:22.108506+08');
INSERT INTO "public"."workspace_members" VALUES (1, 1, 'owner', '2026-07-09 00:01:22.110778+08');
INSERT INTO "public"."workspace_members" VALUES (1, 2, 'admin', '2026-07-09 00:01:22.113025+08');
INSERT INTO "public"."workspace_members" VALUES (2, 1, 'admin', '2026-07-09 00:01:22.114869+08');
INSERT INTO "public"."workspace_members" VALUES (2, 4, 'owner', '2026-07-09 00:01:22.11625+08');
INSERT INTO "public"."workspace_members" VALUES (2, 2, 'member', '2026-07-09 00:01:22.118427+08');
INSERT INTO "public"."workspace_members" VALUES (3, 1, 'owner', '2026-07-09 00:01:22.121457+08');
INSERT INTO "public"."workspace_members" VALUES (3, 3, 'admin', '2026-07-09 00:01:22.127884+08');

-- ----------------------------
-- Table structure for project_members
-- ----------------------------
DROP TABLE IF EXISTS "public"."project_members";
CREATE TABLE "public"."project_members" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "role" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'member'::text,
  "joined_at" timestamptz(6) NOT NULL DEFAULT now(),
  "created_by" int8,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.project_members IS '项目成员表（user↔project 多对多，含 admin/member 角色；workspace_members 的子集）';
COMMENT ON COLUMN public.project_members.id IS '主键 ID';
COMMENT ON COLUMN public.project_members.workspace_id IS '工作空间 FK';
COMMENT ON COLUMN public.project_members.project_id IS '项目 FK';
COMMENT ON COLUMN public.project_members.user_id IS '用户 FK';
COMMENT ON COLUMN public.project_members.role IS '项目级角色: admin / member；admin 可管理项目成员与设置';
COMMENT ON COLUMN public.project_members.joined_at IS '加入时间';
COMMENT ON COLUMN public.project_members.created_by IS '添加人 FK';
COMMENT ON COLUMN public.project_members.created_at IS '创建时间';
COMMENT ON COLUMN public.project_members.updated_at IS '修改时间（触发器自动维护）';

-- ----------------------------
-- Table structure for workspaces
-- ----------------------------
DROP TABLE IF EXISTS "public"."workspaces";
CREATE TABLE "public"."workspaces" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "slug" text COLLATE "pg_catalog"."default" NOT NULL,
  "logo_url" text COLLATE "pg_catalog"."default",
  "timezone" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'Asia/Shanghai'::text,
  "language" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'zh-CN'::text,
  "owner_id" int8 NOT NULL,
  "status" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'active'::text,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now()
)
;
COMMENT ON TABLE public.workspaces IS '工作空间表（多租户顶层容器，RLS 依据、通知配置、品牌化、SSO）';
COMMENT ON COLUMN public.workspaces.id IS '主键 ID';
COMMENT ON COLUMN public.workspaces.name IS '工作空间名称（唯一标识租户）';
COMMENT ON COLUMN public.workspaces.slug IS 'URL 友好唯一标识（小写 + 连字符；用于子域名/API 路由）';
COMMENT ON COLUMN public.workspaces.description IS '工作空间简介（可选）';
COMMENT ON COLUMN public.workspaces.logo IS '品牌 Logo 附件 ID';
COMMENT ON COLUMN public.workspaces.brand_colors IS '品牌色 JSONB {primary, secondary, accent} HEX 值';
COMMENT ON COLUMN public.workspaces.default_role IS '邀请新用户默认角色: member / guest';
COMMENT ON COLUMN public.workspaces.is_active IS '工作空间激活状态: true=正常, false=暂停（SSO/限流配置）';
COMMENT ON COLUMN public.workspaces.settings IS '工作空间设置 JSONB（SSO/安全策略/通知通道/附件配置）';
COMMENT ON COLUMN public.workspaces.created_by IS '创建人 FK（owner；默认拥有 owner 角色）';
COMMENT ON COLUMN public.workspaces.created_at IS '创建时间';
COMMENT ON COLUMN public.workspaces.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.workspaces.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of workspaces
-- ----------------------------
INSERT INTO "public"."workspaces" OVERRIDING SYSTEM VALUE VALUES (1, '核心产品', 'core', NULL, 'Asia/Shanghai', 'zh-CN', 1, 'active', '2026-08-08 00:01:22.137523+08', '2026-08-08 00:01:22.137523+08');
INSERT INTO "public"."workspaces" OVERRIDING SYSTEM VALUE VALUES (2, '设计系统', 'design-system', NULL, 'Asia/Shanghai', 'zh-CN', 1, 'active', '2026-08-08 00:01:22.140236+08', '2026-08-08 00:01:22.140236+08');
INSERT INTO "public"."workspaces" OVERRIDING SYSTEM VALUE VALUES (3, '基础设施', 'infra', NULL, 'Asia/Shanghai', 'zh-CN', 1, 'active', '2026-08-08 00:01:22.14201+08', '2026-08-08 00:01:22.14201+08');

-- ----------------------------
-- Function structure for armor
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."armor"(bytea, _text, _text);
CREATE FUNCTION "public"."armor"(bytea, _text, _text)
  RETURNS "pg_catalog"."text" AS '$libdir/pgcrypto', 'pg_armor'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for armor
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."armor"(bytea);
CREATE FUNCTION "public"."armor"(bytea)
  RETURNS "pg_catalog"."text" AS '$libdir/pgcrypto', 'pg_armor'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for bump_version
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."bump_version"();
CREATE FUNCTION "public"."bump_version"()
  RETURNS "pg_catalog"."trigger" AS $BODY$
BEGIN
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;

-- ----------------------------
-- Function structure for crypt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."crypt"(text, text);
CREATE FUNCTION "public"."crypt"(text, text)
  RETURNS "pg_catalog"."text" AS '$libdir/pgcrypto', 'pg_crypt'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for dearmor
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."dearmor"(text);
CREATE FUNCTION "public"."dearmor"(text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pg_dearmor'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for decrypt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."decrypt"(bytea, bytea, text);
CREATE FUNCTION "public"."decrypt"(bytea, bytea, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pg_decrypt'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for decrypt_iv
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."decrypt_iv"(bytea, bytea, bytea, text);
CREATE FUNCTION "public"."decrypt_iv"(bytea, bytea, bytea, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pg_decrypt_iv'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for digest
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."digest"(text, text);
CREATE FUNCTION "public"."digest"(text, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pg_digest'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for digest
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."digest"(bytea, text);
CREATE FUNCTION "public"."digest"(bytea, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pg_digest'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for encrypt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."encrypt"(bytea, bytea, text);
CREATE FUNCTION "public"."encrypt"(bytea, bytea, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pg_encrypt'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for encrypt_iv
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."encrypt_iv"(bytea, bytea, bytea, text);
CREATE FUNCTION "public"."encrypt_iv"(bytea, bytea, bytea, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pg_encrypt_iv'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for fips_mode
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."fips_mode"();
CREATE FUNCTION "public"."fips_mode"()
  RETURNS "pg_catalog"."bool" AS '$libdir/pgcrypto', 'pg_check_fipsmode'
  LANGUAGE c VOLATILE STRICT
  COST 1;

-- ----------------------------
-- Function structure for fn_cleanup_search_document
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."fn_cleanup_search_document"();
CREATE FUNCTION "public"."fn_cleanup_search_document"()
  RETURNS "pg_catalog"."trigger" AS $BODY$
DECLARE
    v_doc_type TEXT;
BEGIN
    CASE TG_TABLE_NAME
        WHEN 'issues'   THEN v_doc_type := 'issue';
        WHEN 'sprints'  THEN v_doc_type := 'sprint';
        WHEN 'versions' THEN v_doc_type := 'version';
        ELSE RETURN OLD;
    END CASE;
    DELETE FROM search_documents
    WHERE doc_type = v_doc_type AND doc_id = OLD.id AND workspace_id = OLD.workspace_id;
    RETURN OLD;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;

-- ----------------------------
-- Function structure for fn_refresh_search_document
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."fn_refresh_search_document"();
CREATE FUNCTION "public"."fn_refresh_search_document"()
  RETURNS "pg_catalog"."trigger" AS $BODY$
DECLARE
    v_title TEXT;
    v_content TEXT;
    v_metadata JSONB;
BEGIN
    v_title := COALESCE(NEW.name, '');
    v_content := COALESCE(NEW.description_stripped, '');
    v_metadata := jsonb_build_object(
        'type_code', NEW.type_code,
        'state_id', NEW.state_id,
        'priority', NEW.priority
    );

    INSERT INTO search_documents (workspace_id, project_id, doc_type, doc_id, title, identifier, content, search_tsv, metadata)
    VALUES (
        NEW.workspace_id, NEW.project_id, 'issue', NEW.id,
        v_title, NEW.sequence_id::text, v_content,
        to_tsvector('simple',
            coalesce(v_title, '') || ' ' ||
            coalesce(v_content, '')
        ),
        v_metadata
    )
    ON CONFLICT (workspace_id, doc_type, doc_id) DO UPDATE SET
        title = EXCLUDED.title,
        identifier = EXCLUDED.identifier,
        content = EXCLUDED.content,
        search_tsv = EXCLUDED.search_tsv,
        metadata = EXCLUDED.metadata,
        updated_at = now();
    RETURN NEW;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;

-- ----------------------------
-- Function structure for fn_refresh_sprint_search_document
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."fn_refresh_sprint_search_document"();
CREATE FUNCTION "public"."fn_refresh_sprint_search_document"()
  RETURNS "pg_catalog"."trigger" AS $BODY$
BEGIN
    INSERT INTO search_documents (workspace_id, project_id, doc_type, doc_id, title, identifier, content, search_tsv, metadata)
    VALUES (
        NEW.workspace_id, NEW.project_id, 'sprint', NEW.id,
        COALESCE(NEW.name, ''),
        NULL,
        COALESCE(NEW.goal, ''),
        to_tsvector('simple',
            coalesce(NEW.name, '') || ' ' ||
            coalesce(NEW.goal, '')
        ),
        jsonb_build_object('status', NEW.status)
    )
    ON CONFLICT (workspace_id, doc_type, doc_id) DO UPDATE SET
        title = EXCLUDED.title,
        identifier = EXCLUDED.identifier,
        content = EXCLUDED.content,
        search_tsv = EXCLUDED.search_tsv,
        metadata = EXCLUDED.metadata,
        updated_at = now();
    RETURN NEW;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;

-- ----------------------------
-- Function structure for fn_refresh_version_search_document
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."fn_refresh_version_search_document"();
CREATE FUNCTION "public"."fn_refresh_version_search_document"()
  RETURNS "pg_catalog"."trigger" AS $BODY$
BEGIN
    INSERT INTO search_documents (workspace_id, project_id, doc_type, doc_id, title, identifier, content, search_tsv, metadata)
    VALUES (
        NEW.workspace_id, NEW.project_id, 'version', NEW.id,
        COALESCE(NEW.name, ''),
        NEW.semver,
        COALESCE(NEW.description, ''),
        to_tsvector('simple',
            coalesce(NEW.name, '') || ' ' ||
            coalesce(NEW.description, '')
        ),
        jsonb_build_object('status', NEW.status)
    )
    ON CONFLICT (workspace_id, doc_type, doc_id) DO UPDATE SET
        title = EXCLUDED.title,
        identifier = EXCLUDED.identifier,
        content = EXCLUDED.content,
        search_tsv = EXCLUDED.search_tsv,
        metadata = EXCLUDED.metadata,
        updated_at = now();
    RETURN NEW;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;

-- ----------------------------
-- Function structure for fn_touch_recent_item
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."fn_touch_recent_item"();
CREATE FUNCTION "public"."fn_touch_recent_item"()
  RETURNS "pg_catalog"."trigger" AS $BODY$
BEGIN
    NEW.accessed_at := now();
    RETURN NEW;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;

-- ----------------------------
-- Function structure for gen_random_bytes
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."gen_random_bytes"(int4);
CREATE FUNCTION "public"."gen_random_bytes"(int4)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pg_random_bytes'
  LANGUAGE c VOLATILE STRICT
  COST 1;

-- ----------------------------
-- Function structure for gen_random_uuid
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."gen_random_uuid"();
CREATE FUNCTION "public"."gen_random_uuid"()
  RETURNS "pg_catalog"."uuid" AS '$libdir/pgcrypto', 'pg_random_uuid'
  LANGUAGE c VOLATILE
  COST 1;

-- ----------------------------
-- Function structure for gen_salt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."gen_salt"(text);
CREATE FUNCTION "public"."gen_salt"(text)
  RETURNS "pg_catalog"."text" AS '$libdir/pgcrypto', 'pg_gen_salt'
  LANGUAGE c VOLATILE STRICT
  COST 1;

-- ----------------------------
-- Function structure for gen_salt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."gen_salt"(text, int4);
CREATE FUNCTION "public"."gen_salt"(text, int4)
  RETURNS "pg_catalog"."text" AS '$libdir/pgcrypto', 'pg_gen_salt_rounds'
  LANGUAGE c VOLATILE STRICT
  COST 1;

-- ----------------------------
-- Function structure for hmac
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."hmac"(text, text, text);
CREATE FUNCTION "public"."hmac"(text, text, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pg_hmac'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for hmac
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."hmac"(bytea, bytea, text);
CREATE FUNCTION "public"."hmac"(bytea, bytea, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pg_hmac'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_armor_headers
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_armor_headers"(text, OUT "key" text, OUT "value" text);
CREATE FUNCTION "public"."pgp_armor_headers"(IN text, OUT "key" text, OUT "value" text)
  RETURNS SETOF "pg_catalog"."record" AS '$libdir/pgcrypto', 'pgp_armor_headers'
  LANGUAGE c IMMUTABLE STRICT
  COST 1
  ROWS 1000;

-- ----------------------------
-- Function structure for pgp_key_id
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_key_id"(bytea);
CREATE FUNCTION "public"."pgp_key_id"(bytea)
  RETURNS "pg_catalog"."text" AS '$libdir/pgcrypto', 'pgp_key_id_w'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_pub_decrypt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_pub_decrypt"(bytea, bytea, text, text);
CREATE FUNCTION "public"."pgp_pub_decrypt"(bytea, bytea, text, text)
  RETURNS "pg_catalog"."text" AS '$libdir/pgcrypto', 'pgp_pub_decrypt_text'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_pub_decrypt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_pub_decrypt"(bytea, bytea);
CREATE FUNCTION "public"."pgp_pub_decrypt"(bytea, bytea)
  RETURNS "pg_catalog"."text" AS '$libdir/pgcrypto', 'pgp_pub_decrypt_text'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_pub_decrypt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_pub_decrypt"(bytea, bytea, text);
CREATE FUNCTION "public"."pgp_pub_decrypt"(bytea, bytea, text)
  RETURNS "pg_catalog"."text" AS '$libdir/pgcrypto', 'pgp_pub_decrypt_text'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_pub_decrypt_bytea
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_pub_decrypt_bytea"(bytea, bytea);
CREATE FUNCTION "public"."pgp_pub_decrypt_bytea"(bytea, bytea)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pgp_pub_decrypt_bytea'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_pub_decrypt_bytea
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_pub_decrypt_bytea"(bytea, bytea, text, text);
CREATE FUNCTION "public"."pgp_pub_decrypt_bytea"(bytea, bytea, text, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pgp_pub_decrypt_bytea'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_pub_decrypt_bytea
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_pub_decrypt_bytea"(bytea, bytea, text);
CREATE FUNCTION "public"."pgp_pub_decrypt_bytea"(bytea, bytea, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pgp_pub_decrypt_bytea'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_pub_encrypt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_pub_encrypt"(text, bytea);
CREATE FUNCTION "public"."pgp_pub_encrypt"(text, bytea)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pgp_pub_encrypt_text'
  LANGUAGE c VOLATILE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_pub_encrypt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_pub_encrypt"(text, bytea, text);
CREATE FUNCTION "public"."pgp_pub_encrypt"(text, bytea, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pgp_pub_encrypt_text'
  LANGUAGE c VOLATILE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_pub_encrypt_bytea
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_pub_encrypt_bytea"(bytea, bytea);
CREATE FUNCTION "public"."pgp_pub_encrypt_bytea"(bytea, bytea)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pgp_pub_encrypt_bytea'
  LANGUAGE c VOLATILE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_pub_encrypt_bytea
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_pub_encrypt_bytea"(bytea, bytea, text);
CREATE FUNCTION "public"."pgp_pub_encrypt_bytea"(bytea, bytea, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pgp_pub_encrypt_bytea'
  LANGUAGE c VOLATILE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_sym_decrypt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_sym_decrypt"(bytea, text, text);
CREATE FUNCTION "public"."pgp_sym_decrypt"(bytea, text, text)
  RETURNS "pg_catalog"."text" AS '$libdir/pgcrypto', 'pgp_sym_decrypt_text'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_sym_decrypt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_sym_decrypt"(bytea, text);
CREATE FUNCTION "public"."pgp_sym_decrypt"(bytea, text)
  RETURNS "pg_catalog"."text" AS '$libdir/pgcrypto', 'pgp_sym_decrypt_text'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_sym_decrypt_bytea
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_sym_decrypt_bytea"(bytea, text, text);
CREATE FUNCTION "public"."pgp_sym_decrypt_bytea"(bytea, text, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pgp_sym_decrypt_bytea'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_sym_decrypt_bytea
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_sym_decrypt_bytea"(bytea, text);
CREATE FUNCTION "public"."pgp_sym_decrypt_bytea"(bytea, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pgp_sym_decrypt_bytea'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_sym_encrypt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_sym_encrypt"(text, text, text);
CREATE FUNCTION "public"."pgp_sym_encrypt"(text, text, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pgp_sym_encrypt_text'
  LANGUAGE c VOLATILE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_sym_encrypt
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_sym_encrypt"(text, text);
CREATE FUNCTION "public"."pgp_sym_encrypt"(text, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pgp_sym_encrypt_text'
  LANGUAGE c VOLATILE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_sym_encrypt_bytea
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_sym_encrypt_bytea"(bytea, text, text);
CREATE FUNCTION "public"."pgp_sym_encrypt_bytea"(bytea, text, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pgp_sym_encrypt_bytea'
  LANGUAGE c VOLATILE STRICT
  COST 1;

-- ----------------------------
-- Function structure for pgp_sym_encrypt_bytea
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."pgp_sym_encrypt_bytea"(bytea, text);
CREATE FUNCTION "public"."pgp_sym_encrypt_bytea"(bytea, text)
  RETURNS "pg_catalog"."bytea" AS '$libdir/pgcrypto', 'pgp_sym_encrypt_bytea'
  LANGUAGE c VOLATILE STRICT
  COST 1;

-- ----------------------------
-- Function structure for set_updated_at
-- ----------------------------
DROP FUNCTION IF EXISTS "public"."set_updated_at"();
CREATE FUNCTION "public"."set_updated_at"()
  RETURNS "pg_catalog"."trigger" AS $BODY$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$BODY$
  LANGUAGE plpgsql VOLATILE
  COST 100;

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."api_tokens_id_seq"
OWNED BY "public"."api_tokens"."id";
SELECT setval(pg_get_serial_sequence('public.api_tokens', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.api_tokens) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."attachments_id_seq"
OWNED BY "public"."attachments"."id";
SELECT setval(pg_get_serial_sequence('public.attachments', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.attachments) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."audit_logs_id_seq"
OWNED BY "public"."audit_logs"."id";
SELECT setval(pg_get_serial_sequence('public.audit_logs', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.audit_logs) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."automation_rules_id_seq"
OWNED BY "public"."automation_rules"."id";
SELECT setval(pg_get_serial_sequence('public.automation_rules', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.automation_rules) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."automation_templates_id_seq"
OWNED BY "public"."automation_templates"."id";
SELECT setval(pg_get_serial_sequence('public.automation_templates', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.automation_templates) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."dashboard_snapshots_id_seq"
OWNED BY "public"."dashboard_snapshots"."id";
SELECT setval(pg_get_serial_sequence('public.dashboard_snapshots', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.dashboard_snapshots) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."dashboard_templates_id_seq"
OWNED BY "public"."dashboard_templates"."id";
SELECT setval(pg_get_serial_sequence('public.dashboard_templates', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.dashboard_templates) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."dashboard_widgets_id_seq"
OWNED BY "public"."dashboard_widgets"."id";
SELECT setval(pg_get_serial_sequence('public.dashboard_widgets', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.dashboard_widgets) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."deployment_events_id_seq"
OWNED BY "public"."deployment_events"."id";
SELECT setval(pg_get_serial_sequence('public.deployment_events', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.deployment_events) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."domain_events_id_seq"
OWNED BY "public"."domain_events"."id";
SELECT setval(pg_get_serial_sequence('public.domain_events', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.domain_events) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."intake_channels_id_seq"
OWNED BY "public"."intake_channels"."id";
SELECT setval(pg_get_serial_sequence('public.intake_channels', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.intake_channels) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."intake_issues_id_seq"
OWNED BY "public"."intake_issues"."id";
SELECT setval(pg_get_serial_sequence('public.intake_issues', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.intake_issues) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."invitations_id_seq"
OWNED BY "public"."invitations"."id";
SELECT setval(pg_get_serial_sequence('public.invitations', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.invitations) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."issue_activities_id_seq"
OWNED BY "public"."issue_activities"."id";
SELECT setval(pg_get_serial_sequence('public.issue_activities', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.issue_activities) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."issue_comments_id_seq"
OWNED BY "public"."issue_comments"."id";
SELECT setval(pg_get_serial_sequence('public.issue_comments', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.issue_comments) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."issue_dependencies_id_seq"
OWNED BY "public"."issue_dependencies"."id";
SELECT setval(pg_get_serial_sequence('public.issue_dependencies', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.issue_dependencies) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."issue_relations_id_seq"
OWNED BY "public"."issue_relations"."id";
SELECT setval(pg_get_serial_sequence('public.issue_relations', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.issue_relations) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."issues_id_seq"
OWNED BY "public"."issues"."id";
SELECT setval(pg_get_serial_sequence('public.issues', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.issues) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."labels_id_seq"
OWNED BY "public"."labels"."id";
SELECT setval(pg_get_serial_sequence('public.labels', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.labels) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."metric_adjustments_id_seq"
OWNED BY "public"."metric_adjustments"."id";
SELECT setval(pg_get_serial_sequence('public.metric_adjustments', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.metric_adjustments) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."metric_snapshots_id_seq"
OWNED BY "public"."metric_snapshots"."id";
SELECT setval(pg_get_serial_sequence('public.metric_snapshots', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.metric_snapshots) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."modules_id_seq"
OWNED BY "public"."modules"."id";
SELECT setval(pg_get_serial_sequence('public.modules', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.modules) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."notification_deliveries_id_seq"
OWNED BY "public"."notification_deliveries"."id";
SELECT setval(pg_get_serial_sequence('public.notification_deliveries', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.notification_deliveries) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."notification_digests_id_seq"
OWNED BY "public"."notification_digests"."id";
SELECT setval(pg_get_serial_sequence('public.notification_digests', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.notification_digests) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."notification_preferences_id_seq"
OWNED BY "public"."notification_preferences"."id";
SELECT setval(pg_get_serial_sequence('public.notification_preferences', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.notification_preferences) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."notifications_id_seq"
OWNED BY "public"."notifications"."id";
SELECT setval(pg_get_serial_sequence('public.notifications', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.notifications) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."password_reset_tokens_id_seq"
OWNED BY "public"."password_reset_tokens"."id";
SELECT setval(pg_get_serial_sequence('public.password_reset_tokens', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.password_reset_tokens) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."projects_id_seq"
OWNED BY "public"."projects"."id";
SELECT setval(pg_get_serial_sequence('public.projects', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.projects) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."recent_items_id_seq"
OWNED BY "public"."recent_items"."id";
SELECT setval(pg_get_serial_sequence('public.recent_items', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.recent_items) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."risk_alerts_id_seq"
OWNED BY "public"."risk_alerts"."id";
SELECT setval(pg_get_serial_sequence('public.risk_alerts', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.risk_alerts) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."risk_rules_id_seq"
OWNED BY "public"."risk_rules"."id";
SELECT setval(pg_get_serial_sequence('public.risk_rules', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.risk_rules) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."rule_executions_id_seq"
OWNED BY "public"."rule_executions"."id";
SELECT setval(pg_get_serial_sequence('public.rule_executions', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.rule_executions) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."search_bookmarks_id_seq"
OWNED BY "public"."search_bookmarks"."id";
SELECT setval(pg_get_serial_sequence('public.search_bookmarks', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.search_bookmarks) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."search_documents_id_seq"
OWNED BY "public"."search_documents"."id";
SELECT setval(pg_get_serial_sequence('public.search_documents', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.search_documents) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."search_history_id_seq"
OWNED BY "public"."search_history"."id";
SELECT setval(pg_get_serial_sequence('public.search_history', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.search_history) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."sprint_snapshots_id_seq"
OWNED BY "public"."sprint_snapshots"."id";
SELECT setval(pg_get_serial_sequence('public.sprint_snapshots', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.sprint_snapshots) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."sprints_id_seq"
OWNED BY "public"."sprints"."id";
SELECT setval(pg_get_serial_sequence('public.sprints', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.sprints) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."state_transitions_id_seq"
OWNED BY "public"."state_transitions"."id";
SELECT setval(pg_get_serial_sequence('public.state_transitions', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.state_transitions) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."states_id_seq"
OWNED BY "public"."states"."id";
SELECT setval(pg_get_serial_sequence('public.states', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.states) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."time_logs_id_seq"
OWNED BY "public"."time_logs"."id";
SELECT setval(pg_get_serial_sequence('public.time_logs', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.time_logs) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."users_id_seq"
OWNED BY "public"."users"."id";
SELECT setval(pg_get_serial_sequence('public.users', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.users) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."version_delivery_snapshots_id_seq"
OWNED BY "public"."version_delivery_snapshots"."id";
SELECT setval(pg_get_serial_sequence('public.version_delivery_snapshots', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.version_delivery_snapshots) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."versions_id_seq"
OWNED BY "public"."versions"."id";
SELECT setval(pg_get_serial_sequence('public.versions', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.versions) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."view_preferences_id_seq"
OWNED BY "public"."view_preferences"."id";
SELECT setval(pg_get_serial_sequence('public.view_preferences', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.view_preferences) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."webhook_logs_id_seq"
OWNED BY "public"."webhook_logs"."id";
SELECT setval(pg_get_serial_sequence('public.webhook_logs', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.webhook_logs) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."webhooks_id_seq"
OWNED BY "public"."webhooks"."id";
SELECT setval(pg_get_serial_sequence('public.webhooks', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.webhooks) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."workbench_configs_id_seq"
OWNED BY "public"."workbench_configs"."id";
SELECT setval(pg_get_serial_sequence('public.workbench_configs', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.workbench_configs) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."workbench_templates_id_seq"
OWNED BY "public"."workbench_templates"."id";
SELECT setval(pg_get_serial_sequence('public.workbench_templates', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.workbench_templates) + 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."workspaces_id_seq"
OWNED BY "public"."workspaces"."id";
SELECT setval(pg_get_serial_sequence('public.workspaces', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.workspaces) + 1, false);

-- ----------------------------
-- Auto increment value for api_tokens
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.api_tokens', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.api_tokens) + 1, false);

-- ----------------------------
-- Indexes structure for table api_tokens
-- ----------------------------
CREATE INDEX "idx_api_tokens_user" ON "public"."api_tokens" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE revoked_at IS NULL;
COMMENT ON INDEX public.idx_api_tokens_user IS '按用户查拥有的 Token（Token 管理页/撤销确认）';

-- ----------------------------
-- Triggers structure for table api_tokens
-- ----------------------------
CREATE TRIGGER "trg_api_tokens_updated_at" BEFORE UPDATE ON "public"."api_tokens"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

-- ----------------------------
-- Uniques structure for table api_tokens
-- ----------------------------
ALTER TABLE "public"."api_tokens" ADD CONSTRAINT "api_tokens_token_hash_key" UNIQUE ("token_hash");

-- ----------------------------
-- Primary Key structure for table api_tokens
-- ----------------------------
ALTER TABLE "public"."api_tokens" ADD CONSTRAINT "api_tokens_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table attachments
-- ----------------------------
CREATE INDEX "idx_attachments_entity" ON "public"."attachments" USING btree (
  "entity_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "entity_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_attachments_entity IS '多态查询: type+id 定位附件列表（工作项详情/评论附件）';
CREATE INDEX "idx_attachments_uploader" ON "public"."attachments" USING btree (
  "uploaded_by" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_attachments_uploader IS '按上传人查附件（配额统计/回收站）';
CREATE INDEX "idx_attachments_workspace" ON "public"."attachments" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_attachments_workspace IS '按工作空间查全部附件（配额/存储用量统计）';

-- ----------------------------
-- Checks structure for table attachments
-- ----------------------------
ALTER TABLE "public"."attachments" ADD CONSTRAINT "attachments_entity_type_check" CHECK (entity_type::text = ANY (ARRAY['issue'::character varying, 'comment'::character varying, 'workspace'::character varying, 'project'::character varying]::text[]));

-- ----------------------------
-- Primary Key structure for table attachments
-- ----------------------------
ALTER TABLE "public"."attachments" ADD CONSTRAINT "attachments_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for audit_logs
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.audit_logs', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.audit_logs) + 1, false);

-- ----------------------------
-- Indexes structure for table audit_logs
-- ----------------------------
CREATE INDEX "idx_audit_logs_action_target" ON "public"."audit_logs" USING btree (
  "action" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "target" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE action ~~ 'version.%'::text;
COMMENT ON INDEX public.idx_audit_logs_action_target IS '按操作类型+目标对象查询审计日志';
CREATE INDEX "idx_audit_logs_ws_time" ON "public"."audit_logs" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_audit_logs_ws_time IS '按工作空间+时间倒序游标分页（安全审计查询）';

-- ----------------------------
-- Primary Key structure for table audit_logs
-- ----------------------------
ALTER TABLE "public"."audit_logs" ADD CONSTRAINT "audit_logs_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for automation_rules
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.automation_rules', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.automation_rules) + 1, false);

-- ----------------------------
-- Indexes structure for table automation_rules
-- ----------------------------
CREATE INDEX "idx_automation_rules_project_status" ON "public"."automation_rules" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE project_id IS NOT NULL;
COMMENT ON INDEX public.idx_automation_rules_project_status IS '按项目+状态查规则列表';
CREATE INDEX "idx_automation_rules_sort" ON "public"."automation_rules" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."int4_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_automation_rules_sort IS '按项目+排序权重查规则（执行优先级）';
CREATE INDEX "idx_automation_rules_trigger" ON "public"."automation_rules" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "trigger_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE status = 'active'::text;
COMMENT ON INDEX public.idx_automation_rules_trigger IS '按项目+trigger 类型快速匹配规则（事件分发索引）';
CREATE INDEX "idx_automation_rules_ws" ON "public"."automation_rules" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_automation_rules_ws IS '按工作空间查规则列表';

-- S11 性能优化: 活跃规则事件分发索引（覆盖索引消除回表）
CREATE INDEX CONCURRENTLY IF NOT EXISTS "idx_automation_rules_trigger_active"
ON "automation_rules" (project_id, trigger_type, sort_order)
WHERE status = 'active' AND project_id IS NOT NULL;
COMMENT ON INDEX public.idx_automation_rules_trigger_active IS 'S11: 活跃规则按 project+trigger_type+sort_order 覆盖索引（事件分发核心路径，消除回表）';

-- ----------------------------
-- Triggers structure for table automation_rules
-- ----------------------------
CREATE TRIGGER "trg_automation_rules_updated_at" BEFORE UPDATE ON "public"."automation_rules"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_automation_rules_updated_at IS 'automation_rules: BEFORE UPDATE 自动维护 updated_at + 触发条件缓存失效';

-- ----------------------------
-- Checks structure for table automation_rules
-- ----------------------------
ALTER TABLE "public"."automation_rules" ADD CONSTRAINT "automation_rules_status_check" CHECK (status = ANY (ARRAY['draft'::text, 'active'::text, 'disabled'::text, 'error'::text]));

-- ----------------------------
-- Primary Key structure for table automation_rules
-- ----------------------------
ALTER TABLE "public"."automation_rules" ADD CONSTRAINT "automation_rules_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for automation_templates
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.automation_templates', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.automation_templates) + 1, false);

-- ----------------------------
-- Indexes structure for table automation_templates
-- ----------------------------
CREATE INDEX "idx_automation_templates_cat" ON "public"."automation_templates" USING btree (
  "category" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."int4_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_automation_templates_cat IS '按分类查模板列表';

-- ----------------------------
-- Uniques structure for table automation_templates
-- ----------------------------
ALTER TABLE "public"."automation_templates" ADD CONSTRAINT "automation_templates_slug_key" UNIQUE ("slug");

-- ----------------------------
-- Primary Key structure for table automation_templates
-- ----------------------------
ALTER TABLE "public"."automation_templates" ADD CONSTRAINT "automation_templates_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for dashboard_snapshots
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.dashboard_snapshots', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.dashboard_snapshots) + 1, false);

-- ----------------------------
-- Indexes structure for table dashboard_snapshots
-- ----------------------------
CREATE INDEX "idx_dashboard_snapshots_project" ON "public"."dashboard_snapshots" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_dashboard_snapshots_project IS '按项目查快照列表';
CREATE INDEX "idx_dashboard_snapshots_refreshed" ON "public"."dashboard_snapshots" USING btree (
  "refreshed_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_dashboard_snapshots_refreshed IS '按刷新时间查询快照（调度器轮询）';

-- ----------------------------
-- Uniques structure for table dashboard_snapshots
-- ----------------------------
ALTER TABLE "public"."dashboard_snapshots" ADD CONSTRAINT "dashboard_snapshots_project_id_widget_type_key" UNIQUE ("project_id", "widget_type");

-- ----------------------------
-- Primary Key structure for table dashboard_snapshots
-- ----------------------------
ALTER TABLE "public"."dashboard_snapshots" ADD CONSTRAINT "dashboard_snapshots_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for dashboard_templates
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.dashboard_templates', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.dashboard_templates) + 1, false);

-- ----------------------------
-- Indexes structure for table dashboard_templates
-- ----------------------------
CREATE INDEX "idx_dashboard_templates_category" ON "public"."dashboard_templates" USING btree (
  "category" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."int4_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_dashboard_templates_category IS '按分类查仪表盘模板列表';

-- ----------------------------
-- Uniques structure for table dashboard_templates
-- ----------------------------
ALTER TABLE "public"."dashboard_templates" ADD CONSTRAINT "dashboard_templates_slug_key" UNIQUE ("slug");

-- ----------------------------
-- Primary Key structure for table dashboard_templates
-- ----------------------------
ALTER TABLE "public"."dashboard_templates" ADD CONSTRAINT "dashboard_templates_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for dashboard_widgets
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.dashboard_widgets', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.dashboard_widgets) + 1, false);

-- ----------------------------
-- Indexes structure for table dashboard_widgets
-- ----------------------------
CREATE INDEX "idx_dashboard_widgets_project" ON "public"."dashboard_widgets" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."int4_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_dashboard_widgets_project IS '按项目查 Widget 配置';
CREATE INDEX "idx_dashboard_widgets_user" ON "public"."dashboard_widgets" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE user_id IS NOT NULL;
COMMENT ON INDEX public.idx_dashboard_widgets_user IS '按用户查询 Widget 个性化配置';

-- ----------------------------
-- Triggers structure for table dashboard_widgets
-- ----------------------------
CREATE TRIGGER "trg_dashboard_widgets_updated_at" BEFORE UPDATE ON "public"."dashboard_widgets"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_dashboard_widgets_updated_at IS 'dashboard_widgets: BEFORE UPDATE 自动将 updated_at 更新为 now()';

-- ----------------------------
-- Checks structure for table dashboard_widgets
-- ----------------------------
ALTER TABLE "public"."dashboard_widgets" ADD CONSTRAINT "dashboard_widgets_widget_type_check" CHECK (widget_type = ANY (ARRAY['progress_overview'::text, 'burndown'::text, 'velocity'::text, 'priority_split'::text, 'state_distribution'::text, 'overdue_list'::text, 'blocked_list'::text, 'risk_alert'::text, 'recent_activity'::text, 'team_workload'::text]));

-- ----------------------------
-- Primary Key structure for table dashboard_widgets
-- ----------------------------
ALTER TABLE "public"."dashboard_widgets" ADD CONSTRAINT "dashboard_widgets_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for deployment_events
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.deployment_events', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.deployment_events) + 1, false);

-- ----------------------------
-- Indexes structure for table deployment_events
-- ----------------------------
CREATE INDEX "idx_deployment_events_project" ON "public"."deployment_events" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "deployed_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE project_id IS NOT NULL;
COMMENT ON INDEX public.idx_deployment_events_project IS '按项目+时间倒序（部署历史）';
CREATE INDEX "idx_deployment_events_ws" ON "public"."deployment_events" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "deployed_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_deployment_events_ws IS '按工作空间查部署事件';
CREATE UNIQUE INDEX "uq_deployment_events_idempotent" ON "public"."deployment_events" USING btree (
  "deployment_id" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "env" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deployment_id IS NOT NULL;
COMMENT ON INDEX public.uq_deployment_events_idempotent IS '投递幂等键（同一事件+目标只投递一次）';

-- S11 性能优化: DORA 部署查询索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS "idx_deployment_dora"
ON "deployment_events" (project_id, status, deployed_at DESC)
WHERE project_id IS NOT NULL;
COMMENT ON INDEX public.idx_deployment_dora IS 'S11: DORA 指标查询按 project+status+deployed_at 覆盖索引（部署成功率/变更前置时间聚合）';

-- ----------------------------
-- Checks structure for table deployment_events
-- ----------------------------
ALTER TABLE "public"."deployment_events" ADD CONSTRAINT "deployment_events_env_check" CHECK (env = ANY (ARRAY['development'::text, 'staging'::text, 'production'::text, 'testing'::text]));
ALTER TABLE "public"."deployment_events" ADD CONSTRAINT "deployment_events_status_check" CHECK (status = ANY (ARRAY['success'::text, 'failed'::text, 'rolled_back'::text]));

-- ----------------------------
-- Primary Key structure for table deployment_events
-- ----------------------------
ALTER TABLE "public"."deployment_events" ADD CONSTRAINT "deployment_events_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for domain_events
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.domain_events', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.domain_events) + 1, false);

-- ----------------------------
-- Indexes structure for table domain_events
-- ----------------------------
CREATE INDEX "idx_events_unpublished" ON "public"."domain_events" USING btree (
  "id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE published_at IS NULL;
COMMENT ON INDEX public.idx_events_unpublished IS 'WHERE published_at IS NULL 未发布事件投递索引（Outbox reader）';

-- ----------------------------
-- Primary Key structure for table domain_events
-- ----------------------------
ALTER TABLE "public"."domain_events" ADD CONSTRAINT "domain_events_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table idempotency_keys
-- ----------------------------
ALTER TABLE "public"."idempotency_keys" ADD CONSTRAINT "idempotency_keys_pkey" PRIMARY KEY ("key");

-- ----------------------------
-- Indexes structure for table intake_channels
-- ----------------------------
CREATE INDEX "idx_intake_channels_project" ON "public"."intake_channels" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE project_id IS NOT NULL;
COMMENT ON INDEX public.idx_intake_channels_project IS '按项目查收件箱频道列表';
CREATE INDEX "idx_intake_channels_slug" ON "public"."intake_channels" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "slug" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE is_active = true;
COMMENT ON INDEX public.idx_intake_channels_slug IS '按 slug 公开门户路由（/intake/{slug} 校验 active）';
CREATE INDEX "idx_intake_channels_workspace" ON "public"."intake_channels" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_intake_channels_workspace IS '按工作空间查频道列表';

-- ----------------------------
-- Uniques structure for table intake_channels
-- ----------------------------
ALTER TABLE "public"."intake_channels" ADD CONSTRAINT "uq_intake_channel_slug" UNIQUE ("workspace_id", "slug");

-- ----------------------------
-- Primary Key structure for table intake_channels
-- ----------------------------
ALTER TABLE "public"."intake_channels" ADD CONSTRAINT "intake_channels_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table intake_issues
-- ----------------------------
CREATE INDEX "idx_intake_issues_channel" ON "public"."intake_issues" USING btree (
  "channel_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_intake_issues_channel IS '按频道查收件箱工单列表';
CREATE INDEX "idx_intake_issues_status" ON "public"."intake_issues" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_intake_issues_status IS '按状态过滤收件箱工单';
CREATE INDEX "idx_intake_issues_submitter" ON "public"."intake_issues" USING btree (
  "submitter_email" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_intake_issues_submitter IS '按提交人查询收件箱工单';
CREATE INDEX "idx_intake_issues_tracking" ON "public"."intake_issues" USING btree (
  "tracking_id" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_intake_issues_tracking IS '按 tracking_id YD-IN-XXXX 查询提交状态（用户跟踪页）';
CREATE INDEX "idx_intake_issues_workspace" ON "public"."intake_issues" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_intake_issues_workspace IS '按工作空间查收件箱工单列表';

-- ----------------------------
-- Uniques structure for table intake_issues
-- ----------------------------
ALTER TABLE "public"."intake_issues" ADD CONSTRAINT "uq_intake_issue_tracking" UNIQUE ("workspace_id", "tracking_id");

-- ----------------------------
-- Checks structure for table intake_issues
-- ----------------------------
ALTER TABLE "public"."intake_issues" ADD CONSTRAINT "chk_intake_status" CHECK (status::text = ANY (ARRAY['open'::character varying, 'accepted'::character varying, 'rejected'::character varying, 'archived'::character varying]::text[]));

-- ----------------------------
-- Primary Key structure for table intake_issues
-- ----------------------------
ALTER TABLE "public"."intake_issues" ADD CONSTRAINT "intake_issues_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for invitations
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.invitations', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.invitations) + 1, false);

-- ----------------------------
-- Indexes structure for table invitations
-- ----------------------------
CREATE INDEX "idx_invitations_email" ON "public"."invitations" USING btree (
  "email" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE status = 'pending'::text;
COMMENT ON INDEX public.idx_invitations_email IS '按邮箱查邀请记录（重复邀请校验）';
CREATE INDEX "idx_invitations_token" ON "public"."invitations" USING btree (
  "token_hash" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_invitations_token IS '按 token 查询邀请详情（注册流程）';
CREATE INDEX "idx_invitations_workspace" ON "public"."invitations" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE status = 'pending'::text;
COMMENT ON INDEX public.idx_invitations_workspace IS '按工作空间查邀请列表';

-- ----------------------------
-- Triggers structure for table invitations
-- ----------------------------
CREATE TRIGGER "trg_invitations_updated_at" BEFORE UPDATE ON "public"."invitations"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_invitations_updated_at IS 'invitations: BEFORE UPDATE 自动将 updated_at 更新为 now()';

-- ----------------------------
-- Uniques structure for table invitations
-- ----------------------------
ALTER TABLE "public"."invitations" ADD CONSTRAINT "invitations_token_hash_key" UNIQUE ("token_hash");

-- ----------------------------
-- Checks structure for table invitations
-- ----------------------------
ALTER TABLE "public"."invitations" ADD CONSTRAINT "invitations_role_check" CHECK (role = ANY (ARRAY['admin'::text, 'member'::text, 'guest'::text]));
ALTER TABLE "public"."invitations" ADD CONSTRAINT "invitations_status_check" CHECK (status = ANY (ARRAY['pending'::text, 'accepted'::text, 'revoked'::text, 'expired'::text]));

-- ----------------------------
-- Primary Key structure for table invitations
-- ----------------------------
ALTER TABLE "public"."invitations" ADD CONSTRAINT "invitations_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for issue_activities
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.issue_activities', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.issue_activities) + 1, false);

-- ----------------------------
-- Indexes structure for table issue_activities
-- ----------------------------
CREATE INDEX "idx_activities_issue" ON "public"."issue_activities" USING btree (
  "issue_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_activities_issue IS '按工作项+时间倒序（详情页活动时间线）';
CREATE INDEX "idx_activities_issue_covering" ON "public"."issue_activities" USING btree (
  "issue_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_activities_issue_covering IS '覆盖索引：按 work activity 高频列表查询';
CREATE INDEX "idx_activities_project" ON "public"."issue_activities" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_activities_project IS '按工作空间+项目查询活动日志';

-- ----------------------------
-- Checks structure for table issue_activities
-- ----------------------------
ALTER TABLE "public"."issue_activities" ADD CONSTRAINT "issue_activities_verb_check" CHECK (verb = ANY (ARRAY['created'::text, 'updated'::text, 'transitioned'::text, 'attached'::text, 'linked'::text, 'unlinked'::text, 'commented'::text]));

-- ----------------------------
-- Primary Key structure for table issue_activities
-- ----------------------------
ALTER TABLE "public"."issue_activities" ADD CONSTRAINT "issue_activities_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table issue_assignees
-- ----------------------------
CREATE INDEX "idx_issue_assignees_covering" ON "public"."issue_assignees" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_issue_assignees_covering IS '覆盖索引：按工作项+用户查指派（含常用字段）';
CREATE INDEX "idx_issue_assignees_user" ON "public"."issue_assignees" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_issue_assignees_user IS '按 user_id 查询"我的指派"（待办列表高频场景）';

-- ----------------------------
-- Primary Key structure for table issue_assignees
-- ----------------------------
ALTER TABLE "public"."issue_assignees" ADD CONSTRAINT "issue_assignees_pkey" PRIMARY KEY ("issue_id", "user_id");

-- ----------------------------
-- Indexes structure for table issue_comments
-- ----------------------------
CREATE INDEX "idx_issue_comments_author" ON "public"."issue_comments" USING btree (
  "created_by" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_issue_comments_author IS '按作者查评论列表';
CREATE INDEX "idx_issue_comments_issue" ON "public"."issue_comments" USING btree (
  "issue_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_issue_comments_issue IS '按工作项查评论列表';

-- ----------------------------
-- Primary Key structure for table issue_comments
-- ----------------------------
ALTER TABLE "public"."issue_comments" ADD CONSTRAINT "issue_comments_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for issue_dependencies
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.issue_dependencies', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.issue_dependencies) + 1, false);

-- ----------------------------
-- Indexes structure for table issue_dependencies
-- ----------------------------
CREATE INDEX "idx_issue_deps_pred" ON "public"."issue_dependencies" USING btree (
  "predecessor_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_issue_deps_pred IS '按前驱工作项查依赖关系';
CREATE INDEX "idx_issue_deps_succ" ON "public"."issue_dependencies" USING btree (
  "successor_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_issue_deps_succ IS '按后继工作项查依赖关系';

-- ----------------------------
-- Uniques structure for table issue_dependencies
-- ----------------------------
ALTER TABLE "public"."issue_dependencies" ADD CONSTRAINT "issue_dependencies_predecessor_id_successor_id_dependency_t_key" UNIQUE ("predecessor_id", "successor_id", "dependency_type");

-- ----------------------------
-- Checks structure for table issue_dependencies
-- ----------------------------
ALTER TABLE "public"."issue_dependencies" ADD CONSTRAINT "issue_dependencies_dependency_type_check" CHECK (dependency_type = ANY (ARRAY['FS'::text, 'SS'::text, 'FF'::text, 'SF'::text]));
ALTER TABLE "public"."issue_dependencies" ADD CONSTRAINT "no_self_dependency" CHECK (predecessor_id <> successor_id);

-- ----------------------------
-- Primary Key structure for table issue_dependencies
-- ----------------------------
ALTER TABLE "public"."issue_dependencies" ADD CONSTRAINT "issue_dependencies_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table issue_labels
-- ----------------------------
ALTER TABLE "public"."issue_labels" ADD CONSTRAINT "issue_labels_pkey" PRIMARY KEY ("issue_id", "label_id");

-- ----------------------------
-- Primary Key structure for table issue_modules
-- ----------------------------
ALTER TABLE "public"."issue_modules" ADD CONSTRAINT "issue_modules_pkey" PRIMARY KEY ("issue_id", "module_id");

-- ----------------------------
-- Auto increment value for issue_relations
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.issue_relations', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.issue_relations) + 1, false);

-- ----------------------------
-- Indexes structure for table issue_relations
-- ----------------------------
CREATE INDEX "idx_issue_relations_source" ON "public"."issue_relations" USING btree (
  "source_issue_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_issue_relations_source IS '按源工作项查关联关系';
CREATE INDEX "idx_issue_relations_target" ON "public"."issue_relations" USING btree (
  "target_issue_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_issue_relations_target IS '按目标工作项查关联关系';

-- ----------------------------
-- Uniques structure for table issue_relations
-- ----------------------------
ALTER TABLE "public"."issue_relations" ADD CONSTRAINT "issue_relations_source_issue_id_target_issue_id_relation_ty_key" UNIQUE ("source_issue_id", "target_issue_id", "relation_type");

-- ----------------------------
-- Checks structure for table issue_relations
-- ----------------------------
ALTER TABLE "public"."issue_relations" ADD CONSTRAINT "issue_relations_relation_type_check" CHECK (relation_type = ANY (ARRAY['duplicate'::text, 'relates_to'::text, 'blocked_by'::text, 'start_before'::text, 'finish_before'::text, 'implemented_by'::text]));
ALTER TABLE "public"."issue_relations" ADD CONSTRAINT "no_self_relation" CHECK (source_issue_id <> target_issue_id);

-- ----------------------------
-- Primary Key structure for table issue_relations
-- ----------------------------
ALTER TABLE "public"."issue_relations" ADD CONSTRAINT "issue_relations_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table issue_watchers
-- ----------------------------
CREATE INDEX "idx_issue_watchers_user" ON "public"."issue_watchers" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_issue_watchers_user IS '按 user_id 查关注的工作项列表';

-- ----------------------------
-- Primary Key structure for table issue_watchers
-- ----------------------------
ALTER TABLE "public"."issue_watchers" ADD CONSTRAINT "issue_watchers_pkey" PRIMARY KEY ("issue_id", "user_id");

-- ----------------------------
-- Indexes structure for table issue_reactions
-- ----------------------------
CREATE INDEX "idx_issue_reactions_issue" ON "public"."issue_reactions" USING btree (
  "issue_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_issue_reactions_issue IS '按工作项查所有反应（详情页聚合）';
CREATE INDEX "idx_issue_reactions_user" ON "public"."issue_reactions" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_issue_reactions_user IS '按用户查其发出的反应';

-- ----------------------------
-- Uniques structure for table issue_reactions
-- ----------------------------
ALTER TABLE "public"."issue_reactions" ADD CONSTRAINT "issue_reactions_issue_user_type_key" UNIQUE ("issue_id", "user_id", "reaction_type");

-- ----------------------------
-- Primary Key structure for table issue_reactions
-- ----------------------------
ALTER TABLE "public"."issue_reactions" ADD CONSTRAINT "issue_reactions_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table issue_votes
-- ----------------------------
CREATE INDEX "idx_issue_votes_issue" ON "public"."issue_votes" USING btree (
  "issue_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_issue_votes_issue IS '按工作项查投票聚合（赞成/反对计数）';

-- ----------------------------
-- Uniques structure for table issue_votes
-- ----------------------------
ALTER TABLE "public"."issue_votes" ADD CONSTRAINT "issue_votes_issue_user_key" UNIQUE ("issue_id", "user_id");

-- ----------------------------
-- Primary Key structure for table issue_votes
-- ----------------------------
ALTER TABLE "public"."issue_votes" ADD CONSTRAINT "issue_votes_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for issues
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.issues', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.issues) + 1, false);

-- ----------------------------
-- Indexes structure for table issues
-- ----------------------------
CREATE INDEX "idx_issues_created" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_issues_created IS '按工作空间+创建时间倒序（最近创建/活动日志查询）';
CREATE INDEX "idx_issues_fix_version" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "fix_version_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND fix_version_id IS NOT NULL;
COMMENT ON INDEX public.idx_issues_fix_version IS '按 fix_version 查询缺陷（版本修复追踪）';
CREATE INDEX "idx_issues_found_version" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "found_version_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND found_version_id IS NOT NULL;
COMMENT ON INDEX public.idx_issues_found_version IS '按 found_version 查询缺陷（发现版本统计）';
CREATE INDEX "idx_issues_list_covering" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "updated_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_issues_list_covering IS '覆盖索引：列表视图高频查询场景';
CREATE INDEX "idx_issues_parent" ON "public"."issues" USING btree (
  "parent_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND parent_id IS NOT NULL;
COMMENT ON INDEX public.idx_issues_parent IS '按 parent_id 查询子项（WBS 树展开/进度回写）';
CREATE INDEX "idx_issues_priority_covering" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "priority" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "updated_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE deleted_at IS NULL AND (priority = ANY (ARRAY['urgent'::text, 'high'::text]));
COMMENT ON INDEX public.idx_issues_priority_covering IS '覆盖索引：按项目+优先级过滤（看板/列表视图）';
CREATE UNIQUE INDEX "idx_issues_project_sequence" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sequence_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_issues_project_sequence IS '按项目+序号唯一（YD-123 展示编号查询）';
CREATE INDEX "idx_issues_project_state" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "state_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."float8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_issues_project_state IS '按工作空间+项目+状态查询（看板列表/过滤器高频场景；排除软删除）';
CREATE UNIQUE INDEX "idx_issues_public_id" ON "public"."issues" USING btree (
  "public_id" "pg_catalog"."uuid_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_issues_public_id IS '按 public_id 查询工作项（API/URL 路由）';
CREATE INDEX "idx_issues_release_version" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "release_version_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND release_version_id IS NOT NULL;
COMMENT ON INDEX public.idx_issues_release_version IS '按 release_version 查询（版本交付范围）';
CREATE INDEX "idx_issues_search_tsv" ON "public"."issues" USING gin (
  "search_tsv" "pg_catalog"."tsvector_ops"
);
COMMENT ON INDEX public.idx_issues_search_tsv IS 'tsvector 索引（search_tsv 列 GIN；PostgreSQL 全文检索）';
CREATE INDEX "idx_issues_state_covering" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "state_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."float8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_issues_state_covering IS '覆盖索引：按项目+状态查询看板列表';
CREATE INDEX "idx_issues_target_date" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "target_date" "pg_catalog"."date_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND completed_at IS NULL;
COMMENT ON INDEX public.idx_issues_target_date IS '按工作空间+项目+目标日期查询（未完成逾期提醒/甘特图）';
CREATE INDEX "idx_issues_target_date_covering" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "target_date" "pg_catalog"."date_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND target_date IS NOT NULL;
COMMENT ON INDEX public.idx_issues_target_date_covering IS '覆盖索引：目标日期查询场景';
CREATE INDEX "idx_issues_type" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "type_code" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_issues_type IS '按工作项类型过滤（需求/任务/缺陷列表视图切换）';
CREATE INDEX "idx_issues_type_covering" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "type_code" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_issues_type_covering IS '覆盖索引：按类型过滤列表查询';
CREATE INDEX "idx_issues_updated" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "updated_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_issues_updated IS '按工作空间+更新时间倒序（最近活动/时间线查询）';
CREATE INDEX "idx_issues_workspace_project" ON "public"."issues" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_issues_workspace_project IS '按工作空间+项目高频过滤查询';

-- ----------------------------
-- Triggers structure for table issues
-- ----------------------------
CREATE TRIGGER "trg_issue_search_cleanup" AFTER UPDATE OF "deleted_at" ON "public"."issues"
FOR EACH ROW
WHEN ((new.deleted_at IS NOT NULL))
EXECUTE PROCEDURE "public"."fn_cleanup_search_document"();
-- FIXED: COMMENT ON TRIGGER trg_issue_search_cleanup IS 'issues: AFTER DELETE（软删除）异步清理 ES 索引对应文档';
CREATE TRIGGER "trg_issue_search_sync" AFTER INSERT OR UPDATE OF "name", "description_stripped" ON "public"."issues"
FOR EACH ROW
WHEN ((new.deleted_at IS NULL))
EXECUTE PROCEDURE "public"."fn_refresh_search_document"();
-- FIXED: COMMENT ON TRIGGER trg_issue_search_sync IS 'issues: AFTER INSERT/UPDATE 异步同步 ES 索引 `ydsz_issues`；routing=workspace_id';
CREATE TRIGGER "trg_issues_updated_at" BEFORE UPDATE ON "public"."issues"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_issues_updated_at IS 'issues: BEFORE UPDATE 自动将 updated_at 更新为 now()；事件监听同步 ES';

-- ----------------------------
-- Checks structure for table issues
-- ----------------------------
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_delay_reason_check" CHECK (delay_reason = ANY (ARRAY['requirement_change'::text, 'resource'::text, 'blocked'::text, 'other'::text]));
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_depth_check" CHECK (depth >= 1 AND depth <= 3);
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_found_phase_check" CHECK (found_phase = ANY (ARRAY['unit'::text, 'integration'::text, 'uat'::text, 'production'::text, 'customer'::text]));
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_point_check" CHECK (point >= 0 AND point <= 12);
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_priority_check" CHECK (priority = ANY (ARRAY['urgent'::text, 'high'::text, 'medium'::text, 'low'::text, 'none'::text]));
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_progress_check" CHECK (progress >= 0 AND progress <= 100);
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_root_cause_category_check" CHECK (root_cause_category = ANY (ARRAY['requirement'::text, 'technical'::text, 'environment'::text, 'data'::text]));
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_severity_check" CHECK (severity >= 1 AND severity <= 5);
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_type_code_check" CHECK (type_code = ANY (ARRAY['epic'::text, 'requirement'::text, 'task'::text, 'defect'::text]));
ALTER TABLE "public"."issues" ADD CONSTRAINT "defect_required" CHECK (type_code <> 'defect'::text OR severity IS NOT NULL AND found_phase IS NOT NULL);

-- ----------------------------
-- Primary Key structure for table issues
-- ----------------------------
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for labels
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.labels', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.labels) + 1, false);

-- ----------------------------
-- Indexes structure for table labels
-- ----------------------------
CREATE INDEX "idx_labels_project" ON "public"."labels" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_labels_project IS '按工作空间+项目查标签列表（项目标签管理/选择器）';

-- ----------------------------
-- Triggers structure for table labels
-- ----------------------------
CREATE TRIGGER "trg_labels_updated_at" BEFORE UPDATE ON "public"."labels"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_labels_updated_at IS 'labels: BEFORE UPDATE 自动将 updated_at 更新为 now()';

-- ----------------------------
-- Primary Key structure for table labels
-- ----------------------------
ALTER TABLE "public"."labels" ADD CONSTRAINT "labels_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for metric_adjustments
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.metric_adjustments', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.metric_adjustments) + 1, false);

-- ----------------------------
-- Indexes structure for table metric_adjustments
-- ----------------------------
CREATE INDEX "idx_metric_adjustments_ws" ON "public"."metric_adjustments" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "snapshot_date" "pg_catalog"."date_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_metric_adjustments_ws IS '按工作空间查询指标调整记录';

-- ----------------------------
-- Triggers structure for table metric_adjustments
-- ----------------------------
CREATE TRIGGER "trg_metric_adjustments_updated_at" BEFORE UPDATE ON "public"."metric_adjustments"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

-- ----------------------------
-- Primary Key structure for table metric_adjustments
-- ----------------------------
ALTER TABLE "public"."metric_adjustments" ADD CONSTRAINT "metric_adjustments_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for metric_snapshots
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.metric_snapshots', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.metric_snapshots) + 1, false);

-- ----------------------------
-- Indexes structure for table metric_snapshots
-- ----------------------------
CREATE INDEX "idx_metric_snap_covering" ON "public"."metric_snapshots" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "metric" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "snapshot_date" "pg_catalog"."date_ops" ASC NULLS LAST,
  "value" "pg_catalog"."numeric_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_metric_snap_covering IS '覆盖指标快照查询（granularity+ref_id+metric+metric_date）';
CREATE INDEX "idx_metric_snap_date" ON "public"."metric_snapshots" USING btree (
  "snapshot_date" "pg_catalog"."date_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_metric_snap_date IS '按日期范围查询指标快照（趋势图）';
CREATE INDEX "idx_metric_snap_lookup" ON "public"."metric_snapshots" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "metric" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "snapshot_date" "pg_catalog"."date_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_metric_snap_lookup IS '按 (granularity, ref_id, metric, metric_date) 查询/覆盖快照';
CREATE INDEX "idx_metric_snap_ws" ON "public"."metric_snapshots" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "metric" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "snapshot_date" "pg_catalog"."date_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_metric_snap_ws IS '按工作空间查询指标快照';
CREATE UNIQUE INDEX "idx_metric_snapshots_project" ON "public"."metric_snapshots" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "granularity" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "ref_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "metric" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "snapshot_date" "pg_catalog"."date_ops" ASC NULLS LAST
) WHERE project_id IS NOT NULL;
COMMENT ON INDEX public.idx_metric_snapshots_project IS '按项目查快照列表';

-- S11 性能优化: 指标趋势覆盖索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS "idx_metric_snap_trend"
ON "metric_snapshots" (workspace_id, project_id, metric, snapshot_date DESC, value);
COMMENT ON INDEX public.idx_metric_snap_trend IS 'S11: 趋势图按 ws+project+metric+date+value 覆盖索引（列表渲染消除回表）';
CREATE UNIQUE INDEX "idx_metric_snapshots_workspace" ON "public"."metric_snapshots" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "granularity" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "metric" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "snapshot_date" "pg_catalog"."date_ops" ASC NULLS LAST
) WHERE project_id IS NULL;
COMMENT ON INDEX public.idx_metric_snapshots_workspace IS '按工作空间查快照列表';

-- ----------------------------
-- Checks structure for table metric_snapshots
-- ----------------------------
ALTER TABLE "public"."metric_snapshots" ADD CONSTRAINT "metric_snapshots_granularity_check" CHECK (granularity = ANY (ARRAY['daily'::text, 'sprint'::text, 'version'::text]));

-- ----------------------------
-- Primary Key structure for table metric_snapshots
-- ----------------------------
ALTER TABLE "public"."metric_snapshots" ADD CONSTRAINT "metric_snapshots_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for modules
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.modules', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.modules) + 1, false);

-- ----------------------------
-- Indexes structure for table modules
-- ----------------------------
CREATE INDEX "idx_modules_project" ON "public"."modules" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_modules_project IS '按工作空间+项目查询模块列表';

-- ----------------------------
-- Triggers structure for table modules
-- ----------------------------
CREATE TRIGGER "trg_modules_updated_at" BEFORE UPDATE ON "public"."modules"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_modules_updated_at IS 'modules: BEFORE UPDATE 自动将 updated_at 更新为 now()';

-- ----------------------------
-- Checks structure for table modules
-- ----------------------------
ALTER TABLE "public"."modules" ADD CONSTRAINT "modules_status_check" CHECK (status = ANY (ARRAY['active'::text, 'completed'::text, 'cancelled'::text]));

-- ----------------------------
-- Primary Key structure for table modules
-- ----------------------------
ALTER TABLE "public"."modules" ADD CONSTRAINT "modules_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table notification_deliveries
-- ----------------------------
CREATE INDEX "idx_deliveries_next_retry" ON "public"."notification_deliveries" USING btree (
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "next_retry_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
) WHERE status::text = 'pending'::text;
COMMENT ON INDEX public.idx_deliveries_next_retry IS '按下次重试时间查待投递记录（调度器轮询）';
CREATE INDEX "idx_deliveries_notification" ON "public"."notification_deliveries" USING btree (
  "notification_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_deliveries_notification IS '按通知记录查投递明细';
CREATE INDEX "idx_deliveries_status" ON "public"."notification_deliveries" USING btree (
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
) WHERE status::text = 'pending'::text;
COMMENT ON INDEX public.idx_deliveries_status IS '按投递状态查询（统计/监控）';

-- ----------------------------
-- Primary Key structure for table notification_deliveries
-- ----------------------------
ALTER TABLE "public"."notification_deliveries" ADD CONSTRAINT "notification_deliveries_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table notification_digests
-- ----------------------------
CREATE INDEX "idx_digests_pending" ON "public"."notification_digests" USING btree (
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "scheduled_for" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
) WHERE status::text = 'pending'::text;
COMMENT ON INDEX public.idx_digests_pending IS '按状态查待生成摘要（调度器轮询）';
CREATE UNIQUE INDEX "idx_notification_digests_pending" ON "public"."notification_digests" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "digest_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE status::text = 'pending'::text;
COMMENT ON INDEX public.idx_notification_digests_pending IS '按状态查待投递摘要队列';

-- ----------------------------
-- Primary Key structure for table notification_digests
-- ----------------------------
ALTER TABLE "public"."notification_digests" ADD CONSTRAINT "notification_digests_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Uniques structure for table notification_preferences
-- ----------------------------
ALTER TABLE "public"."notification_preferences" ADD CONSTRAINT "notification_preferences_user_id_workspace_id_key" UNIQUE ("user_id", "workspace_id");

-- ----------------------------
-- Primary Key structure for table notification_preferences
-- ----------------------------
ALTER TABLE "public"."notification_preferences" ADD CONSTRAINT "notification_preferences_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table notifications
-- ----------------------------
CREATE INDEX "idx_notifications_archived" ON "public"."notifications" USING btree (
  "created_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
) WHERE is_archived = true;
COMMENT ON INDEX public.idx_notifications_archived IS '按已读状态过滤通知列表（活跃通知筛选）';
CREATE INDEX "idx_notifications_entity" ON "public"."notifications" USING btree (
  "entity_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "entity_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_notifications_entity IS '按关联实体查通知记录';
CREATE INDEX "idx_notifications_recipient_unread" ON "public"."notifications" USING btree (
  "recipient_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "is_read" "pg_catalog"."bool_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE is_archived = false;
COMMENT ON INDEX public.idx_notifications_recipient_unread IS '按接收人+未读过滤（站内信列表未读数；按时间倒序）';

-- ----------------------------
-- Primary Key structure for table notifications
-- ----------------------------
ALTER TABLE "public"."notifications" ADD CONSTRAINT "notifications_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table password_reset_tokens
-- ----------------------------
CREATE INDEX "idx_password_reset_tokens_expires" ON "public"."password_reset_tokens" USING btree (
  "expires_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_password_reset_tokens_expires IS '按过期时间查询 Token（定时清理）';
CREATE UNIQUE INDEX "idx_password_reset_tokens_user_active" ON "public"."password_reset_tokens" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE used_at IS NULL;
COMMENT ON INDEX public.idx_password_reset_tokens_user_active IS '按用户查有效 Token（只能有一个 active）';

-- ----------------------------
-- Primary Key structure for table password_reset_tokens
-- ----------------------------
ALTER TABLE "public"."password_reset_tokens" ADD CONSTRAINT "password_reset_tokens_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table project_sequences
-- ----------------------------
ALTER TABLE "public"."project_sequences" ADD CONSTRAINT "project_sequences_pkey" PRIMARY KEY ("project_id");

-- ----------------------------
-- Auto increment value for projects
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.projects', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.projects) + 1, false);

-- ----------------------------
-- Indexes structure for table projects
-- ----------------------------
CREATE INDEX "idx_projects_template" ON "public"."projects" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "template" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_projects_template IS '按模板查询项目列表';
CREATE INDEX "idx_projects_workspace" ON "public"."projects" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_projects_workspace IS '按工作空间查项目列表（含激活/归档状态过滤）';
CREATE UNIQUE INDEX "idx_projects_workspace_identifier" ON "public"."projects" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "identifier" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_projects_workspace_identifier IS '按工作空间+标识符唯一（URL 路由定位 YD 项目）';
CREATE UNIQUE INDEX "idx_projects_workspace_slug" ON "public"."projects" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "slug" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_projects_workspace_slug IS '按工作空间+slug 唯一（URL 友好标识）';

-- ----------------------------
-- Triggers structure for table projects
-- ----------------------------
CREATE TRIGGER "trg_projects_updated_at" BEFORE UPDATE ON "public"."projects"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_projects_updated_at IS 'projects: BEFORE UPDATE 自动将 updated_at 更新为 now()';

-- ----------------------------
-- Uniques structure for table projects
-- ----------------------------
ALTER TABLE "public"."projects" ADD CONSTRAINT "projects_public_id_key" UNIQUE ("public_id");

-- ----------------------------
-- Checks structure for table projects
-- ----------------------------
ALTER TABLE "public"."projects" ADD CONSTRAINT "projects_network_check" CHECK (network = ANY (ARRAY['public'::text, 'private'::text]));
ALTER TABLE "public"."projects" ADD CONSTRAINT "projects_status_check" CHECK (status = ANY (ARRAY['active'::text, 'archived'::text]));
ALTER TABLE "public"."projects" ADD CONSTRAINT "projects_template_check" CHECK (template = ANY (ARRAY['agile'::text, 'waterfall'::text, 'generic'::text]));

-- ----------------------------
-- Checks structure for table project_members
-- ----------------------------
ALTER TABLE "public"."project_members" ADD CONSTRAINT "project_members_role_check" CHECK (role = ANY (ARRAY['admin'::text, 'member'::text]));
ALTER TABLE "public"."project_members" ADD CONSTRAINT "project_members_uniq" UNIQUE ("workspace_id", "project_id", "user_id");

-- ----------------------------
-- Primary Key structure for table projects
-- ----------------------------
ALTER TABLE "public"."projects" ADD CONSTRAINT "projects_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table project_members
-- ----------------------------
ALTER TABLE "public"."project_members" ADD CONSTRAINT "project_members_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table project_members
-- ----------------------------
CREATE INDEX "idx_project_members_project" ON "public"."project_members" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_project_members_project IS '按工作空间+项目查询成员列表';
CREATE INDEX "idx_project_members_user" ON "public"."project_members" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_project_members_user IS '按用户查询其参与的项目';

-- ----------------------------
-- Auto increment value for recent_items
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.recent_items', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.recent_items) + 1, false);

-- ----------------------------
-- Indexes structure for table recent_items
-- ----------------------------
CREATE INDEX "idx_recent_items_user" ON "public"."recent_items" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "accessed_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_recent_items_user IS '按用户+最后访问时间倒序（首页最近列表 Top N）';
CREATE INDEX "idx_recent_items_ws" ON "public"."recent_items" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_recent_items_ws IS '按工作空间查最近访问记录';

-- ----------------------------
-- Triggers structure for table recent_items
-- ----------------------------
CREATE TRIGGER "trg_recent_items_touch" BEFORE UPDATE ON "public"."recent_items"
FOR EACH ROW
EXECUTE PROCEDURE "public"."fn_touch_recent_item"();
-- FIXED: COMMENT ON TRIGGER trg_recent_items_touch IS 'recent_items: BEFORE UPDATE 每次访问重置 updated_at = now()；服务于"最近访问"列表排序';

-- ----------------------------
-- Uniques structure for table recent_items
-- ----------------------------
ALTER TABLE "public"."recent_items" ADD CONSTRAINT "recent_items_user_id_item_type_item_id_key" UNIQUE ("user_id", "item_type", "item_id");

-- ----------------------------
-- Checks structure for table recent_items
-- ----------------------------
ALTER TABLE "public"."recent_items" ADD CONSTRAINT "recent_items_item_type_check" CHECK (item_type = ANY (ARRAY['project'::text, 'issue'::text, 'sprint'::text, 'version'::text]));

-- ----------------------------
-- Primary Key structure for table recent_items
-- ----------------------------
ALTER TABLE "public"."recent_items" ADD CONSTRAINT "recent_items_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for risk_alerts
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.risk_alerts', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.risk_alerts) + 1, false);

-- ----------------------------
-- Indexes structure for table risk_alerts
-- ----------------------------
CREATE INDEX "idx_risk_alerts_project" ON "public"."risk_alerts" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE NOT is_resolved;
COMMENT ON INDEX public.idx_risk_alerts_project IS '按项目查风险告警列表';
CREATE INDEX "idx_risk_alerts_unresolved" ON "public"."risk_alerts" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "severity" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE NOT is_resolved;
COMMENT ON INDEX public.idx_risk_alerts_unresolved IS '未解决风险告警查询（Dashboard widget 视图）';
CREATE INDEX "idx_risk_alerts_workspace" ON "public"."risk_alerts" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE NOT is_resolved;
COMMENT ON INDEX public.idx_risk_alerts_workspace IS '按工作空间查风险告警列表';

-- ----------------------------
-- Checks structure for table risk_alerts
-- ----------------------------
ALTER TABLE "public"."risk_alerts" ADD CONSTRAINT "risk_alerts_severity_check" CHECK (severity = ANY (ARRAY['info'::text, 'low'::text, 'medium'::text, 'high'::text, 'critical'::text]));

-- ----------------------------
-- Primary Key structure for table risk_alerts
-- ----------------------------
ALTER TABLE "public"."risk_alerts" ADD CONSTRAINT "risk_alerts_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for risk_rules
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.risk_rules', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.risk_rules) + 1, false);

-- ----------------------------
-- Indexes structure for table risk_rules
-- ----------------------------
CREATE INDEX "idx_risk_rules_active" ON "public"."risk_rules" USING btree (
  "is_active" "pg_catalog"."bool_ops" ASC NULLS LAST
) WHERE is_active = true;
COMMENT ON INDEX public.idx_risk_rules_active IS '按启用状态查规则列表';
CREATE INDEX "idx_risk_rules_project" ON "public"."risk_rules" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE project_id IS NOT NULL;
COMMENT ON INDEX public.idx_risk_rules_project IS '按项目查规则列表';
CREATE INDEX "idx_risk_rules_workspace" ON "public"."risk_rules" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_risk_rules_workspace IS '按工作空间查规则列表';

-- ----------------------------
-- Triggers structure for table risk_rules
-- ----------------------------
CREATE TRIGGER "trg_risk_rules_updated_at" BEFORE UPDATE ON "public"."risk_rules"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_risk_rules_updated_at IS 'risk_rules: BEFORE UPDATE 自动维护 updated_at';

-- ----------------------------
-- Checks structure for table risk_rules
-- ----------------------------
ALTER TABLE "public"."risk_rules" ADD CONSTRAINT "risk_rules_rule_type_check" CHECK (rule_type = ANY (ARRAY['overdue_issue'::text, 'overdue_sprint'::text, 'blocked_count'::text, 'sla_breach'::text, 'stalled_progress'::text, 'high_priority_open'::text]));

-- ----------------------------
-- Primary Key structure for table risk_rules
-- ----------------------------
ALTER TABLE "public"."risk_rules" ADD CONSTRAINT "risk_rules_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for rule_executions
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.rule_executions', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.rule_executions) + 1, false);

-- ----------------------------
-- Indexes structure for table rule_executions
-- ----------------------------
CREATE INDEX "idx_rule_executions_event" ON "public"."rule_executions" USING btree (
  "trigger_event_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE trigger_event_id IS NOT NULL;
COMMENT ON INDEX public.idx_rule_executions_event IS '按事件反查规则执行记录';
CREATE INDEX "idx_rule_executions_project" ON "public"."rule_executions" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE project_id IS NOT NULL;
COMMENT ON INDEX public.idx_rule_executions_project IS '按项目查规则执行列表';
CREATE INDEX "idx_rule_executions_rule" ON "public"."rule_executions" USING btree (
  "rule_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_rule_executions_rule IS '按规则反查执行历史（调试/监控）';
CREATE INDEX "idx_rule_executions_ws" ON "public"."rule_executions" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_rule_executions_ws IS '按工作空间查规则执行列表';

-- S11 性能优化: 规则执行历史 + 幂等去重
CREATE INDEX CONCURRENTLY IF NOT EXISTS "idx_rule_executions_rule_created"
ON "rule_executions" (rule_id DESC, created_at DESC)
WHERE trigger_event_id IS NOT NULL;
COMMENT ON INDEX public.idx_rule_executions_rule_created IS 'S11: 按 rule+created_at 倒序覆盖索引（最近执行历史分页）';
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS "idx_rule_executions_idempotent"
ON "rule_executions" (rule_id, trigger_event_id)
WHERE trigger_event_id IS NOT NULL;
COMMENT ON INDEX public.idx_rule_executions_idempotent IS 'S11: 幂等去重键 UNIQUE 索引（防重复投递）';

-- ----------------------------
-- Checks structure for table rule_executions
-- ----------------------------
ALTER TABLE "public"."rule_executions" ADD CONSTRAINT "rule_executions_status_check" CHECK (status = ANY (ARRAY['matched'::text, 'skipped'::text, 'success'::text, 'failed'::text, 'dry_run'::text]));

-- ----------------------------
-- Primary Key structure for table rule_executions
-- ----------------------------
ALTER TABLE "public"."rule_executions" ADD CONSTRAINT "rule_executions_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table schema_migrations
-- ----------------------------
ALTER TABLE "public"."schema_migrations" ADD CONSTRAINT "schema_migrations_pkey" PRIMARY KEY ("version");

-- ----------------------------
-- Auto increment value for search_bookmarks
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.search_bookmarks', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.search_bookmarks) + 1, false);

-- ----------------------------
-- Indexes structure for table search_bookmarks
-- ----------------------------
CREATE INDEX "idx_search_bookmarks_project" ON "public"."search_bookmarks" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_search_bookmarks_project IS '按项目查搜索收藏';
CREATE INDEX "idx_search_bookmarks_user" ON "public"."search_bookmarks" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."float8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_search_bookmarks_user IS '按用户查收藏列表（搜索收藏管理页）';
CREATE INDEX "idx_search_bookmarks_ws" ON "public"."search_bookmarks" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_search_bookmarks_ws IS '按工作空间查搜索收藏';

-- ----------------------------
-- Triggers structure for table search_bookmarks
-- ----------------------------
CREATE TRIGGER "trg_search_bookmarks_updated_at" BEFORE UPDATE ON "public"."search_bookmarks"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_search_bookmarks_updated_at IS 'search_bookmarks: BEFORE UPDATE 自动将 updated_at 更新为 now()';

-- ----------------------------
-- Primary Key structure for table search_bookmarks
-- ----------------------------
ALTER TABLE "public"."search_bookmarks" ADD CONSTRAINT "search_bookmarks_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for search_documents
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.search_documents', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.search_documents) + 1, false);

-- ----------------------------
-- Indexes structure for table search_documents
-- ----------------------------
CREATE INDEX "idx_search_documents_project" ON "public"."search_documents" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "doc_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "updated_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_search_documents_project IS '按项目查搜索文档列表';
CREATE INDEX "idx_search_documents_tsv" ON "public"."search_documents" USING gin (
  "search_tsv" "pg_catalog"."tsvector_ops"
);
COMMENT ON INDEX public.idx_search_documents_tsv IS 'GIN 全文索引（search_tsv 列；ES 不可用时兜底）';
CREATE UNIQUE INDEX "idx_search_documents_unique" ON "public"."search_documents" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "doc_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "doc_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_search_documents_unique IS '按文档唯一性约束检索';
CREATE INDEX "idx_search_documents_ws" ON "public"."search_documents" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "doc_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_search_documents_ws IS '按工作空间查搜索文档';

-- ----------------------------
-- Triggers structure for table search_documents
-- ----------------------------
CREATE TRIGGER "trg_search_documents_updated_at" BEFORE UPDATE ON "public"."search_documents"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_search_documents_updated_at IS 'search_documents: BEFORE UPDATE 自动将 updated_at 更新为 now()';

-- ----------------------------
-- Checks structure for table search_documents
-- ----------------------------
ALTER TABLE "public"."search_documents" ADD CONSTRAINT "search_documents_doc_type_check" CHECK (doc_type = ANY (ARRAY['issue'::text, 'sprint'::text, 'version'::text]));

-- ----------------------------
-- Primary Key structure for table search_documents
-- ----------------------------
ALTER TABLE "public"."search_documents" ADD CONSTRAINT "search_documents_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for search_history
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.search_history', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.search_history) + 1, false);

-- ----------------------------
-- Indexes structure for table search_history
-- ----------------------------
CREATE INDEX "idx_search_history_user" ON "public"."search_history" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "searched_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_search_history_user IS '按用户查搜索历史（最近搜索自动补全）';
CREATE INDEX "idx_search_history_ws_user" ON "public"."search_history" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_search_history_ws_user IS '按工作空间+用户查搜索历史';

-- ----------------------------
-- Primary Key structure for table search_history
-- ----------------------------
ALTER TABLE "public"."search_history" ADD CONSTRAINT "search_history_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table sprint_issues
-- ----------------------------
CREATE INDEX "idx_sprint_issues_issue" ON "public"."sprint_issues" USING btree (
  "issue_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_sprint_issues_issue IS '按 issue_id 反查所属迭代';

-- ----------------------------
-- Primary Key structure for table sprint_issues
-- ----------------------------
ALTER TABLE "public"."sprint_issues" ADD CONSTRAINT "sprint_issues_pkey" PRIMARY KEY ("sprint_id", "issue_id");

-- ----------------------------
-- Auto increment value for sprint_snapshots
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.sprint_snapshots', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.sprint_snapshots) + 1, false);

-- ----------------------------
-- Indexes structure for table sprint_snapshots
-- ----------------------------
CREATE UNIQUE INDEX "idx_sprint_snapshots_unique" ON "public"."sprint_snapshots" USING btree (
  "sprint_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "snapshot_date" "pg_catalog"."date_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_sprint_snapshots_unique IS '按 sprint+时间查燃尽图快照序列';
CREATE INDEX "idx_sprintsnapshots_project" ON "public"."sprint_snapshots" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_sprintsnapshots_project IS '按项目查燃尽图快照列表';
CREATE INDEX "idx_sprintsnapshots_sprint_date" ON "public"."sprint_snapshots" USING btree (
  "sprint_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "snapshot_date" "pg_catalog"."date_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_sprintsnapshots_sprint_date IS '按 sprint+日期查快照（趋势图）';

-- ----------------------------
-- Primary Key structure for table sprint_snapshots
-- ----------------------------
ALTER TABLE "public"."sprint_snapshots" ADD CONSTRAINT "sprint_snapshots_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for sprints
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.sprints', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.sprints) + 1, false);

-- ----------------------------
-- Indexes structure for table sprints
-- ----------------------------
CREATE UNIQUE INDEX "idx_one_active_sprint_per_project" ON "public"."sprints" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE status = 'active'::text AND deleted_at IS NULL;
COMMENT ON INDEX public.idx_one_active_sprint_per_project IS '每个项目仅一个 active 迭代（部分索引 WHERE active）';
CREATE INDEX "idx_sprints_active_unique" ON "public"."sprints" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE status = 'active'::text AND deleted_at IS NULL;
COMMENT ON INDEX public.idx_sprints_active_unique IS 'active 迭代唯一性校验（部分索引）';
CREATE INDEX "idx_sprints_project_status" ON "public"."sprints" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_sprints_project_status IS '按项目+迭代状态查询（active 唯一性校验 / 迭代列表）';
CREATE INDEX "idx_sprints_version" ON "public"."sprints" USING btree (
  "version_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_sprints_version IS '按 version 反查关联迭代';

-- ----------------------------
-- Triggers structure for table sprints
-- ----------------------------
CREATE TRIGGER "trg_sprints_updated_at" BEFORE UPDATE ON "public"."sprints"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_sprints_updated_at IS 'sprints: BEFORE UPDATE 自动将 updated_at 更新为 now()';
CREATE TRIGGER "trg_sprint_search_sync" AFTER INSERT OR UPDATE OF "name", "goal" ON "public"."sprints"
FOR EACH ROW
WHEN ((new.deleted_at IS NULL))
EXECUTE PROCEDURE "public"."fn_refresh_sprint_search_document"();
-- FIXED: COMMENT ON TRIGGER trg_sprint_search_sync IS 'sprints: AFTER INSERT/UPDATE 同步 ES 降级索引 search_documents';
CREATE TRIGGER "trg_sprint_search_cleanup" AFTER UPDATE OF "deleted_at" ON "public"."sprints"
FOR EACH ROW
WHEN ((new.deleted_at IS NOT NULL))
EXECUTE PROCEDURE "public"."fn_cleanup_search_document"();
-- FIXED: COMMENT ON TRIGGER trg_sprint_search_cleanup IS 'sprints: AFTER DELETE 清理 ES 及 search_documents 中对应文档';

-- ----------------------------
-- Checks structure for table sprints
-- ----------------------------
ALTER TABLE "public"."sprints" ADD CONSTRAINT "sprints_status_check" CHECK (status = ANY (ARRAY['planned'::text, 'active'::text, 'completed'::text]));
ALTER TABLE "public"."sprints" ADD CONSTRAINT "sprint_date_range" CHECK (start_date IS NULL OR end_date IS NULL OR start_date <= end_date);

-- ----------------------------
-- Primary Key structure for table sprints
-- ----------------------------
ALTER TABLE "public"."sprints" ADD CONSTRAINT "sprints_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for state_transitions
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.state_transitions', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.state_transitions) + 1, false);

-- ----------------------------
-- Indexes structure for table state_transitions
-- ----------------------------
CREATE INDEX "idx_state_transitions_lookup" ON "public"."state_transitions" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "type_code" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "from_state_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_state_transitions_lookup IS '(project, type, from_state, to_state) 唯一；防重复定义流转';

-- ----------------------------
-- Uniques structure for table state_transitions
-- ----------------------------
ALTER TABLE "public"."state_transitions" ADD CONSTRAINT "state_transitions_project_id_type_code_from_state_id_to_sta_key" UNIQUE ("project_id", "type_code", "from_state_id", "to_state_id");

-- ----------------------------
-- Primary Key structure for table state_transitions
-- ----------------------------
ALTER TABLE "public"."state_transitions" ADD CONSTRAINT "state_transitions_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for states
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.states', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.states) + 1, false);

-- ----------------------------
-- Indexes structure for table states
-- ----------------------------
CREATE INDEX "idx_states_project" ON "public"."states" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sequence" "pg_catalog"."float8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_states_project IS '按项目+类型查状态集列表（不同类型独立状态模板）';

-- ----------------------------
-- Triggers structure for table states
-- ----------------------------
CREATE TRIGGER "trg_states_updated_at" BEFORE UPDATE ON "public"."states"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_states_updated_at IS 'states: BEFORE UPDATE 自动将 updated_at 更新为 now()';

-- ----------------------------
-- Checks structure for table states
-- ----------------------------
ALTER TABLE "public"."states" ADD CONSTRAINT "states_group_check" CHECK ("group" = ANY (ARRAY['backlog'::text, 'started'::text, 'completed'::text, 'cancelled'::text]));
ALTER TABLE "public"."states" ADD CONSTRAINT "states_template_set_check" CHECK (template_set = ANY (ARRAY['dev_flow'::text, 'defect_flow'::text, 'requirement_flow'::text, 'custom'::text]));

-- ----------------------------
-- Primary Key structure for table states
-- ----------------------------
ALTER TABLE "public"."states" ADD CONSTRAINT "states_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for time_logs
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.time_logs', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.time_logs) + 1, false);

-- ----------------------------
-- Indexes structure for table time_logs
-- ----------------------------
CREATE INDEX "idx_time_logs_issue" ON "public"."time_logs" USING btree (
  "issue_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_time_logs_issue IS '按工作项查工时明细（详情页时间线 / sum 重算 actual_effort）';
CREATE INDEX "idx_time_logs_user_date" ON "public"."time_logs" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "spent_date" "pg_catalog"."date_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_time_logs_user_date IS '按用户+日期查工时（成员工时报表 / 负载热力图）';

-- ----------------------------
-- Triggers structure for table time_logs
-- ----------------------------
CREATE TRIGGER "trg_time_logs_updated_at" BEFORE UPDATE ON "public"."time_logs"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_time_logs_updated_at IS 'time_logs: BEFORE UPDATE 自动将 updated_at 更新为 now()';

-- ----------------------------
-- Checks structure for table time_logs
-- ----------------------------
ALTER TABLE "public"."time_logs" ADD CONSTRAINT "time_logs_duration_minutes_check" CHECK (duration_minutes > 0 AND duration_minutes <= 1440);

-- ----------------------------
-- Primary Key structure for table time_logs
-- ----------------------------
ALTER TABLE "public"."time_logs" ADD CONSTRAINT "time_logs_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for users
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.users', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.users) + 1, false);

-- ----------------------------
-- Triggers structure for table users
-- ----------------------------
CREATE TRIGGER "trg_users_updated_at" BEFORE UPDATE ON "public"."users"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_users_updated_at IS 'users: BEFORE UPDATE 自动将 updated_at 更新为 now()';

-- ----------------------------
-- Uniques structure for table users
-- ----------------------------
ALTER TABLE "public"."users" ADD CONSTRAINT "users_public_id_key" UNIQUE ("public_id");
ALTER TABLE "public"."users" ADD CONSTRAINT "users_email_key" UNIQUE ("email");

-- ----------------------------
-- Primary Key structure for table users
-- ----------------------------
ALTER TABLE "public"."users" ADD CONSTRAINT "users_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for version_delivery_snapshots
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.version_delivery_snapshots', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.version_delivery_snapshots) + 1, false);

-- ----------------------------
-- Indexes structure for table version_delivery_snapshots
-- ----------------------------
CREATE INDEX "idx_vds_version" ON "public"."version_delivery_snapshots" USING btree (
  "version_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "snapshot_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_vds_version IS '按版本查交付范围快照';
CREATE INDEX "idx_vds_workspace" ON "public"."version_delivery_snapshots" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_vds_workspace IS '按工作空间查快照列表';

-- ----------------------------
-- Primary Key structure for table version_delivery_snapshots
-- ----------------------------
ALTER TABLE "public"."version_delivery_snapshots" ADD CONSTRAINT "version_delivery_snapshots_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for versions
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.versions', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.versions) + 1, false);

-- ----------------------------
-- Indexes structure for table versions
-- ----------------------------
CREATE INDEX "idx_versions_project_status" ON "public"."versions" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_versions_project_status IS '按项目+状态查询版本列表（planning/active/released/archived）';
CREATE UNIQUE INDEX "idx_versions_unique_semver" ON "public"."versions" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "semver" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_versions_unique_semver IS '按项目+semver 发布用（唯一；发布后只读）';
CREATE INDEX "idx_versions_workspace" ON "public"."versions" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_versions_workspace IS '按工作空间查版本列表';

-- ----------------------------
-- Triggers structure for table versions
-- ----------------------------
CREATE TRIGGER "trg_versions_bump_version" BEFORE UPDATE ON "public"."versions"
FOR EACH ROW
EXECUTE PROCEDURE "public"."bump_version"();
-- FIXED: COMMENT ON TRIGGER trg_versions_bump_version IS 'versions: BEFORE UPDATE 乐观锁版本号 version = version + 1';
CREATE TRIGGER "trg_versions_updated_at" BEFORE UPDATE ON "public"."versions"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_versions_updated_at IS 'versions: BEFORE UPDATE 自动将 updated_at 更新为 now()';
CREATE TRIGGER "trg_version_search_sync" AFTER INSERT OR UPDATE OF "name", "description" ON "public"."versions"
FOR EACH ROW
WHEN ((new.deleted_at IS NULL))
EXECUTE PROCEDURE "public"."fn_refresh_version_search_document"();
-- FIXED: COMMENT ON TRIGGER trg_version_search_sync IS 'versions: AFTER INSERT/UPDATE 同步 ES 降级索引';
CREATE TRIGGER "trg_version_search_cleanup" AFTER UPDATE OF "deleted_at" ON "public"."versions"
FOR EACH ROW
WHEN ((new.deleted_at IS NOT NULL))
EXECUTE PROCEDURE "public"."fn_cleanup_search_document"();
-- FIXED: COMMENT ON TRIGGER trg_version_search_cleanup IS 'versions: AFTER DELETE（软删除）清理 ES 及 search_documents';

-- ----------------------------
-- Checks structure for table versions
-- ----------------------------
ALTER TABLE "public"."versions" ADD CONSTRAINT "versions_date_range" CHECK (start_date IS NULL OR end_date IS NULL OR start_date <= end_date);
ALTER TABLE "public"."versions" ADD CONSTRAINT "versions_semver_valid" CHECK (semver ~ '^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-[0-9A-Za-z\-.]+)?(\+[0-9A-Za-z\-.]+)?$'::text);
ALTER TABLE "public"."versions" ADD CONSTRAINT "versions_status_check" CHECK (status = ANY (ARRAY['planning'::text, 'active'::text, 'released'::text, 'archived'::text]));
ALTER TABLE "public"."versions" ADD CONSTRAINT "versions_checklist_limit" CHECK (jsonb_array_length(checklist) <= 50);

-- ----------------------------
-- Primary Key structure for table versions
-- ----------------------------
ALTER TABLE "public"."versions" ADD CONSTRAINT "versions_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table view_preferences
-- ----------------------------
CREATE INDEX "idx_view_prefs_user" ON "public"."view_preferences" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_view_prefs_user IS '按用户查视图偏好设置';

-- ----------------------------
-- Uniques structure for table view_preferences
-- ----------------------------
ALTER TABLE "public"."view_preferences" ADD CONSTRAINT "view_preferences_workspace_id_project_id_user_id_view_type_key" UNIQUE ("workspace_id", "project_id", "user_id", "view_type");

-- ----------------------------
-- Primary Key structure for table view_preferences
-- ----------------------------
ALTER TABLE "public"."view_preferences" ADD CONSTRAINT "view_preferences_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table webhook_logs
-- ----------------------------
CREATE INDEX "idx_webhook_logs_delivery" ON "public"."webhook_logs" USING btree (
  "delivery_id" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_webhook_logs_delivery IS '按投递 ID 查询 Webhook 日志';
CREATE INDEX "idx_webhook_logs_occurred" ON "public"."webhook_logs" USING btree (
  "occurred_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_webhook_logs_occurred IS '按发生时间倒序游标分页（管理页面）';
CREATE INDEX "idx_webhook_logs_webhook" ON "public"."webhook_logs" USING btree (
  "webhook_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "occurred_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_webhook_logs_webhook IS '按 webhook+时间查投递日志（管理页面/监控面板）';
CREATE INDEX "idx_webhook_logs_workspace" ON "public"."webhook_logs" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "occurred_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
COMMENT ON INDEX public.idx_webhook_logs_workspace IS '按工作空间查 Webhook 日志';

-- ----------------------------
-- Primary Key structure for table webhook_logs
-- ----------------------------
ALTER TABLE "public"."webhook_logs" ADD CONSTRAINT "webhook_logs_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table webhooks
-- ----------------------------
CREATE INDEX "idx_webhooks_active" ON "public"."webhooks" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "is_active" "pg_catalog"."bool_ops" ASC NULLS LAST
) WHERE is_active = true;
COMMENT ON INDEX public.idx_webhooks_active IS '按项目+启用状态查 Webhook 配置（事件投递匹配）';
CREATE INDEX "idx_webhooks_project" ON "public"."webhooks" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE project_id IS NOT NULL;
COMMENT ON INDEX public.idx_webhooks_project IS '按项目查 Webhook 列表';
CREATE INDEX "idx_webhooks_workspace" ON "public"."webhooks" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_webhooks_workspace IS '按工作空间查 Webhook 列表';

-- ----------------------------
-- Checks structure for table webhooks
-- ----------------------------
ALTER TABLE "public"."webhooks" ADD CONSTRAINT "chk_target_url_protocol" CHECK (target_url ~ '^https?://'::text);

-- ----------------------------
-- Primary Key structure for table webhooks
-- ----------------------------
ALTER TABLE "public"."webhooks" ADD CONSTRAINT "webhooks_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for workbench_configs
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.workbench_configs', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.workbench_configs) + 1, false);

-- ----------------------------
-- Indexes structure for table workbench_configs
-- ----------------------------
CREATE INDEX "idx_workbench_configs_project" ON "public"."workbench_configs" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE project_id IS NOT NULL;
COMMENT ON INDEX public.idx_workbench_configs_project IS '按项目查工作台配置';
CREATE INDEX "idx_workbench_configs_user" ON "public"."workbench_configs" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_workbench_configs_user IS '按用户查个性化工作台配置';
CREATE UNIQUE INDEX "idx_workbench_configs_user_project" ON "public"."workbench_configs" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  COALESCE(project_id, 0::bigint) "pg_catalog"."int8_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_workbench_configs_user_project IS '按用户+项目查工作台配置';

-- ----------------------------
-- Triggers structure for table workbench_configs
-- ----------------------------
CREATE TRIGGER "trg_workbench_configs_updated_at" BEFORE UPDATE ON "public"."workbench_configs"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_workbench_configs_updated_at IS 'workbench_configs: BEFORE UPDATE 自动将 updated_at 更新为 now()';

-- ----------------------------
-- Primary Key structure for table workbench_configs
-- ----------------------------
ALTER TABLE "public"."workbench_configs" ADD CONSTRAINT "workbench_configs_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for workbench_templates
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.workbench_templates', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.workbench_templates) + 1, false);

-- ----------------------------
-- Indexes structure for table workbench_templates
-- ----------------------------
CREATE INDEX "idx_workbench_templates_default" ON "public"."workbench_templates" USING btree (
  "is_default" "pg_catalog"."bool_ops" ASC NULLS LAST
);
COMMENT ON INDEX public.idx_workbench_templates_default IS '按默认状态查询工作台模板';

-- ----------------------------
-- Uniques structure for table workbench_templates
-- ----------------------------
ALTER TABLE "public"."workbench_templates" ADD CONSTRAINT "workbench_templates_slug_key" UNIQUE ("slug");

-- ----------------------------
-- Primary Key structure for table workbench_templates
-- ----------------------------
ALTER TABLE "public"."workbench_templates" ADD CONSTRAINT "workbench_templates_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Checks structure for table workspace_members
-- ----------------------------
ALTER TABLE "public"."workspace_members" ADD CONSTRAINT "workspace_members_role_check" CHECK (role = ANY (ARRAY['owner'::text, 'admin'::text, 'member'::text, 'guest'::text]));

-- ----------------------------
-- Primary Key structure for table workspace_members
-- ----------------------------
ALTER TABLE "public"."workspace_members" ADD CONSTRAINT "workspace_members_pkey" PRIMARY KEY ("workspace_id", "user_id");

-- ----------------------------
-- Auto increment value for workspaces
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.workspaces', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.workspaces) + 1, false);

-- ----------------------------
-- Indexes structure for table workspaces
-- ----------------------------
CREATE UNIQUE INDEX "uq_workspaces_slug" ON "public"."workspaces" USING btree (
  "slug" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE status <> 'archived'::text;
COMMENT ON INDEX public.uq_workspaces_slug IS 'slug 唯一（URL 路由依据；软删除排除）';

-- ----------------------------
-- Triggers structure for table workspaces
-- ----------------------------
CREATE TRIGGER "trg_workspaces_updated_at" BEFORE UPDATE ON "public"."workspaces"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_workspaces_updated_at IS 'workspaces: BEFORE UPDATE 自动将 updated_at 更新为 now()';

-- ----------------------------
-- Checks structure for table workspaces
-- ----------------------------
ALTER TABLE "public"."workspaces" ADD CONSTRAINT "workspaces_status_check" CHECK (status = ANY (ARRAY['active'::text, 'archived'::text]));

-- ----------------------------
-- Primary Key structure for table workspaces
-- ----------------------------
ALTER TABLE "public"."workspaces" ADD CONSTRAINT "workspaces_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Foreign Keys structure for table api_tokens
-- ----------------------------
ALTER TABLE "public"."api_tokens" ADD CONSTRAINT "api_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table audit_logs
-- ----------------------------
ALTER TABLE "public"."audit_logs" ADD CONSTRAINT "audit_logs_actor_id_fkey" FOREIGN KEY ("actor_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table automation_rules
-- ----------------------------
ALTER TABLE "public"."automation_rules" ADD CONSTRAINT "automation_rules_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."automation_rules" ADD CONSTRAINT "automation_rules_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."automation_rules" ADD CONSTRAINT "automation_rules_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table dashboard_snapshots
-- ----------------------------
ALTER TABLE "public"."dashboard_snapshots" ADD CONSTRAINT "dashboard_snapshots_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table dashboard_widgets
-- ----------------------------
ALTER TABLE "public"."dashboard_widgets" ADD CONSTRAINT "dashboard_widgets_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."dashboard_widgets" ADD CONSTRAINT "dashboard_widgets_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table deployment_events
-- ----------------------------
ALTER TABLE "public"."deployment_events" ADD CONSTRAINT "deployment_events_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."deployment_events" ADD CONSTRAINT "deployment_events_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table idempotency_keys
-- ----------------------------
ALTER TABLE "public"."idempotency_keys" ADD CONSTRAINT "idempotency_keys_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table intake_channels
-- ----------------------------
ALTER TABLE "public"."intake_channels" ADD CONSTRAINT "fk_intake_channel_project" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."intake_channels" ADD CONSTRAINT "fk_intake_channel_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table intake_issues
-- ----------------------------
ALTER TABLE "public"."intake_issues" ADD CONSTRAINT "fk_intake_issue_channel" FOREIGN KEY ("channel_id") REFERENCES "public"."intake_channels" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."intake_issues" ADD CONSTRAINT "fk_intake_issue_project" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE SET NULL ON UPDATE NO ACTION;
ALTER TABLE "public"."intake_issues" ADD CONSTRAINT "fk_intake_issue_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table invitations
-- ----------------------------
ALTER TABLE "public"."invitations" ADD CONSTRAINT "invitations_inviter_id_fkey" FOREIGN KEY ("inviter_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."invitations" ADD CONSTRAINT "invitations_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table issue_activities
-- ----------------------------
ALTER TABLE "public"."issue_activities" ADD CONSTRAINT "issue_activities_actor_id_fkey" FOREIGN KEY ("actor_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_activities" ADD CONSTRAINT "issue_activities_issue_id_fkey" FOREIGN KEY ("issue_id") REFERENCES "public"."issues" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_activities" ADD CONSTRAINT "issue_activities_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_activities" ADD CONSTRAINT "issue_activities_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table issue_assignees
-- ----------------------------
ALTER TABLE "public"."issue_assignees" ADD CONSTRAINT "issue_assignees_assigned_by_fkey" FOREIGN KEY ("assigned_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_assignees" ADD CONSTRAINT "issue_assignees_issue_id_fkey" FOREIGN KEY ("issue_id") REFERENCES "public"."issues" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_assignees" ADD CONSTRAINT "issue_assignees_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table issue_comments
-- ----------------------------
ALTER TABLE "public"."issue_comments" ADD CONSTRAINT "issue_comments_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_comments" ADD CONSTRAINT "issue_comments_issue_id_fkey" FOREIGN KEY ("issue_id") REFERENCES "public"."issues" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_comments" ADD CONSTRAINT "issue_comments_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "public"."issue_comments" ("id") ON DELETE SET NULL ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_comments" ADD CONSTRAINT "issue_comments_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_comments" ADD CONSTRAINT "issue_comments_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table issue_dependencies
-- ----------------------------
ALTER TABLE "public"."issue_dependencies" ADD CONSTRAINT "issue_dependencies_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_dependencies" ADD CONSTRAINT "issue_dependencies_predecessor_id_fkey" FOREIGN KEY ("predecessor_id") REFERENCES "public"."issues" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_dependencies" ADD CONSTRAINT "issue_dependencies_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_dependencies" ADD CONSTRAINT "issue_dependencies_successor_id_fkey" FOREIGN KEY ("successor_id") REFERENCES "public"."issues" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_dependencies" ADD CONSTRAINT "issue_dependencies_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table issue_labels
-- ----------------------------
ALTER TABLE "public"."issue_labels" ADD CONSTRAINT "issue_labels_issue_id_fkey" FOREIGN KEY ("issue_id") REFERENCES "public"."issues" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_labels" ADD CONSTRAINT "issue_labels_label_id_fkey" FOREIGN KEY ("label_id") REFERENCES "public"."labels" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table issue_modules
-- ----------------------------
ALTER TABLE "public"."issue_modules" ADD CONSTRAINT "issue_modules_issue_id_fkey" FOREIGN KEY ("issue_id") REFERENCES "public"."issues" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_modules" ADD CONSTRAINT "issue_modules_module_id_fkey" FOREIGN KEY ("module_id") REFERENCES "public"."modules" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table issue_relations
-- ----------------------------
ALTER TABLE "public"."issue_relations" ADD CONSTRAINT "issue_relations_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_relations" ADD CONSTRAINT "issue_relations_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_relations" ADD CONSTRAINT "issue_relations_source_issue_id_fkey" FOREIGN KEY ("source_issue_id") REFERENCES "public"."issues" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_relations" ADD CONSTRAINT "issue_relations_target_issue_id_fkey" FOREIGN KEY ("target_issue_id") REFERENCES "public"."issues" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_relations" ADD CONSTRAINT "issue_relations_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table issue_watchers
-- ----------------------------
ALTER TABLE "public"."issue_watchers" ADD CONSTRAINT "issue_watchers_issue_id_fkey" FOREIGN KEY ("issue_id") REFERENCES "public"."issues" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issue_watchers" ADD CONSTRAINT "issue_watchers_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table issues
-- ----------------------------
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_fix_version_id_fkey" FOREIGN KEY ("fix_version_id") REFERENCES "public"."versions" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_found_version_id_fkey" FOREIGN KEY ("found_version_id") REFERENCES "public"."versions" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "public"."issues" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_release_version_id_fkey" FOREIGN KEY ("release_version_id") REFERENCES "public"."versions" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_state_id_fkey" FOREIGN KEY ("state_id") REFERENCES "public"."states" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_verifier_id_fkey" FOREIGN KEY ("verifier_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table labels
-- ----------------------------
ALTER TABLE "public"."labels" ADD CONSTRAINT "labels_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."labels" ADD CONSTRAINT "labels_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."labels" ADD CONSTRAINT "labels_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table metric_adjustments
-- ----------------------------
ALTER TABLE "public"."metric_adjustments" ADD CONSTRAINT "metric_adjustments_adjusted_by_fkey" FOREIGN KEY ("adjusted_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."metric_adjustments" ADD CONSTRAINT "metric_adjustments_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."metric_adjustments" ADD CONSTRAINT "metric_adjustments_snapshot_id_fkey" FOREIGN KEY ("snapshot_id") REFERENCES "public"."metric_snapshots" ("id") ON DELETE SET NULL ON UPDATE NO ACTION;
ALTER TABLE "public"."metric_adjustments" ADD CONSTRAINT "metric_adjustments_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table metric_snapshots
-- ----------------------------
ALTER TABLE "public"."metric_snapshots" ADD CONSTRAINT "metric_snapshots_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."metric_snapshots" ADD CONSTRAINT "metric_snapshots_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table modules
-- ----------------------------
ALTER TABLE "public"."modules" ADD CONSTRAINT "modules_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."modules" ADD CONSTRAINT "modules_lead_id_fkey" FOREIGN KEY ("lead_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."modules" ADD CONSTRAINT "modules_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."modules" ADD CONSTRAINT "modules_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table notification_deliveries
-- ----------------------------
ALTER TABLE "public"."notification_deliveries" ADD CONSTRAINT "notification_deliveries_notification_id_fkey" FOREIGN KEY ("notification_id") REFERENCES "public"."notifications" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table notification_digests
-- ----------------------------
ALTER TABLE "public"."notification_digests" ADD CONSTRAINT "notification_digests_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."notification_digests" ADD CONSTRAINT "notification_digests_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table notification_preferences
-- ----------------------------
ALTER TABLE "public"."notification_preferences" ADD CONSTRAINT "notification_preferences_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."notification_preferences" ADD CONSTRAINT "notification_preferences_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table notifications
-- ----------------------------
ALTER TABLE "public"."notifications" ADD CONSTRAINT "notifications_actor_id_fkey" FOREIGN KEY ("actor_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."notifications" ADD CONSTRAINT "notifications_recipient_id_fkey" FOREIGN KEY ("recipient_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."notifications" ADD CONSTRAINT "notifications_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table password_reset_tokens
-- ----------------------------
ALTER TABLE "public"."password_reset_tokens" ADD CONSTRAINT "password_reset_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table project_sequences
-- ----------------------------
ALTER TABLE "public"."project_sequences" ADD CONSTRAINT "project_sequences_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table projects
-- ----------------------------
ALTER TABLE "public"."projects" ADD CONSTRAINT "projects_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."projects" ADD CONSTRAINT "projects_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table project_members
-- ----------------------------
ALTER TABLE "public"."project_members" ADD CONSTRAINT "project_members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."project_members" ADD CONSTRAINT "project_members_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."project_members" ADD CONSTRAINT "project_members_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."project_members" ADD CONSTRAINT "project_members_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table recent_items
-- ----------------------------
ALTER TABLE "public"."recent_items" ADD CONSTRAINT "recent_items_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."recent_items" ADD CONSTRAINT "recent_items_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."recent_items" ADD CONSTRAINT "recent_items_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table risk_alerts
-- ----------------------------
ALTER TABLE "public"."risk_alerts" ADD CONSTRAINT "risk_alerts_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."risk_alerts" ADD CONSTRAINT "risk_alerts_resolved_by_fkey" FOREIGN KEY ("resolved_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."risk_alerts" ADD CONSTRAINT "risk_alerts_rule_id_fkey" FOREIGN KEY ("rule_id") REFERENCES "public"."risk_rules" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."risk_alerts" ADD CONSTRAINT "risk_alerts_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table risk_rules
-- ----------------------------
ALTER TABLE "public"."risk_rules" ADD CONSTRAINT "risk_rules_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."risk_rules" ADD CONSTRAINT "risk_rules_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table rule_executions
-- ----------------------------
ALTER TABLE "public"."rule_executions" ADD CONSTRAINT "rule_executions_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."rule_executions" ADD CONSTRAINT "rule_executions_rule_id_fkey" FOREIGN KEY ("rule_id") REFERENCES "public"."automation_rules" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."rule_executions" ADD CONSTRAINT "rule_executions_trigger_event_id_fkey" FOREIGN KEY ("trigger_event_id") REFERENCES "public"."domain_events" ("id") ON DELETE SET NULL ON UPDATE NO ACTION;
ALTER TABLE "public"."rule_executions" ADD CONSTRAINT "rule_executions_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table search_bookmarks
-- ----------------------------
ALTER TABLE "public"."search_bookmarks" ADD CONSTRAINT "search_bookmarks_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."search_bookmarks" ADD CONSTRAINT "search_bookmarks_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."search_bookmarks" ADD CONSTRAINT "search_bookmarks_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table search_documents
-- ----------------------------
ALTER TABLE "public"."search_documents" ADD CONSTRAINT "search_documents_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."search_documents" ADD CONSTRAINT "search_documents_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table search_history
-- ----------------------------
ALTER TABLE "public"."search_history" ADD CONSTRAINT "search_history_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."search_history" ADD CONSTRAINT "search_history_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table sprint_issues
-- ----------------------------
ALTER TABLE "public"."sprint_issues" ADD CONSTRAINT "sprint_issues_added_by_fkey" FOREIGN KEY ("added_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."sprint_issues" ADD CONSTRAINT "sprint_issues_issue_id_fkey" FOREIGN KEY ("issue_id") REFERENCES "public"."issues" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."sprint_issues" ADD CONSTRAINT "sprint_issues_sprint_id_fkey" FOREIGN KEY ("sprint_id") REFERENCES "public"."sprints" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table sprint_snapshots
-- ----------------------------
ALTER TABLE "public"."sprint_snapshots" ADD CONSTRAINT "sprint_snapshots_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."sprint_snapshots" ADD CONSTRAINT "sprint_snapshots_sprint_id_fkey" FOREIGN KEY ("sprint_id") REFERENCES "public"."sprints" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."sprint_snapshots" ADD CONSTRAINT "sprint_snapshots_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table sprints
-- ----------------------------
ALTER TABLE "public"."sprints" ADD CONSTRAINT "sprints_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."sprints" ADD CONSTRAINT "sprints_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."sprints" ADD CONSTRAINT "sprints_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."sprints" ADD CONSTRAINT "sprints_version_id_fkey" FOREIGN KEY ("version_id") REFERENCES "public"."versions" ("id") ON DELETE SET NULL ON UPDATE NO ACTION;
ALTER TABLE "public"."sprints" ADD CONSTRAINT "sprints_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table state_transitions
-- ----------------------------
ALTER TABLE "public"."state_transitions" ADD CONSTRAINT "state_transitions_from_state_id_fkey" FOREIGN KEY ("from_state_id") REFERENCES "public"."states" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."state_transitions" ADD CONSTRAINT "state_transitions_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."state_transitions" ADD CONSTRAINT "state_transitions_to_state_id_fkey" FOREIGN KEY ("to_state_id") REFERENCES "public"."states" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."state_transitions" ADD CONSTRAINT "state_transitions_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table states
-- ----------------------------
ALTER TABLE "public"."states" ADD CONSTRAINT "states_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."states" ADD CONSTRAINT "states_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table time_logs
-- ----------------------------
ALTER TABLE "public"."time_logs" ADD CONSTRAINT "time_logs_issue_id_fkey" FOREIGN KEY ("issue_id") REFERENCES "public"."issues" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."time_logs" ADD CONSTRAINT "time_logs_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."time_logs" ADD CONSTRAINT "time_logs_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."time_logs" ADD CONSTRAINT "time_logs_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table version_delivery_snapshots
-- ----------------------------
ALTER TABLE "public"."version_delivery_snapshots" ADD CONSTRAINT "version_delivery_snapshots_version_id_fkey" FOREIGN KEY ("version_id") REFERENCES "public"."versions" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."version_delivery_snapshots" ADD CONSTRAINT "version_delivery_snapshots_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table versions
-- ----------------------------
ALTER TABLE "public"."versions" ADD CONSTRAINT "versions_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."versions" ADD CONSTRAINT "versions_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."versions" ADD CONSTRAINT "versions_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table webhook_logs
-- ----------------------------
ALTER TABLE "public"."webhook_logs" ADD CONSTRAINT "fk_webhook" FOREIGN KEY ("webhook_id") REFERENCES "public"."webhooks" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table webhooks
-- ----------------------------
ALTER TABLE "public"."webhooks" ADD CONSTRAINT "fk_project" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."webhooks" ADD CONSTRAINT "fk_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table workbench_configs
-- ----------------------------
ALTER TABLE "public"."workbench_configs" ADD CONSTRAINT "workbench_configs_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."workbench_configs" ADD CONSTRAINT "workbench_configs_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."workbench_configs" ADD CONSTRAINT "workbench_configs_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table workspace_members
-- ----------------------------
ALTER TABLE "public"."workspace_members" ADD CONSTRAINT "workspace_members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."workspace_members" ADD CONSTRAINT "workspace_members_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table workspaces
-- ----------------------------
ALTER TABLE "public"."workspaces" ADD CONSTRAINT "workspaces_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Table structure for pages
-- ----------------------------
DROP TABLE IF EXISTS "public"."pages";
CREATE TABLE "public"."pages" (
  "id" int8 NOT NULL GENERATED ALWAYS AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1
),
  "public_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "workspace_id" int8 NOT NULL,
  "project_id" int8 NOT NULL,
  "name" text COLLATE "pg_catalog"."default" NOT NULL,
  "description_json" jsonb,
  "description_html" text COLLATE "pg_catalog"."default",
  "description_stripped" text COLLATE "pg_catalog"."default",
  "parent_id" int8,
  "sort_order" float8 NOT NULL DEFAULT 65535,
  "created_by" int8 NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT now(),
  "updated_at" timestamptz(6) NOT NULL DEFAULT now(),
  "deleted_at" timestamptz(6),
  "version" int4 NOT NULL DEFAULT 1
)
;
COMMENT ON TABLE public.pages IS '页面表（项目内自定义页面；TipTap 富文本内容；支持嵌入看板/指标）';
COMMENT ON COLUMN public.pages.id IS '主键 ID';
COMMENT ON COLUMN public.pages.workspace_id IS '工作空间 FK（RLS 依据）';
COMMENT ON COLUMN public.pages.project_id IS '所属项目 FK';
COMMENT ON COLUMN public.pages.title IS '页面标题';
COMMENT ON COLUMN public.pages.content_json IS 'TipTap 编辑器内容 JSON（ProseMirror 格式；富文本 + 嵌入卡片）';
COMMENT ON COLUMN public.pages.content_html IS '从 content_json 渲染的 HTML（展示层）';
COMMENT ON COLUMN public.pages.cover_image IS '封面图片附件 ID';
COMMENT ON COLUMN public.pages.icon IS '页面图标（Emoji / Lucide 图标）';
COMMENT ON COLUMN public.pages.parent_id IS '父页面 FK（pages.id）；层级≤3';
COMMENT ON COLUMN public.pages.sort_order IS '同级排序权重';
COMMENT ON COLUMN public.pages.is_pinned IS '是否置顶: true=在侧边栏始终展示';
COMMENT ON COLUMN public.pages.created_by IS '创建人 FK';
COMMENT ON COLUMN public.pages.created_at IS '创建时间';
COMMENT ON COLUMN public.pages.updated_at IS '修改时间（触发器自动维护）';
COMMENT ON COLUMN public.pages.deleted_at IS '软删除时间戳';

-- ----------------------------
-- Records of pages
-- ----------------------------

-- ----------------------------
-- Indexes structure for table pages
-- ----------------------------
CREATE INDEX "idx_pages_project" ON "public"."pages" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_pages_project IS '按项目查知识页面列表';
CREATE INDEX "idx_pages_parent" ON "public"."pages" USING btree (
  "parent_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND parent_id IS NOT NULL;
COMMENT ON INDEX public.idx_pages_parent IS '按 parent_id 查询子页面（树展开）';
CREATE INDEX "idx_pages_project_sort" ON "public"."pages" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."float8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_pages_project_sort IS '按项目+排序权重（拖拽排序）';
CREATE UNIQUE INDEX "idx_pages_public_id" ON "public"."pages" USING btree (
  "public_id" "pg_catalog"."uuid_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
COMMENT ON INDEX public.idx_pages_public_id IS '按 public_id 查询页面（API/URL 路由）';

-- ----------------------------
-- Triggers structure for table pages
-- ----------------------------
CREATE TRIGGER "trg_pages_updated_at" BEFORE UPDATE ON "public"."pages"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
-- FIXED: COMMENT ON TRIGGER trg_pages_updated_at IS 'pages: BEFORE UPDATE 自动将 updated_at 更新为 now()';

-- ----------------------------
-- Primary Key structure for table pages
-- ----------------------------
ALTER TABLE "public"."pages" ADD CONSTRAINT "pages_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Auto increment value for pages
-- ----------------------------
SELECT setval(pg_get_serial_sequence('public.pages', 'id'), (SELECT COALESCE(MAX(id), 0) FROM public.pages) + 1, false);

-- ----------------------------
-- Foreign Keys structure for table pages
-- ----------------------------
ALTER TABLE "public"."pages" ADD CONSTRAINT "pages_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."pages" ADD CONSTRAINT "pages_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "public"."pages" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "public"."pages" ADD CONSTRAINT "pages_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "public"."pages" ADD CONSTRAINT "pages_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- ============================================================================
-- 数据库注释补丁 (合并自 add-comments-patch.sql)
-- 用途: 为所有表、字段、索引、触发器
-- ============================================================================

-- ============================================================================
-- 补丁名称: ydsz-plane 全库注释补丁 (add-comments-patch.sql)
-- 生成日期: 2026-08-08
-- 目标数据库: PostgreSQL 18, schema: public
-- 用途: 为 ydsz-plane 项目全部 58 张表及其字段、索引、触发器添加中文注释，
--       实现 100% 注释覆盖率。
--
-- 使用方法:
--   psql -U ydsz_app -d ydsz-plane -f add-comments-patch.sql
--
-- 注意事项:
--   1. 已存在注释的表/issue_comments, notification_preferences, notifications, versions
--      仅补充缺失字段的注释，不覆盖已有内容。
--   2. 所有 COMMENT 语句均为幂等操作，可重复执行。
--   3. 部分索引注释中包含 WHERE 条件含义说明。
-- ============================================================================

-- ============================================================================
-- 核心工作项域
-- ============================================================================

-- ============================================================
-- 表: issues — 核心工作项表
-- ============================================================


-- ============================================================
-- 表: labels — 标签表
-- ============================================================


-- ============================================================
-- 表: modules — 模块/组件表
-- ============================================================


-- ============================================================
-- 表: issue_labels — 工作项-标签关联表
-- ============================================================


-- ============================================================
-- 表: issue_modules — 工作项-模块关联表
-- ============================================================


-- ============================================================
-- 表: issue_assignees — 工作项指派人员表
-- ============================================================


-- ============================================================
-- 表: issue_watchers — 工作项关注者表
-- ============================================================


-- ============================================================
-- 表: issue_comments — 工作项评论表（已有注释，补充缺失字段）
-- ============================================================

-- ============================================================
-- 表: issue_activities — 工作项活动表（按月分区）
-- ============================================================


-- ============================================================
-- 表: issue_dependencies — 工作项依赖关系表
-- ============================================================


-- ============================================================
-- 表: issue_relations — 工作项关联关系表
-- ============================================================


-- ============================================================
-- 表: time_logs — 工时记录表
-- ============================================================


-- ============================================================
-- 表: attachments — 附件表
-- ============================================================


-- ============================================================================
-- 项目管理域
-- ============================================================================

-- ============================================================
-- 表: projects — 项目表
-- ============================================================



-- ============================================================
-- 表: sprints — 迭代/冲刺表
-- ============================================================


-- ============================================================
-- 表: sprint_issues — 迭代-工作项关联表
-- ============================================================


-- ============================================================
-- 表: sprint_snapshots — 迭代快照表
-- ============================================================


-- ============================================================
-- 表: versions — 版本发布表（已有 version 字段注释，补充其他字段）
-- ============================================================


-- ============================================================
-- 表: version_delivery_snapshots — 版本交付快照表
-- ============================================================


-- ============================================================================
-- 工作流域
-- ============================================================================

-- ============================================================
-- 表: states — 状态表
-- ============================================================


-- ============================================================
-- 表: state_transitions — 状态流转规则表
-- ============================================================


-- ============================================================================
-- 租户与权限域
-- ============================================================================

-- ============================================================
-- 表: users — 用户表
-- ============================================================


-- ============================================================
-- 表: workspaces — 工作空间/租户表
-- ============================================================


-- ============================================================
-- 表: invitations — 邀请表
-- ============================================================


-- ============================================================
-- 表: workspace_members — 工作空间成员表
-- ============================================================




-- ============================================================
-- 表: sprints — 迭代/冲刺表
-- ============================================================


-- ============================================================
-- 表: sprint_issues — 迭代-工作项关联表
-- ============================================================


-- ============================================================
-- 表: sprint_snapshots — 迭代快照表
-- ============================================================


-- ============================================================
-- 表: versions — 版本发布表（已有 version 字段注释，补充其他字段）
-- ============================================================


-- ============================================================
-- 表: version_delivery_snapshots — 版本交付快照表
-- ============================================================


-- ============================================================================
-- 工作流域
-- ============================================================================

-- ============================================================
-- 表: states — 状态表
-- ============================================================


-- ============================================================
-- 表: state_transitions — 状态流转规则表
-- ============================================================


-- ============================================================================
-- 租户与权限域
-- ============================================================================

-- ============================================================
-- 表: users — 用户表
-- ============================================================


-- ============================================================
-- 表: workspaces — 工作空间/租户表
-- ============================================================


-- ============================================================
-- 表: invitations — 邀请表
-- ============================================================


-- ============================================================
-- 表: workspace_members — 工作空间成员表
-- ============================================================



-- ============================================================================
-- 通知域
-- ============================================================================

-- ============================================================
-- 表: notifications — 通知消息表（已有注释，补充缺失字段）
-- ============================================================

-- ============================================================
-- 表: notification_preferences — 通知偏好表（已有表注释，补充字段）
-- ============================================================

-- ============================================================
-- 表: notification_deliveries — 通知投递记录表
-- ============================================================


-- ============================================================
-- 表: notification_digests — 通知摘要/聚合
-- ============================================================


-- ============================================================================
-- 入口工单域
-- ============================================================================

-- ============================================================
-- 表: intake_channels — 入口渠道表
-- ============================================================


-- ============================================================
-- 表: intake_issues — 入口工单表
-- ============================================================



-- ============================================================================
-- 自动化域
-- ============================================================================

-- ============================================================
-- 表: automation_rules — 自动化规则表
-- ============================================================


-- ============================================================
-- 表: automation_templates — 自动化模板表
-- ============================================================


-- ============================================================
-- 表: rule_executions — 规则执行日志表
-- ============================================================


-- ============================================================================
-- 仪表盘域
-- ============================================================================

-- ============================================================
-- 表: dashboard_widgets — 仪表盘组件表
-- ============================================================


-- ============================================================
-- 表: dashboard_templates — 仪表盘模板表
-- ============================================================


-- ============================================================
-- 表: dashboard_snapshots — 仪表盘快照表
-- ============================================================


-- ============================================================================
-- 功能区
-- ============================================================================

-- ============================================================
-- 表: workbench_configs — 工作台配置表
-- ============================================================


-- ============================================================
-- 表: workbench_templates — 工作台模板表
-- ============================================================


-- ============================================================
-- 表: view_preferences — 视图偏好表
-- ============================================================


-- ============================================================
-- 表: recent_items — 最近访问记录表
-- ============================================================


-- ============================================================
-- 表: search_bookmarks — 搜索书签表
-- ============================================================


-- ============================================================
-- 表: search_documents — 搜索文档索引表
-- ============================================================


-- ============================================================
-- 表: search_history — 搜索历史表
-- ============================================================



-- ============================================================================
-- 风险与度量域
-- ============================================================================

-- ============================================================
-- 表: risk_rules — 风险规则表
-- ============================================================


-- ============================================================
-- 表: risk_alerts — 风险告警表
-- ============================================================


-- ============================================================
-- 表: metric_snapshots — 指标快照表
-- ============================================================


-- ============================================================
-- 表: metric_adjustments — 指标调整记录表
-- ============================================================


-- ============================================================================
-- 集成与扩展域
-- ============================================================================

-- ============================================================
-- 表: webhooks — Webhook 配置表
-- ============================================================


-- ============================================================
-- 表: webhook_logs — Webhook 日志表
-- ============================================================


-- ============================================================
-- 表: api_tokens — API 令牌表
-- ============================================================


-- ============================================================
-- 表: deployment_events — 部署事件表
-- ============================================================


-- ============================================================================
-- 系统基础设施域
-- ============================================================================

-- ============================================================
-- 表: domain_events — 领域事件表（Outbox 模式）
-- ============================================================


-- ============================================================
-- 表: idempotency_keys — API 幂等键表
-- ============================================================


-- ============================================================
-- 表: password_reset_tokens — 密码重置令牌表
-- ============================================================


-- ============================================================
-- 表: audit_logs — 审计日志表
-- ============================================================


-- ============================================================
-- 表: schema_migrations — 数据库迁移版本表
-- ============================================================


-- ============================================================
-- 表: project_sequences — 项目序列发号器表
-- ============================================================




-- ============================================================================
-- 触发器注释（统一区块）
-- ============================================================================

-- ============================================================
-- updated_at 自动维护触发器（适用于所有含 updated_at 的表）
-- ============================================================

-- ============================================================
-- 乐观锁版本号自增触发器
-- ============================================================

-- ============================================================
-- 搜索索引同步触发器
-- ============================================================

-- ============================================================
-- 最近访问时间更新触发器
-- ============================================================

-- ============================================================================
-- 部分索引注释（说明 WHERE 条件含义）
-- ============================================================================

-- ================ issues 表索引 ================

-- ================ sprints 表索引 ================

-- ================ versions 表索引 ================

-- ================ issues 关联表索引 ================

-- ================ notification 模块索引 ================

-- ================ search 模块索引 ================

-- ================ 其他模块索引 ================

-- ============================================================================
-- 模块二: S11 性能优化索引回滚脚本（含原 0020_s11_perf_indexes.down.sql）
-- 用途: 回滚 S11 引入的覆盖索引（仅在需要时执行此区块）
-- ============================================================================

-- DROP INDEX CONCURRENTLY IF EXISTS "public"."idx_automation_rules_trigger_active";
-- DROP INDEX CONCURRENTLY IF EXISTS "public"."idx_rule_executions_rule_created";
-- DROP INDEX CONCURRENTLY IF EXISTS "public"."idx_rule_executions_idempotent";
-- DROP INDEX CONCURRENTLY IF EXISTS "public"."idx_metric_snap_trend";
-- DROP INDEX CONCURRENTLY IF EXISTS "public"."idx_deployment_dora";

-- ============================================================================
-- 模块三: 注释覆盖率检查函数（含原 check-comments.sql）
-- 用途: CI 流水线质量门禁。调用方式: SELECT public.ydsz_check_comment_coverage();
-- 返回: RAISE NOTICE 输出报告；存在缺失时 RAISE EXCEPTION
-- ============================================================================

CREATE OR REPLACE FUNCTION public.ydsz_check_comment_coverage()
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    v_uncommented_tables  int := 0;
    v_uncommented_columns  int := 0;
    v_uncommented_triggers int := 0;
    v_total_tables         int := 0;
    v_total_columns        int := 0;
    v_total_triggers       int := 0;
    v_rec                  record;
BEGIN
    -- 1. 检查未注释的表
    SELECT count(*) INTO v_total_tables
    FROM information_schema.tables t
    WHERE t.table_schema = 'public'
      AND t.table_type = 'BASE TABLE'
      AND t.table_name NOT IN ('schema_migrations');

    FOR v_rec IN
        SELECT t.table_name
        FROM information_schema.tables t
        LEFT JOIN pg_catalog.pg_description d
            ON d.objoid = (quote_ident(t.table_schema)||'.'||quote_ident(t.table_name))::regclass::oid
           AND d.objsubid = 0
        WHERE t.table_schema = 'public'
          AND t.table_type = 'BASE TABLE'
          AND t.table_name NOT IN ('schema_migrations')
          AND d.description IS NULL
        ORDER BY t.table_name
    LOOP
        v_uncommented_tables := v_uncommented_tables + 1;
        RAISE NOTICE '❌ 缺表注释: %', v_rec.table_name;
    END LOOP;

    -- 2. 检查未注释的字段
    SELECT count(*) INTO v_total_columns
    FROM information_schema.columns c
    WHERE c.table_schema = 'public'
      AND c.table_name NOT IN ('schema_migrations');

    FOR v_rec IN
        SELECT c.table_name, c.column_name
        FROM information_schema.columns c
        LEFT JOIN pg_catalog.pg_description d
            ON d.objoid = (quote_ident(c.table_schema)||'.'||quote_ident(c.table_name))::regclass::oid
           AND d.objsubid = c.ordinal_position
        WHERE c.table_schema = 'public'
          AND c.table_name NOT IN ('schema_migrations')
          AND d.description IS NULL
        ORDER BY c.table_name, c.column_name
    LOOP
        v_uncommented_columns := v_uncommented_columns + 1;
        RAISE NOTICE '❌ 缺字段注释: %.%', v_rec.table_name, v_rec.column_name;
    END LOOP;

    -- 3. 检查未注释的触发器
    SELECT count(*) INTO v_total_triggers
    FROM information_schema.triggers t
    WHERE t.trigger_schema = 'public';

    FOR v_rec IN
        SELECT t.trigger_name, t.event_object_table
        FROM information_schema.triggers t
        LEFT JOIN pg_catalog.pg_description d
            ON d.objoid = (quote_ident(t.trigger_schema)||'.'||quote_ident(t.event_object_table))::regclass::oid
           AND d.classoid = 'pg_trigger'::regclass::oid
           AND d.objsubid = (
               SELECT tgfoid FROM pg_trigger tg
               WHERE tg.tgname = t.trigger_name
                 AND tg.tgrelid = (quote_ident(t.trigger_schema)||'.'||quote_ident(t.event_object_table))::regclass::oid
               LIMIT 1
           )
        WHERE t.trigger_schema = 'public'
          AND d.description IS NULL
        ORDER BY t.event_object_table, t.trigger_name
    LOOP
        v_uncommented_triggers := v_uncommented_triggers + 1;
        RAISE NOTICE '❌ 缺触发器注释: % (表: %)', v_rec.trigger_name, v_rec.event_object_table;
    END LOOP;

    -- 4. 汇总
    RAISE NOTICE '============================================';
    RAISE NOTICE '📊 注释覆盖率报告 (ydsz-plane)';
    RAISE NOTICE '============================================';
    RAISE NOTICE '表注释: % / (% %)',
        v_total_tables - v_uncommented_tables,
        v_total_tables,
        round((v_total_tables - v_uncommented_tables)::numeric / nullif(v_total_tables,0) * 100, 1);
    RAISE NOTICE '字段注释: % / (% %)',
        v_total_columns - v_uncommented_columns,
        v_total_columns,
        round((v_total_columns - v_uncommented_columns)::numeric / nullif(v_total_columns,0) * 100, 1);
    RAISE NOTICE '触发器注释: % / (% %)',
        v_total_triggers - v_uncommented_triggers,
        v_total_triggers,
        round((v_total_triggers - v_uncommented_triggers)::numeric / nullif(v_total_triggers,0) * 100, 1);
    RAISE NOTICE '============================================';

    -- 5. 质量门禁
    IF v_uncommented_tables > 0 OR v_uncommented_columns > 0 OR v_uncommented_triggers > 0 THEN
        RAISE EXCEPTION '发现 % 张表缺注释、% 个字段缺注释、% 个触发器缺注释',
            v_uncommented_tables, v_uncommented_columns, v_uncommented_triggers;
    ELSE
        RAISE NOTICE '✅ 全部通过！注释覆盖率 100%%';
    END IF;
END $$;

COMMENT ON FUNCTION public.ydsz_check_comment_coverage() IS 'CI 质量门禁: 检查 public schema 下所有表/字段/触发器注释覆盖率';

-- ============================================================================
-- 模块四: RLS 租户隔离策略（含原 0021_rls_tenant_isolation.up.sql）
-- 用途: 为所有带 workspace_id 的表启用 PostgreSQL 原生行级安全
-- 策略: 使用 current_setting('app.workspace_id') 做等值过滤
-- 参考: 等保三级、OWASP Database Security Cheat Sheet
-- ============================================================================

DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY[
        'workspaces', 'workspace_members', 'projects', 'modules', 'labels',
        'states', 'state_transitions', 'issues', 'issue_assignees',
        'issue_labels', 'issue_modules', 'issue_activities', 'issue_comments',
        'issue_relations', 'issue_dependencies', 'sprints', 'sprint_issues',
        'sprint_snapshots', 'versions', 'automation_rules', 'rule_executions',
        'automation_templates', 'notifications', 'notification_preferences',
        'api_tokens', 'attachments', 'invitations', 'intake_channels',
        'intake_issues', 'workbench_configs', 'search_documents',
        'dashboard_widgets', 'dashboard_snapshots', 'metric_snapshots',
        'audit_logs', 'domain_events', 'webhooks', 'webhook_logs',
        'view_preferences', 'workbench_templates'
    ];
BEGIN
    FOREACH tbl IN ARRAY tables
    LOOP
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = tbl) THEN
            EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', tbl);
            IF NOT EXISTS (
                SELECT 1 FROM pg_policies
                WHERE schemaname = 'public' AND tablename = tbl AND policyname = 'tenant_isolation'
            ) THEN
                EXECUTE format(
                    'CREATE POLICY tenant_isolation ON %I USING (workspace_id::text = current_setting(''app.workspace_id'', true))',
                    tbl
                );
            END IF;
        END IF;
    END LOOP;
END
$$;

-- ============================================================================
-- 模块五: RLS 回滚脚本（含原 0021_rls_tenant_isolation.down.sql）
-- 用途: 禁用所有表的 RLS（仅在需要时取消注释执行）
-- ============================================================================

/*
DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY[
        'workspaces', 'workspace_members', 'projects', 'modules', 'labels',
        'states', 'state_transitions', 'issues', 'issue_assignees',
        'issue_labels', 'issue_modules', 'issue_activities', 'issue_comments',
        'issue_relations', 'issue_dependencies', 'sprints', 'sprint_issues',
        'sprint_snapshots', 'versions', 'automation_rules', 'rule_executions',
        'automation_templates', 'notifications', 'notification_preferences',
        'api_tokens', 'attachments', 'invitations', 'intake_channels',
        'intake_issues', 'workbench_configs', 'search_documents',
        'dashboard_widgets', 'dashboard_snapshots', 'metric_snapshots',
        'audit_logs', 'domain_events', 'webhooks', 'webhook_logs',
        'view_preferences', 'workbench_templates'
    ];
BEGIN
    FOREACH tbl IN ARRAY tables
    LOOP
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = tbl) THEN
            EXECUTE format('ALTER TABLE %I DISABLE ROW LEVEL SECURITY', tbl);
        END IF;
    END LOOP;
END
$$;
*/

-- ============================================================================
-- 全部 schema 定义 + 注释 + 辅助函数 + RLS 完毕。
-- 调用 ydsz_check_comment_coverage() 验证注释完整性。
-- ============================================================================


-- ============================================================================
-- 以下为增量迁移整合（原 0023_rbac_roles_and_permissions / 0024_work_item_split /
--                0025_work_item_archive / 0025_epic_and_modules /
--                0026_work_item_migration 五个迁移的 up 内容合并）。
-- 调整说明：
--   1. 所有 CREATE TABLE 改为 CREATE TABLE IF NOT EXISTS，保证幂等。
--   2. modules 表原本缺少 public_id 列与 RLS，通过 ALTER 补全。
--   3. workspace_members 旧 role 约束 (owner/admin/member/guest) 替换为新版
--      (admin/owner/pm/po/techlead/qalead/dev/guest)。
-- ============================================================================


-- ============================================================================
-- [原 0023] RBAC 体系重构 —— 角色/权限 DB 化
-- ============================================================================

-- -----------------------------------------------------------------------------
-- 1. roles 表
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.roles (
    slug        VARCHAR(20)  PRIMARY KEY,
    name        VARCHAR(50)  NOT NULL UNIQUE,
    description TEXT         NOT NULL DEFAULT '',
    level       SMALLINT     NOT NULL,
    is_system   BOOLEAN      NOT NULL DEFAULT true,
    icon        VARCHAR(20)  DEFAULT '',
    sort_order  SMALLINT     DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMENT ON TABLE  public.roles           IS '工作空间级角色定义（全局共享，通过 role_permissions 表热更新权限映射）';
COMMENT ON COLUMN public.roles.slug      IS '角色枚举标识：admin/owner/pm/po/techlead/qalead/dev/guest';
COMMENT ON COLUMN public.roles.level     IS '角色层级：admin=100(系统) owner=80(空间) pm/po/techlead/qalead=50 dev=30 guest=10';
COMMENT ON COLUMN public.roles.is_system IS '是否系统内置角色；true=不可删改 slug，仅能改 name/description/权限矩阵';

-- -----------------------------------------------------------------------------
-- 2. role_permissions 表
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.role_permissions (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    role_slug       VARCHAR(20)  NOT NULL REFERENCES public.roles(slug) ON DELETE CASCADE,
    permission_code VARCHAR(64)  NOT NULL,
    description     TEXT         DEFAULT '',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (role_slug, permission_code)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON public.role_permissions (role_slug);
CREATE INDEX IF NOT EXISTS idx_role_permissions_perm ON public.role_permissions (permission_code);

COMMENT ON TABLE  public.role_permissions                  IS '角色-权限映射表（一张表承载全部 RBAC 矩阵，支持运行时增删权限）';
COMMENT ON COLUMN public.role_permissions.role_slug       IS '关联 roles.slug；CASCADE 删除保证一致性';
COMMENT ON COLUMN public.role_permissions.permission_code IS '权限点标识，与 internal/auth/rbac.go 的 PermXxx 常量严格对齐';

-- -----------------------------------------------------------------------------
-- 3. workspace_members 表结构增强（追加列 + 替换 role 约束）
-- -----------------------------------------------------------------------------
ALTER TABLE public.workspace_members
    ADD COLUMN IF NOT EXISTS is_active  BOOLEAN     NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS created_by BIGINT      REFERENCES public.users(id),
    ADD COLUMN IF NOT EXISTS created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS updated_at  TIMESTAMPTZ NOT NULL DEFAULT now();

-- 替换旧的 role CHECK 约束（旧: owner/admin/member/guest → 新: 8 级 RBAC）
ALTER TABLE public.workspace_members DROP CONSTRAINT IF EXISTS workspace_members_role_check;
ALTER TABLE public.workspace_members DROP CONSTRAINT IF EXISTS chk_workspace_member_role;
ALTER TABLE public.workspace_members
    ADD CONSTRAINT chk_workspace_member_role
    CHECK (role IN ('admin','owner','pm','po','techlead','qalead','dev','guest'));

COMMENT ON COLUMN public.workspace_members.is_active  IS '成员激活状态；false=暂停访问';
COMMENT ON COLUMN public.workspace_members.created_by IS '添加人（邀请人）';

-- -----------------------------------------------------------------------------
-- 4. 种子数据 —— 8 个系统角色 + 完整权限矩阵
-- -----------------------------------------------------------------------------
INSERT INTO public.roles (slug, name, description, level, is_system, icon, sort_order) VALUES
    ('admin',    '系统管理员',    '平台最高权限。可管理所有工作空间、任意进入某个空间行使空间管理员权力、管理系统通用配置（SMTP/SSO/安全策略）。', 100, true, '🛡️', 1),
    ('owner',    '空间管理员',    '工作空间最高权限。可管理工作空间下所有事项：项目、版本、迭代、需求、任务、缺陷、成员、设置、审计日志。可删除空间或转移所有权。',  80, true, '👑', 2),
    ('pm',       '项目经理',      '项目全生命周期负责人：创建/归档项目、管理迭代、查看效能报表、管理工作项状态。',  50, true, '📋', 3),
    ('po',       '产品经理',      '需求侧负责人：创建/编辑需求、设置优先级与验收标准、管理产品路线图。',             50, true, '🎯', 4),
    ('techlead', '技术经理',      '技术侧负责人：管理 Sprint 排期、自动化规则、Webhook、效能度量与代码集成。',       50, true, '🛠️', 5),
    ('qalead',   '测试经理',      '质量侧负责人：创建/编辑缺陷、管理缺陷分类与严重度、查看缺陷分析报表。',           50, true, '🔍', 6),
    ('dev',      '开发',          '执行者：创建/编辑分配给自己的工作项、更新状态、记录工时、参与迭代。',              30, true, '💻', 7),
    ('guest',    '访客',          '只读协作者：浏览指定项目、添加评论，无任何编辑与管理权限。',                     10, true, '👁️', 8)
ON CONFLICT (slug) DO UPDATE SET
    name        = EXCLUDED.name,
    description = EXCLUDED.description,
    level       = EXCLUDED.level,
    icon        = EXCLUDED.icon,
    sort_order  = EXCLUDED.sort_order,
    updated_at  = now();

-- owner = 空间级全部权限
INSERT INTO public.role_permissions (role_slug, permission_code) VALUES
    ('owner','workspace:read'),            ('owner','workspace:update'),           ('owner','workspace:delete'),
    ('owner','workspace:transfer'),        ('owner','project:read'),               ('owner','project:create'),
    ('owner','project:update'),            ('owner','project:delete'),             ('owner','issue:read'),
    ('owner','issue:create'),              ('owner','issue:edit_own'),             ('owner','issue:edit_all'),
    ('owner','issue:delete'),              ('owner','issue:transition'),           ('owner','issue:reassign'),
    ('owner','issue:change_priority'),     ('owner','issue:manage_sprint'),        ('owner','member:invite'),
    ('owner','member:remove'),             ('owner','member:change_role'),         ('owner','sprint:read'),
    ('owner','sprint:create'),             ('owner','sprint:update'),              ('owner','sprint:delete'),
    ('owner','sprint:lifecycle'),          ('owner','sprint:plan'),                ('owner','version:read'),
    ('owner','version:create'),            ('owner','version:update'),             ('owner','version:release'),
    ('owner','version:delete'),            ('owner','defect:create'),              ('owner','qa:report'),
    ('owner','analytics:read'),            ('owner','analytics:export'),           ('owner','automation:manage'),
    ('owner','deploy:report'),             ('owner','audit:read'),                 ('owner','webhook:manage'),
    ('owner','intake:manage'),             ('owner','pages:manage'),               ('owner','comment:moderate'),
    ('owner','relation:manage'),           ('owner','field:edit_severity'),        ('owner','field:edit_effort'),
    ('owner','field:edit_deadline'),       ('owner','menu:settings'),              ('owner','menu:audit')
ON CONFLICT (role_slug, permission_code) DO NOTHING;

-- admin = owner 全部权限 + 系统级权限
INSERT INTO public.role_permissions (role_slug, permission_code) VALUES
    ('admin','system:config'),             ('admin','system:user:read'),           ('admin','system:user:manage'),
    ('admin','system:workspace:list'),     ('admin','system:workspace:manage'),    ('admin','system:audit:read'),
    ('admin','workspace:read'),            ('admin','workspace:update'),           ('admin','workspace:delete'),
    ('admin','workspace:transfer'),        ('admin','project:read'),               ('admin','project:create'),
    ('admin','project:update'),            ('admin','project:delete'),             ('admin','issue:read'),
    ('admin','issue:create'),              ('admin','issue:edit_own'),             ('admin','issue:edit_all'),
    ('admin','issue:delete'),              ('admin','issue:transition'),           ('admin','issue:reassign'),
    ('admin','issue:change_priority'),     ('admin','issue:manage_sprint'),        ('admin','member:invite'),
    ('admin','member:remove'),             ('admin','member:change_role'),         ('admin','sprint:read'),
    ('admin','sprint:create'),             ('admin','sprint:update'),              ('admin','sprint:delete'),
    ('admin','sprint:lifecycle'),          ('admin','sprint:plan'),                ('admin','version:read'),
    ('admin','version:create'),            ('admin','version:update'),             ('admin','version:release'),
    ('admin','version:delete'),            ('admin','defect:create'),              ('admin','qa:report'),
    ('admin','analytics:read'),            ('admin','analytics:export'),           ('admin','automation:manage'),
    ('admin','deploy:report'),             ('admin','audit:read'),                 ('admin','webhook:manage'),
    ('admin','intake:manage'),             ('admin','pages:manage'),               ('admin','comment:moderate'),
    ('admin','relation:manage'),           ('admin','field:edit_severity'),        ('admin','field:edit_effort'),
    ('admin','field:edit_deadline'),       ('admin','menu:settings'),              ('admin','menu:audit')
ON CONFLICT (role_slug, permission_code) DO NOTHING;

-- pm 项目经理
INSERT INTO public.role_permissions (role_slug, permission_code) VALUES
    ('pm','workspace:read'),
    ('pm','project:read'),                 ('pm','project:create'),                ('pm','project:update'),
    ('pm','project:delete'),               ('pm','issue:read'),                    ('pm','issue:create'),
    ('pm','issue:edit_own'),               ('pm','issue:edit_all'),                ('pm','issue:delete'),
    ('pm','issue:transition'),             ('pm','issue:reassign'),                ('pm','issue:change_priority'),
    ('pm','issue:manage_sprint'),          ('pm','sprint:read'),                   ('pm','sprint:create'),
    ('pm','sprint:update'),                ('pm','sprint:delete'),                 ('pm','sprint:lifecycle'),
    ('pm','sprint:plan'),                  ('pm','version:read'),                  ('pm','version:create'),
    ('pm','version:update'),               ('pm','version:release'),               ('pm','version:delete'),
    ('pm','defect:create'),                ('pm','analytics:read'),                ('pm','analytics:export'),
    ('pm','automation:manage'),            ('pm','intake:manage'),                 ('pm','pages:manage'),
    ('pm','relation:manage'),              ('pm','field:edit_effort'),             ('pm','field:edit_deadline')
ON CONFLICT (role_slug, permission_code) DO NOTHING;

-- po 产品经理
INSERT INTO public.role_permissions (role_slug, permission_code) VALUES
    ('po','workspace:read'),
    ('po','project:read'),                 ('po','issue:read'),                    ('po','issue:create'),
    ('po','issue:edit_own'),               ('po','issue:edit_all'),
    ('po','issue:transition'),             ('po','issue:reassign'),                ('po','issue:change_priority'),
    ('po','version:read'),                 ('po','version:create'),                ('po','version:update'),
    ('po','version:release'),
    ('po','analytics:read'),               ('po','intake:manage'),                 ('po','pages:manage'),
    ('po','relation:manage'),              ('po','field:edit_deadline')
ON CONFLICT (role_slug, permission_code) DO NOTHING;

-- techlead 技术经理
INSERT INTO public.role_permissions (role_slug, permission_code) VALUES
    ('techlead','workspace:read'),
    ('techlead','project:read'),           ('techlead','issue:read'),              ('techlead','issue:create'),
    ('techlead','issue:edit_own'),         ('techlead','issue:edit_all'),          ('techlead','issue:delete'),
    ('techlead','issue:transition'),       ('techlead','issue:reassign'),
    ('techlead','issue:manage_sprint'),    ('techlead','sprint:read'),             ('techlead','sprint:create'),
    ('techlead','sprint:update'),          ('techlead','sprint:lifecycle'),        ('techlead','sprint:plan'),
    ('techlead','version:read'),           ('techlead','version:create'),          ('techlead','version:update'),
    ('techlead','version:release'),        ('techlead','defect:create'),           ('techlead','qa:report'),
    ('techlead','analytics:read'),         ('techlead','automation:manage'),       ('techlead','deploy:report'),
    ('techlead','pages:manage'),           ('techlead','relation:manage'),
    ('techlead','field:edit_severity'),    ('techlead','field:edit_effort')
ON CONFLICT (role_slug, permission_code) DO NOTHING;

-- qalead 测试经理
INSERT INTO public.role_permissions (role_slug, permission_code) VALUES
    ('qalead','workspace:read'),
    ('qalead','project:read'),             ('qalead','issue:read'),                ('qalead','issue:create'),
    ('qalead','issue:edit_own'),           ('qalead','issue:edit_all'),
    ('qalead','issue:transition'),         ('qalead','issue:reassign'),
    ('qalead','qa:report'),                ('qalead','analytics:read'),
    ('qalead','relation:manage'),          ('qalead','field:edit_severity')
ON CONFLICT (role_slug, permission_code) DO NOTHING;

-- dev 开发
INSERT INTO public.role_permissions (role_slug, permission_code) VALUES
    ('dev','workspace:read'),
    ('dev','project:read'),                ('dev','issue:read'),                   ('dev','issue:create'),
    ('dev','issue:edit_own'),              ('dev','issue:transition'),
    ('dev','sprint:read'),                 ('dev','version:read'),
    ('dev','defect:create'),               ('dev','relation:manage'),
    ('dev','field:edit_severity'),         ('dev','field:edit_effort')
ON CONFLICT (role_slug, permission_code) DO NOTHING;

-- guest 访客
INSERT INTO public.role_permissions (role_slug, permission_code) VALUES
    ('guest','workspace:read'),
    ('guest','project:read'),              ('guest','issue:read'),
    ('guest','sprint:read'),               ('guest','version:read')
ON CONFLICT (role_slug, permission_code) DO NOTHING;


-- ============================================================================
-- [原 0024] 工作项分表拆分 —— task / requirement / defect
-- ============================================================================

-- 1. task 表
CREATE TABLE IF NOT EXISTS task (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id       UUID NOT NULL DEFAULT gen_random_uuid(),
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    sequence_id     BIGINT NOT NULL,
    parent_id       BIGINT REFERENCES task(id),
    depth           SMALLINT NOT NULL DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    name            TEXT NOT NULL,
    description_json    JSONB,
    description_html    TEXT,
    description_stripped TEXT,
    state_id        BIGINT NOT NULL REFERENCES states(id),
    priority        TEXT NOT NULL DEFAULT 'none'
                    CHECK (priority IN ('urgent','high','medium','low','none')),
    category        TEXT CHECK (category IN ('frontend','backend','qa','doc','design','devops','other')),
    actual_effort   NUMERIC(8,2),
    remaining_effort NUMERIC(8,2),
    delay_reason    TEXT CHECK (delay_reason IN ('requirement_change','resource','blocked','other')),
    point           SMALLINT CHECK (point BETWEEN 0 AND 12),
    estimate_point_id BIGINT REFERENCES estimate_points(id),
    sprint_id       BIGINT REFERENCES sprints(id),
    version_id      BIGINT REFERENCES versions(id),
    progress        SMALLINT NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    start_date      DATE,
    target_date     DATE,
    completed_at    TIMESTAMPTZ,
    is_draft        BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      DOUBLE PRECISION NOT NULL DEFAULT 65535,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    version         INT NOT NULL DEFAULT 1,
    UNIQUE (project_id, sequence_id),
);

-- 2. requirement 表
CREATE TABLE IF NOT EXISTS requirement (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id       UUID NOT NULL DEFAULT gen_random_uuid(),
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    sequence_id     BIGINT NOT NULL,
    parent_id       BIGINT REFERENCES requirement(id),
    depth           SMALLINT NOT NULL DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    name            TEXT NOT NULL,
    description_json    JSONB,
    description_html    TEXT,
    description_stripped TEXT,
    state_id        BIGINT NOT NULL REFERENCES states(id),
    priority        TEXT NOT NULL DEFAULT 'none'
                    CHECK (priority IN ('urgent','high','medium','low','none')),
    source          TEXT CHECK (source IN ('customer','internal','competitor','other')),
    acceptance_criteria JSONB,
    business_value  TEXT,
    review_status   TEXT CHECK (review_status IN ('draft','reviewing','accepted','rejected','verified')),
    point           SMALLINT CHECK (point BETWEEN 0 AND 12),
    estimate_point_id BIGINT REFERENCES estimate_points(id),
    sprint_id       BIGINT REFERENCES sprints(id),
    version_id      BIGINT REFERENCES versions(id),
    progress        SMALLINT NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    start_date      DATE,
    target_date     DATE,
    completed_at    TIMESTAMPTZ,
    is_draft        BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      DOUBLE PRECISION NOT NULL DEFAULT 65535,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    version         INT NOT NULL DEFAULT 1,
    UNIQUE (project_id, sequence_id),
);

-- 3. defect 表
CREATE TABLE IF NOT EXISTS defect (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id       UUID NOT NULL DEFAULT gen_random_uuid(),
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    sequence_id     BIGINT NOT NULL,
    parent_id       BIGINT REFERENCES defect(id),
    depth           SMALLINT NOT NULL DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    name            TEXT NOT NULL,
    description_json    JSONB,
    description_html    TEXT,
    description_stripped TEXT,
    state_id        BIGINT NOT NULL REFERENCES states(id),
    priority        TEXT NOT NULL DEFAULT 'none'
                    CHECK (priority IN ('urgent','high','medium','low','none')),
    severity        SMALLINT NOT NULL CHECK (severity BETWEEN 1 AND 5),
    found_phase     TEXT NOT NULL CHECK (found_phase IN ('unit','integration','uat','production','customer')),
    found_version_id BIGINT REFERENCES versions(id),
    fix_version_id   BIGINT REFERENCES versions(id),
    root_cause_category TEXT CHECK (root_cause_category IN ('requirement','technical','environment','data')),
    verifier_id     BIGINT REFERENCES users(id),
    environment     JSONB,
    reproduce_steps JSONB NOT NULL,
    fix_steps       JSONB,
    regression_risk TEXT CHECK (regression_risk IN ('low','medium','high')),
    point           SMALLINT CHECK (point BETWEEN 0 AND 12),
    estimate_point_id BIGINT REFERENCES estimate_points(id),
    sprint_id       BIGINT REFERENCES sprints(id),
    version_id      BIGINT REFERENCES versions(id),
    progress        SMALLINT NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    start_date      DATE,
    target_date     DATE,
    completed_at    TIMESTAMPTZ,
    is_draft        BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      DOUBLE PRECISION NOT NULL DEFAULT 65535,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    version         INT NOT NULL DEFAULT 1,
    UNIQUE (project_id, sequence_id),
);

-- 4. task_ext 表
CREATE TABLE IF NOT EXISTS task_ext (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    task_id         BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    field_name      TEXT NOT NULL,
    field_value     JSONB NOT NULL,
    field_schema    JSONB NOT NULL,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, field_name)
);

-- 5. requirement_ext 表
CREATE TABLE IF NOT EXISTS requirement_ext (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    requirement_id  BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    field_name      TEXT NOT NULL,
    field_value     JSONB NOT NULL,
    field_schema    JSONB NOT NULL,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (requirement_id, field_name)
);

-- 6. defect_ext 表
CREATE TABLE IF NOT EXISTS defect_ext (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    defect_id       BIGINT NOT NULL REFERENCES defect(id) ON DELETE CASCADE,
    field_name      TEXT NOT NULL,
    field_value     JSONB NOT NULL,
    field_schema    JSONB NOT NULL,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (defect_id, field_name)
);

-- 7. biz_entity_relation 表
CREATE TABLE IF NOT EXISTS biz_entity_relation (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    source_type     TEXT NOT NULL CHECK (source_type IN ('task','requirement','defect')),
    source_id       BIGINT NOT NULL,
    target_type     TEXT NOT NULL CHECK (target_type IN ('task','requirement','defect')),
    target_id       BIGINT NOT NULL,
    relation_type   TEXT NOT NULL CHECK (relation_type IN ('implemented_by','relates_to','duplicate','blocked_by','parent_child','found_in','fixed_in','verified_in')),
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_type, source_id, target_type, target_id, relation_type)
);

-- 8. states 表扩展
ALTER TABLE states ADD COLUMN IF NOT EXISTS applicable_types TEXT[] NOT NULL DEFAULT '{"all"}';

-- 9. state_transitions 表扩展
ALTER TABLE state_transitions ALTER COLUMN type_code DROP DEFAULT;
UPDATE state_transitions SET type_code = 'all' WHERE type_code = '';


-- ============================================================================
-- [原 0025] 工作项归档支持（原 0025_work_item_archive）
-- ============================================================================

-- 给 task/requirement/defect 表添加归档时间戳和索引
ALTER TABLE task ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ NULL;
ALTER TABLE requirement ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ NULL;
ALTER TABLE defect ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_task_archived       ON task(archived_at)       WHERE archived_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_requirement_archived ON requirement(archived_at) WHERE archived_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_defect_archived      ON defect(archived_at)      WHERE archived_at IS NOT NULL;


-- ============================================================================
-- [原 0026] 工作项数据迁移（issues → task/requirement/defect）
-- 一次性数据同步，旧库数据拆分到新分表；全新空库执行无影响（ON CONFLICT DO NOTHING）。
-- ============================================================================

-- 1. task 类型工作项
INSERT INTO task (
    public_id, workspace_id, project_id, sequence_id, parent_id, depth,
    name, description_json, description_html, state_id, priority,
    category, actual_effort, remaining_effort, delay_reason,
    point, estimate_point_id, sprint_id, version_id, progress,
    start_date, target_date, completed_at, is_draft, sort_order,
    created_by, created_at, updated_at, deleted_at
)
SELECT
    public_id, workspace_id, project_id, sequence_id, parent_id, depth,
    name, description_json, description_html, state_id, priority,
    category, actual_effort, remaining_effort, delay_reason,
    point, estimate_point_id, sprint_id, version_id, progress,
    start_date, target_date, completed_at, is_draft, sort_order,
    created_by, created_at, updated_at, deleted_at
FROM issues
WHERE type_code = 'task'
ON CONFLICT (project_id, sequence_id) DO NOTHING;

-- 2. requirement 类型工作项
INSERT INTO requirement (
    public_id, workspace_id, project_id, sequence_id, parent_id, depth,
    name, description_json, description_html, state_id, priority,
    source, point, estimate_point_id, sprint_id, version_id, progress,
    start_date, target_date, completed_at, is_draft, sort_order,
    created_by, created_at, updated_at, deleted_at
)
SELECT
    public_id, workspace_id, project_id, sequence_id, parent_id, depth,
    name, description_json, description_html, state_id, priority,
    source, point, estimate_point_id, sprint_id, version_id, progress,
    start_date, target_date, completed_at, is_draft, sort_order,
    created_by, created_at, updated_at, deleted_at
FROM issues
WHERE type_code = 'requirement'
ON CONFLICT (project_id, sequence_id) DO NOTHING;

-- 3. defect 类型工作项
INSERT INTO defect (
    public_id, workspace_id, project_id, sequence_id, parent_id, depth,
    name, description_json, description_html, state_id, priority,
    severity, found_phase, found_version_id, fix_version_id, root_cause_category,
    verifier_id, environment, reproduce_steps,
    point, estimate_point_id, sprint_id, version_id, progress,
    start_date, target_date, completed_at, is_draft, sort_order,
    created_by, created_at, updated_at, deleted_at
)
SELECT
    public_id, workspace_id, project_id, sequence_id, parent_id, depth,
    name, description_json, description_html, state_id, priority,
    severity, found_phase, found_version_id, fix_version_id, root_cause_category,
    verifier_id, environment, reproduce_steps,
    point, estimate_point_id, sprint_id, version_id, progress,
    start_date, target_date, completed_at, is_draft, sort_order,
    created_by, created_at, updated_at, deleted_at
FROM issues
WHERE type_code = 'defect'
ON CONFLICT (project_id, sequence_id) DO NOTHING;

-- 4. 迁移关联关系数据
INSERT INTO biz_entity_relation (
    workspace_id, project_id, source_type, source_id, target_type, target_id,
    relation_type, created_by, created_at
)
SELECT
    ir.workspace_id, ir.project_id,
    CASE WHEN i1.type_code = 'task' THEN 'task' WHEN i1.type_code = 'requirement' THEN 'requirement' ELSE 'defect' END,
    ir.source_id,
    CASE WHEN i2.type_code = 'task' THEN 'task' WHEN i2.type_code = 'requirement' THEN 'requirement' ELSE 'defect' END,
    ir.target_id,
    ir.relation_type, ir.created_by, ir.created_at
FROM issue_relations ir
JOIN issues i1 ON ir.source_id = i1.id
JOIN issues i2 ON ir.target_id = i2.id
ON CONFLICT (source_type, source_id, target_type, target_id, relation_type) DO NOTHING;


-- ============================================================================
-- [原 0026] Issue 版本快照审计（对标 Plane IssueVersion / Activity History）
-- ============================================================================

CREATE TABLE IF NOT EXISTS issue_versions (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    issue_id        BIGINT NOT NULL,
    version         INT NOT NULL,
    snapshot        JSONB NOT NULL,
    changed_fields  TEXT[] DEFAULT '{}',
    change_type     TEXT NOT NULL DEFAULT 'update'
                    CHECK (change_type IN ('create','update','delete','transition')),
    created_by      BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (issue_id, version)
);

ALTER TABLE issue_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE issue_versions FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON issue_versions CREATE POLICY tenant_isolation ON issue_versions
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

CREATE INDEX IF NOT EXISTS idx_issue_versions_issue   ON issue_versions(workspace_id, issue_id, version DESC);
CREATE INDEX IF NOT EXISTS idx_issue_versions_project ON issue_versions(workspace_id, project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_issue_versions_actor   ON issue_versions(workspace_id, created_by) WHERE created_by IS NOT NULL;

COMMENT ON TABLE issue_versions IS '工作项版本快照：记录每次变更前的字段状态，支撑审计回溯与变更对比';
COMMENT ON COLUMN issue_versions.snapshot IS '变更前完整字段快照（JSONB，对应 BaseWorkitem 结构）';
COMMENT ON COLUMN issue_versions.changed_fields IS '本次变更涉及的字段名，便于 diff 视图渲染';
COMMENT ON COLUMN issue_versions.change_type IS '变更类型：create(创建) / update(字段更新) / delete(软删除) / transition(状态流转)';
COMMENT ON COLUMN issue_versions.version IS '递增版本号；与 issues.version 一一对应';


-- ============================================================================
-- [原 0025] Epic 类型 + Module 模块体系
-- ============================================================================

-- 1. issues.type_code 约束已含 epic（init 文件 4518 行已对齐），此处确保约束
ALTER TABLE issues DROP CONSTRAINT IF EXISTS issues_type_code_check;
ALTER TABLE issues ADD CONSTRAINT issues_type_code_check
    CHECK (type_code = ANY (ARRAY['epic'::text, 'requirement'::text, 'task'::text, 'defect'::text]));

COMMENT ON COLUMN issues.type_code IS '工作项类型: epic(史诗) / requirement(需求) / task(任务) / defect(缺陷)';

-- 2. modules 表升级：补齐 public_id 列与缺失的约束/索引
ALTER TABLE modules ADD COLUMN IF NOT EXISTS public_id UUID NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE modules DROP CONSTRAINT IF EXISTS modules_unique_project_name;
CREATE UNIQUE INDEX IF NOT EXISTS modules_unique_project_name ON modules(project_id, name) WHERE deleted_at IS NULL;
ALTER TABLE modules DROP CONSTRAINT IF EXISTS modules_unique_public_id;
CREATE UNIQUE INDEX IF NOT EXISTS modules_unique_public_id ON modules(public_id) WHERE deleted_at IS NULL;
-- 替换旧的 status check: cancelled → archived
ALTER TABLE modules DROP CONSTRAINT IF EXISTS modules_status_check;
ALTER TABLE modules ADD CONSTRAINT modules_status_check CHECK (status = ANY (ARRAY['active'::text, 'completed'::text, 'archived'::text]));

-- 3. module_issues 关联表（工作项 × 模块 M:N）
CREATE TABLE IF NOT EXISTS module_issues (
    module_id       BIGINT NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    issue_id        BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (module_id, issue_id)
);

COMMENT ON TABLE module_issues IS '工作项-模块 M:N 关联表（一个工作项可属于多个模块，一个模块包含多个工作项）';


-- ============================================================================
-- 分表 / RLS / 索引（为新增表补全租户隔离与索引）
-- ============================================================================

-- task RLS
ALTER TABLE task ENABLE ROW LEVEL SECURITY;
ALTER TABLE task FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON task CREATE POLICY tenant_isolation ON task
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- requirement RLS
ALTER TABLE requirement ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON requirement CREATE POLICY tenant_isolation ON requirement
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- defect RLS
ALTER TABLE defect ENABLE ROW LEVEL SECURITY;
ALTER TABLE defect FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON defect CREATE POLICY tenant_isolation ON defect
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- task_ext / requirement_ext / defect_ext RLS
ALTER TABLE task_ext ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_ext FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON task_ext CREATE POLICY tenant_isolation ON task_ext
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

ALTER TABLE requirement_ext ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement_ext FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON requirement_ext CREATE POLICY tenant_isolation ON requirement_ext
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

ALTER TABLE defect_ext ENABLE ROW LEVEL SECURITY;
ALTER TABLE defect_ext FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON defect_ext CREATE POLICY tenant_isolation ON defect_ext
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- biz_entity_relation RLS
ALTER TABLE biz_entity_relation ENABLE ROW LEVEL SECURITY;
ALTER TABLE biz_entity_relation FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON biz_entity_relation CREATE POLICY tenant_isolation ON biz_entity_relation
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- modules RLS（升级后补 RLS + tenant_isolation 策略）
ALTER TABLE modules ENABLE ROW LEVEL SECURITY;
ALTER TABLE modules FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON modules CREATE POLICY tenant_isolation ON modules
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- module_issues RLS（通过 modules 继承 workspace 隔离，无需独立 RLS；模块由 modules 管）


-- ============================================================================
-- 索引（新增表 + 升级表）
-- ============================================================================

-- task 索引
CREATE INDEX IF NOT EXISTS idx_task_project_state   ON task(workspace_id, project_id, state_id)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_task_project_sprint  ON task(workspace_id, project_id, sprint_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_task_parent          ON task(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_task_target_date     ON task(workspace_id, project_id, target_date) WHERE deleted_at IS NULL AND completed_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_task_updated         ON task(workspace_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_fts             ON task USING gin(to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(description_stripped,'')));
CREATE INDEX IF NOT EXISTS idx_task_sort            ON task(project_id, state_id, sort_order);

-- requirement 索引
CREATE INDEX IF NOT EXISTS idx_requirement_project_state   ON requirement(workspace_id, project_id, state_id)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_requirement_project_sprint  ON requirement(workspace_id, project_id, sprint_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_requirement_parent          ON requirement(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_requirement_target_date     ON requirement(workspace_id, project_id, target_date) WHERE deleted_at IS NULL AND completed_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_requirement_updated         ON requirement(workspace_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_requirement_fts             ON requirement USING gin(to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(description_stripped,'')));
CREATE INDEX IF NOT EXISTS idx_requirement_sort            ON requirement(project_id, state_id, sort_order);

-- defect 索引
CREATE INDEX IF NOT EXISTS idx_defect_project_state   ON defect(workspace_id, project_id, state_id)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_defect_project_sprint  ON defect(workspace_id, project_id, sprint_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_defect_parent          ON defect(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_defect_target_date     ON defect(workspace_id, project_id, target_date) WHERE deleted_at IS NULL AND completed_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_defect_updated         ON defect(workspace_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_defect_fts             ON defect USING gin(to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(description_stripped,'')));
CREATE INDEX IF NOT EXISTS idx_defect_sort            ON defect(project_id, state_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_defect_severity        ON defect(workspace_id, project_id, severity) WHERE deleted_at IS NULL AND completed_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_defect_root_cause      ON defect(workspace_id, project_id, root_cause_category) WHERE deleted_at IS NULL;

-- biz_entity_relation 索引
CREATE INDEX IF NOT EXISTS idx_biz_entity_relation_source ON biz_entity_relation(workspace_id, source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_biz_entity_relation_target ON biz_entity_relation(workspace_id, target_type, target_id);

-- states 索引
CREATE INDEX IF NOT EXISTS idx_states_applicable_types ON states USING gin(applicable_types);

-- state_transitions 索引
CREATE INDEX IF NOT EXISTS idx_state_transitions_type ON state_transitions(project_id, type_code);

-- module_issues 索引
CREATE INDEX IF NOT EXISTS idx_module_issues_issue ON module_issues(issue_id);

-- modules 升级索引
CREATE INDEX IF NOT EXISTS idx_modules_project_sort ON modules(project_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_modules_lead         ON modules(lead_id) WHERE deleted_at IS NULL;


-- ============================================================================
-- 触发器（modules updated_at 已在 init 中创建，此处用 CREATE OR REPLACE 保安全）
-- ============================================================================

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 确保分表有 updated_at 触发器（若已存在则跳过）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_task_updated_at') THEN
        CREATE TRIGGER trg_task_updated_at BEFORE UPDATE ON task FOR EACH ROW EXECUTE FUNCTION set_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_requirement_updated_at') THEN
        CREATE TRIGGER trg_requirement_updated_at BEFORE UPDATE ON requirement FOR EACH ROW EXECUTE FUNCTION set_updated_at();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_defect_updated_at') THEN
        CREATE TRIGGER trg_defect_updated_at BEFORE UPDATE ON defect FOR EACH ROW EXECUTE FUNCTION set_updated_at();
    END IF;
    -- modules 触发器已在 init 中定义（trg_modules_updated_at），无需重复创建
END $$;


-- ============================================================================
-- 迁移整合完毕。
-- 至此 sql/ 目录下仅保留 ydsz-plane-init.sql 一个文件，
-- migrate 程序将自动进入全量 dump 模式（运行 migrate up → 执行本文件）。
-- 原增量迁移文件：
--   0023_rbac_roles_and_permissions、0024_work_item_split、
--   0025_work_item_archive、0025_epic_and_modules、
--   0026_work_item_migration、0026_issue_version_snapshots
-- 已全部合并到本文件中。
-- ============================================================================

-- ============================================================================
-- 迁移脚本整合：0027_work_item_views.up.sql
-- 创建三个独立的工作项视图（task_view, requirement_view, defect_view）
-- ============================================================================

-- 创建三个独立的工作项视图，分别对应需求、任务、缺陷表的全字段，无历史兼容逻辑
-- 视图只读，写入需要操作对应主表

-- 1. 任务工作项视图
CREATE OR REPLACE VIEW task_view AS
SELECT
    id,
    public_id,
    workspace_id,
    project_id,
    sequence_id,
    'task'::text as type_code, -- 固定类型标识
    parent_id,
    depth,
    name,
    description_json,
    description_html,
    description_stripped,
    state_id,
    priority,
    -- task专属字段
    category,
    actual_effort,
    remaining_effort,
    delay_reason,
    -- 公共字段
    point,
    estimate_point_id,
    sprint_id,
    version_id,
    progress,
    start_date,
    target_date,
    completed_at,
    is_draft,
    sort_order,
    assignee_ids,
    label_ids,
    module_ids,
    watcher_ids,
    created_by,
    created_at,
    updated_at,
    deleted_at,
    version,
    archived_at
FROM task;

-- 2. 需求工作项视图
CREATE OR REPLACE VIEW requirement_view AS
SELECT
    id,
    public_id,
    workspace_id,
    project_id,
    sequence_id,
    'requirement'::text as type_code, -- 固定类型标识
    parent_id,
    depth,
    name,
    description_json,
    description_html,
    description_stripped,
    state_id,
    priority,
    -- requirement专属字段
    source,
    acceptance_criteria,
    business_value,
    review_status,
    -- 公共字段
    point,
    estimate_point_id,
    sprint_id,
    version_id,
    progress,
    start_date,
    target_date,
    completed_at,
    is_draft,
    sort_order,
    assignee_ids,
    label_ids,
    module_ids,
    watcher_ids,
    created_by,
    created_at,
    updated_at,
    deleted_at,
    version,
    archived_at
FROM requirement;

-- 3. 缺陷工作项视图
CREATE OR REPLACE VIEW defect_view AS
SELECT
    id,
    public_id,
    workspace_id,
    project_id,
    sequence_id,
    'defect'::text as type_code, -- 固定类型标识
    parent_id,
    depth,
    name,
    description_json,
    description_html,
    description_stripped,
    state_id,
    priority,
    -- defect专属字段
    severity,
    found_phase,
    found_version_id,
    fix_version_id,
    root_cause_category,
    verifier_id,
    environment,
    reproduce_steps,
    fix_steps,
    regression_risk,
    -- 公共字段
    point,
    estimate_point_id,
    sprint_id,
    version_id,
    progress,
    start_date,
    target_date,
    completed_at,
    is_draft,
    sort_order,
    assignee_ids,
    label_ids,
    module_ids,
    watcher_ids,
    created_by,
    created_at,
    updated_at,
    deleted_at,
    version,
    archived_at
FROM defect;

-- 创建视图RLS权限，和安全策略对齐
ALTER VIEW task_view SET (security_barrier = true);
ALTER VIEW requirement_view SET (security_barrier = true);
ALTER VIEW defect_view SET (security_barrier = true);


-- ============================================================================
-- 全部迁移脚本整合完毕。
-- sql/ 目录下原迁移文件已合并到本文件中，可删除原文件。
-- ============================================================================


-- ============================================================================
-- 迁移脚本整合：0027_work_item_views.down.sql（回滚脚本，仅供参考）
-- 以下内容已注释，不执行，仅作回滚参考
-- ============================================================================
-- 删除三个独立工作项视图的脚本
-- DROP VIEW IF EXISTS defect_view;
-- DROP VIEW IF EXISTS requirement_view;
-- DROP VIEW IF EXISTS task_view;


-- ============================================================================
-- 全部迁移脚本整合完毕。
-- sql/ 目录下原迁移文件 (0027_work_item_views.up.sql / .down.sql) 
-- 已合并到本文件中，原文件可安全删除。
-- ============================================================================


-- ============================================================================
-- 迁移脚本整合：0028_drop_legacy_issues.up.sql
-- 彻底删除所有旧 issues 相关表、索引、约束、触发器，切换到新的三表结构
-- ============================================================================

-- 彻底删除所有旧issues相关表、索引、约束、触发器，完全切换到新的三表结构
-- 执行前请确保所有数据已迁移到新表，且所有代码已适配新结构
-- 该操作不可逆，请勿在生产环境未测试的情况下直接执行

-- 删除依赖表（按依赖顺序倒序删除）
DROP TABLE IF EXISTS issue_reactions;
DROP TABLE IF EXISTS issue_votes;
DROP TABLE IF EXISTS issue_subscriptions;
DROP TABLE IF EXISTS issue_comments;
DROP TABLE IF EXISTS issue_activities;
DROP TABLE IF EXISTS issue_dependencies;
DROP TABLE IF EXISTS issue_relations;
DROP TABLE IF EXISTS issue_watchers;
DROP TABLE IF EXISTS issue_modules;
DROP TABLE IF EXISTS issue_labels;
DROP TABLE IF EXISTS issue_assignees;
DROP TABLE IF EXISTS issue_sequences;
DROP TABLE IF EXISTS sprint_issues;
DROP TABLE IF EXISTS intake_issues;
DROP TABLE IF EXISTS project_sequences;

-- 删除主表
DROP TABLE IF EXISTS issues;

-- 删除所有旧issues相关的触发器
DROP TRIGGER IF EXISTS issues_set_updated_at ON issues;
DROP TRIGGER IF EXISTS issues_set_sequence ON issues;
DROP TRIGGER IF EXISTS issues_elasticsearch_sync ON issues;
DROP TRIGGER IF EXISTS issues_outbox ON issues;

-- 删除所有旧issues相关的函数
DROP FUNCTION IF EXISTS generate_issue_identifier;


-- ============================================================================
-- 迁移脚本整合：0027_processed_events_idempotency.up.sql
-- 消费者幂等去重表，支持 at-least-once 投递下的重复处理跳过
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


-- ============================================================================
-- 迁移脚本整合：0028_dlq_monitoring_metadata.up.sql
-- DLQ 死信事件元数据表，记录 Relay 发布失败 + 消费者 NACK 路由到 DLX 的消息
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


-- ============================================================================
-- 迁移脚本整合：0029_split_relation_tables.up.sql
-- 将旧 issues 通用关联表拆分为 task/requirement/defect 三套独立关联表
-- ============================================================================

-- 任务工作项关联表
CREATE TABLE IF NOT EXISTS task_assignees (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, user_id)
);
ALTER TABLE task_assignees ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_assignees FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_assignees
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_assignee_task ON task_assignees(workspace_id, task_id) ;
CREATE INDEX idx_task_assignee_user ON task_assignees(workspace_id, user_id) ;

CREATE TABLE IF NOT EXISTS task_labels (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    label_id BIGINT NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, label_id)
);
ALTER TABLE task_labels ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_labels FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_labels
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_label_task ON task_labels(workspace_id, task_id) ;
CREATE INDEX idx_task_label_label ON task_labels(workspace_id, label_id) ;

CREATE TABLE IF NOT EXISTS task_modules (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    module_id BIGINT NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, module_id)
);
ALTER TABLE task_modules ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_modules FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_modules
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_module_task ON task_modules(workspace_id, task_id) ;
CREATE INDEX idx_task_module_module ON task_modules(workspace_id, module_id) ;

CREATE TABLE IF NOT EXISTS task_watchers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, user_id)
);
ALTER TABLE task_watchers ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_watchers FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_watchers
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_watcher_task ON task_watchers(workspace_id, task_id) ;
CREATE INDEX idx_task_watcher_user ON task_watchers(workspace_id, user_id) ;

CREATE TABLE IF NOT EXISTS task_relations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    source_task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    target_task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL CHECK (relation_type IN ('duplicate','relates_to','blocked_by','start_before','finish_before')),
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_task_id, target_task_id, relation_type)
);
ALTER TABLE task_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_relations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_relations
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_relation_source ON task_relations(workspace_id, source_task_id) ;
CREATE INDEX idx_task_relation_target ON task_relations(workspace_id, target_task_id) ;

CREATE TABLE IF NOT EXISTS task_comments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    content_json JSONB NOT NULL,
    content_html TEXT NOT NULL,
    parent_id BIGINT REFERENCES task_comments(id) ON DELETE CASCADE,
    created_by BIGINT NOT NULL REFERENCES users(id),
    updated_by BIGINT REFERENCES users(id),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE task_comments ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_comments FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_comments
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_comment_task ON task_comments(workspace_id, task_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS task_activities (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    verb TEXT NOT NULL CHECK (verb IN ('created','updated','transitioned','attached','linked','commented')),
    field_name TEXT,
    old_value TEXT,
    new_value TEXT,
    actor_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
ALTER TABLE task_activities ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_activities FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_activities
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_activity_task ON task_activities(workspace_id, task_id, created_at DESC);
CREATE TABLE task_activities_default PARTITION OF task_activities DEFAULT;


-- 需求工作项关联表
CREATE TABLE IF NOT EXISTS requirement_labels (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    requirement_id BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    label_id BIGINT NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (requirement_id, label_id)
);
ALTER TABLE requirement_labels ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement_labels FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON requirement_labels
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_requirement_label_req ON requirement_labels(workspace_id, requirement_id) ;
CREATE INDEX idx_requirement_label_label ON requirement_labels(workspace_id, label_id) ;

CREATE TABLE IF NOT EXISTS requirement_watchers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    requirement_id BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (requirement_id, user_id)
);
ALTER TABLE requirement_watchers ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement_watchers FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON requirement_watchers
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_requirement_watcher_req ON requirement_watchers(workspace_id, requirement_id) ;
CREATE INDEX idx_requirement_watcher_user ON requirement_watchers(workspace_id, user_id) ;

CREATE TABLE IF NOT EXISTS requirement_relations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    source_requirement_id BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    target_requirement_id BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL CHECK (relation_type IN ('duplicate','relates_to','implemented_by')),
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_requirement_id, target_requirement_id, relation_type)
);
ALTER TABLE requirement_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement_relations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON requirement_relations
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_requirement_relation_source ON requirement_relations(workspace_id, source_requirement_id) ;
CREATE INDEX idx_requirement_relation_target ON requirement_relations(workspace_id, target_requirement_id) ;

CREATE TABLE IF NOT EXISTS requirement_comments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    requirement_id BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    content_json JSONB NOT NULL,
    content_html TEXT NOT NULL,
    parent_id BIGINT REFERENCES requirement_comments(id) ON DELETE CASCADE,
    created_by BIGINT NOT NULL REFERENCES users(id),
    updated_by BIGINT REFERENCES users(id),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE requirement_comments ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement_comments FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON requirement_comments
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_requirement_comment_req ON requirement_comments(workspace_id, requirement_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS requirement_activities (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    requirement_id BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    verb TEXT NOT NULL CHECK (verb IN ('created','updated','transitioned','attached','linked','commented')),
    field_name TEXT,
    old_value TEXT,
    new_value TEXT,
    actor_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
ALTER TABLE requirement_activities ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement_activities FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON requirement_activities
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_requirement_activity_req ON requirement_activities(workspace_id, requirement_id, created_at DESC);
CREATE TABLE requirement_activities_default PARTITION OF requirement_activities DEFAULT;


-- 缺陷工作项关联表
CREATE TABLE IF NOT EXISTS defect_watchers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    defect_id BIGINT NOT NULL REFERENCES defect(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (defect_id, user_id)
);
ALTER TABLE defect_watchers ENABLE ROW LEVEL SECURITY;
ALTER TABLE defect_watchers FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON defect_watchers
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_defect_watcher_defect ON defect_watchers(workspace_id, defect_id) ;
CREATE INDEX idx_defect_watcher_user ON defect_watchers(workspace_id, user_id) ;

CREATE TABLE IF NOT EXISTS defect_relations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    source_defect_id BIGINT NOT NULL REFERENCES defect(id) ON DELETE CASCADE,
    target_defect_id BIGINT NOT NULL REFERENCES defect(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL CHECK (relation_type IN ('duplicate','relates_to','blocked_by','found_in','fixed_in','verified_in')),
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_defect_id, target_defect_id, relation_type)
);
ALTER TABLE defect_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE defect_relations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON defect_relations
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_defect_relation_source ON defect_relations(workspace_id, source_defect_id) ;
CREATE INDEX idx_defect_relation_target ON defect_relations(workspace_id, target_defect_id) ;

CREATE TABLE IF NOT EXISTS defect_comments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    defect_id BIGINT NOT NULL REFERENCES defect(id) ON DELETE CASCADE,
    content_json JSONB NOT NULL,
    content_html TEXT NOT NULL,
    parent_id BIGINT REFERENCES defect_comments(id) ON DELETE CASCADE,
    created_by BIGINT NOT NULL REFERENCES users(id),
    updated_by BIGINT REFERENCES users(id),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE defect_comments ENABLE ROW LEVEL SECURITY;
ALTER TABLE defect_comments FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON defect_comments
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_defect_comment_defect ON defect_comments(workspace_id, defect_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS defect_activities (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    defect_id BIGINT NOT NULL REFERENCES defect(id) ON DELETE CASCADE,
    verb TEXT NOT NULL CHECK (verb IN ('created','updated','transitioned','attached','linked','commented','verified')),
    field_name TEXT,
    old_value TEXT,
    new_value TEXT,
    actor_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
ALTER TABLE defect_activities ENABLE ROW LEVEL SECURITY;
ALTER TABLE defect_activities FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON defect_activities
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_defect_activity_defect ON defect_activities(workspace_id, defect_id, created_at DESC);
CREATE TABLE defect_activities_default PARTITION OF defect_activities DEFAULT;

-- 跨类型关联表
CREATE TABLE IF NOT EXISTS biz_entity_relations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    project_id BIGINT NOT NULL REFERENCES projects(id),
    source_type TEXT NOT NULL CHECK (source_type IN ('task','requirement','defect')),
    source_id BIGINT NOT NULL,
    target_type TEXT NOT NULL CHECK (target_type IN ('task','requirement','defect')),
    target_id BIGINT NOT NULL,
    relation_type TEXT NOT NULL CHECK (relation_type IN ('implemented_by','relates_to','duplicate','blocked_by','parent_child','found_in','fixed_in','verified_in')),
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_type, source_id, target_type, target_id, relation_type)
);
ALTER TABLE biz_entity_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE biz_entity_relations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON biz_entity_relations
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_biz_entity_rel_source ON biz_entity_relations(workspace_id, source_type, source_id) ;
CREATE INDEX idx_biz_entity_rel_target ON biz_entity_relations(workspace_id, target_type, target_id) ;


-- ============================================================================
-- 全部迁移脚本整合完毕。
-- 至此 sql/ 目录下仅保留 ydsz-plane-init.sql 一个文件。
-- ============================================================================
