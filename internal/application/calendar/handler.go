// Package calendar — 项目日历 HTTP handlers。
//
// 路由挂载于 /api/v1/workspaces/:wid/calendar 与 /projects/:pid/calendar，
// 提供日历事件的 CRUD 端点。
package calendar

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps 日历 handler 依赖。
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

// RegisterRead 注册只读路由（列表 / 详情）。
func (h *Handler) RegisterRead(r *gin.RouterGroup) {
	g := r.Group("/calendar")
	g.GET("/events", h.ListEvents)
	g.GET("/events/:event_id", h.GetEvent)
}

// RegisterWrite 注册写路由（创建 / 更新 / 归档）。
func (h *Handler) RegisterWrite(r *gin.RouterGroup) {
	g := r.Group("/calendar")
	g.POST("/events", h.CreateEvent)
	g.PATCH("/events/:event_id", h.UpdateEvent)
	g.DELETE("/events/:event_id", h.ArchiveEvent)
}

// Register 注册全部路由（读 + 写），供单一权限层级的场景使用。
func (h *Handler) Register(r *gin.RouterGroup) {
	h.RegisterRead(r)
	h.RegisterWrite(r)
}

// ListEvents godoc
//
//	@Summary		日历事件列表
//	@Description	按时间范围 + 可选项目/类型过滤查询活跃事件
//	@Tags			calendar
//	@Produce		json
//	@Param			start		query	string	false	"开始时间（RFC3339）"
//	@Param			end		query	string	false	"结束时间（RFC3339）"
//	@Param			event_type	query	string	false	"事件类型"
//	@Param			limit		query	int		false	"每页条数，默认 50"
//	@Param			offset		query	int		false	"偏移量，默认 0"
//	@Success		200	{object}	map[string]any
//	@Router			/calendar/events [get]
func (h *Handler) ListEvents(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	filter := ListEventsFilter{
		WorkspaceID: wsID,
		Limit:       intQuery(c, "limit", 50),
		Offset:      intQuery(c, "offset", 0),
	}
	if projectID > 0 {
		filter.ProjectID = &projectID
	}
	if startStr := c.Query("start"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.Start = t
		}
	}
	if endStr := c.Query("end"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.End = t
		}
	}
	if et := c.Query("event_type"); et != "" {
		filter.EventType = EventType(et)
	}

	events, err := h.d.Svc.List(c.Request.Context(), filter)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": events})
}

// CreateEvent godoc
//
//	@Summary		创建日历事件
//	@Description	支持幂等键去重；event_type 必须是 meeting/milestone/release/deadline/other
//	@Tags			calendar
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	CalendarEvent
//	@Router			/calendar/events [post]
func (h *Handler) CreateEvent(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		Title          string  `json:"title" binding:"required"`
		Description    string  `json:"description"`
		StartTime      string  `json:"start_time" binding:"required"`
		EndTime        string  `json:"end_time" binding:"required"`
		IsAllDay       bool    `json:"is_all_day"`
		Location       string  `json:"location"`
		EventType      string  `json:"event_type" binding:"required"`
		SourceType     string  `json:"source_type"`
		SourceID       *int64  `json:"source_id"`
		IdempotencyKey string  `json:"idempotency_key"`
		OrganizerID    *int64  `json:"organizer_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "start_time", Reason: "invalid RFC3339"},
		))
		return
	}
	end, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "end_time", Reason: "invalid RFC3339"},
		))
		return
	}

	var projID *int64
	if projectID > 0 {
		projID = &projectID
	}
	organizer := req.OrganizerID
	if organizer == nil || *organizer == 0 {
		uid := userID
		organizer = &uid
	}

	event, err := h.d.Svc.Create(c.Request.Context(), CreateEventInput{
		TenantID:       1,
		WorkspaceID:    wsID,
		ProjectID:      projID,
		Title:          req.Title,
		Description:    req.Description,
		StartTime:      start,
		EndTime:        end,
		IsAllDay:       req.IsAllDay,
		Location:       req.Location,
		EventType:      EventType(req.EventType),
		SourceType:     req.SourceType,
		SourceID:       req.SourceID,
		IdempotencyKey: req.IdempotencyKey,
		OrganizerID:    organizer,
		CreatedBy:      userID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, event)
}

// GetEvent godoc
//
//	@Summary		日历事件详情
//	@Tags			calendar
//	@Produce		json
//	@Success		200	{object}	CalendarEvent
//	@Router			/calendar/events/{event_id} [get]
func (h *Handler) GetEvent(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("event_id"), 10, 64)
	if err != nil || id <= 0 {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "event_id", Reason: "invalid"},
		))
		return
	}
	event, err := h.d.Svc.GetByID(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
}

// UpdateEvent godoc
//
//	@Summary		更新日历事件
//	@Description	乐观锁：version 不匹配返回 409
//	@Tags			calendar
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	CalendarEvent
//	@Router			/calendar/events/{event_id} [patch]
func (h *Handler) UpdateEvent(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("event_id"), 10, 64)
	if err != nil || id <= 0 {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "event_id", Reason: "invalid"},
		))
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		StartTime   string `json:"start_time" binding:"required"`
		EndTime     string `json:"end_time" binding:"required"`
		IsAllDay    bool   `json:"is_all_day"`
		Location    string `json:"location"`
		EventType   string `json:"event_type" binding:"required"`
		OrganizerID *int64 `json:"organizer_id"`
		Version     int    `json:"version" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "start_time", Reason: "invalid RFC3339"},
		))
		return
	}
	end, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "end_time", Reason: "invalid RFC3339"},
		))
		return
	}

	event, err := h.d.Svc.Update(c.Request.Context(), UpdateEventInput{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		StartTime:   start,
		EndTime:     end,
		IsAllDay:    req.IsAllDay,
		Location:    req.Location,
		EventType:   EventType(req.EventType),
		OrganizerID: req.OrganizerID,
		Version:     req.Version,
		UpdatedBy:   userID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
}

// ArchiveEvent godoc
//
//	@Summary		归档日历事件
//	@Description	软删除语义：置为 archived，保留历史
//	@Tags			calendar
//	@Success		204
//	@Router			/calendar/events/{event_id} [delete]
func (h *Handler) ArchiveEvent(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("event_id"), 10, 64)
	if err != nil || id <= 0 {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "event_id", Reason: "invalid"},
		))
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)
	if err := h.d.Svc.Archive(c.Request.Context(), id, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Helpers ---

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
