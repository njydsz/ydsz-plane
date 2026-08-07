
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
