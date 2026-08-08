-- ==========================================================================
-- ydsz-plane 知识库（Knowledge Base）迁移文件
--
-- 包含 4 张表：
--   knowledge_spaces       知识库空间
--   knowledge_pages        文档（无限层级树，parent_id 自引用）
--   knowledge_page_versions  版本快照
--   knowledge_page_relations 文档与工作项关联
--
-- 设计约束：
--   - 所有表带 workspace_id 以支持 RLS set_config 租户隔离
--   - spaces / pages 支持软删除（deleted_at timestamptz）
--   - 文档支持乐观锁（version 字段）
-- ==========================================================================

BEGIN;

-- ==========================================================================
-- knowledge_spaces — 知识库空间
-- ==========================================================================
CREATE TABLE IF NOT EXISTS public.knowledge_spaces (
    id                   BIGSERIAL   PRIMARY KEY,
    workspace_id         BIGINT      NOT NULL REFERENCES public.workspaces(id) ON DELETE CASCADE,
    project_id           BIGINT      REFERENCES public.projects(id) ON DELETE SET NULL,
    name                 VARCHAR(255) NOT NULL,
    slug                 VARCHAR(128) NOT NULL UNIQUE,
    description          TEXT        DEFAULT '',
    owner_id             BIGINT      REFERENCES public.users(id) ON DELETE SET NULL,
    default_permission   VARCHAR(32) NOT NULL DEFAULT 'viewer'
                                  CHECK (default_permission IN ('viewer','editor','admin','owner')),
    is_private           BOOLEAN     NOT NULL DEFAULT TRUE,
    cover_image          VARCHAR(512) DEFAULT '',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at           TIMESTAMPTZ
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_ks_workspace ON public.knowledge_spaces(workspace_id);
CREATE INDEX IF NOT EXISTS idx_ks_project   ON public.knowledge_spaces(project_id) WHERE project_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_ks_deleted   ON public.knowledge_spaces(deleted_at) WHERE deleted_at IS NULL;

COMMENT ON TABLE  public.knowledge_spaces               IS '知识库空间：可挂在工作空间或项目下';
COMMENT ON COLUMN public.knowledge_spaces.workspace_id IS '所属工作空间（RLS 租户隔离键）';
COMMENT ON COLUMN public.knowledge_spaces.project_id   IS '所属项目（可空 = 工作空间级空间）';
COMMENT ON COLUMN public.knowledge_spaces.slug         IS '空间标识（URL 友好）';
COMMENT ON COLUMN public.knowledge_spaces.deleted_at   IS '软删除时间戳（NULL = 未删除）';

-- ==========================================================================
-- knowledge_pages — 文档（自引用 parent_id 实现无限层级树）
-- ==========================================================================
CREATE TABLE IF NOT EXISTS public.knowledge_pages (
    id                   BIGSERIAL   PRIMARY KEY,
    workspace_id         BIGINT      NOT NULL REFERENCES public.workspaces(id) ON DELETE CASCADE,
    space_id             BIGINT      NOT NULL REFERENCES public.knowledge_spaces(id) ON DELETE CASCADE,
    parent_id            BIGINT      REFERENCES public.knowledge_pages(id) ON DELETE SET NULL,
    lft                  BIGINT      NOT NULL DEFAULT 0,
    rgt                  BIGINT      NOT NULL DEFAULT 0,
    depth                INTEGER     NOT NULL DEFAULT 0,
    title                VARCHAR(512) NOT NULL,
    path                 VARCHAR(2048) DEFAULT '',
    content_md           TEXT        DEFAULT '',
    content_html         TEXT        DEFAULT '',
    version              BIGINT      NOT NULL DEFAULT 1,
    status               VARCHAR(32) NOT NULL DEFAULT 'draft'
                                  CHECK (status IN ('draft','published','archived')),
    sort_order           BIGINT      NOT NULL DEFAULT 0,
    is_pinned            BOOLEAN     NOT NULL DEFAULT FALSE,
    is_featured          BOOLEAN     NOT NULL DEFAULT FALSE,
    view_count           BIGINT      NOT NULL DEFAULT 0,
    created_by           BIGINT      REFERENCES public.users(id) ON DELETE SET NULL,
    updated_by           BIGINT      REFERENCES public.users(id) ON DELETE SET NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at           TIMESTAMPTZ,

    -- 自引用外键约束
    CONSTRAINT fk_kp_parent FOREIGN KEY (parent_id)
        REFERENCES public.knowledge_pages(id) ON DELETE SET NULL
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_kp_ws_space  ON public.knowledge_pages(workspace_id, space_id);
CREATE INDEX IF NOT EXISTS idx_kp_parent    ON public.knowledge_pages(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_kp_status    ON public.knowledge_pages(status);
CREATE INDEX IF NOT EXISTS idx_kp_deleted   ON public.knowledge_pages(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_kp_created   ON public.knowledge_pages(created_at DESC);

COMMENT ON TABLE  public.knowledge_pages               IS '知识库文档（无限层级树）';
COMMENT ON COLUMN public.knowledge_pages.workspace_id IS '所属工作空间（RLS 租户隔离键）';
COMMENT ON COLUMN public.knowledge_pages.space_id     IS '所属空间';
COMMENT ON COLUMN public.knowledge_pages.parent_id    IS '父文档 ID（自引用）';
COMMENT ON COLUMN public.knowledge_pages.lft          IS '嵌套集合左值';
COMMENT ON COLUMN public.knowledge_pages.rgt          IS '嵌套集合右值';
COMMENT ON COLUMN public.knowledge_pages.depth        IS '文档层级深度（0 = 根）';
COMMENT ON COLUMN public.knowledge_pages.path        IS '完整路径（如 /features/login）';
COMMENT ON COLUMN public.knowledge_pages.version      IS '乐观锁版本号（每次内容更新自动 +1）';
COMMENT ON COLUMN public.knowledge_pages.deleted_at   IS '软删除时间戳';

-- ==========================================================================
-- knowledge_page_versions — 版本快照（内容变更时自动记录）
-- ==========================================================================
CREATE TABLE IF NOT EXISTS public.knowledge_page_versions (
    id                   BIGSERIAL   PRIMARY KEY,
    page_id              BIGINT      NOT NULL REFERENCES public.knowledge_pages(id) ON DELETE CASCADE,
    version              BIGINT      NOT NULL,
    title                VARCHAR(512) NOT NULL,
    content_md           TEXT        DEFAULT '',
    content_html         TEXT        DEFAULT '',
    change_summary       VARCHAR(256) DEFAULT '',
    created_by           BIGINT      REFERENCES public.users(id) ON DELETE SET NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- 同一文档下版本号唯一
    CONSTRAINT uq_kpv_page_version UNIQUE (page_id, version)
);

CREATE INDEX IF NOT EXISTS idx_kpv_page     ON public.knowledge_page_versions(page_id);
CREATE INDEX IF NOT EXISTS idx_kpv_version  ON public.knowledge_page_versions(page_id, version DESC);

COMMENT ON TABLE  public.knowledge_page_versions               IS '文档版本快照（回滚历史）';
COMMENT ON COLUMN public.knowledge_page_versions.page_id       IS '所属文档';
COMMENT ON COLUMN public.knowledge_page_versions.version       IS '快照版本号';
COMMENT ON COLUMN public.knowledge_page_versions.change_summary IS '变更摘要';

-- ==========================================================================
-- knowledge_page_relations — 文档与工作项关联
-- ==========================================================================
CREATE TABLE IF NOT EXISTS public.knowledge_page_relations (
    id                   BIGSERIAL   PRIMARY KEY,
    page_id              BIGINT      NOT NULL REFERENCES public.knowledge_pages(id) ON DELETE CASCADE,
    issue_id             BIGINT      NOT NULL REFERENCES public.issues(id) ON DELETE CASCADE,
    relation_type        VARCHAR(32) NOT NULL DEFAULT 'referenced'
                                  CHECK (relation_type IN ('referenced','referencing')),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- 同一文档+工作项+类型组合唯一
    CONSTRAINT uq_kpr_page_issue_type UNIQUE (page_id, issue_id, relation_type)
);

CREATE INDEX IF NOT EXISTS idx_kpr_page  ON public.knowledge_page_relations(page_id);
CREATE INDEX IF NOT EXISTS idx_kpr_issue ON public.knowledge_page_relations(issue_id);

COMMENT ON TABLE  public.knowledge_page_relations            IS '文档与工作项关联表';
COMMENT ON COLUMN public.knowledge_page_relations.page_id    IS '文档 ID';
COMMENT ON COLUMN public.knowledge_page_relations.issue_id   IS '工作项 ID';
COMMENT ON COLUMN public.knowledge_page_relations.relation_type IS '关联类型（referenced=被引用 / referencing=引用方）';

COMMIT;
