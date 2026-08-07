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
COMMENT ON TABLE public.issues IS '[核心工作项域]工作项主表（承载需求/任务/缺陷三类工作项，项目内 sequence_id 唯一标识如 YD-123，支持3级父子层级，乐观锁并发控制）';

COMMENT ON COLUMN public.issues.id IS '自增主键（内部使用，不对外暴露）';
COMMENT ON COLUMN public.issues.public_id IS '对外暴露的 UUID 主键，API 与前端使用此字段';
COMMENT ON COLUMN public.issues.workspace_id IS '关联 workspaces.id（租户隔离列，RLS 策略依据）';
COMMENT ON COLUMN public.issues.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.issues.sequence_id IS '项目内自增序列号，配合 identifier 展示为 YD-123 格式';
COMMENT ON COLUMN public.issues.type_code IS '工作项类型: requirement(需求) / task(任务) / defect(缺陷)';
COMMENT ON COLUMN public.issues.parent_id IS '关联 issues.id（父工作项，支持层级结构）';
COMMENT ON COLUMN public.issues.depth IS '层级深度: 1=顶层 / 2=子项 / 3=孙项，最大 3';
COMMENT ON COLUMN public.issues.name IS '工作项标题';
COMMENT ON COLUMN public.issues.description_json IS '工作项描述的 TipTap JSON 格式（富文本编辑器原始数据，文档节点树结构）';
COMMENT ON COLUMN public.issues.description_html IS '工作项描述的 HTML 渲染结果';
COMMENT ON COLUMN public.issues.description_stripped IS '纯文本摘要（去除富文本标记，供全文检索使用）';
COMMENT ON COLUMN public.issues.state_id IS '关联 states.id（当前状态）';
COMMENT ON COLUMN public.issues.priority IS '优先级: urgent(紧急) / high(高) / medium(中) / low(低) / none(无)';
COMMENT ON COLUMN public.issues.severity IS '严重程度（缺陷专有）: 1=致命 / 2=严重 / 3=一般 / 4=轻微 / 5=建议';
COMMENT ON COLUMN public.issues.found_phase IS '发现阶段（缺陷专有）: unit(单元测试) / integration(集成测试) / uat(验收测试) / production(生产环境) / customer(客户反馈)';
COMMENT ON COLUMN public.issues.root_cause_category IS '根因分类（缺陷专有）: requirement(需求问题) / technical(技术问题) / environment(环境问题) / data(数据问题)';
COMMENT ON COLUMN public.issues.verifier_id IS '关联 users.id（缺陷验证人）';
COMMENT ON COLUMN public.issues.environment IS '环境信息 JSON（缺陷专有），结构: {os, browser, version, ...}';
COMMENT ON COLUMN public.issues.reproduce_steps IS '复现步骤 JSON（缺陷专有），结构: {steps: [], expected: "", actual: ""}';
COMMENT ON COLUMN public.issues.category IS '工作项分类（任务专有）: frontend / backend / qa / doc / design / other';
COMMENT ON COLUMN public.issues.actual_effort IS '实际花费工时（小时，NUMERIC(8,2)）';
COMMENT ON COLUMN public.issues.remaining_effort IS '剩余预估工时（小时，NUMERIC(8,2)）';
COMMENT ON COLUMN public.issues.delay_reason IS '延期原因（任务专有）: requirement_change(需求变更) / resource(资源不足) / blocked(被阻塞) / other(其他)';
COMMENT ON COLUMN public.issues.source IS '需求来源（需求专有）: customer(客户) / internal(内部) / competitor(竞品)';
COMMENT ON COLUMN public.issues.point IS '故事点（0-12，敏捷估算用）';
COMMENT ON COLUMN public.issues.sprint_id IS '关联 sprints.id（所属迭代）';
COMMENT ON COLUMN public.issues.progress IS '完成进度百分比（0-100）';
COMMENT ON COLUMN public.issues.start_date IS '实际开始日期';
COMMENT ON COLUMN public.issues.target_date IS '目标完成日期';
COMMENT ON COLUMN public.issues.completed_at IS '实际完成时间（含时区）';
COMMENT ON COLUMN public.issues.is_draft IS '是否为草稿: true=草稿 / false=已发布';
COMMENT ON COLUMN public.issues.sort_order IS '看板列内排序权重（默认 65535，越小越靠前）';
COMMENT ON COLUMN public.issues.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.issues.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.issues.updated_at IS '最后更新时间（含时区，由触发器自动维护）';
COMMENT ON COLUMN public.issues.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.issues.version IS '乐观锁版本号（每次更新自增，冲突返回 409）';
COMMENT ON COLUMN public.issues.found_version_id IS '关联 versions.id（发现缺陷时的版本）';
COMMENT ON COLUMN public.issues.fix_version_id IS '关联 versions.id（计划修复的版本）';
COMMENT ON COLUMN public.issues.release_version_id IS '关联 versions.id（首次发布的版本）';
COMMENT ON COLUMN public.issues.search_tsv IS '全文检索向量（自动生成，simple 配置，供 ES 降级使用）';

-- ============================================================
-- 表: labels — 标签表
-- ============================================================
COMMENT ON TABLE public.labels IS '[核心工作项域]项目级标签（用于工作项分类、筛选与可视化标识）';

COMMENT ON COLUMN public.labels.id IS '自增主键';
COMMENT ON COLUMN public.labels.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.labels.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.labels.name IS '标签名称（项目内唯一展示名）';
COMMENT ON COLUMN public.labels.color IS '标签颜色（十六进制，如 #8DA2C2）';
COMMENT ON COLUMN public.labels.description IS '标签说明文本';
COMMENT ON COLUMN public.labels.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.labels.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.labels.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.labels.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: modules — 模块/组件表
-- ============================================================
COMMENT ON TABLE public.modules IS '[核心工作项域]项目模块/组件分解结构（将项目划分为功能模块，每模块可关联工作项）';

COMMENT ON COLUMN public.modules.id IS '自增主键';
COMMENT ON COLUMN public.modules.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.modules.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.modules.name IS '模块名称';
COMMENT ON COLUMN public.modules.description IS '模块描述';
COMMENT ON COLUMN public.modules.lead_id IS '关联 users.id（模块负责人）';
COMMENT ON COLUMN public.modules.status IS '模块状态: active(进行中) / completed(已完成) / cancelled(已取消)';
COMMENT ON COLUMN public.modules.start_date IS '模块开始日期';
COMMENT ON COLUMN public.modules.target_date IS '模块目标日期';
COMMENT ON COLUMN public.modules.sort_order IS '排序权重（默认 65535，越小越靠前）';
COMMENT ON COLUMN public.modules.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.modules.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.modules.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.modules.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: issue_labels — 工作项-标签关联表
-- ============================================================
COMMENT ON TABLE public.issue_labels IS '[核心工作项域]工作项与标签的多对多关联关系';

COMMENT ON COLUMN public.issue_labels.issue_id IS '关联 issues.id（工作项）';
COMMENT ON COLUMN public.issue_labels.label_id IS '关联 labels.id（标签）';

-- ============================================================
-- 表: issue_modules — 工作项-模块关联表
-- ============================================================
COMMENT ON TABLE public.issue_modules IS '[核心工作项域]工作项与模块的多对多关联关系';

COMMENT ON COLUMN public.issue_modules.issue_id IS '关联 issues.id（工作项）';
COMMENT ON COLUMN public.issue_modules.module_id IS '关联 modules.id（模块）';

-- ============================================================
-- 表: issue_assignees — 工作项指派人员表
-- ============================================================
COMMENT ON TABLE public.issue_assignees IS '[核心工作项域]工作项负责人多对多关联（一个工作项可指派多人）';

COMMENT ON COLUMN public.issue_assignees.issue_id IS '关联 issues.id（工作项）';
COMMENT ON COLUMN public.issue_assignees.user_id IS '关联 users.id（被指派人员）';
COMMENT ON COLUMN public.issue_assignees.assigned_at IS '指派时间（含时区）';
COMMENT ON COLUMN public.issue_assignees.assigned_by IS '关联 users.id（执行指派操作的管理员）';

-- ============================================================
-- 表: issue_watchers — 工作项关注者表
-- ============================================================
COMMENT ON TABLE public.issue_watchers IS '[核心工作项域]工作项关注者多对多关联（关注者接收该工作项变更通知）';

COMMENT ON COLUMN public.issue_watchers.issue_id IS '关联 issues.id（被关注的工作项）';
COMMENT ON COLUMN public.issue_watchers.user_id IS '关联 users.id（关注者）';
COMMENT ON COLUMN public.issue_watchers.created_at IS '关注时间（含时区）';

-- ============================================================
-- 表: issue_comments — 工作项评论表（已有注释，补充缺失字段）
-- ============================================================
COMMENT ON COLUMN public.issue_comments.id IS '自增主键';
COMMENT ON COLUMN public.issue_comments.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.issue_comments.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.issue_comments.issue_id IS '关联 issues.id（所属工作项）';
COMMENT ON COLUMN public.issue_comments.content_json IS 'TipTap 编辑器的 JSON 输出';
COMMENT ON COLUMN public.issue_comments.content_html IS '评论的 HTML 渲染结果';
COMMENT ON COLUMN public.issue_comments.content_stripped IS '评论纯文本摘要（去除富文本标记，供搜索使用）';
COMMENT ON COLUMN public.issue_comments.created_by IS '关联 users.id（评论作者）';
COMMENT ON COLUMN public.issue_comments.mentions IS '@提及的用户 ID 数组';
COMMENT ON COLUMN public.issue_comments.parent_id IS '父评论 ID（嵌套回复）';
COMMENT ON COLUMN public.issue_comments.is_edited IS '是否已被编辑: true=已编辑 / false=原文';
COMMENT ON COLUMN public.issue_comments.edited_at IS '最后编辑时间（含时区）';
COMMENT ON COLUMN public.issue_comments.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.issue_comments.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: issue_activities — 工作项活动表（按月分区）
-- ============================================================
COMMENT ON TABLE public.issue_activities IS '[核心工作项域]工作项变更历史流水（按月分区归档，记录所有字段变更事件）';

COMMENT ON COLUMN public.issue_activities.id IS '自增主键';
COMMENT ON COLUMN public.issue_activities.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.issue_activities.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.issue_activities.issue_id IS '关联 issues.id（所属工作项）';
COMMENT ON COLUMN public.issue_activities.verb IS '操作类型: created(创建) / updated(更新) / transitioned(状态流转) / attached(附件) / linked(关联) / unlinked(取消关联) / commented(评论)';
COMMENT ON COLUMN public.issue_activities.field IS '变更的字段名（verb=updated 时有效）';
COMMENT ON COLUMN public.issue_activities.old_value IS '字段变更前的值（纯文本）';
COMMENT ON COLUMN public.issue_activities.new_value IS '字段变更后的值（纯文本）';
COMMENT ON COLUMN public.issue_activities.old_ref IS '变更前的复杂引用 JSON（如状态对象 {id, name, group}）';
COMMENT ON COLUMN public.issue_activities.new_ref IS '变更后的复杂引用 JSON';
COMMENT ON COLUMN public.issue_activities.actor_id IS '关联 users.id（操作执行人）';
COMMENT ON COLUMN public.issue_activities.actor_email IS '操作人邮箱（冗余存储，方便审计展示）';
COMMENT ON COLUMN public.issue_activities.actor_name IS '操作人显示名（冗余存储）';
COMMENT ON COLUMN public.issue_activities.created_at IS '活动发生时间（含时区）';

-- ============================================================
-- 表: issue_dependencies — 工作项依赖关系表
-- ============================================================
COMMENT ON TABLE public.issue_dependencies IS '[核心工作项域]工作项间依赖关系（FS/SS/FF/SF 四种工程依赖类型，支持 lag_days 延迟）';

COMMENT ON COLUMN public.issue_dependencies.id IS '自增主键';
COMMENT ON COLUMN public.issue_dependencies.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.issue_dependencies.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.issue_dependencies.predecessor_id IS '关联 issues.id（前置工作项，即依赖的源头）';
COMMENT ON COLUMN public.issue_dependencies.successor_id IS '关联 issues.id（后续工作项，即被依赖方）';
COMMENT ON COLUMN public.issue_dependencies.dependency_type IS '依赖类型: FS(完成-开始) / SS(开始-开始) / FF(完成-完成) / SF(开始-完成)';
COMMENT ON COLUMN public.issue_dependencies.lag_days IS '延迟天数（正数表示延后，负数表示提前）';
COMMENT ON COLUMN public.issue_dependencies.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.issue_dependencies.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: issue_relations — 工作项关联关系表
-- ============================================================
COMMENT ON TABLE public.issue_relations IS '[核心工作项域]工作项间的语义关联（重复、相关、阻塞、顺序、实现等多种关系类型）';

COMMENT ON COLUMN public.issue_relations.id IS '自增主键';
COMMENT ON COLUMN public.issue_relations.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.issue_relations.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.issue_relations.source_issue_id IS '关联 issues.id（源工作项）';
COMMENT ON COLUMN public.issue_relations.target_issue_id IS '关联 issues.id（目标工作项）';
COMMENT ON COLUMN public.issue_relations.relation_type IS '关联类型: duplicate(重复) / relates_to(相关) / blocked_by(被阻塞) / start_before(先于开始) / finish_before(先于完成) / implemented_by(由...实现)';
COMMENT ON COLUMN public.issue_relations.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.issue_relations.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: time_logs — 工时记录表
-- ============================================================
COMMENT ON TABLE public.time_logs IS '[核心工作项域]工作项工时记录（用户每日填报的实际花费时间，单位: 分钟）';

COMMENT ON COLUMN public.time_logs.id IS '自增主键';
COMMENT ON COLUMN public.time_logs.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.time_logs.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.time_logs.issue_id IS '关联 issues.id（工时所属工作项）';
COMMENT ON COLUMN public.time_logs.user_id IS '关联 users.id（工时填报人）';
COMMENT ON COLUMN public.time_logs.spent_date IS '工时消耗日期（默认当天）';
COMMENT ON COLUMN public.time_logs.duration_minutes IS '花费时长（分钟，范围 1-1440）';
COMMENT ON COLUMN public.time_logs.description IS '工时描述（工作内容说明）';
COMMENT ON COLUMN public.time_logs.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.time_logs.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.time_logs.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: attachments — 附件表
-- ============================================================
COMMENT ON TABLE public.attachments IS '[核心工作项域]多态附件存储（支持关联到 issue/comment/workspace/project 四种实体）';

COMMENT ON COLUMN public.attachments.id IS '自增主键';
COMMENT ON COLUMN public.attachments.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.attachments.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.attachments.entity_type IS '关联实体类型: issue(工作项) / comment(评论) / workspace(工作空间) / project(项目)';
COMMENT ON COLUMN public.attachments.entity_id IS '关联实体的 ID（配合 entity_type 组成多态关联）';
COMMENT ON COLUMN public.attachments.file_name IS '原始文件名';
COMMENT ON COLUMN public.attachments.file_size IS '文件大小（字节）';
COMMENT ON COLUMN public.attachments.content_type IS 'MIME 类型（默认 application/octet-stream）';
COMMENT ON COLUMN public.attachments.storage_key IS '对象存储中的文件路径/key';
COMMENT ON COLUMN public.attachments.storage_url IS '文件访问 URL';
COMMENT ON COLUMN public.attachments.thumb_key IS '缩略图存储 key（图片/视频文件专用）';
COMMENT ON COLUMN public.attachments.uploaded_by IS '关联 users.id（上传者）';
COMMENT ON COLUMN public.attachments.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.attachments.created_at IS '上传时间（含时区）';
COMMENT ON COLUMN public.attachments.updated_at IS '最后更新时间（含时区）';

-- ============================================================================
-- 项目管理域
-- ============================================================================

-- ============================================================
-- 表: projects — 项目表
-- ============================================================
COMMENT ON TABLE public.projects IS '[项目管理域]工作空间下的项目（最小业务容器，所有工作项归属项目）';

COMMENT ON COLUMN public.projects.id IS '自增主键';
COMMENT ON COLUMN public.projects.workspace_id IS '关联 workspaces.id（所属工作空间）';
COMMENT ON COLUMN public.projects.public_id IS '对外暴露的 UUID 主键';
COMMENT ON COLUMN public.projects.name IS '项目名称';
COMMENT ON COLUMN public.projects.slug IS '项目 URL 标识（小写字母+数字+短横线，空间内唯一）';

COMMENT ON COLUMN public.projects.identifier IS '项目标识符（大写字母，如 YD，用于工作项编号前缀）';
COMMENT ON COLUMN public.projects.description IS '项目描述';
COMMENT ON COLUMN public.projects.network IS '可见性: public(空间内所有成员可见) / private(仅项目成员可见)';
COMMENT ON COLUMN public.projects.icon IS '项目图标（Lucide 图标名或 emoji）';
COMMENT ON COLUMN public.projects.color IS '项目颜色（十六进制）';
COMMENT ON COLUMN public.projects.status IS '项目状态: active(进行中) / archived(已归档)';
COMMENT ON COLUMN public.projects.sort_order IS '排序权重（默认 65535）';
COMMENT ON COLUMN public.projects.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.projects.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.projects.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.projects.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.projects.template IS '项目模板: agile(敏捷) / waterfall(瀑布) / generic(通用)';

-- ============================================================
-- 表: sprints — 迭代/冲刺表
-- ============================================================
COMMENT ON TABLE public.sprints IS '[项目管理域]敏捷迭代/冲刺（包含容量规划、目标设定、状态流转，乐观锁并发控制）';

COMMENT ON COLUMN public.sprints.id IS '自增主键';
COMMENT ON COLUMN public.sprints.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.sprints.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.sprints.name IS '迭代名称';
COMMENT ON COLUMN public.sprints.description IS '迭代描述';
COMMENT ON COLUMN public.sprints.goal IS '迭代目标（Sprint Goal）';
COMMENT ON COLUMN public.sprints.status IS '迭代状态: planned(未开始) / active(进行中) / completed(已完成)';
COMMENT ON COLUMN public.sprints.start_date IS '迭代开始日期';
COMMENT ON COLUMN public.sprints.end_date IS '迭代结束日期';
COMMENT ON COLUMN public.sprints.capacity IS '迭代容量（团队可用工时，NUMERIC(10,2) 小时）';
COMMENT ON COLUMN public.sprints.owner_id IS '关联 users.id（迭代负责人/Scrum Master）';
COMMENT ON COLUMN public.sprints.viewport IS '迭代视图配置 JSON（组件布局、筛选条件等）';
COMMENT ON COLUMN public.sprints.review_snapshot IS '评审快照 JSON（评审会议数据）';
COMMENT ON COLUMN public.sprints.started_at IS '实际启动时间（含时区）';
COMMENT ON COLUMN public.sprints.completed_at IS '完成时间（含时区）';
COMMENT ON COLUMN public.sprints.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.sprints.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.sprints.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.sprints.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.sprints.version_id IS '关联 versions.id（首次发布版本）';

-- ============================================================
-- 表: sprint_issues — 迭代-工作项关联表
-- ============================================================
COMMENT ON TABLE public.sprint_issues IS '[项目管理域]迭代与工作项的多对多关联（含中途加入标记与排序）';

COMMENT ON COLUMN public.sprint_issues.sprint_id IS '关联 sprints.id（所属迭代）';
COMMENT ON COLUMN public.sprint_issues.issue_id IS '关联 issues.id（工作项）';
COMMENT ON COLUMN public.sprint_issues.added_midway IS '是否中途加入迭代: true=迭代启动后加入 / false=规划时加入';
COMMENT ON COLUMN public.sprint_issues.sort_order IS '迭代内排序权重（默认 65535）';
COMMENT ON COLUMN public.sprint_issues.added_at IS '加入迭代时间（含时区）';
COMMENT ON COLUMN public.sprint_issues.added_by IS '关联 users.id（执行加入操作的人）';

-- ============================================================
-- 表: sprint_snapshots — 迭代快照表
-- ============================================================
COMMENT ON TABLE public.sprint_snapshots IS '[项目管理域]迭代燃尽图/速率等每日快照数据（JSONB 压缩存储历史指标）';

COMMENT ON COLUMN public.sprint_snapshots.id IS '自增主键';
COMMENT ON COLUMN public.sprint_snapshots.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.sprint_snapshots.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.sprint_snapshots.sprint_id IS '关联 sprints.id（所属迭代）';
COMMENT ON COLUMN public.sprint_snapshots.snapshot_date IS '快照日期';
COMMENT ON COLUMN public.sprint_snapshots.data IS '快照数据 JSON（burndown 点数、完成率、速率等指标）';
COMMENT ON COLUMN public.sprint_snapshots.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.sprint_snapshots.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: versions — 版本发布表（已有 version 字段注释，补充其他字段）
-- ============================================================
COMMENT ON TABLE public.versions IS '[项目管理域]版本发布管理（支持 SemVer 规范，含检查清单与发布说明）';

COMMENT ON COLUMN public.versions.id IS '自增主键';
COMMENT ON COLUMN public.versions.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.versions.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.versions.name IS '版本名称';
COMMENT ON COLUMN public.versions.semver IS '语义化版本号（如 1.2.3-beta.1，符合 SemVer 规范）';
COMMENT ON COLUMN public.versions.description IS '版本描述';
COMMENT ON COLUMN public.versions.status IS '版本状态: planning(规划中) / active(开发中) / released(已发布) / archived(已归档)';
COMMENT ON COLUMN public.versions.checklist IS '发布检查清单 JSON 数组（最多 50 项）';
COMMENT ON COLUMN public.versions.release_notes IS '发布说明';
COMMENT ON COLUMN public.versions.delivered_at IS '实际交付时间（含时区）';
COMMENT ON COLUMN public.versions.target_date IS '目标发布日期';
COMMENT ON COLUMN public.versions.archived_at IS '归档时间（含时区）';
COMMENT ON COLUMN public.versions.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.versions.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.versions.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.versions.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.versions.start_date IS '开发开始日期';
COMMENT ON COLUMN public.versions.end_date IS '开发结束日期';

-- ============================================================
-- 表: version_delivery_snapshots — 版本交付快照表
-- ============================================================
COMMENT ON TABLE public.version_delivery_snapshots IS '[项目管理域]版本交付过程记录（保存进度与质量维度的时间线数据）';

COMMENT ON COLUMN public.version_delivery_snapshots.id IS '自增主键';
COMMENT ON COLUMN public.version_delivery_snapshots.version_id IS '关联 versions.id（所属版本）';
COMMENT ON COLUMN public.version_delivery_snapshots.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.version_delivery_snapshots.progress IS '进度快照 JSON（完成率、阻塞数、剩余工作项等）';
COMMENT ON COLUMN public.version_delivery_snapshots.quality IS '质量快照 JSON（缺陷数、解决率、回归率等）';
COMMENT ON COLUMN public.version_delivery_snapshots.release_notes IS '当时版本的发布说明快照';
COMMENT ON COLUMN public.version_delivery_snapshots.snapshot_at IS '快照记录时间（含时区）';

-- ============================================================================
-- 工作流域
-- ============================================================================

-- ============================================================
-- 表: states — 状态表
-- ============================================================
COMMENT ON TABLE public.states IS '[工作流域]项目工作项状态定义（每个项目可自定义状态集，状态归属到 4 大分组之一）';

COMMENT ON COLUMN public.states.id IS '自增主键';
COMMENT ON COLUMN public.states.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.states.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.states.name IS '状态名称（如 进行中、待测试）';
COMMENT ON COLUMN public.states."group" IS '状态分组: backlog(待办) / started(进行中) / completed(已完成) / cancelled(已取消)';
COMMENT ON COLUMN public.states.color IS '状态颜色标识（十六进制，如 #8DA2C2）';
COMMENT ON COLUMN public.states.sequence IS '状态排序权重（默认 65535，越小越靠前）';
COMMENT ON COLUMN public.states.is_default IS '是否为新建工作项的默认状态: true=默认 / false=非默认';
COMMENT ON COLUMN public.states.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.states.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.states.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.states.template_set IS '模板归属集合: dev_flow(研发流) / defect_flow(缺陷流) / requirement_flow(需求评审流) / custom(自定义)';

-- ============================================================
-- 表: state_transitions — 状态流转规则表
-- ============================================================
COMMENT ON TABLE public.state_transitions IS '[工作流域]状态流转规则（定义 from→to 的合法路径、必填字段与权限约束）';

COMMENT ON COLUMN public.state_transitions.id IS '自增主键';
COMMENT ON COLUMN public.state_transitions.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.state_transitions.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.state_transitions.type_code IS '适用的工作项类型: requirement/task/defect/all（all 表示所有类型）';
COMMENT ON COLUMN public.state_transitions.from_state_id IS '关联 states.id（起始状态）';
COMMENT ON COLUMN public.state_transitions.to_state_id IS '关联 states.id（目标状态）';
COMMENT ON COLUMN public.state_transitions.required_fields IS '流转时需要填写的字段列表 JSON 数组（如 ["root_cause_category"]）';
COMMENT ON COLUMN public.state_transitions.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.state_transitions.updated_at IS '最后更新时间（含时区）';

-- ============================================================================
-- 租户与权限域
-- ============================================================================

-- ============================================================
-- 表: users — 用户表
-- ============================================================
COMMENT ON TABLE public.users IS '[租户与权限域]系统用户（平台级账号，可跨工作空间加入多个空间）';

COMMENT ON COLUMN public.users.id IS '自增主键';
COMMENT ON COLUMN public.users.public_id IS '对外暴露的 UUID 主键';
COMMENT ON COLUMN public.users.email IS '用户邮箱（系统内唯一，登录凭证之一）';
COMMENT ON COLUMN public.users.password_hash IS '密码 bcrypt 哈希（OIDC/OAuth 登录用户可为空）';
COMMENT ON COLUMN public.users.display_name IS '用户显示名';
COMMENT ON COLUMN public.users.avatar_url IS '头像图片 URL';
COMMENT ON COLUMN public.users.is_active IS '账号是否激活: true=激活 / false=已禁用';
COMMENT ON COLUMN public.users.timezone IS '用户时区（默认 Asia/Shanghai）';
COMMENT ON COLUMN public.users.created_at IS '注册时间（含时区）';
COMMENT ON COLUMN public.users.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.users.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: workspaces — 工作空间/租户表
-- ============================================================
COMMENT ON TABLE public.workspaces IS '[租户与权限域]工作空间/租户（顶级隔离容器，所有业务表 workspace_id 指向此表）';

COMMENT ON COLUMN public.workspaces.id IS '自增主键';
COMMENT ON COLUMN public.workspaces.name IS '工作空间名称';
COMMENT ON COLUMN public.workspaces.slug IS '工作空间 URL 标识（全局唯一，非归档状态不可重复）';
COMMENT ON COLUMN public.workspaces.logo_url IS '工作空间 Logo URL';
COMMENT ON COLUMN public.workspaces.timezone IS '默认时区（默认 Asia/Shanghai）';
COMMENT ON COLUMN public.workspaces.language IS '默认语言（如 zh-CN, en-US）';
COMMENT ON COLUMN public.workspaces.owner_id IS '关联 users.id（空间所有者）';
COMMENT ON COLUMN public.workspaces.status IS '空间状态: active(活跃) / archived(已归档)';
COMMENT ON COLUMN public.workspaces.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.workspaces.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: invitations — 邀请表
-- ============================================================
COMMENT ON TABLE public.invitations IS '[租户与权限域]工作空间成员邀请（通过邮件发送邀请链接，7 天有效）';

COMMENT ON COLUMN public.invitations.id IS '自增主键';
COMMENT ON COLUMN public.invitations.workspace_id IS '关联 workspaces.id（被邀请的工作空间）';
COMMENT ON COLUMN public.invitations.inviter_id IS '关联 users.id（发送邀请的人）';
COMMENT ON COLUMN public.invitations.email IS '被邀请人邮箱';
COMMENT ON COLUMN public.invitations.role IS '邀请角色: admin(管理员) / member(成员) / guest(访客)';
COMMENT ON COLUMN public.invitations.token_hash IS '邀请令牌哈希（用于验证邀请链接）';
COMMENT ON COLUMN public.invitations.message IS '邀请附言';
COMMENT ON COLUMN public.invitations.status IS '邀请状态: pending(待接受) / accepted(已接受) / revoked(已撤销) / expired(已过期)';
COMMENT ON COLUMN public.invitations.expires_at IS '过期时间（含时区，默认创建后 7 天）';
COMMENT ON COLUMN public.invitations.accepted_at IS '接受时间（含时区）';
COMMENT ON COLUMN public.invitations.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.invitations.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: workspace_members — 工作空间成员表
-- ============================================================
COMMENT ON TABLE public.workspace_members IS '[租户与权限域]工作空间成员关系（用户在不同空间有不同角色）';

COMMENT ON COLUMN public.workspace_members.workspace_id IS '关联 workspaces.id（工作空间）';
COMMENT ON COLUMN public.workspace_members.user_id IS '关联 users.id（成员用户）';
COMMENT ON COLUMN public.workspace_members.role IS '成员角色: owner(所有者) / admin(管理员) / member(普通成员) / guest(访客)';
COMMENT ON COLUMN public.workspace_members.joined_at IS '加入时间（含时区）';



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
COMMENT ON COLUMN public.notifications.id IS '自增主键';
COMMENT ON COLUMN public.notifications.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.notifications.recipient_id IS '关联 users.id（通知接收人）';
COMMENT ON COLUMN public.notifications.event_type IS '事件类型: issue.created/issue.assigned/issue.status_changed/comment.created/sprint.started/sprint.completed/version.released/member.added';
COMMENT ON COLUMN public.notifications.entity_type IS '关联对象类型: issue/sprint/version/project/workspace/comment';
COMMENT ON COLUMN public.notifications.entity_id IS '关联对象的 ID';
COMMENT ON COLUMN public.notifications.title IS '通知标题';
COMMENT ON COLUMN public.notifications.body IS '通知正文（富文本）';
COMMENT ON COLUMN public.notifications.action_url IS '点击通知跳转的 URL';
COMMENT ON COLUMN public.notifications.actor_id IS '关联 users.id（触发该通知的操作人）';
COMMENT ON COLUMN public.notifications.actor_name IS '触发通知的操作人显示名（冗余存储）';
COMMENT ON COLUMN public.notifications.is_read IS '是否已读: true=已读 / false=未读';
COMMENT ON COLUMN public.notifications.is_archived IS '是否归档: true=已归档 / false=正常';
COMMENT ON COLUMN public.notifications.read_at IS '阅读时间（含时区）';
COMMENT ON COLUMN public.notifications.channel IS '通知渠道: in_app(站内)/email/sms/wecom/dingtalk/feishu';
COMMENT ON COLUMN public.notifications.payload IS '通知附加数据 JSON（包含跳转上下文等）';
COMMENT ON COLUMN public.notifications.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: notification_preferences — 通知偏好表（已有表注释，补充字段）
-- ============================================================
COMMENT ON COLUMN public.notification_preferences.id IS '自增主键';
COMMENT ON COLUMN public.notification_preferences.user_id IS '关联 users.id（用户）';
COMMENT ON COLUMN public.notification_preferences.workspace_id IS '关联 workspaces.id（工作空间）';
COMMENT ON COLUMN public.notification_preferences.event_types IS '订阅的事件类型列表 JSON 数组';
COMMENT ON COLUMN public.notification_preferences.channels IS '通知渠道配置 JSON 数组（如 ["in_app", "email"]）';
COMMENT ON COLUMN public.notification_preferences.digest IS '聚合发送频率: realtime(实时) / daily(每日) / weekly(每周)';
COMMENT ON COLUMN public.notification_preferences.dnd_enabled IS '是否启用免打扰: true=启用 / false=关闭';
COMMENT ON COLUMN public.notification_preferences.dnd_start IS '免打扰开始时间（默认 22:00）';
COMMENT ON COLUMN public.notification_preferences.dnd_end IS '免打扰结束时间（默认 08:00）';
COMMENT ON COLUMN public.notification_preferences.is_enabled IS '是否启用通知: true=启用 / false=关闭';
COMMENT ON COLUMN public.notification_preferences.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.notification_preferences.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: notification_deliveries — 通知投递记录表
-- ============================================================
COMMENT ON TABLE public.notification_deliveries IS '[通知域]通知多渠道投递记录（追踪每条通知的发送状态与重试）';

COMMENT ON COLUMN public.notification_deliveries.id IS '自增主键';
COMMENT ON COLUMN public.notification_deliveries.notification_id IS '关联 notifications.id（所属通知）';
COMMENT ON COLUMN public.notification_deliveries.channel IS '投递渠道: in_app / email / sms / wecom / dingtalk / feishu';
COMMENT ON COLUMN public.notification_deliveries.status IS '投递状态: pending(待发送) / sent(已发送) / failed(失败) / skipped(跳过)';
COMMENT ON COLUMN public.notification_deliveries.recipient IS '接收方标识（邮箱/手机号/用户ID等）';
COMMENT ON COLUMN public.notification_deliveries.sent_at IS '实际发送时间（含时区）';
COMMENT ON COLUMN public.notification_deliveries.error_msg IS '发送失败时的错误信息';
COMMENT ON COLUMN public.notification_deliveries.retry_count IS '已重试次数';
COMMENT ON COLUMN public.notification_deliveries.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.notification_deliveries.next_retry_at IS '下次重试时间（含时区）';

-- ============================================================
-- 表: notification_digests — 通知摘要/聚合
-- ============================================================
COMMENT ON TABLE public.notification_digests IS '[通知域]通知聚合摘要（将多条通知打包为每日/每周摘要定时发送）';

COMMENT ON COLUMN public.notification_digests.id IS '自增主键';
COMMENT ON COLUMN public.notification_digests.user_id IS '关联 users.id（摘要接收人）';
COMMENT ON COLUMN public.notification_digests.workspace_id IS '关联 workspaces.id（工作空间）';
COMMENT ON COLUMN public.notification_digests.digest_type IS '摘要类型: daily(每日) / weekly(每周)';
COMMENT ON COLUMN public.notification_digests.notification_ids IS '聚合的通知 ID 数组';
COMMENT ON COLUMN public.notification_digests.status IS '摘要状态: pending(待发送) / sent(已发送) / failed(失败)';
COMMENT ON COLUMN public.notification_digests.scheduled_for IS '计划发送时间（含时区）';
COMMENT ON COLUMN public.notification_digests.sent_at IS '实际发送时间（含时区）';
COMMENT ON COLUMN public.notification_digests.created_at IS '创建时间（含时区）';

-- ============================================================================
-- 入口工单域
-- ============================================================================

-- ============================================================
-- 表: intake_channels — 入口渠道表
-- ============================================================
COMMENT ON TABLE public.intake_channels IS '[入口工单域]工单接收渠道（公开提交入口，可配置默认类型、指派规则和自定义字段）';

COMMENT ON COLUMN public.intake_channels.id IS '自增主键';
COMMENT ON COLUMN public.intake_channels.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.intake_channels.project_id IS '关联 projects.id（工单转入的项目，可为空）';
COMMENT ON COLUMN public.intake_channels.slug IS '渠道 URL 标识（空间内唯一，is_active 时不可重复）';
COMMENT ON COLUMN public.intake_channels.name IS '渠道名称';
COMMENT ON COLUMN public.intake_channels.description IS '渠道说明';
COMMENT ON COLUMN public.intake_channels.is_public IS '是否公开访问: true=无需登录即可提交 / false=需要认证';
COMMENT ON COLUMN public.intake_channels.default_issue_type IS '工单默认类型: requirement / task / defect';
COMMENT ON COLUMN public.intake_channels.default_priority IS '默认优先级（0=无, 1=紧急, 2=高, 3=中, 4=低）';
COMMENT ON COLUMN public.intake_channels.auto_assign_rules IS '自动指派规则 JSON 数组';
COMMENT ON COLUMN public.intake_channels.rate_limit_per_min IS '每分钟提交限流数（默认 20）';
COMMENT ON COLUMN public.intake_channels.require_captcha IS '是否需要验证码: true=需要 / false=不需要';
COMMENT ON COLUMN public.intake_channels.custom_fields IS '自定义字段配置 JSON 数组';
COMMENT ON COLUMN public.intake_channels.branding IS '品牌配置 JSON（标题、Logo、说明文字）';
COMMENT ON COLUMN public.intake_channels.notify_on_submit IS '提交时是否通知管理员: true=通知 / false=不通知';
COMMENT ON COLUMN public.intake_channels.notify_users IS '提交时通知的管理员用户 ID 数组';
COMMENT ON COLUMN public.intake_channels.is_active IS '是否启用: true=启用 / false=禁用';
COMMENT ON COLUMN public.intake_channels.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.intake_channels.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.intake_channels.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: intake_issues — 入口工单表
-- ============================================================
COMMENT ON TABLE public.intake_issues IS '[入口工单域]通过入口渠道提交的工单（可转换为正式 issues 工作项）';

COMMENT ON COLUMN public.intake_issues.id IS '自增主键';
COMMENT ON COLUMN public.intake_issues.channel_id IS '关联 intake_channels.id（提交渠道）';
COMMENT ON COLUMN public.intake_issues.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.intake_issues.project_id IS '关联 projects.id（转入项目，可为空）';
COMMENT ON COLUMN public.intake_issues.tracking_id IS '工单追踪编号（空间内唯一，对外展示）';
COMMENT ON COLUMN public.intake_issues.submitter_name IS '提交者姓名';
COMMENT ON COLUMN public.intake_issues.submitter_email IS '提交者邮箱';
COMMENT ON COLUMN public.intake_issues.submitter_user_id IS '关联 users.id（如果提交者是系统用户）';
COMMENT ON COLUMN public.intake_issues.title IS '工单标题';
COMMENT ON COLUMN public.intake_issues.description IS '工单描述（纯文本）';
COMMENT ON COLUMN public.intake_issues.issue_type IS '工单类型: requirement / task / defect';
COMMENT ON COLUMN public.intake_issues.priority IS '优先级（0=无, 1=紧急, 2=高, 3=中, 4=低）';
COMMENT ON COLUMN public.intake_issues.custom_fields IS '自定义字段值 JSON';
COMMENT ON COLUMN public.intake_issues.attachment_ids IS '关联附件 ID 数组';
COMMENT ON COLUMN public.intake_issues.status IS '工单状态: open(待处理) / accepted(已接受) / rejected(已拒绝) / archived(已归档)';
COMMENT ON COLUMN public.intake_issues.status_reason IS '状态变更原因';
COMMENT ON COLUMN public.intake_issues.converted_issue_id IS '关联 issues.id（转换后的正式工作项 ID）';
COMMENT ON COLUMN public.intake_issues.assigned_to IS '关联 users.id（工单指定处理人）';
COMMENT ON COLUMN public.intake_issues.reviewed_by IS '关联 users.id（审核人）';
COMMENT ON COLUMN public.intake_issues.reviewed_at IS '审核时间（含时区）';
COMMENT ON COLUMN public.intake_issues.notify_on_status IS '状态变更时是否通知提交者: true=通知 / false=不通知';
COMMENT ON COLUMN public.intake_issues.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.intake_issues.updated_at IS '最后更新时间（含时区）';


-- ============================================================================
-- 自动化域
-- ============================================================================

-- ============================================================
-- 表: automation_rules — 自动化规则表
-- ============================================================
COMMENT ON TABLE public.automation_rules IS '[自动化域]工作项自动化规则（DSL 定义 trigger+condition+action，支持多类型触发器）';

COMMENT ON COLUMN public.automation_rules.id IS '自增主键';
COMMENT ON COLUMN public.automation_rules.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.automation_rules.project_id IS '关联 projects.id（项目级规则为空表示空间级）';
COMMENT ON COLUMN public.automation_rules.name IS '规则名称';
COMMENT ON COLUMN public.automation_rules.description IS '规则描述';
COMMENT ON COLUMN public.automation_rules.dsl IS '规则 DSL JSON（trigger / conditions / actions 三段式结构）';
COMMENT ON COLUMN public.automation_rules.trigger_type IS '触发器类型: issue.created / issue.updated / issue.status_changed / version.released / scheduled 等';
COMMENT ON COLUMN public.automation_rules.action_count IS '规则包含的动作数量';
COMMENT ON COLUMN public.automation_rules.status IS '规则状态: draft(草稿) / active(启用) / disabled(禁用) / error(执行出错)';
COMMENT ON COLUMN public.automation_rules.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.automation_rules.last_run_at IS '最近一次执行时间（含时区）';
COMMENT ON COLUMN public.automation_rules.last_error IS '最近一次错误信息';
COMMENT ON COLUMN public.automation_rules.consecutive_failures IS '连续失败次数（达到阈值自动置为 error）';
COMMENT ON COLUMN public.automation_rules.execution_count IS '累计执行次数';
COMMENT ON COLUMN public.automation_rules.sort_order IS '排序权重（越小越靠前）';
COMMENT ON COLUMN public.automation_rules.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.automation_rules.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: automation_templates — 自动化模板表
-- ============================================================
COMMENT ON TABLE public.automation_templates IS '[自动化域]预置自动化规则模板（用户可一键启用，快速配置常见自动化场景）';

COMMENT ON COLUMN public.automation_templates.id IS '自增主键';
COMMENT ON COLUMN public.automation_templates.name IS '模板名称';
COMMENT ON COLUMN public.automation_templates.slug IS '模板标识符（系统内唯一）';
COMMENT ON COLUMN public.automation_templates.description IS '模板说明';
COMMENT ON COLUMN public.automation_templates.category IS '模板分类: efficiency(效率) / quality(质量) / notification(通知)';
COMMENT ON COLUMN public.automation_templates.dsl_template IS '模板 DSL JSON（预置的 trigger+condition+action 结构）';
COMMENT ON COLUMN public.automation_templates.icon IS '模板图标名（Lucide）';
COMMENT ON COLUMN public.automation_templates.sort_order IS '排序权重';
COMMENT ON COLUMN public.automation_templates.is_recommended IS '是否推荐: true=在首页推荐展示 / false=普通';
COMMENT ON COLUMN public.automation_templates.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: rule_executions — 规则执行日志表
-- ============================================================
COMMENT ON TABLE public.rule_executions IS '[自动化域]自动化规则执行日志（记录每次触发的匹配、执行和耗时信息）';

COMMENT ON COLUMN public.rule_executions.id IS '自增主键';
COMMENT ON COLUMN public.rule_executions.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.rule_executions.project_id IS '关联 projects.id（项目级规则执行）';
COMMENT ON COLUMN public.rule_executions.rule_id IS '关联 automation_rules.id（执行规则）';
COMMENT ON COLUMN public.rule_executions.trigger_event_id IS '关联 domain_events.id（触发事件 ID）';
COMMENT ON COLUMN public.rule_executions.status IS '执行状态: matched(已匹配但未执行) / skipped(跳过条件不满足) / success(成功) / failed(失败) / dry_run(试运行)';
COMMENT ON COLUMN public.rule_executions.duration_ms IS '执行耗时（毫秒）';
COMMENT ON COLUMN public.rule_executions.error_message IS '执行失败时的错误信息';
COMMENT ON COLUMN public.rule_executions.context_json IS '执行上下文 JSON（当时的 issue 快照等）';
COMMENT ON COLUMN public.rule_executions.trigger_depth IS '触发深度（防止递归触发，0=直接触发）';
COMMENT ON COLUMN public.rule_executions.via_automation IS '是否由其他自动化规则间接触发';
COMMENT ON COLUMN public.rule_executions.created_at IS '创建时间（含时区）';

-- ============================================================================
-- 仪表盘域
-- ============================================================================

-- ============================================================
-- 表: dashboard_widgets — 仪表盘组件表
-- ============================================================
COMMENT ON TABLE public.dashboard_widgets IS '[仪表盘域]仪表盘组件实例（项目级可配置的图表/列表/数据卡组件，支持布局定位）';

COMMENT ON COLUMN public.dashboard_widgets.id IS '自增主键';
COMMENT ON COLUMN public.dashboard_widgets.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.dashboard_widgets.widget_type IS '组件类型: progress_overview(进度概览) / burndown(燃尽图) / velocity(速率图) / priority_split(优先级分布) / state_distribution(状态分布) / overdue_list(逾期列表) / blocked_list(阻塞列表) / risk_alert(风险告警) / recent_activity(近期活动) / team_workload(团队负载)';
COMMENT ON COLUMN public.dashboard_widgets.title IS '组件标题';
COMMENT ON COLUMN public.dashboard_widgets.grid_x IS '网格 X 坐标（列位置）';
COMMENT ON COLUMN public.dashboard_widgets.grid_y IS '网格 Y 坐标（行位置）';
COMMENT ON COLUMN public.dashboard_widgets.grid_w IS '网格宽度（占几列，默认 4）';
COMMENT ON COLUMN public.dashboard_widgets.grid_h IS '网格高度（占几行，默认 3）';
COMMENT ON COLUMN public.dashboard_widgets.config IS '组件配置 JSON（数据源、过滤条件、展示参数）';
COMMENT ON COLUMN public.dashboard_widgets.is_visible IS '是否可见: true=显示 / false=隐藏';
COMMENT ON COLUMN public.dashboard_widgets.sort_order IS '排序权重';
COMMENT ON COLUMN public.dashboard_widgets.user_id IS '关联 users.id（NULL 表示项目级组件，非空表示个人组件）';
COMMENT ON COLUMN public.dashboard_widgets.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.dashboard_widgets.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: dashboard_templates — 仪表盘模板表
-- ============================================================
COMMENT ON TABLE public.dashboard_templates IS '[仪表盘域]预置仪表盘布局模板（一键初始化项目仪表盘，如项目概览、质量看板等）';

COMMENT ON COLUMN public.dashboard_templates.id IS '自增主键';
COMMENT ON COLUMN public.dashboard_templates.name IS '模板名称';
COMMENT ON COLUMN public.dashboard_templates.slug IS '模板标识符（系统内唯一）';
COMMENT ON COLUMN public.dashboard_templates.description IS '模板说明';
COMMENT ON COLUMN public.dashboard_templates.layout IS '布局配置 JSON（widgets 数组的默认位置与尺寸）';
COMMENT ON COLUMN public.dashboard_templates.icon IS '模板图标名';
COMMENT ON COLUMN public.dashboard_templates.category IS '模板分类: agile(敏捷) / pmo(项目管理) / quality(质量)';
COMMENT ON COLUMN public.dashboard_templates.is_default IS '是否为默认模板: true=创建项目时自动应用';
COMMENT ON COLUMN public.dashboard_templates.sort_order IS '排序权重';
COMMENT ON COLUMN public.dashboard_templates.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: dashboard_snapshots — 仪表盘快照表
-- ============================================================
COMMENT ON TABLE public.dashboard_snapshots IS '[仪表盘域]仪表盘组件数据快照（定时刷新缓存，避免实时查询开销）';

COMMENT ON COLUMN public.dashboard_snapshots.id IS '自增主键';
COMMENT ON COLUMN public.dashboard_snapshots.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.dashboard_snapshots.widget_type IS '组件类型（与 dashboard_widgets.widget_type 枚举值相同）';
COMMENT ON COLUMN public.dashboard_snapshots.data IS '快照数据 JSON（组件渲染所需的全部数据）';
COMMENT ON COLUMN public.dashboard_snapshots.refreshed_at IS '最后刷新时间（含时区）';

-- ============================================================================
-- 功能区
-- ============================================================================

-- ============================================================
-- 表: workbench_configs — 工作台配置表
-- ============================================================
COMMENT ON TABLE public.workbench_configs IS '[功能区]用户工作台个性化布局配置（每人/每项目可有独立工作台布局）';

COMMENT ON COLUMN public.workbench_configs.id IS '自增主键';
COMMENT ON COLUMN public.workbench_configs.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.workbench_configs.project_id IS '关联 projects.id（NULL 表示空间级工作台）';
COMMENT ON COLUMN public.workbench_configs.user_id IS '关联 users.id（配置所属用户）';
COMMENT ON COLUMN public.workbench_configs.layout IS '工作台布局 JSON（组件排列、尺寸等）';
COMMENT ON COLUMN public.workbench_configs.widget_states IS '组件状态 JSON（折叠/展开/过滤条件等）';
COMMENT ON COLUMN public.workbench_configs.focus_enabled IS '是否启用专注模式: true=仅显示待办和专注计时器';
COMMENT ON COLUMN public.workbench_configs.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: workbench_templates — 工作台模板表
-- ============================================================
COMMENT ON TABLE public.workbench_templates IS '[功能区]预置工作台布局模板（如敏捷开发、项目监控、个人专注等模式一键切换）';

COMMENT ON COLUMN public.workbench_templates.id IS '自增主键';
COMMENT ON COLUMN public.workbench_templates.name IS '模板名称';
COMMENT ON COLUMN public.workbench_templates.slug IS '模板标识符（系统内唯一）';
COMMENT ON COLUMN public.workbench_templates.description IS '模板说明';
COMMENT ON COLUMN public.workbench_templates.layout IS '布局 JSON（预置组件排列方案）';
COMMENT ON COLUMN public.workbench_templates.icon IS '模板图标名';
COMMENT ON COLUMN public.workbench_templates.is_default IS '是否为默认模板: true=新用户默认使用';
COMMENT ON COLUMN public.workbench_templates.sort_order IS '排序权重';
COMMENT ON COLUMN public.workbench_templates.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: view_preferences — 视图偏好表
-- ============================================================
COMMENT ON TABLE public.view_preferences IS '[功能区]视图展示偏好（如列表/看板的列定义、筛选条件、排序方式）';

COMMENT ON COLUMN public.view_preferences.id IS '自增主键';
COMMENT ON COLUMN public.view_preferences.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.view_preferences.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.view_preferences.user_id IS '关联 users.id（偏好所属用户）';
COMMENT ON COLUMN public.view_preferences.view_type IS '视图类型: list(列表) / kanban(看板) / calendar(日历) / spreadsheet(表格) / gantt(甘特图)';
COMMENT ON COLUMN public.view_preferences.layout IS '布局模式: list / kanban / calendar / spreadsheet / gantt 等';
COMMENT ON COLUMN public.view_preferences.columns IS '已配置列的 JSON 数组（列宽、可见性、顺序）';
COMMENT ON COLUMN public.view_preferences.filters IS '视图过滤条件 JSON';
COMMENT ON COLUMN public.view_preferences.sort IS '排序规则 JSON（字段、方向）';
COMMENT ON COLUMN public.view_preferences.extra IS '额外视图参数 JSON（视图特有配置）';
COMMENT ON COLUMN public.view_preferences.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.view_preferences.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: recent_items — 最近访问记录表
-- ============================================================
COMMENT ON TABLE public.recent_items IS '[功能区]用户最近访问记录（工作台"最近访问"功能的数据来源）';

COMMENT ON COLUMN public.recent_items.id IS '自增主键';
COMMENT ON COLUMN public.recent_items.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.recent_items.user_id IS '关联 users.id（记录所属用户）';
COMMENT ON COLUMN public.recent_items.item_type IS '访问对象类型: project / issue / sprint / version';
COMMENT ON COLUMN public.recent_items.item_id IS '访问对象的 ID';
COMMENT ON COLUMN public.recent_items.project_id IS '关联 projects.id（便于按项目筛选）';
COMMENT ON COLUMN public.recent_items.title IS '对象标题（冗余存储，避免关联查询）';
COMMENT ON COLUMN public.recent_items.identifier IS '对象标识符（如 YD-123，冗余存储）';
COMMENT ON COLUMN public.recent_items.accessed_at IS '访问时间（含时区，触发器自动更新）';

-- ============================================================
-- 表: search_bookmarks — 搜索书签表
-- ============================================================
COMMENT ON TABLE public.search_bookmarks IS '[功能区]用户保存的搜索书签（常用搜索条件的快捷入口，可共享给团队）';

COMMENT ON COLUMN public.search_bookmarks.id IS '自增主键';
COMMENT ON COLUMN public.search_bookmarks.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.search_bookmarks.project_id IS '关联 projects.id（项目级书签，可为空）';
COMMENT ON COLUMN public.search_bookmarks.user_id IS '关联 users.id（书签所属用户）';
COMMENT ON COLUMN public.search_bookmarks.name IS '书签名称';
COMMENT ON COLUMN public.search_bookmarks.query IS '搜索关键词';
COMMENT ON COLUMN public.search_bookmarks.filters IS '搜索过滤条件 JSON';
COMMENT ON COLUMN public.search_bookmarks.is_shared IS '是否共享: true=项目成员可见 / false=仅个人可见';
COMMENT ON COLUMN public.search_bookmarks.sort_order IS '排序权重';
COMMENT ON COLUMN public.search_bookmarks.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.search_bookmarks.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.search_bookmarks.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: search_documents — 搜索文档索引表
-- ============================================================
COMMENT ON TABLE public.search_documents IS '[功能区]全文检索文档索引（触发器自动从 issues/sprints/versions 同步，供 PG 全文检索降级使用）';

COMMENT ON COLUMN public.search_documents.id IS '自增主键';
COMMENT ON COLUMN public.search_documents.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.search_documents.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.search_documents.doc_type IS '文档类型: issue / sprint / version';
COMMENT ON COLUMN public.search_documents.doc_id IS '原始表中的 ID（与 doc_type 组成唯一索引）';
COMMENT ON COLUMN public.search_documents.title IS '文档标题（工作项名/迭代名/版本名）';
COMMENT ON COLUMN public.search_documents.identifier IS '文档标识符（如 YD-123 或版本号）';
COMMENT ON COLUMN public.search_documents.content IS '文档正文内容（纯文本，用于全文检索）';
COMMENT ON COLUMN public.search_documents.search_tsv IS '全文检索 tsvector（simple 配置，由触发器维护）';
COMMENT ON COLUMN public.search_documents.metadata IS '元数据 JSON（type_code、state_id、priority 等可筛选字段）';
COMMENT ON COLUMN public.search_documents.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: search_history — 搜索历史表
-- ============================================================
COMMENT ON TABLE public.search_history IS '[功能区]用户搜索历史记录（用于搜索建议和最近搜索展示）';

COMMENT ON COLUMN public.search_history.id IS '自增主键';
COMMENT ON COLUMN public.search_history.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.search_history.user_id IS '关联 users.id（搜索用户）';
COMMENT ON COLUMN public.search_history.query IS '搜索关键词原文';
COMMENT ON COLUMN public.search_history.filters IS '搜索当时使用的过滤条件 JSON';
COMMENT ON COLUMN public.search_history.result_count IS '搜索结果数量';
COMMENT ON COLUMN public.search_history.searched_at IS '搜索时间（含时区）';


-- ============================================================================
-- 风险与度量域
-- ============================================================================

-- ============================================================
-- 表: risk_rules — 风险规则表
-- ============================================================
COMMENT ON TABLE public.risk_rules IS '[风险与度量域]风险检测规则（预定义阈值条件，触发时自动生成 risk_alerts 告警）';

COMMENT ON COLUMN public.risk_rules.id IS '自增主键';
COMMENT ON COLUMN public.risk_rules.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.risk_rules.project_id IS '关联 projects.id（项目级规则，为空表示空间级）';
COMMENT ON COLUMN public.risk_rules.rule_name IS '规则名称';
COMMENT ON COLUMN public.risk_rules.rule_type IS '规则类型: overdue_issue(逾期工作项) / overdue_sprint(逾期迭代) / blocked_count(阻塞数超标) / sla_breach(SLA违约) / stalled_progress(进度停滞) / high_priority_open(高优未关闭)';
COMMENT ON COLUMN public.risk_rules.condition_json IS '规则条件配置 JSON（阈值、比较运算符、检测频率等）';
COMMENT ON COLUMN public.risk_rules.notify_channels IS '告警通知渠道数组（in_app / email / webhook）';
COMMENT ON COLUMN public.risk_rules.is_active IS '规则是否启用: true=启用监控 / false=暂停';
COMMENT ON COLUMN public.risk_rules.last_triggered IS '最近一次触发时间（含时区）';
COMMENT ON COLUMN public.risk_rules.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.risk_rules.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: risk_alerts — 风险告警表
-- ============================================================
COMMENT ON TABLE public.risk_alerts IS '[风险与度量域]风险告警记录（由 risk_rules 自动生成，需要人工确认和解决）';

COMMENT ON COLUMN public.risk_alerts.id IS '自增主键';
COMMENT ON COLUMN public.risk_alerts.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.risk_alerts.project_id IS '关联 projects.id（告警所属项目）';
COMMENT ON COLUMN public.risk_alerts.rule_id IS '关联 risk_rules.id（触发规则）';
COMMENT ON COLUMN public.risk_alerts.severity IS '告警严重度: info(信息) / low(低) / medium(中) / high(高) / critical(严重)';
COMMENT ON COLUMN public.risk_alerts.title IS '告警标题';
COMMENT ON COLUMN public.risk_alerts.description IS '告警描述（详细说明了触发原因）';
COMMENT ON COLUMN public.risk_alerts.metadata IS '告警元数据 JSON（触发时的上下文快照）';
COMMENT ON COLUMN public.risk_alerts.is_resolved IS '是否已解决: true=已解决 / false=未解决';
COMMENT ON COLUMN public.risk_alerts.resolved_at IS '解决时间（含时区）';
COMMENT ON COLUMN public.risk_alerts.resolved_by IS '关联 users.id（解决人）';
COMMENT ON COLUMN public.risk_alerts.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: metric_snapshots — 指标快照表
-- ============================================================
COMMENT ON TABLE public.metric_snapshots IS '[风险与度量域]项目/空间级效能指标每日快照（支撑趋势分析、燃尽图、速率图）';

COMMENT ON COLUMN public.metric_snapshots.id IS '自增主键';
COMMENT ON COLUMN public.metric_snapshots.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.metric_snapshots.project_id IS '关联 projects.id（项目级指标，为空表示空间级）';
COMMENT ON COLUMN public.metric_snapshots.granularity IS '粒度: daily(日级) / sprint(迭代级) / version(版本级)';
COMMENT ON COLUMN public.metric_snapshots.ref_id IS '关联对象 ID（sprint 粒度时为 sprints.id，version 粒度时为 versions.id）';
COMMENT ON COLUMN public.metric_snapshots.metric IS '指标名称（如 velocity、burndown_points、bug_count）';
COMMENT ON COLUMN public.metric_snapshots.value IS '指标值（NUMERIC(12,4)）';
COMMENT ON COLUMN public.metric_snapshots.dimensions IS '维度标签 JSON（按工作项类型、优先级等切分的子指标）';
COMMENT ON COLUMN public.metric_snapshots.snapshot_date IS '快照日期';
COMMENT ON COLUMN public.metric_snapshots.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: metric_adjustments — 指标调整记录表
-- ============================================================
COMMENT ON TABLE public.metric_adjustments IS '[风险与度量域]指标人工修正记录（对异常指标进行手动调整的审计日志）';

COMMENT ON COLUMN public.metric_adjustments.id IS '自增主键';
COMMENT ON COLUMN public.metric_adjustments.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.metric_adjustments.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.metric_adjustments.snapshot_id IS '关联 metric_snapshots.id（被修正的快照）';
COMMENT ON COLUMN public.metric_adjustments.metric IS '被修正的指标名';
COMMENT ON COLUMN public.metric_adjustments.snapshot_date IS '快照日期';
COMMENT ON COLUMN public.metric_adjustments.original_value IS '原始指标值';
COMMENT ON COLUMN public.metric_adjustments.adjusted_value IS '修正后的指标值';
COMMENT ON COLUMN public.metric_adjustments.reason IS '修正原因（必填）';
COMMENT ON COLUMN public.metric_adjustments.adjusted_by IS '关联 users.id（修正人）';
COMMENT ON COLUMN public.metric_adjustments.created_at IS '创建时间（含时区）';

-- ============================================================================
-- 集成与扩展域
-- ============================================================================

-- ============================================================
-- 表: webhooks — Webhook 配置表
-- ============================================================
COMMENT ON TABLE public.webhooks IS '[集成与扩展域]Webhook 配置（项目级 HTTP 回调，事件触发时推送 JSON 报文）';

COMMENT ON COLUMN public.webhooks.id IS '自增主键';
COMMENT ON COLUMN public.webhooks.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.webhooks.project_id IS '关联 projects.id（项目级 webhook，可为空表示空间级）';
COMMENT ON COLUMN public.webhooks.name IS 'Webhook 名称';
COMMENT ON COLUMN public.webhooks.target_url IS '回调目标 URL（必须 http:// 或 https:// 开头）';
COMMENT ON COLUMN public.webhooks.secret IS '签名密钥（用于 HMAC 签名验证）';
COMMENT ON COLUMN public.webhooks.events IS '订阅事件类型数组（如 issue.created, issue.updated）';
COMMENT ON COLUMN public.webhooks.is_active IS '是否启用: true=启用 / false=禁用';
COMMENT ON COLUMN public.webhooks.last_error IS '最近一次发送失败的错误信息';
COMMENT ON COLUMN public.webhooks.last_triggered IS '最近一次触发时间（含时区）';
COMMENT ON COLUMN public.webhooks.last_status IS '最近一次执行状态: success / failed';
COMMENT ON COLUMN public.webhooks.unhealthy_at IS '判定为不健康的时间（含时区连续失败超过阈值）';
COMMENT ON COLUMN public.webhooks.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.webhooks.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.webhooks.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: webhook_logs — Webhook 日志表
-- ============================================================
COMMENT ON TABLE public.webhook_logs IS '[集成与扩展域]Webhook 投递日志（每次回调尝试的完整请求/响应记录，支持重放排错）';

COMMENT ON COLUMN public.webhook_logs.id IS '自增主键';
COMMENT ON COLUMN public.webhook_logs.webhook_id IS '关联 webhooks.id（所属 webhook）';
COMMENT ON COLUMN public.webhook_logs.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.webhook_logs.delivery_id IS '投递 ID（用于去重和追踪）';
COMMENT ON COLUMN public.webhook_logs.event_type IS '触发事件类型';
COMMENT ON COLUMN public.webhook_logs.event_id IS '关联 domain_events.id（触发事件 ID）';
COMMENT ON COLUMN public.webhook_logs.request_url IS '请求目标 URL';
COMMENT ON COLUMN public.webhook_logs.request_method IS 'HTTP 方法（默认 POST）';
COMMENT ON COLUMN public.webhook_logs.request_headers IS '请求头 JSON';
COMMENT ON COLUMN public.webhook_logs.request_body IS '请求体（JSON 序列化的事件数据）';
COMMENT ON COLUMN public.webhook_logs.response_status IS 'HTTP 响应状态码';
COMMENT ON COLUMN public.webhook_logs.response_body IS 'HTTP 响应体';
COMMENT ON COLUMN public.webhook_logs.response_headers IS '响应头 JSON';
COMMENT ON COLUMN public.webhook_logs.status IS '投递状态: success / failed / pending / retrying';
COMMENT ON COLUMN public.webhook_logs.attempt IS '尝试次数';
COMMENT ON COLUMN public.webhook_logs.duration_ms IS '请求耗时（毫秒）';
COMMENT ON COLUMN public.webhook_logs.error IS '错误信息';
COMMENT ON COLUMN public.webhook_logs.occurred_at IS '发生时间（含时区）';

-- ============================================================
-- 表: api_tokens — API 令牌表
-- ============================================================
COMMENT ON TABLE public.api_tokens IS '[集成与扩展域]用户 API 令牌（用于编程访问 API，支持 scope 权限控制与撤销）';

COMMENT ON COLUMN public.api_tokens.id IS '自增主键';
COMMENT ON COLUMN public.api_tokens.user_id IS '关联 users.id（令牌所属用户）';
COMMENT ON COLUMN public.api_tokens.name IS '令牌名称（仅展示用途）';
COMMENT ON COLUMN public.api_tokens.token_hash IS '令牌哈希（存储 hash，原始值仅创建时返回一次）';
COMMENT ON COLUMN public.api_tokens.scopes IS '权限范围 JSON 数组（如 ["read:workspace", "write:issues"]）';
COMMENT ON COLUMN public.api_tokens.last_used_at IS '最近一次使用时间（含时区）';
COMMENT ON COLUMN public.api_tokens.expires_at IS '过期时间（含时区）';
COMMENT ON COLUMN public.api_tokens.revoked_at IS '撤销时间（含时区），NULL 表示未撤销';
COMMENT ON COLUMN public.api_tokens.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.api_tokens.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: deployment_events — 部署事件表
-- ============================================================
COMMENT ON TABLE public.deployment_events IS '[集成与扩展域]部署流水线事件（接收 CI/CD 回调，追踪代码部署状态与关联）';

COMMENT ON COLUMN public.deployment_events.id IS '自增主键';
COMMENT ON COLUMN public.deployment_events.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.deployment_events.project_id IS '关联 projects.id（部署项目）';
COMMENT ON COLUMN public.deployment_events.deployment_id IS '部署 ID（用于幂等去重）';
COMMENT ON COLUMN public.deployment_events.env IS '部署环境: development / staging / production / testing';
COMMENT ON COLUMN public.deployment_events.status IS '部署状态: success(成功) / failed(失败) / rolled_back(已回滚)';
COMMENT ON COLUMN public.deployment_events.commit_sha IS '部署的 Git commit SHA';
COMMENT ON COLUMN public.deployment_events.started_at IS '部署开始时间（含时区）';
COMMENT ON COLUMN public.deployment_events.deployed_at IS '部署完成时间（含时区）';
COMMENT ON COLUMN public.deployment_events.source IS '事件来源: webhook(回调) / api(接口) / cli(命令行)';
COMMENT ON COLUMN public.deployment_events.metadata IS '部署元数据 JSON（流水线名称、执行人等）';
COMMENT ON COLUMN public.deployment_events.created_at IS '创建时间（含时区）';

-- ============================================================================
-- 系统基础设施域
-- ============================================================================

-- ============================================================
-- 表: domain_events — 领域事件表（Outbox 模式）
-- ============================================================
COMMENT ON TABLE public.domain_events IS '[系统基础设施域]领域事件表（Transactional Outbox 模式，保证业务操作与事件发布的原子性）';

COMMENT ON COLUMN public.domain_events.id IS '自增主键';
COMMENT ON COLUMN public.domain_events.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.domain_events.aggregate_type IS '聚合根类型（如 issue、sprint、version）';
COMMENT ON COLUMN public.domain_events.aggregate_id IS '聚合根 ID';
COMMENT ON COLUMN public.domain_events.event_type IS '事件类型（如 issue.status_changed、issue.created）';
COMMENT ON COLUMN public.domain_events.payload IS '事件数据 JSON（事件详情）';
COMMENT ON COLUMN public.domain_events.occurred_at IS '事件发生时间（含时区）';
COMMENT ON COLUMN public.domain_events.published_at IS '事件发布时间（含时区），NULL 表示未发布';

-- ============================================================
-- 表: idempotency_keys — API 幂等键表
-- ============================================================
COMMENT ON TABLE public.idempotency_keys IS '[系统基础设施域]API 幂等键（防止重复提交，存储首次响应用于重放）';

COMMENT ON COLUMN public.idempotency_keys.key IS '幂等键（UUID，客户端生成）';
COMMENT ON COLUMN public.idempotency_keys.user_id IS '关联 users.id（请求用户）';
COMMENT ON COLUMN public.idempotency_keys.response IS '首次响应 JSON（用于重复请求时直接返回）';
COMMENT ON COLUMN public.idempotency_keys.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: password_reset_tokens — 密码重置令牌表
-- ============================================================
COMMENT ON TABLE public.password_reset_tokens IS '[系统基础设施域]密码重置令牌（一次性使用，过期后失效）';

COMMENT ON COLUMN public.password_reset_tokens.id IS '自增主键';
COMMENT ON COLUMN public.password_reset_tokens.user_id IS '关联 users.id（申请重置的用户）';
COMMENT ON COLUMN public.password_reset_tokens.token_hash IS '令牌哈希（通过邮件发送原始值）';
COMMENT ON COLUMN public.password_reset_tokens.expires_at IS '过期时间（含时区）';
COMMENT ON COLUMN public.password_reset_tokens.used_at IS '使用时间（含时区），NULL 表示未使用';
COMMENT ON COLUMN public.password_reset_tokens.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: audit_logs — 审计日志表
-- ============================================================
COMMENT ON TABLE public.audit_logs IS '[系统基础设施域]操作审计日志（记录所有实体变更和关键操作，供合规审查）';

COMMENT ON COLUMN public.audit_logs.id IS '自增主键';
COMMENT ON COLUMN public.audit_logs.workspace_id IS '关联 workspaces.id（工作空间上下文）';
COMMENT ON COLUMN public.audit_logs.actor_id IS '关联 users.id（操作执行人，NULL 表示系统操作）';
COMMENT ON COLUMN public.audit_logs.action IS '操作类型（如 issue.status_changed、version.released）';
COMMENT ON COLUMN public.audit_logs.target IS '操作目标标识（如工作项编号、版本号）';
COMMENT ON COLUMN public.audit_logs.detail IS '操作详情 JSON（变更前后值等）';
COMMENT ON COLUMN public.audit_logs.ip IS '操作来源 IP 地址（inet 类型）';
COMMENT ON COLUMN public.audit_logs.created_at IS '操作时间（含时区）';

-- ============================================================
-- 表: schema_migrations — 数据库迁移版本表
-- ============================================================
COMMENT ON TABLE public.schema_migrations IS '[系统基础设施域]数据库迁移版本（记录已应用的迁移编号和脏标记）';

COMMENT ON COLUMN public.schema_migrations.version IS '迁移版本号（对应迁移文件名中的数字前缀）';
COMMENT ON COLUMN public.schema_migrations.dirty IS '是否处于脏状态: true=上次迁移中途失败 / false=正常';

-- ============================================================
-- 表: project_sequences — 项目序列发号器表
-- ============================================================
COMMENT ON TABLE public.project_sequences IS '[系统基础设施域]项目工作项序列发号器（每个项目一行，原子递增生成 sequence_id）';

COMMENT ON COLUMN public.project_sequences.project_id IS '关联 projects.id（所属项目，主键）';
COMMENT ON COLUMN public.project_sequences.next_value IS '下一个序列号（从 1 开始，允许跳号）';



-- ============================================================================
-- 触发器注释（统一区块）
-- ============================================================================

-- ============================================================
-- updated_at 自动维护触发器（适用于所有含 updated_at 的表）
-- ============================================================
COMMENT ON TRIGGER trg_api_tokens_updated_at ON public.api_tokens IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_automation_rules_updated_at ON public.automation_rules IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_dashboard_widgets_updated_at ON public.dashboard_widgets IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_invitations_updated_at ON public.invitations IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_issues_updated_at ON public.issues IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_labels_updated_at ON public.labels IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_metric_adjustments_updated_at ON public.metric_adjustments IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_modules_updated_at ON public.modules IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_projects_updated_at ON public.projects IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_risk_rules_updated_at ON public.risk_rules IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_search_bookmarks_updated_at ON public.search_bookmarks IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_search_documents_updated_at ON public.search_documents IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_sprints_updated_at ON public.sprints IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_states_updated_at ON public.states IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_time_logs_updated_at ON public.time_logs IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_users_updated_at ON public.users IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_versions_updated_at ON public.versions IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_workbench_configs_updated_at ON public.workbench_configs IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_workspaces_updated_at ON public.workspaces IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';

-- ============================================================
-- 乐观锁版本号自增触发器
-- ============================================================
COMMENT ON TRIGGER trg_versions_bump_version ON public.versions IS '乐观锁版本号自增: NEW.version = OLD.version + 1，BEFORE UPDATE 时触发';

-- ============================================================
-- 搜索索引同步触发器
-- ============================================================
COMMENT ON TRIGGER trg_issue_search_sync ON public.issues IS '工作项插入/更新时同步到 search_documents 全文检索表（仅 deleted_at IS NULL 时）';
COMMENT ON TRIGGER trg_issue_search_cleanup ON public.issues IS '工作项软删除（deleted_at 变为非空）时清理 search_documents 对应记录';
COMMENT ON TRIGGER trg_sprint_search_sync ON public.sprints IS '迭代插入/更新时同步到 search_documents 全文检索表（仅 deleted_at IS NULL 时）';
COMMENT ON TRIGGER trg_sprint_search_cleanup ON public.sprints IS '迭代软删除（deleted_at 变为非空）时清理 search_documents 对应记录';
COMMENT ON TRIGGER trg_version_search_sync ON public.versions IS '版本插入/更新时同步到 search_documents 全文检索表（仅 deleted_at IS NULL 时）';
COMMENT ON TRIGGER trg_version_search_cleanup ON public.versions IS '版本软删除（deleted_at 变为非空）时清理 search_documents 对应记录';

-- ============================================================
-- 最近访问时间更新触发器
-- ============================================================
COMMENT ON TRIGGER trg_recent_items_touch ON public.recent_items IS '每次 UPDATE 时自动将 accessed_at 重置为 now()';

-- ============================================================================
-- 部分索引注释（说明 WHERE 条件含义）
-- ============================================================================

-- ================ issues 表索引 ================
COMMENT ON INDEX public.idx_issues_project_state IS '按项目+状态+排序查询工作项列表（WHERE deleted_at IS NULL，看板视图主索引）';
COMMENT ON INDEX public.idx_issues_list_covering IS '工作项列表排序查询（WHERE deleted_at IS NULL，按 updated_at 倒序）';
COMMENT ON INDEX public.idx_issues_project_sequence IS '工作项编号唯一性保证（WHERE deleted_at IS NULL，sequence_id 在项目内唯一）';
COMMENT ON INDEX public.idx_issues_public_id IS '按 public_id 查询工作项（WHERE deleted_at IS NULL，UUID 部分唯一索引）';
COMMENT ON INDEX public.idx_issues_parent IS '查找子工作项（WHERE deleted_at IS NULL AND parent_id IS NOT NULL，层级关系查询）';
COMMENT ON INDEX public.idx_issues_target_date IS '按目标日期查询未完成工作项（WHERE deleted_at IS NULL AND completed_at IS NULL）';
COMMENT ON INDEX public.idx_issues_target_date_covering IS '目标日期覆盖索引（WHERE deleted_at IS NULL AND target_date IS NOT NULL）';
COMMENT ON INDEX public.idx_issues_search_tsv IS '工作项全文检索 GIN 索引（基于 search_tsv 字段，ES 降级使用）';
COMMENT ON INDEX public.idx_issues_priority_covering IS '高优工作项覆盖索引（WHERE deleted_at IS NULL AND priority IN (urgent, high)，优先级筛选视图）';
COMMENT ON INDEX public.idx_issues_fix_version IS '按修复版本查询工作项（WHERE deleted_at IS NULL AND fix_version_id IS NOT NULL）';
COMMENT ON INDEX public.idx_issues_found_version IS '按发现版本查询缺陷（WHERE deleted_at IS NULL AND found_version_id IS NOT NULL）';
COMMENT ON INDEX public.idx_issues_release_version IS '按发布版本查询工作项（WHERE deleted_at IS NULL AND release_version_id IS NOT NULL）';
COMMENT ON INDEX public.idx_issues_type IS '按类型查询工作项（WHERE deleted_at IS NULL，类型筛选）';
COMMENT ON INDEX public.idx_issues_type_covering IS '按类型查询并排序（WHERE deleted_at IS NULL，类型视图）';
COMMENT ON INDEX public.idx_issues_workspace_project IS '按空间+项目查询工作项（WHERE deleted_at IS NULL，列表视图）';
COMMENT ON INDEX public.idx_issues_created IS '按创建时间倒序查询（项目内排序）';
COMMENT ON INDEX public.idx_issues_state_covering IS '看板列内按状态+排序字段查询（WHERE deleted_at IS NULL）';

-- ================ sprints 表索引 ================
COMMENT ON INDEX public.idx_one_active_sprint_per_project IS '保证每项目最多一个激活迭代（WHERE status = active AND deleted_at IS NULL）';
COMMENT ON INDEX public.idx_sprints_active_unique IS '激活迭代唯一性+按项目查询（WHERE status = active AND deleted_at IS NULL）';
COMMENT ON INDEX public.idx_sprints_project_status IS '按项目+状态查询迭代列表（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_sprints_version IS '按发布版本查找关联迭代（WHERE deleted_at IS NULL）';

-- ================ versions 表索引 ================
COMMENT ON INDEX public.idx_versions_project_status IS '按项目+状态查询版本列表（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_versions_unique_semver IS '同项目下 SemVer 唯一性保证（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_versions_workspace IS '按空间查询版本列表（WHERE deleted_at IS NULL）';

-- ================ issues 关联表索引 ================
COMMENT ON INDEX public.idx_issue_deps_pred IS '查找指定工作项的所有后继依赖（按 predecessor_id）';
COMMENT ON INDEX public.idx_issue_deps_succ IS '查找指定工作项的所有前驱依赖（按 successor_id）';
COMMENT ON INDEX public.idx_issue_relations_source IS '查找源工作项的所有关联关系（按 source_issue_id）';
COMMENT ON INDEX public.idx_issue_relations_target IS '查找目标工作项的所有关联关系（按 target_issue_id）';
COMMENT ON INDEX public.idx_issue_assignees_user IS '查找用户负责的所有工作项（按 user_id）';
COMMENT ON INDEX public.idx_issue_watchers_user IS '查找用户关注的所有工作项（按 user_id）';
COMMENT ON INDEX public.idx_sprint_issues_issue IS '按工作项反查所在迭代（按 issue_id）';

-- ================ notification 模块索引 ================
COMMENT ON INDEX public.idx_notifications_recipient_unread IS '查询用户未读通知（WHERE is_archived = false，按 created_at 倒序）';
COMMENT ON INDEX public.idx_notifications_entity IS '按关联实体类型+ID 查找通知';
COMMENT ON INDEX public.idx_notifications_archived IS '查询归档通知（WHERE is_archived = true）';
COMMENT ON INDEX public.idx_deliveries_next_retry IS '查询待重试投递（WHERE status = pending，按 next_retry_at）';
COMMENT ON INDEX public.idx_deliveries_notification IS '按通知 ID 查找投递记录';
COMMENT ON INDEX public.idx_deliveries_status IS '按投递状态查询（WHERE status = pending）';
COMMENT ON INDEX public.idx_digests_pending IS '查询待发送的聚合摘要（WHERE status = pending，按 scheduled_for）';

-- ================ search 模块索引 ================
COMMENT ON INDEX public.idx_search_documents_tsv IS '搜索文档 GIN 全文检索索引';
COMMENT ON INDEX public.idx_search_documents_unique IS '搜索文档唯一性保证（workspace_id + doc_type + doc_id）';
COMMENT ON INDEX public.idx_search_documents_workspace IS '按空间+类型查找搜索文档';
COMMENT ON INDEX public.idx_search_documents_project IS '按项目+类型查找搜索文档';
COMMENT ON INDEX public.idx_search_history_user IS '按用户查找搜索历史（按 searched_at 倒序）';

-- ================ 其他模块索引 ================
COMMENT ON INDEX public.idx_attachments_entity IS '按实体类型+ID 查找附件（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_attachments_uploader IS '按上传者查找附件（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_audit_logs_ws_time IS '按空间+时间查询审计日志';
COMMENT ON INDEX public.idx_time_logs_issue IS '按工作项查询工时记录（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_time_logs_user_date IS '按用户+日期查询工时记录（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_risk_alerts_project IS '按项目查询未解决风险告警（WHERE NOT is_resolved）';
COMMENT ON INDEX public.idx_risk_alerts_unresolved IS '按空间+严重度查询未解决告警（WHERE NOT is_resolved）';
COMMENT ON INDEX public.idx_recent_items_user IS '按用户查询最近访问记录（按 accessed_at 倒序）';
COMMENT ON INDEX public.idx_intake_issues_channel IS '按渠道查询入口工单（按 created_at 倒序）';
COMMENT ON INDEX public.idx_intake_issues_status IS '按空间+状态查询入口工单';
COMMENT ON INDEX public.idx_events_unpublished IS '查询未发布领域事件（WHERE published_at IS NULL，Outbox 轮询）';
COMMENT ON INDEX public.idx_password_reset_tokens_expires IS '按过期时间查询令牌（用于清理任务）';
COMMENT ON INDEX public.idx_password_reset_tokens_user_active IS '查找用户有效重置令牌（WHERE used_at IS NULL，唯一约束）';
COMMENT ON INDEX public.idx_deployment_events_project IS '按项目查询部署事件（按 deployed_at 倒序）';
COMMENT ON INDEX public.idx_deployment_events_ws IS '按空间查询部署事件';

-- ============================================================================
-- 全部对象注释注入完毕。补丁执行完毕。
-- ============================================================================

