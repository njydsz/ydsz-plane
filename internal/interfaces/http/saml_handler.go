// Package httpapi — SAML 2.0 ACS (Assertion Consumer Service) HTTP Handler。
//
// 路由:
//   POST /api/v1/auth/saml/acs — IdP POST-back 的 Assertion Consumer Service
//
// 注意: 生产环境需引入 xml-sec 库 (gosaml2/go-saml) 验证 SAML Response 签名。
package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// handleSAMLACS 处理 IdP POST-back 的 SAML Response（SAML HTTP-POST Binding）。
//
// IdP 完成认证后，将 Base64 编码的 SAML Response 通过 form POST 发送到此端点。
// 后端验证 Response → 创建/查找用户 → 签发 Token → 重定向到前端。
func handleSAMLACS(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if d.OIDCService == nil {
			middleware.AbortWithError(c, errs.New("SSO.NOT_CONFIGURED", "SSO 未配置", 503))
			return
		}

		// 从 POST body 读取 SAML Response
		samlResponse := c.PostForm("SAMLResponse")
		if samlResponse == "" {
			c.Redirect(302, "/login?error=saml_missing_response")
			return
		}

		if err := d.OIDCService.HandleSAMLACS(c.Request.Context(), samlResponse); err != nil {
			c.Redirect(302, "/login?error=saml_auth_failed")
			return
		}

		// TODO: 实际生产流程应建立会话后重定向到前端
		//       当前骨架实现仅完成 SAML Response 解析桩
		c.Redirect(302, "/login?info=saml_not_fully_implemented")
	}
}

// handleSAMLMetadata 暴露 SP 元数据（供 IdP 管理员导入）。
//
// GET /api/v1/auth/saml/metadata → SP 实体描述 XML（entity descriptor）。
func handleSAMLMetadata(d *Deps) gin.HandlerFunc {
	_ = d
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/xml")
		c.String(http.StatusOK, `<?xml version="1.0"?>
<!-- S13: ydsz-plane SAML SP Metadata (stub) -->
<!-- 生产环境应由 go-saml 库根据 SP 配置动态生成 -->
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
  entityID="ydsz-plane">
  <SPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <AssertionConsumerService
      Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
      Location="https://ydsz-plane.example.com/api/v1/auth/saml/acs"
      index="0" isDefault="true"/>
  </SPSSODescriptor>
</EntityDescriptor>`)
	}
}
