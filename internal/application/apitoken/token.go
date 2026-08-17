// Package apitoken — 个人 API Token（Personal Access Token）领域。
//
// 设计参考：GitHub Fine-grained PAT、GitLab Personal Access Token。
// 核心安全约束：
//   - 原始 token 仅创建时返回一次，数据库只存 SHA-256 hash；
//   - token 带 scope 白名单（scopes 收敛），鉴权时按 RBAC 权限二次校验；
//   - 支持过期时间与 last_used_at 审计，吊销为软删除（revoked_at）。
//
// 本文件仅包含纯函数（生成/哈希/校验），不依赖数据库，便于单元测试。
package apitoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// TokenPrefix 是 API Token 的固定前缀，用于：
//   - 快速识别请求携带的凭证是否为 API Token（而非 JWT）；
//   - 日志/审计中不泄露 token 本体。
const TokenPrefix = "ydz_"

// MaxActiveTokens 是单个用户可同时持有的活跃令牌上限，
// 防止令牌表无限膨胀（GitHub 限制为 10 个，此处放宽至 100）。
const MaxActiveTokens = 100

/* ------------------------------------------------------------------ */
/* Scope 白名单                                                         */
/* ------------------------------------------------------------------ */

const (
	// ScopeWorkspaceRead 读取空间/成员/项目（默认 scope）。
	ScopeWorkspaceRead = "read:workspace"
	// ScopeWorkspaceWrite 修改空间设置、成员管理、项目管理。
	ScopeWorkspaceWrite = "write:workspace"
	// ScopeIssuesRead 读取需求/任务/缺陷。
	ScopeIssuesRead = "read:issues"
	// ScopeIssuesWrite 创建/修改/删除需求/任务/缺陷。
	ScopeIssuesWrite = "write:issues"
	// ScopeSprintsRead 读取迭代。
	ScopeSprintsRead = "read:sprints"
	// ScopeSprintsWrite 管理迭代。
	ScopeSprintsWrite = "write:sprints"
	// ScopeVersionsRead 读取版本。
	ScopeVersionsRead = "read:versions"
	// ScopeVersionsWrite 管理版本。
	ScopeVersionsWrite = "write:versions"
	// ScopeAuditRead 读取审计日志。
	ScopeAuditRead = "read:audit"
	// ScopeAll 全权限通配（谨慎授予）。
	ScopeAll = "*"
)

// AllScopes 是创建向导中可选的完整 scope 列表（按展示顺序）。
var AllScopes = []string{
	ScopeWorkspaceRead,
	ScopeWorkspaceWrite,
	ScopeIssuesRead,
	ScopeIssuesWrite,
	ScopeSprintsRead,
	ScopeSprintsWrite,
	ScopeVersionsRead,
	ScopeVersionsWrite,
	ScopeAuditRead,
	ScopeAll,
}

// validScopes 是 scope 白名单（O(1) 校验）。
var validScopes = func() map[string]struct{} {
	m := make(map[string]struct{}, len(AllScopes))
	for _, s := range AllScopes {
		m[s] = struct{}{}
	}
	return m
}()

// ValidateScopes 校验 scope 列表：非空、无重复、全部命中白名单。
// 返回 nil 表示合法；否则返回第一个非法 scope 的字段错误描述。
func ValidateScopes(scopes []string) (string, bool) {
	if len(scopes) == 0 {
		return "scopes", false
	}
	seen := make(map[string]struct{}, len(scopes))
	for _, s := range scopes {
		if _, ok := validScopes[s]; !ok {
			return s, false
		}
		if _, dup := seen[s]; dup {
			return s, false
		}
		seen[s] = struct{}{}
	}
	return "", true
}

// ScopeCovers 报告 token 已拥有的 scope 集合是否覆盖 required scope。
//
// 覆盖规则（参考 GitHub scope 层次）：
//   - 完全相等即覆盖；
//   - `write:<域>` 隐含覆盖 `read:<域>`（写权限包含读权限）；
//   - `*` 覆盖一切。
func ScopeCovers(owned []string, required string) bool {
	for _, o := range owned {
		if o == ScopeAll || o == required {
			return true
		}
		if strings.HasPrefix(required, "read:") && o == "write:"+strings.TrimPrefix(required, "read:") {
			return true
		}
	}
	return false
}

// PermissionScope 将 RBAC 权限常量映射为执行该操作所需的最小 token scope。
// 返回值 ok=false 表示该权限尚未纳入 scope 映射（保守策略：API Token 拒绝）。
//
// 映射原则：写操作要求 write scope，读操作要求 read scope（write 隐含 read）。
func PermissionScope(perm string) (string, bool) {
	switch perm {
	case "workspace:read":
		return ScopeWorkspaceRead, true
	case "workspace:update", "workspace:delete",
		"member:invite", "member:remove", "member:change_role",
		"project:create", "project:delete":
		return ScopeWorkspaceWrite, true
	case "audit:read":
		return ScopeAuditRead, true
	case "issue:create", "issue:delete":
		return ScopeIssuesWrite, true
	case "version:create", "version:release", "version:delete", "version:update":
		return ScopeVersionsWrite, true
	default:
		return "", false
	}
}

/* ------------------------------------------------------------------ */
/* Token 生成与哈希                                                     */
/* ------------------------------------------------------------------ */

// GenerateToken 生成 `ydz_` 前缀的随机 token。
//
// 载荷为 32 字节 CSPRNG 随机数 + base64url 无填充编码（43 字符），
// 总长度 47 字符，熵 256 bit。原始值只出现一次，随后仅以 hash 落库。
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return TokenPrefix + base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken 计算 token 的 SHA-256 hex 摘要（与邀请 token 的存储方式一致）。
func HashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// LooksLikeAPIToken 报告该凭证是否携带 API Token 前缀。
// 用于解析器快速分流：非本前缀的凭证直接交给 JWT 路径。
func LooksLikeAPIToken(raw string) bool {
	return strings.HasPrefix(raw, TokenPrefix)
}
