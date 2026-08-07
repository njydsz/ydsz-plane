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

-- ----------------------------
-- Records of api_tokens
-- ----------------------------

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

-- ----------------------------
-- Records of issue_watchers
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
  "template" text COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'generic'::text
)
;

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
CREATE INDEX "idx_attachments_uploader" ON "public"."attachments" USING btree (
  "uploaded_by" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
CREATE INDEX "idx_attachments_workspace" ON "public"."attachments" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;

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
CREATE INDEX "idx_audit_logs_ws_time" ON "public"."audit_logs" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);

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
CREATE INDEX "idx_automation_rules_sort" ON "public"."automation_rules" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."int4_ops" ASC NULLS LAST
);
CREATE INDEX "idx_automation_rules_trigger" ON "public"."automation_rules" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "trigger_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE status = 'active'::text;
CREATE INDEX "idx_automation_rules_ws" ON "public"."automation_rules" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table automation_rules
-- ----------------------------
CREATE TRIGGER "trg_automation_rules_updated_at" BEFORE UPDATE ON "public"."automation_rules"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
CREATE INDEX "idx_dashboard_snapshots_refreshed" ON "public"."dashboard_snapshots" USING btree (
  "refreshed_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

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
CREATE INDEX "idx_dashboard_widgets_user" ON "public"."dashboard_widgets" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE user_id IS NOT NULL;

-- ----------------------------
-- Triggers structure for table dashboard_widgets
-- ----------------------------
CREATE TRIGGER "trg_dashboard_widgets_updated_at" BEFORE UPDATE ON "public"."dashboard_widgets"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
CREATE INDEX "idx_deployment_events_ws" ON "public"."deployment_events" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "deployed_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
CREATE UNIQUE INDEX "uq_deployment_events_idempotent" ON "public"."deployment_events" USING btree (
  "deployment_id" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "env" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deployment_id IS NOT NULL;

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
CREATE INDEX "idx_intake_channels_slug" ON "public"."intake_channels" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "slug" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE is_active = true;
CREATE INDEX "idx_intake_channels_workspace" ON "public"."intake_channels" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

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
CREATE INDEX "idx_intake_issues_status" ON "public"."intake_issues" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "idx_intake_issues_submitter" ON "public"."intake_issues" USING btree (
  "submitter_email" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "idx_intake_issues_tracking" ON "public"."intake_issues" USING btree (
  "tracking_id" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "idx_intake_issues_workspace" ON "public"."intake_issues" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);

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
CREATE INDEX "idx_invitations_token" ON "public"."invitations" USING btree (
  "token_hash" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "idx_invitations_workspace" ON "public"."invitations" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE status = 'pending'::text;

-- ----------------------------
-- Triggers structure for table invitations
-- ----------------------------
CREATE TRIGGER "trg_invitations_updated_at" BEFORE UPDATE ON "public"."invitations"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
CREATE INDEX "idx_activities_issue_covering" ON "public"."issue_activities" USING btree (
  "issue_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
CREATE INDEX "idx_activities_project" ON "public"."issue_activities" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);

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
CREATE INDEX "idx_issue_assignees_user" ON "public"."issue_assignees" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

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
CREATE INDEX "idx_issue_comments_issue" ON "public"."issue_comments" USING btree (
  "issue_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

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
CREATE INDEX "idx_issue_deps_succ" ON "public"."issue_dependencies" USING btree (
  "successor_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

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
CREATE INDEX "idx_issue_relations_target" ON "public"."issue_relations" USING btree (
  "target_issue_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

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

-- ----------------------------
-- Primary Key structure for table issue_watchers
-- ----------------------------
ALTER TABLE "public"."issue_watchers" ADD CONSTRAINT "issue_watchers_pkey" PRIMARY KEY ("issue_id", "user_id");

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
CREATE INDEX "idx_issues_fix_version" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "fix_version_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND fix_version_id IS NOT NULL;
CREATE INDEX "idx_issues_found_version" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "found_version_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND found_version_id IS NOT NULL;
CREATE INDEX "idx_issues_list_covering" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "updated_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE deleted_at IS NULL;
CREATE INDEX "idx_issues_parent" ON "public"."issues" USING btree (
  "parent_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND parent_id IS NOT NULL;
CREATE INDEX "idx_issues_priority_covering" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "priority" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "updated_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE deleted_at IS NULL AND (priority = ANY (ARRAY['urgent'::text, 'high'::text]));
CREATE UNIQUE INDEX "idx_issues_project_sequence" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sequence_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
CREATE INDEX "idx_issues_project_state" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "state_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."float8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX "idx_issues_public_id" ON "public"."issues" USING btree (
  "public_id" "pg_catalog"."uuid_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
CREATE INDEX "idx_issues_release_version" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "release_version_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND release_version_id IS NOT NULL;
CREATE INDEX "idx_issues_search_tsv" ON "public"."issues" USING gin (
  "search_tsv" "pg_catalog"."tsvector_ops"
);
CREATE INDEX "idx_issues_state_covering" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "state_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."float8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
CREATE INDEX "idx_issues_target_date" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "target_date" "pg_catalog"."date_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND completed_at IS NULL;
CREATE INDEX "idx_issues_target_date_covering" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "target_date" "pg_catalog"."date_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND target_date IS NOT NULL;
CREATE INDEX "idx_issues_type" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "type_code" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
CREATE INDEX "idx_issues_type_covering" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "type_code" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE deleted_at IS NULL;
CREATE INDEX "idx_issues_updated" ON "public"."issues" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "updated_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
CREATE INDEX "idx_issues_workspace_project" ON "public"."issues" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;

-- ----------------------------
-- Triggers structure for table issues
-- ----------------------------
CREATE TRIGGER "trg_issue_search_cleanup" AFTER UPDATE OF "deleted_at" ON "public"."issues"
FOR EACH ROW
WHEN ((new.deleted_at IS NOT NULL))
EXECUTE PROCEDURE "public"."fn_cleanup_search_document"();
CREATE TRIGGER "trg_issue_search_sync" AFTER INSERT OR UPDATE OF "name", "description_stripped" ON "public"."issues"
FOR EACH ROW
WHEN ((new.deleted_at IS NULL))
EXECUTE PROCEDURE "public"."fn_refresh_search_document"();
CREATE TRIGGER "trg_issues_updated_at" BEFORE UPDATE ON "public"."issues"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
ALTER TABLE "public"."issues" ADD CONSTRAINT "issues_type_code_check" CHECK (type_code = ANY (ARRAY['requirement'::text, 'task'::text, 'defect'::text]));
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

-- ----------------------------
-- Triggers structure for table labels
-- ----------------------------
CREATE TRIGGER "trg_labels_updated_at" BEFORE UPDATE ON "public"."labels"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
CREATE INDEX "idx_metric_snap_date" ON "public"."metric_snapshots" USING btree (
  "snapshot_date" "pg_catalog"."date_ops" ASC NULLS LAST
);
CREATE INDEX "idx_metric_snap_lookup" ON "public"."metric_snapshots" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "metric" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "snapshot_date" "pg_catalog"."date_ops" ASC NULLS LAST
);
CREATE INDEX "idx_metric_snap_ws" ON "public"."metric_snapshots" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "metric" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "snapshot_date" "pg_catalog"."date_ops" DESC NULLS FIRST
);
CREATE UNIQUE INDEX "idx_metric_snapshots_project" ON "public"."metric_snapshots" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "granularity" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "ref_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "metric" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "snapshot_date" "pg_catalog"."date_ops" ASC NULLS LAST
) WHERE project_id IS NOT NULL;
CREATE UNIQUE INDEX "idx_metric_snapshots_workspace" ON "public"."metric_snapshots" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "granularity" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "metric" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "snapshot_date" "pg_catalog"."date_ops" ASC NULLS LAST
) WHERE project_id IS NULL;

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

-- ----------------------------
-- Triggers structure for table modules
-- ----------------------------
CREATE TRIGGER "trg_modules_updated_at" BEFORE UPDATE ON "public"."modules"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
CREATE INDEX "idx_deliveries_notification" ON "public"."notification_deliveries" USING btree (
  "notification_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_deliveries_status" ON "public"."notification_deliveries" USING btree (
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
) WHERE status::text = 'pending'::text;

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
CREATE UNIQUE INDEX "idx_notification_digests_pending" ON "public"."notification_digests" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "digest_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE status::text = 'pending'::text;

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
CREATE INDEX "idx_notifications_entity" ON "public"."notifications" USING btree (
  "entity_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "entity_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_notifications_recipient_unread" ON "public"."notifications" USING btree (
  "recipient_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "is_read" "pg_catalog"."bool_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE is_archived = false;

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
CREATE UNIQUE INDEX "idx_password_reset_tokens_user_active" ON "public"."password_reset_tokens" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE used_at IS NULL;

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
CREATE INDEX "idx_projects_workspace" ON "public"."projects" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX "idx_projects_workspace_identifier" ON "public"."projects" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "identifier" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX "idx_projects_workspace_slug" ON "public"."projects" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "slug" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;

-- ----------------------------
-- Triggers structure for table projects
-- ----------------------------
CREATE TRIGGER "trg_projects_updated_at" BEFORE UPDATE ON "public"."projects"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
-- Primary Key structure for table projects
-- ----------------------------
ALTER TABLE "public"."projects" ADD CONSTRAINT "projects_pkey" PRIMARY KEY ("id");

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
CREATE INDEX "idx_recent_items_ws" ON "public"."recent_items" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table recent_items
-- ----------------------------
CREATE TRIGGER "trg_recent_items_touch" BEFORE UPDATE ON "public"."recent_items"
FOR EACH ROW
EXECUTE PROCEDURE "public"."fn_touch_recent_item"();

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
CREATE INDEX "idx_risk_alerts_unresolved" ON "public"."risk_alerts" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "severity" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE NOT is_resolved;
CREATE INDEX "idx_risk_alerts_workspace" ON "public"."risk_alerts" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE NOT is_resolved;

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
CREATE INDEX "idx_risk_rules_project" ON "public"."risk_rules" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE project_id IS NOT NULL;
CREATE INDEX "idx_risk_rules_workspace" ON "public"."risk_rules" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table risk_rules
-- ----------------------------
CREATE TRIGGER "trg_risk_rules_updated_at" BEFORE UPDATE ON "public"."risk_rules"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
CREATE INDEX "idx_rule_executions_project" ON "public"."rule_executions" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
) WHERE project_id IS NOT NULL;
CREATE INDEX "idx_rule_executions_rule" ON "public"."rule_executions" USING btree (
  "rule_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
CREATE INDEX "idx_rule_executions_ws" ON "public"."rule_executions" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "created_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);

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
CREATE INDEX "idx_search_bookmarks_user" ON "public"."search_bookmarks" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."float8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_search_bookmarks_ws" ON "public"."search_bookmarks" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table search_bookmarks
-- ----------------------------
CREATE TRIGGER "trg_search_bookmarks_updated_at" BEFORE UPDATE ON "public"."search_bookmarks"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
CREATE INDEX "idx_search_documents_tsv" ON "public"."search_documents" USING gin (
  "search_tsv" "pg_catalog"."tsvector_ops"
);
CREATE UNIQUE INDEX "idx_search_documents_unique" ON "public"."search_documents" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "doc_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST,
  "doc_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_search_documents_ws" ON "public"."search_documents" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "doc_type" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table search_documents
-- ----------------------------
CREATE TRIGGER "trg_search_documents_updated_at" BEFORE UPDATE ON "public"."search_documents"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
CREATE INDEX "idx_search_history_ws_user" ON "public"."search_history" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

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
CREATE INDEX "idx_sprintsnapshots_project" ON "public"."sprint_snapshots" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
CREATE INDEX "idx_sprintsnapshots_sprint_date" ON "public"."sprint_snapshots" USING btree (
  "sprint_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "snapshot_date" "pg_catalog"."date_ops" ASC NULLS LAST
);

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
CREATE INDEX "idx_sprints_active_unique" ON "public"."sprints" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE status = 'active'::text AND deleted_at IS NULL;
CREATE INDEX "idx_sprints_project_status" ON "public"."sprints" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "status" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
CREATE INDEX "idx_sprints_version" ON "public"."sprints" USING btree (
  "version_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;

-- ----------------------------
-- Triggers structure for table sprints
-- ----------------------------
CREATE TRIGGER "trg_sprints_updated_at" BEFORE UPDATE ON "public"."sprints"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
CREATE TRIGGER "trg_sprint_search_sync" AFTER INSERT OR UPDATE OF "name", "goal" ON "public"."sprints"
FOR EACH ROW
WHEN ((new.deleted_at IS NULL))
EXECUTE PROCEDURE "public"."fn_refresh_sprint_search_document"();
CREATE TRIGGER "trg_sprint_search_cleanup" AFTER UPDATE OF "deleted_at" ON "public"."sprints"
FOR EACH ROW
WHEN ((new.deleted_at IS NOT NULL))
EXECUTE PROCEDURE "public"."fn_cleanup_search_document"();

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

-- ----------------------------
-- Triggers structure for table states
-- ----------------------------
CREATE TRIGGER "trg_states_updated_at" BEFORE UPDATE ON "public"."states"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
CREATE INDEX "idx_time_logs_user_date" ON "public"."time_logs" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "spent_date" "pg_catalog"."date_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;

-- ----------------------------
-- Triggers structure for table time_logs
-- ----------------------------
CREATE TRIGGER "trg_time_logs_updated_at" BEFORE UPDATE ON "public"."time_logs"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
CREATE INDEX "idx_vds_workspace" ON "public"."version_delivery_snapshots" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

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
CREATE UNIQUE INDEX "idx_versions_unique_semver" ON "public"."versions" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "semver" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
CREATE INDEX "idx_versions_workspace" ON "public"."versions" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;

-- ----------------------------
-- Triggers structure for table versions
-- ----------------------------
CREATE TRIGGER "trg_versions_bump_version" BEFORE UPDATE ON "public"."versions"
FOR EACH ROW
EXECUTE PROCEDURE "public"."bump_version"();
CREATE TRIGGER "trg_versions_updated_at" BEFORE UPDATE ON "public"."versions"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();
CREATE TRIGGER "trg_version_search_sync" AFTER INSERT OR UPDATE OF "name", "description" ON "public"."versions"
FOR EACH ROW
WHEN ((new.deleted_at IS NULL))
EXECUTE PROCEDURE "public"."fn_refresh_version_search_document"();
CREATE TRIGGER "trg_version_search_cleanup" AFTER UPDATE OF "deleted_at" ON "public"."versions"
FOR EACH ROW
WHEN ((new.deleted_at IS NOT NULL))
EXECUTE PROCEDURE "public"."fn_cleanup_search_document"();

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
CREATE INDEX "idx_webhook_logs_occurred" ON "public"."webhook_logs" USING btree (
  "occurred_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE INDEX "idx_webhook_logs_webhook" ON "public"."webhook_logs" USING btree (
  "webhook_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "occurred_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);
CREATE INDEX "idx_webhook_logs_workspace" ON "public"."webhook_logs" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "occurred_at" "pg_catalog"."timestamptz_ops" DESC NULLS FIRST
);

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
CREATE INDEX "idx_webhooks_project" ON "public"."webhooks" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE project_id IS NOT NULL;
CREATE INDEX "idx_webhooks_workspace" ON "public"."webhooks" USING btree (
  "workspace_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

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
CREATE INDEX "idx_workbench_configs_user" ON "public"."workbench_configs" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_workbench_configs_user_project" ON "public"."workbench_configs" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  COALESCE(project_id, 0::bigint) "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Triggers structure for table workbench_configs
-- ----------------------------
CREATE TRIGGER "trg_workbench_configs_updated_at" BEFORE UPDATE ON "public"."workbench_configs"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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

-- ----------------------------
-- Triggers structure for table workspaces
-- ----------------------------
CREATE TRIGGER "trg_workspaces_updated_at" BEFORE UPDATE ON "public"."workspaces"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
CREATE INDEX "idx_pages_parent" ON "public"."pages" USING btree (
  "parent_id" "pg_catalog"."int8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL AND parent_id IS NOT NULL;
CREATE INDEX "idx_pages_project_sort" ON "public"."pages" USING btree (
  "project_id" "pg_catalog"."int8_ops" ASC NULLS LAST,
  "sort_order" "pg_catalog"."float8_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX "idx_pages_public_id" ON "public"."pages" USING btree (
  "public_id" "pg_catalog"."uuid_ops" ASC NULLS LAST
) WHERE deleted_at IS NULL;

-- ----------------------------
-- Triggers structure for table pages
-- ----------------------------
CREATE TRIGGER "trg_pages_updated_at" BEFORE UPDATE ON "public"."pages"
FOR EACH ROW
EXECUTE PROCEDURE "public"."set_updated_at"();

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
