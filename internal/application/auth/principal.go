// Package auth — 认证主体（Principal）定义。
//
// Principal 统一描述一条已认证请求的调用方身份：它既可以来自
// 会话 JWT（浏览器 SPA），也可以来自个人 API Token（脚本/集成）。
// 中间件链据此区分认证来源，并对 API Token 施加 scope 收敛校验
// （见 internal/application/apitoken 与 middleware/rbac.go）。
package auth

// PrincipalKind 标识认证凭证的类型。
type PrincipalKind string

const (
	// PrincipalJWT 表示会话 JWT access token（浏览器 SPA）。
	PrincipalJWT PrincipalKind = "jwt"
	// PrincipalAPIToken 表示个人 API Token（ydz_ 前缀）。
	PrincipalAPIToken PrincipalKind = "api_token"
)

// Principal 是 RequireAuth 解析后的认证主体。
//
// UserID 恒有值；Kind 区分凭证类型；Scopes 仅在 Kind == PrincipalAPIToken
// 时非空（来自 api_tokens.scopes，是授权收敛的唯一依据）。
type Principal struct {
	UserID int64
	Kind   PrincipalKind
	Scopes []string
}
