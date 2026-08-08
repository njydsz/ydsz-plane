// Package auth — SAML 2.0 企业认证集成（骨架实现）。
//
// 设计参考:
//   - SAML 2.0 Core (OASIS Standard)
//   - ADFS / Azure AD / Okta / OneLogin SAML SSO 流程
//   - OWASP SAML 安全指南 (Cheat Sheet)
//
// 流程:
//   GET /api/v1/auth/sso/:workspace_id/providers/:pid/login (protocol=saml)
//     → 生成 SAML AuthnRequest XML → Base64+URL 编码 → 302 重定向到 IdP SSO URL
//
//   POST /api/v1/auth/saml/acs (IdP Assertion Consumer Service)
//     → 解码 SAML Response XML → 验证签名 (xmlsec) → 提取 NameID/Attributes
//     → 创建会话 → 重定向到前端
//
// 注意: 此为 MVP 骨架。完整的 SAML Response 签名验证依赖 xml-sec 库
//   (如 /atro32/go-saml 或 /russellhaering/gosaml2)，生产部署时引入。
//   当前 validateSAMLResponse 为桩实现，仅解析非签名 NameID 与 Attribute。
package auth

import (
	"bytes"
	"compress/flate"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// SAMLProviderConfig SAML Provider 配置（复用 OIDCProviderConfig + SAML 专有字段）。
type SAMLProviderConfig struct {
	OIDCProviderConfig
	// SAML 专有：IdP 元数据 XML 或单点登录 URL
	SSOURL          string `json:"sso_url"`
	IDPIssuer       string `json:"idp_issuer"`
	IDPCertificate  string `json:"idp_certificate"`
	SkipSignature   bool   `json:"skip_signature"` // 生产必须为 false
	WantAuthnReqSig bool   `json:"want_authn_req_sig"`
}

// SAMLInitiateResult SAML 登录发起结果。
type SAMLInitiateResult struct {
	RedirectURL string `json:"redirect_url"`
}

// samlAuthnRequest SAML AuthnRequest XML 顶层结构（简化）。
type samlAuthnRequest struct {
	XMLName                        xml.Name  `xml:"urn:oasis:names:tc:SAML:2.0:protocol AuthnRequest"`
	ID                             string    `xml:"ID,attr"`
	Version                        string    `xml:"Version,attr"`
	IssueInstant                   string    `xml:"IssueInstant,attr"`
	Destination                    string    `xml:"Destination,omitempty"`
	AssertionConsumerServiceURL    string    `xml:"AssertionConsumerServiceURL,attr"`
	Issuer                         samlIssuer `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
}

type samlIssuer struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
	Value   string   `xml:",chardata"`
}

// samlResponseWrapper SAML Response XML 简化解析结构。
type samlResponseWrapper struct {
	XMLName   xml.Name        `xml:"urn:oasis:names:tc:SAML:2.0:protocol Response"`
	Assertion samlAssertion   `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
	Status    samlStatusField `xml:"urn:oasis:names:tc:SAML:2.0:protocol Status"`
}

type samlAssertion struct {
	Subject samlSubject  `xml:"urn:oasis:names:tc:SAML:2.0:assertion Subject"`
	NameID  samlNameID   `xml:"urn:oasis:names:tc:SAML:2.0:assertion Subject>NameID"`
	Attrs   []samlAttr   `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeStatement>Attribute"`
}

type samlSubject struct {
	NameID samlNameID `xml:"urn:oasis:names:tc:SAML:2.0:assertion NameID"`
}

type samlNameID struct {
	Value string `xml:",chardata"`
	Format string `xml:"Format,attr"`
}

type samlAttr struct {
	Name   string   `xml:"Name,attr"`
	Values []string `xml:"AttributeValue"`
}

type samlStatusField struct {
	StatusCode samlStatusCode `xml:"urn:oasis:names:tc:SAML:2.0:protocol StatusCode"`
}

type samlStatusCode struct {
	Value string `xml:"Value,attr"`
}

// SAMLInitiateLogin 生成 SAML AuthnRequest → 重定向到 IdP SSO 端点。
func (s *OIDCService) SAMLInitiateLogin(ctx context.Context, providerID int64, redirectTo string) (*SAMLInitiateResult, error) {
	provider, err := s.loadSAMLProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}
	if provider == nil || provider.SSOURL == "" {
		return nil, errs.New("SSO.SAML_NOT_CONFIGURED", "SAML Provider 的 SSO URL 未配置", 400)
	}

	// 生成 state、SAML ID
	state, err := generateRandomString(32)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	samlID, err := generateSAMLID()
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 生成 ACS URL
	acsURL := s.samlACSURL(provider.ID)

	// 存储 SSO session
	_, err = s.db.Exec(ctx, `
		INSERT INTO sso_sessions (state, nonce, provider_id, redirect_to, status, expires_at)
		VALUES ($1, $2, $3, $4, 'pending', now() + interval '10 minutes')`,
		state, samlID, providerID, redirectTo)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 生成 SAML AuthnRequest XML
	reqXML, err := buildSAMLAuthnRequest(samlID, acsURL)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// Deflate + Base64 + URL 编码
	encoded := encodeSAMLRequest(reqXML)

	// 构造 IdP 重定向 URL（SAML GET binding）
	redirectURL := buildIdPRedirectURL(provider.SSOURL, state, encoded)

	return &SAMLInitiateResult{RedirectURL: redirectURL}, nil
}

// HandleSAMLACS 处理 SAML Assertion Consumer Service 回调。
func (s *OIDCService) HandleSAMLACS(ctx context.Context, samlResponse string) error {
	// Base64 解码
	raw, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return errs.New("SSO.SAML_INVALID_RESPONSE", "SAML Response 解析失败", 400)
	}

	// 解析 SAML Response XML
	resp, err := parseSAMLResponse(raw)
	if err != nil {
		return errs.New("SSO.SAML_INVALID_RESPONSE", "SAML Response XML 解析失败: "+err.Error(), 400)
	}

	// 验证状态 — 成功才继续
	if !strings.Contains(resp.Status.StatusCode.Value, "Success") {
		return errs.New("SSO.SAML_AUTH_FAILED", "SAML 认证失败", 401)
	}

	// 提取用户身份标识
	nameID := strings.TrimSpace(resp.Assertion.NameID.Value)
	if nameID == "" {
		return errs.New("SSO.SAML_NO_NAMEID", "SAML Response 缺少 NameID", 400)
	}

	// 提取属性（email, display name 等）
	attributes := map[string]string{}
	for _, attr := range resp.Assertion.Attrs {
		if len(attr.Values) > 0 {
			attributes[attr.Name] = attr.Values[0]
		}
	}

	// 存 session（关联 Provider 与待完成用户）
	state, _ := generateRandomString(32)
	_, _ = s.db.Exec(ctx, `
		INSERT INTO sso_sessions (state, nonce, provider_id, redirect_to, status, expires_at)
		VALUES ($1, $2, $3, '', 'pending', now() + interval '5 minutes')`,
		state, nameID, 0)

	// NOTE: 实际生产流程应通过 state 匹配原始 session → ProviderID
	//       → Create/Find user → Issue JWT token
	//       此处为骨架实现，返回 nil 表示框架已就位。
	return nil
}

// validateSAMLResponse 桩实现 — 生产环境引入 xml-security 库进行签名验证。
//
// 推荐实现: github.com/russellhaering/gosaml2 / github.com/atro32/go-saml
// 需: 加载 IDP X.509 证书 → 验证 <ds:Signature> → 检查 Conditions/NotOnOrAfter
func validateSAMLResponse(_ []byte, _ *SAMLProviderConfig) error {
	// TODO: 替换为真正的 xml-sec 校验
	return nil
}

// --- private helpers ---

func (s *OIDCService) loadSAMLProvider(ctx context.Context, providerID int64) (*SAMLProviderConfig, error) {
	var cfg OIDCProviderConfig
	var protocol, ssoURL, idpIssuer, idpCert string
	var skipSig bool
	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(id, 0), COALESCE(name, ''), COALESCE(issuer_url, ''),
		       COALESCE(client_id, ''), COALESCE(client_secret, ''),
		       COALESCE(redirect_uri, ''), COALESCE(auth_url, ''),
		       COALESCE(token_url, ''), COALESCE(jwks_url, ''),
		       COALESCE(scopes, ''), COALESCE(protocol, ''),
		       COALESCE(attribute_mapping::text, '{}'),
		       '', '', sso_url, idp_issuer, idp_certificate, skip_signature
		FROM sso_providers WHERE id = $1 AND protocol = 'saml' AND enabled = TRUE`,
		providerID).Scan(&cfg.ID, &cfg.Name, &cfg.IssuerURL, &cfg.ClientID,
		&cfg.ClientSecret, &cfg.RedirectURI, &cfg.AuthURL, &cfg.TokenURL,
		&cfg.JWKSURL, &cfg.Scopes, &protocol, &cfg.AttributeMapping,
		&ssoURL, &idpIssuer, &idpCert, &skipSig)
	if err != nil {
		return nil, errs.New("SSO.SAML_NOT_FOUND", "SAML Provider 未找到或未启用", 404)
	}
	return &SAMLProviderConfig{
		OIDCProviderConfig: cfg,
		SSOURL:             ssoURL,
		IDPIssuer:          idpIssuer,
		IDPCertificate:     idpCert,
		SkipSignature:      skipSig,
	}, nil
}

func (s *OIDCService) samlACSURL(_ int64) string {
	return s.appBaseURL + "/api/v1/auth/saml/acs"
}

// generateSAMLID 生成 SAML 唯一 ID（格式: _<hex32>）。
func generateSAMLID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "_" + hex.EncodeToString(bytes), nil
}

// buildSAMLAuthnRequest 构造 SAML AuthnRequest XML。
func buildSAMLAuthnRequest(id, acsURL string) (string, error) {
	req := samlAuthnRequest{
		ID:                          id,
		Version:                     "2.0",
		IssueInstant:                time.Now().UTC().Format(time.RFC3339),
		AssertionConsumerServiceURL: acsURL,
		Issuer:                      samlIssuer{Value: "ydsz-plane"},
	}

	output, err := xml.MarshalIndent(req, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(output), nil
}

// encodeSAMLRequest Deflate + Base64 + URL 编码 SAML Request。
func encodeSAMLRequest(xmlStr string) string {
	var buf bytes.Buffer
	w, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		// 失败时回退为无压缩 base64
		return base64.StdEncoding.EncodeToString([]byte(xmlStr))
	}
	_, _ = w.Write([]byte(xmlStr))
	_ = w.Close()
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// buildIdPRedirectURL 构造 IdP SSO 重定向 URL（SAML HTTP-Redirect Binding）。
func buildIdPRedirectURL(ssoURL, state, samlRequest string) string {
	separator := "?"
	if strings.Contains(ssoURL, "?") {
		separator = "&"
	}
	return ssoURL + separator + "SAMLRequest=" + url.QueryEscape(samlRequest) +
		"&RelayState=" + url.QueryEscape(state)
}

// parseSAMLResponse 解析 SAML Response XML。
func parseSAMLResponse(data []byte) (*samlResponseWrapper, error) {
	var resp samlResponseWrapper
	decoder := xml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&resp); err != nil && err != io.EOF {
		return nil, err
	}
	return &resp, nil
}
