-- 测试数据种子：补充项目 + 项目成员 + 每项目状态集
-- 前提：已有 users (5 rows, id 1=admin), workspaces (3 rows), roles, role_permissions, workspace_members

-- 1. 项目（每个工作空间创建一个测试项目，admin 作为 created_by）
INSERT INTO projects (workspace_id, name, slug, identifier, created_by, created_at, updated_at)
VALUES
  (3, 'E2E Test Project', 'e2e-test', 'E2E', 1, now(), now()),
  (1, 'Core Platform', 'core', 'CORE', 1, now(), now()),
  (2, 'Design System', 'design-system', 'DS', 1, now(), now())
ON CONFLICT DO NOTHING;

-- 2. 项目成员（admin 是 admin 角色，其他是 member）
INSERT INTO project_members (workspace_id, project_id, user_id, role, joined_at, created_at, updated_at)
SELECT p.ws_id, p.proj_id, u.id,
  CASE WHEN u.email = 'admin@ydsz.dev' THEN 'admin' ELSE 'member' END,
  now(), now(), now()
FROM (
  SELECT 3 AS ws_id, (SELECT id FROM projects WHERE workspace_id=3 LIMIT 1) AS proj_id
  UNION ALL SELECT 1, (SELECT id FROM projects WHERE workspace_id=1 LIMIT 1)
  UNION ALL SELECT 2, (SELECT id FROM projects WHERE workspace_id=2 LIMIT 1)
) p
CROSS JOIN users u
WHERE u.email IN ('admin@ydsz.dev', 'pm@ydsz.dev', 'dev@ydsz.dev', 'designer@ydsz.dev', 'viewer@ydsz.dev')
  AND p.proj_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- 3. 每项目插入状态集（OVERRIDING SYSTEM VALUE 允许指定 identity 列）
INSERT INTO states (id, workspace_id, project_id, name, "group", color, sequence, created_at, updated_at)
OVERRIDING SYSTEM VALUE
SELECT
  (p.ws_id * 1000 + s.seq) AS id,
  p.ws_id,
  p.proj_id,
  s.name,
  s.grp,
  s.color,
  s.seq,
  now(),
  now()
FROM (
  SELECT 3 AS ws_id, (SELECT id FROM projects WHERE workspace_id=3 LIMIT 1) AS proj_id
  UNION ALL SELECT 1, (SELECT id FROM projects WHERE workspace_id=1 LIMIT 1)
  UNION ALL SELECT 2, (SELECT id FROM projects WHERE workspace_id=2 LIMIT 1)
) p
CROSS JOIN (VALUES
  (0, '待办',     'backlog',    '#6B7280'),
  (1, '未开始',   'unstarted',  '#9CA3AF'),
  (2, '进行中',   'started',    '#3B82F6'),
  (3, '已完成',   'completed',  '#10B981'),
  (4, '已取消',   'cancelled',  '#EF4444')
) AS s(seq, name, grp, color)
WHERE p.proj_id IS NOT NULL
ON CONFLICT DO NOTHING;
