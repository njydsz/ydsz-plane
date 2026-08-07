-- 0011_issue_comments: 工作项评论系统
-- 表: issue_comments
-- 支持富文本评论（JSON 格式）+ @提及

CREATE TABLE IF NOT EXISTS issue_comments (
    id BIGSERIAL PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    project_id BIGINT NOT NULL REFERENCES projects(id),
    issue_id BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    -- 富文本内容
    content_json JSONB NOT NULL DEFAULT '{}',  -- TipTap JSON 格式
    content_html TEXT,                          -- 渲染后的 HTML
    content_stripped TEXT,                      -- 纯文本（用于搜索/摘要）
    -- 评论人
    created_by BIGINT NOT NULL REFERENCES users(id),
    -- 元数据
    mentions BIGINT[] DEFAULT '{}',             -- @提及的用户 ID 列表
    parent_id BIGINT REFERENCES issue_comments(id) ON DELETE SET NULL,  -- 回复评论
    is_edited BOOLEAN NOT NULL DEFAULT false,
    edited_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引：按工作项查询评论（时间线）
CREATE INDEX idx_issue_comments_issue
    ON issue_comments(issue_id, created_at);

-- 索引：按评论人查询
CREATE INDEX idx_issue_comments_author
    ON issue_comments(created_by, created_at DESC);

-- RLS
ALTER TABLE issue_comments ENABLE ROW LEVEL SECURITY;
CREATE POLICY comments_isolation ON issue_comments
    USING (workspace_id = current_setting('app.workspace_id')::bigint);

COMMENT ON TABLE issue_comments IS '工作项评论表（支持富文本 + @提及 + 嵌套回复）';
COMMENT ON COLUMN issue_comments.content_json IS 'TipTap 编辑器的 JSON 输出';
COMMENT ON COLUMN issue_comments.mentions IS '@提及的用户 ID 数组';
COMMENT ON COLUMN issue_comments.parent_id IS '父评论 ID（嵌套回复）';
