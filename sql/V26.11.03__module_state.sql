-- V26.11.03__module_state.sql — 模块状态机（对标 Plane Module States）
--
-- Plane 模块状态：Backlog / Planned / In Progress / Paused / Completed / Cancelled
-- 本迁移新增 workflow_status 列与索引，不改动既有 status（entity_status：active/archived）。

ALTER TABLE modules ADD COLUMN IF NOT EXISTS workflow_status VARCHAR(20) NOT NULL DEFAULT 'backlog';

-- 修正历史值（防御性）
UPDATE modules SET workflow_status = 'backlog' WHERE workflow_status IS NULL;

ALTER TABLE modules DROP CONSTRAINT IF EXISTS chk_modules_workflow_status;
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_modules_workflow_status'
    ) THEN
        ALTER TABLE modules ADD CONSTRAINT chk_modules_workflow_status
            CHECK (workflow_status IN ('backlog','planned','in_progress','paused','completed','cancelled'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_modules_workflow_status
    ON modules (project_id, workflow_status) WHERE NOT deleted;

COMMENT ON COLUMN modules.workflow_status IS '模块工作流状态（backlog/planned/in_progress/paused/completed/cancelled）';
