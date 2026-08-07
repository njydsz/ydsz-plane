// Package httpapi — workspace / member / invitation / project 端点处理函数。
//
// Sprint 2 (M1) 实施：可创建空间 → 邀请成员 → 建项目，RBAC 生效。
package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/application/workspace"
	"github.com/njydsz/ydsz-plane/internal/interfaces/http/dto"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ==================================================================
// Workspaces 集合路由（不依赖 :workspace_id）
// ==================================================================

func getWorkspaceBySlug(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		ws, err := d.WorkspaceSvc.GetBySlug(c.Request.Context(), slug)
		if err != nil {
			writeError(c, err)
			return
		}
		uid := c.GetInt64(middleware.CtxUserID)
		if m, err := d.WorkspaceStore.ResolveRole(c.Request.Context(), ws.ID, uid); err == nil {
			ws.Role = string(m.Role)
		}
		c.JSON(http.StatusOK, ws)
	}
}

func listWorkspaces(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64(middleware.CtxUserID)
		items, err := d.WorkspaceSvc.ListByUser(c.Request.Context(), uid)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, items)
	}
}

func createWorkspace(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CreateWorkspaceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		ws, err := d.WorkspaceSvc.Create(c.Request.Context(), workspace.CreateInput{
			Name: req.Name, Slug: req.Slug, Timezone: req.Timezone, Language: req.Language,
			OwnerID: c.GetInt64(middleware.CtxUserID),
		})
		if err != nil {
			writeError(c, err)
			return
		}
		d.AuditSvc.RecordFromGin(c, ws.ID, "workspace.create", ws.Name, map[string]any{
			"slug": ws.Slug, "timezone": ws.Timezone,
		})
		c.JSON(http.StatusCreated, ws)
	}
}

// ==================================================================
// 工作空间详情 & 设置
// ==================================================================

func getWorkspace(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		ws, err := d.WorkspaceSvc.Get(c.Request.Context(), wsID)
		if err != nil {
			writeError(c, err)
			return
		}
		uid := c.GetInt64(middleware.CtxUserID)
		if m, err := d.WorkspaceStore.ResolveRole(c.Request.Context(), wsID, uid); err == nil {
			ws.Role = string(m.Role)
		}
		c.JSON(http.StatusOK, ws)
	}
}

func updateWorkspace(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		var req dto.UpdateWorkspaceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		ws, err := d.WorkspaceSvc.Update(c.Request.Context(), wsID, workspace.UpdateInput{
			Name: req.Name, Timezone: req.Timezone, Language: req.Language, LogoURL: req.LogoURL,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		d.AuditSvc.RecordFromGin(c, wsID, "workspace.update", ws.Name, map[string]any{
			"fields": req,
		})
		c.JSON(http.StatusOK, ws)
	}
}

func archiveWorkspace(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		if err := d.WorkspaceSvc.Archive(c.Request.Context(), wsID); err != nil {
			writeError(c, err)
			return
		}
		d.AuditSvc.RecordFromGin(c, wsID, "workspace.archive", "", nil)
		c.Status(http.StatusNoContent)
	}
}

// ==================================================================
// 成员管理
// ==================================================================

func listMembers(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		members, err := d.MemberSvc.List(c.Request.Context(), wsID)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, members)
	}
}

func changeMemberRole(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		targetID, _ := strconv.ParseInt(c.Param("user_id"), 10, 64)
		var req dto.ChangeRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		if targetID == c.GetInt64(middleware.CtxUserID) {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field: "user_id", Reason: "不可修改自己的角色",
			}))
			return
		}
		if err := d.MemberSvc.ChangeRole(c.Request.Context(), wsID, targetID, req.Role); err != nil {
			writeError(c, err)
			return
		}
		d.AuditSvc.RecordFromGin(c, wsID, "member.role_change", strconv.FormatInt(targetID, 10), map[string]any{
			"new_role": req.Role,
		})
		c.Status(http.StatusNoContent)
	}
}

func removeMember(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		targetID, _ := strconv.ParseInt(c.Param("user_id"), 10, 64)
		if targetID == c.GetInt64(middleware.CtxUserID) {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field: "user_id", Reason: "不可移除自己",
			}))
			return
		}
		if err := d.MemberSvc.RemoveMember(c.Request.Context(), wsID, targetID); err != nil {
			writeError(c, err)
			return
		}
		d.AuditSvc.RecordFromGin(c, wsID, "member.remove", strconv.FormatInt(targetID, 10), nil)
		c.Status(http.StatusNoContent)
	}
}

// ==================================================================
// 邀请
// ==================================================================

func sendInvitation(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		var req dto.SendInvitationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		inv, err := d.InvitationSvc.Invite(c.Request.Context(), workspace.InviteInput{
			WorkspaceID: wsID,
			InviterID:   c.GetInt64(middleware.CtxUserID),
			Email:       req.Email,
			Role:        req.Role,
			Message:     req.Message,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		d.AuditSvc.RecordFromGin(c, wsID, "invitation.send", req.Email, map[string]any{
			"role": inv.Role,
		})
		c.JSON(http.StatusCreated, inv)
	}
}

func listInvitations(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		status := c.Query("status")
		items, err := d.InvitationSvc.ListByWorkspace(c.Request.Context(), wsID, status)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, items)
	}
}

func revokeInvitation(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		invID, _ := strconv.ParseInt(c.Param("invitation_id"), 10, 64)
		if err := d.InvitationSvc.Revoke(c.Request.Context(), invID, wsID); err != nil {
			writeError(c, err)
			return
		}
		d.AuditSvc.RecordFromGin(c, wsID, "invitation.revoke", strconv.FormatInt(invID, 10), nil)
		c.Status(http.StatusNoContent)
	}
}

func acceptInvitation(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.AcceptInvitationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		inv, err := d.InvitationSvc.Accept(c.Request.Context(), req.Token, c.GetInt64(middleware.CtxUserID))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, inv)
	}
}

func getInvitationPreview(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		preview, err := d.InvitationSvc.Preview(c.Request.Context(), token)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, preview)
	}
}

// ==================================================================
// 项目
// ==================================================================

func listProjects(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		items, err := d.ProjectSvc.ListByWorkspace(c.Request.Context(), wsID)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, items)
	}
}

func createProject(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		var req dto.CreateProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		p, err := d.ProjectSvc.Create(c.Request.Context(), workspace.ProjectCreateInput{
			WorkspaceID: wsID,
			Name:        req.Name,
			Slug:        req.Slug,
			Identifier:  req.Identifier,
			Description: req.Description,
			Network:     req.Network,
			Icon:        req.Icon,
			Color:       req.Color,
			CreatedBy:   c.GetInt64(middleware.CtxUserID),
		})
		if err != nil {
			writeError(c, err)
			return
		}
		d.AuditSvc.RecordFromGin(c, wsID, "project.create", p.Identifier, map[string]any{
			"name": p.Name, "slug": p.Slug,
		})
		c.JSON(http.StatusCreated, p)
	}
}

func getProject(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		projectID, _ := strconv.ParseInt(c.Param("project_id"), 10, 64)
		p, err := d.ProjectSvc.Get(c.Request.Context(), wsID, projectID)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, p)
	}
}

func updateProject(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		projectID, _ := strconv.ParseInt(c.Param("project_id"), 10, 64)
		var req dto.UpdateProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		p, err := d.ProjectSvc.Update(c.Request.Context(), wsID, projectID, workspace.ProjectUpdateInput{
			Name: req.Name, Slug: req.Slug, Description: req.Description,
			Network: req.Network, Icon: req.Icon, Color: req.Color,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, p)
	}
}

func archiveProject(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		projectID, _ := strconv.ParseInt(c.Param("project_id"), 10, 64)
		if err := d.ProjectSvc.Archive(c.Request.Context(), wsID, projectID); err != nil {
			writeError(c, err)
			return
		}
		d.AuditSvc.RecordFromGin(c, wsID, "project.archive", strconv.FormatInt(projectID, 10), nil)
		c.Status(http.StatusNoContent)
	}
}

// ==================================================================
// 审计（仅 owner/admin 可见）
// ==================================================================

func listAuditLogs(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		limit := 50
		if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 200 {
			limit = l
		}
		rows, err := d.AuditSvc.List(c.Request.Context(), wsID, limit)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, rows)
	}
}
