-- V26.11.03__work_item_type.sql — 自定义工作项类型（Custom Work Item Types）
--
-- 对标 Plane Issue Type：项目可定义自定义工作项类型（Story / Bug / ...），
-- 每种类型可配置一组自定义属性（字段），工作项按类型携带属性值。
--
--   - work_item_types            ：项目级工作项类型定义
--   - work_item_type_properties  ：类型下的自定义属性定义
--   - work_item_custom_values    ：工作项的自定义属性值（EAV 模型）

-- ---------------------------------------------------------------------------
-- work_item_types
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS work_item_types (
    id          BIGINT       PRIMARY KEY,
    project_id  BIGINT       NOT NULL,
    name        VARCHAR(200) NOT NULL,
    code        VARCHAR(50)  NOT NULL,
    color       VARCHAR(20)  NOT NULL DEFAULT '#6B7280',
    icon        VARCHAR(100) NOT NULL DEFAULT '',
    description TEXT         NOT NULL DEFAULT '',
    level       SMALLINT     NOT NULL DEFAULT 0,          -- 0=任务级, 1=史诗级
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    is_default  BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    archived_at TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,
    -- 每个项目内 code 唯一（软删除/归档记录仍占用唯一性）
    CONSTRAINT uq_work_item_types_project_code UNIQUE (project_id, code),
    -- level 仅允许 0 / 1
    CONSTRAINT chk_work_item_type_level CHECK (level IN (0, 1))
);

CREATE INDEX IF NOT EXISTS idx_work_item_types_project ON work_item_types (project_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_work_item_types_project_active ON work_item_types (project_id, is_active)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_work_item_types_code ON work_item_types (project_id, code)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- work_item_type_properties
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS work_item_type_properties (
    id         BIGINT       PRIMARY KEY,
    type_id    BIGINT       NOT NULL,
    name       VARCHAR(200) NOT NULL,
    code       VARCHAR(50)  NOT NULL,
    field_type VARCHAR(20)  NOT NULL,                       -- text/number/select/multiselect/date/checkbox/url
    options    JSONB,                                       -- select/multiselect 的候选项
    required   BOOLEAN      NOT NULL DEFAULT FALSE,
    sort_order INTEGER      NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    -- 同一类型内属性 code 唯一
    CONSTRAINT uq_work_item_type_prop_code UNIQUE (type_id, code),
    CONSTRAINT chk_work_item_type_field_type CHECK (
        field_type IN ('text', 'number', 'select', 'multiselect', 'date', 'checkbox', 'url')
    )
);

CREATE INDEX IF NOT EXISTS idx_work_item_type_properties_type ON work_item_type_properties (type_id);

-- ---------------------------------------------------------------------------
-- work_item_custom_values
-- ---------------------------------------------------------------------------
-- EAV 模型：每个工作项对每个属性一行，value 统一存 TEXT，由 field_type 解释。
CREATE TABLE IF NOT EXISTS work_item_custom_values (
    id              BIGINT       PRIMARY KEY,
    property_id     BIGINT       NOT NULL,
    work_item_id    BIGINT       NOT NULL,
    work_item_type  VARCHAR(50)  NOT NULL,                  -- requirement/task/defect/自定义类型 code
    value           TEXT         NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    -- 同一工作项对同一属性仅一行
    CONSTRAINT uq_work_item_custom_value UNIQUE (property_id, work_item_id)
);

CREATE INDEX IF NOT EXISTS idx_work_item_custom_values_work_item ON work_item_custom_values (work_item_id);
CREATE INDEX IF NOT EXISTS idx_work_item_custom_values_type ON work_item_custom_values (work_item_type);

COMMENT ON TABLE work_item_types IS '项目级自定义工作项类型定义';
COMMENT ON TABLE work_item_type_properties IS '工作项类型的自定义属性定义';
COMMENT ON TABLE work_item_custom_values IS '工作项的自定义属性值（EAV）';
