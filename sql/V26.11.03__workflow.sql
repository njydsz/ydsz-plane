-- V26.11.03__workflow.sql — 工作流与审批
--
-- 对标 Plane / Jira Workflow：
--   - workflows            ：项目级工作流定义（状态流转规则容器）
--   - workflow_transitions ：状态流转规则（含审批配置）
--   - workflow_approvals   ：审批记录（pending / approved / rejected）

-- ---------------------------------------------------------------------------
-- workflows
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS workflows (
    id          BIGINT       PRIMARY KEY,
    project_id  BIGINT       NOT NULL,
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    is_default  BOOLEAN      NOT NULL DEFAULT FALSE,
    created_by  BIGINT,
    updated_by  BIGINT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_workflows_project_id ON workflows (project_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_workflows_project_default ON workflows (project_id, is_default) WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- workflow_transitions
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS workflow_transitions (
    id                BIGINT       PRIMARY KEY,
    workflow_id       BIGINT       NOT NULL,
    from_state_id     BIGINT       NOT NULL,
    to_state_id       BIGINT       NOT NULL,
    name              VARCHAR(200),
    requires_approval BOOLEAN      NOT NULL DEFAULT FALSE,
    approver_role     VARCHAR(50),
    approver_ids      BIGINT[],
    conditions        JSONB,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_workflow_transition_states_differ CHECK (from_state_id <> to_state_id)
);

CREATE INDEX IF NOT EXISTS idx_workflow_transitions_workflow_id ON workflow_transitions (workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_transitions_lookup
    ON workflow_transitions (workflow_id, from_state_id, to_state_id);

-- ---------------------------------------------------------------------------
-- workflow_approvals
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS workflow_approvals (
    id            BIGINT       PRIMARY KEY,
    transition_id BIGINT       NOT NULL,
    work_item_id  BIGINT       NOT NULL,
    work_item_type VARCHAR(20) NOT NULL,
    requester_id  BIGINT       NOT NULL,
    approver_id   BIGINT       NOT NULL,
    status        VARCHAR(20)  NOT NULL DEFAULT 'pending',
    comment       TEXT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    resolved_at   TIMESTAMPTZ,
    CONSTRAINT chk_workflow_approval_status CHECK (status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT chk_workflow_approval_type CHECK (work_item_type IN ('requirement', 'task', 'defect'))
);

CREATE INDEX IF NOT EXISTS idx_workflow_approvals_transition ON workflow_approvals (transition_id);
CREATE INDEX IF NOT EXISTS idx_workflow_approvals_work_item ON workflow_approvals (work_item_id);
CREATE INDEX IF NOT EXISTS idx_workflow_approvals_pending
    ON workflow_approvals (approver_id, status) WHERE status = 'pending';

COMMENT ON TABLE workflows IS '项目级工作流定义';
COMMENT ON TABLE workflow_transitions IS '工作流状态流转规则（含审批门禁配置）';
COMMENT ON TABLE workflow_approvals IS '工作项流转审批记录';
