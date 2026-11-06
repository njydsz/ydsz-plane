-- V26.11.03__project_code.sql — Projects-as-Code（YAML 定义项目）
--
-- 对标 Plane Compose：用户以 YAML 声明式定义项目（状态 / 标签 / 模块 / 成员 / 设置），
-- 可纳入 Git 版本管理，并通过 API 或 CLI 拉取（pull/export）与.apply（apply）。
--
-- 本迁移只新增「项目配置存储」表，不修改既有 projects / states / labels / modules 表结构：
--   - project_code_configs：保存原始 YAML 文本 + 解析后的规范 JSON + sha256 + 版本号。
--     解析 / 校验 / diff / apply 的规范结构由应用层（internal/application/projectcode）负责，
--     DB 仅作持久化与版本追踪。

-- ---------------------------------------------------------------------------
-- project_code_configs
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS project_code_configs (
    id           BIGINT       PRIMARY KEY,
    workspace_id BIGINT       NOT NULL,
    project_id   BIGINT,                -- 关联的既有项目（apply 后回填 / 或创建时绑定）
    name         VARCHAR(200) NOT NULL, -- 配置名（通常等于 spec.name）
    spec_key     VARCHAR(64),           -- spec.key（项目短标识）
    version      VARCHAR(20)  NOT NULL DEFAULT '1.0', -- spec version
    yaml         TEXT         NOT NULL, -- 原始 YAML 文本
    spec_json    JSONB        NOT NULL, -- 解析后的规范结构（ProjectSpec）
    sha256       VARCHAR(64)  NOT NULL, -- yaml 文本的 sha256（幂等 / 变更检测）
    config_version INTEGER    NOT NULL DEFAULT 1, -- 保存次数（每次 SaveConfig 自增）
    created_by   BIGINT       NOT NULL DEFAULT 0,
    updated_by   BIGINT       NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted      BOOLEAN      NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_project_code_configs_workspace
    ON project_code_configs (workspace_id, deleted, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_project_code_configs_project
    ON project_code_configs (project_id) WHERE project_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_project_code_configs_sha
    ON project_code_configs (workspace_id, sha256);

COMMENT ON TABLE  project_code_configs                IS 'Projects-as-Code 项目配置（YAML + 解析 JSON + 版本）';
COMMENT ON COLUMN project_code_configs.yaml           IS '原始 YAML 定义文本';
COMMENT ON COLUMN project_code_configs.spec_json      IS '解析后的规范结构（ProjectSpec JSON）';
COMMENT ON COLUMN project_code_configs.sha256         IS 'yaml 文本 sha256，用于变更检测与幂等保存';
COMMENT ON COLUMN project_code_configs.config_version IS '同一配置的保存版本号（自增）';
