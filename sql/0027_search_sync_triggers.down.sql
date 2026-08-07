-- 0027_search_sync_triggers: 回滚。
-- 删除 sprints / versions 触发器，并将函数定义还原为 0008 的"仅 issues"版本。

DROP TRIGGER IF EXISTS trg_sprint_search_sync ON sprints;
DROP TRIGGER IF EXISTS trg_sprint_search_cleanup ON sprints;
DROP TRIGGER IF EXISTS trg_version_search_sync ON versions;
DROP TRIGGER IF EXISTS trg_version_search_cleanup ON versions;

-- 还原 fn_refresh_search_document 为 0008 的"仅 issues"实现。
CREATE OR REPLACE FUNCTION fn_refresh_search_document()
RETURNS TRIGGER AS $$
DECLARE
    v_title TEXT;
    v_content TEXT;
    v_metadata JSONB;
BEGIN
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
            RETURN NEW;
    END CASE;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 还原 fn_cleanup_search_document 为 0008 的"仅 issue"实现。
CREATE OR REPLACE FUNCTION fn_cleanup_search_document()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM search_documents
    WHERE doc_type = 'issue' AND doc_id = OLD.id AND workspace_id = OLD.workspace_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;