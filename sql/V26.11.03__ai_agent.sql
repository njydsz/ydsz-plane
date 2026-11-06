-- V26.11.03__ai_agent.sql — Plane AI 风格 Agent 能力
--
-- 对标 Plane AI 的三种模式（Ask / Build / Autopilot）+ AI Memory + Agent 定义。

-- ---------------------------------------------------------------------------
-- ai_agents
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ai_agents (
    id                    BIGINT       PRIMARY KEY,
    workspace_id          BIGINT       NOT NULL,
    name                  VARCHAR(200) NOT NULL,
    description           TEXT,
    instructions          TEXT,
    triggers              TEXT[]       NOT NULL DEFAULT '{}',
    skills                TEXT[]       NOT NULL DEFAULT '{}',
    scope_project_ids     BIGINT[]     NOT NULL DEFAULT '{}',
    scope_work_item_types TEXT[]       NOT NULL DEFAULT '{}',
    model                 VARCHAR(100) NOT NULL DEFAULT 'gpt-4o-mini',
    temperature           DOUBLE PRECISION NOT NULL DEFAULT 0.7,
    daily_credit_limit    INTEGER      NOT NULL DEFAULT 100,
    used_credits_today    INTEGER      NOT NULL DEFAULT 0,
    status                VARCHAR(20)  NOT NULL DEFAULT 'active',
    created_by            BIGINT       NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_ai_agent_status CHECK (status IN ('active','paused','archived'))
);

CREATE INDEX IF NOT EXISTS idx_ai_agents_workspace ON ai_agents (workspace_id);
CREATE INDEX IF NOT EXISTS idx_ai_agents_status ON ai_agents (workspace_id, status);

-- ---------------------------------------------------------------------------
-- ai_memory
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ai_memory (
    id           BIGINT       PRIMARY KEY,
    workspace_id BIGINT       NOT NULL,
    scope        VARCHAR(20)  NOT NULL,
    scope_id     BIGINT       NOT NULL DEFAULT 0,
    key          VARCHAR(200) NOT NULL DEFAULT '',
    content      TEXT         NOT NULL,
    source       VARCHAR(20)  NOT NULL DEFAULT 'user',
    created_by   BIGINT       NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_ai_memory_scope CHECK (scope IN ('user','project','workspace')),
    UNIQUE (workspace_id, scope, scope_id, key)
);

CREATE INDEX IF NOT EXISTS idx_ai_memory_lookup ON ai_memory (workspace_id, scope, scope_id);

-- ---------------------------------------------------------------------------
-- ai_conversations
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ai_conversations (
    id           BIGINT      PRIMARY KEY,
    workspace_id BIGINT      NOT NULL,
    agent_id     BIGINT      NOT NULL DEFAULT 0,
    user_id      BIGINT      NOT NULL,
    mode         VARCHAR(20) NOT NULL,
    title        VARCHAR(300) NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_ai_conversation_mode CHECK (mode IN ('ask','build','autopilot'))
);

CREATE INDEX IF NOT EXISTS idx_ai_conversations_user ON ai_conversations (workspace_id, user_id, updated_at DESC);

-- ---------------------------------------------------------------------------
-- ai_messages
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ai_messages (
    id              BIGINT      PRIMARY KEY,
    conversation_id BIGINT      NOT NULL,
    role            VARCHAR(20) NOT NULL,
    content         TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_ai_message_role CHECK (role IN ('user','assistant','system'))
);

CREATE INDEX IF NOT EXISTS idx_ai_messages_conversation ON ai_messages (conversation_id, created_at);

-- ---------------------------------------------------------------------------
-- ai_actions
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ai_actions (
    id            BIGINT      PRIMARY KEY,
    message_id    BIGINT      NOT NULL,
    type          VARCHAR(40) NOT NULL,
    payload       TEXT        NOT NULL DEFAULT '{}',
    status        VARCHAR(20) NOT NULL DEFAULT 'pending',
    result        TEXT,
    error_message TEXT,
    executed_by   BIGINT,
    executed_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_ai_action_status CHECK (status IN ('pending','approved','rejected','executed','failed'))
);

CREATE INDEX IF NOT EXISTS idx_ai_actions_message ON ai_actions (message_id);
CREATE INDEX IF NOT EXISTS idx_ai_actions_pending ON ai_actions (status) WHERE status = 'pending';

COMMENT ON TABLE ai_agents IS 'AI 助手定义（身份/指令/触发器/技能/额度）';
COMMENT ON TABLE ai_memory IS 'AI 持久化记忆（用户/项目/工作空间作用域）';
COMMENT ON TABLE ai_conversations IS 'AI 会话（Ask/Build/Autopilot）';
COMMENT ON TABLE ai_messages IS 'AI 会话消息';
COMMENT ON TABLE ai_actions IS 'AI 动作（Build 模式待确认 / Autopilot 已执行）';

-- ---------------------------------------------------------------------------
-- ai_agent_runs（Run 生命周期）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ai_agent_runs (
    id                BIGINT       PRIMARY KEY,
    workspace_id      BIGINT       NOT NULL,
    agent_id          BIGINT       NOT NULL,
    user_id           BIGINT       NOT NULL,
    conversation_id   BIGINT       NOT NULL DEFAULT 0,
    mode              VARCHAR(20)  NOT NULL,
    status            VARCHAR(20)  NOT NULL DEFAULT 'created',
    trigger           VARCHAR(20)  NOT NULL DEFAULT 'user',
    trigger_payload   TEXT,
    input             TEXT,
    output            TEXT,
    error_message     TEXT,
    prompt_tokens     INTEGER      NOT NULL DEFAULT 0,
    completion_tokens INTEGER      NOT NULL DEFAULT 0,
    total_tokens      INTEGER      NOT NULL DEFAULT 0,
    tool_call_count   INTEGER      NOT NULL DEFAULT 0,
    needs_input_reason TEXT,
    started_at        TIMESTAMPTZ,
    completed_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_ai_run_mode CHECK (mode IN ('ask','build','autopilot')),
    CONSTRAINT chk_ai_run_status CHECK (status IN
        ('created','in_progress','awaiting','completed','stopping','stopped','failed','stale'))
);

CREATE INDEX IF NOT EXISTS idx_ai_agent_runs_agent ON ai_agent_runs (agent_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_agent_runs_status ON ai_agent_runs (status) WHERE status IN ('in_progress','awaiting','stopping');
CREATE INDEX IF NOT EXISTS idx_ai_agent_runs_workspace ON ai_agent_runs (workspace_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- ai_run_steps（可观测执行轨迹）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ai_run_steps (
    id         BIGINT      PRIMARY KEY,
    run_id     BIGINT      NOT NULL,
    step_type  VARCHAR(20) NOT NULL,
    summary    VARCHAR(500) NOT NULL DEFAULT '',
    detail     TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_ai_run_step_type CHECK (step_type IN ('thought','action','response','elicitation'))
);

CREATE INDEX IF NOT EXISTS idx_ai_run_steps_run ON ai_run_steps (run_id, created_at);

COMMENT ON TABLE ai_agent_runs IS 'Agent 执行记录（Run 生命周期）';
COMMENT ON TABLE ai_run_steps IS 'Agent 执行步骤（可观测轨迹）';
