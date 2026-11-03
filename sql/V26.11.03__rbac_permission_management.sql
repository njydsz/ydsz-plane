-- ===========================================================================
-- V26.11.03 · RBAC 权限管理基座（序列 + 索引）
-- 依据：docs/optimization/S20-数据库表·接口测试·前后端贯通现状分析与优化建议.md §1.3.2 D7
-- ===========================================================================
-- 背景：
--   role_permissions 表此前仅由 init.sql 中的静态种子数据填充，缺少：
--     1. id 列的序列默认值（导致运行时写入必须显式提供 id）
--     2. role_slug 索引（ListRolePermissions 每次全表扫描）
--   本迁移补齐这两项，使 RBAC 权限可由管理端动态配置。
-- ===========================================================================

-- 1. 为 role_permissions.id 创建序列并设为默认值。
--    START WITH 1 与现有种子数据的 abs(hashtext(...)) % 2^62 不冲突，
--    因为种子 ID 是大整数，序列从 1 开始不会碰撞。
CREATE SEQUENCE IF NOT EXISTS role_permissions_id_seq START 1 INCREMENT 1;

ALTER TABLE role_permissions
    ALTER COLUMN id SET DEFAULT nextval('role_permissions_id_seq');

-- 将序列所有权绑定到列，确保表删除时序列一并清理。
ALTER SEQUENCE role_permissions_id_seq OWNED BY role_permissions.id;

-- 2. role_slug 索引：加速 ListRolePermissions / RoleHasPermission 查询。
--    部分索引排除已软删除的行，减小索引体积。
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_slug
    ON role_permissions (role_slug)
    WHERE deleted = false;

-- 3. permission_code 索引：加速权限码维度的反查（如"哪些角色拥有 issue:delete"）。
CREATE INDEX IF NOT EXISTS idx_role_permissions_code
    ON role_permissions (permission_code)
    WHERE deleted = false;
