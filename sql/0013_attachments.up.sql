-- 0013_attachments.up.sql
-- 附件表：支持 S3/MinIO 对象存储，统一管理工作项与评论的附件。

CREATE TABLE IF NOT EXISTS attachments (
    id              BIGSERIAL       PRIMARY KEY,
    workspace_id    BIGINT          NOT NULL,
    project_id      BIGINT          NOT NULL,
    -- 所属实体（多态关联）
    entity_type     VARCHAR(20)     NOT NULL CHECK (entity_type IN ('issue', 'comment', 'workspace', 'project')),
    entity_id       BIGINT          NOT NULL,
    -- 文件信息
    file_name       VARCHAR(512)    NOT NULL,
    file_size       BIGINT          NOT NULL,
    content_type    VARCHAR(128)    NOT NULL DEFAULT 'application/octet-stream',
    -- 存储信息
    storage_key     VARCHAR(512)    NOT NULL,  -- S3/MinIO object key
    storage_url     VARCHAR(2048),             -- 预签名 URL（按需生成，不持久化完整 URL）
    -- 缩略图（图片类附件）
    thumb_key       VARCHAR(512),
    -- 上传者
    uploaded_by     BIGINT          NOT NULL,
    -- 软删除
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 按实体查询附件
CREATE INDEX idx_attachments_entity ON attachments (entity_type, entity_id) WHERE deleted_at IS NULL;

-- 按工作空间范围查询
CREATE INDEX idx_attachments_workspace ON attachments (workspace_id) WHERE deleted_at IS NULL;

-- 按上传者查询
CREATE INDEX idx_attachments_uploader ON attachments (uploaded_by) WHERE deleted_at IS NULL;

-- 行级安全策略：仅允许工作空间成员访问
ALTER TABLE attachments ENABLE ROW LEVEL SECURITY;
CREATE POLICY attachments_workspace_isolation ON attachments
    USING (workspace_id = current_setting('app.workspace_id')::bigint);
