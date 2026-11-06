-- V26.11.03__conditional_grants.sql — 条件授权（属性基授权，ABAC）
--
-- 对标 Plane Enterprise 的 Conditional Grants：在角色权限之上叠加
-- 基于属性的约束（如「只能编辑自己创建的工作项」「只能访问指定项目」
-- 「只能看到某状态分组的工作项」）。
--
-- 设计：
--   - grant_rules 定义规则（角色 + 权限 + 条件表达式）
--   - 条件表达式为受限 JSON（仅允许字段比较/逻辑组合），由服务端求值
--   - 规则按 workspace_id 隔离，project_id 可选（NULL = 全工作空间生效）

-- ---------------------------------------------------------------------------
-- grant_rules
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS grant_rules (
    id           BIGINT       PRIMARY KEY,
    workspace_id BIGINT       NOT NULL,
    name         VARCHAR(200) NOT NULL,
    description  TEXT,
    -- 生效角色（可多选，空数组 = 全部角色）
    role_slugs   TEXT[]       NOT NULL DEFAULT '{}',
    -- 授予的权限（在满足条件时追加）
    permissions  TEXT[]       NOT NULL DEFAULT '{}',
    -- 受限的 scope（限定生效的领域，可选）
    scope        VARCHAR(50),
    -- 条件表达式（JSON，见下方约束说明）
    condition    JSONB        NOT NULL,
    -- 是否启用
    enabled      BOOLEAN      NOT NULL DEFAULT TRUE,
    priority     INTEGER      NOT NULL DEFAULT 100,
    created_by   BIGINT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_grant_rules_workspace ON grant_rules (workspace_id, enabled);
CREATE INDEX IF NOT EXISTS idx_grant_rules_scope ON grant_rules (workspace_id, scope) WHERE enabled;

COMMENT ON TABLE grant_rules IS '条件授权规则（角色 + 权限 + 条件表达式）';
COMMENT ON COLUMN grant_rules.condition IS
    '条件表达式 JSON：{"op":"and|or|not|eq|ne|in|contains|gt|lt","field":"...","value":...,"left":...,"right":...}';

-- 权限点白名单（与 auth 包常量对齐，防御注入）
-- 说明：服务端在写入时校验 permissions 必须属于已注册权限点。
