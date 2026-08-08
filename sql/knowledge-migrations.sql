-- ============================================================================
-- knowledge 域迁移 — 全文检索（PostgreSQL tsvector + GIN 索引）
-- ============================================================================
-- 前置：ydsz-plane-init.sql 已创建 knowledge_pages 表。
-- 本文件幂等，可重复执行（IF NOT EXISTS / ADD COLUMN IF NOT EXISTS）。

-- 1) tsv 列（如果 CREATE TABLE 中未定义则补充）
ALTER TABLE public.knowledge_pages ADD COLUMN IF NOT EXISTS tsv tsvector;

-- 2) 自动维护 tsv 的触发器（INSERT / UPDATE of title, content_md）
CREATE OR REPLACE FUNCTION public.knowledge_pages_tsv_trigger_fn() RETURNS trigger AS $$
BEGIN
  NEW.tsv := to_tsvector('simple', coalesce(NEW.title,'') || ' ' || coalesce(NEW.content_md,''));
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS knowledge_pages_tsv_trigger ON public.knowledge_pages;
CREATE TRIGGER knowledge_pages_tsv_trigger
  BEFORE INSERT OR UPDATE OF title, content_md ON public.knowledge_pages
  FOR EACH ROW EXECUTE FUNCTION public.knowledge_pages_tsv_trigger_fn();

-- 3) GIN 索引加速 websearch_to_tsquery / @@ 匹配
CREATE INDEX IF NOT EXISTS idx_kp_tsv ON public.knowledge_pages USING GIN (tsv);

-- 4) 一次性回填（新表未触发时执行；注释状态，按需启用）
-- UPDATE public.knowledge_pages
--   SET tsv = to_tsvector('simple', coalesce(title,'') || ' ' || coalesce(content_md,''))
--   WHERE tsv IS NULL;
