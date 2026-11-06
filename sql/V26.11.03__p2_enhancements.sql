-- V26.11.03__p2_enhancements.sql — P2 增强：评论审核 / 重复工作项 / 里程碑 / 工作空间视图
--
-- 本迁移包含四个相对独立的子域，统一放在同一文件中以便一次性部署：
--   1. page_comments 增加审核列（hidden / moderated_at / moderated_by / moderation_reason）；
--   2. recurring_rules         ：重复工作项规则（cron 模板 → 自动生成工作项）；
--   3. milestones + milestone_work_items：项目里程碑及工作项关联；
--   4. workspace_views         ：工作空间级命名视图（board/list/spreadsheet/gantt/calendar）。

-- ===========================================================================
-- 1. 评论审核（page_comments 增量列）
-- ===========================================================================

-- 若列已存在（重复执行场景）先忽略；PostgreSQL 11+ 支持 ADD COLUMN IF NOT EXISTS。
ALTER TABLE page_comments ADD COLUMN IF NOT EXISTS hidden            BOOLEAN      NOT NULL DEFAULT FALSE;
ALTER TABLE page_comments ADD COLUMN IF NOT EXISTS moderated_at     TIMESTAMPTZ;
ALTER TABLE page_comments ADD COLUMN IF NOT EXISTS moderated_by     BIGINT       NOT NULL DEFAULT 0;
ALTER TABLE page_comments ADD COLUMN IF NOT EXISTS moderation_reason TEXT;

-- 列表过滤：默认排除 hidden，显式 hidden=true 时才包含（含隐藏的审核视图）。
CREATE INDEX IF NOT EXISTS idx_page_comments_hidden
    ON page_comments (workspace_id, page_id)
    WHERE deleted_at IS NULL AND hidden = FALSE;

COMMENT ON COLUMN page_comments.hidden             IS '是否被审核隐藏（true 时不在普通列表返回）';
COMMENT ON COLUMN page_comments.moderated_at      IS '最近一次审核操作时间';
COMMENT ON COLUMN page_comments.moderated_by      IS '最近一次审核操作人 ID（0 表示系统）';
COMMENT ON COLUMN page_comments.moderation_reason IS '审核原因（hide/unhide 备注）';

-- ===========================================================================
-- 2. 重复工作项（recurring_rules）
-- ===========================================================================
CREATE TABLE IF NOT EXISTS recurring_rules (
    id                   BIGINT       PRIMARY KEY,
    workspace_id         BIGINT       NOT NULL,
    project_id           BIGINT       NOT NULL,
    work_item_type       VARCHAR(32)  NOT NULL,               -- task | requirement | defect
    title_template       VARCHAR(500) NOT NULL,
    description_template TEXT,
    cron_expr            VARCHAR(128) NOT NULL,               -- 5 段式 cron（分 时 日 月 周）
    next_run_at          TIMESTAMPTZ  NOT NULL,
    last_run_at          TIMESTAMPTZ,
    assignee_ids         BIGINT[]     NOT NULL DEFAULT '{}',
    label_ids            BIGINT[]     NOT NULL DEFAULT '{}',
    state_id             BIGINT,
    enabled              BOOLEAN      NOT NULL DEFAULT TRUE,
    created_by           BIGINT       NOT NULL,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT chk_recurring_rules_type CHECK (work_item_type IN ('task','requirement','defect'))
);

CREATE INDEX IF NOT EXISTS idx_recurring_rules_due
    ON recurring_rules (enabled, next_run_at)
    WHERE enabled = TRUE;

CREATE INDEX IF NOT EXISTS idx_recurring_rules_project
    ON recurring_rules (workspace_id, project_id);

COMMENT ON TABLE  recurring_rules                    IS '重复工作项规则（按 cron 模板定时生成工作项）';
COMMENT ON COLUMN recurring_rules.work_item_type     IS '生成的工作项类型：task|requirement|defect';
COMMENT ON COLUMN recurring_rules.title_template     IS '标题模板（支持 {{date}} 占位符）';
COMMENT ON COLUMN recurring_rules.description_template IS '描述模板（纯文本 / HTML）';
COMMENT ON COLUMN recurring_rules.cron_expr          IS '5 段式 cron 表达式';
COMMENT ON COLUMN recurring_rules.next_run_at        IS '下次触发时间（UTC）';
COMMENT ON COLUMN recurring_rules.assignee_ids       IS '生成工作项的指派用户 ID 列表';
COMMENT ON COLUMN recurring_rules.label_ids          IS '生成工作项的标签 ID 列表';
COMMENT ON COLUMN recurring_rules.state_id           IS '生成工作项的初始状态 ID（0/NULL 取项目默认）';
COMMENT ON COLUMN recurring_rules.enabled            IS '是否启用（false 时调度器跳过）';

-- ===========================================================================
-- 3. 里程碑（milestones + milestone_work_items）
-- ===========================================================================
CREATE TABLE IF NOT EXISTS milestones (
    id           BIGINT       PRIMARY KEY,
    workspace_id BIGINT       NOT NULL,
    project_id   BIGINT       NOT NULL,
    name         VARCHAR(200) NOT NULL,
    description  TEXT,
    target_date  DATE,
    status       VARCHAR(20)  NOT NULL DEFAULT 'not_started',
    created_by   BIGINT       NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT chk_milestone_status CHECK (status IN ('not_started','in_progress','completed','missed'))
);

CREATE INDEX IF NOT EXISTS idx_milestones_project
    ON milestones (workspace_id, project_id, target_date);

CREATE INDEX IF NOT EXISTS idx_milestones_status
    ON milestones (project_id, status);

-- 说明：不使用软删除列，物理删除里程碑时由应用层先清理关联表。

-- 里程碑 ↔ 工作项关联表
CREATE TABLE IF NOT EXISTS milestone_work_items (
    milestone_id   BIGINT      NOT NULL,
    work_item_id   BIGINT      NOT NULL,
    work_item_type VARCHAR(32) NOT NULL,               -- task | requirement | defect
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (milestone_id, work_item_id),
    CONSTRAINT chk_milestone_work_item_type CHECK (work_item_type IN ('task','requirement','defect'))
);

CREATE INDEX IF NOT EXISTS idx_milestone_work_items_item
    ON milestone_work_items (work_item_id);

COMMENT ON TABLE  milestones                 IS '项目里程碑';
COMMENT ON COLUMN milestones.target_date     IS '目标日期（用于进度预测与逾期判定）';
COMMENT ON COLUMN milestones.status          IS '状态：not_started|in_progress|completed|missed';
COMMENT ON TABLE  milestone_work_items      IS '里程碑关联的工作项（跨类型）';
COMMENT ON COLUMN milestone_work_items.work_item_type IS '工作项类型：task|requirement|defect';

-- ===========================================================================
-- 4. 工作空间视图（workspace_views）
-- ===========================================================================
CREATE TABLE IF NOT EXISTS workspace_views (
    id              BIGINT       PRIMARY KEY,
    workspace_id    BIGINT       NOT NULL,
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    layout          VARCHAR(32)  NOT NULL DEFAULT 'list',  -- board|list|spreadsheet|gantt|calendar
    filters         JSONB        NOT NULL DEFAULT '{}'::jsonb,
    display_options JSONB        NOT NULL DEFAULT '{}'::jsonb,
    sort            JSONB        NOT NULL DEFAULT '{}'::jsonb,
    is_public       BOOLEAN      NOT NULL DEFAULT FALSE,
    is_favorite     BOOLEAN      NOT NULL DEFAULT FALSE,
    created_by      BIGINT       NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT chk_workspace_view_layout CHECK (layout IN ('board','list','spreadsheet','gantt','calendar'))
);

CREATE INDEX IF NOT EXISTS idx_workspace_views_workspace
    ON workspace_views (workspace_id, is_favorite DESC, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_workspace_views_owner
    ON workspace_views (created_by);

COMMENT ON TABLE  workspace_views                  IS '工作空间级命名视图';
COMMENT ON COLUMN workspace_views.layout           IS '布局：board|list|spreadsheet|gantt|calendar';
COMMENT ON COLUMN workspace_views.filters          IS '过滤器配置（JSONB）';
COMMENT ON COLUMN workspace_views.display_options  IS '展示选项（列、分组等，JSONB）';
COMMENT ON COLUMN workspace_views.sort             IS '排序配置（JSONB）';
COMMENT ON COLUMN workspace_views.is_public        IS '是否对全空间可见';
COMMENT ON COLUMN workspace_views.is_favorite      IS '当前用户是否收藏';
