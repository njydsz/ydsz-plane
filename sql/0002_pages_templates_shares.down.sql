-- 0002 down: 回滚文档模板与公开分享链接表

BEGIN;

DROP TABLE IF EXISTS page_shares;
DROP TABLE IF EXISTS page_templates;

COMMIT;
