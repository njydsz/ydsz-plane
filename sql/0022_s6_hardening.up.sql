-- 0013_s6_hardening: S6 大厂标准加固（乐观锁 + 清单上限 + 审计追踪）
-- 变更:
--   1. versions 表增加 version 列（乐观锁 CAS）
--   2. versions 增加 checklist 长度上限约束
--   3. versions 增加 auto-increment version 的触发器
--   4. 版本相关 audit_logs 索引优化

-- -----------------------------------------------------------------
-- Step 1: versions 表增加 version 列（乐观锁）
-- -----------------------------------------------------------------
ALTER TABLE versions ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

COMMENT ON COLUMN versions.version IS '乐观锁版本号，每次 UPDATE 自增';

-- -----------------------------------------------------------------
-- Step 2: checklist 长度约束（防止 DoS 巨型 payload）
-- -----------------------------------------------------------------
ALTER TABLE versions ADD CONSTRAINT versions_checklist_limit
    CHECK (jsonb_array_length(checklist) <= 50);

-- -----------------------------------------------------------------
-- Step 3: 确保 existing 行的 version 为有效的正数
-- -----------------------------------------------------------------
UPDATE versions SET version = 1 WHERE version < 1;

-- -----------------------------------------------------------------
-- Step 4: 生成 version 自增触发器函数（项目级通用）
-- -----------------------------------------------------------------
CREATE OR REPLACE FUNCTION bump_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- -----------------------------------------------------------------
-- Step 5: 为 versions 表挂载 version 自增触发器
-- 每次 UPDATE 自动 version+1，无需应用层维护
-- -----------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_versions_bump_version ON versions;
CREATE TRIGGER trg_versions_bump_version
    BEFORE UPDATE ON versions
    FOR EACH ROW
    EXECUTE FUNCTION bump_version();

-- -----------------------------------------------------------------
-- Step 6: 版本交付快照表（发布时事务内记录，用于问题追溯）
-- -----------------------------------------------------------------
CREATE TABLE IF NOT EXISTS version_delivery_snapshots (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    version_id      BIGINT NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    progress        JSONB NOT NULL DEFAULT '{}'::jsonb,
    quality         JSONB NOT NULL DEFAULT '{}'::jsonb,
    release_notes   TEXT,
    snapshot_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_vds_version ON version_delivery_snapshots(version_id, snapshot_at DESC);
CREATE INDEX IF NOT EXISTS idx_vds_workspace ON version_delivery_snapshots(workspace_id);

ALTER TABLE version_delivery_snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE version_delivery_snapshots FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON version_delivery_snapshots
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- Step 7: 版本审计日志索引
-- -----------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_audit_logs_action_target
    ON audit_logs(action, target) WHERE action LIKE 'version.%';
