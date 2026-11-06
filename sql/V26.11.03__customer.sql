-- V26.11.03__customer.sql — 客户管理（CRM 层）
--
-- 对标 Plane Business / HubSpot 轻量 CRM：
--   - customers         ：客户记录（含行业、规模、收入、阶段、套餐、区域等自定义属性）；
--   - customer_requests ：客户请求（可关联到 intake 工单或正式工作项）。
--
-- 应用层生成 id（雪花式 BIGINT），两表均使用 deleted 软删除。

-- ---------------------------------------------------------------------------
-- customers
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS customers (
    id                BIGINT       PRIMARY KEY,
    workspace_id      BIGINT       NOT NULL,
    name              VARCHAR(255) NOT NULL,
    email             VARCHAR(255),
    website           VARCHAR(255),
    industry          VARCHAR(100),
    stage             VARCHAR(20)  NOT NULL DEFAULT 'lead',
    tier              VARCHAR(20)  NOT NULL DEFAULT 'free',
    region            VARCHAR(100),
    employees         INTEGER      NOT NULL DEFAULT 0,
    revenue           VARCHAR(50),
    account_manager_id BIGINT,
    notes             TEXT,
    deleted           BOOLEAN      NOT NULL DEFAULT FALSE,
    created_by        BIGINT,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_customers_stage CHECK (stage IN ('lead', 'prospect', 'customer', 'churned')),
    CONSTRAINT chk_customers_tier  CHECK (tier  IN ('free', 'pro', 'enterprise'))
);

CREATE INDEX IF NOT EXISTS idx_customers_workspace_id ON customers (workspace_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_customers_workspace_stage ON customers (workspace_id, stage) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_customers_workspace_tier  ON customers (workspace_id, tier)  WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_customers_workspace_industry ON customers (workspace_id, industry) WHERE NOT deleted;
-- 名称前缀检索（服务层 ILIKE 模糊搜索的辅助索引；pg_trgm 可用时可由 V26.09.25 补充 trgm 索引）。
CREATE INDEX IF NOT EXISTS idx_customers_name ON customers (name text_pattern_ops) WHERE NOT deleted;

-- ---------------------------------------------------------------------------
-- customer_requests
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS customer_requests (
    id             BIGINT       PRIMARY KEY,
    workspace_id   BIGINT       NOT NULL,
    customer_id    BIGINT       NOT NULL,
    title          VARCHAR(255) NOT NULL,
    description    TEXT,
    status         VARCHAR(20)  NOT NULL DEFAULT 'new',
    priority       VARCHAR(20)  NOT NULL DEFAULT 'medium',
    source         VARCHAR(20)  NOT NULL DEFAULT 'email',
    work_item_id   BIGINT,
    work_item_type VARCHAR(20),
    assigned_to    BIGINT,
    deleted        BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_customer_requests_status   CHECK (status IN ('new', 'acknowledged', 'in_progress', 'resolved', 'closed')),
    CONSTRAINT chk_customer_requests_priority CHECK (priority IN ('urgent', 'high', 'medium', 'low')),
    CONSTRAINT chk_customer_requests_source   CHECK (source IN ('email', 'form', 'in_app', 'phone'))
);

CREATE INDEX IF NOT EXISTS idx_customer_requests_workspace_id ON customer_requests (workspace_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_customer_requests_customer_id ON customer_requests (customer_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_customer_requests_status ON customer_requests (workspace_id, status) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_customer_requests_priority ON customer_requests (workspace_id, priority) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_customer_requests_work_item ON customer_requests (work_item_id) WHERE work_item_id IS NOT NULL;

COMMENT ON TABLE customers IS '客户记录';
COMMENT ON TABLE customer_requests IS '客户请求（关联客户，可关联工作项）';

-- ---------------------------------------------------------------------------
-- 客户管理权限种子 — 幂等
-- 角色：owner / pm / po / techlead / qalead 授予读+管理
-- ---------------------------------------------------------------------------
INSERT INTO role_permissions (id, tenant_id, role_slug, permission_code)
SELECT (abs(hashtext('customer:' || r.role_slug || ':' || x.p)) % 4611686018427387904)::bigint,
       1, r.role_slug, x.p
FROM (VALUES ('owner'), ('pm'), ('po'), ('techlead'), ('qalead')) AS r(role_slug),
     (VALUES ('customer:read'), ('customer:manage')) AS x(p)
ON CONFLICT DO NOTHING;
