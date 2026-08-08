// Package httpapi — OIDC / SSO 认证 HTTP 端点。
//
// 设计参考:
//   - Microsoft identity platform OAuth 2.0 authorization code flow
//   - Google / Okta / Auth0 SPA OIDC 集成最佳实践
//   - PKCE 强制: 所有 SPA 登录请求必须带 code_challenge (S256)
//
// 路由:
//   POST  /api/v1/auth/oidc/:provider_id/login   → 返回 IdP redirect URL（JSON）
//   GET   /api/v1/auth/sso/:wid/providers/:pid/login → 浏览器重定向到 IdP
//   GET   /api/v1/auth/oidc/callback              → IdP 回调 → Cookie → 重定向前端
package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// handleSSORedirect GET-based SSO 登录入口（浏览器直跳场景）。
// 前端 window.location.href → 此端点 → 302 重定向到 IdP 授权端点。
func handleSSORedirect(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d.OIDCService == nil {
			c.Redirect(302, "/login?error=sso_not_configured")
			return
		}
		providerID, err := strconv.ParseInt(c.Param("provider_id"), 10, 64)
		if err != nil || providerID <= 0 {
			c.Redirect(302, "/login?error=sso_invalid_provider")
			return
		}
		redirectTo := c.Query("redirect_to")
		result, err := d.OIDCService.InitiateLogin(
			c.Request.Context(), providerID, redirectTo, c.ClientIP(), c.Request.UserAgent())
		if err != nil {
			c.Redirect(302, "/login?error=sso_initiate_failed")
			return
		}
		c.Redirect(302, result.RedirectURL)
	}
}

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
				Field: "provider_id", Reason: "无效的 Provider ID",
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

// handleSSOCallback OIDC 回调处理。
//
// 验证 state/code → 签发 JWT → 设置 HTTP-only Cookie → 重定向到前端 SSO 页。
// 使用 HttpOnly Secure Cookie（与密码登录一致），避免 token 泄漏到日志/Referer。
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

		// 设置认证 Cookie（与密码登录一致：HttpOnly + Secure + SameSite=Lax）
		secure := !d.Cfg.IsDev()
		http.SetCookie(c.Writer, &http.Cookie{
			Name: "ydsz_access", Value: pair.AccessToken, Path: "/",
			HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
			MaxAge: int(d.Cfg.Auth.AccessTokenTTL.Seconds()),
		})
		http.SetCookie(c.Writer, &http.Cookie{
			Name: "ydsz_refresh", Value: pair.RefreshToken, Path: "/api/v1/auth",
			HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
			MaxAge: int(d.Cfg.Auth.RefreshTokenTTL.Seconds()),
		})
		// SSO 建立会话后同步设置 CSRF 双提交令牌 Cookie
		middleware.SetCSRFTokenCookie(c, d.Cfg, d.Cfg.Auth.JWTSecret, pair.AccessToken)

		// 重定向到前端 SSO 回调页，前端仅验证 Cookie 是否生效
		frontendURL := d.Cfg.Email.AppBaseURL
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}
		c.Redirect(302, frontendURL+"/sso/callback")
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
