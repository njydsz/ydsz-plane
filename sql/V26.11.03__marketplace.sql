-- V26.11.03__marketplace.sql — 应用商店（Marketplace）
--
-- 对标 Plane 的 Marketplace：应用/导入器/模板的目录、评分、安装统计。
--
-- 设计：
--   - marketplace_apps      ：应用目录（内置 + 第三方提交）
--   - marketplace_installs  ：安装记录（工作空间 × 应用，唯一）
--   - marketplace_reviews   ：评分与评论
--   - marketplace_categories：分类（应用 / 导入器 / 模板 / 集成）

-- ---------------------------------------------------------------------------
-- marketplace_categories
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS marketplace_categories (
    slug        VARCHAR(50)  PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    icon        VARCHAR(50),
    sort_order  INTEGER      NOT NULL DEFAULT 0
);

INSERT INTO marketplace_categories (slug, name, description, icon, sort_order) VALUES
    ('apps',      '应用',   '扩展平台能力的应用',     'Plug',     10),
    ('importers', '导入器', '从外部平台迁移数据',     'Upload',   20),
    ('templates', '模板',   '项目与工作项模板',       'LayoutTemplate', 30),
    ('integrations', '集成', '第三方服务连接',        'Blocks',   40)
ON CONFLICT (slug) DO NOTHING;

-- ---------------------------------------------------------------------------
-- marketplace_apps
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS marketplace_apps (
    id            BIGINT       PRIMARY KEY,
    slug          VARCHAR(100) NOT NULL UNIQUE,
    name          VARCHAR(200) NOT NULL,
    description   TEXT,
    -- app / importer / template / integration
    category      VARCHAR(50)  NOT NULL DEFAULT 'apps',
    -- 提供方（官方 / 第三方）
    publisher     VARCHAR(100) NOT NULL DEFAULT 'official',
    publisher_url VARCHAR(500),
    icon_url      VARCHAR(500),
    -- 截图
    screenshots   TEXT[]       NOT NULL DEFAULT '{}',
    -- 标签
    tags          TEXT[]       NOT NULL DEFAULT '{}',
    -- 定价：free / paid / freemium
    pricing       VARCHAR(20)  NOT NULL DEFAULT 'free',
    -- 安装方式：builtin / oauth / webhook / cli
    install_type  VARCHAR(20)  NOT NULL DEFAULT 'builtin',
    -- 安装配置（JSON，如 OAuth client_id、webhook URL 模板）
    install_config JSONB       NOT NULL DEFAULT '{}',
    -- 文档链接
    docs_url      VARCHAR(500),
    -- 评分聚合（0-5，保留两位）
    rating_avg    NUMERIC(3,2) NOT NULL DEFAULT 0,
    rating_count  INTEGER      NOT NULL DEFAULT 0,
    install_count INTEGER      NOT NULL DEFAULT 0,
    -- 状态：published / draft / deprecated
    status        VARCHAR(20)  NOT NULL DEFAULT 'published',
    created_by    BIGINT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_marketplace_app_category CHECK (category IN ('apps','importers','templates','integrations')),
    CONSTRAINT chk_marketplace_app_pricing CHECK (pricing IN ('free','paid','freemium')),
    CONSTRAINT chk_marketplace_app_status CHECK (status IN ('published','draft','deprecated'))
);

CREATE INDEX IF NOT EXISTS idx_marketplace_apps_category ON marketplace_apps (category, status);
CREATE INDEX IF NOT EXISTS idx_marketplace_apps_rating ON marketplace_apps (rating_avg DESC);
CREATE INDEX IF NOT EXISTS idx_marketplace_apps_installs ON marketplace_apps (install_count DESC);

-- ---------------------------------------------------------------------------
-- marketplace_installs
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS marketplace_installs (
    id           BIGINT      PRIMARY KEY,
    workspace_id BIGINT      NOT NULL,
    app_id       BIGINT      NOT NULL,
    installed_by BIGINT      NOT NULL,
    -- 安装后的配置（如 OAuth 凭证引用、webhook 配置）
    config       JSONB       NOT NULL DEFAULT '{}',
    enabled      BOOLEAN     NOT NULL DEFAULT TRUE,
    installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (workspace_id, app_id)
);

CREATE INDEX IF NOT EXISTS idx_marketplace_installs_workspace ON marketplace_installs (workspace_id, enabled);

-- ---------------------------------------------------------------------------
-- marketplace_reviews
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS marketplace_reviews (
    id           BIGINT      PRIMARY KEY,
    app_id       BIGINT      NOT NULL,
    workspace_id BIGINT      NOT NULL,
    user_id      BIGINT      NOT NULL,
    rating       INTEGER     NOT NULL,
    comment      TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_marketplace_review_rating CHECK (rating BETWEEN 1 AND 5),
    UNIQUE (app_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_marketplace_reviews_app ON marketplace_reviews (app_id);

COMMENT ON TABLE marketplace_apps IS '应用商店目录';
COMMENT ON TABLE marketplace_installs IS '应用安装记录';
COMMENT ON TABLE marketplace_reviews IS '应用评分与评论';
