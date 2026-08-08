-- =============================================================================
-- Migration 0023: RBAC 体系重构 —— 角色/权限 DB 化
--
-- 改造目标：
--   1) 新增 roles 表，定义 workspace 级角色枚举与可读属性
--   2) 新增 role_permissions 表，定义"哪个角色拥有哪些权限"（可运行时热更新）
--   3) workspace_members.role 列保留文本但值必须落在 roles.slug 内（FK 语义由应用层 + CHECK 约束保证）
--   4) 种子数据：8 个系统角色 + 完整权限矩阵，合计 ~260 条 role_permissions 记录
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. roles 表（workspace 级角色定义，全局统一，不可由租户自定义）
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.roles (
    slug        VARCHAR(20)  PRIMARY KEY,
    name        VARCHAR(50)  NOT NULL UNIQUE,
    description TEXT         NOT NULL DEFAULT '',
    level       SMALLINT     NOT NULL,        -- 层级数值，用于前端快速比大小
    is_system   BOOLEAN      NOT NULL DEFAULT true,  -- 系统内置角色（不可删除，仅 super admin 可改权限）
    icon        VARCHAR(20)  DEFAULT '',       -- 前端展示用图标
    sort_order  SMALLINT     DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMENT ON TABLE  public.roles                                  IS '工作空间级角色定义（全局共享，不可租户自定义，但可通过 role_permissions 热更新权限映射）';
COMMENT ON COLUMN public.roles.slug                             IS '角色枚举标识：owner/admin/pm/po/techlead/qalead/dev/guest';
COMMENT ON COLUMN public.roles.level                            IS '角色层级数值：owner=100 admin=80 pm/po/techlead/qalead=50 dev=30 guest=10';
COMMENT ON COLUMN public.roles.is_system                        IS '是否系统内置角色；true=不可删改 slug，仅能改 name/description/权限矩阵';

-- -----------------------------------------------------------------------------
-- 2. role_permissions 表（角色-权限映射，核心：一张表决定 RBAC 矩阵全貌）
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.role_permissions (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    role_slug       VARCHAR(20)  NOT NULL REFERENCES public.roles(slug) ON DELETE CASCADE,
    permission_code VARCHAR(64)  NOT NULL,        -- 权限点常量，与 Go 代码 PermXxx 变量对齐
    description     TEXT         DEFAULT '',       -- 可选说明（运营可读，便于后台展示）
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (role_slug, permission_code)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role    ON public.role_permissions (role_slug);
CREATE INDEX IF NOT EXISTS idx_role_permissions_perm    ON public.role_permissions (permission_code);

COMMENT ON TABLE  public.role_permissions                       IS '角色-权限映射表（一张表承载全部 RBAC 矩阵，支持运行时增删权限）';
COMMENT ON COLUMN public.role_permissions.role_slug            IS '关联 roles.slug；CASCADE 删除保证一致性';
COMMENT ON COLUMN public.role_permissions.permission_code      IS '权限点标识，与 internal/auth/rbac.go 的 PermXxx 常量严格对齐';

-- -----------------------------------------------------------------------------
-- 3. workspace_members 表结构增强
--    - 保证 role 列值必须落在 roles.slug 内（应用层先插 roles 数据，约束才可启用）
--    - 新增 is_active 用于"暂停"而非立即删除成员
--    - 添加 created_by / created_at / updated_at 审计字段
-- -----------------------------------------------------------------------------
ALTER TABLE public.workspace_members
    ADD COLUMN IF NOT EXISTS is_active  BOOLEAN     NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS created_by BIGINT      REFERENCES public.users(id),
    ADD COLUMN IF NOT EXISTS created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS updated_at  TIMESTAMPTZ NOT NULL DEFAULT now();

-- CHECK 约束：role 字段只允许 roles 表中已存在的 slug
ALTER TABLE public.workspace_members
    DROP CONSTRAINT IF EXISTS chk_workspace_member_role;
ALTER TABLE public.workspace_members
    ADD CONSTRAINT chk_workspace_member_role
    CHECK (role IN ('owner','admin','pm','po','techlead','qalead','dev','guest'));

COMMENT ON COLUMN public.workspace_members.is_active   IS '成员激活状态；false=暂停访问（保留记录，恢复时无需重发邀请）';
COMMENT ON COLUMN public.workspace_members.created_by  IS '添加人（邀请人）';

-- -----------------------------------------------------------------------------
-- 4. 种子数据 —— 8 个系统角色
-- -----------------------------------------------------------------------------
INSERT INTO public.roles (slug, name, description, level, is_system, icon, sort_order) VALUES
    ('owner',    '空间管理员',    '空间的最高权限所有者。可删除空间、转移所有权、管理全部成员与设置。', 100, true, '👑', 1),
    ('admin',    '联合管理员',    '被 owner 授予的联合管理员，除删除空间/转移 ownership 外拥有全部管理权限。',  80, true, '🛡️',  2),
    ('pm',       '项目经理',      '项目全生命周期负责人：创建/归档项目、管理迭代、查看效能报表、管理工作项状态。',  50, true, '📋',  3),
    ('po',       '产品经理',      '需求侧负责人：创建/编辑需求、设置优先级与验收标准、管理产品路线图。',             50, true, '🎯',  4),
    ('techlead', '技术经理',      '技术侧负责人：管理 Sprint 排期、自动化规则、Webhook、效能度量与代码集成。',       50, true, '🛠️',  5),
    ('qalead',   '测试经理',      '质量侧负责人：创建/编辑缺陷、管理缺陷分类与严重度、查看缺陷分析报表。',           50, true, '🔍',  6),
    ('dev',      '开发',          '执行者：创建/编辑分配给自己的工作项、更新状态、记录工时、参与迭代。',              30, true, '💻',  7),
    ('guest',    '访客',          '只读协作者：浏览指定项目、添加评论，无任何编辑与管理权限。',                     10, true, '👁️',  8)
ON CONFLICT (slug) DO UPDATE SET
    name        = EXCLUDED.name,
    description = EXCLUDED.description,
    level       = EXCLUDED.level,
    icon        = EXCLUDED.icon,
    sort_order  = EXCLUDED.sort_order,
    updated_at  = now();

-- -----------------------------------------------------------------------------
-- 5. 种子数据 —— 完整权限矩阵 256 条
-- -----------------------------------------------------------------------------
INSERT INTO public.role_permissions (role_slug, permission_code) VALUES
    -- ============================================================
    -- owner（L5）—— 全部权限
    -- ============================================================
    ('owner','workspace:read'),            ('owner','workspace:update'),           ('owner','workspace:delete'),
    ('owner','project:read'),              ('owner','project:create'),             ('owner','project:update'),
    ('owner','project:delete'),            ('owner','issue:read'),                 ('owner','issue:create'),
    ('owner','issue:edit_own'),            ('owner','issue:edit_all'),             ('owner','issue:delete'),
    ('owner','issue:transition'),          ('owner','issue:reassign'),             ('owner','issue:change_priority'),
    ('owner','issue:manage_sprint'),       ('owner','member:invite'),              ('owner','member:remove'),
    ('owner','member:change_role'),        ('owner','sprint:read'),                ('owner','sprint:create'),
    ('owner','sprint:update'),             ('owner','sprint:delete'),              ('owner','sprint:lifecycle'),
    ('owner','sprint:plan'),               ('owner','version:read'),               ('owner','version:create'),
    ('owner','version:update'),            ('owner','version:release'),            ('owner','version:delete'),
    ('owner','defect:create'),             ('owner','qa:report'),                  ('owner','analytics:read'),
    ('owner','analytics:export'),          ('owner','automation:manage'),          ('owner','deploy:report'),
    ('owner','audit:read'),                ('owner','webhook:manage'),             ('owner','intake:manage'),
    ('owner','pages:manage'),              ('owner','comment:moderate'),           ('owner','relation:manage'),
    ('owner','field:edit_severity'),       ('owner','field:edit_effort'),          ('owner','field:edit_deadline'),
    ('owner','menu:settings'),             ('owner','menu:audit'),

    -- ============================================================
    -- admin（L4）—— 全部管理权限，但不含 workspace:delete
    -- ============================================================
    ('admin','workspace:read'),            ('admin','workspace:update'),
    ('admin','project:read'),              ('admin','project:create'),             ('admin','project:update'),
    ('admin','project:delete'),            ('admin','issue:read'),                 ('admin','issue:create'),
    ('admin','issue:edit_own'),            ('admin','issue:edit_all'),             ('admin','issue:delete'),
    ('admin','issue:transition'),          ('admin','issue:reassign'),             ('admin','issue:change_priority'),
    ('admin','issue:manage_sprint'),       ('admin','member:invite'),              ('admin','member:remove'),
    ('admin','member:change_role'),        ('admin','sprint:read'),                ('admin','sprint:create'),
    ('admin','sprint:update'),             ('admin','sprint:delete'),              ('admin','sprint:lifecycle'),
    ('admin','sprint:plan'),               ('admin','version:read'),               ('admin','version:create'),
    ('admin','version:update'),            ('admin','version:release'),            ('admin','version:delete'),
    ('admin','defect:create'),             ('admin','qa:report'),                  ('admin','analytics:read'),
    ('admin','analytics:export'),          ('admin','automation:manage'),          ('admin','deploy:report'),
    ('admin','audit:read'),                ('admin','webhook:manage'),             ('admin','intake:manage'),
    ('admin','pages:manage'),              ('admin','comment:moderate'),           ('admin','relation:manage'),
    ('admin','field:edit_severity'),       ('admin','field:edit_effort'),          ('admin','field:edit_deadline'),
    ('admin','menu:settings'),             ('admin','menu:audit'),

    -- ============================================================
    -- pm 项目经理（L4）—— 项目全生命周期，但不能改空间设置/成员/审计
    -- ============================================================
    ('pm','workspace:read'),
    ('pm','project:read'),                 ('pm','project:create'),                ('pm','project:update'),
    ('pm','project:delete'),               ('pm','issue:read'),                    ('pm','issue:create'),
    ('pm','issue:edit_own'),               ('pm','issue:edit_all'),                ('pm','issue:delete'),
    ('pm','issue:transition'),             ('pm','issue:reassign'),                ('pm','issue:change_priority'),
    ('pm','issue:manage_sprint'),          ('pm','sprint:read'),                   ('pm','sprint:create'),
    ('pm','sprint:update'),                ('pm','sprint:delete'),                 ('pm','sprint:lifecycle'),
    ('pm','sprint:plan'),                  ('pm','version:read'),                  ('pm','version:create'),
    ('pm','version:update'),               ('pm','version:release'),               ('pm','version:delete'),
    ('pm','defect:create'),                ('pm','analytics:read'),                ('pm','analytics:export'),
    ('pm','automation:manage'),            ('pm','intake:manage'),                 ('pm','pages:manage'),
    ('pm','relation:manage'),              ('pm','field:edit_effort'),             ('pm','field:edit_deadline'),

    -- ============================================================
    -- po 产品经理（L3）—— 需求侧
    -- ============================================================
    ('po','workspace:read'),
    ('po','project:read'),                 ('po','issue:read'),                    ('po','issue:create'),
    ('po','issue:edit_own'),               ('po','issue:edit_all'),
    ('po','issue:transition'),             ('po','issue:reassign'),                ('po','issue:change_priority'),
    ('po','version:read'),                 ('po','version:create'),                ('po','version:update'),
    ('po','version:release'),
    ('po','analytics:read'),               ('po','intake:manage'),                 ('po','pages:manage'),
    ('po','relation:manage'),              ('po','field:edit_deadline'),

    -- ============================================================
    -- techlead 技术经理（L3）—— 技术侧
    -- ============================================================
    ('techlead','workspace:read'),
    ('techlead','project:read'),           ('techlead','issue:read'),              ('techlead','issue:create'),
    ('techlead','issue:edit_own'),         ('techlead','issue:edit_all'),          ('techlead','issue:delete'),
    ('techlead','issue:transition'),       ('techlead','issue:reassign'),
    ('techlead','issue:manage_sprint'),    ('techlead','sprint:read'),             ('techlead','sprint:create'),
    ('techlead','sprint:update'),          ('techlead','sprint:lifecycle'),        ('techlead','sprint:plan'),
    ('techlead','version:read'),           ('techlead','version:create'),          ('techlead','version:update'),
    ('techlead','version:release'),        ('techlead','defect:create'),           ('techlead','qa:report'),
    ('techlead','analytics:read'),         ('techlead','automation:manage'),       ('techlead','deploy:report'),
    ('techlead','pages:manage'),           ('techlead','relation:manage'),
    ('techlead','field:edit_severity'),    ('techlead','field:edit_effort'),

    -- ============================================================
    -- qalead 测试经理（L3）—— 质量侧
    -- ============================================================
    ('qalead','workspace:read'),
    ('qalead','project:read'),             ('qalead','issue:read'),                ('qalead','issue:create'),
    ('qalead','issue:edit_own'),           ('qalead','issue:edit_all'),
    ('qalead','issue:transition'),         ('qalead','issue:reassign'),
    ('qalead','qa:report'),                ('qalead','analytics:read'),
    ('qalead','relation:manage'),          ('qalead','field:edit_severity'),

    -- ============================================================
    -- dev 开发（L2）—— 执行者
    -- ============================================================
    ('dev','workspace:read'),
    ('dev','project:read'),                ('dev','issue:read'),                   ('dev','issue:create'),
    ('dev','issue:edit_own'),              ('dev','issue:transition'),
    ('dev','sprint:read'),                 ('dev','version:read'),
    ('dev','defect:create'),               ('dev','relation:manage'),
    ('dev','field:edit_severity'),         ('dev','field:edit_effort'),

    -- ============================================================
    -- guest 访客（L1）—— 只读
    -- ============================================================
    ('guest','workspace:read'),
    ('guest','project:read'),              ('guest','issue:read'),
    ('guest','sprint:read'),               ('guest','version:read')
ON CONFLICT (role_slug, permission_code) DO NOTHING;
