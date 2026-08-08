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
	"fmt"
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

// HandleSAMLACS 处理 SAML Assertion Consumer Service 回调：
// 解码 Response → 校验状态 → 关联 Provider → 校验签名开关 → 提取身份 →
// 查找/创建用户 → 签发令牌对。返回的 *TokenPair 由 HTTP 层设置认证 Cookie 并重定向前端。
//
// 这是 SAML 登录链路真正闭环的关键：此前实现仅解析断言后重定向到占位提示页，
// 用户永远不会被真正登录。
func (s *OIDCService) HandleSAMLACS(ctx context.Context, samlResponse, relayState string) (*TokenPair, error) {
	// 1. Base64 解码
	raw, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, errs.New("SSO.SAML_INVALID_RESPONSE", "SAML Response 不是合法的 Base64 编码", 400)
	}

	// 2. 解析 SAML Response XML
	resp, err := parseSAMLResponse(raw)
	if err != nil {
		return nil, errs.New("SSO.SAML_INVALID_RESPONSE", "SAML Response XML 解析失败: "+err.Error(), 400)
	}

	// 3. 校验状态 — 成功才继续
	if !strings.Contains(resp.Status.StatusCode.Value, "Success") {
		return nil, errs.New("SSO.SAML_AUTH_FAILED", "SAML 认证状态非 Success", 401)
	}

	// 4. 通过 RelayState 关联发起时存储的 SSO 会话，取得 ProviderID
	providerID, err := s.resolveSAMLProvider(ctx, relayState)
	if err != nil {
		return nil, err
	}

	// 5. 加载 Provider 配置（含签名校验开关与 IdP 证书）
	provider, err := s.loadSAMLProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}

	// 6. 签名校验（无可用校验库时，要求显式 SkipSignature=true 方可放行）
	if err := s.validateSAMLResponse(raw, provider); err != nil {
		return nil, err
	}

	// 7. 提取身份标识与属性
	nameID := strings.TrimSpace(resp.Assertion.NameID.Value)
	if nameID == "" {
		return nil, errs.New("SSO.SAML_NO_NAMEID", "SAML Response 缺少 NameID", 400)
	}
	email, displayName, avatar := extractSAMLAttributes(resp.Assertion.Attrs, nameID)

	// 8. 查找或创建用户
	user, err := s.samlFindOrCreateUser(ctx, provider.ID, nameID, email, displayName, avatar,
		provider.AutoCreateUser, provider.DefaultRole)
	if err != nil {
		return nil, err
	}

	// 9. 签发令牌对
	pair, err := s.authSvc.issuePair(user.ID, user.Email, user.DisplayName, user.AvatarURL)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 10. 标记会话完成 + 维护 SSO 绑定
	_, _ = s.db.Exec(ctx,
		`UPDATE sso_sessions SET status = 'completed', user_id = $1, completed_at = now() WHERE state = $2`,
		user.ID, relayState)
	_, _ = s.db.Exec(ctx, `
		INSERT INTO sso_links (user_id, provider_id, sso_subject, sso_email, sso_display_name, last_login_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (provider_id, sso_subject) DO UPDATE SET
			sso_email = EXCLUDED.sso_email,
			sso_display_name = EXCLUDED.sso_display_name,
			last_login_at = now(),
			updated_at = now()`,
		user.ID, provider.ID, nameID, email, displayName)

	return pair, nil
}

// validateSAMLResponse 校验 SAML 断言签名。
//
// SAML 断言签名验证需要 xml-sec 库（如 crewjam/saml / russellhaering/gosaml2）执行
// 规范的 exc-c14n + RSA 验签。Go 标准库无法保留 XML 前缀，无法正确实现该算法，
// 因此本构建不内置验签。为保障安全，默认要求运维在受信环境**显式**设置 skip_signature=true
// 以放行；否则视为配置错误并拒绝，避免静默接受伪造断言。
func (s *OIDCService) validateSAMLResponse(_ []byte, provider *SAMLProviderConfig) error {
	if provider.SkipSignature {
		// 运维在受信网络/测试环境显式关闭校验
		return nil
	}
	return errs.New("SSO.SAML_SIGNATURE_REQUIRED",
		"SAML 签名校验未关闭，但当前构建未集成 xml-sec 校验库；请在受信环境设置 skip_signature=true，或集成 SAML 校验库后再启用严格校验", 500)
}

// resolveSAMLProvider 通过 RelayState 关联发起 SSO 时保存的会话，定位 Provider。
func (s *OIDCService) resolveSAMLProvider(ctx context.Context, relayState string) (int64, error) {
	if relayState == "" {
		return 0, errs.New("SSO.SAML_NO_RELAYSTATE", "SAML 响应缺少 RelayState，无法定位 Provider", 400)
	}
	var providerID int64
	err := s.db.QueryRow(ctx,
		`SELECT provider_id FROM sso_sessions WHERE state = $1 AND status = 'pending' AND expires_at > now()`,
		relayState).Scan(&providerID)
	if err != nil {
		return 0, errs.New("SSO.SAML_SESSION_NOT_FOUND", "SAML 会话不存在或已过期，请重新发起登录", 401)
	}
	return providerID, nil
}

// extractSAMLAttributes 从 SAML AttributeStatement 中提取 email / displayName / avatar，
// 兼容 OID/claims/常见简写等多种属性命名约定。
func extractSAMLAttributes(attrs []samlAttr, nameID string) (email, displayName, avatar string) {
	emailCandidates := []string{
		"urn:oid:0.9.2342.19200300.100.1.3",
		"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
		"email", "mail", "emailaddress", "user.email",
	}
	nameCandidates := []string{
		"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
		"displayname", "name", "cn", "urn:oid:2.5.4.3", "givenname", "username",
	}
	avatarCandidates := []string{"avatar", "picture", "photo", "urn:oid:0.9.2342.19200300.100.1.7"}

	attrMap := map[string]string{}
	for _, a := range attrs {
		if len(a.Values) > 0 && strings.TrimSpace(a.Values[0]) != "" {
			attrMap[strings.ToLower(a.Name)] = a.Values[0]
		}
	}
	for _, c := range emailCandidates {
		if v, ok := attrMap[strings.ToLower(c)]; ok {
			email = v
			break
		}
	}
	for _, c := range nameCandidates {
		if v, ok := attrMap[strings.ToLower(c)]; ok {
			displayName = v
			break
		}
	}
	for _, c := range avatarCandidates {
		if v, ok := attrMap[strings.ToLower(c)]; ok {
			avatar = v
			break
		}
	}
	if displayName == "" {
		displayName = nameID
	}
	return email, displayName, avatar
}

// samlFindOrCreateUser 按 SAML Provider + NameID 查找用户，不存在则按 email 回退，
// 仍不存在且开启自动创建时新建用户（与 OIDC 的 findOrCreateUser 行为一致）。
func (s *OIDCService) samlFindOrCreateUser(ctx context.Context, providerID int64, subject, email, displayName, avatar string, autoCreate bool, defaultRole string) (*UserBrief, error) {
	if defaultRole == "" {
		defaultRole = "member"
	}
	providerKey := fmt.Sprintf("saml:%d", providerID)

	var user UserBrief
	err := s.db.QueryRow(ctx, `
		SELECT id, email, display_name, coalesce(avatar_url, '')
		FROM users WHERE sso_provider = $1 AND sso_subject = $2 AND deleted_at IS NULL`,
		providerKey, subject).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL)
	if err == nil {
		return &user, nil
	}

	// 按邮箱回退查找（邮箱可能已通过其他方式注册）
	if email != "" {
		err = s.db.QueryRow(ctx, `
			SELECT id, email, display_name, coalesce(avatar_url, '')
			FROM users WHERE email = $1 AND deleted_at IS NULL`, email).
			Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL)
		if err == nil {
			_, _ = s.db.Exec(ctx,
				`UPDATE users SET sso_provider = $1, sso_subject = $2 WHERE id = $3`,
				providerKey, subject, user.ID)
			return &user, nil
		}
	}

	if !autoCreate {
		return nil, errs.New("SSO.USER_NOT_FOUND", "该 SAML 账号未关联到平台用户，请联系管理员", 403)
	}
	if email == "" {
		email = subject // 至少保证唯一可联系
	}
	err = s.db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name, avatar_url, sso_provider, sso_subject, is_active)
		VALUES ($1, NULL, $2, $3, $4, $5, TRUE)
		RETURNING id, email, display_name, coalesce(avatar_url, '')`,
		email, displayName, avatar, providerKey, subject).
		Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &user, nil
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
		       '', '', sso_url, idp_issuer, idp_certificate, skip_signature,
		       COALESCE(auto_create_user, false), COALESCE(default_role, 'member')
		FROM sso_providers WHERE id = $1 AND protocol = 'saml' AND enabled = TRUE`,
		providerID).Scan(&cfg.ID, &cfg.Name, &cfg.IssuerURL, &cfg.ClientID,
		&cfg.ClientSecret, &cfg.RedirectURI, &cfg.AuthURL, &cfg.TokenURL,
		&cfg.JWKSURL, &cfg.Scopes, &protocol, &cfg.AttributeMapping,
		&ssoURL, &idpIssuer, &idpCert, &skipSig, &cfg.AutoCreateUser, &cfg.DefaultRole)
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
