-- 0016_intake.up.sql — Intake 收件箱：提交通道与工单评审
-- 对标: GitHub Issues Forms, Jira Service Management portals, Linear Inbox
BEGIN;

-- intake_channels 收件通道（公开表单）
CREATE TABLE IF NOT EXISTS intake_channels (
    id                  BIGSERIAL PRIMARY KEY,
    workspace_id        BIGINT NOT NULL,
    project_id          BIGINT,                                 -- NULL 表示工作空间级
    slug                VARCHAR(64) NOT NULL,                   -- URL slug: /intake/{slug}
    name                VARCHAR(100) NOT NULL,
    description         TEXT,
    is_public           BOOLEAN NOT NULL DEFAULT true,          -- 是否公开免登录提交
    -- 默认行为
    default_issue_type  VARCHAR(20) NOT NULL DEFAULT 'requirement', -- requirement|defect|task
    default_priority    SMALLINT NOT NULL DEFAULT 0,           -- 默认优先级
    -- 自动分配规则（JSONB 数组，按顺序匹配）
    auto_assign_rules   JSONB NOT NULL DEFAULT '[]'::jsonb,    -- [{match:{keyword:"x"},assign_to:123}]
    -- 限流与安全
    rate_limit_per_min  SMALLINT NOT NULL DEFAULT 20,           -- 每分钟每 IP 提交上限
    require_captcha     BOOLEAN NOT NULL DEFAULT true,           -- 是否需要验证码
    custom_fields       JSONB NOT NULL DEFAULT '[]'::jsonb,     -- 自定义字段定义
    branding            JSONB NOT NULL DEFAULT '{}'::jsonb,     -- {logo_url,heading,subheading}
    -- 通知
    notify_on_submit    BOOLEAN NOT NULL DEFAULT true,           -- 提交后通知管理员
    notify_users        BIGINT[] NOT NULL DEFAULT '{}',         -- 通知的用户 ID 列表
    -- 状态
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_by          BIGINT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_intake_channel_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    CONSTRAINT fk_intake_channel_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT uq_intake_channel_slug UNIQUE (workspace_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_intake_channels_workspace ON intake_channels (workspace_id);
CREATE INDEX IF NOT EXISTS idx_intake_channels_project ON intake_channels (project_id) WHERE project_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_intake_channels_slug ON intake_channels (workspace_id, slug) WHERE is_active = true;

-- intake_issues 收件工单
CREATE TABLE IF NOT EXISTS intake_issues (
    id                  BIGSERIAL PRIMARY KEY,
    channel_id          BIGINT NOT NULL,
    workspace_id        BIGINT NOT NULL,
    project_id          BIGINT,                                 -- 分配后更新
    -- 跟踪回执
    tracking_id         VARCHAR(32) NOT NULL,                   -- YD-IN-XXXX 提交回执
    -- 提交者（可能未注册）
    submitter_name      VARCHAR(100) NOT NULL,
    submitter_email     VARCHAR(255) NOT NULL,
    submitter_user_id   BIGINT,                                 -- 如果提交时是登录用户
    -- 内容
    title               VARCHAR(255) NOT NULL,
    description         TEXT NOT NULL DEFAULT '',
    issue_type          VARCHAR(20) NOT NULL DEFAULT 'requirement',
    priority            SMALLINT NOT NULL DEFAULT 0,
    custom_fields       JSONB NOT NULL DEFAULT '{}'::jsonb,     -- 用户填写的自定义字段值
    -- 附件 IDs
    attachment_ids      BIGINT[] NOT NULL DEFAULT '{}',
    -- 状态机：open|accepted|rejected|archived
    status              VARCHAR(20) NOT NULL DEFAULT 'open',
    status_reason       TEXT,                                   -- 拒绝原因等
    -- 转正关联
    converted_issue_id  BIGINT,                                 -- 转正后的正式工作项 ID
    assigned_to         BIGINT,                                 -- 处理人（自动或手动分配）
    -- 审核
    reviewed_by         BIGINT,                                 -- 管理员 user ID
    reviewed_at         TIMESTAMPTZ,
    -- 提交者通知
    notify_on_status    BOOLEAN NOT NULL DEFAULT true,           -- 状态变化时通知提交者
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_intake_issue_channel FOREIGN KEY (channel_id) REFERENCES intake_channels(id) ON DELETE CASCADE,
    CONSTRAINT fk_intake_issue_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    CONSTRAINT fk_intake_issue_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL,
    CONSTRAINT uq_intake_issue_tracking UNIQUE (workspace_id, tracking_id),
    CONSTRAINT chk_intake_status CHECK (status IN ('open','accepted','rejected','archived'))
);

CREATE INDEX IF NOT EXISTS idx_intake_issues_channel ON intake_issues (channel_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_intake_issues_workspace ON intake_issues (workspace_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_intake_issues_tracking ON intake_issues (tracking_id);
CREATE INDEX IF NOT EXISTS idx_intake_issues_status ON intake_issues (workspace_id, status);
CREATE INDEX IF NOT EXISTS idx_intake_issues_submitter ON intake_issues (submitter_email);

COMMIT;
