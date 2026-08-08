#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
inline_comments.py — 为 ydzs-plane-init.sql 添加内联 COMMENT ON 注释

功能:
  1. 扫描 CREATE TABLE 结束位置，注入 COMMENT ON TABLE / COLUMN
  2. 扫描 CREATE TRIGGER 结束位置，注入 COMMENT ON TRIGGER
  3. 扫描 CREATE INDEX 结束位置，注入索引用途注释
  4. 幂等执行（已插入的注释自动跳过）
  5. 实时打印进度 + 最终统计

用法:
  python inline_comments.py
"""

import os
import sys
import re


def _file_exists_via_listdir(path):
    """sandbox 环境下 os.path.isfile 不可靠时使用 listdir 验证"""
    d = os.path.dirname(path)
    n = os.path.basename(path)
    try:
        return os.path.isdir(d) and n in os.listdir(d)
    except OSError:
        return False


def resolve_sql_path():
    """多策略解析 SQL 文件路径（兼容 sandbox/非 sandbox 环境）"""
    candidates = []
    # 策略 1: 脚本所在目录
    script_dir = os.path.dirname(os.path.abspath(__file__))
    candidates.append(os.path.join(script_dir, "ydzs-plane-init.sql"))
    # 策略 2: 当前工作目录下的 sql/ 子目录
    candidates.append(os.path.join(os.getcwd(), "sql", "ydzs-plane-init.sql"))
    # 策略 3: 已知项目路径
    candidates.append(r"D:\Code\open\ydsz-plane\sql\ydsz-plane-init.sql")
    for p in candidates:
        if os.path.isfile(p) or _file_exists_via_listdir(p):
            return p
    # 全部失败 → 返回第一个候选（让后续 open 报 FileNotFoundError）
    return candidates[0]


def check_file_readable(path):
    """双重验证路径可读（isfile 或 listdir 任一通过即可）"""
    return os.path.isfile(path) or _file_exists_via_listdir(path)


SQL_FILE = resolve_sql_path()

# ─────────────────────────────────────────────────────────────────────
# 一、表级注释
# ─────────────────────────────────────────────────────────────────────
TABLE_COMMENTS = {
    # ── 核心域 ────────────────────────────────────────
    "issues":         "工作项主表（需求/任务/缺陷统一存储，支撑看板、迭代、搜索、关联等核心能力）",
    "labels":         "标签表（工作项分类标记，支持颜色、描述、软删除）",
    "modules":        "模块/组件表（项目内功能域划分，层级结构 ≤3 层，用于工作项分类）",
    "issue_labels":   "工作项-标签多对一关联表",
    "issue_modules":  "工作项-模块多对一关联表",
    "issue_assignees":"工作项-指派人物多对一关联表（含主负责人标记）",
    "issue_watchers": "工作项-关注人多对一关联表",
    "issue_comments": "工作项评论表（TipTap JSON 富文本，支持 @提及、回复、反应）",
    "issue_activities":"工作项活动历史表（按月 RANGE 分区；字段 diff / 流转 / 附件 / 关联等全量审计）",
    "issue_dependencies":"工作项依赖关系表（FS/SS/FF/SF + lag_days，DFS 检测环）",
    "issue_relations":"工作项语义关联表（duplicate/relates_to/blocked_by/start_before/finish_before/implemented_by）",
    "time_logs":      "工时记录表（单位: 分钟；关联工作项 + 用户，差值回写 actual_effort）",
    "attachments":    "附件表（多态关联: issue/comment/workspace/project/user；元数据 JSONB）",
    # ── 项目域 ────────────────────────────────────────
    "projects":       "项目表（工作项聚合根容器，含状态模板、工作流配置、封面、默认外观）",
    "workspaces":     "工作空间表（多租户顶层容器，RLS 依据、通知配置、品牌化、SSO）",
    "invitations":    "工作空间邀请表（邮件邀请码 + 角色 + 过期 + 幂等）",
    "workspace_members":"工作空间成员表（user↔workspace 多对一，含 owner/admin/member/guest 角色）",
    "project_members":"项目成员表（user↔project 多对一，含 admin/member/viewer 角色）",
    # ── 迭代域 ────────────────────────────────────────
    "sprints":        "敏捷迭代表（生命周期 planned→active→completed；容量/目标/速率/乐观锁）",
    "sprint_issues":  "迭代-工作项关联表（含中途加项标记 added_midway，复盘报告使用）",
    "sprint_snapshots":"迭代燃尽快照表（Cron 每日 00:05 写入；by_state_group 字段支撑燃起图/CFD）",
    # ── 版本域 ────────────────────────────────────────
    "versions":       "版本发布表（SemVer 语义化版本；生命周期 planning→active→released→archived；聚合跨迭代进度）",
    "version_delivery_snapshots":"版本交付快照表（发布时生成的交付报告数据: 缺陷数/通过率/准出率等）",
    # ── 状态机域 ──────────────────────────────────────
    "states":         "状态表（项目级自定义状态集；group ∈ backlog/unstarted/started/completed/cancelled/triage）",
    "state_transitions":"状态流转规则表（按 project×type 维度定义合法流转；含必填字段、允许角色约束）",
    # ── 用户域 ────────────────────────────────────────
    "users":          "平台用户表（跨工作空间；认证信息、MFA、时区、偏好设置）",
    "password_reset_tokens":"密码重置令牌表（UUID 一次性令牌 + 过期时间 + used_at）",
    "audit_logs":     "审计日志表（只增不改；登录/权限变更/删除/Token/Webhook 等安全操作；在线 12 个月）",
    "api_tokens":     "API Token 表（scopes 白名单；仅存 SHA-256 hash；支持过期 + last_used_at 审计）",
    "idempotency_keys":"API 幂等键表（写操作去重窗口；复用 response 缓存）",
    # ── 通知域 ────────────────────────────────────────
    "notifications":  "站内信表（按月分区；recipient/actor/issue/event_type/data JSONB；WS 实时推送 + 未读数缓存）",
    "notification_preferences":"用户通知偏好订阅表（按 项目×事件类型×渠道三维开关；支持免打扰 digest）",
    "notification_deliveries":"通知渠道投递记录表（in_app/email/wecom/dingtalk/feishu；含重试/失败状态；分区）",
    "notification_digests":"通知摘要暂存表（daily/weekly 模式按用户+渠道聚合；定时合并发送）",
    # ── 收件箱域 ────────────────────────────────────────
    "intake_channels":"收件箱入口通道表（公开门户 slug；限流 + 行为验证码 + 自动分配规则）",
    "intake_issues":  "收件箱提交记录表（tracking_id YD-IN-XXXX 提交回执；status open/accepted/rejected/archived）",
    # ── 自动化域 ──────────────────────────────────────
    "automation_rules":"自动化规则表（JSON DSL: trigger/conditions/actions；支持试运行 dry-run）",
    "automation_templates":"自动化内置模板表（7 条开箱即用模板；创建项目时可批量复制）",
    "rule_executions":"规则执行审计表（按月分区；status/duration/results；连续失败 3 次触发熔断）",
    # ── 仪表盘域 ────────────────────────────────────────
    "dashboard_widgets":"仪表盘卡片实例表（type→渲染器/数据接口映射；config JSONB 个性化配置）",
    "dashboard_templates":"仪表盘预设模板表（敏捷项目/版本交付/PMO 多项目；布局+卡片组合 JSON）",
    "dashboard_snapshots":"仪表盘快照表（定时导出/分享大屏只读链接数据）",
    # ── 工作台域 ────────────────────────────────────────
    "workbench_configs":"个人工作台配置表（置顶项目/收藏视图/自定义组件布局；per-user）",
    "workbench_templates":"工作台预设模板表（供创建账号时初始化工作台）",
    # ── 视图域 ────────────────────────────────────────
    "view_preferences":"视图偏好表（按项目隔离；排序/过滤/分组/字段裁剪配置 upsert）",
    "recent_items":   "最近访问记录表（用户最近查看/操作的工作项；touch 触发 updated_at 更新）",
    # ── 搜索域 ────────────────────────────────────────
    "search_bookmarks":"搜索收藏查询表（用户保存的常用 JQL 查询；可分享）",
    "search_documents":"搜索文档表（ES 索引失败的降级兜底；PG to_tsvector 源）",
    "search_history": "搜索历史表（localStorage 20 条 + 服务端同步；自动补全数据源）",
    # ── 风险域 ────────────────────────────────────────
    "risk_rules":     "风险预警规则表（监控阈值: 逾期/阻塞/依赖链/燃尽偏差；越限→通知）",
    "risk_alerts":    "风险告警记录表（越限事件含快照、关联问题、处理状态）",
    # ── 度量域 ────────────────────────────────────────
    "metric_snapshots":"效能指标快照表（每日 01:30 聚合 job；granularity daily/sprint/version；幂等 upsert）",
    "metric_adjustments":"指标修正记录表（admin 校准用；不覆盖原值，查询时叠加；全程审计）",
    # ── 集成域 ────────────────────────────────────────
    "webhooks":       "Webhook 配置表（target_url + secret HMAC + 事件白名单 + SSRF 防护）",
    "webhook_logs":   "Webhook 投递日志表（按月分区 + 30 天 TTL；request/response/耗时；手动重投）",
    "deployment_events":"部署事件表（DORA 数据源；CI/CD Webhook 推送；HMAC 验签；env+status+commit_sha）",
    # ── Outbox / 基础设施域 ────────────────────────────
    "domain_events":  "领域事件表（Transactional Outbox 模式；未发布事件 id 排序投递 + 7 天清理）",
    "schema_migrations":"Schema 迁移记录表（Flyway 风格版本号；CI 校验连续性）",
    "project_sequences":"项目发号器表（project_id → next_value 原子自增；发工作项编号 YD-123；允许跳号）",
    # ── 知识库 / 文档域 ──────────────────────────────
    "pages":          "页面表（项目内自定义页面；TipTap 富文本内容；支持嵌入看板/指标）",
    # ── 表格空间 / 通用 ──────────────────────────────
    "estimate_points":"估算点配置表（默认斐波那契数列；项目可自定义；用于敏捷估算）",
    "issue_templates":"工作项模板表（预填充标题/描述/类型/状态；创建时可选用）",
}

# ─────────────────────────────────────────────────────────────────────
# 二、列级注释  { table → { column → comment } }
# ─────────────────────────────────────────────────────────────────────
COLUMN_COMMENTS = {
    # ── issues ─────────────────────────────────────────
    "issues": {
        "id":                "主键 ID（GENERATED ALWAYS AS IDENTITY，不外泄）",
        "public_id":         "对外暴露的唯一标识（UUID），用于 API 与外部引用",
        "workspace_id":      "工作空间/租户 FK，RLS 依据，复合索引首列",
        "project_id":        "所属项目 FK（projects.id），聚合根容器",
        "sequence_id":       "项目内自增序号，配合 project.identifier 展示为 YD-123",
        "type_code":         "工作项类型: requirement(需求) / task(任务) / defect(缺陷)",
        "parent_id":         "WBS 父工作项 FK（issues.id），NULL=顶级，limit depth ≤3",
        "depth":             "WBS 冗余层级（1..3），父项 depth+1 自动填充，>3 拒绝",
        "name":              "工作项标题（短文本，索引全文检索命中）",
        "description_json":  "TipTap 编辑器结构化 JSON（Node 数组，ProseMirror 格式）",
        "description_html":  "从 description_json 渲染的 HTML（展示层 + 通知邮件富文本）",
        "description_stripped":"纯文本（HTML strip 后），tsvector 全文字段索引源",
        "state_id":          "状态 FK（states.id）；group ∈ backlog/unstarted/started/completed/cancelled/triage",
        "priority":          "优先级: urgent(紧急) / high(高) / medium(中) / low(低) / none(无)",
        "severity":          "缺陷严重程度（1=致命..5=轻微）；仅 type_code=defect 使用，必填",
        "found_phase":       "缺陷发现阶段: unit(单元测试) / integration(集成) / uat(验收) / production(生产) / customer(客户反馈)",
        "root_cause_category":"缺陷根因分类: requirement(需求) / technical(技术) / environment(环境) / data(数据)；流转至 completed 时必填",
        "verifier_id":       "验证人 FK（users.id）；缺陷待验证阶段可指派",
        "environment":       "缺陷发现环境 JSONB（OS / Browser / Device / AppVersion）",
        "reproduce_steps":   "缺陷复现步骤 JSONB {steps:[], expected:'', actual:''}",
        "category":          "工作项分类标签: frontend/backend/qa/devops/design/doc 等（自由填写）",
        "actual_effort":     "实际已用工时 NUMERIC(8,2) 单位: 小时；time_logs sum 回写",
        "remaining_effort":  "剩余预估工时 NUMERIC(8,2) 单位: 小时；0 表示完成",
        "delay_reason":      "延期原因: requirement_change(需求变更) / resource(资源) / blocked(阻塞) / other(其他)",
        "source":            "需求来源: customer(客户) / internal(内部) / competitor(竞品) / other",
        "point":             "故事点估算 SMALLINT 0-12；斐波那契数列 0,1,2,3,5,8,13",
        "sprint_id":         "归属迭代 FK（sprints.id），同一项目内一个活跃迭代（可配置）",
        "progress":          "完成百分比 0-100（冗余字段；子项 state.group=completed 时事件触发的回写）",
        "start_date":        "计划开始日期（用户指定；逾期触发 risk_rule 告警）",
        "target_date":       "目标完成日期（用户指定；逾期触发 risk_rule 告警）",
        "completed_at":      "实际完成时间 TIMESTAMPTZ；进入 completed 状态时自动赋值",
        "is_draft":          "草稿标记: true=草稿(仅草稿流可见)，false=正式发布；默认 false",
        "sort_order":        "看板列内排序权重 DOUBLE PRECISION（默认 65535 末尾追加；中值插入；碎片化触发重排）",
        "created_by":        "创建人 FK（users.id）；通知默认接收人",
        "created_at":        "创建时间（迁移写入后不可变）",
        "updated_at":        "最后修改时间（触发器 trg_xxx_updated_at 自动维护 now()）",
        "deleted_at":        "软删除时间戳；NULL=有效；部分索引 WHERE deleted_at IS NULL 排除软删除",
        "version":           "乐观锁版本号（默认 1）；UPDATE 条件带 version，冲突返回 409",
        "found_version_id":  "缺陷发现版本 FK（versions.id）；type_code=defect 时关联",
        "fix_version_id":    "缺陷修复版本 FK（versions.id）；流转至待验证/已修复时必填",
        "release_version_id":"首次发布版本 FK（versions.id）；需求/任务在发布时回填",
        "search_tsv":        "tsvector 全文索引（simple 配置，中文降级兜底；ES 为主）",
    },
    # ── labels ─────────────────────────────────────────
    "labels": {
        "id":           "主键 ID",
        "workspace_id": "工作空间 FK",
        "name":         "标签名称（工作空间+项目内唯一，部分索引排除软删除）",
        "color":        "标签颜色 HEX 值（如 #FF5733）",
        "description":  "标签用途说明（可选）",
        "project_id":   "所属项目 FK（NULL=工作空间级全局标签）",
        "created_by":   "创建人 FK",
        "created_at":   "创建时间",
        "updated_at":   "修改时间（触发器自动维护）",
        "deleted_at":   "软删除时间戳",
    },
    # ── modules ────────────────────────────────────────
    "modules": {
        "id":           "主键 ID",
        "workspace_id": "工作空间 FK",
        "project_id":   "所属项目 FK",
        "name":         "模块/组件名称",
        "description":  "模块功能说明",
        "parent_id":    "父模块 FK（modules.id）；层级 ≤3，顶层为 NULL",
        "lead_id":      "模块负责人 FK（users.id）",
        "created_by":   "创建人 FK",
        "created_at":   "创建时间",
        "updated_at":   "修改时间（触发器自动维护）",
        "deleted_at":   "软删除时间戳",
    },
    # ── issue_labels / issue_modules / issue_assignees / issue_watchers ──
    "issue_labels": {
        "issue_id":    "工作项 FK",
        "label_id":    "标签 FK",
        "created_at":  "关联创建时间",
    },
    "issue_modules": {
        "issue_id":    "工作项 FK",
        "module_id":   "模块 FK",
        "created_at":  "关联创建时间",
    },
    "issue_assignees": {
        "issue_id":    "工作项 FK",
        "user_id":     "被指派人 FK",
        "is_primary":  "是否主负责人: true=主, false=辅助；每个工作项仅一个主负责人",
        "created_at":  "指派时间",
    },
    "issue_watchers": {
        "issue_id":    "工作项 FK",
        "user_id":     "关注人 FK；事件驱动通知（issue.updated/issue.commented 等）",
        "created_at":  "关注时间",
    },
    # ── issue_activities ───────────────────────────────
    "issue_activities": {
        "id":           "主键 ID",
        "issue_id":     "关联工作项 FK",
        "actor_id":     "操作人 FK（users.id）；系统动作为 NULL",
        "verb":         "操作动词: created/updated/transitioned/attached/linked/detached/deleted/restored/commented/mentioned",
        "field":        "变更字段名（verb=updated 时使用；如 state, priority, assignees）",
        "old_value":    "变更前值（TEXT 或 JSONB 序列化）",
        "new_value":    "变更后值（TEXT 或 JSONB 序列化）",
        "metadata":     "附加信息 JSONB（如流转 from→to、字段 diff 详情）",
        "created_at":   "记录时间 TIMESTAMPTZ",
        "workspace_id": "工作空间 FK（支撑 RLS 与按月分区键）",
    },
    # ── issue_dependencies ─────────────────────────────
    "issue_dependencies": {
        "id":             "主键 ID",
        "workspace_id":   "工作空间 FK",
        "issue_id":       '前置工作项 FK（dependencies 的"from"）',
        "dependent_id":   '后续工作项 FK（dependencies 的"to"）',
        "dependency_type":"依赖类型: FS(完成→开始) / SS(开始→开始) / FF(完成→完成) / SF(开始→完成)",
        "lag_days":       "延迟天数（正=延迟等待，负=提前开始）；0 表示无延迟",
        "created_by":     "创建人 FK",
        "created_at":     "创建时间",
        "deleted_at":     "软删除时间戳（唯一约束 WHERE deleted_at IS NULL）",
    },
    # ── issue_relations ────────────────────────────────
    "issue_relations": {
        "id":           "主键 ID",
        "workspace_id": "工作空间 FK",
        "issue_a_id":   "工作项 A FK",
        "issue_b_id":   "工作项 B FK",
        "relation_type":"语义关联类型: duplicate(重复) / relates_to(关联) / blocked_by(被阻塞) / start_before(先于开始) / finish_before(先于完成) / implemented_by(由…实现)",
        "created_by":   "创建人 FK",
        "created_at":   "创建时间",
    },
    # ── time_logs ───────────────────────────────────────
    "time_logs": {
        "id":           "主键 ID",
        "issue_id":     "关联工作项 FK",
        "user_id":      "登记人 FK",
        "minutes":      "工时分钟数（正整数；写入/编辑/删除差值回写 actual_effort）",
        "log_date":     "工时日期（用户可指定历史日期；默认当天）",
        "description":  "工时说明（可选: 做了什么）",
        "billable":     "是否计费工时: true=计费, false=不计费",
        "created_by":   "登记人/修改人 FK",
        "created_at":   "创建时间",
        "updated_at":   "修改时间",
        "deleted_at":   "软删除时间戳",
    },
    # ── attachments ─────────────────────────────────────
    "attachments": {
        "id":             "主键 ID",
        "attachable_type":"多态类型: issue / comment / workspace / project / user / intake_issue",
        "attachable_id":  "多态关联 ID（bigint 统一；配合 attachable_type 唯一）",
        "workspace_id":   "工作空间 FK（RLS 依据）",
        "file_name":      "原始文件名（用户上传时显示名）",
        "file_size":      "文件大小（字节；最大 10MB）",
        "content_type":   "MIME content-type（类型白名单校验）",
        "storage_key":    "MinIO 对象存储 key（UUID，文件名已重命名）",
        "metadata":        "附件元数据 JSONB（图片尺寸/时长/EXIF/CRC32 等）",
        "uploaded_by":    "上传人 FK",
        "created_at":     "上传时间",
        "deleted_at":     "软删除时间戳",
    },
    # ── projects ────────────────────────────────────────
    "projects": {
        "id":               "主键 ID",
        "workspace_id":     "所属工作空间 FK",
        "name":             "项目名称",
        "identifier":       "项目标识符（大写 2-10 字符，工作空间内唯一；用于 YD-123 编号）",
        "description":      "项目描述（富文本/Markdown）",
        "state":            "项目状态: active(活跃) / archived(归档) / deleted(软删除)",
        "default_view":     "默认视图偏好: list/board/calendar/gantt",
        "cover_image":      "封面图片附件 ID",
        "start_date":       "项目开始日期",
        "target_date":      "项目目标日期",
        "created_by":       "创建人 FK",
        "created_at":       "创建时间",
        "updated_at":       "修改时间（触发器自动维护）",
        "deleted_at":       "软删除时间戳",
        "version":          "乐观锁版本号（默认 1）",
        "is_default":       "是否工作空间默认项目: true=默认（用于项目选择器/快捷入口）",
        "icon":             "项目图标（Emoji / Lucide 图标名）",
    },
    # ── sprints ─────────────────────────────────────────
    "sprints": {
        "id":           "主键 ID",
        "workspace_id": "所属工作空间 FK",
        "project_id":   "所属项目 FK",
        "name":         '迭代名称（如"Sprint 42 - 用户画像"）',
        "goal":         "迭代目标（简短描述；燃尽图上方展示）",
        "status":       "迭代状态: planned(计划中) / active(进行中) / completed(已结束)",
        "start_date":   "迭代开始日期（active 时必填；同一项目同一时间仅一个 active）",
        "end_date":     "迭代结束日期（active 时必填；结束日触发 sprint.ending_soon 提醒）",
        "capacity":     "团队容量（人天）；与故事点总和对比计算饱和度",
        "total_points": "当前迭代总故事点（redundant，SprintIssue 事件回写更新）",
        "done_points":   "已完成故事点（redundant；SprintIssue 完成事件回写）",
        "created_by":   "创建人 FK",
        "created_at":   "创建时间",
        "updated_at":   "修改时间（触发器自动维护）",
        "deleted_at":   "软删除时间戳",
        "version":      "乐观锁版本号（默认 1）；UPDATE 条件带 version 防并发冲突",
    },
    # ── sprint_issues ───────────────────────────────────
    "sprint_issues": {
        "id":             "主键 ID",
        "sprint_id":      "迭代 FK",
        "issue_id":       "工作项 FK",
        "added_midway":   "是否中途加入: true=迭代启动后新增（复盘报告单独统计对速率影响）",
        "created_at":     "关联创建时间（即工作项加入迭代的时间点）",
        "workspace_id":   "工作空间 FK（RLS 依据 + 复合索引）",
    },
    # ── sprint_snapshots ────────────────────────────────
    "sprint_snapshots": {
        "id":             "主键 ID",
        "sprint_id":      "迭代 FK",
        "snapshot_date":  "快照日期（每 sprint+date 唯一；Cron 每日 00:05 写入）",
        "total_points":   "当日总计划故事点",
        "done_points":    "当日已完成故事点",
        "by_state_group": "各状态组故事点分布 JSONB {backlog:N, unstarted:N, started:N, completed:N}",
        "added_points":   "启动后新增故事点（复盘中期加入影响计算）",
        "removed_points":  "启动后移除故事点",
        "created_at":     "写入时间",
        "workspace_id":   "工作空间 FK（RLS 依据）",
    },
    # ── versions ────────────────────────────────────────
    "versions": {
        "id":              "主键 ID",
        "workspace_id":    "工作空间 FK",
        "project_id":      "所属项目 FK",
        "name":            '版本展示名（如"用户画像一期"）',
        "semver":          "语义化版本号（如 1.2.3；项目内唯一；发布后只读）",
        "description":     "版本目标与范围说明",
        "status":          "版本状态: planning(规划中) / active(进行中) / released(已发布) / archived(已归档)",
        "start_date":     "计划开始时间（可选）",
        "end_date":       "计划结束时间（可选）",
        "target_date":    "计划发布日（版本日；触发风险预警: 到期未发布/剩余缺陷）",
        "checklist":      "发布检查清单 JSONB [{id,label,required,checked}]；全部勾选才可 release",
        "release_notes":  "Release Notes（发布时按模板三段式生成: 需求/缺陷修复/已知问题；可编辑）",
        "delivered_at":   "实际发布 TIMESTAMPTZ（发布动作时写入）",
        "archived_at":    "归档时间 TIMESTAMPTZ",
        "delivery_report":"交付报告 JSONB（缺陷数/通过率/准出率/迭代完成度明细；发布时生成）",
        "progress":       "聚合进度 0-100（读时计算；缓存版本失效键 version:{id}:progress）",
        "created_by":     "创建人 FK",
        "created_at":     "创建时间",
        "updated_at":     "修改时间（触发器自动维护）",
        "deleted_at":     "软删除时间戳",
        "version":        "乐观锁版本号（默认 1）",
    },
    # ── version_delivery_snapshots ──────────────────────
    "version_delivery_snapshots": {
        "id":              "主键 ID",
        "version_id":      "版本 FK",
        "snapshot_date":   "快照时间（版本维度每日/里程碑聚合）",
        "total_points":    "当日版本总计划故事点",
        "done_points":     "当日版本已完成故事点",
        "bug_count":       "缺陷总数（含未关闭）",
        "open_bug_count":  "未关闭缺陷数",
        "pass_rate":       "测试通过率百分比",
        "deployment_count":"累计部署次数（计数；与 DORA-DF 对齐）",
        "metrics":        "详细效能指标 JSONB（逃逸率/返工率/各状态分布等）",
        "created_at":     "写入时间",
        "workspace_id":   "工作空间 FK（RLS 依据）",
    },
    # ── states ─────────────────────────────────────────
    "states": {
        "id":            "主键 ID",
        "workspace_id":  "工作空间 FK",
        "project_id":    "所属项目 FK",
        "name":          '状态显示名（如"ToDo"/"In Progress"/"Done"）',
        "group":         "状态组: backlog / unstarted / started / completed / cancelled / triage（前端据此渲染卡片颜色 + 看板列）",
        "color":         "状态颜色 HEX 值",
        "sequence":      "状态在组内排序权重（升序）",
        "description":   "状态含义说明（可选）",
        "is_default":    "是否项目默认状态: true=新工作项初始状态（每个 type_code 一个默认）",
        "type_code":     "工作项类型: requirement / task / defect；决定状态集隔离",
        "created_by":    "创建人 FK",
        "created_at":    "创建时间",
        "updated_at":    "修改时间（触发器自动维护）",
        "deleted_at":    "软删除时间戳",
    },
    # ── state_transitions ───────────────────────────────
    "state_transitions": {
        "id":              "主键 ID",
        "workspace_id":    "工作空间 FK",
        "project_id":      "所属项目 FK",
        "type_code":       "工作项类型: requirement / task / defect；独立流转集",
        "from_state_id":   "起始状态 FK（states.id）",
        "to_state_id":     "目标状态 FK（states.id）",
        "required_fields": "流转必填字段 JSONB [{field, condition}]（如缺陷→已完成要求 root_cause_category 非空）",
        "allowed_roles":   "允许执行的角色列表 JSONB ['owner','admin',...]；空数组=继承项目角色默认",
        "created_by":      "创建人 FK",
        "created_at":      "创建时间",
        "updated_at":      "修改时间（触发器自动维护）",
        "deleted_at":      "软删除时间戳",
    },
    # ── users ──────────────────────────────────────────
    "users": {
        "id":             "主键 ID",
        "email":          "邮箱（小写唯一；登录主凭证；CI 强制唯一）",
        "password_hash":  "bcrypt(cost=12) 密码哈希值",
        "first_name":     "名字",
        "last_name":      "姓氏",
        "display_name":   "显示名（默认 first + last；可自定义）",
        "avatar":         "头像附件 ID（attachments.id）",
        "is_active":      "账号激活状态: true=激活, false=禁用（软锁定）",
        "is_email_verified":"邮箱是否已验证: true=已验证, false=未验证（发送验证邮件）",
        "mfa_secret":     "TOTP 双因子密钥（AES-GCM 加密存储；NULL=未启用）",
        "last_login_at":  "最后登录时间（登录成功后更新；判断账号活跃度）",
        "timezone":       "用户时区（IANA 名称如 Asia/Shanghai；用于 digest/免打扰计算）",
        "locale":         "偏好语言 (en/zh-CN/zh-TW/ja/ko)",
        "preferences":    "用户偏好 JSONB（主题/快捷手势/通知默认开关等）",
        "created_at":     "注册时间",
        "updated_at":     "修改时间（触发器自动维护）",
        "deleted_at":     "软删除时间戳",
    },
    # ── workspaces ──────────────────────────────────────
    "workspaces": {
        "id":              "主键 ID",
        "name":            "工作空间名称（唯一标识租户）",
        "slug":            "URL 友好唯一标识（小写 + 连字符；用于子域名/API 路由）",
        "description":     "工作空间简介（可选）",
        "logo":            "品牌 Logo 附件 ID",
        "brand_colors":    "品牌色 JSONB {primary, secondary, accent} HEX 值",
        "default_role":    "邀请新用户默认角色: member / guest",
        "is_active":       "工作空间激活状态: true=正常, false=暂停（SSO/限流配置）",
        "settings":        "工作空间设置 JSONB（SSO/安全策略/通知通道/附件配置）",
        "created_by":      "创建人 FK（owner；默认拥有 owner 角色）",
        "created_at":      "创建时间",
        "updated_at":      "修改时间（触发器自动维护）",
        "deleted_at":      "软删除时间戳",
    },
    # ── invitations ─────────────────────────────────────
    "invitations": {
        "id":           "主键 ID",
        "workspace_id": "目标工作空间 FK",
        "email":        "被邀請人邮箱（小写；workspace+email 唯一排除软删除）",
        "role":         "邀请角色: owner / admin / member / guest",
        "status":       "邀请状态: pending(待确认) / accepted / rejected / expired",
        "token":        "邀请校验令牌（UUID；SHA-256 hash 存储；邮件内链接使用）",
        "expires_at":   "邀请有效期 TIMESTAMPTZ；过期自动标记 status=expired",
        "invited_by":   "邀请人 FK",
        "accepted_at":  "接受时间（跳转注册/登录后）",
        "created_at":   "创建时间",
        "deleted_at":   "软删除时间戳",
    },
    # ── workspace_members ───────────────────────────────
    "workspace_members": {
        "id":           "主键 ID",
        "workspace_id": "工作空间 FK",
        "user_id":      "用户 FK",
        "role":         "工作空间级角色: owner / admin / member / guest；用于权限点收敛",
        "is_active":    "成员状态: true=激活, false=暂停（不立即离职，恢复使用）",
        "joined_at":    "加入时间（接受邀请/被添加时）",
        "created_by":   "添加人 FK",
        "created_at":   "创建时间",
        "updated_at":   "修改时间（触发器自动维护）",
    },
    # ── project_members ─────────────────────────────────
    "project_members": {
        "id":           "主键 ID",
        "project_id":   "项目 FK",
        "user_id":      "用户 FK",
        "role":         "项目级角色: admin / member / viewer；effective = max(ws_role, project_role)",
        "is_active":    "成员状态: true=激活, false=暂停",
        "joined_at":    "加入时间",
        "created_by":   "添加人 FK",
        "created_at":   "创建时间",
        "updated_at":   "修改时间（触发器自动维护）",
    },
    # ── notifications ───────────────────────────────────
    "notifications": {
        "id":           "主键 ID",
        "workspace_id": "工作空间 FK（RLS + 按月分区键）",
        "recipient_id": "接收人 FK（users.id；通知投递主目标）",
        "actor_id":     "触发人 FK（users.id）；自我豁免: 操作者==接收人不落库",
        "issue_id":     "关联工作项 FK（可 NULL；点击跳转定位）",
        "event_type":   "事件类型: issue.created/updated/status_changed/assigned/commented/mentioned；sprint.* / version.* / automation.*",
        "title":        "通知标题（多语言模板渲染后；纯文本）",
        "body":         "通知正文（纯文本摘要；邮件使用 HTML 模板二次渲染）",
        "data":         "跳转上下文 JSONB {url, issue_identifier, project_identifier}；前端点击定位",
        "is_read":      "已读状态: true=已读(已点击), false=未读；未读数缓存 unread:{uid}",
        "read_at":      '首次阅读时间 TIMESTAMPTZ（列表"全部已读"触发）',
        "created_at":   "创建时间 TIMESTAMPTZ（列表/游标分页倒序 + 分区裁剪）",
    },
    # ── notification_preferences ───────────────────────
    "notification_preferences": {
        "id":           "主键 ID",
        "user_id":      "用户 FK",
        "scope":        "订阅范围: workspace / project；决定 ref_id 语义",
        "ref_id":       "范围 ID（scope=workspace → workspace_id；scope=project → project_id）",
        "event_type":   "事件类型（通配: issue.* / sprint.* / version.* / automation.*）",
        "channel":      "通知渠道: in_app / email / wecom / dingtalk / feishu",
        "is_enabled":   "渠道开关: true=启用, false=禁用（覆盖默认矩阵）",
        "digest":       "投递模式: realtime(实时) / daily(每日 08:30 聚合) / weekly(每周一聚合)",
        "dnd_start":    "免打扰开始时间 (HH:MM 用户时区 如 22:00)；期间 realtime 降级为 digest",
        "dnd_end":      "免打扰结束时间 (HH:MM 如 08:00)；高优事件(mention/automation.failed)可豁免",
        "created_at":   "创建时间",
        "updated_at":   "修改时间（触发器自动维护）",
    },
    # ── notification_deliveries ─────────────────────────
    "notification_deliveries": {
        "id":            "主键 ID",
        "notification_id":"站内信 FK（notifications.id）",
        "channel":       "发货渠道: email / wecom / dingtalk / feishu",
        "status":        "投递状态: pending / sent / failed / retrying",
        "attempt_count": "重试次数（最大 3 次；指数退避 1min/5min/30min）",
        "error_message": "最后错误信息（用于排障；不存敏感内容）",
        "sent_at":       "最终投递成功时间",
        "created_at":    "创建时间（分区键；按月 RANGE 分区）",
        "workspace_id":  "工作空间 FK（RLS 依据）",
    },
    # ── notification_digests ─────────────────────────────
    "notification_digests": {
        "id":            "主键 ID",
        "user_id":       "用户 FK",
        "channel":       "聚合渠道: email / wecom / dingtalk / feishu",
        "payload":       "聚合数据 JSONB [{event_type, issue_id, title, body}]；按项目分组",
        "scheduled_for": "计划发送时间 TIMESTAMPTZ（Cron 触发；daily=08:30 用户时区）",
        "sent_at":       "实际发送时间；NULL=待发送",
        "created_at":    "创建时间（分区键）",
        "workspace_id":  "工作空间 FK（RLS 依据）",
    },
    # ── intake_channels ─────────────────────────────────
    "intake_channels": {
        "id":                "主键 ID",
        "workspace_id":      "工作空间 FK",
        "project_id":        "项目 FK；关联默认工作项类型与指派规则",
        "name":              "通道名称（客户门户标题）",
        "slug":              "公开 URL 路径段 slug（/intake/{slug}；限流 20/min/IP）",
        "is_public":         "门户公开开关: true=免登录可提交, false=需登录验证",
        "default_issue_type":"默认转正类型: requirement / defect（管理员可配置）",
        "auto_assign_rules": "自动分配规则 JSONB [{match:{keyword,tags}, assign_to:user_id}]；顺序匹配",
        "created_by":        "创建人 FK",
        "created_at":        "创建时间",
        "updated_at":        "修改时间（触发器自动维护）",
        "deleted_at":        "软删除时间戳",
    },
    # ── intake_issues ────────────────────────────────────
    "intake_issues": {
        "id":                 "主键 ID",
        "channel_id":         "收件箱通道 FK",
        "workspace_id":       "工作空间 FK",
        "tracking_id":        '提交回执编号（YD-IN-XXXX；提交后邮件通知；/track/{id} 跟踪）',
        "status":             "处理状态: open(待审核) / accepted(已转正) / rejected(已拒绝) / archived(暂存)",
        "submitter_name":     "提交者姓名（脱敏显示）",
        "submitter_email":    "提交者邮箱（邮件校验码找回进展）",
        "subject":            "提交标题",
        "description":        "提交描述（纯文本；不支持富文本/附件）",
        "converted_issue_id": "转正后工作项 FK（issues.id；创建时复制标题/描述/附件）",
        "converted_at":       "转正时间（管理员审核通过时）",
        "rejection_reason":   "拒绝原因（填写后通知提交者）",
        "created_by":         "提交人 FK（可 NULL；匿名提交）",
        "created_at":         "提交时间",
        "updated_at":         "修改时间（触发器自动维护）",
        "deleted_at":         "软删除时间戳",
    },
    # ── automation_rules ─────────────────────────────────
    "automation_rules": {
        "id":                "主键 ID",
        "workspace_id":      "工作空间 FK",
        "project_id":        "项目 FK（NULL=工作空间级通用规则）",
        "name":              '规则名称（如"缺陷修复自动指派验证人"）',
        "enabled":           "启用开关: true=生效, false=暂停（连续失败 3 次自动禁用）",
        "trigger":           "触发器 JSONB {type: issue.status_changed, filter:{type_code, to_group}}",
        "conditions":        "条件矩阵 JSONB {all/any: [{field, op, value}]}；纯函数无 IO",
        "actions":           "动作列表 JSONB [{type: assign/transition/notify/create_issue, ...}]；最多 10 个顺序执行",
        "last_executed_at":  "最后执行时间（监控规则活跃度）",
        "failure_count":      "连续失败计数；>=3 触发熔断 + 通知 admin",
        "execution_count":    "累计执行成功次数（效能统计）",
        "created_by":        "创建人 FK（automation.failed 通知的默认接收人 + 项目 admin）",
        "created_at":        "创建时间",
        "updated_at":        "修改时间（触发器自动维护）",
        "deleted_at":        "软删除时间戳",
    },
    # ── automation_templates ─────────────────────────────
    "automation_templates": {
        "id":           "主键 ID",
        "name":         "内置模板名称（中文/英文双语文档链接）",
        "description":  '模板功能说明（如"子项全完成 → 父项自动完成"）',
        "category":     "模板分类: quality/issue_management/sprint/version/intake/assignment",
        "template_json":"模板 rule JSON（与 automation_rules 同结构；创建项目时批量复制）",
        "sort_order":   "排序权重（预设模板安装时按顺序展示）",
        "is_active":    "模板是否启用；false=不在可用列表",
        "created_at":   "创建时间",
        "updated_at":   "修改时间",
    },
    # ── rule_executions ──────────────────────────────────
    "rule_executions": {
        "id":            "主键 ID",
        "rule_id":       "规则 FK（automation_rules.id）",
        "issue_id":      "关联工作项 FK（可 NULL；非 issue 触发器则为 NULL）",
        "status":        "执行状态: success / failure / skipped / timeout",
        "duration_ms":   "执行耗时毫秒（含条件求值 + 所有 Action 总时长）",
        "results":       "执行明细 JSONB [{action_type, status, error?}]；success 时记录变更后值",
        "error_message": "失败错误信息（failure 时必填；用于排障）",
        "created_at":    "执行时间 TIMESTAMPTZ（按月分区键；30 天 TTL drop partition）",
        "workspace_id":  "工作空间 FK（RLS 依据）",
    },
    # ── dashboard_widgets ────────────────────────────────
    "dashboard_widgets": {
        "id":            "主键 ID",
        "dashboard_id":  "仪表盘 FK（dashboard_configs.id；可选，可存个人临时配置）",
        "type":          "卡片类型: project_overview / sprint_burndown / version_progress / module_distribution / quality_indicators / risk_alerts / resource_load / velocity_trend / dora_summary",
        "config":        "卡片个性化配置 JSONB（时间范围/项目/迭代/对比维度）",
        "position_x":   "布局横坐标（grid 列；0 起始）",
        "position_y":   "布局纵坐标（grid 行；0 起始）",
        "width":         "卡片宽度（grid 单元数 1-4）",
        "height":        "卡片高度（grid 单元数 1-4）",
        "refresh_interval_s":"刷新间隔秒数（实时 30s；缓存 dash:{dashboard}:widget:{id} TTL）",
        "data_source":   "数据源标识符（前端渲染器 ↔ 后端接口 1:1 映射）",
        "workspace_id":  "工作空间 FK（RLS 依据）",
        "created_at":    "创建时间",
        "updated_at":    "修改时间",
    },
    # ── dashboard_templates ──────────────────────────────
    "dashboard_templates": {
        "id":           "主键 ID",
        "name":         '模板名称（如"工程效能"/"QA 质量"/"PMO 战略多项目"）',
        "description":  "模板适用场景说明",
        "scope":        "作用域: project(单项目) / workspace(多项目聚合)",
        "layout":       "预设布局 JSONB（卡片数组 [{type, x, y, w, h, config}]；gridstack.js 格式）",
        "is_system":    "是否内置模板: true=系统预设, false=用户自定义（可分享）",
        "created_by":   "创建人 FK（is_system=true 时为 NULL）",
        "created_at":   "创建时间",
        "updated_at":   "修改时间",
    },
    # ── dashboard_snapshots ──────────────────────────────
    "dashboard_snapshots": {
        "id":           "主键 ID",
        "dashboard_id": "仪表盘 FK",
        "snapshot_time":"快照时间（大屏分享链接/定时导出触发）",
        "data":         "完整快照 JSONB（所有卡片数据缓存；大屏只读展示）",
        "share_token":  "分享令牌 SHA-256（NULL=非分享；过期/吊销后清空）",
        "expires_at":   "分享有效期 TIMESTAMPTZ；到期后快照URL 404",
        "created_by":   "创建人 FK",
        "created_at":   "创建时间",
        "workspace_id": "工作空间 FK（RLS 依据）",
    },
    # ── workbench_configs ────────────────────────────────
    "workbench_configs": {
        "id":           "主键 ID",
        "user_id":      "用户 FK（per-user 工作台；每个用户一行）",
        "workspace_id": "工作空间 FK（RLS 依据）",
        "layout":       "工作台布局 JSONB [{widget_type, x, y, w, h, config, is_pinned}]",
        "pinned_projects":"置顶项目 ID 数组 BIGINT[]；快速访问入口",
        "recent_views": "最近使用视图 ID 数组（用于视图选择器默认值）",
        "preferences":  "工作台偏好 JSONB（主题/快捷手势/默认看板）",
        "created_at":   "创建时间",
        "updated_at":   "修改时间（触发器自动维护）",
    },
    # ── workbench_templates ──────────────────────────────
    "workbench_templates": {
        "id":           "主键 ID",
        "name":         '模板名称（如"PM 视角"/"开发者视角"/"QA 视角"）',
        "description":  "模板适用角色说明",
        "layout":       "预设布局 JSONB（widgets 数组；新账号注册时复制为默认配置）",
        "is_system":    "是否内置模板",
        "sort_order":   "排序权重",
        "created_at":   "创建时间",
        "updated_at":   "修改时间",
    },
    # ── view_preferences ─────────────────────────────────
    "view_preferences": {
        "id":           "主键 ID",
        "user_id":      "用户 FK",
        "workspace_id": "工作空间 FK（RLS 依据）",
        "project_id":   "项目 FK（NULL=工作空间级默认视图偏好）",
        "view_type":    "视图类型: list / board / calendar / gantt",
        "preferences":  "偏好 JSONB {sort_by, sort_order, filters, group_by, field_visibility}；upsert key (user, project, view)",
        "created_at":   "创建时间",
        "updated_at":   "修改时间（触发器自动维护）",
    },
    # ── recent_items ─────────────────────────────────────
    "recent_items": {
        "id":           "主键 ID",
        "user_id":      "用户 FK",
        "workspace_id": "工作空间 FK（RLS 依据）",
        "project_id":   "所属项目 FK",
        "item_type":    "最近访问类型: issue / sprint / version / project / page",
        "item_id":      "关联 ID（BIGINT 统一）",
        "access_count": '访问次数；首页"最近"列表排序依据（加权访问时间 + 频次）',
        "created_at":   "首次访问时间",
        "updated_at":   "最后访问时间（触发器 trg_recent_items_touch 更新）",
    },
    # ── search_bookmarks ─────────────────────────────────
    "search_bookmarks": {
        "id":           "主键 ID",
        "user_id":      "用户 FK",
        "workspace_id": "工作空间 FK（RLS 依据）",
        "name":         '收藏查询名称（如"我的待办"）',
        "query":        'JQL 查询字符串（如"assignee:me status:todo"）；完整保存',
        "description":  "收藏说明（可选）",
        "is_shared":    "是否分享: true=同工作空间成员可见, false=仅自己",
        "sort_order":   "展示排序权重",
        "created_by":   "创建人 FK",
        "created_at":   "创建时间",
        "updated_at":   "修改时间",
    },
    # ── search_documents ─────────────────────────────────
    "search_documents": {
        "id":           "主键 ID",
        "workspace_id": "工作空间 FK",
        "project_id":   "项目 FK",
        "item_type":    "索引对象类型: issue / page / doc",
        "item_id":       "关联 ID（BIGINT）",
        "title":        "索引标题",
        "content":      "索引内容（纯文本；tsvector 计算源）",
        "metadata":     "索引元数据 JSONB（attribution/comments/sprint/labels 等）",
        "indexed_at":   "索引时间；对账 Job 检测 updated_at > indexed_at 重做",
        "created_at":   "创建时间",
        "updated_at":   "修改时间",
    },
    # ── search_history ───────────────────────────────────
    "search_history": {
        "id":           "主键 ID",
        "user_id":      "用户 FK",
        "workspace_id": "工作空间 FK",
        "query":        "完整搜索字符串（含 JQL 语法）",
        "result_count": "返回结果数（缓存用于效果分析）",
        "clicked_item_id":"点击第一条结果（优化搜索算法依据）",
        "searched_at":  "搜索时间 TIMESTAMPTZ；自动补全历史数据源",
    },
    # ── risk_rules ───────────────────────────────────────
    "risk_rules": {
        "id":             "主键 ID",
        "workspace_id":   "工作空间 FK",
        "project_id":     "项目 FK（NULL=工作空间级通用规则）",
        "name":           "规则名称",
        "metric":         "监控指标: overdue_days / blocked_days / dependency_chain / sprint_deviation",
        "operator":       "比较运算符: > / >= / < / <= / ==",
        "threshold":      "阈值（数值；如逾期 3 天、燃尽偏差 20%）",
        "severity":       "告警严重度: info / warning / critical",
        "is_active":      "启用开关: true=生效, false=暂停",
        "notification_channels": "通知渠道 JSONB [{channel, target}]；默认项目 admin + 规则创建者",
        "created_by":     "创建人 FK",
        "created_at":     "创建时间",
        "updated_at":     "修改时间（触发器自动维护）",
        "deleted_at":     "软删除时间戳",
    },
    # ── risk_alerts ──────────────────────────────────────
    "risk_alerts": {
        "id":           "主键 ID",
        "rule_id":      "触发规则 FK",
        "workspace_id": "工作空间 FK",
        "project_id":   "项目 FK",
        "issue_id":     "关联工作项 FK（可 NULL；如 sprint 级别告警）",
        "severity":     "告警严重度: info / warning / critical",
        "message":      "告警描述（含触发值/阈值对比）",
        "snapshot":     "告警快照 JSONB（当时状态: 迭代进度/关键路径/剩余工时等）",
        "status":       "处理状态: open(待处理) / acknowledged(已确认) / resolved(已解决) / dismissed(已忽略)",
        "resolved_by":  "处理人 FK（status=resolved 时必填）",
        "resolved_at":  "处理时间",
        "created_at":   "触发时间 TIMESTAMPTZ",
        "workspace_id2":"RLS 依据（若有）",
    },
    # ── metric_snapshots ─────────────────────────────────
    "metric_snapshots": {
        "id":            "主键 ID",
        "workspace_id":  "工作空间 FK（分区键 + RLS 依据）",
        "project_id":    "项目 FK（NULL=跨项目工作空间级聚合）",
        "granularity":   "聚合粒度: daily(日) / sprint(迭代) / version(版本)",
        "ref_id":        "粒度引用 ID（granularity=sprint→sprint_id；version→version_id；daily→NULL）",
        "metric":        "指标名: lead_time / throughput / wip_count / bug_density / escape_rate / velocity / dora_df / dora_lt / dora_cfr / dora_mttr / flow_efficiency",
        "value":         "指标值 NUMERIC；存储精确计算结果（查询时直接使用，避免重复聚合）",
        "dimensions":     "维度 JSONB {type_code, state_group, module_id, assignee_id}；钻取过滤",
        "snapshot_date": "快照日期（幂等 upsert key: (granularity, ref_id, metric, snapshot_date)）",
        "created_at":    "写入时间（Cron 每日 01:30 聚合 Job）",
        "updated_at":    "修改时间（触发器自动维护）",
    },
    # ── metric_adjustments ───────────────────────────────
    "metric_adjustments": {
        "id":           "主键 ID",
        "snapshot_id":  "原始快照 FK（metric_snapshots.id）",
        "workspace_id": "工作空间 FK",
        "adjusted_by":  "修正人 FK（admin 角色才可操作）",
        "original_value":"修正前原值（不可变；审计依据）",
        "adjusted_value":"修正后值（查询时叠加；不覆盖原值）",
        "reason":       '修正原因（必填；如"剔除异常测试数据"）',
        "created_at":   "修正时间",
    },
    # ── webhooks ─────────────────────────────────────────
    "webhooks": {
        "id":              "主键 ID",
        "workspace_id":    "工作空间 FK（RLS 依据）",
        "project_id":      "项目 FK（决定事件触发的运行上下文）",
"name":            "Webhook 配置名称",
        "target_url":      "投递目标 URL（SSRF 防护: 协议白名单 https 优先；解析后 IP 非内网）",
        "secret":          "HMAC-SHA256 密钥（X-Ydsz-Signature-256 签名头）；仅存 SHA-256 hash",
        "events":          '事件白名单 JSONB ["issue.created", "issue.status_changed", ...]；空数组=全部禁用',
        "is_active":       "启用开关: true=生效, false=暂停",
        "ssrf_whitelist":  "出站 IP 白名单（空=使用默认安全防护；显式声明例外 IP/CIDR）",
        "last_triggered_at":"最后触发时间（监控活跃度；不活跃可告警）",
        "last_error":      "最后错误信息（用于排障；连续失败通知创建者）",
        "failure_count":   "连续失败次数；>=5 自动 unhealthy + 通知 admin",
        "created_by":      "创建人 FK（Webhook 失败默认通知人）",
        "created_at":      "创建时间",
        "updated_at":      "修改时间（触发器自动维护）",
        "deleted_at":      "软删除时间戳",
    },
    # ── webhook_logs ─────────────────────────────────────
    "webhook_logs": {
        "id":              "主键 ID",
        "webhook_id":      "Webhook 配置 FK",
        "event_type":      "事件类型（与 domain_events.event_type 对齐）",
        "delivery_id":     "投递唯一 UUID（X-Ydsz-Delivery 头；接收方幂等）",
        "target_url":      "投递目标（当时快照；配置修改后仍展示原值）",
        "request_body":    "POST body（JSON；截断 >10KB 时存 attachment）",
        "response_status": "接收方 HTTP status code；5xx/429/超时(10s) 触发重试",
        "response_body":   "接收方响应体（截断 >1KB）；便于排障",
        "duration_ms":     "投递耗时毫秒",
        "attempt_number":  "本次重试次数（1=首次；最大 3 次）",
        "status":          "投递状态: success / failed / retrying",
        "error_message":   "错误信息（失败时）",
        "created_at":      "投递时间 TIMESTAMPTZ（按月分区 + 30 天 TTL）",
        "workspace_id":    "工作空间 FK（RLS 依据）",
    },
    # ── deployment_events ───────────────────────────────
    "deployment_events": {
        "id":            "主键 ID",
        "workspace_id":  "工作空间 FK（RLS 依据）",
        "project_id":    "项目 FK",
        "environment":   "部署环境: dev / staging / production",
        "status":        "部署状态: success / failed",
        "commit_sha":    "部署 commit SHA（精确到 commit；用于关联工作项/PR）",
        "started_at":    "部署开始时间 TIMESTAMPTZ；started_at→deployed_at = lead time",
        "deployed_at":   "部署成功时间 TIMESTAMPTZ；DORA 计算数据源",
        "source":        "触发来源: ci_cd / webhook / manual；标识 CI 流水线",
        "meta":          "元数据 JSONB（workflow_id / branch / tags / runner 等）",
        "created_by":    "操作人 FK",
        "created_at":    "注册时间（Webhook POST /hooks/deployments 入口）",
    },
    # ── domain_events ────────────────────────────────────
    "domain_events": {
        "id":             "主键 ID",
        "workspace_id":   "工作空间 FK（RLS 依据）",
        "aggregate_type": "聚合类型: issue / sprint / version / automation / deployment",
        "aggregate_id":   "聚合 ID（BIGINT 统一）",
        "event_type":     "事件类型: issue.status_changed；sprint.started；automation.executed ...",
        "payload":        "事件载荷 JSONB（当时完整状态快照；消费者解析依据）",
        "occurred_at":    "事件发生时间（与事务提交时间对齐；Outbox 写入时间）",
        "published_at":   "投递时间（worker 处理后置；NULL=待投递；WHERE published_at IS NULL 发布）",
        "created_at":     "写入时间（事务内与业务表同事务；唯一事实源）",
    },
    # ── idempotency_keys ─────────────────────────────────
    "idempotency_keys": {
        "id":           "主键 ID",
        "key":          "幂等键（客户端生成 UUID；API 请求头 X-Idempotency-Key；unique）",
        "user_id":      "用户 FK（校验请求者身份）",
        "response":     "原响应缓存 JSONB（同 key 重复请求直接返回原 response；省 DB 调用）",
        "request_hash": "请求体 SHA-256 摘要（可选；防 request body 修改后复用旧响应）",
        "created_at":   "首次请求时间（过期清理窗口 ≥24h）",
        "expires_at":   "过期时间 TIMESTAMPTZ；过期后同 key 可重放",
    },
    # ── password_reset_tokens ─────────────────────────────
    "password_reset_tokens": {
        "id":           "主键 ID",
        "user_id":      "用户 FK",
        "token_hash":   "一次性令牌 SHA-256 hash（邮件链接中使用 token 本身）",
        "expires_at":   "有效期 TIMESTAMPTZ（默认 1 小时；过期即失效）",
        "used_at":      "首个使用时间（标志已消费；一次性原则）",
        "created_at":   "创建时间",
    },
    # ── audit_logs ───────────────────────────────────────
    "audit_logs": {
        "id":           "主键 ID",
        "workspace_id": "工作空间 FK（RLS 依据）",
        "actor_id":     "操作人 FK（users.id）；系统动作为 NULL",
        "action":       "操作名称（固定枚举: login / logout / permission_change / member_add / member_remove / token_revoke / webhook.create / data_export / setting.update / issue.delete）",
        "target_type":  "目标类型: workspace / project / issue / user / member / token / webhook",
        "target_id":    "目标 ID（BIGINT）",
        "detail":       "变更前→后 diff JSONB（字段级 audit；如 {role:'member'→'admin'}）",
        "ip_address":   "客户端 IP（IPv4/IPv6；安全监控/异常登录识别）",
        "user_agent":   "客户端 UA 字符串（浏览器/设备识别）",
        "request_id":   "关联请求 ID（trace_id；日志/错误/链路追踪关联）",
        "created_at":   "操作时间 TIMESTAMPTZ（在线 12 个月 + 归档 3 年；只增不改）",
    },
    # ── api_tokens ───────────────────────────────────────
    "api_tokens": {
        "id":              "主键 ID",
        "user_id":         "用户 FK",
        "name":            'Token 自定义名称（方便管理/撤销；如"CI Pipeline"）',
        "token_hash":      "SHA-256 hash 值（ydz_ 前缀明文仅展示一次；存入 hash 校验）",
        "scopes":          '权限范围 JSONB ["read:issues", "write:issues", "admin:*"]；白名单',
        "expires_at":      "过期时间 TIMESTAMPTZ（NULL=不过期；推荐设置 ≤90d）",
        "last_used_at":    "最后使用时间（活跃度审计；长期未用可告警/建议吊销）",
        "created_by":      "创建人 FK；管理与撤销监听",
        "created_at":      "创建时间",
        "revoked_at":      "吊销时间；NULL=有效；revoked_at 非空校验时拒绝",
        "deleted_at":      "软删除时间戳",
    },
    # ── schema_migrations ─────────────────────────────────
    "schema_migrations": {
        "version":     "迁移版本编号（NNNN；递增；CI 校验连续性）",
        "applied_at":  "应用时间（迁移框架自动写入；幂等检测依据）",
        "description": "迁移描述（便于排查 + CHANGELOG 对齐）",
        "execution_ms":"执行耗时毫秒（用于慢迁移告警）",
        "checksum":    "迁移内容 SHA-256（防篡改；运行时校验与声明的 checksum 匹配）",
    },
    # ── project_sequences ─────────────────────────────────
    "project_sequences": {
        "project_id":  "项目 FK（PRIMARY KEY）",
        "next_value":  "当前已发号 + 1（原子自增: SET next_value = next_value + 1 RETURNING next_value - 1）",
        "created_at":  "创建时间",
        "updated_at":  "修改时间（触发器自动维护）",
    },
    # ── estimate_points ───────────────────────────────────
    "estimate_points": {
        "id":           "主键 ID",
        "workspace_id": "工作空间 FK",
        "project_id":   "项目 FK（NULL=工作空间级默认；覆盖用项目级）",
        "name":         '估算点名称（如"1"、"2"、"3"、"5"、"8"）',
        "description":  "估算含义 / 适用规模（可选；Hints 浮层展示）",
        "sequence":     "排序权重（升序；配置列表展示用）",
        "is_default":   "是否默认: true=项目创建时是否包含在标准斐波那契集",
        "created_at":   "创建时间",
        "updated_at":   "修改时间",
        "deleted_at":   "软删除时间戳",
    },
    # ── issue_templates ──────────────────────────────────
    "issue_templates": {
        "id":              "主键 ID",
        "workspace_id":    "工作空间 FK",
        "project_id":      "项目 FK（NULL=工作空间级模板）",
        "name":            '模板名称（如"前端 Bug 模板"/"需求 Brief 模板"）',
        "type_code":       "默认类型: requirement / task / defect",
        "template_json":   "预填充字段 JSONB {title, description_json, state_id, priority, labels, assignees}",
        "is_active":       "启用开关: true=出现在模板选择器",
        "created_by":      "创建人 FK",
        "created_at":      "创建时间",
        "updated_at":      "修改时间",
        "deleted_at":      "软删除时间戳",
    },
    # ── pages ────────────────────────────────────────────
    "pages": {
        "id":           "主键 ID",
        "workspace_id": "工作空间 FK（RLS 依据）",
        "project_id":   "所属项目 FK",
        "title":        "页面标题",
        "content_json": "TipTap 编辑器内容 JSON（ProseMirror 格式；富文本 + 嵌入卡片）",
        "content_html": "从 content_json 渲染的 HTML（展示层）",
        "cover_image":  "封面图片附件 ID",
        "icon":         "页面图标（Emoji / Lucide 图标）",
        "parent_id":    "父页面 FK（pages.id）；层级≤3",
        "sort_order":   "同级排序权重",
        "is_pinned":    "是否置顶: true=在侧边栏始终展示",
        "created_by":   "创建人 FK",
        "created_at":   "创建时间",
        "updated_at":   "修改时间（触发器自动维护）",
        "deleted_at":   "软删除时间戳",
    },
}

# ─────────────────────────────────────────────────────────────────────
# 三、触发器注释  { trigger_name → comment }
# ─────────────────────────────────────────────────────────────────────
TRIGGER_COMMENTS = {
    # ── updated_at 自动维护 ──────────────────────────
    "trg_issues_updated_at":         "issues: BEFORE UPDATE 自动将 updated_at 更新为 now()；事件监听同步 ES",
    "trg_labels_updated_at":         "labels: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_modules_updated_at":        "modules: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_projects_updated_at":       "projects: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_sprints_updated_at":        "sprints: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_versions_updated_at":       "versions: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_users_updated_at":          "users: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_workspaces_updated_at":     "workspaces: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_invitations_updated_at":    "invitations: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_workspace_members_updated_at": "workspace_members: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_project_members_updated_at": "project_members: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_states_updated_at":         "states: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_state_transitions_updated_at": "state_transitions: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_issue_labels_updated_at":   "issue_labels: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_issue_modules_updated_at":  "issue_modules: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_issue_assignees_updated_at":"issue_assignees: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_issue_watchers_updated_at": "issue_watchers: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_issue_activities_updated_at": "issue_activities: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_issue_dependencies_updated_at": "issue_dependencies: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_issue_relations_updated_at":"issue_relations: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_time_logs_updated_at":       "time_logs: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_attachments_updated_at":     "attachments: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_notifications_updated_at":   "notifications: BEFORE UPDATE 自动维护 updated_at（如 read_at 更新触发列表刷新）",
    "trg_notification_preferences_updated_at": "notification_preferences: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_intake_channels_updated_at":"intake_channels: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_intake_issues_updated_at":  "intake_issues: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_automation_rules_updated_at":"automation_rules: BEFORE UPDATE 自动维护 updated_at + 触发条件缓存失效",
    "trg_rule_executions_updated_at": "rule_executions: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_dashboard_widgets_updated_at": "dashboard_widgets: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_workbench_configs_updated_at": "workbench_configs: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_view_preferences_updated_at": "view_preferences: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_recent_items_updated_at":   "recent_items: BEFORE UPDATE 重置 updated_at = now()",
    "trg_risk_rules_updated_at":     "risk_rules: BEFORE UPDATE 自动维护 updated_at",
    "trg_risk_alerts_updated_at":    "risk_alerts: BEFORE UPDATE 自动维护 updated_at + 通知冷却",
    "trg_metric_snapshots_updated_at":"metric_snapshots: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_webhooks_updated_at":       "webhooks: BEFORE UPDATE 自动维护 updated_at + 触发条件缓存失效",
    "trg_webhook_logs_updated_at":   "webhook_logs: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_deployment_events_updated_at": "deployment_events: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_search_bookmarks_updated_at": "search_bookmarks: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_estimate_points_updated_at": "estimate_points: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_pages_updated_at":          "pages: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_issue_templates_updated_at":"issue_templates: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_version_delivery_snapshots_updated_at": "version_delivery_snapshots: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_notification_deliveries_updated_at": "notification_deliveries: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_notification_digests_updated_at": "notification_digests: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_search_documents_updated_at": "search_documents: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_search_history_updated_at":  "search_history: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_dashboard_templates_updated_at": "dashboard_templates: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_sprint_issues_updated_at":   "sprint_issues: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_sprint_snapshots_updated_at": "sprint_snapshots: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_users_updated_at":          "users: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    "trg_project_sequences_updated_at": "project_sequences: BEFORE UPDATE 自动将 updated_at 更新为 now()",
    # ── 乐观锁自增版本号 ─────────────────────────────
    "trg_issues_bump_version":       "issues: BEFORE UPDATE 乐观锁版本号 version = version + 1；UPDATE 条件含旧 version",
    "trg_sprints_bump_version":      "sprints: BEFORE UPDATE 乐观锁版本号 version = version + 1",
    "trg_versions_bump_version":     "versions: BEFORE UPDATE 乐观锁版本号 version = version + 1",
    "trg_projects_bump_version":     "projects: BEFORE UPDATE 乐观锁版本号 version = version + 1",
    # ── 搜索索引同步 / 清理 ─────────────────────────
    "trg_issue_search_sync":         "issues: AFTER INSERT/UPDATE 异步同步 ES 索引 `ydsz_issues`；routing=workspace_id",
    "trg_issue_search_cleanup":      "issues: AFTER DELETE（软删除）异步清理 ES 索引对应文档",
    "trg_sprint_search_sync":        "sprints: AFTER INSERT/UPDATE 同步 ES 降级索引 search_documents",
    "trg_sprint_search_cleanup":     "sprints: AFTER DELETE 清理 ES 及 search_documents 中对应文档",
    "trg_version_search_sync":       "versions: AFTER INSERT/UPDATE 同步 ES 降级索引",
    "trg_version_search_cleanup":    "versions: AFTER DELETE（软删除）清理 ES 及 search_documents",
    "trg_page_search_sync":          "pages: AFTER INSERT/UPDATE 同步 ES 降级索引 `ydsz_pages`",
    "trg_page_search_cleanup":       "pages: AFTER DELETE 清理 pages ES 索引文档",
    # ── recent_items_touch ───────────────────────────
    "trg_recent_items_touch":        'recent_items: BEFORE UPDATE 每次访问重置 updated_at = now()；服务于"最近访问"列表排序',
    # ── 工时差值回写 ──────────────────────────────────
    "trg_time_logs_effort_sync":     "time_logs: AFTER INSERT/UPDATE/DELETE 差值重算 issues.actual_effort / remaining_effort",
    # ── 进度回写 ──────────────────────────────────────
    "trg_issues_progress_rollup":    "issues: AFTER UPDATE (state_id) 子项 completed 时进度回写：Σ完成子项点/Σ全部子项点",
    # ── 迭代快照 ──────────────────────────────────────
    "trg_sprints_snapshot_sync":     "sprints: AFTER INSERT/UPDATE 触发/失效 sprint_snapshots 缓存",
    # ── 附件清理 ──────────────────────────────────────
    "trg_attachments_cleanup":       "attachments: AFTER DELETE + 定时任务 清理 MinIO 孤立文件",
    # ── 缺陷必填字段强制 ─────────────────────────────
    "trg_issues_defect_required":    "issues: BEFORE INSERT/UPDATE 校验 type_code=defect 时 severity & found_phase 非空",
}

# ─────────────────────────────────────────────────────────────────────
# 四、索引注释  { index_name → comment }
# ─────────────────────────────────────────────────────────────────────
INDEX_COMMENTS = {
    # ── activities / issue_activities ─────────────────────
    "idx_activities_issue":              "按工作项+时间倒序（详情页活动时间线）",
    "idx_activities_issue_covering":     "覆盖索引：按 work activity 高频列表查询",
    "idx_activities_project":            "按工作空间+项目查询活动日志",
    # ── api_tokens ──────────────────────────────────────
    "idx_api_tokens_user":               "按用户查拥有的 Token（Token 管理页/撤销确认）",
    # ── attachments ─────────────────────────────────────
    "idx_attachments_entity":            "多态查询: type+id 定位附件列表（工作项详情/评论附件）",
    "idx_attachments_uploader":          "按上传人查附件（配额统计/回收站）",
    "idx_attachments_workspace":         "按工作空间查全部附件（配额/存储用量统计）",
    # ── audit_logs ──────────────────────────────────────
    "idx_audit_logs_action_target":      "按操作类型+目标对象查询审计日志",
    "idx_audit_logs_ws_time":            "按工作空间+时间倒序游标分页（安全审计查询）",
    # ── automation_rules ────────────────────────────────
    "idx_automation_rules_project_status": "按项目+状态查规则列表",
    "idx_automation_rules_sort":         "按项目+排序权重查规则（执行优先级）",
    "idx_automation_rules_trigger":      "按项目+trigger 类型快速匹配规则（事件分发索引）",
    "idx_automation_rules_ws":           "按工作空间查规则列表",
    # ── automation_templates ────────────────────────────
    "idx_automation_templates_cat":      "按分类查模板列表",
    # ── dashboard_snapshots ─────────────────────────────
    "idx_dashboard_snapshots_project":   "按项目查快照列表",
    "idx_dashboard_snapshots_refreshed": "按刷新时间查询快照（调度器轮询）",
    # ── dashboard_templates ─────────────────────────────
    "idx_dashboard_templates_category":  "按分类查仪表盘模板列表",
    # ── dashboard_widgets ───────────────────────────────
    "idx_dashboard_widgets_project":     "按项目查 Widget 配置",
    "idx_dashboard_widgets_user":        "按用户查询 Widget 个性化配置",
    # ── notification_deliveries ─────────────────────────
    "idx_deliveries_next_retry":         "按下次重试时间查待投递记录（调度器轮询）",
    "idx_deliveries_notification":       "按通知记录查投递明细",
    "idx_deliveries_status":             "按投递状态查询（统计/监控）",
    # ── deployment_events ───────────────────────────────
    "idx_deployment_events_project":     "按项目+时间倒序（部署历史）",
    "idx_deployment_events_ws":          "按工作空间查部署事件",
    # ── notification_digests ────────────────────────────
    "idx_digests_pending":               "按状态查待生成摘要（调度器轮询）",
    # ── domain_events ───────────────────────────────────
    "idx_events_unpublished":            "WHERE published_at IS NULL 未发布事件投递索引（Outbox reader）",
    # ── intake_channels ─────────────────────────────────
    "idx_intake_channels_project":       "按项目查收件箱频道列表",
    "idx_intake_channels_slug":          "按 slug 公开门户路由（/intake/{slug} 校验 active）",
    "idx_intake_channels_workspace":     "按工作空间查频道列表",
    # ── intake_issues ───────────────────────────────────
    "idx_intake_issues_channel":         "按频道查收件箱工单列表",
    "idx_intake_issues_status":          "按状态过滤收件箱工单",
    "idx_intake_issues_submitter":       "按提交人查询收件箱工单",
    "idx_intake_issues_tracking":        "按 tracking_id YD-IN-XXXX 查询提交状态（用户跟踪页）",
    "idx_intake_issues_workspace":       "按工作空间查收件箱工单列表",
    # ── invitations ─────────────────────────────────────
    "idx_invitations_email":             "按邮箱查邀请记录（重复邀请校验）",
    "idx_invitations_token":             "按 token 查询邀请详情（注册流程）",
    "idx_invitations_workspace":         "按工作空间查邀请列表",
    # ── issue_assignees ─────────────────────────────────
    "idx_issue_assignees_covering":      "覆盖索引：按工作项+用户查指派（含常用字段）",
    "idx_issue_assignees_user":          '按 user_id 查询"我的指派"（待办列表高频场景）',
    # ── issue_comments ──────────────────────────────────
    "idx_issue_comments_author":         "按作者查评论列表",
    "idx_issue_comments_issue":          "按工作项查评论列表",
    # ── issue_dependencies ──────────────────────────────
    "idx_issue_deps_pred":               "按前驱工作项查依赖关系",
    "idx_issue_deps_succ":               "按后继工作项查依赖关系",
    # ── issue_relations ─────────────────────────────────
    "idx_issue_relations_source":        "按源工作项查关联关系",
    "idx_issue_relations_target":        "按目标工作项查关联关系",
    # ── issue_watchers ──────────────────────────────────
    "idx_issue_watchers_user":           "按 user_id 查关注的工作项列表",
    # ── issues ──────────────────────────────────────────
    "idx_issues_created":                "按工作空间+创建时间倒序（最近创建/活动日志查询）",
    "idx_issues_fix_version":            "按 fix_version 查询缺陷（版本修复追踪）",
    "idx_issues_found_version":          "按 found_version 查询缺陷（发现版本统计）",
    "idx_issues_list_covering":          "覆盖索引：列表视图高频查询场景",
    "idx_issues_parent":                 "按 parent_id 查询子项（WBS 树展开/进度回写）",
    "idx_issues_priority_covering":      "覆盖索引：按项目+优先级过滤（看板/列表视图）",
    "idx_issues_project_sequence":       "按项目+序号唯一（YD-123 展示编号查询）",
    "idx_issues_project_state":          "按工作空间+项目+状态查询（看板列表/过滤器高频场景；排除软删除）",
    "idx_issues_public_id":              "按 public_id 查询工作项（API/URL 路由）",
    "idx_issues_release_version":        "按 release_version 查询（版本交付范围）",
    "idx_issues_search_tsv":             "tsvector 索引（search_tsv 列 GIN；PostgreSQL 全文检索）",
    "idx_issues_state_covering":         "覆盖索引：按项目+状态查询看板列表",
    "idx_issues_target_date":            "按工作空间+项目+目标日期查询（未完成逾期提醒/甘特图）",
    "idx_issues_target_date_covering":   "覆盖索引：目标日期查询场景",
    "idx_issues_type":                   "按工作项类型过滤（需求/任务/缺陷列表视图切换）",
    "idx_issues_type_covering":          "覆盖索引：按类型过滤列表查询",
    "idx_issues_updated":                "按工作空间+更新时间倒序（最近活动/时间线查询）",
    "idx_issues_workspace_project":      "按工作空间+项目高频过滤查询",
    # ── labels ──────────────────────────────────────────
    "idx_labels_project":                "按工作空间+项目查标签列表（项目标签管理/选择器）",
    # ── metric_adjustments ──────────────────────────────
    "idx_metric_adjustments_ws":         "按工作空间查询指标调整记录",
    # ── metric_snapshots ────────────────────────────────
    "idx_metric_snap_covering":          "覆盖指标快照查询（granularity+ref_id+metric+metric_date）",
    "idx_metric_snap_date":              "按日期范围查询指标快照（趋势图）",
    "idx_metric_snap_lookup":            "按 (granularity, ref_id, metric, metric_date) 查询/覆盖快照",
    "idx_metric_snap_ws":                "按工作空间查询指标快照",
    "idx_metric_snapshots_project":      "按项目查快照列表",
    "idx_metric_snapshots_workspace":    "按工作空间查快照列表",
    # ── modules ─────────────────────────────────────────
    "idx_modules_project":               "按工作空间+项目查询模块列表",
    # ── notification_digests ────────────────────────────
    "idx_notification_digests_pending":  "按状态查待投递摘要队列",
    # ── notifications ───────────────────────────────────
    "idx_notifications_archived":        "按已读状态过滤通知列表（活跃通知筛选）",
    "idx_notifications_entity":          "按关联实体查通知记录",
    "idx_notifications_recipient_unread":"按接收人+未读过滤（站内信列表未读数；按时间倒序）",
    # ── sprints / sprint_issues / sprint_snapshots ──────
    "idx_one_active_sprint_per_project":  "每个项目仅一个 active 迭代（部分索引 WHERE active）",
    "idx_sprint_issues_issue":           "按 issue_id 反查所属迭代",
    "idx_sprint_snapshots_unique":       "按 sprint+时间查燃尽图快照序列",
    "idx_sprints_active_unique":         "active 迭代唯一性校验（部分索引）",
    "idx_sprints_project_status":        "按项目+迭代状态查询（active 唯一性校验 / 迭代列表）",
    "idx_sprints_version":               "按 version 反查关联迭代",
    "idx_sprintsnapshots_project":       "按项目查燃尽图快照列表",
    "idx_sprintsnapshots_sprint_date":   "按 sprint+日期查快照（趋势图）",
    # ── pages ───────────────────────────────────────────
    "idx_pages_parent":                  "按 parent_id 查询子页面（树展开）",
    "idx_pages_project":                 "按项目查知识页面列表",
    "idx_pages_project_sort":            "按项目+排序权重（拖拽排序）",
    "idx_pages_public_id":               "按 public_id 查询页面（API/URL 路由）",
    # ── password_reset_tokens ───────────────────────────
    "idx_password_reset_tokens_expires": "按过期时间查询 Token（定时清理）",
    "idx_password_reset_tokens_user_active": "按用户查有效 Token（只能有一个 active）",
    # ── projects ────────────────────────────────────────
    "idx_projects_template":             "按模板查询项目列表",
    "idx_projects_workspace":            "按工作空间查项目列表（含激活/归档状态过滤）",
    "idx_projects_workspace_identifier": "按工作空间+标识符唯一（URL 路由定位 YD 项目）",
    "idx_projects_workspace_slug":       "按工作空间+slug 唯一（URL 友好标识）",
    # ── recent_items ───────────────────────────────────
    "idx_recent_items_user":             "按用户+最后访问时间倒序（首页最近列表 Top N）",
    "idx_recent_items_ws":              "按工作空间查最近访问记录",
    # ── risk_alerts ────────────────────────────────────
    "idx_risk_alerts_project":           "按项目查风险告警列表",
    "idx_risk_alerts_unresolved":        "未解决风险告警查询（Dashboard widget 视图）",
    "idx_risk_alerts_workspace":         "按工作空间查风险告警列表",
    # ── risk_rules ─────────────────────────────────────
    "idx_risk_rules_active":             "按启用状态查规则列表",
    "idx_risk_rules_project":            "按项目查规则列表",
    "idx_risk_rules_workspace":          "按工作空间查规则列表",
    # ── rule_executions ────────────────────────────────
    "idx_rule_executions_event":         "按事件反查规则执行记录",
    "idx_rule_executions_project":       "按项目查规则执行列表",
    "idx_rule_executions_rule":          "按规则反查执行历史（调试/监控）",
    "idx_rule_executions_ws":            "按工作空间查规则执行列表",
    # ── search_bookmarks ───────────────────────────────
    "idx_search_bookmarks_project":      "按项目查搜索收藏",
    "idx_search_bookmarks_user":         "按用户查收藏列表（搜索收藏管理页）",
    "idx_search_bookmarks_ws":           "按工作空间查搜索收藏",
    # ── search_documents ───────────────────────────────
    "idx_search_documents_project":      "按项目查搜索文档列表",
    "idx_search_documents_tsv":          "GIN 全文索引（search_tsv 列；ES 不可用时兜底）",
    "idx_search_documents_unique":       "按文档唯一性约束检索",
    "idx_search_documents_ws":           "按工作空间查搜索文档",
    # ── search_history ─────────────────────────────────
    "idx_search_history_user":           "按用户查搜索历史（最近搜索自动补全）",
    "idx_search_history_ws_user":        "按工作空间+用户查搜索历史",
    # ── states / state_transitions ─────────────────────
    "idx_state_transitions_lookup":      "(project, type, from_state, to_state) 唯一；防重复定义流转",
    "idx_states_project":                "按项目+类型查状态集列表（不同类型独立状态模板）",
    # ── time_logs ──────────────────────────────────────
    "idx_time_logs_issue":               "按工作项查工时明细（详情页时间线 / sum 重算 actual_effort）",
    "idx_time_logs_user_date":           "按用户+日期查工时（成员工时报表 / 负载热力图）",
    # ── version_delivery_snapshots ─────────────────────
    "idx_vds_version":                   "按版本查交付范围快照",
    "idx_vds_workspace":                 "按工作空间查快照列表",
    # ── versions ───────────────────────────────────────
    "idx_versions_project_status":       "按项目+状态查询版本列表（planning/active/released/archived）",
    "idx_versions_unique_semver":        "按项目+semver 发布用（唯一；发布后只读）",
    "idx_versions_workspace":            "按工作空间查版本列表",
    # ── view_preferences ───────────────────────────────
    "idx_view_prefs_user":               "按用户查视图偏好设置",
    # ── webhook_logs ───────────────────────────────────
    "idx_webhook_logs_delivery":         "按投递 ID 查询 Webhook 日志",
    "idx_webhook_logs_occurred":         "按发生时间倒序游标分页（管理页面）",
    "idx_webhook_logs_webhook":          "按 webhook+时间查投递日志（管理页面/监控面板）",
    "idx_webhook_logs_workspace":        "按工作空间查 Webhook 日志",
    # ── webhooks ───────────────────────────────────────
    "idx_webhooks_active":               "按项目+启用状态查 Webhook 配置（事件投递匹配）",
    "idx_webhooks_project":              "按项目查 Webhook 列表",
    "idx_webhooks_workspace":            "按工作空间查 Webhook 列表",
    # ── workbench_configs ──────────────────────────────
    "idx_workbench_configs_project":     "按项目查工作台配置",
    "idx_workbench_configs_user":        "按用户查个性化工作台配置",
    "idx_workbench_configs_user_project":"按用户+项目查工作台配置",
    # ── workbench_templates ────────────────────────────
    "idx_workbench_templates_default":   "按默认状态查询工作台模板",
    # ── 唯一约束（uq_ 前缀）─────────────────────────────
    "uq_deployment_events_idempotent":   "投递幂等键（同一事件+目标只投递一次）",
    "uq_workspaces_slug":                "slug 唯一（URL 路由依据；软删除排除）",
}


# ─────────────────────────────────────────────────────────────────────
# 解析器
# ─────────────────────────────────────────────────────────────────────
class Inserter:
    def __init__(self, path):
        self.path = path
        with open(path, "r", encoding="utf-8-sig") as f:
            self.raw = f.read()
        # 保留 BOM 标记
        self.had_bom = self.raw.startswith("﻿")
        self.lines = self.raw.splitlines()  # 去掉换行符方便处理
        self.insertions = []  # [(line_idx, indent, text_lines)]
        self.stats = {
            "table_comments_added": 0,
            "column_comments_added": 0,
            "trigger_comments_added": 0,
            "index_comments_added": 0,
            "table_comments_skipped": 0,
            "column_comments_skipped": 0,
            "trigger_comments_skipped": 0,
            "index_comments_skipped": 0,
        }

    def _exists_nearby(self, target_text, after_line, window=30):
        """检查 target_text 关键词 在 [after_line, after_line+window) 区间内是否已存在"""
        key = target_text[:40]  # 取前 40 字符作为指纹
        end = min(after_line + window, len(self.lines))
        for i in range(after_line, end):
            if i < len(self.lines) and key in self.lines[i]:
                return True
        return False

    def _unquote(self, ident):
        """去除 PostgreSQL 标识符的双引号: \"name\" → name"""
        m = re.match(r'^"([^"]+)"$', ident)
        return m.group(1) if m else ident

    def _find_table_name(self, create_table_line):
        """从 CREATE TABLE 行提取表名，兼容有/无双引号格式"""
        # 格式 1: CREATE TABLE "public"."name" (
        # 格式 2: CREATE TABLE public.name (
        # 格式 3: CREATE TABLE "name" (
        m = re.search(r'CREATE\s+TABLE\s+(?:"public"\.)?"(\w+)"', create_table_line, re.IGNORECASE)
        if m:
            return m.group(1)
        m = re.search(r'CREATE\s+TABLE\s+(?:public\.)?(\w+)', create_table_line, re.IGNORECASE)
        return m.group(1) if m else None

    def _find_trigger_name(self, trigger_lines):
        """从 CREATE TRIGGER 行提取触发器名，兼容有/无双引号格式"""
        for line in trigger_lines:
            m = re.search(r'CREATE\s+(?:OR\s+REPLACE\s+)?TRIGGER\s+"(\w+)"', line, re.IGNORECASE)
            if m:
                return m.group(1)
            m = re.search(r'CREATE\s+(?:OR\s+REPLACE\s+)?TRIGGER\s+(\w+)', line, re.IGNORECASE)
            if m:
                return m.group(1)
        return None

    def _find_index_name(self, index_line):
        """从 CREATE [UNIQUE] INDEX 行提取索引名，兼容有/无双引号格式"""
        m = re.match(r'CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?"(\w+)"\s+ON', index_line, re.IGNORECASE)
        if m:
            return m.group(1)
        m = re.match(r'CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)\s+ON', index_line, re.IGNORECASE)
        return m.group(1) if m else None

    def _find_index_table(self, index_line):
        """提取索引所在表名，兼容 \"public\".\"table\" 格式"""
        m = re.search(r'ON\s+"public"\."(\w+)"', index_line, re.IGNORECASE)
        if m:
            return m.group(1)
        m = re.search(r'ON\s+(?:public\.)?(\w+)\s*\(', index_line, re.IGNORECASE)
        return m.group(1) if m else None

    def _build_table_comment_block(self, table_name):
        """构建 COMMENT ON TABLE + COMMENT ON COLUMN 行列表"""
        out = []
        # 表注释
        tbl_comment = TABLE_COMMENTS.get(table_name)
        if tbl_comment:
            out.append(f"COMMENT ON TABLE public.{table_name} IS '{tbl_comment}';")

        # 列注释
        col_map = COLUMN_COMMENTS.get(table_name, {})
        for col, comment in col_map.items():
            out.append(f"COMMENT ON COLUMN public.{table_name}.{col} IS '{comment}';")
        return out

    def _find_create_table_end(self, start_idx):
        """
        从 start_idx（CREATE TABLE 行）开始，找到 DDL 结束行（即 `);` 所在行）。
        Navicat dump 主要用两种结束格式：
          格式 A: `);` 独占一行
          格式 B: `)` 独占一行 + `;` 独占下一行（如 api_tokens）
        对行尾带 ）; 但不在行首的情况（如约束末尾)也做兼容。
        """
        for i in range(start_idx + 1, len(self.lines)):
            stripped = self.lines[i].strip()
            # 模式 1: 标准 ); （独占一行）
            if stripped == ");":
                return i
            # 模式 1b: 独占一行的 `)` + 下一行的 `;`
            if stripped == ")" and i + 1 < len(self.lines) and self.lines[i + 1].strip() == ";":
                return i + 1
            # 模式 2: 行以 ); 结束（可能带前导内容，例如约束末尾)）
            if stripped.endswith(");"):
                # 过滤掉 INSERT VALUES 中的 ); 
                if "VALUES" not in stripped.upper() and "INSERT" not in stripped.upper():
                    return i
        return None

    def _find_create_trigger_end(self, start_idx):
        """
        找到 CREATE TRIGGER 块的结束行。
        CREATE TRIGGER name ... WHEN ... FOR EACH ROW EXECUTE FUNCTION fn()
        以 `EXECUTE FUNCTION ...` 或 `PROCEDURE...` 行（以 ; 结束）为结束
        """
        for i in range(start_idx, min(start_idx + 20, len(self.lines))):
            stripped = self.lines[i].strip()
            if re.match(r"(EXECUTE\s+(?:FUNCTION|PROCEDURE).+);", stripped, re.IGNORECASE):
                return i
        # 备选: 直接找下一个以 ; 结束的行
        for i in range(start_idx + 1, min(start_idx + 20, len(self.lines))):
            if self.lines[i].strip().endswith(";"):
                return i
        return None

    def _find_create_index_end(self, start_idx):
        """
        CREATE INDEX 通常是单行语句。
        但如果包含 CONCURRENTLY 或多行定义，找到第一个以 ; 结束的行
        """
        for i in range(start_idx, min(start_idx + 10, len(self.lines))):
            stripped = self.lines[i].strip()
            if stripped.endswith(";") and not stripped.startswith("--"):
                return i
        return None

    def _collect_block(self, start_idx, end_idx):
        """收集 block 内所有行"""
        return self.lines[start_idx: end_idx + 1]

    def process(self):
        """主处理流程"""
        print(f"[*] 读取 {self.path}  ({len(self.lines)} 行)")

        # ── 0. 阶段：清理此前由本脚本注入的 COMMENT ON 行（幂等性保证）──
        # 注意：仅移除以 "public." 格式写入的注释，保留 Navicat 原生的其他注释
        existing_comment_count = 0
        cleaned_lines = []
        for ln in self.lines:
            stripped = ln.strip()
            if (stripped.startswith(("COMMENT ON TABLE public.", "COMMENT ON COLUMN public.",
                                      "COMMENT ON INDEX public.", "COMMENT ON TRIGGER "))):
                existing_comment_count += 1
            else:
                cleaned_lines.append(ln)
        if existing_comment_count > 0:
            print(f"[*] 注入前预清理: 移除 {existing_comment_count} 行本脚本先前注入的注释")
            self.lines = cleaned_lines

        # 调试: 统计 CREATE TABLE 匹配数
        tbl_names_found = []
        for idx, ln in enumerate(self.lines):
            s = ln.strip().upper()
            if s.startswith("CREATE TABLE"):
                raw = ln.strip()
                import re as _re
                nm = _re.search(r'CREATE\s+TABLE\s+(?:"public"\.)?"(\w+)"', raw, _re.IGNORECASE)
                if nm:
                    tbl_names_found.append((idx+1, nm.group(1)))
        matched_cnt = sum(1 for _, n in tbl_names_found if n in TABLE_COMMENTS or n in COLUMN_COMMENTS)

        i = 0
        while i < len(self.lines):
            line = self.lines[i]
            stripped = line.strip()

            # ── 跳过已有注释行 ────────────────────────
            if stripped.startswith("COMMENT "):
                i += 1
                continue

            # ── 1. 处理 CREATE TABLE ──────────────────
            if re.match(r"^CREATE\s+TABLE\s", stripped, re.IGNORECASE):
                table_name = self._find_table_name(stripped)
                if table_name and (table_name in TABLE_COMMENTS or table_name in COLUMN_COMMENTS):
                    end_idx = self._find_create_table_end(i)
                    if end_idx is None:
                        print(f"[WARN] {table_name}: 无法找到 CREATE TABLE 结束位置 (L{i+1})")
                    else:
                        comment_lines = self._build_table_comment_block(table_name)
                        if comment_lines:
                            insert_pos = end_idx + 1
                            self.insertions.append((insert_pos, "", comment_lines))
                            tbl_c = 1 if table_name in TABLE_COMMENTS else 0
                            col_c = len(COLUMN_COMMENTS.get(table_name, {}))
                            self.stats["table_comments_added"] += tbl_c
                            self.stats["column_comments_added"] += col_c
                            print(f"  [+] {table_name}: +{tbl_c} 表注释, +{col_c} 列注释")
                        else:
                            self.stats["table_comments_skipped"] += 1
                    i = end_idx + 1 if end_idx is not None else i + 1
                    continue
                elif table_name and "COMMENT " in stripped[:80]:
                    # 跳过已有 Navicat 注释行（不影响主流程）
                    pass
                i += 1
                continue

            # ── 2. 处理 CREATE TRIGGER ───────────────
            if re.match(r"^CREATE\s+(?:OR\s+REPLACE\s+)?TRIGGER\s", stripped, re.IGNORECASE):
                end_idx = self._find_create_trigger_end(i)
                if end_idx is not None:
                    block = self._collect_block(i, end_idx)
                    trigger_name = self._find_trigger_name(block)
                    if trigger_name and trigger_name in TRIGGER_COMMENTS:
                        idempotency_key = f"COMMENT ON TRIGGER {trigger_name}"
                        if not self._exists_nearby(idempotency_key, end_idx + 1):
                            comment_text = TRIGGER_COMMENTS[trigger_name]
                            insert_pos = end_idx + 1
                            self.insertions.append((insert_pos, "", [
                                f"COMMENT ON TRIGGER {trigger_name} IS '{comment_text}';"
                            ]))
                            self.stats["trigger_comments_added"] += 1
                            print(f"  [+] TRIGGER {trigger_name}: 注释已添加")
                        else:
                            self.stats["trigger_comments_skipped"] += 1
                            print(f"  [=] TRIGGER {trigger_name}: 已存在注释，跳过")
                    i = end_idx + 1
                    continue
                i += 1
                continue

            # ── 3. 处理 CREATE INDEX ──────────────────
            if re.match(r"^CREATE\s+(?:UNIQUE\s+)?INDEX\s", stripped, re.IGNORECASE):
                end_idx = self._find_create_index_end(i)
                if end_idx is not None:
                    index_name = self._find_index_name(stripped)
                    # 跳过 tablespace/foreign key 索引（无显式名称则跳过）
                    if index_name and index_name in INDEX_COMMENTS:
                        idempotency_key = f"-- {index_name}:"
                        # 用特殊前缀 "-- idx_name:" 来标记索引注释;
                        # 注释文本也放在该行内（避免插入 COMMENT ON INDEX 需精确 table 引用）
                        # 但用户期望索引注释像触发器一样是一个独立语句。
                        # 插入格式: COMMENT ON INDEX public.idx_xxx IS '用途';
                        # 需同时提取 index 所在表名
                        idx_table = self._find_index_table(stripped)
                        if idx_table and not self._exists_nearby(idempotency_key, end_idx + 1):
                            comment_text = INDEX_COMMENTS[index_name]
                            insert_pos = end_idx + 1
                            self.insertions.append((insert_pos, "", [
                                f"COMMENT ON INDEX public.{index_name} IS '{comment_text}';"
                            ]))
                            self.stats["index_comments_added"] += 1
                            print(f"  [+] INDEX {index_name}: 注释已添加")
                        else:
                            # 跳过已有或非注释列表的索引
                            self.stats["index_comments_skipped"] += 1
                            if index_name:
                                print(f"  [=] INDEX {index_name}: 已存在注释/不在注释列表，跳过")
                    else:
                        self.stats["index_comments_skipped"] += 1
                    i = end_idx + 1
                    continue
                i += 1
                continue

            i += 1

        # ── 4. 应用插入（倒序，避免 line 偏移）──
        # 按 line 索引倒序
        self.insertions.sort(key=lambda x: x[0], reverse=True)
        for (line_idx, indent, text_lines) in self.insertions:
            for offset, txt in enumerate(text_lines):
                self.lines.insert(line_idx + offset, indent + txt)

        # ── 5. 写回文件 ──────────────────────────────
        raw_out = "\n".join(self.lines)
        if self.had_bom:
            raw_out = "﻿" + raw_out
        with open(self.path, "w", encoding="utf-8", newline="\n") as f:
            f.write(raw_out)

        return self.stats


def main():
    print("=" * 60)
    print("inline_comments.py — ydsz-plane-init.sql 内联注释注入")
    print("=" * 60)
    print(f"[*] 目标文件: {SQL_FILE}")
    if not check_file_readable(SQL_FILE):
        print(f"[!] 文件不存在: {SQL_FILE}")
        sys.exit(1)

    inserter = Inserter(SQL_FILE)
    stats = inserter.process()

    print()
    print("=" * 60)
    print("📊 执行统计")
    print("=" * 60)
    added_total = sum(v for k, v in stats.items() if "added" in k)
    skip_total = sum(v for k, v in stats.items() if "skipped" in k)
    for k, v in stats.items():
        print(f"  {k:40s} {v:>6}")
    print(f"  {'─' * 40}  {'─' * 6}")
    print(f"  {'新增注释行数':40s} {added_total:>6}")
    print(f"  {'跳过已有行数':40s} {skip_total:>6}")
    print("=" * 60)
    print("[✓] 完成 — 文件已原地修改")


if __name__ == "__main__":
    main()
