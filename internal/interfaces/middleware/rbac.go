// Package middleware — 工作空间级 RBAC 中间件。
//
// 中间件链顺序（外层 → 内层）：
//
//	RequireAuth → RequireWorkspaceParam → RequirePermission / RequireProjectParam → handler
//
// RequirePermission 校验用户在当前 workspace 拥有指定 permission；
// 校验通过后把 workspace_role 写入 ctx，供 handler 做二次鉴权。
//
// 错误语义：
//
//	401 Unauthorized — 未登录或 token 无效（由 RequireAuth 返回）。
//	403 Forbidden — 已登录但权限不足。
//	422 — workspace_id / project_id 格式错误。
package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/application/apitoken"
	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/rbac"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// RequireWorkspaceParam 是一个便利中间件：把路径 :workspace_id 解析成 int64
// 并写入 ctx (CtxWorkspaceID)。解析失败直接 422。
func RequireWorkspaceParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseBigInt(c.Param("workspace_id"))
		if !ok {
			respondError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field: "workspace_id", Reason: "无效的工作空间 ID",
			}))
			c.Abort()
			return
		}
		c.Set(CtxWorkspaceID, id)
		c.Next()
	}
}

// RequirePermission 校验当前用户在当前 workspace 上拥有特定权限。
// 必须配合 RequireAuth（设置 CtxUserID）与 RequireWorkspaceParam（设置 CtxWorkspaceID）使用。
func RequirePermission(store *auth.WorkspaceMembershipStore, perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawUID, _ := c.Get(CtxUserID)
		userID, ok := rawUID.(int64)
		if !ok || userID == 0 {
			respondError(c, errs.ErrUnauthorized)
			c.Abort()
			return
		}
		rawWS, ok := c.Get(CtxWorkspaceID)
		if !ok {
			respondError(c, errs.ErrForbidden)
			c.Abort()
			return
		}
		wsID := rawWS.(int64)

		m, err := store.ResolveRole(c.Request.Context(), wsID, userID)
		if err != nil {
			var appErr *errs.AppError
			if errors.As(err, &appErr) {
				respondError(c, appErr)
			} else {
				respondError(c, errs.ErrForbidden)
			}
			c.Abort()
			return
		}
		// 旧版 HasPermission 已废弃（硬编码角色矩阵迁至 DB），
		// 后端路由统一使用 RequirePermissionFromDB 做权限校验。
		// 这里仅做"是否为该工作空间成员"判定。
		_ = perm

		// API Token 双重要求：RBAC 角色通过后，还必须持有覆盖该权限的 scope。
		// 这是"个人令牌收敛"的关键防线（GitHub PAT 模型）：
		// 即使角色允许，token scope 不足依然 403。
		if c.GetString(CtxAuthKind) == string(auth.PrincipalAPIToken) {
			rawScopes, _ := c.Get(CtxAuthScopes)
			owned, _ := rawScopes.([]string)
			required, ok := apitoken.PermissionScope(perm)
			// 未纳入 scope 映射的权限：仅全权限（*）令牌放行，否则保守拒绝。
			if !ok {
				if !apitoken.ScopeCovers(owned, apitoken.ScopeAll) {
					respondError(c, errs.ErrForbidden)
					c.Abort()
					return
				}
			} else if !apitoken.ScopeCovers(owned, required) {
				respondError(c, errs.ErrForbidden)
				c.Abort()
				return
			}
		}

	c.Set("workspace_role", string(m.Role))
	c.Set("workspace_is_owner", m.Role == auth.RoleOwner)
	c.Next()
}
}

// RequirePermissionFromDB 从 DB-backed rbac.Store 解析权限，是 RequirePermission 的 DB 版替代。
// 校验通过后额外写入 ctx：workspace_role / workspace_is_owner / workspace_permissions（[]string）。
func RequirePermissionFromDB(rbacStore *rbac.Store, perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawUID, _ := c.Get(CtxUserID)
		userID, ok := rawUID.(int64)
		if !ok || userID == 0 {
			respondError(c, errs.ErrUnauthorized)
			c.Abort()
			return
		}
		rawWS, ok := c.Get(CtxWorkspaceID)
		if !ok {
			respondError(c, errs.ErrForbidden)
			c.Abort()
			return
		}
		wsID := rawWS.(int64)

		role, permissions, err := rbacStore.ResolveMembership(c.Request.Context(), wsID, userID)
		if err != nil {
			var appErr *errs.AppError
			if errors.As(err, &appErr) {
				respondError(c, appErr)
			} else {
				respondError(c, errs.ErrForbidden)
			}
			c.Abort()
			return
		}
		hasPerm, err := rbacStore.RoleHasPermission(c.Request.Context(), role.Slug, perm)
		if err != nil || !hasPerm {
			respondError(c, errs.ErrForbidden)
			c.Abort()
			return
		}

		// API Token scope 收敛防线（沿用旧逻辑）
		if c.GetString(CtxAuthKind) == string(auth.PrincipalAPIToken) {
			rawScopes, _ := c.Get(CtxAuthScopes)
			owned, _ := rawScopes.([]string)
			required, ok := apitoken.PermissionScope(perm)
			if !ok {
				if !apitoken.ScopeCovers(owned, apitoken.ScopeAll) {
					respondError(c, errs.ErrForbidden)
					c.Abort()
					return
				}
			} else if !apitoken.ScopeCovers(owned, required) {
				respondError(c, errs.ErrForbidden)
				c.Abort()
				return
			}
		}

		c.Set("workspace_role", role.Slug)
		c.Set("workspace_is_owner", role.Slug == "owner")
		c.Set("workspace_permissions", permissions)
		c.Next()
	}
}

// RequireProjectParam 把路径 :project_id 解析成 int64 并写入 ctx (CtxProjectID)。
// 解析失败直接 422。
//
// 注意：必须嵌套在 RequireWorkspaceParam 之后的子路由，因为项目是工作空间下的二级资源。
func RequireProjectParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseBigInt(c.Param("project_id"))
		if !ok {
			respondError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field: "project_id", Reason: "无效的项目 ID",
			}))
			c.Abort()
			return
		}
		c.Set(CtxProjectID, id)
		c.Next()
	}
}

// parseBigInt 解析 int64（简化版，无依赖 strconv 防止多引入）。
func parseBigInt(s string) (int64, bool) {
	if s == "" {
		return 0, false
	}
	var n int64
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int64(r-'0')
		if n < 0 { // 溢出保护
			return 0, false
		}
	}
	return n, true
}
