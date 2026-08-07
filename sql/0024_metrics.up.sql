-- 0016_metrics: 效能度量基础 (Sprint 11 — M6)
-- 参考 docs/architecture/11-仪表盘与效能度量设计.md §2 效能度量
-- DORA: https://dora.dev/devops-capabilities/

-- -----------------------------------------------------------------
-- metric_snapshots: 每日效能指标快照（幂等 upsert，口径冻结）
-- granularity: daily | sprint | version
-- -----------------------------------------------------------------
CREATE TABLE metric_snapshots (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    granularity     TEXT NOT NULL CHECK (granularity IN ('daily','sprint','version')),
    ref_id          BIGINT,                  -- sprint_id 或 version_id
    metric          TEXT NOT NULL,           -- 指标名: velocity | lead_time_p50 | lead_time_p85 | ...
    value           NUMERIC(12,4),           -- 指标值（NULL 表示不可用）
    -- 多维度分解: { "by_module": {"支付": 5, "订单": 3}, "by_member": { ... } }
    dimensions      JSONB NOT NULL DEFAULT '{}'::jsonb,
    -- 元数据
    snapshot_date   DATE NOT NULL,           -- 快照日期（UTC）
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_metric_snapshots_project ON metric_snapshots(workspace_id, project_id, granularity, ref_id, metric, snapshot_date) WHERE project_id IS NOT NULL;
CREATE UNIQUE INDEX idx_metric_snapshots_workspace ON metric_snapshots(workspace_id, granularity, metric, snapshot_date) WHERE project_id IS NULL;
-- 常用查询模式: WHERE metric=X AND snapshot_date BETWEEN ...
CREATE INDEX idx_metric_snap_lookup ON metric_snapshots(project_id, metric, snapshot_date);
CREATE INDEX idx_metric_snap_ws ON metric_snapshots(workspace_id, metric, snapshot_date DESC);
CREATE INDEX idx_metric_snap_date ON metric_snapshots(snapshot_date);
-- 覆盖索引：避免回表查 dimensions
CREATE INDEX idx_metric_snap_covering ON metric_snapshots(project_id, metric, snapshot_date, value);

ALTER TABLE metric_snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE metric_snapshots FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON metric_snapshots
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- deployment_events: DORA 部署事件（CI/CD 推送）
-- -----------------------------------------------------------------
CREATE TABLE deployment_events (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    -- 部署标识
    deployment_id   TEXT,                    -- 外部部署系统 ID（去重用）
    env             TEXT NOT NULL DEFAULT 'production'
                    CHECK (env IN ('development','staging','production','testing')),
    status          TEXT NOT NULL CHECK (status IN ('success','failed','rolled_back')),
    commit_sha      TEXT,
    -- 时间（DORA 计算所需）
    started_at      TIMESTAMPTZ,             -- 部署开始
    deployed_at     TIMESTAMPTZ,             -- 部署完成（生产环境时间）
    -- 来源: github_actions | gitlab_ci | jenkins | webhook | manual
    source          TEXT NOT NULL DEFAULT 'webhook',
    -- 关联元数据
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    -- 幂等键：(deployment_id, env, project_id)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_deployment_events_project ON deployment_events(project_id, deployed_at DESC) WHERE project_id IS NOT NULL;
CREATE INDEX idx_deployment_events_ws ON deployment_events(workspace_id, deployed_at DESC);
-- 幂等唯一约束
CREATE UNIQUE INDEX uq_deployment_events_idempotent ON deployment_events(deployment_id, env, project_id)
    WHERE deployment_id IS NOT NULL;

ALTER TABLE deployment_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE deployment_events FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON deployment_events
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- metric_adjustments: 管理员数据校准记录（不覆盖原快照，叠加修正）
-- -----------------------------------------------------------------
CREATE TABLE metric_adjustments (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    snapshot_id     BIGINT REFERENCES metric_snapshots(id) ON DELETE SET NULL,
    metric          TEXT NOT NULL,
    snapshot_date   DATE NOT NULL,
    original_value  NUMERIC(12,4),
    adjusted_value  NUMERIC(12,4) NOT NULL,
    reason          TEXT NOT NULL,
    adjusted_by     BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_metric_adjustments_ws ON metric_adjustments(workspace_id, snapshot_date DESC);

ALTER TABLE metric_adjustments ENABLE ROW LEVEL SECURITY;
ALTER TABLE metric_adjustments FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON metric_adjustments
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint);

CREATE TRIGGER trg_metric_adjustments_updated_at BEFORE UPDATE ON metric_adjustments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
