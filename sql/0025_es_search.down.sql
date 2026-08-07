-- 0025_es_search.down: 回滚 ES 搜索升级
DROP TRIGGER IF EXISTS trg_issue_es_sync ON issues;
DROP TRIGGER IF EXISTS trg_sprint_es_sync ON sprints;
DROP TRIGGER IF EXISTS trg_version_es_sync ON versions;
DROP FUNCTION IF EXISTS fn_mark_es_sync();
DROP FUNCTION IF EXISTS fn_es_reconcile(BIGINT, INT);
DROP TRIGGER IF EXISTS trg_search_suggestion_update ON search_history;
DROP FUNCTION IF EXISTS fn_update_search_suggestion();
DROP TABLE IF EXISTS search_suggestions;
ALTER TABLE search_documents DROP COLUMN IF EXISTS es_synced_at;
DROP TABLE IF EXISTS es_sync_log;
