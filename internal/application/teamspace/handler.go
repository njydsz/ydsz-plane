// Package teamspace — 团队空间 HTTP handlers。
//
// 路由挂载于 /api/v1/workspaces/:workspace_id/teamspaces。
package teamspace

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
)

// HandlerDeps handler 依赖。
type HandlerDeps struct {
	Svc *Service
}

// Handler Gin handler 集合。
type Handler struct {
	d *HandlerDeps
}

// NewHandler 构造 handler。
func NewHandler(d *HandlerDeps) *Handler {
	return &Handler{d: d}
}

// RegisterRead 注册只读路由。
func (h *Handler) RegisterRead(r *gin.RouterGroup) {
	g := r.Group("/teamspaces")
	g.GET("", h.List)
	g.GET("/:teamspace_id", h.Get)
	g.GET("/:teamspace_id/members", h.ListMembers)
	g.GET("/:teamspace_id/projects", h.ListProjects)
	g.GET("/:teamspace_id/progress", h.GetProgress)
	g.GET("/:teamspace_id/workloads", h.ListWorkloads)
}

// RegisterWrite 注册写路由。
func (h *Handler) RegisterWrite(r *gin.RouterGroup) {
	g := r.Group("/teamspaces")
	g.POST("", h.Create)
	g.PATCH("/:teamspace_id", h.Update)
	g.DELETE("/:teamspace_id", h.Delete)
	g.POST("/:teamspace_id/archive", h.Archive)
	g.POST("/:teamspace_id/restore", h.Restore)
	g.POST("/:teamspace_id/members", h.AddMember)
	g.DELETE("/:teamspace_id/members/:user_id", h.RemoveMember)
	g.POST("/:teamspace_id/projects", h.AddProject)
	g.DELETE("/:teamspace_id/projects/:project_id", h.RemoveProject)
}

// Register 注册全部路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	h.RegisterRead(r)
	h.RegisterWrite(r)
}

// --- helpers ---

func (h *Handler) fail(c *gin.Context, err error) {
	switch err {
	case ErrTeamspaceNotFound, ErrTeamspaceArchived:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		if ec, ok := err.(interface{ HTTPStatus() int }); ok {
			c.JSON(ec.HTTPStatus(), gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func intQuery(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func paramInt(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return n
}

// --- handlers ---

// List godoc
//
//	@Summary		团队空间列表
//	@Tags			teamspace
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/teamspaces [get]
func (h *Handler) List(c *gin.Context) {
	items, err := h.d.Svc.List(c.Request.Context(), ListTeamspacesOptions{
		WorkspaceID: c.GetInt64(middleware.CtxWorkspaceID),
		Limit:       intQuery(c, "limit", 100),
		Offset:      intQuery(c, "offset", 0),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Get godoc
//
//	@Summary		团队空间详情
//	@Tags			teamspace
//	@Produce		json
//	@Success		200	{object}	Teamspace
//	@Router			/teamspaces/{teamspace_id} [get]
func (h *Handler) Get(c *gin.Context) {
	ts, err := h.d.Svc.GetByID(c.Request.Context(), paramInt(c, "teamspace_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, ts)
}

// Create godoc
//
//	@Summary		创建团队空间
//	@Tags			teamspace
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	Teamspace
//	@Router			/teamspaces [post]
func (h *Handler) Create(c *gin.Context) {
	var body struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Color       string  `json:"color"`
		LeadID      int64   `json:"lead_id"`
		MemberIDs   []int64 `json:"member_ids"`
		ProjectIDs  []int64 `json:"project_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ts, err := h.d.Svc.Create(c.Request.Context(), CreateTeamspaceInput{
		WorkspaceID: c.GetInt64(middleware.CtxWorkspaceID),
		Name:        body.Name,
		Description: body.Description,
		Color:       body.Color,
		LeadID:      body.LeadID,
		CreatedBy:   c.GetInt64(middleware.CtxUserID),
		MemberIDs:   body.MemberIDs,
		ProjectIDs:  body.ProjectIDs,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, ts)
}

// Update godoc
//
//	@Summary		更新团队空间
//	@Tags			teamspace
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Teamspace
//	@Router			/teamspaces/{teamspace_id} [patch]
func (h *Handler) Update(c *gin.Context) {
	var body struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Color       *string `json:"color"`
		LeadID      *int64  `json:"lead_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ts, err := h.d.Svc.Update(c.Request.Context(), paramInt(c, "teamspace_id"), UpdateTeamspaceInput{
		Name:        body.Name,
		Description: body.Description,
		Color:       body.Color,
		LeadID:      body.LeadID,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, ts)
}

// Delete godoc
//
//	@Summary		删除团队空间
//	@Tags			teamspace
//	@Produce		json
//	@Success		204
//	@Router			/teamspaces/{teamspace_id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if err := h.d.Svc.Delete(c.Request.Context(), paramInt(c, "teamspace_id")); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Archive godoc
//
//	@Summary		归档团队空间
//	@Tags			teamspace
//	@Produce		json
//	@Success		200	{object}	Teamspace
//	@Router			/teamspaces/{teamspace_id}/archive [post]
func (h *Handler) Archive(c *gin.Context) {
	if err := h.d.Svc.Archive(c.Request.Context(), paramInt(c, "teamspace_id")); err != nil {
		h.fail(c, err)
		return
	}
	ts, _ := h.d.Svc.GetByID(c.Request.Context(), paramInt(c, "teamspace_id"))
	c.JSON(http.StatusOK, ts)
}

// Restore godoc
//
//	@Summary		恢复团队空间
//	@Tags			teamspace
//	@Produce		json
//	@Success		200	{object}	Teamspace
//	@Router			/teamspaces/{teamspace_id}/restore [post]
func (h *Handler) Restore(c *gin.Context) {
	if err := h.d.Svc.Restore(c.Request.Context(), paramInt(c, "teamspace_id")); err != nil {
		h.fail(c, err)
		return
	}
	ts, _ := h.d.Svc.GetByID(c.Request.Context(), paramInt(c, "teamspace_id"))
	c.JSON(http.StatusOK, ts)
}

// AddMember godoc
//
//	@Summary		添加团队成员
//	@Tags			teamspace
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Router			/teamspaces/{teamspace_id}/members [post]
func (h *Handler) AddMember(c *gin.Context) {
	var body struct {
		UserID int64  `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.d.Svc.AddMember(c.Request.Context(), paramInt(c, "teamspace_id"), body.UserID, MemberRole(body.Role)); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveMember godoc
//
//	@Summary		移除团队成员
//	@Tags			teamspace
//	@Produce		json
//	@Success		204
//	@Router			/teamspaces/{teamspace_id}/members/{user_id} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	if err := h.d.Svc.RemoveMember(c.Request.Context(), paramInt(c, "teamspace_id"), paramInt(c, "user_id")); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListMembers godoc
//
//	@Summary		团队成员列表
//	@Tags			teamspace
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/teamspaces/{teamspace_id}/members [get]
func (h *Handler) ListMembers(c *gin.Context) {
	items, err := h.d.Svc.ListMembers(c.Request.Context(), paramInt(c, "teamspace_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// AddProject godoc
//
//	@Summary		关联项目
//	@Tags			teamspace
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Router			/teamspaces/{teamspace_id}/projects [post]
func (h *Handler) AddProject(c *gin.Context) {
	var body struct {
		ProjectID int64 `json:"project_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.d.Svc.AddProject(c.Request.Context(), paramInt(c, "teamspace_id"), body.ProjectID); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveProject godoc
//
//	@Summary		解除项目关联
//	@Tags			teamspace
//	@Produce		json
//	@Success		204
//	@Router			/teamspaces/{teamspace_id}/projects/{project_id} [delete]
func (h *Handler) RemoveProject(c *gin.Context) {
	if err := h.d.Svc.RemoveProject(c.Request.Context(), paramInt(c, "teamspace_id"), paramInt(c, "project_id")); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListProjects godoc
//
//	@Summary		关联项目列表
//	@Tags			teamspace
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/teamspaces/{teamspace_id}/projects [get]
func (h *Handler) ListProjects(c *gin.Context) {
	items, err := h.d.Svc.ListProjects(c.Request.Context(), paramInt(c, "teamspace_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GetProgress godoc
//
//	@Summary		团队进度汇总
//	@Tags			teamspace
//	@Produce		json
//	@Success		200	{object}	TeamspaceProgress
//	@Router			/teamspaces/{teamspace_id}/progress [get]
func (h *Handler) GetProgress(c *gin.Context) {
	p, err := h.d.Svc.GetProgress(c.Request.Context(), paramInt(c, "teamspace_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

// ListWorkloads godoc
//
//	@Summary		成员工作负载
//	@Tags			teamspace
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/teamspaces/{teamspace_id}/workloads [get]
func (h *Handler) ListWorkloads(c *gin.Context) {
	items, err := h.d.Svc.ListMemberWorkloads(c.Request.Context(), paramInt(c, "teamspace_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
