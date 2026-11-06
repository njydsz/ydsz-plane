// Package milestone — 项目里程碑 HTTP handlers。
//
// 路由挂载：
//   /api/v1/workspaces/:workspace_id/projects/:project_id/milestones
package milestone

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
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
	g := r.Group("/milestones")
	g.GET("", h.list)
	g.GET("/:milestone_id", h.get)
	g.GET("/:milestone_id/progress", h.getProgress)
	g.GET("/:milestone_id/work-items", h.listWorkItems)
}

// RegisterWrite 注册写入路由。
func (h *Handler) RegisterWrite(r *gin.RouterGroup) {
	g := r.Group("/milestones")
	g.POST("", h.create)
	g.PATCH("/:milestone_id", h.update)
	g.DELETE("/:milestone_id", h.delete)
	g.POST("/:milestone_id/work-items", h.linkWorkItem)
	g.DELETE("/:milestone_id/work-items/:work_item_id", h.unlinkWorkItem)
}

// Register 注册全部路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	h.RegisterRead(r)
	h.RegisterWrite(r)
}

// --- helpers ---

func writeErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	switch err {
	case ErrMilestoneNotFound, ErrWorkItemNotFound:
		middleware.AbortWithError(c, errs.ErrNotFound)
	case ErrAlreadyLinked:
		middleware.AbortWithError(c, errs.New("MILESTONE.ALREADY_LINKED", "工作项已关联到该里程碑", http.StatusConflict))
	default:
		middleware.AbortWithError(c, errs.ErrInternal)
	}
}

func intParam(c *gin.Context, key string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil || id <= 0 {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: key, Reason: "无效的 ID"})
	}
	return id, nil
}

func intQuery(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func fieldDetail(err error) errs.FieldDetail {
	return errs.FieldDetail{Field: "body", Reason: err.Error()}
}

func parseDateParam(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "target_date", Reason: "无效的日期格式（YYYY-MM-DD）"})
	}
	return &t, nil
}

// --- handlers ---

// list 列出项目下的里程碑。
func (h *Handler) list(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	opts := ListOptions{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		Status:      MilestoneStatus(c.Query("status")),
		Limit:       intQuery(c, "limit", 100),
		Offset:      intQuery(c, "offset", 0),
	}
	list, err := h.d.Svc.List(c.Request.Context(), opts)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": list})
}

// get 查询里程碑详情。
func (h *Handler) get(c *gin.Context) {
	id, err := intParam(c, "milestone_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	m, err := h.d.Svc.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}

// create 创建里程碑。
func (h *Handler) create(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		TargetDate  string `json:"target_date"`
		Status      string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}
	targetDate, err := parseDateParam(req.TargetDate)
	if err != nil {
		writeErr(c, err)
		return
	}

	m, err := h.d.Svc.Create(c.Request.Context(), CreateInput{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
		TargetDate:  targetDate,
		Status:      MilestoneStatus(req.Status),
		CreatedBy:   userID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, m)
}

// update 更新里程碑。
func (h *Handler) update(c *gin.Context) {
	id, err := intParam(c, "milestone_id")
	if err != nil {
		writeErr(c, err)
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		TargetDate  *string `json:"target_date"`
		Status      *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	in := UpdateInput{
		Name:        req.Name,
		Description: req.Description,
	}
	if req.TargetDate != nil {
		td, err := parseDateParam(*req.TargetDate)
		if err != nil {
			writeErr(c, err)
			return
		}
		in.TargetDate = td
	}
	if req.Status != nil {
		st := MilestoneStatus(*req.Status)
		in.Status = &st
	}

	m, err := h.d.Svc.Update(c.Request.Context(), id, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}

// delete 删除里程碑。
func (h *Handler) delete(c *gin.Context) {
	id, err := intParam(c, "milestone_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	if err := h.d.Svc.Delete(c.Request.Context(), id); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// linkWorkItem 关联工作项到里程碑。
func (h *Handler) linkWorkItem(c *gin.Context) {
	milestoneID, err := intParam(c, "milestone_id")
	if err != nil {
		writeErr(c, err)
		return
	}

	var req struct {
		WorkItemID   int64  `json:"work_item_id"`
		WorkItemType string `json:"work_item_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}
	if req.WorkItemID <= 0 {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_id", Reason: "无效的工作项 ID"}))
		return
	}

	if err := h.d.Svc.LinkWorkItem(c.Request.Context(), LinkInput{
		MilestoneID:  milestoneID,
		WorkItemID:   req.WorkItemID,
		WorkItemType: req.WorkItemType,
	}); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// unlinkWorkItem 解除关联。
func (h *Handler) unlinkWorkItem(c *gin.Context) {
	milestoneID, err := intParam(c, "milestone_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	workItemID, err := intParam(c, "work_item_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	if err := h.d.Svc.UnlinkWorkItem(c.Request.Context(), milestoneID, workItemID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// listWorkItems 列出里程碑关联的工作项。
func (h *Handler) listWorkItems(c *gin.Context) {
	milestoneID, err := intParam(c, "milestone_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	items, err := h.d.Svc.ListWorkItems(c.Request.Context(), milestoneID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": items})
}

// getProgress 获取里程碑进度聚合。
func (h *Handler) getProgress(c *gin.Context) {
	milestoneID, err := intParam(c, "milestone_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	p, err := h.d.Svc.GetProgress(c.Request.Context(), milestoneID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}
