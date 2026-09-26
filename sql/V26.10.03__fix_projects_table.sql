-- ===========================================================================
-- V26.10.03 · 修复 projects / project_members 表结构缺陷
--
-- 故障现象：创建项目报"服务内部错误"（HTTP 500）
--
-- 根因分析：
--   1. projects 表缺少 color 列，但 Go ProjectService.Create 的 INSERT 包含 color
--      → PG: column "color" of relation "projects" does not exist
--   2. projects.id 无序列/默认值，Go INSERT 又未传 id
--      → PG: null value in column "id" violates not-null constraint
--   3. project_members.id 同病，Go INSERT 未传 id
--
-- 修复方案：补列 + 建序列 + 设默认值
-- ===========================================================================

-- ═══════════════════════════════════════════════════════════════════════
-- 1. projects 表 — 补 color 列
-- ═══════════════════════════════════════════════════════════════════════
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'projects' AND column_name = 'color'
    ) THEN
        ALTER TABLE projects ADD COLUMN color VARCHAR(20);
    END IF;
END$$;

-- ═══════════════════════════════════════════════════════════════════════
-- 2. projects 表 — 创建序列并设 id 默认值
-- ═══════════════════════════════════════════════════════════════════════
CREATE SEQUENCE IF NOT EXISTS projects_id_seq START 1 INCREMENT 1;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'projects' AND column_name = 'id'
          AND column_default IS NOT NULL
    ) THEN
        ALTER TABLE projects ALTER COLUMN id SET DEFAULT nextval('projects_id_seq');
    END IF;
END$$;

-- ═══════════════════════════════════════════════════════════════════════
-- 3. project_members 表 — 创建序列并设 id 默认值
-- ═══════════════════════════════════════════════════════════════════════
CREATE SEQUENCE IF NOT EXISTS project_members_id_seq START 1 INCREMENT 1;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'project_members' AND column_name = 'id'
          AND column_default IS NOT NULL
    ) THEN
        ALTER TABLE project_members ALTER COLUMN id SET DEFAULT nextval('project_members_id_seq');
    END IF;
END$$;

-- ═══════════════════════════════════════════════════════════════════════
-- 4. risk_rules 表 — 创建序列并设 id 默认值（项目创建异步初始化用）
-- ═══════════════════════════════════════════════════════════════════════
CREATE SEQUENCE IF NOT EXISTS risk_rules_id_seq START 1 INCREMENT 1;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'risk_rules' AND column_name = 'id'
          AND column_default IS NOT NULL
    ) THEN
        ALTER TABLE risk_rules ALTER COLUMN id SET DEFAULT nextval('risk_rules_id_seq');
    END IF;
END$$;

-- ═══════════════════════════════════════════════════════════════════════
-- 5. 校准序列值（避免与已有数据冲突）
-- ═══════════════════════════════════════════════════════════════════════
DO $$
DECLARE
    max_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) INTO max_id FROM projects;
    IF max_id > 0 THEN
        PERFORM setval('projects_id_seq', max_id);
    END IF;

    SELECT COALESCE(MAX(id), 0) INTO max_id FROM project_members;
    IF max_id > 0 THEN
        PERFORM setval('project_members_id_seq', max_id);
    END IF;

    SELECT COALESCE(MAX(id), 0) INTO max_id FROM risk_rules;
    IF max_id > 0 THEN
        PERFORM setval('risk_rules_id_seq', max_id);
    END IF;
END$$;
