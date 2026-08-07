-- 0008_search.down.sql: Revert search schema changes
DROP TRIGGER IF EXISTS trg_issue_search_cleanup ON issues;
DROP FUNCTION IF EXISTS fn_cleanup_search_document();
DROP TRIGGER IF EXISTS trg_issue_search_sync ON issues;
DROP FUNCTION IF EXISTS fn_refresh_search_document();
DROP TRIGGER IF EXISTS trg_search_documents_updated_at ON search_documents;
DROP TABLE IF EXISTS search_documents;
DROP TRIGGER IF EXISTS trg_search_bookmarks_updated_at ON search_bookmarks;
DROP TABLE IF EXISTS search_bookmarks;
DROP TABLE IF EXISTS search_history;
DROP INDEX IF EXISTS idx_issues_search_tsv;
ALTER TABLE issues DROP COLUMN IF EXISTS search_tsv;
