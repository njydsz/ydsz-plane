-- 大表差异化归档支持：给三类工作项表添加归档字段
ALTER TABLE task ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ NULL;
ALTER TABLE requirement ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ NULL;
ALTER TABLE defect ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ NULL;

-- 创建归档状态索引，方便查询未归档数据
CREATE INDEX idx_task_archived ON task(archived_at) WHERE archived_at IS NOT NULL;
CREATE INDEX idx_requirement_archived ON requirement(archived_at) WHERE archived_at IS NOT NULL;
CREATE INDEX idx_defect_archived ON defect(archived_at) WHERE archived_at IS NOT NULL;
