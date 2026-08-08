-- 0001: 页面模块增强 — 添加 category 字段、版本历史和文档关联表
--
-- 变更内容:
--   1. pages 表添加 category 字段（PRD / design / api / test / checklist）
--   2. 创建 document_versions 表（版本快照）
--   3. 创建 document_links 表（关联 Issue / Sprint / Version 等）

BEGIN;

-- 1. 添加 category 字段到 pages 表（使用 TEXT 允许任意扩展，应用层校验枚举）
ALTER TABLE pages ADD COLUMN IF NOT EXISTS category TEXT;

COMMENT ON COLUMN pages.category IS '文档分类枚举：PRD / design / api / test / checklist';

-- 2. 创建 document_versions 版本快照表
CREATE TABLE IF NOT EXISTS document_versions (
    id              BIGSERIAL PRIMARY KEY,
    page_id         BIGINT NOT NULL,
    version_number  INT NOT NULL DEFAULT 1,
    content_md      TEXT,
    content_html    TEXT,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_doc_version_page FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_doc_versions_page ON document_versions(page_id, version_number DESC);

COMMENT ON TABLE document_versions IS '文档版本快照表';
COMMENT ON COLUMN document_versions.page_id IS '关联页面 FK';
COMMENT ON COLUMN document_versions.version_number IS '版本序号（从 1 开始递增）';
COMMENT ON COLUMN document_versions.content_md IS '纯文本内容（用于 diff / 搜索）';
COMMENT ON COLUMN document_versions.content_html IS '富文本 HTML 内容（用于渲染）';

-- 3. 创建 document_links 文档关联表
CREATE TABLE IF NOT EXISTS document_links (
    id              BIGSERIAL PRIMARY KEY,
    page_id         BIGINT NOT NULL,
    linkable_type   TEXT NOT NULL,
    linkable_id     BIGINT NOT NULL,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_doc_link_page FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_doc_links_page ON document_links(page_id);
CREATE INDEX IF NOT EXISTS idx_doc_links_target ON document_links(linkable_type, linkable_id);

COMMENT ON TABLE document_links IS '文档关联表（关联 Issue / Sprint / Version 等）';
COMMENT ON COLUMN document_links.page_id IS '关联页面 FK';
COMMENT ON COLUMN document_links.linkable_type IS '关联实体类型：issue / sprint / version';
COMMENT ON COLUMN document_links.linkable_id IS '关联实体 ID';

COMMIT;
