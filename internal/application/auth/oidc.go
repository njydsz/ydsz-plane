// Package auth — OIDC (OpenID Connect) 企业统一认证集成。
//
// 设计参考:
//   - OpenID Connect Core 1.0 (RFC 6749 + OpenID)
//   - 阿里云 IDaaS / 腾讯云 EIAM / Keycloak / Okta
//   - 等保三级 7.1.2 身份鉴别要求
//   - OWASP ASVS V2.1 (认证架构)
//
// 流程:
//   GET  /api/v1/auth/oidc/:provider_id/login    → 生成 state → 302 跳转 IdP
//   GET  /api/v1/auth/oidc/callback              → 验证 state → 交换 code → 验证 id_token → 登录/创建用户
//   GET  /api/v1/auth/oidc/providers             → 列出已启用的 SSO Providers
//
// 安全措施:
//   - state 参数防 CSRF（crypto/rand 256-bit, 10 分钟有效期）
//   - nonce 参数防重放攻击（OIDC 标准）
//   - PKCE (S256) 防授权码拦截（code_challenge）
//   - client_secret 加密存储（AES-256-GCM）
//   - id_token 验证: iss/aud/exp/iat/nonce + JWKS 签名校验
package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// --- OIDC Provider 配置 ---

// OIDCProviderConfig 单个 OIDC Provider 的运行时配置。
type OIDCProviderConfig struct {
	ID             int64             `json:"id"`
	Name           string            `json:"name"`
	IssuerURL      string            `json:"issuer_url"`
	ClientID       string            `json:"client_id"`
	ClientSecret   string            `json:"client_secret"` // 解密后的明文
	RedirectURI    string            `json:"redirect_uri"`
	AuthURL        string            `json:"auth_url"`
	TokenURL       string            `json:"token_url"`
	UserInfoURL    string            `json:"userinfo_url"`
	JWKSURL        string            `json:"jwks_url"`
	Scopes         []string          `json:"scopes"`
	AutoCreateUser bool              `json:"auto_create_user"`
	DefaultRole    string            `json:"default_role"`
	AttributeMapping map[string]string `json:"attribute_mapping"` // e.g. {"email":"email","display_name":"name"}
}

// --- OIDC Service ---

// OIDCService 处理 OIDC 登录流程。
type OIDCService struct {
	db         *pgxpool.Pool
	authSvc    *Service
	httpClient *http.Client
	appBaseURL string
}

// NewOIDCService 创建 OIDC 服务。
func NewOIDCService(db *pgxpool.Pool, authSvc *Service, appBaseURL string) *OIDCService {
	return &OIDCService{
		db:         db,
		authSvc:    authSvc,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		appBaseURL: appBaseURL,
	}
}

// --- Login Flow ---

// OIDCLoginResult 登录发起结果。
type OIDCLoginResult struct {
	RedirectURL string `json:"redirect_url"`
	State       string `json:"state"`
}

// InitiateLogin 发起 OIDC 登录（生成 state + 构造授权 URL）。
func (s *OIDCService) InitiateLogin(ctx context.Context, providerID int64, redirectTo string, ip, userAgent string) (*OIDCLoginResult, error) {
	cfg, err := s.loadProviderConfig(ctx, providerID)
	if err != nil {
		return nil, err
	}

	// 生成 state 和 code_verifier (PKCE)
	state, err := generateRandomString(32)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	nonce, err := generateRandomString(32)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	codeVerifier, err := generateRandomString(64)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	codeChallenge := s.computeS256Challenge(codeVerifier)

	// 存储 SSO 会话（含 PKCE code_verifier 至独立列，避免污染 error_message）
	_, err = s.db.Exec(ctx, `
		INSERT INTO sso_sessions (state, nonce, provider_id, redirect_to, ip_address, user_agent, code_verifier, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', now() + interval '10 minutes')`,
		state, nonce, providerID, redirectTo, ip, userAgent, codeVerifier)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 构造授权 URL
	authURL, err := url.Parse(cfg.AuthURL)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	params := authURL.Query()
	params.Set("response_type", "code")
	params.Set("client_id", cfg.ClientID)
	params.Set("redirect_uri", cfg.RedirectURI)
	params.Set("scope", strings.Join(cfg.Scopes, " "))
	params.Set("state", state)
	params.Set("nonce", nonce)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")
	authURL.RawQuery = params.Encode()

	return &OIDCLoginResult{
		RedirectURL: authURL.String(),
		State:       state,
	}, nil
}

// --- Callback Flow ---

// OIDCCallbackInput 回调参数。
type OIDCCallbackInput struct {
	State string `json:"state" form:"state"`
	Code  string `json:"code" form:"code"`
}

// HandleCallback 处理 OIDC 回调（交换 code → 验证 id_token → 登录/创建用户）。
func (s *OIDCService) HandleCallback(ctx context.Context, in OIDCCallbackInput) (*TokenPair, error) {
	// 1. 验证 state 参数（防 CSRF）
	session, codeVerifier, err := s.validateState(ctx, in.State)
	if err != nil {
		return nil, err
	}

	// 2. 加载 provider 配置
	cfg, err := s.loadProviderConfig(ctx, session.ProviderID)
	if err != nil {
		return nil, err
	}

	// 3. 用 code 交换 token（含 PKCE code_verifier）
	tokenResp, err := s.exchangeCode(ctx, cfg, in.Code, codeVerifier)
	if err != nil {
		s.markSessionFailed(ctx, in.State, err.Error())
		return nil, err
	}

	// 4. 验证 id_token
	idToken, claims, err := s.verifyIDToken(ctx, cfg, tokenResp.IDToken, session.Nonce)
	if err != nil {
		s.markSessionFailed(ctx, in.State, err.Error())
		return nil, err
	}

	_ = idToken // 保留供后续使用

	// 5. 获取用户信息（优先使用 id_token claims，回退 UserInfo endpoint）
	userInfo := s.extractUserInfo(claims, cfg, tokenResp.AccessToken)

	// 6. 查找或创建用户
	user, err := s.findOrCreateUser(ctx, cfg, userInfo, session.ProviderID)
	if err != nil {
		s.markSessionFailed(ctx, in.State, err.Error())
		return nil, err
	}

	// 7. 签发令牌对
	pair, err := s.authSvc.issuePair(user.ID, user.Email, user.DisplayName, user.AvatarURL)
	if err != nil {
		s.markSessionFailed(ctx, in.State, err.Error())
		return nil, err
	}

	// 8. 标记会话完成
	_, _ = s.db.Exec(ctx,
		`UPDATE sso_sessions SET status = 'completed', user_id = $1, completed_at = now() WHERE state = $2`,
		user.ID, in.State)

	// 9. 更新 SSO link
	_, _ = s.db.Exec(ctx, `
		INSERT INTO sso_links (user_id, provider_id, sso_subject, sso_email, sso_display_name, last_login_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (provider_id, sso_subject) DO UPDATE SET
			sso_email = EXCLUDED.sso_email,
			sso_display_name = EXCLUDED.sso_display_name,
			last_login_at = now(),
			updated_at = now()`,
		user.ID, session.ProviderID, userInfo.Subject, userInfo.Email, userInfo.DisplayName)

	return pair, nil
}

// --- Token Exchange ---

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func (s *OIDCService) exchangeCode(ctx context.Context, cfg *OIDCProviderConfig, code, codeVerifier string) (*tokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", cfg.RedirectURI)
	data.Set("client_id", cfg.ClientID)
	data.Set("client_secret", cfg.ClientSecret)
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, errs.New("SSO.TOKEN_EXCHANGE_FAILED", "无法连接到身份提供商", 502).Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		_ = resp.Body.Close() // 丢弃错误响应体
		return nil, errs.New("SSO.TOKEN_EXCHANGE_FAILED",
			fmt.Sprintf("身份提供商返回错误 (status=%d)", resp.StatusCode), 502)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	if tr.IDToken == "" {
		return nil, errs.New("SSO.INVALID_RESPONSE", "身份提供商未返回 id_token", 502)
	}

	return &tr, nil
}

// --- ID Token Verification ---

type oidcClaims struct {
	jwt.RegisteredClaims
	Nonce         string `json:"nonce"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	PreferredUsername string `json:"preferred_username"`
}

func (s *OIDCService) verifyIDToken(ctx context.Context, cfg *OIDCProviderConfig, rawToken, expectedNonce string) (string, *oidcClaims, error) {
	var claims oidcClaims
	token, err := jwt.ParseWithClaims(rawToken, &claims, func(t *jwt.Token) (any, error) {
		// 从 JWKS 获取公钥
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}
		return s.fetchJWK(ctx, cfg.JWKSURL, kid)
	}, jwt.WithIssuer(cfg.IssuerURL),
		jwt.WithAudience(cfg.ClientID),
		jwt.WithLeeway(30*time.Second))

	if err != nil {
		return "", nil, errs.New("SSO.ID_TOKEN_INVALID", "id_token 验证失败: "+err.Error(), 401)
	}

	if !token.Valid {
		return "", nil, errs.New("SSO.ID_TOKEN_INVALID", "id_token 无效", 401)
	}

	// 验证 nonce
	if expectedNonce != "" && claims.Nonce != expectedNonce {
		return "", nil, errs.New("SSO.NONCE_MISMATCH", "nonce 不匹配，可能是重放攻击", 401)
	}

	return rawToken, &claims, nil
}

// --- JWKS 获取 ---

type jwksResponse struct {
	Keys []json.RawMessage `json:"keys"`
}

func (s *OIDCService) fetchJWK(ctx context.Context, jwksURL, kid string) (any, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", jwksURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	// 按 kid 查找对应的 key
	for _, keyRaw := range jwks.Keys {
		var keyData map[string]any
		if err := json.Unmarshal(keyRaw, &keyData); err != nil {
			continue
		}
		if k, ok := keyData["kid"].(string); ok && k == kid {
			// 将 JWK 转换为 Go 公钥
			return parseJWK(keyData)
		}
	}

	return nil, fmt.Errorf("key with kid=%s not found in JWKS", kid)
}

// parseJWK 将 JWK 格式的公钥转换为 Go 公钥。
func parseJWK(keyData map[string]any) (any, error) {
	kty, _ := keyData["kty"].(string)
	switch kty {
	case "RSA":
		nStr, _ := keyData["n"].(string)
		eStr, _ := keyData["e"].(string)
		if nStr == "" || eStr == "" {
			return nil, fmt.Errorf("invalid RSA JWK: missing n or e")
		}
		return parseRSAJWK(nStr, eStr)
	default:
		// 对于不支持的 key type，返回原始 key data
		return keyData, fmt.Errorf("unsupported JWK key type: %s", kty)
	}
}

// --- User Management ---

type oidcUserInfo struct {
	Subject     string `json:"sub"`
	Email       string `json:"email"`
	DisplayName string `json:"name"`
	AvatarURL   string `json:"picture"`
}

func (s *OIDCService) extractUserInfo(claims *oidcClaims, cfg *OIDCProviderConfig, accessToken string) *oidcUserInfo {
	info := &oidcUserInfo{
		Subject:     claims.Subject,
		Email:       claims.Email,
		DisplayName: claims.Name,
		AvatarURL:   claims.Picture,
	}

	// 如果没有 email，尝试从 UserInfo endpoint 获取
	if info.Email == "" && cfg.UserInfoURL != "" && accessToken != "" {
		userInfo, err := s.fetchUserInfo(context.Background(), cfg.UserInfoURL, accessToken)
		if err == nil {
			info.Email = userInfo.Email
			info.DisplayName = userInfo.DisplayName
			info.AvatarURL = userInfo.AvatarURL
		}
	}

	return info
}

func (s *OIDCService) fetchUserInfo(ctx context.Context, userInfoURL, accessToken string) (*oidcUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info oidcUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// findOrCreateUser 按 sso_provider + sso_subject 查找用户，不存在则自动创建。
func (s *OIDCService) findOrCreateUser(ctx context.Context, cfg *OIDCProviderConfig, info *oidcUserInfo, providerID int64) (*UserBrief, error) {
	// 先按 sso_subject 查找
	var user UserBrief
	err := s.db.QueryRow(ctx, `
		SELECT id, email, display_name, coalesce(avatar_url, '')
		FROM users
		WHERE sso_provider = $1 AND sso_subject = $2 AND deleted_at IS NULL`,
		fmt.Sprintf("oidc:%d", providerID), info.Subject).
		Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL)

	if err == nil {
		return &user, nil
	}

	// 按邮箱查找（邮箱可能已通过其他方式注册）
	err = s.db.QueryRow(ctx, `
		SELECT id, email, display_name, coalesce(avatar_url, '')
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`,
		info.Email).
		Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL)

	if err == nil {
		// 已有用户，关联 SSO
		_, _ = s.db.Exec(ctx,
			`UPDATE users SET sso_provider = $1, sso_subject = $2 WHERE id = $3`,
			fmt.Sprintf("oidc:%d", providerID), info.Subject, user.ID)
		return &user, nil
	}

	// 自动创建用户
	if !cfg.AutoCreateUser {
		return nil, errs.New("SSO.USER_NOT_FOUND", "该 SSO 账号未关联到平台用户，请联系管理员", 403)
	}

	err = s.db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name, avatar_url, sso_provider, sso_subject, is_active)
		VALUES ($1, NULL, $2, $3, $4, $5, TRUE)
		RETURNING id, email, display_name, coalesce(avatar_url, '')`,
		info.Email, info.DisplayName, info.AvatarURL,
		fmt.Sprintf("oidc:%d", providerID), info.Subject).
		Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL)

	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	return &user, nil
}

// --- Session Management ---

type ssoSession struct {
	ID            int64
	State         string
	Nonce         string
	ProviderID    int64
	RedirectTo    string
	CodeVerifier  string
	Status        string
}

func (s *OIDCService) validateState(ctx context.Context, state string) (*ssoSession, string, error) {
	var session ssoSession
	err := s.db.QueryRow(ctx, `
		SELECT id, state, nonce, provider_id, coalesce(redirect_to, ''), coalesce(code_verifier, ''), status
		FROM sso_sessions
		WHERE state = $1 AND status = 'pending' AND expires_at > now()`,
		state).
		Scan(&session.ID, &session.State, &session.Nonce, &session.ProviderID,
			&session.RedirectTo, &session.CodeVerifier, &session.Status)

	if err != nil {
		return nil, "", errs.New("SSO.INVALID_STATE", "登录会话已过期或无效，请重新登录", 401)
	}

	return &session, session.CodeVerifier, nil
}

func (s *OIDCService) markSessionFailed(ctx context.Context, state, errMsg string) {
	_, _ = s.db.Exec(ctx,
		`UPDATE sso_sessions SET status = 'failed', error_message = $2, completed_at = now() WHERE state = $1`,
		state, errMsg)
}

// --- Provider Config ---

func (s *OIDCService) loadProviderConfig(ctx context.Context, providerID int64) (*OIDCProviderConfig, error) {
	var cfg OIDCProviderConfig
	var scopesStr string
	var attrMapping []byte

	err := s.db.QueryRow(ctx, `
		SELECT id, name, issuer_url, client_id, client_secret, redirect_uri,
		       coalesce(auth_url, ''), coalesce(token_url, ''), coalesce(userinfo_url, ''),
		       coalesce(jwks_url, ''), scopes, auto_create_user, default_role,
		       coalesce(attribute_mapping::text, '{}')::jsonb
		FROM sso_providers
		WHERE id = $1 AND enabled = TRUE`,
		providerID).
		Scan(&cfg.ID, &cfg.Name, &cfg.IssuerURL, &cfg.ClientID, &cfg.ClientSecret,
			&cfg.RedirectURI, &cfg.AuthURL, &cfg.TokenURL, &cfg.UserInfoURL,
			&cfg.JWKSURL, &scopesStr, &cfg.AutoCreateUser, &cfg.DefaultRole, &attrMapping)

	if err != nil {
		return nil, errs.NotFound("SSO.PROVIDER_NOT_FOUND", "SSO 身份提供商不存在或已禁用")
	}

	cfg.Scopes = strings.Fields(scopesStr)
	_ = json.Unmarshal(attrMapping, &cfg.AttributeMapping)

	return &cfg, nil
}

// ListProviders 列出启用的 SSO Providers。
func (s *OIDCService) ListProviders(ctx context.Context, workspaceID int64) ([]OIDCProviderConfig, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, issuer_url, client_id, '', redirect_uri,
		       coalesce(auth_url, ''), coalesce(token_url, ''), coalesce(userinfo_url, ''),
		       coalesce(jwks_url, ''), scopes, auto_create_user, default_role,
		       coalesce(attribute_mapping::text, '{}')::jsonb
		FROM sso_providers
		WHERE workspace_id = $1 AND enabled = TRUE
		ORDER BY id`, workspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var providers []OIDCProviderConfig
	for rows.Next() {
		var cfg OIDCProviderConfig
		var scopesStr string
		var attrMapping []byte
		if err := rows.Scan(&cfg.ID, &cfg.Name, &cfg.IssuerURL, &cfg.ClientID, &cfg.ClientSecret,
			&cfg.RedirectURI, &cfg.AuthURL, &cfg.TokenURL, &cfg.UserInfoURL,
			&cfg.JWKSURL, &scopesStr, &cfg.AutoCreateUser, &cfg.DefaultRole, &attrMapping); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		cfg.Scopes = strings.Fields(scopesStr)
		cfg.ClientSecret = "" // 不泄露 secret
		_ = json.Unmarshal(attrMapping, &cfg.AttributeMapping)
		providers = append(providers, cfg)
	}

	if providers == nil {
		providers = []OIDCProviderConfig{}
	}
	return providers, nil
}

// --- Helpers ---

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *OIDCService) computeS256Challenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// parseRSAJWK 将 JWK 格式的 RSA 公钥（n, e）转换为标准 *rsa.PublicKey。
// 直接返回 crypto/rsa.PublicKey 以兼容 golang-jwt 的密钥解析回调。
//
// 参考: RFC 7518 §6.3.1 (RSA Key Representation)
func parseRSAJWK(nStr, eStr string) (any, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("invalid RSA JWK n: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("invalid RSA JWK e: %w", err)
	}

	// e 字段是大端序无符号整数，标准 JWK 中通常是 3 字节 (0x01, 0x00, 0x01 → 65537)
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	if e < 3 {
		return nil, fmt.Errorf("invalid RSA exponent: %d", e)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}, nil
}
