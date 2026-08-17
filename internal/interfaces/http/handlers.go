// Package httpapi — workspace / member / invitation / project 端点处理函数。
//
// Sprint 2 (M1) 实施：可创建空间 → 邀请成员 → 建项目，RBAC 生效。
package httpapi

import (
	"context"
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/application/workspace"
	"github.com/njydsz/ydsz-plane/internal/interfaces/http/dto"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ==================================================================
// Workspaces 集合路由（不依赖 :workspace_id）
// ==================================================================

// getWorkspaceBySlug 根据 URL 中的 slug 查询工作空间，并附带当前用户的角色。
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

// listWorkspaces 返回当前用户参与的所有工作空间。
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

// createWorkspace 创建新工作空间，并将当前用户设为 owner，记录审计日志。
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

// getWorkspace 返回指定工作空间的详情（含当前用户角色）。
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

// updateWorkspace 更新工作空间的名称/时区/语言/Logo，并记录审计日志。
func updateWorkspace(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		var req dto.UpdateWorkspaceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		ws, err := d.WorkspaceSvc.Update(c.Request.Context(), wsID, workspace.UpdateInput{
			Name: req.Name, Timezone: req.Timezone, Language: req.Language, LogoURL: req.LogoURL, BrandColor: req.BrandColor,
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

// ==================================================================
// 工作空间 Logo 上传
// ==================================================================

// LogoUploadResult 上传 Logo 的响应。
type LogoUploadResult struct {
	LogoURL string `json:"logo_url"`
}

// uploadLogo 处理工作空间 Logo 上传：
// POST /api/v1/workspaces/:workspace_id/logo
// Content-Type: multipart/form-data，字段名 "file"
// 需要 workspace:update 权限。
func uploadLogo(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d.Storage == nil {
			middleware.AbortWithError(c, errs.ErrInternal.WithDetails(errs.FieldDetail{
				Field: "storage", Reason: "未配置对象存储",
			}))
			return
		}

		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		header, err := c.FormFile("file")
		if err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field: "file", Reason: "请上传文件（form 字段名 'file'）",
			}))
			return
		}

		f, err := header.Open()
		if err != nil {
			middleware.AbortWithError(c, errs.ErrInternal.Wrap(err))
			return
		}
		defer f.Close()

		logoSvc := workspace.NewLogoService(d.WorkspaceSvc, d.Storage)
		logoURL, err := logoSvc.SaveLogo(c.Request.Context(), wsID, f, header.Size, header.Header.Get("Content-Type"))
		if err != nil {
			writeError(c, err)
			return
		}

		d.AuditSvc.RecordFromGin(c, wsID, "workspace.logo_upload", "", nil)
		c.JSON(http.StatusOK, LogoUploadResult{LogoURL: logoURL})
	}
}

// removeLogo 清除工作空间 Logo：
// DELETE /api/v1/workspaces/:workspace_id/logo
// 需要 workspace:update 权限。
func removeLogo(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		logoSvc := workspace.NewLogoService(d.WorkspaceSvc, d.Storage)
		if err := logoSvc.RemoveLogo(c.Request.Context(), wsID); err != nil {
			writeError(c, err)
			return
		}
		d.AuditSvc.RecordFromGin(c, wsID, "workspace.logo_remove", "", nil)
		c.Status(http.StatusNoContent)
	}
}

// archiveWorkspace 归档指定工作空间（软删除），返回 204。
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

// listMembers 返回工作空间的所有成员列表。
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

// changeMemberRole 调整指定成员角色；禁止修改自己的角色。
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

// removeMember 从工作空间移除指定成员；禁止移除自己。
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

// sendInvitation 向指定邮箱发送工作空间邀请，并记录审计日志。
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

// listInvitations 按可选状态过滤返回工作空间的邀请列表。
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

// revokeInvitation 撤销一条未使用的邀请，返回 204。
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

// acceptInvitation 使用邀请令牌接受工作空间邀请，加入对应工作空间。
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

// getInvitationPreview 根据邀请令牌返回邀请预览信息（不校验登录）。
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

// listProjects 返回指定工作空间下的所有项目列表。
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

// --- project modules DTO ↔ domain 转换 ---

// modulesDTOToDomain 将创建请求中的 Modules DTO 转换为域模型指针。
// nil 输入表示使用默认值（全部启用）。
func modulesDTOToDomain(m *struct {
	Sprint   *bool `json:"sprint,omitempty"`
	Version  *bool `json:"version,omitempty"`
	Estimate *bool `json:"estimate,omitempty"`
}) *workspace.ProjectModuleToggles {
	if m == nil {
		return nil
	}
	t := workspace.ProjectModuleAllEnabled()
	if m.Sprint != nil {
		t.Sprint = *m.Sprint
	}
	if m.Version != nil {
		t.Version = *m.Version
	}
	if m.Estimate != nil {
		t.Estimate = *m.Estimate
	}
	return &t
}

// modulesDTOToUpdateDomain 将更新请求中的 Modules DTO 转换为域模型指针。
// nil 输入表示不更新（返回 nil）。
func modulesDTOToUpdateDomain(m *struct {
	Sprint   *bool `json:"sprint,omitempty"`
	Version  *bool `json:"version,omitempty"`
	Estimate *bool `json:"estimate,omitempty"`
}) *workspace.ProjectModuleToggles {
	if m == nil {
		return nil
	}
	t := workspace.ProjectModuleAllEnabled()
	if m.Sprint != nil {
		t.Sprint = *m.Sprint
	}
	if m.Version != nil {
		t.Version = *m.Version
	}
	if m.Estimate != nil {
		t.Estimate = *m.Estimate
	}
	return &t
}

// createProject 在工作空间下创建项目并初始化状态模板，记录审计日志。
func createProject(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		var req dto.CreateProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		var coverImagePtr *string
		if req.CoverImageUrl != "" {
			coverImagePtr = &req.CoverImageUrl
		}
		p, err := d.ProjectSvc.Create(c.Request.Context(), workspace.ProjectCreateInput{
			WorkspaceID:  wsID,
			Name:         req.Name,
			Slug:         req.Slug,
			Identifier:   req.Identifier,
			Description:  req.Description,
			Network:      req.Network,
			Icon:         req.Icon,
			Color:        req.Color,
			Template:     req.Template,
			CreatedBy:    c.GetInt64(middleware.CtxUserID),
			Modules:      modulesDTOToDomain(req.Modules),
			CoverImageUrl: coverImagePtr,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		// 异步初始化项目状态模板（失败仅记录日志，不阻塞响应）。
		// 注意：调用方需确保 ProjectInitSvc 已在 Deps 中装配。
		if d.ProjectInitSvc != nil {
			// 捕获必要变量，避免 goroutine 闭包引用循环变量
			identifier := p.Identifier
			tpl := req.Template
			go func() {
				initCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				if err := d.ProjectInitSvc.InitializeForProject(initCtx, wsID, p.ID, tpl); err != nil {
					if d.Log != nil {
						d.Log.Error("project init failed",
							zap.String("template", tpl),
							zap.String("identifier", identifier),
							zap.Int64("project_id", p.ID),
							zap.Error(err),
						)
					}
				}
			}()
		}
		d.AuditSvc.RecordFromGin(c, wsID, "project.create", p.Identifier, map[string]any{
			"name": p.Name, "slug": p.Slug, "template": req.Template,
		})
		c.JSON(http.StatusCreated, p)
	}
}

// listProjectTemplates 返回预置项目模板列表（供前端模板选择器）。
func listProjectTemplates(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		tpls := d.TemplateSvc.ListTemplates()
		c.JSON(http.StatusOK, tpls)
	}
}
// getProject 返回指定项目的详情。
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

// updateProject 更新项目名称/描述/网络/图标/颜色/模块开关等信息。
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
			Name:          req.Name,
			Slug:          req.Slug,
			Description:   req.Description,
			Network:       req.Network,
			Icon:          req.Icon,
			Color:         req.Color,
			Modules:       modulesDTOToUpdateDomain(req.Modules),
			CoverImageUrl: req.CoverImageUrl,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, p)
	}
}

// ==================================================================
// 项目成员
// ==================================================================

// listProjectMembers 返回指定项目内的所有成员列表。
func listProjectMembers(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		projectID, _ := strconv.ParseInt(c.Param("project_id"), 10, 64)
		if projectID <= 0 {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field: "project_id", Reason: "无效的项目 ID",
			}))
			return
		}
		members, err := d.ProjectMemberSvc.List(c.Request.Context(), wsID, projectID)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, members)
	}
}

// addProjectMember 将工作空间成员加入项目。
func addProjectMember(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		projectID, _ := strconv.ParseInt(c.Param("project_id"), 10, 64)
		adderID := c.GetInt64(middleware.CtxUserID)
		if projectID <= 0 {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field: "project_id", Reason: "无效的项目 ID",
			}))
			return
		}
		var req dto.AddProjectMemberRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		if err := d.ProjectMemberSvc.AddMember(c.Request.Context(), wsID, projectID, req.UserID, adderID, req.Role); err != nil {
			writeError(c, err)
			return
		}
		d.AuditSvc.RecordFromGin(c, wsID, "project_member.add", strconv.FormatInt(req.UserID, 10), map[string]any{
			"project_id": projectID, "role": req.Role,
		})
		c.Status(http.StatusCreated)
	}
}

// changeProjectMemberRole 修改项目成员角色。
func changeProjectMemberRole(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		projectID, _ := strconv.ParseInt(c.Param("project_id"), 10, 64)
		targetID, _ := strconv.ParseInt(c.Param("user_id"), 10, 64)
		if projectID <= 0 {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field: "project_id", Reason: "无效的项目 ID",
			}))
			return
		}
		var req dto.ChangeRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		// 项目角色仅支持 admin / member
		if req.Role != "admin" && req.Role != "member" {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field: "role", Reason: "无效的项目角色",
			}))
			return
		}
		if err := d.ProjectMemberSvc.ChangeRole(c.Request.Context(), wsID, projectID, targetID, req.Role); err != nil {
			writeError(c, err)
			return
		}
		d.AuditSvc.RecordFromGin(c, wsID, "project_member.role_change", strconv.FormatInt(targetID, 10), map[string]any{
			"project_id": projectID, "new_role": req.Role,
		})
		c.Status(http.StatusNoContent)
	}
}

// removeProjectMember 从项目中移除成员。
func removeProjectMember(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		projectID, _ := strconv.ParseInt(c.Param("project_id"), 10, 64)
		targetID, _ := strconv.ParseInt(c.Param("user_id"), 10, 64)
		if projectID <= 0 {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field: "project_id", Reason: "无效的项目 ID",
			}))
			return
		}
		if err := d.ProjectMemberSvc.RemoveMember(c.Request.Context(), wsID, projectID, targetID); err != nil {
			writeError(c, err)
			return
		}
		d.AuditSvc.RecordFromGin(c, wsID, "project_member.remove", strconv.FormatInt(targetID, 10), map[string]any{
			"project_id": projectID,
		})
		c.Status(http.StatusNoContent)
	}
}

// archiveProject 归档指定项目，返回 204 并记录审计日志。
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

// listAuditLogs 返回工作空间的审计日志（默认 50 条，最多 200 条），
// 仅 owner/admin 可访问。
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

// ==================================================================
// 成员 CSV 批量导入
// ==================================================================

// importMembers 从上传的 CSV 文件批量导入成员。
func importMembers(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		inviterID := c.GetInt64(middleware.CtxUserID)

		file, _, err := c.Request.FormFile("file")
		if err != nil {
			writeError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "file", Reason: "请上传 CSV 文件"}))
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.LazyQuotes = true
		records, err := reader.ReadAll()
		if err != nil {
			writeError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "file", Reason: "CSV 文件解析失败: " + err.Error()}))
			return
		}

		if len(records) < 2 {
			writeError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "file", Reason: "CSV 格式错误：至少需要标题行+数据行"}))
			return
		}

		// 解析标题行，查找 email 和 name 列
		headers := records[0]
		emailIdx, nameIdx := -1, -1
		for i, h := range headers {
			switch strings.ToLower(strings.TrimSpace(h)) {
			case "email", "e-mail", "邮箱":
				emailIdx = i
			case "name", "姓名", "名称":
				nameIdx = i
			}
		}
		if emailIdx < 0 {
			writeError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "file", Reason: "CSV 缺少 email 列"}))
			return
		}

		type result struct {
			Email  string `json:"email"`
			Name   string `json:"name,omitempty"`
			Status string `json:"status"` // ok / exists / fail
			Reason string `json:"reason,omitempty"`
		}

		var results []result

		for i := 1; i < len(records); i++ {
			row := records[i]
			email := strings.TrimSpace(row[emailIdx])
			if email == "" {
				continue
			}
			name := ""
			if nameIdx >= 0 && nameIdx < len(row) {
				name = strings.TrimSpace(row[nameIdx])
			}

			// 发送邀请（使用默认 member 角色）
			_, invErr := d.InvitationSvc.Invite(c.Request.Context(), workspace.InviteInput{
				WorkspaceID: wsID,
				InviterID:   inviterID,
				Email:       email,
				Role:        "member",
			})
			if invErr != nil {
				results = append(results, result{Email: email, Name: name, Status: "fail", Reason: invErr.Error()})
			} else {
				results = append(results, result{Email: email, Name: name, Status: "ok"})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"results": results,
			"total":   len(results),
		})
	}
}
