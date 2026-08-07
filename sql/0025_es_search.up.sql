-- 0025_es_search: Elasticsearch 搜索升级 — 索引同步状态追踪 + JQL 解析器支撑
-- 设计参考: Elasticsearch 8.x + IK 分词器 + 类 Jira JQL 语法
--
-- 核心变更:
--   1. es_sync_log: ES 索引同步状态追踪（增量同步 + 对账兜底）
--   2. search_suggestions: 搜索建议缓存（基于用户历史 + 热门搜索）
--   3. search_documents 增加 es_synced_at 字段追踪 ES 同步状态

-- -----------------------------------------------------------------
-- es_sync_log: ES 索引同步状态追踪
-- 记录每次 ES 同步操作的结果，支持增量同步 + 对账兜底
-- -----------------------------------------------------------------
CREATE TABLE es_sync_log (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    doc_type        TEXT NOT NULL CHECK (doc_type IN ('issue','sprint','version')),
    doc_id          BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    action          TEXT NOT NULL CHECK (action IN ('index','delete','update')),
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','synced','failed','skipped')),
    es_version      BIGINT,                                -- ES 文档 _version
    error_message   TEXT,                                  -- 失败原因
    retry_count     INT NOT NULL DEFAULT 0,
    last_synced_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 索引：按状态+重试次数查询待同步项
CREATE INDEX idx_es_sync_pending ON es_sync_log(status, retry_count, created_at)
    WHERE status IN ('pending', 'failed');
-- 索引：按文档类型+ID 查询最新同步状态
CREATE INDEX idx_es_sync_doc ON es_sync_log(doc_type, doc_id, workspace_id);
-- 索引：按工作空间+项目 对账查询
CREATE INDEX idx_es_sync_project ON es_sync_log(project_id, doc_type, last_synced_at DESC);

-- 清理：30 天前的成功日志
CREATE INDEX idx_es_sync_cleanup ON es_sync_log(created_at) WHERE status = 'synced';

-- RLS
ALTER TABLE es_sync_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE es_sync_log FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON es_sync_log
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- search_documents 增加 es_synced_at 字段
-- 用于对账：比较 updated_at > es_synced_at 找出未同步到 ES 的文档
-- -----------------------------------------------------------------
ALTER TABLE search_documents ADD COLUMN IF NOT EXISTS es_synced_at TIMESTAMPTZ;

-- 对账查询专用索引
CREATE INDEX idx_search_docs_es_sync ON search_documents(workspace_id, doc_type, es_synced_at)
    WHERE es_synced_at IS NULL OR updated_at > es_synced_at;

-- -----------------------------------------------------------------
-- search_suggestions: 搜索建议缓存
-- 基于用户历史 + 全局热门搜索，用于搜索框自动补全
-- -----------------------------------------------------------------
CREATE TABLE search_suggestions (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    keyword         TEXT NOT NULL,
    source          TEXT NOT NULL DEFAULT 'history' CHECK (source IN ('history','admin','system')),
    weight          DOUBLE PRECISION NOT NULL DEFAULT 1.0,  -- 排序权重
    last_used_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, keyword)
);

CREATE INDEX idx_search_suggestions_wskey ON search_suggestions(workspace_id, weight DESC, last_used_at DESC);
CREATE INDEX idx_search_suggestions_prefix ON search_suggestions(workspace_id, keyword text_pattern_ops);

-- RLS
ALTER TABLE search_suggestions ENABLE ROW LEVEL SECURITY;
ALTER TABLE search_suggestions FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON search_suggestions
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- 自动记录搜索建议的触发器函数
-- 搜索历史记录时同步更新 suggestions 权重
-- -----------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_update_search_suggestion()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO search_suggestions (workspace_id, keyword, source, weight, last_used_at)
    VALUES (NEW.workspace_id, NEW.query, 'history', 1.0, now())
    ON CONFLICT (workspace_id, keyword) DO UPDATE SET
        weight = search_suggestions.weight + 1.0,
        last_used_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_search_suggestion_update
    AFTER INSERT ON search_history
    FOR EACH ROW
    EXECUTE FUNCTION fn_update_search_suggestion();

-- -----------------------------------------------------------------
-- ES 同步函数：标记文档待同步
-- 在 issues/sprints/versions 变更时由触发器调用
-- -----------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_mark_es_sync()
RETURNS TRIGGER AS $$
DECLARE
    v_doc_type TEXT;
    v_doc_id BIGINT;
BEGIN
    CASE TG_TABLE_NAME
        WHEN 'issues' THEN
            v_doc_type := 'issue';
            v_doc_id := NEW.id;
        WHEN 'sprints' THEN
            v_doc_type := 'sprint';
            v_doc_id := NEW.id;
        WHEN 'versions' THEN
            v_doc_type := 'version';
            v_doc_id := NEW.id;
        ELSE
            RETURN NEW;
    END CASE;

    INSERT INTO es_sync_log (doc_type, doc_id, workspace_id, project_id, action, status)
    VALUES (
        v_doc_type, v_doc_id,
        NEW.workspace_id, NEW.project_id,
        CASE
            WHEN TG_OP = 'DELETE' OR (TG_TABLE_NAME = 'issues' AND NEW.deleted_at IS NOT NULL) THEN 'delete'
            WHEN TG_OP = 'UPDATE' THEN 'update'
            ELSE 'index'
        END,
        'pending'
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 在各表上创建 ES 同步触发器
-- issues 表（已有 search_documents 触发器，追加 ES 同步）
CREATE TRIGGER trg_issue_es_sync
    AFTER INSERT OR UPDATE OF name, description_stripped, type_code, state_id, priority, deleted_at ON issues
    FOR EACH ROW
    EXECUTE FUNCTION fn_mark_es_sync();

-- sprints 表
CREATE TRIGGER trg_sprint_es_sync
    AFTER INSERT OR UPDATE OF name, goal, status ON sprints
    FOR EACH ROW
    EXECUTE FUNCTION fn_mark_es_sync();

-- versions 表
CREATE TRIGGER trg_version_es_sync
    AFTER INSERT OR UPDATE OF name, description, status ON versions
    FOR EACH ROW
    EXECUTE FUNCTION fn_mark_es_sync();

-- -----------------------------------------------------------------
-- ES 同步对账函数：找出未同步到 ES 的文档
-- 比较 search_documents.updated_at > search_documents.es_synced_at
-- 由 Worker 定时调用（每 10 分钟）
-- -----------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_es_reconcile(
    p_workspace_id BIGINT,
    p_limit INT DEFAULT 500
)
RETURNS TABLE (
    doc_type TEXT,
    doc_id BIGINT,
    project_id BIGINT,
    title TEXT,
    identifier TEXT,
    content TEXT,
    metadata JSONB,
    updated_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        d.doc_type,
        d.doc_id,
        d.project_id,
        d.title,
        d.identifier,
        d.content,
        d.metadata,
        d.updated_at
    FROM search_documents d
    WHERE d.workspace_id = p_workspace_id
      AND (d.es_synced_at IS NULL OR d.updated_at > d.es_synced_at)
    ORDER BY d.updated_at ASC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;
