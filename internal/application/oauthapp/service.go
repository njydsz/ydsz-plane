// Package oauthapp — OAuth 应用与授权应用服务。
//
// 提供：
//   - OAuth 应用 CRUD（client_id / client_secret 签发、回调白名单、scope）；
//   - OAuth2 Authorization Code 流程（Authorize → ExchangeCode → RefreshToken → RevokeToken）；
//   - 访问令牌校验（ValidateAccessToken）与 scope 校验（HasScope）。
package oauthapp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

const (
	// TokenPrefix 是 OAuth 访问令牌的固定前缀，便于与 JWT / API Token 区分。
	TokenPrefix = "oat_"
	// RefreshPrefix 是 OAuth 刷新令牌的固定前缀。
	RefreshPrefix = "ort_"
	// CodePrefix 是授权码的固定前缀。
	CodePrefix = "oac_"

	// AccessTokenTTL 访问令牌有效期。
	AccessTokenTTL = time.Hour
	// RefreshTokenTTL 刷新令牌有效期。
	RefreshTokenTTL = 30 * 24 * time.Hour
	// AuthorizationCodeTTL 授权码有效期。
	AuthorizationCodeTTL = 10 * time.Minute
)

// --- 错误 ---

var (
	// ErrAppNotFound OAuth 应用不存在。
	ErrAppNotFound = errors.New("oauth app not found")
	// ErrAppRevoked OAuth 应用已吊销。
	ErrAppRevoked = errors.New("oauth app revoked")
	// ErrRedirectURINotAllowed 回调地址不在白名单内。
	ErrRedirectURINotAllowed = errors.New("redirect uri not allowed")
	// ErrCodeNotFound 授权码不存在或已失效。
	ErrCodeNotFound = errors.New("authorization code not found or expired")
	// ErrTokenNotFound 令牌不存在或已失效。
	ErrTokenNotFound = errors.New("oauth token not found or expired")
	// ErrInvalidClientSecret client_secret 校验失败。
	ErrInvalidClientSecret = errors.New("invalid client secret")
	// ErrScopeNotGranted 令牌未被授予请求的 scope。
	ErrScopeNotGranted = errors.New("scope not granted")
)

// Service OAuth 应用与授权服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

/* ------------------------------------------------------------------ */
/* 应用 CRUD                                                           */
/* ------------------------------------------------------------------ */

// CreateApp 创建 OAuth 应用，返回 client_id 与 client_secret（仅此一次可见）。
func (s *Service) CreateApp(ctx context.Context, in CreateAppInput) (*OAuthApp, error) {
	if in.WorkspaceID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id", Reason: "required"})
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "required"})
	}
	if len(in.RedirectURIs) == 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "redirect_uris", Reason: "至少需要一个回调地址"})
	}
	for _, uri := range in.RedirectURIs {
		if !strings.HasPrefix(uri, "https://") && !strings.HasPrefix(uri, "http://localhost") {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "redirect_uris", Reason: "回调地址必须为 https:// 或 http://localhost"})
		}
	}

	clientID, err := generatePublicID()
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	clientSecret, err := generateSecret()
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	var app OAuthApp
	err = s.db.QueryRow(ctx, `
		INSERT INTO oauth_apps (id, workspace_id, name, description, client_id, client_secret_hash,
			redirect_uris, scopes, icon_url, status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7::text[],$8::text[],$9,'active',$10)
		RETURNING id, workspace_id, name, COALESCE(description,''), client_id,
			redirect_uris, scopes, COALESCE(icon_url,''), status, created_by, created_at, updated_at`,
		genID(), in.WorkspaceID, in.Name, nullIfEmpty(in.Description), clientID, HashToken(clientSecret),
		in.RedirectURIs, in.Scopes, nullIfEmpty(in.IconURL), in.CreatedBy,
	).Scan(&app.ID, &app.WorkspaceID, &app.Name, &app.Description, &app.ClientID,
		&app.RedirectURIs, &app.Scopes, &app.IconURL, &app.Status, &app.CreatedBy,
		&app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	app.ClientSecret = clientSecret // 仅创建时回传一次
	return &app, nil
}

// ListApps 列出工作空间的 OAuth 应用（不含 secret）。
func (s *Service) ListApps(ctx context.Context, opts ListAppsOptions) ([]*OAuthApp, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, name, COALESCE(description,''), client_id,
			redirect_uris, scopes, COALESCE(icon_url,''), status, created_by, created_at, updated_at
		FROM oauth_apps
		WHERE workspace_id = $1 AND status <> 'revoked'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		opts.WorkspaceID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*OAuthApp
	for rows.Next() {
		var app OAuthApp
		if err := rows.Scan(&app.ID, &app.WorkspaceID, &app.Name, &app.Description, &app.ClientID,
			&app.RedirectURIs, &app.Scopes, &app.IconURL, &app.Status, &app.CreatedBy,
			&app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &app)
	}
	return out, rows.Err()
}

// GetApp 查询应用（不含 secret）。
func (s *Service) GetApp(ctx context.Context, id int64) (*OAuthApp, error) {
	return s.getApp(ctx, id, "")
}

// GetAppByClientID 按 client_id 查询应用（不含 secret）。
func (s *Service) GetAppByClientID(ctx context.Context, clientID string) (*OAuthApp, error) {
	return s.getApp(ctx, 0, clientID)
}

// getApp 内部查询：id 或 clientID 二选一（另一个为零值）。
func (s *Service) getApp(ctx context.Context, id int64, clientID string) (*OAuthApp, error) {
	var app OAuthApp
	var err error
	if clientID != "" {
		err = s.db.QueryRow(ctx, `
			SELECT id, workspace_id, name, COALESCE(description,''), client_id,
				redirect_uris, scopes, COALESCE(icon_url,''), status, created_by, created_at, updated_at
			FROM oauth_apps WHERE client_id = $1`, clientID,
		).Scan(&app.ID, &app.WorkspaceID, &app.Name, &app.Description, &app.ClientID,
			&app.RedirectURIs, &app.Scopes, &app.IconURL, &app.Status, &app.CreatedBy,
			&app.CreatedAt, &app.UpdatedAt)
	} else {
		err = s.db.QueryRow(ctx, `
			SELECT id, workspace_id, name, COALESCE(description,''), client_id,
				redirect_uris, scopes, COALESCE(icon_url,''), status, created_by, created_at, updated_at
			FROM oauth_apps WHERE id = $1`, id,
		).Scan(&app.ID, &app.WorkspaceID, &app.Name, &app.Description, &app.ClientID,
			&app.RedirectURIs, &app.Scopes, &app.IconURL, &app.Status, &app.CreatedBy,
			&app.CreatedAt, &app.UpdatedAt)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAppNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &app, nil
}

// UpdateApp 更新应用（可更新 name/description/redirect_uris/scopes/icon_url）。
func (s *Service) UpdateApp(ctx context.Context, id int64, in UpdateAppInput) (*OAuthApp, error) {
	sets := []string{}
	args := []any{}
	add := func(expr string, val any) {
		args = append(args, val)
		sets = append(sets, expr+"$"+itoa(len(args)))
	}
	if in.Name != nil && strings.TrimSpace(*in.Name) != "" {
		add("name = ", *in.Name)
	}
	if in.Description != nil {
		add("description = ", nullIfEmpty(*in.Description))
	}
	if in.RedirectURIs != nil {
		add("redirect_uris = ", *in.RedirectURIs)
	}
	if in.Scopes != nil {
		add("scopes = ", *in.Scopes)
	}
	if in.IconURL != nil {
		add("icon_url = ", nullIfEmpty(*in.IconURL))
	}
	if len(sets) == 0 {
		return s.GetApp(ctx, id)
	}
	args = append(args, id)
	query := "UPDATE oauth_apps SET " + joinComma(sets) + ", updated_at = now() WHERE id = $" + itoa(len(args))
	if _, err := s.db.Exec(ctx, query, args...); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return s.GetApp(ctx, id)
}

// DeleteApp 删除应用（含关联令牌与授权记录）。
func (s *Service) DeleteApp(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM oauth_apps WHERE id = $1`, id)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAppNotFound
	}
	_, _ = s.db.Exec(ctx, `DELETE FROM oauth_tokens WHERE app_id = $1`, id)
	_, _ = s.db.Exec(ctx, `DELETE FROM oauth_authorizations WHERE app_id = $1`, id)
	_, _ = s.db.Exec(ctx, `DELETE FROM oauth_codes WHERE app_id = $1`, id)
	return nil
}

// RotateSecret 轮换 client_secret，返回新 secret（仅此一次可见）。
func (s *Service) RotateSecret(ctx context.Context, id int64) (string, error) {
	secret, err := generateSecret()
	if err != nil {
		return "", errs.ErrInternal.Wrap(err)
	}
	tag, err := s.db.Exec(ctx,
		`UPDATE oauth_apps SET client_secret_hash = $1, updated_at = now() WHERE id = $2 AND status = 'active'`,
		HashToken(secret), id)
	if err != nil {
		return "", errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return "", ErrAppNotFound
	}
	return secret, nil
}

/* ------------------------------------------------------------------ */
/* OAuth2 Authorization Code 流程                                      */
/* ------------------------------------------------------------------ */

// Authorize 校验 client_id / redirect_uri 后签发一次性授权码。
// 用户已同意授权（consent）后调用；scope 必须在应用已申请范围内。
func (s *Service) Authorize(ctx context.Context, in AuthorizeInput) (*OAuthCode, error) {
	app, err := s.GetAppByClientID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	if app.Status == AppStatusRevoked {
		return nil, ErrAppRevoked
	}
	if !containsString(app.RedirectURIs, in.RedirectURI) {
		return nil, ErrRedirectURINotAllowed
	}
	// 请求的 scope 必须是应用已申请范围的子集。
	scopes := splitScopes(in.Scope)
	if !subsetScopes(app.Scopes, scopes) {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "scope", Reason: "请求的权限范围超出应用已申请范围",
		})
	}

	code, err := generatePublicID()
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	expiresAt := time.Now().Add(AuthorizationCodeTTL)

	var oc OAuthCode
	err = s.db.QueryRow(ctx, `
		INSERT INTO oauth_codes (id, code, app_id, user_id, workspace_id, redirect_uri, scopes, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7::text[],$8)
		RETURNING id, code, app_id, user_id, workspace_id, redirect_uri, scopes, expires_at, created_at`,
		genID(), code, app.ID, in.UserID, in.WorkspaceID, in.RedirectURI, scopes, expiresAt,
	).Scan(&oc.ID, &oc.Code, &oc.AppID, &oc.UserID, &oc.WorkspaceID, &oc.RedirectURI,
		&oc.Scopes, &oc.ExpiresAt, &oc.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &oc, nil
}

// ExchangeCode 用授权码换取访问/刷新令牌（授权码一次性、绑定 client 与 redirect_uri）。
func (s *Service) ExchangeCode(ctx context.Context, in ExchangeInput) (*TokenResponse, error) {
	// 校验 client_secret。
	app, err := s.GetAppByClientID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	if !s.verifySecret(ctx, app.ID, in.ClientSecret) {
		return nil, ErrInvalidClientSecret
	}

	// 取并立即消费授权码（一次性）。
	var (
		codeID      int64
		userID      int64
		wsID        int64
		redirectURI string
		scopes      []string
	)
	err = s.db.QueryRow(ctx, `
		DELETE FROM oauth_codes
		WHERE code = $1 AND app_id = $2 AND redirect_uri = $3 AND expires_at > now()
		RETURNING id, user_id, workspace_id, redirect_uri, scopes`,
		in.Code, app.ID, in.RedirectURI,
	).Scan(&codeID, &userID, &wsID, &redirectURI, &scopes)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCodeNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	return s.issueTokenPair(ctx, app.ID, userID, wsID, scopes)
}

// RefreshToken 用刷新令牌换取新的访问令牌（刷新令牌轮换）。
func (s *Service) RefreshToken(ctx context.Context, in RefreshInput) (*TokenResponse, error) {
	app, err := s.GetAppByClientID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	if !s.verifySecret(ctx, app.ID, in.ClientSecret) {
		return nil, ErrInvalidClientSecret
	}

	// 校验 refresh token（hash 命中、未过期、属于该应用）。
	var (
		tokenID int64
		userID  int64
		wsID    int64
		scopes  []string
	)
	err = s.db.QueryRow(ctx, `
		SELECT id, user_id, workspace_id, scopes FROM oauth_tokens
		WHERE app_id = $1 AND refresh_hash = $2 AND refresh_expires_at > now()`,
		app.ID, HashToken(in.RefreshToken),
	).Scan(&tokenID, &userID, &wsID, &scopes)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 轮换：吊销旧 refresh token，签发新对。
	if _, err := s.db.Exec(ctx,
		`UPDATE oauth_tokens SET revoked_at = now() WHERE id = $1`, tokenID); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return s.issueTokenPair(ctx, app.ID, userID, wsID, scopes)
}

// RevokeToken 吊销访问令牌（按 access_token hash）。
func (s *Service) RevokeToken(ctx context.Context, token string) error {
	if token == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "token", Reason: "required"})
	}
	_, err := s.db.Exec(ctx,
		`UPDATE oauth_tokens SET revoked_at = now() WHERE access_hash = $1 AND revoked_at IS NULL`,
		HashToken(token))
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// issueTokenPair 签发一对访问/刷新令牌（原始值仅返回一次）。
func (s *Service) issueTokenPair(ctx context.Context, appID, userID, wsID int64, scopes []string) (*TokenResponse, error) {
	accessRaw, err := GenerateToken()
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	refreshRaw, err := GenerateToken()
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	now := time.Now()
	accessExp := now.Add(AccessTokenTTL)
	refreshExp := now.Add(RefreshTokenTTL)

	_, err = s.db.Exec(ctx, `
		INSERT INTO oauth_tokens (id, app_id, user_id, workspace_id, access_hash, refresh_hash,
			scopes, expires_at, refresh_expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7::text[],$8,$9)`,
		genID(), appID, userID, wsID, HashToken(accessRaw), HashToken(refreshRaw),
		scopes, accessExp, refreshExp)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 记录用户授权（幂等 upsert）。
	_, _ = s.db.Exec(ctx, `
		INSERT INTO oauth_authorizations (id, app_id, user_id, scopes)
		VALUES ($1,$2,$3,$4::text[])
		ON CONFLICT (app_id, user_id) DO UPDATE SET scopes = EXCLUDED.scopes`,
		genID(), appID, userID, scopes)

	return &TokenResponse{
		AccessToken:  accessRaw,
		RefreshToken: refreshRaw,
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessTokenTTL.Seconds()),
		Scope:        strings.Join(scopes, " "),
	}, nil
}

/* ------------------------------------------------------------------ */
/* 令牌校验                                                            */
/* ------------------------------------------------------------------ */

// ValidateAccessToken 校验访问令牌并返回声明（供鉴权中间件使用）。
func (s *Service) ValidateAccessToken(ctx context.Context, raw string) (*TokenClaims, error) {
	if !strings.HasPrefix(raw, TokenPrefix) {
		return nil, errs.ErrUnauthorized
	}
	var (
		appID, userID, wsID int64
		scopes              []string
		expiresAt           time.Time
		revokedAt           *time.Time
	)
	err := s.db.QueryRow(ctx, `
		SELECT app_id, user_id, workspace_id, scopes, expires_at, revoked_at
		FROM oauth_tokens WHERE access_hash = $1`,
		HashToken(raw),
	).Scan(&appID, &userID, &wsID, &scopes, &expiresAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.ErrUnauthorized
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if revokedAt != nil {
		return nil, errs.ErrUnauthorized
	}
	if time.Now().After(expiresAt) {
		return nil, errs.ErrTokenExpired
	}
	return &TokenClaims{
		AppID:       appID,
		UserID:      userID,
		WorkspaceID: wsID,
		Scopes:      scopes,
		ExpiresAt:   expiresAt,
	}, nil
}

// HasScope 报告令牌声明的 scope 集合是否覆盖 required scope。
func HasScope(c *TokenClaims, required string) bool {
	if c == nil {
		return false
	}
	for _, s := range c.Scopes {
		if s == "*" || s == required {
			return true
		}
		// write:<域> 隐含 read:<域>。
		if strings.HasPrefix(required, "read:") && s == "write:"+strings.TrimPrefix(required, "read:") {
			return true
		}
	}
	return false
}

/* ------------------------------------------------------------------ */
/* helpers                                                             */
/* ------------------------------------------------------------------ */

// verifySecret 校验 client_secret（比较 hash）。
func (s *Service) verifySecret(ctx context.Context, appID int64, secret string) bool {
	var hash string
	err := s.db.QueryRow(ctx,
		`SELECT client_secret_hash FROM oauth_apps WHERE id = $1`, appID).Scan(&hash)
	if err != nil {
		return false
	}
	return hash == HashToken(secret)
}

// generatePublicID 生成 client_id / 授权码（前缀 + 32 字节随机 base64url）。
func generatePublicID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return CodePrefix + base64.RawURLEncoding.EncodeToString(b), nil
}

// generateSecret 生成 client_secret（无前缀，48 字节随机 hex）。
func generateSecret() (string, error) {
	b := make([]byte, 48)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateToken 生成 OAuth 访问/刷新令牌原始值（前缀 + 32 字节随机 base64url）。
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return TokenPrefix + base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken 计算 token 的 SHA-256 hex 摘要。
func HashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// LooksLikeOAuthToken 报告凭证是否携带 OAuth 访问令牌前缀。
func LooksLikeOAuthToken(raw string) bool {
	return strings.HasPrefix(raw, TokenPrefix)
}

// splitScopes 将空格分隔的 scope 字符串解析为切片。
func splitScopes(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return strings.Fields(s)
}

// subsetScopes 报告 required 是否为 owned 的子集。
func subsetScopes(owned, required []string) bool {
	if len(required) == 0 {
		return true
	}
	set := make(map[string]struct{}, len(owned))
	for _, s := range owned {
		set[s] = struct{}{}
	}
	for _, r := range required {
		if _, ok := set[r]; !ok {
			return false
		}
	}
	return true
}

func containsString(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	buf := []byte{}
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}

func joinComma(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}
