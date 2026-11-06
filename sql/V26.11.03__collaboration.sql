-- V26.11.03__collaboration.sql — 实时协作（编辑会话 + 乐观锁 + 操作日志）
--
-- 对标 Plane 的实时协作体验：
--   - collaboration_sessions     ：谁在编辑哪篇文档
--   - collaboration_documents    ：文档版本向量（乐观锁）
--   - collaboration_operations   ：操作日志（用于广播与冲突检测）

-- ---------------------------------------------------------------------------
-- collaboration_sessions
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS collaboration_sessions (
    id            BIGINT       PRIMARY KEY,
    workspace_id  BIGINT       NOT NULL,
    document_type VARCHAR(40)  NOT NULL,
    document_id   BIGINT       NOT NULL,
    user_id       BIGINT       NOT NULL,
    user_name     VARCHAR(200),
    display_name  VARCHAR(200),
    status        VARCHAR(20)  NOT NULL DEFAULT 'active',
    base_version  BIGINT       NOT NULL DEFAULT 0,
    last_seen_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_collab_session_status CHECK (status IN ('active','idle','closed')),
    CONSTRAINT chk_collab_session_type CHECK (document_type IN
        ('project_page','knowledge_page','issue_description','sprint_goal'))
);

CREATE INDEX IF NOT EXISTS idx_collab_session_doc
    ON collaboration_sessions (workspace_id, document_type, document_id, status);
CREATE INDEX IF NOT EXISTS idx_collab_session_user
    ON collaboration_sessions (user_id) WHERE status IN ('active','idle');

-- ---------------------------------------------------------------------------
-- collaboration_documents
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS collaboration_documents (
    id            BIGINT      PRIMARY KEY,
    document_type VARCHAR(40) NOT NULL,
    document_id   BIGINT      NOT NULL,
    version       BIGINT      NOT NULL DEFAULT 0,
    updated_by    BIGINT,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (document_type, document_id)
);

CREATE INDEX IF NOT EXISTS idx_collab_document_lookup
    ON collaboration_documents (document_type, document_id);

-- ---------------------------------------------------------------------------
-- collaboration_operations
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS collaboration_operations (
    id             BIGINT      PRIMARY KEY,
    workspace_id   BIGINT      NOT NULL,
    document_type  VARCHAR(40) NOT NULL,
    document_id    BIGINT      NOT NULL,
    user_id        BIGINT      NOT NULL,
    user_name      VARCHAR(200),
    op_type        VARCHAR(20) NOT NULL,
    base_version   BIGINT      NOT NULL DEFAULT 0,
    target_version BIGINT      NOT NULL DEFAULT 0,
    payload        TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_collab_op_doc
    ON collaboration_operations (document_type, document_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_collab_op_workspace
    ON collaboration_operations (workspace_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_collab_op_user
    ON collaboration_operations (user_id, created_at DESC);

COMMENT ON TABLE collaboration_sessions IS '文档编辑会话（在线编辑者）';
COMMENT ON TABLE collaboration_documents IS '文档版本向量（乐观锁）';
COMMENT ON TABLE collaboration_operations IS '协作操作日志';
