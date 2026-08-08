-- 迁移 0025：Epic 类型支持 + Module 模块体系
-- 遵循项目多租户约束：所有新表含 workspace_id NOT NULL，首列索引，开启 RLS
-- 遵循项目主键约束：id BIGINT 自增主键
-- 遵循项目时间字段约束：created_at/updated_at 为 TIMESTAMPTZ
-- 遵循项目软删除约束：deleted_at TIMESTAMPTZ NULL，唯一索引带 WHERE deleted_at IS NULL

-- ============================================================
-- 1. issues 表 type_check 约束扩展：追加 epic
-- ============================================================
ALTER TABLE issues DROP CONSTRAINT IF EXISTS issues_type_code_check;
ALTER TABLE issues ADD CONSTRAINT issues_type_code_check
    CHECK (type_code = ANY (ARRAY['epic'::text, 'requirement'::text, 'task'::text, 'defect'::text]));

COMMENT ON COLUMN issues.type_code IS '工作项类型: epic(史诗) / requirement(需求) / task(任务) / defect(缺陷)';


-- ============================================================
-- 2. modules 表（项目模块，用于分组工作项）
-- ============================================================
CREATE TABLE IF NOT EXISTS modules (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id       UUID NOT NULL DEFAULT gen_random_uuid(),
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    name            TEXT NOT NULL,
    description     TEXT,
    lead_id         BIGINT REFERENCES users(id),
    status          TEXT NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active','completed','archived')),
    start_date      DATE,
    target_date     DATE,
    sort_order      DOUBLE PRECISION NOT NULL DEFAULT 65535,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE (project_id, name) WHERE deleted_at IS NULL,
    UNIQUE (public_id) WHERE deleted_at IS NULL
);

-- modules 表 RLS 配置（对齐项目现有租户隔离策略）
ALTER TABLE modules ENABLE ROW LEVEL SECURITY;
ALTER TABLE modules FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON modules
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- modules 表索引
CREATE INDEX idx_modules_project      ON modules(workspace_id, project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_modules_project_sort ON modules(project_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_modules_lead         ON modules(lead_id) WHERE deleted_at IS NULL;

COMMENT ON TABLE modules IS '项目模块：将工作项按功能/组件分组，对标 Plane 的 Module 概念';
COMMENT ON COLUMN modules.status IS '模块状态: active(活跃) / completed(已完成) / archived(已归档)';


-- ============================================================
-- 3. module_issues 关联表（工作项 × 模块 M:N）
-- ============================================================
CREATE TABLE IF NOT EXISTS module_issues (
    module_id       BIGINT NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    issue_id        BIGINT NOT NULL,  -- 指向 issues.id（兼容分表后逻辑 ID）
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (module_id, issue_id)
);

-- module_issues 索引
CREATE INDEX idx_module_issues_issue ON module_issues(issue_id);

COMMENT ON TABLE module_issues IS '工作项-模块 M:N 关联表（一个工作项可属于多个模块，一个模块包含多个工作项）';


-- ============================================================
-- 4. updated_at 触发器（对齐项目规范）
-- ============================================================
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_modules_updated_at
    BEFORE UPDATE ON modules
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
