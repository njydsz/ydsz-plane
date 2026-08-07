-- 0008_search: full-text search + search history/favorites (Sprint 8 — M5 全局搜索)
-- 参考: PostgreSQL FTS + Plane/Jira search architecture

-- -----------------------------------------------------------------
-- issues 全文搜索增强: tsvector 列 + GIN 索引
-- 现有 issues 表已有 description_stripped 列作为检索源
-- 新增 generated tsvector 列 + 触发器自动维护
-- -----------------------------------------------------------------
ALTER TABLE issues ADD COLUMN IF NOT EXISTS search_tsv tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(description_stripped, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(type_code, '')), 'C')
    ) STORED;

CREATE INDEX idx_issues_search_tsv ON issues USING GIN(search_tsv);

-- 简单语种配置（中文场景后续接 IK 分词 ES，英文/拼音用 'simple'）
-- 加权: A=标识符/名称 > B=描述 > C=类型

-- -----------------------------------------------------------------
-- search_history: 用户搜索历史（最近 50 条 per user）
-- 用于搜索页"最近搜索"快速回显
-- -----------------------------------------------------------------
CREATE TABLE search_history (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    query           TEXT NOT NULL,                        -- 原始查询文本
    filters         JSONB NOT NULL DEFAULT '{}'::jsonb,  -- 过滤条件快照
    result_count    INT NOT NULL DEFAULT 0,             -- 上次搜索结果数
    searched_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_search_history_user ON search_history(user_id, searched_at DESC);
CREATE INDEX idx_search_history_ws_user ON search_history(workspace_id, user_id);

-- RLS
ALTER TABLE search_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE search_history FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON search_history
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- search_bookmarks: 搜索收藏/保存的过滤器
-- 常用复杂查询可命名保存，供重复使用（类似 Jira 的 Saved Filter / Plane 的 View）
-- -----------------------------------------------------------------
CREATE TABLE search_bookmarks (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT REFERENCES projects(id) ON DELETE CASCADE,  -- NULL = 全局
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,                        -- 收藏名（如"我本周待处理的高优任务"）
    query           TEXT,                                 -- 原始查询文本
    filters         JSONB NOT NULL DEFAULT '{}'::jsonb,  -- 过滤条件
    is_shared       BOOLEAN NOT NULL DEFAULT FALSE,     -- 是否共享给项目成员
    sort_order      DOUBLE PRECISION NOT NULL DEFAULT 65535,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX idx_search_bookmarks_user ON search_bookmarks(user_id, sort_order);
CREATE INDEX idx_search_bookmarks_project ON search_bookmarks(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_search_bookmarks_ws ON search_bookmarks(workspace_id);

-- updated_at 触发器
CREATE TRIGGER trg_search_bookmarks_updated_at BEFORE UPDATE ON search_bookmarks
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- RLS
ALTER TABLE search_bookmarks ENABLE ROW LEVEL SECURITY;
ALTER TABLE search_bookmarks FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON search_bookmarks
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- search_documents: 多对象统一搜索索引（issues / sprints / versions）
-- 类型字段区分来源，便于搜索结果分组/排名
-- 注：此为 Postgresql 原生方案，未来可平滑升级为 ES 方案
-- -----------------------------------------------------------------
CREATE TABLE search_documents (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    doc_type        TEXT NOT NULL CHECK (doc_type IN ('issue','sprint','version')),
    doc_id          BIGINT NOT NULL,
    title           TEXT NOT NULL,                        -- 可搜索标题
    identifier      TEXT,                                 -- 可读标识（如 YD-123）
    content         TEXT,                                 -- 可搜索内容摘要
    search_tsv      tsvector,                             -- 全文检索向量
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,  -- 类型相关元数据
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_search_documents_unique ON search_documents(workspace_id, doc_type, doc_id);
CREATE INDEX idx_search_documents_tsv ON search_documents USING GIN(search_tsv);
CREATE INDEX idx_search_documents_project ON search_documents(project_id, doc_type, updated_at DESC);
CREATE INDEX idx_search_documents_ws ON search_documents(workspace_id, doc_type);

-- RLS
ALTER TABLE search_documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE search_documents FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON search_documents
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- updated_at 触发器
CREATE TRIGGER trg_search_documents_updated_at BEFORE UPDATE ON search_documents
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -----------------------------------------------------------------
-- 刷新 search_documents 的辅助函数（事件驱动刷新，由 worker 异步调用）
-- 在 issues/sprints/versions 表变更后触发
-- -----------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_refresh_search_document()
RETURNS TRIGGER AS $$
DECLARE
    v_title TEXT;
    v_content TEXT;
    v_metadata JSONB;
BEGIN
    -- 按表名分支
    CASE TG_TABLE_NAME
        WHEN 'issues' THEN
            v_title := COALESCE(NEW.name, '');
            v_content := COALESCE(NEW.description_stripped, '');
            v_metadata := jsonb_build_object(
                'type_code', NEW.type_code,
                'state_id', NEW.state_id,
                'priority', NEW.priority
            );
            INSERT INTO search_documents (workspace_id, project_id, doc_type, doc_id, title, identifier, content, search_tsv, metadata)
            VALUES (
                NEW.workspace_id, NEW.project_id, 'issue', NEW.id,
                v_title, NEW.sequence_id::text, v_content,
                to_tsvector('simple',
                    coalesce(v_title, '') || ' ' ||
                    coalesce(v_content, '')
                ),
                v_metadata
            )
            ON CONFLICT (workspace_id, doc_type, doc_id) DO UPDATE SET
                title = EXCLUDED.title,
                identifier = EXCLUDED.identifier,
                content = EXCLUDED.content,
                search_tsv = EXCLUDED.search_tsv,
                metadata = EXCLUDED.metadata,
                updated_at = now();
        ELSE
            -- 其他类型延后处理
            RETURN NEW;
    END CASE;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 工作项创建/更新时自动同步搜索文档
CREATE TRIGGER trg_issue_search_sync
    AFTER INSERT OR UPDATE OF name, description_stripped ON issues
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NULL)
    EXECUTE FUNCTION fn_refresh_search_document();

-- 工作项软删除时清理搜索文档
CREATE OR REPLACE FUNCTION fn_cleanup_search_document()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM search_documents
    WHERE doc_type = 'issue' AND doc_id = OLD.id AND workspace_id = OLD.workspace_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_issue_search_cleanup
    AFTER UPDATE OF deleted_at ON issues
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NOT NULL)
    EXECUTE FUNCTION fn_cleanup_search_document();
