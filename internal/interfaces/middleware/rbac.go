// Package middleware — 工作空间级 RBAC 中间件。
//
// RequirePermission(store, perm) 校验用户在指定 workspace 上是否拥有 perm。
// 必须配合 RequireAuth 使用（前置中间件设置 CtxUserID），
// 调用方需先通过路径参数或查询字符串将 workspace_id 注入到 ctx。
package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/application/auth"
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
		if !m.HasPermission(perm) {
			respondError(c, errs.ErrForbidden)
			c.Abort()
			return
		}
		c.Set("workspace_role", string(m.Role))
		c.Next()
	}
}

// RequireProjectParam 把路径 :project_id 解析成 int64 并写入 ctx (CtxProjectID)。
// 解析失败直接 422。通常嵌套在 RequireWorkspaceParam 之后的子路由组。
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
		if n < 0 { // overflow
			return 0, false
		}
	}
	return n, true
}
