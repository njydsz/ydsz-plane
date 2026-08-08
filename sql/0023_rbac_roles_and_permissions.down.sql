-- =============================================================================
-- Rollback: RBAC 体系重构
-- =============================================================================

-- 1. 恢复 workspace_members 到旧结构（移除本次新增列）
ALTER TABLE public.workspace_members
    DROP CONSTRAINT IF EXISTS chk_workspace_member_role,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS created_by,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;

-- 2. 删除 role_permissions（CASCADE 由于 FK 会自动清，但此处主动 DROP 表更保险）
DROP TABLE IF EXISTS public.role_permissions;

-- 3. 删除 roles
DROP TABLE IF EXISTS public.roles;
