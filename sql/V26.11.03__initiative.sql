-- ===========================================================================
-- V26.11.03 · 倡议（Initiative）
-- 对标：Plane / Jira Advanced Roadmaps 的 Initiative 层级
-- ===========================================================================
-- 背景：
--   倡议是跨项目的战略性目标，可关联多个项目与工作项（需求/任务/缺陷），
--   并通过聚合这些项目/迭代/版本/工作项的进度来向上汇报整体健康度。
--
-- 三张表：
--   1. initiatives            —— 主表（id 由应用层 genID 生成，无序列默认值）
--   2. initiative_projects    —— 倡议 ↔ 项目 多对多
--   3. initiative_work_items  —— 倡议 ↔ 工作项 多对多（带工作项类型）
--
-- 设计要点：
--   - 软删除：initiatives.archived_at 标记归档（非物理删除）；
--   - 关联表使用复合主键防重；
--   - 索引覆盖 workspace_id / status / priority / owner_id 等高频过滤维度。
-- ===========================================================================

-- 1. initiatives 主表。
CREATE TABLE IF NOT EXISTS initiatives (
    id            BIGINT PRIMARY KEY,
    workspace_id  BIGINT NOT NULL DEFAULT 0,
    title         VARCHAR(200) NOT NULL,
    description   TEXT,
    status        VARCHAR(20) NOT NULL DEFAULT 'on_track',
    priority      VARCHAR(20) NOT NULL DEFAULT 'medium',
    start_date    DATE,
    target_date   DATE,
    owner_id      BIGINT,
    created_by    BIGINT NOT NULL DEFAULT 0,
    archived_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now(),

    -- 状态枚举约束（on_track / at_risk / off_track / completed）
    CONSTRAINT chk_initiatives_status
        CHECK (status IN ('on_track', 'at_risk', 'off_track', 'completed')),
    -- 优先级枚举约束（urgent / high / medium / low / none）
    CONSTRAINT chk_initiatives_priority
        CHECK (priority IN ('urgent', 'high', 'medium', 'low', 'none'))
);

-- 2. initiative_projects 关联表。
CREATE TABLE IF NOT EXISTS initiative_projects (
    initiative_id  BIGINT NOT NULL,
    project_id     BIGINT NOT NULL,
    created_at     TIMESTAMPTZ DEFAULT now(),

    PRIMARY KEY (initiative_id, project_id)
);

-- 3. initiative_work_items 关联表。
CREATE TABLE IF NOT EXISTS initiative_work_items (
    initiative_id   BIGINT NOT NULL,
    work_item_id    BIGINT NOT NULL,
    work_item_type  VARCHAR(20) NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT now(),

    PRIMARY KEY (initiative_id, work_item_id, work_item_type),
    -- 工作项类型枚举约束（requirement / task / defect）
    CONSTRAINT chk_initiative_work_items_type
        CHECK (work_item_type IN ('requirement', 'task', 'defect'))
);

-- ===========================================================================
-- 索引
-- ===========================================================================

-- initiatives：按工作空间过滤（最高频查询）
CREATE INDEX IF NOT EXISTS idx_initiatives_workspace_id
    ON initiatives (workspace_id)
    WHERE archived_at IS NULL;

-- initiatives：按状态过滤
CREATE INDEX IF NOT EXISTS idx_initiatives_status
    ON initiatives (status)
    WHERE archived_at IS NULL;

-- initiatives：按优先级过滤
CREATE INDEX IF NOT EXISTS idx_initiatives_priority
    ON initiatives (priority)
    WHERE archived_at IS NULL;

-- initiatives：按负责人过滤
CREATE INDEX IF NOT EXISTS idx_initiatives_owner_id
    ON initiatives (owner_id)
    WHERE archived_at IS NULL AND owner_id IS NOT NULL;

-- initiatives：按目标日期排序（看板/时间线视图）
CREATE INDEX IF NOT EXISTS idx_initiatives_target_date
    ON initiatives (target_date)
    WHERE archived_at IS NULL;

-- initiatives：工作空间 + 状态组合（列表页最常见过滤组合）
CREATE INDEX IF NOT EXISTS idx_initiatives_workspace_status
    ON initiatives (workspace_id, status)
    WHERE archived_at IS NULL;

-- initiative_projects：按项目反查倡议
CREATE INDEX IF NOT EXISTS idx_initiative_projects_project_id
    ON initiative_projects (project_id);

-- initiative_work_items：按工作项反查倡议
CREATE INDEX IF NOT EXISTS idx_initiative_work_items_work_item
    ON initiative_work_items (work_item_id, work_item_type);
