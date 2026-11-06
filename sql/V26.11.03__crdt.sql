-- V26.11.03__crdt.sql — CRDT 文档协同（Y.js over WebSocket）
--
-- 存储协同文档的 Y.js 更新累积状态（顺序拼接，客户端按序 apply）。

CREATE TABLE IF NOT EXISTS crdt_documents (
    doc_id      TEXT        PRIMARY KEY,
    -- Y.js update 二进制累积（bytea）
    state       BYTEA       NOT NULL DEFAULT ''::bytea,
    -- 文档类型（project_page / knowledge_page / issue_description）
    doc_type    VARCHAR(40),
    workspace_id BIGINT,
    version     BIGINT      NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_crdt_documents_workspace ON crdt_documents (workspace_id);
CREATE INDEX IF NOT EXISTS idx_crdt_documents_updated ON crdt_documents (updated_at DESC);

COMMENT ON TABLE crdt_documents IS 'CRDT 协同文档状态（Y.js update 累积）';
