-- V26.11.06__dlp.sql — 数据防泄漏（DLP）
--
-- 对标 Plane 的 Argus 敏感数据扫描。

-- ---------------------------------------------------------------------------
-- dlp_findings — 扫描发现
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS dlp_findings (
    id           BIGINT       PRIMARY KEY,
    job_id       BIGINT       NOT NULL,
    workspace_id BIGINT       NOT NULL,
    scope        VARCHAR(20)  NOT NULL,   -- project / page / issue / global
    target_id    BIGINT,
    target_title VARCHAR(500),
    category     VARCHAR(30)  NOT NULL,   -- id_card / bank_card / phone / email / ...
    severity     VARCHAR(20)  NOT NULL,   -- critical / high / medium / low
    rule_name    VARCHAR(100) NOT NULL,
    masked_value TEXT         NOT NULL,   -- 脱敏后的值
    offset_pos   INTEGER      NOT NULL DEFAULT 0,
    confidence   DOUBLE PRECISION NOT NULL DEFAULT 0,
    action       VARCHAR(20)  NOT NULL DEFAULT 'flagged', -- flagged / blocked / resolved
    actioned_by  BIGINT,
    resolved_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_dlp_severity CHECK (severity IN ('critical','high','medium','low')),
    CONSTRAINT chk_dlp_action CHECK (action IN ('flagged','blocked','resolved'))
);

CREATE INDEX IF NOT EXISTS idx_dlp_findings_workspace ON dlp_findings (workspace_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_dlp_findings_job ON dlp_findings (job_id);
CREATE INDEX IF NOT EXISTS idx_dlp_findings_category ON dlp_findings (workspace_id, category);
CREATE INDEX IF NOT EXISTS idx_dlp_findings_action ON dlp_findings (workspace_id, action) WHERE action = 'flagged';

-- ---------------------------------------------------------------------------
-- dlp_rules — 自定义规则（用户可配置）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS dlp_rules (
    id           BIGINT       PRIMARY KEY,
    workspace_id BIGINT,                  -- NULL = 全局规则
    name         VARCHAR(200) NOT NULL,
    category     VARCHAR(30)  NOT NULL,
    severity     VARCHAR(20)  NOT NULL,
    pattern      TEXT,                    -- 正则表达式
    keywords     TEXT[],                  -- 关键词列表
    enabled      BOOLEAN      NOT NULL DEFAULT TRUE,
    created_by   BIGINT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, name)
);

CREATE INDEX IF NOT EXISTS idx_dlp_rules_workspace ON dlp_rules (workspace_id, enabled);

COMMENT ON TABLE dlp_findings IS '敏感数据扫描发现';
COMMENT ON TABLE dlp_rules IS 'DLP 检测规则（内置 + 自定义）';
