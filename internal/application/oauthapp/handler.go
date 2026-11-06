// Package oauthapp — OAuth 应用与授权 HTTP handlers。
//
// 路由挂载：
//   - /api/v1/workspaces/:workspace_id/oauth-apps/...（应用管理，工作空间级）
//   - /api/v1/oauth/authorize、/oauth/token、/oauth/revoke、/oauth/apps/:client_id（OAuth2 端点）
package oauthapp

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

// --- 应用管理（工作空间级） ---

// RegisterWorkspace 注册工作空间级应用管理路由。
func (h *Handler) RegisterWorkspace(r *gin.RouterGroup) {
	g := r.Group("/oauth-apps")
	g.GET("", h.ListApps)
	g.POST("", h.CreateApp)
	g.GET("/:app_id", h.GetApp)
	g.PATCH("/:app_id", h.UpdateApp)
	g.DELETE("/:app_id", h.DeleteApp)
	g.POST("/:app_id/rotate", h.RotateSecret)
}

// --- OAuth2 端点 ---

// RegisterOAuth 注册 OAuth2 端点路由。
func (h *Handler) RegisterOAuth(r *gin.RouterGroup) {
	r.GET("/authorize", h.Authorize)
	r.POST("/authorize", h.ApproveAuthorize)
	r.POST("/token", h.ExchangeOrRefresh)
	r.POST("/revoke", h.RevokeToken)
	r.GET("/apps/:client_id", h.PublicAppInfo)
}

// --- helpers ---

func (h *Handler) fail(c *gin.Context, err error) {
	switch err {
	case ErrAppNotFound, ErrCodeNotFound, ErrTokenNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrAppRevoked, ErrInvalidClientSecret, ErrRedirectURINotAllowed:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
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

// ListApps godoc
//
//	@Summary		OAuth 应用列表
//	@Tags			oauthapp
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/oauth-apps [get]
func (h *Handler) ListApps(c *gin.Context) {
	items, err := h.d.Svc.ListApps(c.Request.Context(), ListAppsOptions{
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

// CreateApp godoc
//
//	@Summary		创建 OAuth 应用
//	@Tags			oauthapp
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	OAuthApp
//	@Router			/oauth-apps [post]
func (h *Handler) CreateApp(c *gin.Context) {
	var body struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		RedirectURIs []string `json:"redirect_uris"`
		Scopes       []string `json:"scopes"`
		IconURL      string   `json:"icon_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	app, err := h.d.Svc.CreateApp(c.Request.Context(), CreateAppInput{
		WorkspaceID:  c.GetInt64(middleware.CtxWorkspaceID),
		Name:         body.Name,
		Description:  body.Description,
		RedirectURIs: body.RedirectURIs,
		Scopes:       body.Scopes,
		IconURL:      body.IconURL,
		CreatedBy:    c.GetInt64(middleware.CtxUserID),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	// client_secret 仅在创建时返回一次（OAuthApp.ClientSecret 标记为 json:"-"）。
	c.JSON(http.StatusCreated, gin.H{
		"app":           app,
		"client_id":     app.ClientID,
		"client_secret": app.ClientSecret,
	})
}

// GetApp godoc
//
//	@Summary		OAuth 应用详情
//	@Tags			oauthapp
//	@Produce		json
//	@Success		200	{object}	OAuthApp
//	@Router			/oauth-apps/{app_id} [get]
func (h *Handler) GetApp(c *gin.Context) {
	app, err := h.d.Svc.GetApp(c.Request.Context(), paramInt(c, "app_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, app)
}

// UpdateApp godoc
//
//	@Summary		更新 OAuth 应用
//	@Tags			oauthapp
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	OAuthApp
//	@Router			/oauth-apps/{app_id} [patch]
func (h *Handler) UpdateApp(c *gin.Context) {
	var body struct {
		Name         *string   `json:"name"`
		Description  *string   `json:"description"`
		RedirectURIs *[]string `json:"redirect_uris"`
		Scopes       *[]string `json:"scopes"`
		IconURL      *string   `json:"icon_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	app, err := h.d.Svc.UpdateApp(c.Request.Context(), paramInt(c, "app_id"), UpdateAppInput{
		Name:         body.Name,
		Description:  body.Description,
		RedirectURIs: body.RedirectURIs,
		Scopes:       body.Scopes,
		IconURL:      body.IconURL,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, app)
}

// DeleteApp godoc
//
//	@Summary		删除 OAuth 应用
//	@Tags			oauthapp
//	@Produce		json
//	@Success		204
//	@Router			/oauth-apps/{app_id} [delete]
func (h *Handler) DeleteApp(c *gin.Context) {
	if err := h.d.Svc.DeleteApp(c.Request.Context(), paramInt(c, "app_id")); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RotateSecret godoc
//
//	@Summary		轮换 client_secret
//	@Tags			oauthapp
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/oauth-apps/{app_id}/rotate [post]
func (h *Handler) RotateSecret(c *gin.Context) {
	secret, err := h.d.Svc.RotateSecret(c.Request.Context(), paramInt(c, "app_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"client_secret": secret})
}

// Authorize godoc
//
//	@Summary		显示授权同意页（GET /oauth/authorize）
//	@Description	校验 client_id 与 redirect_uri 后返回待授权信息（由前端渲染 consent 页）。
//	@Tags			oauthapp
//	@Produce		json
//	@Router			/oauth/authorize [get]
func (h *Handler) Authorize(c *gin.Context) {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	scope := c.Query("scope")
	state := c.Query("state")

	app, err := h.d.Svc.GetAppByClientID(c.Request.Context(), clientID)
	if err != nil {
		h.fail(c, err)
		return
	}
	if app.Status == AppStatusRevoked {
		h.fail(c, ErrAppRevoked)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"app": gin.H{
			"id":          app.ID,
			"name":        app.Name,
			"description": app.Description,
			"icon_url":    app.IconURL,
		},
		"redirect_uri": redirectURI,
		"scope":        scope,
		"state":        state,
	})
}

// ApproveAuthorize godoc
//
//	@Summary		用户同意授权（POST /oauth/authorize）
//	@Description	校验后签发一次性授权码，返回 code（前端再 302 回 redirect_uri）。
//	@Tags			oauthapp
//	@Accept			json
//	@Produce		json
//	@Router			/oauth/authorize [post]
func (h *Handler) ApproveAuthorize(c *gin.Context) {
	var body struct {
		ClientID    string `json:"client_id"`
		RedirectURI string `json:"redirect_uri"`
		Scope       string `json:"scope"`
		State       string `json:"state"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	code, err := h.d.Svc.Authorize(c.Request.Context(), AuthorizeInput{
		ClientID:    body.ClientID,
		RedirectURI: body.RedirectURI,
		Scope:       body.Scope,
		State:       body.State,
		UserID:      c.GetInt64(middleware.CtxUserID),
		WorkspaceID: c.GetInt64(middleware.CtxWorkspaceID),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": code.Code, "state": body.State})
}

// ExchangeOrRefresh godoc
//
//	@Summary		令牌端点（POST /oauth/token）
//	@Description	支持 authorization_code 换取与 refresh_token 刷新两种 grant_type。
//	@Tags			oauthapp
//	@Accept			json
//	@Produce		json
//	@Router			/oauth/token [post]
func (h *Handler) ExchangeOrRefresh(c *gin.Context) {
	var body struct {
		GrantType    string `json:"grant_type"`
		Code         string `json:"code"`
		RefreshToken string `json:"refresh_token"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURI  string `json:"redirect_uri"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	switch body.GrantType {
	case "authorization_code":
		resp, err := h.d.Svc.ExchangeCode(c.Request.Context(), ExchangeInput{
			Code:         body.Code,
			ClientID:     body.ClientID,
			ClientSecret: body.ClientSecret,
			RedirectURI:  body.RedirectURI,
		})
		if err != nil {
			h.fail(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	case "refresh_token":
		resp, err := h.d.Svc.RefreshToken(c.Request.Context(), RefreshInput{
			RefreshToken: body.RefreshToken,
			ClientID:     body.ClientID,
			ClientSecret: body.ClientSecret,
		})
		if err != nil {
			h.fail(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported grant_type"})
	}
}

// RevokeToken godoc
//
//	@Summary		吊销令牌（POST /oauth/revoke）
//	@Tags			oauthapp
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Router			/oauth/revoke [post]
func (h *Handler) RevokeToken(c *gin.Context) {
	var body struct {
		Token string `json:"token"`
	}
	_ = c.ShouldBindJSON(&body)
	if err := h.d.Svc.RevokeToken(c.Request.Context(), body.Token); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// PublicAppInfo godoc
//
//	@Summary		公开应用信息（GET /oauth/apps/:client_id）
//	@Description	按 client_id 返回应用公开信息（不含 secret）。
//	@Tags			oauthapp
//	@Produce		json
//	@Router			/oauth/apps/{client_id} [get]
func (h *Handler) PublicAppInfo(c *gin.Context) {
	app, err := h.d.Svc.GetAppByClientID(c.Request.Context(), c.Param("client_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":          app.ID,
		"name":        app.Name,
		"description": app.Description,
		"icon_url":    app.IconURL,
		"scopes":      app.Scopes,
	})
}
