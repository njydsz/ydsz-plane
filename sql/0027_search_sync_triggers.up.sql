-- 0027_search_sync_triggers: 为 sprints / versions 补齐 search_documents 自动索引触发器。
--
-- 背景（P0 缺陷）:
--   0008 迁移仅为 issues 建立了 fn_refresh_search_document 的索引触发器，
--   sprints / versions 从不会被自动索引，存量数据也无回填。
--   本迁移:
--     1. 用 CREATE OR REPLACE FUNCTION 覆盖 fn_refresh_search_document（完整含 issues/sprints/versions 三分支），
--        以便新老环境（无论 0008 是否已应用）都能获得完整实现 —— 触发器共用同一函数。
--     2. 覆盖 fn_cleanup_search_document，按 TG_TABLE_NAME 分发清理三个 doc_type。
--     3. 为 sprints / versions 建立 AFTER 触发器（创建/更新 + 软删除清理）。
--
-- 注: 已应用过 0008 的环境不会重跑 0008，因此这里用 CREATE OR REPLACE 覆盖函数定义。

-- -----------------------------------------------------------------
-- 覆盖刷新函数：issues / sprints / versions 三分支
-- -----------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_refresh_search_document()
RETURNS TRIGGER AS $$
DECLARE
    v_title TEXT;
    v_content TEXT;
    v_metadata JSONB;
    v_doc_type TEXT;
BEGIN
    -- 按表名分发
    CASE TG_TABLE_NAME
        WHEN 'issues' THEN
            v_doc_type := 'issue';
            v_title := COALESCE(NEW.name, '');
            v_content := COALESCE(NEW.description_stripped, '');
            v_metadata := jsonb_build_object(
                'type_code', NEW.type_code,
                'state_id', NEW.state_id,
                'priority', NEW.priority
            );
        WHEN 'sprints' THEN
            v_doc_type := 'sprint';
            v_title := COALESCE(NEW.name, '');
            v_content := COALESCE(NEW.goal, '');
            v_metadata := jsonb_build_object('status', NEW.status);
        WHEN 'versions' THEN
            v_doc_type := 'version';
            v_title := COALESCE(NEW.name, '');
            v_content := COALESCE(NEW.description, '');
            v_metadata := jsonb_build_object('status', NEW.status);
        ELSE
            RETURN NEW;
    END CASE;

    INSERT INTO search_documents (workspace_id, project_id, doc_type, doc_id, title, identifier, content, search_tsv, metadata)
    VALUES (
        NEW.workspace_id, NEW.project_id, v_doc_type, NEW.id,
        v_title,
        -- issues 用 sequence_id 作为可读标识；versions 用 semver；sprints 无自然标识留空
        CASE WHEN TG_TABLE_NAME = 'issues' THEN NEW.sequence_id::text
             WHEN TG_TABLE_NAME = 'versions' THEN NEW.semver
             ELSE NULL END,
        v_content,
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
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- -----------------------------------------------------------------
-- 覆盖软删除清理函数：issues / sprints / versions 三分支
-- -----------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_cleanup_search_document()
RETURNS TRIGGER AS $$
DECLARE
    v_doc_type TEXT;
BEGIN
    CASE TG_TABLE_NAME
        WHEN 'issues'   THEN v_doc_type := 'issue';
        WHEN 'sprints'  THEN v_doc_type := 'sprint';
        WHEN 'versions' THEN v_doc_type := 'version';
        ELSE RETURN OLD;
    END CASE;
    DELETE FROM search_documents
    WHERE doc_type = v_doc_type AND doc_id = OLD.id AND workspace_id = OLD.workspace_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- -----------------------------------------------------------------
-- sprints 触发器（镜像 issues）
-- -----------------------------------------------------------------
CREATE TRIGGER trg_sprint_search_sync
    AFTER INSERT OR UPDATE OF name, goal ON sprints
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NULL)
    EXECUTE FUNCTION fn_refresh_search_document();

CREATE TRIGGER trg_sprint_search_cleanup
    AFTER UPDATE OF deleted_at ON sprints
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NOT NULL)
    EXECUTE FUNCTION fn_cleanup_search_document();

-- -----------------------------------------------------------------
-- versions 触发器（镜像 issues）
-- -----------------------------------------------------------------
CREATE TRIGGER trg_version_search_sync
    AFTER INSERT OR UPDATE OF name, description ON versions
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NULL)
    EXECUTE FUNCTION fn_refresh_search_document();

CREATE TRIGGER trg_version_search_cleanup
    AFTER UPDATE OF deleted_at ON versions
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NOT NULL)
    EXECUTE FUNCTION fn_cleanup_search_document();