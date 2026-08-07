// Package version — Version HTTP handlers（REST API）。
package version

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Handler 是 Gin handler 集合。
type Handler struct {
	svc *Service
}

// NewHandler 构造 handler。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 注册 Version 路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	// 集合
	r.GET("/versions", h.list)
	r.POST("/versions", h.create)
	r.GET("/versions/defects", h.filterDefects)

	// 单资源
	v := r.Group("/versions/:version_id")
	{
		v.GET("", h.get)
		v.PATCH("", h.update)
		v.DELETE("", h.delete)

		// 状态机
		v.POST("/activate", h.activate)
		v.POST("/release", h.release)
		v.POST("/archive", h.archive)

		// 进度 / 质量 / 交付
		v.GET("/progress", h.progress)
		v.GET("/quality", h.quality)
		v.GET("/delivery-report", h.deliveryReport)
		v.GET("/release-notes", h.releaseNotes)
		v.POST("/release-notes/regenerate", h.regenerateNotes)

		// 缺陷面板
		v.GET("/defects", h.defectPanel)

		// 迭代聚合
		v.GET("/sprints", h.listSprints)
		v.POST("/sprints", h.addSprint)
		v.DELETE("/sprints/:sprint_id", h.removeSprint)
	}
}

// --- request/response types ---

type createVersionRequest struct {
	Name        string          `json:"name" binding:"required,min=1,max=120"`
	Semver      string          `json:"semver" binding:"required,min=1,max=50"`
	Description string          `json:"description" binding:"max=2000"`
	TargetDate  *string         `json:"target_date" binding:"omitempty,datetime=2006-01-02"`
	Checklist   []ChecklistItem `json:"checklist"`
}

type updateVersionRequest struct {
	Name        *string         `json:"name" binding:"omitempty,min=1,max=120"`
	Description *string         `json:"description" binding:"omitempty,max=2000"`
	Semver      *string         `json:"semver" binding:"omitempty,min=1,max=50"`
	TargetDate  *string         `json:"target_date" binding:"omitempty,datetime=2006-01-02"`
	Checklist   []ChecklistItem `json:"checklist"`
	Version     int             `json:"version"`
}

type releaseVersionRequest struct {
	DraftOverride         string `json:"draft_override"`
	ForceChecklist        bool   `json:"force_checklist"`
	AddKnownIssuesToNotes bool   `json:"add_known_issues_to_notes"`
}

type addSprintRequest struct {
	SprintID int64 `json:"sprint_id" binding:"required,min=1"`
}

// --- 集合操作 ---

func (h *Handler) list(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	opts := ListVersionsOptions{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		Limit:       intQuery(c, "limit", 50),
		Offset:      intQuery(c, "offset", 0),
	}
	if v := c.Query("status"); v != "" {
		s := VersionStatusCode(v)
		if !s.IsValid() {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "status", Reason: "无效的状态值"}))
			return
		}
		opts.Status = &s
	}

	versions, total, err := h.svc.List(c.Request.Context(), opts)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": versions, "total": total})
}

func (h *Handler) create(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req createVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	in := CreateVersionInput{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		Name:        req.Name,
		Semver:      req.Semver,
		Description: req.Description,
		TargetDate:  req.TargetDate,
		Checklist:   req.Checklist,
		CreatedBy:   userID,
	}
	v, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, v)
}

// --- 单资源 ---

func (h *Handler) get(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")

	v, err := h.svc.GetByID(c.Request.Context(), wsID, versionID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *Handler) update(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")

	var req updateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}
	in := UpdateVersionInput{
		Name:        req.Name,
		Description: req.Description,
		Semver:      req.Semver,
		TargetDate:  req.TargetDate,
		Checklist:   req.Checklist,
		Version:     req.Version,
	}
	v, err := h.svc.Update(c.Request.Context(), wsID, versionID, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *Handler) delete(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")

	if err := h.svc.SoftDelete(c.Request.Context(), wsID, versionID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 状态机 ---

func (h *Handler) activate(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")

	v, err := h.svc.Activate(c.Request.Context(), wsID, versionID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *Handler) release(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")

	var req releaseVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	in := ReleaseVersionInput{
		DraftOverride:         req.DraftOverride,
		ForceChecklist:        req.ForceChecklist,
		AddKnownIssuesToNotes: req.AddKnownIssuesToNotes,
	}
	v, err := h.svc.Release(c.Request.Context(), wsID, versionID, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *Handler) archive(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")

	v, err := h.svc.Archive(c.Request.Context(), wsID, versionID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

// --- 进度 / 质量 / 报告 ---

func (h *Handler) progress(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")

	p, err := h.svc.Progress(c.Request.Context(), wsID, versionID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) quality(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")
	v, err := h.svc.GetByID(c.Request.Context(), wsID, versionID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, v.Quality)
}

func (h *Handler) deliveryReport(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")
	v, err := h.svc.GetByID(c.Request.Context(), wsID, versionID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, v.DeliveryReport)
}

func (h *Handler) releaseNotes(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")
	v, err := h.svc.GetByID(c.Request.Context(), wsID, versionID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"version_id": v.ID, "release_notes": v.ReleaseNotes})
}

func (h *Handler) regenerateNotes(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")

	v, err := h.svc.GetByID(c.Request.Context(), wsID, versionID)
	if err != nil {
		writeErr(c, err)
		return
	}
	// 仅 released 允许重生成（或 active 预览）
	if v.Status != VersionReleased && v.Status != VersionActive {
		middleware.AbortWithError(c, errs.ErrVersionInvalidLifecycle)
		return
	}
	src := h.svc.buildReleaseNotesSource(c.Request.Context(), wsID, v, true)
	notes := renderReleaseNotes(v, src)
	c.JSON(http.StatusOK, gin.H{"release_notes": notes})
}

// --- 缺陷面板 / 跨版本过滤 ---

func (h *Handler) defectPanel(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")

	views, total, err := h.svc.DefectPanel(c.Request.Context(), wsID, versionID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": views, "total": total})
}

func (h *Handler) filterDefects(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	f := BugVersionFilter{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		Limit:       intQuery(c, "limit", 50),
		Offset:      intQuery(c, "offset", 0),
	}
	if v := c.Query("found_version_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.FoundVersionID = &id
		}
	}
	if v := c.Query("fix_version_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.FixVersionID = &id
		}
	}
	if v := c.Query("state_group"); v != "" {
		f.StateGroup = &v
	}
	if v := c.Query("severity"); v != "" {
		if sev, err := strconv.Atoi(v); err == nil {
			f.Severity = &sev
		}
	}

	views, total, err := h.svc.FilterDefects(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": views, "total": total})
}

// --- 迭代聚合 ---

func (h *Handler) listSprints(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")

	v, err := h.svc.GetByID(c.Request.Context(), wsID, versionID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": v.Sprints})
}

func (h *Handler) addSprint(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")
	userID := c.GetInt64(middleware.CtxUserID)

	var req addSprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	if err := h.svc.AddSprint(c.Request.Context(), wsID, AddSprintInput{
		VersionID: versionID,
		SprintID:  req.SprintID,
		AddedBy:   userID,
	}); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) removeSprint(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	versionID := int64Param(c, "version_id")
	sprintID := int64Param(c, "sprint_id")

	if err := h.svc.RemoveSprint(c.Request.Context(), wsID, versionID, sprintID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- helpers ---

func int64Param(c *gin.Context, key string) int64 {
	v, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return v
}

func intQuery(c *gin.Context, key string, def int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return def
	}
	return v
}

func fieldDetail(err error) errs.FieldDetail {
	return errs.FieldDetail{Field: "body", Reason: err.Error()}
}

func writeErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}

