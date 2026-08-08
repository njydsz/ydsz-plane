// Package httpapi — OIDC / SSO 认证 HTTP 端点。
//
// 设计参考:
//   - Microsoft identity platform OAuth 2.0 authorization code flow
//   - Google / Okta / Auth0 SPA OIDC 集成最佳实践
//   - PKCE 强制: 所有 SPA 登录请求必须带 code_challenge (S256)
//
// 路由 (注册于 /api/v1/auth/oidc):
//   GET  /providers            → 列出当前工作空间的 SSO Providers
//   POST /:provider_id/login   → 发起登录 → 返回 IdP 重定向 URL
//   GET  /callback             → OIDC 回调 → 重定向到前端 SSO 页（token 在 fragment 中）
package httpapi

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ssoProviderItem 暴露给前端的 SSO Provider 概要（不含 secret）。
type ssoProviderItem struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	IssuerURL      string `json:"issuer_url"`
	ClientID       string `json:"client_id"`
	AuthURL        string `json:"auth_url"`
	Scopes         string `json:"scopes"`
	AutoCreateUser bool   `json:"auto_create_user"`
}

// listSSOProviders 列出当前工作空间启用的 SSO Providers（不含 secret）。
func listSSOProviders(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID := c.GetInt64(middleware.CtxWorkspaceID)
		if wsID == 0 || d.OIDCService == nil {
			c.JSON(200, gin.H{"items": []ssoProviderItem{}})
			return
		}

		providers, err := d.OIDCService.ListProviders(c.Request.Context(), wsID)
		if err != nil {
			writeError(c, err)
			return
		}

		items := make([]ssoProviderItem, 0, len(providers))
		for _, p := range providers {
			items = append(items, ssoProviderItem{
				ID:             p.ID,
				Name:           p.Name,
				IssuerURL:      p.IssuerURL,
				ClientID:       p.ClientID,
				AuthURL:        p.AuthURL,
				Scopes:         joinScopes(p.Scopes),
				AutoCreateUser: p.AutoCreateUser,
			})
		}
		c.JSON(200, gin.H{"items": items})
	}
}

// ssoLoginRequest SSO 登录请求。
type ssoLoginRequest struct {
	RedirectTo string `json:"redirect_to"`
}

// initiateSSOLogin 发起 OIDC 登录 → 返回 IdP 重定向 URL。
func initiateSSOLogin(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d.OIDCService == nil {
			middleware.AbortWithError(c, errs.New("SSO.NOT_CONFIGURED", "SSO 未配置", 503))
			return
		}

		providerID, err := strconv.ParseInt(c.Param("provider_id"), 10, 64)
		if err != nil || providerID <= 0 {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field:  "provider_id", Reason: "无效的 Provider ID",
			}))
			return
		}

		var req ssoLoginRequest
		_ = c.ShouldBindJSON(&req)

		result, err := d.OIDCService.InitiateLogin(
			c.Request.Context(), providerID, req.RedirectTo, c.ClientIP(), c.Request.UserAgent())
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(200, result)
	}
}

// handleSSOCallback OIDC 回调处理 → 重定向到前端 SSO 回调页。
//
// 注意: 使用 #fragment 传递 token，避免 token 泄漏到 Access Log / Referer。
func handleSSOCallback(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d.OIDCService == nil {
			middleware.AbortWithError(c, errs.New("SSO.NOT_CONFIGURED", "SSO 未配置", 503))
			return
		}

		state := c.Query("state")
		code := c.Query("code")

		if state == "" || code == "" {
			c.Redirect(302, "/login?error=sso_missing_params")
			return
		}

		pair, err := d.OIDCService.HandleCallback(c.Request.Context(), auth.OIDCCallbackInput{
			State: state,
			Code:  code,
		})
		if err != nil {
			c.Redirect(302, "/login?error=sso_callback_failed")
			return
		}

		// 重定向到前端 SSO 回调页，token 使用 URL fragment 传递
		frontendURL := d.Cfg.Email.AppBaseURL
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}
		redirectURL := frontendURL + "/sso/callback#access=" + pair.AccessToken + "&refresh=" + pair.RefreshToken
		c.Redirect(302, redirectURL)
	}
}

// joinScopes 将 scope 列表拼接为空格分隔的字符串。
func joinScopes(scopes []string) string {
	result := ""
	for i, s := range scopes {
		if i > 0 {
			result += " "
		}
		result += s
	}
	return result
}
