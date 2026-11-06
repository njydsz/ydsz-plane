// Package oauthapp — OAuth 应用与授权域。
//
// 对标 Plane Developer Apps / GitHub OAuth Apps：
//   - OAuthApp             —— 开发者注册的 OAuth 应用（client_id / client_secret + 回调白名单 + scope）；
//   - OAuthToken           —— 授权码换取的访问/刷新令牌（原始值仅换取时返回一次）；
//   - OAuthAuthorization   —— 用户对应用的授权记录（scope 收敛、可吊销）；
//   - OAuthCode            —— 授权码（一次性、短时效）。
//
// 数据表：oauth_apps / oauth_tokens / oauth_authorizations / oauth_codes（id 由应用层 genID 生成）。
//
// 安全约束：
//   - client_secret 与 access_token / refresh_token 原始值仅创建/换取时返回一次，
//     数据库只存 SHA-256 hash；
//   - 授权码为一次性、10 分钟过期，且绑定 client_id / redirect_uri；
//   - 令牌带 scope 白名单，鉴权时按 scope 收敛校验；
//   - access_token 1 小时、refresh_token 30 天（refresh 轮换）。
package oauthapp

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// AppStatus OAuth 应用状态。
type AppStatus string

const (
	// AppStatusActive 正常可用。
	AppStatusActive AppStatus = "active"
	// AppStatusRevoked 已吊销（不可再授权/换取令牌）。
	AppStatusRevoked AppStatus = "revoked"
)

// IsValid 校验应用状态是否合法。
func (s AppStatus) IsValid() bool {
	switch s {
	case AppStatusActive, AppStatusRevoked:
		return true
	default:
		return false
	}
}

// TokenKind 标识令牌类型。
type TokenKind string

const (
	// TokenKindAccess 访问令牌。
	TokenKindAccess TokenKind = "access"
	// TokenKindRefresh 刷新令牌。
	TokenKindRefresh TokenKind = "refresh"
)

// OAuthApp OAuth 应用注册。
type OAuthApp struct {
	ID          int64  `json:"id"`
	WorkspaceID int64  `json:"workspace_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// ClientID 应用标识（公开）。
	ClientID string `json:"client_id"`
	// ClientSecret 应用密钥（仅创建/轮换时返回一次）。
	ClientSecret string `json:"-"`
	// RedirectURIs 回调地址白名单。
	RedirectURIs []string `json:"redirect_uris"`
	// Scopes 应用申请的权限范围。
	Scopes []string `json:"scopes"`
	// IconURL 图标/品牌。
	IconURL string `json:"icon_url,omitempty"`
	// Status active / revoked。
	Status    AppStatus `json:"status"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OAuthToken 颁发的访问令牌。
type OAuthToken struct {
	ID           int64     `json:"id"`
	AppID        int64     `json:"app_id"`
	UserID       int64     `json:"user_id"`
	WorkspaceID  int64     `json:"workspace_id"`
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
	Scopes       []string  `json:"scopes"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// OAuthAuthorization 用户授权记录（用户授权某应用访问其数据）。
type OAuthAuthorization struct {
	ID        int64     `json:"id"`
	AppID     int64     `json:"app_id"`
	UserID    int64     `json:"user_id"`
	Scopes    []string  `json:"scopes"`
	CreatedAt time.Time `json:"created_at"`
}

// OAuthCode 一次性授权码（短时效、绑定 client_id 与 redirect_uri）。
type OAuthCode struct {
	ID          int64     `json:"id"`
	Code        string    `json:"code"`
	AppID       int64     `json:"app_id"`
	UserID      int64     `json:"user_id"`
	WorkspaceID int64     `json:"workspace_id"`
	RedirectURI string    `json:"redirect_uri"`
	Scopes      []string  `json:"scopes"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// TokenClaims 是访问令牌校验后解析出的声明（供鉴权中间件使用）。
type TokenClaims struct {
	AppID       int64
	UserID      int64
	WorkspaceID int64
	Scopes      []string
	ExpiresAt   time.Time
}

// --- Input Types ---

// CreateAppInput 创建 OAuth 应用入参。
type CreateAppInput struct {
	WorkspaceID  int64
	Name         string
	Description  string
	RedirectURIs []string
	Scopes       []string
	IconURL      string
	CreatedBy    int64
}

// UpdateAppInput 更新 OAuth 应用入参（字段为指针，nil 表示不更新）。
type UpdateAppInput struct {
	Name         *string
	Description  *string
	RedirectURIs *[]string
	Scopes       *[]string
	IconURL      *string
}

// ListAppsOptions 应用列表查询选项。
type ListAppsOptions struct {
	WorkspaceID int64
	Limit       int
	Offset      int
}

// AuthorizeInput 授权码请求入参。
type AuthorizeInput struct {
	ClientID    string
	RedirectURI string
	Scope       string
	State       string
	UserID      int64
	WorkspaceID int64
}

// ExchangeInput 授权码换取令牌入参。
type ExchangeInput struct {
	Code         string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// RefreshInput 刷新令牌入参。
type RefreshInput struct {
	RefreshToken string
	ClientID     string
	ClientSecret string
}

// TokenResponse 令牌接口响应。
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID（与 calendar / intake / datajob 包策略一致）。
//
// 结构：高 42 位为毫秒时间戳，低 22 位为密码学随机数。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
