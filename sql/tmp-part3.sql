
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
