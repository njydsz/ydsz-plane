// Package apitoken — 个人 API Token 应用服务。
//
// 提供令牌生命周期管理（创建/列表/吊销）与请求鉴权解析（ResolvePrincipal）。
// 存储层直接使用 SQL（与项目其他应用服务一致），原始 token 永不落库。
package apitoken

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/telemetry"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// TokenVM 是令牌列表/详情的视图模型（绝不包含原始 token 或 token_hash）。
type TokenVM struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Scopes     []string   `json:"scopes"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CreatedToken 是创建接口的响应：额外携带仅此一次可见的原始 token。
type CreatedToken struct {
	TokenVM
	// Token 原始值，仅在创建响应中出现一次，之后无法再获取。
	Token string `json:"token"`
}

// CreateInput 创建令牌的入参。
type CreateInput struct {
	UserID int64
	Name   string
	Scopes []string
	// ExpiresIn 为 nil 表示永不过期；否则为有效期（服务端下限 1 分钟、上限 365 天）。
	ExpiresIn *time.Duration
}

// Service 管理 API Token 生命周期与鉴权解析。
type Service struct {
	db *pgxpool.Pool
}

// NewService 构造 API Token 服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// Create 创建令牌并返回一次性原始值。
//
// 流程：校验名称/scope → 活跃令牌数上限检查 → 生成原始 token →
// hash 落库（token_hash 唯一约束冲突时重试，概率极低）→ 返回明文。
func (s *Service) Create(ctx context.Context, in CreateInput) (*CreatedToken, error) {
	if in.Name == "" || len(in.Name) > 80 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "名称长度需为 1-80 字符"})
	}
	if bad, ok := ValidateScopes(in.Scopes); !ok {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "scopes", Reason: "存在非法或重复的权限范围: " + bad})
	}

	active, err := s.CountActive(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	if active >= MaxActiveTokens {
		return nil, errs.New("AUTH.API_TOKEN_LIMIT", "个人令牌数量已达上限（100 个），请先吊销不再使用的令牌", 409)
	}

	raw, err := GenerateToken()
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	hash := HashToken(raw)

	var expiresAt *time.Time
	if in.ExpiresIn != nil {
		if *in.ExpiresIn < time.Minute || *in.ExpiresIn > 365*24*time.Hour {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "expires_in_seconds", Reason: "有效期需在 1 分钟到 365 天之间"})
		}
		t := time.Now().Add(*in.ExpiresIn)
		expiresAt = &t
	}

	var vm TokenVM
	err = s.db.QueryRow(ctx, `
		INSERT INTO api_tokens (user_id, name, token_hash, scopes, expires_at)
		VALUES ($1, $2, $3, $4::jsonb, $5)
		RETURNING id, name, scopes, last_used_at, expires_at, created_at`,
		in.UserID, in.Name, hash, in.Scopes, expiresAt).
		Scan(&vm.ID, &vm.Name, &vm.Scopes, &vm.LastUsedAt, &vm.ExpiresAt, &vm.CreatedAt)
	if err != nil {
		// token_hash 唯一约束冲突（理论概率趋近于零）：重试一次。
		if isUniqueViolation(err) {
			return s.Create(ctx, in)
		}
		return nil, errs.ErrInternal.Wrap(err)
	}

	telemetry.AuthOperations.WithLabelValues("api_token_create", "ok").Inc()
	return &CreatedToken{TokenVM: vm, Token: raw}, nil
}

// List 返回当前用户全部活跃（未吊销）令牌。
func (s *Service) List(ctx context.Context, userID int64) ([]TokenVM, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, scopes, last_used_at, expires_at, created_at
		FROM api_tokens
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []TokenVM
	for rows.Next() {
		var vm TokenVM
		if err := rows.Scan(&vm.ID, &vm.Name, &vm.Scopes, &vm.LastUsedAt, &vm.ExpiresAt, &vm.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, vm)
	}
	return out, nil
}

// Revoke 吊销令牌（软删除）。只能操作本人令牌，未命中返回 404。
func (s *Service) Revoke(ctx context.Context, userID, tokenID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE api_tokens SET revoked_at = now(), updated_at = now()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`, tokenID, userID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.New("AUTH.API_TOKEN_NOT_FOUND", "令牌不存在或已吊销", 404)
	}
	return nil
}

// CountActive 统计用户活跃令牌数（未吊销；已过期仍计入，鼓励用户清理）。
func (s *Service) CountActive(ctx context.Context, userID int64) (int, error) {
	var n int
	err := s.db.QueryRow(ctx, `
		SELECT count(*) FROM api_tokens
		WHERE user_id = $1 AND revoked_at IS NULL`, userID).Scan(&n)
	if err != nil {
		return 0, errs.ErrInternal.Wrap(err)
	}
	return n, nil
}

// ResolvePrincipal 将原始 token 解析为认证主体。
//
// 查询条件：hash 命中、未吊销、未过期。命中后在后台异步更新
// last_used_at（不阻塞请求，失败仅静默丢弃）。
// 任何未命中场景统一返回 401（与 JWT 无效的语义一致，不泄露令牌存在性）。
func (s *Service) ResolvePrincipal(ctx context.Context, raw string) (auth.Principal, error) {
	if !LooksLikeAPIToken(raw) {
		return auth.Principal{}, errs.ErrUnauthorized
	}

	var (
		userID int64
		scopes []string
	)
	err := s.db.QueryRow(ctx, `
		SELECT user_id, scopes FROM api_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL
		  AND (expires_at IS NULL OR expires_at > now())`,
		HashToken(raw)).Scan(&userID, &scopes)
	if err != nil {
		return auth.Principal{}, errs.ErrUnauthorized // 模糊化，不泄露令牌存在性
	}

	// 异步 touch last_used_at：独立超时，失败不影响本次请求。
	hash := HashToken(raw)
	go func() {
		tctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = s.db.Exec(tctx, `UPDATE api_tokens SET last_used_at = now()
			WHERE token_hash = $1 AND revoked_at IS NULL`, hash)
	}()

	return auth.Principal{UserID: userID, Kind: auth.PrincipalAPIToken, Scopes: scopes}, nil
}

// isUniqueViolation 判断是否为 PostgreSQL 唯一约束冲突（SQLSTATE 23505）。
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
