// Package httpapi — SAML 2.0 ACS (Assertion Consumer Service) HTTP Handler。
//
// 路由:
//   POST /api/v1/auth/saml/acs — IdP POST-back 的 Assertion Consumer Service
//   GET  /api/v1/auth/saml/metadata — 暴露 SP 元数据（供 IdP 管理员导入）
package httpapi

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// handleSAMLACS 处理 IdP POST-back 的 SAML Response（SAML HTTP-POST Binding）。
//
// 流程: 读取 SAMLResponse/RelayState → 应用层解析断言 → 查找/创建用户 → 签发令牌 →
// 设置 HttpOnly Cookie → 重定向前端 /sso/callback。与 OIDC 回调行为完全一致，
// 使 SAML 登录真正闭环（此前仅重定向到占位提示页）。
func handleSAMLACS(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d.OIDCService == nil {
			middleware.AbortWithError(c, errs.New("SSO.NOT_CONFIGURED", "SSO 未配置", 503))
			return
		}

		samlResponse := c.PostForm("SAMLResponse")
		relayState := c.PostForm("RelayState")
		if samlResponse == "" {
			c.Redirect(302, "/login?error=saml_missing_response")
			return
		}

		pair, err := d.OIDCService.HandleSAMLACS(c.Request.Context(), samlResponse, relayState)
		if err != nil {
			c.Redirect(302, "/login?error=saml_auth_failed")
			return
		}

		// 设置认证 Cookie（与密码/OIDC 登录一致：HttpOnly + Secure + SameSite=Lax）
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
		middleware.SetCSRFTokenCookie(c, d.Cfg, d.Cfg.Auth.JWTSecret, pair.AccessToken)

		// 重定向前端 SSO 回调页，前端仅验证 Cookie 是否生效
		frontendURL := d.Cfg.Email.AppBaseURL
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}
		c.Redirect(302, frontendURL+"/sso/callback")
	}
}

// handleSAMLMetadata 暴露 SP 元数据（供 IdP 管理员导入）。
//
// GET /api/v1/auth/saml/metadata → 根据当前部署地址动态生成的 SP 实体描述 XML。
// 早期实现为硬编码桩（固定 entityID 与 example.com 地址），此处改为基于配置的公网
// 地址（Cfg.AppBaseURL）动态生成，确保 IdP 侧导入的 ACS 端点与实际部署一致。
func handleSAMLMetadata(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		base := d.Cfg.Email.AppBaseURL
		if base == "" {
			scheme := "https"
			if c.Request.TLS == nil {
				scheme = "http"
			}
			base = scheme + "://" + c.Request.Host
		}
		base = strings.TrimRight(base, "/")
		acsURL := base + "/api/v1/auth/saml/acs"
		entityID := base

		c.Header("Content-Type", "application/xml")
		c.String(http.StatusOK, samlMetadataXML(entityID, acsURL))
	}
}

// samlMetadataXML 生成符合 SAML 2.0 元数据的 SP EntityDescriptor（XML 已转义）。
func samlMetadataXML(entityID, acsURL string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(entityID))
	escEntity := b.String()
	b.Reset()
	_ = xml.EscapeText(&b, []byte(acsURL))
	escACS := b.String()

	return `<?xml version="1.0" encoding="UTF-8"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata"
  entityID="` + escEntity + `">
  <md:SPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol" AuthnRequestsSigned="false" WantAssertionsSigned="true">
    <md:NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:emailAddress</md:NameIDFormat>
    <md:AssertionConsumerService
      Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
      Location="` + escACS + `"
      index="0" isDefault="true"/>
  </md:SPSSODescriptor>
</md:EntityDescriptor>`
}
