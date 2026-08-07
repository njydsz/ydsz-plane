-- 0018_project_templates: 项目模板支撑 (Sprint 12 — v0.2)
-- 用于记录项目创建时使用的预设模板类型，便于后续复制/复用。

BEGIN;

-- projects 表新增 template 列（默认 'generic' 通用模板）
ALTER TABLE projects ADD COLUMN IF NOT EXISTS template TEXT NOT NULL DEFAULT 'generic'
    CHECK (template IN ('agile', 'waterfall', 'generic'));

-- 索引（按工作空间 + 模板类型过滤查询）
CREATE INDEX IF NOT EXISTS idx_projects_template ON projects(workspace_id, template) WHERE deleted_at IS NULL;

COMMIT;
