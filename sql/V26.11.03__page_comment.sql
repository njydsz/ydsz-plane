-- V26.11.03__page_comment.sql — 页面评论与行内评论
--
-- 对标 Plane Page Comments / Inline Comments：
--   - page_comments         ：页面级 / 行内评论（含线程回复、解决状态、提及）
--   - page_comment_reactions：评论表情回应
--
-- 设计要点：
--   - 评论同时服务「项目文档页面（pages）」与「知识库文档（knowledge_pages）」，
--     通过 page_type 区分归属，page_id 指向对应表的主键；
--   - 线程模型：parent_id 自引用形成树；resolved_at / resolved_by 标记线程已解决；
--   - 行内评论：block_id + quote_text 锚定文档块与原文引用；
--   - 提及：mentions 存储 @username 解析后的用户 ID 数组（BIGINT[]）；
--   - 软删除：deleted_at 标记，列表 / 详情查询统一过滤。

-- ---------------------------------------------------------------------------
-- page_comments
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS page_comments (
    id           BIGINT       PRIMARY KEY,
    workspace_id BIGINT       NOT NULL,
    page_id      BIGINT       NOT NULL,
    page_type    VARCHAR(20)  NOT NULL DEFAULT 'project',
    parent_id    BIGINT,
    author_id    BIGINT       NOT NULL,
    content      TEXT         NOT NULL,
    block_id     VARCHAR(200),
    quote_text   TEXT,
    resolved_at  TIMESTAMPTZ,
    resolved_by  BIGINT       NOT NULL DEFAULT 0,
    mentions     BIGINT[]     NOT NULL DEFAULT '{}',
    is_edited    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,

    CONSTRAINT chk_page_comments_page_type CHECK (page_type IN ('project', 'knowledge')),
    CONSTRAINT chk_page_comments_parent_not_self CHECK (parent_id IS NULL OR parent_id <> id)
);

-- 页面 + 类型：列表查询主路径
CREATE INDEX IF NOT EXISTS idx_page_comments_page
    ON page_comments (workspace_id, page_id, page_type)
    WHERE deleted_at IS NULL;

-- 线程父评论：查询某评论的回复
CREATE INDEX IF NOT EXISTS idx_page_comments_parent
    ON page_comments (parent_id)
    WHERE deleted_at IS NULL AND parent_id IS NOT NULL;

-- 行内评论：按文档块过滤
CREATE INDEX IF NOT EXISTS idx_page_comments_block
    ON page_comments (page_id, block_id)
    WHERE deleted_at IS NULL AND block_id IS NOT NULL;

-- 作者维度：查询某用户的评论历史
CREATE INDEX IF NOT EXISTS idx_page_comments_author
    ON page_comments (author_id)
    WHERE deleted_at IS NULL;

-- 已解决线程：resolved_at 非空标记
CREATE INDEX IF NOT EXISTS idx_page_comments_resolved
    ON page_comments (page_id, resolved_at)
    WHERE deleted_at IS NULL AND resolved_at IS NOT NULL;

-- ---------------------------------------------------------------------------
-- page_comment_reactions
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS page_comment_reactions (
    id         BIGINT       PRIMARY KEY,
    comment_id BIGINT       NOT NULL,
    user_id    BIGINT       NOT NULL,
    emoji      VARCHAR(32)  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),

    -- 同一评论 + 同一用户 + 同一 emoji 唯一（幂等添加）
    CONSTRAINT uq_page_comment_reactions UNIQUE (comment_id, user_id, emoji)
);

-- 按评论列出全部回应
CREATE INDEX IF NOT EXISTS idx_page_comment_reactions_comment
    ON page_comment_reactions (comment_id);

-- 按用户反查其回应
CREATE INDEX IF NOT EXISTS idx_page_comment_reactions_user
    ON page_comment_reactions (user_id);

COMMENT ON TABLE page_comments IS '页面评论（页面级 / 行内，含线程回复与解决状态）';
COMMENT ON TABLE page_comment_reactions IS '评论表情回应';
COMMENT ON COLUMN page_comments.page_id IS '关联页面 ID（pages 或 knowledge_pages 主键）';
COMMENT ON COLUMN page_comments.page_type IS '页面类型：project（项目文档）| knowledge（知识库）';
COMMENT ON COLUMN page_comments.parent_id IS '父评论 ID（线程回复，NULL 表示根评论）';
COMMENT ON COLUMN page_comments.block_id IS '行内评论锚定的文档块标识';
COMMENT ON COLUMN page_comments.quote_text IS '行内评论引用的原文';
COMMENT ON COLUMN page_comments.mentions IS '@提及的用户 ID 列表';
COMMENT ON COLUMN page_comments.resolved_at IS '线程解决时间（NULL 表示未解决）';
