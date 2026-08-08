// Package rbac 提供基于 PostgreSQL 的 DB-backed RBAC 数据存储与查询能力。
//
// 架构分层：
//   - roles 表：角色枚举（owner/admin/pm/po/techlead/qalead/dev/guest）
//   - role_permissions 表：动态的角色-权限映射（可运行时热更新）
//   - 缓存层：TTL 内存缓存 + singleflight 风格的懒加载，避免每次请求都查 DB
//
// 与旧 code（internal/auth/rbac.go）的兼容：
//   - 角色的枚举值、权限点常量字符串保持不变（通过 PermXxx 常量导出）
//   - ResolveRole / HasPermission 语义一致，下层实现从"内存 map"改为"DB + 缓存"
package rbac

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Role 代表 roles 表中的一行。
type Role struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int    `json:"level"`
	IsSystem    bool   `json:"is_system"`
	Icon        string `json:"icon"`
}

// Store 提供角色和权限的 DB 访问 + 内存缓存。
type Store struct {
	db       *pgxpool.Pool
	log      *zap.Logger
	cacheTTL time.Duration

	mu        sync.RWMutex
	roleCache map[string]*Role               // slug -> Role
	permCache map[string]map[string]struct{} // slug -> set(perm)
	cacheExp  map[string]time.Time           // slug -> expire
	initOnce  sync.Once
	initErr   error
}

// NewStore 构造 Store。
func NewStore(db *pgxpool.Pool, log *zap.Logger) *Store {
	return &Store{
		db:        db,
		log:       log.Named("rbac.store"),
		cacheTTL:  5 * time.Minute,
		roleCache: make(map[string]*Role),
		permCache: make(map[string]map[string]struct{}),
		cacheExp:  make(map[string]time.Time),
	}
}

// InitCache 在启动时预热全部 roles + role_permissions 到内存缓存。
// 调用失败不致命（降级为每次查 DB），但会记录错误。
func (s *Store) InitCache(ctx context.Context) error {
	s.initOnce.Do(func() {
		s.initErr = s.refreshAllLocked(ctx)
		if s.initErr != nil {
			s.log.Warn("RBAC cache warm-up failed, falling back to DB queries", zap.Error(s.initErr))
		}
	})
	return s.initErr
}

// ListRoles 返回全部角色定义（按 level 降序）。
func (s *Store) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := s.db.Query(ctx, `
		SELECT slug, name, description, level, is_system, icon
		FROM roles ORDER BY level DESC, sort_order ASC`)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []Role
	for rows.Next() {
		var r Role
		if err := rows.Scan(&r.Slug, &r.Name, &r.Description, &r.Level, &r.IsSystem, &r.Icon); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, r)
	}
	return out, nil
}

// ListRolePermissions 返回指定角色的所有权限码。
func (s *Store) ListRolePermissions(ctx context.Context, roleSlug string) ([]string, error) {
	// 1. 先尝试读缓存
	if perms := s.getCachedPerms(roleSlug); perms != nil {
		out := make([]string, 0, len(perms))
		for p := range perms {
			out = append(out, p)
		}
		sort.Strings(out)
		return out, nil
	}

	// 2. 缓存未命中：查 DB 并回填缓存
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refreshPermsLocked(ctx, roleSlug); err != nil {
		return nil, err
	}
	perms := s.permCache[roleSlug]
	out := make([]string, 0, len(perms))
	for p := range perms {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

// RoleHasPermission 判断指定角色是否拥有目标权限。
func (s *Store) RoleHasPermission(ctx context.Context, roleSlug, perm string) (bool, error) {
	if perms := s.getCachedPerms(roleSlug); perms != nil {
		_, ok := perms[perm]
		return ok, nil
	}
	// 缓存未命中：查 DB
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refreshPermsLocked(ctx, roleSlug); err != nil {
		return false, err
	}
	_, ok := s.permCache[roleSlug][perm]
	return ok, nil
}

// ResolveMembership 返回用户在指定 workspace 的角色与权限列表。
func (s *Store) ResolveMembership(ctx context.Context, wsID, userID int64) (Role, []string, error) {
	var roleSlug string
	err := s.db.QueryRow(ctx, `
		SELECT role FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2 AND is_active = true`, wsID, userID).
		Scan(&roleSlug)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Role{}, nil, errs.ErrForbidden
		}
		return Role{}, nil, errs.ErrInternal.Wrap(err)
	}
	if roleSlug == "" {
		return Role{}, nil, errs.ErrForbidden
	}

	// 加载角色信息
	role, err := s.getRole(ctx, roleSlug)
	if err != nil {
		return Role{}, nil, err
	}

	// 加载权限
	perms, err := s.ListRolePermissions(ctx, roleSlug)
	if err != nil {
		return Role{}, nil, err
	}
	return role, perms, nil
}

// InvalidateCache 使指定角色的缓存失效（供权限变更后调用）。
func (s *Store) InvalidateCache(roleSlug string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.permCache, roleSlug)
	delete(s.cacheExp, roleSlug)
	s.log.Info("RBAC cache invalidated", zap.String("role", roleSlug))
}

// ResetCache 清空全部缓存（供测试或全量刷新）。
func (s *Store) ResetCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.permCache = make(map[string]map[string]struct{})
	s.cacheExp = make(map[string]time.Time)
	s.roleCache = make(map[string]*Role)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (s *Store) getCachedPerms(roleSlug string) map[string]struct{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if exp, ok := s.cacheExp[roleSlug]; ok && time.Now().Before(exp) {
		if perms, ok := s.permCache[roleSlug]; ok {
			return perms
		}
	}
	return nil
}

func (s *Store) refreshPermsLocked(ctx context.Context, roleSlug string) error {
	rows, err := s.db.Query(ctx, `SELECT permission_code FROM role_permissions WHERE role_slug = $1`, roleSlug)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	set := make(map[string]struct{})
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		set[p] = struct{}{}
	}
	s.permCache[roleSlug] = set
	s.cacheExp[roleSlug] = time.Now().Add(s.cacheTTL)
	return nil
}

func (s *Store) refreshAllLocked(ctx context.Context) error {
	// 加载全部角色
	rRows, err := s.db.Query(ctx, `SELECT slug, name, description, level, is_system, icon FROM roles`)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	for rRows.Next() {
		var r Role
		if err := rRows.Scan(&r.Slug, &r.Name, &r.Description, &r.Level, &r.IsSystem, &r.Icon); err != nil {
			rRows.Close()
			return errs.ErrInternal.Wrap(err)
		}
		s.roleCache[r.Slug] = &r
	}
	rRows.Close()

	// 加载全部角色-权限映射
	pRows, err := s.db.Query(ctx, `SELECT role_slug, permission_code FROM role_permissions`)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	for pRows.Next() {
		var slug, perm string
		if err := pRows.Scan(&slug, &perm); err != nil {
			pRows.Close()
			return errs.ErrInternal.Wrap(err)
		}
		if s.permCache[slug] == nil {
			s.permCache[slug] = make(map[string]struct{})
		}
		s.permCache[slug][perm] = struct{}{}
	}
	pRows.Close()
	return nil
}

func (s *Store) getRole(ctx context.Context, slug string) (Role, error) {
	s.mu.RLock()
	if r, ok := s.roleCache[slug]; ok {
		s.mu.RUnlock()
		return *r, nil
	}
	s.mu.RUnlock()

	var r Role
	err := s.db.QueryRow(ctx, `SELECT slug, name, description, level, is_system, icon FROM roles WHERE slug = $1`, slug).
		Scan(&r.Slug, &r.Name, &r.Description, &r.Level, &r.IsSystem, &r.Icon)
	if err != nil {
		return Role{}, errs.ErrNotFound
	}
	s.mu.Lock()
	s.roleCache[slug] = &r
	s.mu.Unlock()
	return r, nil
}
