// Package integration — 集成 HTTP handlers。
//
// 路由挂载于 /api/v1/workspaces/:workspace_id/integrations：
//   - GET    /                     列表
//   - POST   /                     创建 / 连接
//   - GET    /providers            可用 provider 元数据
//   - GET    /:id                  详情
//   - PATCH  /:id                  更新
//   - DELETE /:id                  删除
//   - POST   /:id/test             测试连接
//   - POST   /:id/sync             触发同步
//   - POST   /:id/pause            暂停
//   - POST   /:id/resume           恢复
//   - POST   /:id/disconnect       断开
//   - GET    /:id/events           事件日志
package integration

import (
	"net/http"
	"strconv"

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
func NewHandler(d *HandlerDeps) *Handler { return &Handler{d: d} }

// Register 注册全部集成路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/integrations")
	{
		g.GET("", h.list)
		g.POST("", h.create)
		g.GET("/providers", h.listProviders)

		g.GET("/:id", h.get)
		g.PATCH("/:id", h.update)
		g.DELETE("/:id", h.remove)

		g.POST("/:id/test", h.testConnection)
		g.POST("/:id/sync", h.syncNow)
		g.POST("/:id/pause", h.pause)
		g.POST("/:id/resume", h.resume)
		g.POST("/:id/disconnect", h.disconnect)

		g.GET("/:id/events", h.listEvents)
	}
}

// RegisterRead 注册只读路由（保留签名兼容）。
func (h *Handler) RegisterRead(r *gin.RouterGroup) {
	r.Group("/integrations").GET("", h.list)
	r.Group("/integrations").GET("/providers", h.listProviders)
	r.Group("/integrations").GET("/:id", h.get)
	r.Group("/integrations").GET("/:id/events", h.listEvents)
}

// RegisterWrite 注册写路由（保留签名兼容）。
func (h *Handler) RegisterWrite(r *gin.RouterGroup) {
	g := r.Group("/integrations")
	g.POST("", h.create)
	g.PATCH("/:id", h.update)
	g.DELETE("/:id", h.remove)
	g.POST("/:id/test", h.testConnection)
	g.POST("/:id/sync", h.syncNow)
	g.POST("/:id/pause", h.pause)
	g.POST("/:id/resume", h.resume)
	g.POST("/:id/disconnect", h.disconnect)
}

// ---- handlers ----

// list GET /integrations?provider=&status=&limit=&offset=
//
//	@Summary		集成列表
//	@Description	按 provider / status 分页列出工作空间内的集成
//	@Tags			integration
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/integrations [get]
func (h *Handler) list(c *gin.Context) {
	opts := ListIntegrationsOptions{
		WorkspaceID: wsID(c),
		Provider:    c.Query("provider"),
		Status:      c.Query("status"),
		Limit:       intQuery(c, "limit", 50),
		Offset:      intQuery(c, "offset", 0),
	}
	items, total, err := h.d.Svc.List(c.Request.Context(), opts)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

// listProviders GET /integrations/providers
//
//	@Summary		可用 provider 列表
//	@Description	返回全部可用 provider 的元数据（名称 / 图标 / 凭证字段 / 文档 URL）
//	@Tags			integration
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/integrations/providers [get]
func (h *Handler) listProviders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": h.d.Svc.ListProviders()})
}

// get GET /integrations/:id
//
//	@Summary		集成详情
//	@Description	返回单个集成配置（不含凭证）
//	@Tags			integration
//	@Produce		json
//	@Param			id	path	int	true	"集成 ID"
//	@Success		200	{object}	Integration
//	@Failure		404	{object}	errs.AppError
//	@Router			/integrations/{id} [get]
func (h *Handler) get(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	itg, err := h.d.Svc.GetByIDPublic(c.Request.Context(), wsID(c), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, itg)
}

type createRequest struct {
	Provider    string            `json:"provider" binding:"required"`
	Name        string            `json:"name"`
	Credentials map[string]string `json:"credentials" binding:"required"`
	Config      map[string]string `json:"config"`
}

// create POST /integrations
//
//	@Summary		创建 / 连接集成
//	@Description	校验凭证 + 测试连接 + 加密存储，创建集成实例
//	@Tags			integration
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	Integration
//	@Failure		422	{object}	errs.AppError
//	@Router			/integrations [post]
func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	itg, err := h.d.Svc.Create(c.Request.Context(), CreateIntegrationInput{
		WorkspaceID: wsID(c),
		Provider:    req.Provider,
		Name:        req.Name,
		Credentials: req.Credentials,
		Config:      req.Config,
		CreatedBy:   actorID(c),
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, itg)
}

type updateRequest struct {
	Name        *string           `json:"name"`
	Credentials map[string]string `json:"credentials"`
	Config      map[string]string `json:"config"`
}

// update PATCH /integrations/:id
//
//	@Summary		更新集成
//	@Description	更新名称 / 凭证 / 配置
//	@Tags			integration
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"集成 ID"
//	@Success		200	{object}	Integration
//	@Router			/integrations/{id} [patch]
func (h *Handler) update(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	itg, err := h.d.Svc.Update(c.Request.Context(), wsID(c), id, UpdateIntegrationInput{
		Name:        req.Name,
		Credentials: req.Credentials,
		Config:      req.Config,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, itg)
}

// remove DELETE /integrations/:id
//
//	@Summary		删除集成
//	@Description	删除集成实例及其事件日志
//	@Tags			integration
//	@Param			id	path	int	true	"集成 ID"
//	@Success		204
//	@Router			/integrations/{id} [delete]
func (h *Handler) remove(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	if err := h.d.Svc.Delete(c.Request.Context(), wsID(c), id); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// testConnection POST /integrations/:id/test
//
//	@Summary		测试连接
//	@Description	使用已存储凭证测试与第三方服务的连通性
//	@Tags			integration
//	@Produce		json
//	@Param			id	path	int	true	"集成 ID"
//	@Success		200	{object}	map[string]any
//	@Router			/integrations/{id}/test [post]
func (h *Handler) testConnection(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	if err := h.d.Svc.TestConnection(c.Request.Context(), wsID(c), id); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "连接测试成功"})
}

// syncNow POST /integrations/:id/sync
//
//	@Summary		触发同步
//	@Description	立即从第三方服务拉取工作项并记录事件
//	@Tags			integration
//	@Produce		json
//	@Param			id	path	int	true	"集成 ID"
//	@Success		200	{object}	SyncResult
//	@Router			/integrations/{id}/sync [post]
func (h *Handler) syncNow(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	result, err := h.d.Svc.SyncNow(c.Request.Context(), wsID(c), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// pause POST /integrations/:id/pause
//
//	@Summary		暂停集成
//	@Tags		integration
//	@Produce	json
//	@Param		id	path	int	true	"集成 ID"
//	@Success	200	{object}	Integration
//	@Router		/integrations/{id}/pause [post]
func (h *Handler) pause(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	itg, err := h.d.Svc.Pause(c.Request.Context(), wsID(c), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, itg)
}

// resume POST /integrations/:id/resume
//
//	@Summary		恢复集成
//	@Tags		integration
//	@Produce	json
//	@Param		id	path	int	true	"集成 ID"
//	@Success	200	{object}	Integration
//	@Router		/integrations/{id}/resume [post]
func (h *Handler) resume(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	itg, err := h.d.Svc.Resume(c.Request.Context(), wsID(c), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, itg)
}

// disconnect POST /integrations/:id/disconnect
//
//	@Summary		断开集成
//	@Description	删除凭证并将状态置为 pending
//	@Tags			integration
//	@Produce		json
//	@Param			id	path	int	true	"集成 ID"
//	@Success		200	{object}	Integration
//	@Router			/integrations/{id}/disconnect [post]
func (h *Handler) disconnect(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	itg, err := h.d.Svc.Disconnect(c.Request.Context(), wsID(c), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, itg)
}

// listEvents GET /integrations/:id/events?direction=&status=&event_type=&limit=&offset=
//
//	@Summary		事件日志
//	@Description	分页查询指定集成的事件日志
//	@Tags			integration
//	@Produce		json
//	@Param			id			path	int		true	"集成 ID"
//	@Param			direction	query	string	false	"inbound / outbound"
//	@Param			status		query	string	false	"success / failed / ignored"
//	@Param			event_type	query	string	false	"事件类型"
//	@Success			200			{object}	map[string]any
//	@Router			/integrations/{id}/events [get]
func (h *Handler) listEvents(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	opts := ListEventsOptions{
		WorkspaceID:   wsID(c),
		IntegrationID: &id,
		Direction:     c.Query("direction"),
		Status:        c.Query("status"),
		EventType:     c.Query("event_type"),
		Limit:         intQuery(c, "limit", 50),
		Offset:        intQuery(c, "offset", 0),
	}
	items, total, err := h.d.Svc.ListEvents(c.Request.Context(), opts)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

// ---- helpers ----

func wsID(c *gin.Context) int64 { return c.GetInt64(middleware.CtxWorkspaceID) }

func actorID(c *gin.Context) int64 { return c.GetInt64(middleware.CtxUserID) }

func idParam(c *gin.Context, name string) (int64, error) {
	v, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || v <= 0 {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: name, Reason: "无效的 ID"})
	}
	return v, nil
}

func intQuery(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func writeErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}
