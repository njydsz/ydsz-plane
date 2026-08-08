// Package httpapi — SSO Provider 管理 HTTP 端点（workspace 管理员配置 SSO 集成）。
//
// 设计参考:
//   - GitHub Organization SAML SSO settings
//   - GitLab Admin Area → OmniAuth / SAML settings
//   - Azure AD Enterprise Applications
//
// 路由 (注册于 /api/v1/workspaces/:wid/sso/providers):
//
//\tGET    /                    → 列出工作空间 SSO Providers
//\tPOST   /                    → 创建 SSO Provider（需 workspace:update 权限）
//\tGET    /:id                 → 获取单个 Provider 详情
//\tPATCH  /:id                 → 更新 Provider 配置
//\tDELETE /:id                 → 删除 Provider
package httpapi

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ssoProviderListItem 列表展示的 Provider 概要（不含 secret）。
type ssoProviderListItem struct {
	ID             int64             `json:"id"`
	Name           string            `json:"name"`
	Protocol       string            `json:"protocol"`
	IssuerURL      string            `json:"issuer_url"`
	ClientID       string            `json:"client_id"`
	AuthURL        string            `json:"auth_url"`
	Scopes         string            `json:"scopes"`
	AutoCreateUser bool              `json:"auto_createUser"`
	DefaultRole    string            `json:"default_role"`
	Enabled        bool              `json:"enabled"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
}

// listSSOProvidersMgmt 列出当前工作空间的 SSO Providers（含 secret 占位提示）。
func listSSOProvidersMgmt(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d.OIDCService == nil {
			c.JSON(200, gin.H{"items": []ssoProviderListItem{}})
			return
		}
		providers, err := d.OIDCService.ListProviders(c.Request.Context(), mustWorkspaceID(c))
		if err != nil {
			writeError(c, err)
			return
		}
		items := make([]ssoProviderListItem, 0, len(providers))
		for _, p := range providers {
			items = append(items, ssoProviderListItem{
				ID: p.ID, Name: p.Name, Protocol: "oidc",
				IssuerURL: p.IssuerURL, ClientID: p.ClientID, AuthURL: p.AuthURL,
				Scopes: joinScopes(p.Scopes), AutoCreateUser: p.AutoCreateUser,
				DefaultRole: p.DefaultRole, Enabled: true,
			})
		}
		c.JSON(200, gin.H{"items": items})
	}
}

// createSSOProvider 创建 SSO Provider。
func createSSOProvider(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d.OIDCService == nil {
			middleware.AbortWithError(c, errs.New("SSO.NOT_CONFIGURED", "SSO 未配置", 503))
			return
		}
		var in auth.ProviderModifyInput
		if err := c.ShouldBindJSON(&in); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field:  "body", Reason: "请求体格式错误",
			}))
			return
		}
		in.WorkspaceID = mustWorkspaceID(c)
		created, err := d.OIDCService.CreateProvider(c.Request.Context(), &in)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(201, toSSOProviderDetail(created))
	}
}

// getSSOProvider 获取单个 SSO Provider 详情。
func getSSOProvider(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d.OIDCService == nil {
			middleware.AbortWithError(c, errs.New("SSO.NOT_CONFIGURED", "SSO 未配置", 503))
			return
		}
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		if id <= 0 {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field:  "id", Reason: "无效的 Provider ID",
			}))
			return
		}
		cfg, err := d.OIDCService.GetProvider(c.Request.Context(), id)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(200, toSSOProviderDetail(cfg))
	}
}

// updateSSOProvider 更新 SSO Provider。
func updateSSOProvider(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d.OIDCService == nil {
			middleware.AbortWithError(c, errs.New("SSO.NOT_CONFIGURED", "SSO 未配置", 503))
			return
		}
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		if id <= 0 {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field:  "id", Reason: "无效的 Provider ID",
			}))
			return
		}
		var in auth.ProviderModifyInput
		if err := c.ShouldBindJSON(&in); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field:  "body", Reason: "请求体格式错误",
			}))
			return
		}
		updated, err := d.OIDCService.UpdateProvider(c.Request.Context(), id, &in)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(200, toSSOProviderDetail(updated))
	}
}

// deleteSSOProvider 删除 SSO Provider。
func deleteSSOProvider(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d.OIDCService == nil {
			middleware.AbortWithError(c, errs.New("SSO.NOT_CONFIGURED", "SSO 未配置", 503))
			return
		}
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		if id <= 0 {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field:  "id", Reason: "无效的 Provider ID",
			}))
			return
		}
		if err := d.OIDCService.DeleteProvider(c.Request.Context(), id); err != nil {
			writeError(c, err)
			return
		}
		c.Status(204)
	}
}

// --- helpers ---

// ssoProviderDetail 返回给前端的 Provider 详情（secret 字段标注是否已设置，但不返回明文）。
type ssoProviderDetail struct {
	ID               int64             `json:"id"`
	Name             string            `json:"name"`
	Protocol         string            `json:"protocol"`
	IssuerURL        string            `json:"issuer_url"`
	ClientID         string            `json:"client_id"`
	HasClientSecret  bool              `json:"has_client_secret"`
	RedirectURI      string            `json:"redirect_uri"`
	AuthURL          string            `json:"auth_url"`
	TokenURL         string            `json:"token_url"`
	UserInfoURL      string            `json:"userinfo_url"`
	JWKSURL          string            `json:"jwks_url"`
	Scopes           string            `json:"scopes"`
	AutoCreateUser   bool              `json:"auto_create_user"`
	DefaultRole      string            `json:"default_role"`
	AttributeMapping map[string]string `json:"attribute_mapping"`
}

func toSSOProviderDetail(cfg *auth.OIDCProviderConfig) *ssoProviderDetail {
	return &ssoProviderDetail{
		ID: cfg.ID, Name: cfg.Name, IssuerURL: cfg.IssuerURL, ClientID: cfg.ClientID,
		HasClientSecret: cfg.ClientSecret != "", RedirectURI: cfg.RedirectURI,
		AuthURL: cfg.AuthURL, TokenURL: cfg.TokenURL, UserInfoURL: cfg.UserInfoURL,
		JWKSURL: cfg.JWKSURL, Scopes: joinScopes(cfg.Scopes),
		AutoCreateUser: cfg.AutoCreateUser, DefaultRole: cfg.DefaultRole,
		AttributeMapping: cfg.AttributeMapping,
	}
}

// mustWorkspaceID 从路由参数提取 workspace ID。
func mustWorkspaceID(c *gin.Context) int64 {
	id, _ := strconv.ParseInt(c.Param("wid"), 10, 64)
	if id <= 0 {
		id = c.GetInt64(middleware.CtxWorkspaceID)
	}
	return id
}
