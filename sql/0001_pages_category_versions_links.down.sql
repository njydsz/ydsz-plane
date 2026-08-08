-- 0001 down: 回滚页面模块增强变更

BEGIN;

DROP TABLE IF EXISTS document_links;
DROP TABLE IF EXISTS document_versions;
ALTER TABLE pages DROP COLUMN IF EXISTS category;

COMMIT;
