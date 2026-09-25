// Package auth 实现认证用例：密码登录、token 签发/刷新与 token 解析。
// 密码使用 bcrypt 散列；token 使用 JWT（MVP 阶段 HS256，Phase 3 切换
// 非对称密钥对 RS256）。
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/telemetry"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 提供认证用例：登录、注册、token 签发/解析。
//
// 安全约束：
//   - secret 长度须 ≥ 32 字节（HS256 最低要求）。
//   - accessTTL / refreshTTL 建议分别为 15min / 7d（refresh 另见 S2 refresh rotation）。
//   - bcryptCost 建议 12（≈ 250ms/哈希，每 +1 约耗时翻倍）。
type Service struct {
	db               *pgxpool.Pool
	secret           []byte        // JWT 签名密钥（HS256 共享密钥）。
	issuer           string        // JWT iss 声明，用于多租户签发者区分。
	accessTTL        time.Duration // access token 有效期。
	refreshTTL       time.Duration // refresh token 有效期。
	bcryptCost       int           // bcrypt 成本因子（4-31）。
	registrationOpen bool          // 是否开放注册；false 时仅允许邀请注册。
}

// NewService 构造认证服务。
func NewService(db *pgxpool.Pool, secret, issuer string, accessTTL, refreshTTL time.Duration, bcryptCost int, registrationOpen bool) *Service {
	return &Service{
		db:               db,
		secret:           []byte(secret),
		issuer:           issuer,
		accessTTL:        accessTTL,
		refreshTTL:       refreshTTL,
		bcryptCost:       bcryptCost,
		registrationOpen: registrationOpen,
	}
}

// TokenPair 是登录/刷新接口的响应负载。
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"` // Bearer
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserBrief `json:"user"`
}

// UserBrief 是认证响应中内嵌的用户概要。
type UserBrief struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type claims struct {
	jwt.RegisteredClaims
	Kind     string `json:"kind"` // access | refresh
	FamilyID string `json:"fid"` // refresh token family 标识（仅 refresh token 携带）
}

// Login 使用邮箱+密码认证并签发令牌对。
func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	var (
		id          int64
		hash        string
		displayName string
		avatarURL   string
		isActive    bool
	)
	err := s.db.QueryRow(ctx,
		`SELECT id, password_hash, display_name, coalesce(avatar_url,''), is_active
		 FROM users WHERE email = $1`, email).
		Scan(&id, &hash, &displayName, &avatarURL, &isActive)
	if err != nil {
		if errors.Is(err, pgxErrNoRows()) {
			telemetry.AuthOperations.WithLabelValues("login", "user_not_found").Inc()
			return nil, errs.ErrInvalidCredentials // 模糊化，不区分账号/密码
		}
		telemetry.AuthOperations.WithLabelValues("login", "error").Inc()
		return nil, errs.ErrInternal.Wrap(err)
	}
	if !isActive || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		telemetry.AuthOperations.WithLabelValues("login", "invalid").Inc()
		return nil, errs.ErrInvalidCredentials
	}
	telemetry.AuthOperations.WithLabelValues("login", "ok").Inc()
	return s.issuePair(id, email, displayName, avatarURL)
}

// Refresh 将刷新令牌轮换为一对新令牌（P0-3：Refresh Token 轮换与家族吊销）。
//
// 流程:
// 1. 解析 old refresh token 取得 family_id（若无 fid claim 则降级为老逻辑）。
// 2. 查 refresh_token_families 表：
//   - 若 revoked_at IS NOT NULL → 该登录 session 已被安全吊销，返回 ErrInvalidCredentials + 吊销提示。
//   - 若未 revoke → 标记为 revoked(reason=rotation)，签发新的 fid token 对，插入新行。
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	c, err := s.parse(refreshToken)
	if err != nil || c.Kind != "refresh" {
		return nil, errs.ErrTokenExpired
	}
	var (
		email, displayName, avatarURL string
		isActive                      bool
	)
	uid, err := parseSubject(c.Subject)
	if err != nil {
		return nil, errs.ErrTokenExpired
	}
	err = s.db.QueryRow(ctx,
		`SELECT email, display_name, coalesce(avatar_url,''), is_active FROM users WHERE id = $1`, uid).
		Scan(&email, &displayName, &avatarURL, &isActive)
	if err != nil || !isActive {
		return nil, errs.ErrTokenExpired
	}

	// P0-3：家族轮换检测（若无 fid 则向后兼容跳过）
	if c.FamilyID != "" {
		revoked, checkErr := s.isFamilyRevoked(ctx, c.FamilyID)
		if checkErr != nil {
			telemetry.AuthOperations.WithLabelValues("refresh", "family_check_error").Inc()
			// 家族表查询失败时不阻断刷新，降级为跳过轮换检测
		} else if revoked {
			// 该家族已被吊销：可能是 refresh token 被盗用后的重放。
			// 吊销信号：提示用户该 session 已被安全吊销，请重新登录。
			telemetry.AuthOperations.WithLabelValues("refresh", "family_replay_detected").Inc()
			return nil, errs.New("AUTH.SESSION_REVOKED",
				"该登录 session 已被安全吊销，可能为 refresh token 重放攻击，请重新登录", 401)
		}

		// 标记旧家族为已轮换（非阻塞：失败不影响新 token 签发）
		if rotateErr := s.revokeFamily(ctx, c.FamilyID, "rotation"); rotateErr != nil {
			telemetry.AuthOperations.WithLabelValues("refresh", "family_rotate_error").Inc()
		}
	}

	return s.issuePair(uid, email, displayName, avatarURL)
}

// Logout 吊销指定 refresh token 对应的家族（reason=logout）。
// 用于"退出登录"场景：仅吊销此 refresh token 族，不影响用户其他会话。
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	c, err := s.parse(refreshToken)
	if err != nil || c.Kind != "refresh" {
		return errs.ErrTokenExpired
	}
	if c.FamilyID == "" {
		// 老版本 token 无 fid，无法关联家族记录
		return nil
	}
	return s.revokeFamily(ctx, c.FamilyID, "logout")
}

// isFamilyRevoked 查询指定 refresh token 家族是否已被吊销。
// 表不存在时返回 (false, nil) 以兼容尚未执行 migration 的环境（降级为无家族追踪模式）。
func (s *Service) isFamilyRevoked(ctx context.Context, familyID string) (bool, error) {
	var revokedAt *time.Time
	err := s.db.QueryRow(ctx,
		`SELECT revoked_at FROM refresh_token_families WHERE family_id = $1`, familyID).
		Scan(&revokedAt)
	if err != nil {
		// 未找到记录：家族不存在，视为未吊销
		return false, nil
	}
	return revokedAt != nil, nil
}

// revokeFamily 将指定家族标记为 revoked（原子 UPDATE，仅当尚未 revoke 时写入）。
func (s *Service) revokeFamily(ctx context.Context, familyID, reason string) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE refresh_token_families
		    SET revoked_at = now(), revoked_reason = $2
		  WHERE family_id = $1 AND revoked_at IS NULL`,
		familyID, reason)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() > 0 {
		telemetry.AuthOperations.WithLabelValues("revoke_family", reason).Inc()
	}
	return nil
}

// ParseAccess 校验访问令牌并返回用户 ID。
func (s *Service) ParseAccess(token string) (int64, error) {
	c, err := s.parse(token)
	if err != nil || c.Kind != "access" {
		return 0, fmt.Errorf("invalid access token")
	}
	return parseSubject(c.Subject)
}

// HashPassword 将明文密码散列（供注册/seed 使用）。
func (s *Service) HashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), s.bcryptCost)
	return string(h), err
}

// RegisterInput 保存注册参数。
type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

// Register 创建新用户并签发访问/刷新令牌对。
func (s *Service) Register(ctx context.Context, in RegisterInput) (*TokenPair, error) {
	if !s.registrationOpen {
		return nil, errs.ErrForbidden
	}
	if len(in.Password) < 8 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "password", Reason: "密码长度至少 8 位",
		})
	}

	hash, err := s.HashPassword(in.Password)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	var (
		userID int64
		email  string
	)
	err = s.db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING
		RETURNING id, email`,
		in.Email, hash, in.DisplayName).Scan(&userID, &email)
	if err != nil {
		// 邮箱冲突 → 已注册
		telemetry.AuthOperations.WithLabelValues("register", "conflict").Inc()
		return nil, errs.New("AUTH.EMAIL_TAKEN", "该邮箱已被注册", 409)
	}

	telemetry.AuthOperations.WithLabelValues("register", "ok").Inc()
	return s.issuePair(userID, email, in.DisplayName, "")
}

func (s *Service) issuePair(userID int64, email, displayName, avatarURL string) (*TokenPair, error) {
	// P0-3：provider_id 默认为 0 表示密码登录（sso callback 场景传真实 providerID）
	return s.issuePairWithProvider(userID, email, displayName, avatarURL, 0)
}

// issuePairWithProvider 签发令牌对，SSO 场景传入 provider_id 以便关联家族记录。
func (s *Service) issuePairWithProvider(userID int64, email, displayName, avatarURL string, providerID int64) (*TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(s.accessTTL)
	refreshExp := now.Add(s.refreshTTL)

	// 为每次 refresh token 生成唯一家族 ID（P0-3：轮换吊销家族标识）
	familyID, err := generateFamilyID()
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	access, err := s.sign(claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: s.issuer, Subject: fmtInt(userID),
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(accessExp),
		},
		Kind: "access",
		// access token 不携带 fid
	})
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	refresh, err := s.sign(claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: s.issuer, Subject: fmtInt(userID),
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(refreshExp),
		},
		Kind:     "refresh",
		FamilyID: familyID,
	})
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// P0-3：记录 refresh token 家族
	// provider_id: 密码登录为 NULL, SSO 场景传真实值
	if insertErr := s.recordRefreshFamily(context.Background(), familyID, userID, providerID, refreshExp); insertErr != nil {
		telemetry.AuthOperations.WithLabelValues("issue_pair", "family_record_error").Inc()
		// 非阻塞：家族记录失败不影响 token 签发（降级为无家族追踪模式）
	}

	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresAt:    accessExp,
		User:         UserBrief{ID: userID, Email: email, DisplayName: displayName, AvatarURL: avatarURL},
	}, nil
}

// generateFamilyID 生成随机的 refresh token 家族 UUID。
func generateFamilyID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// recordRefreshFamily 向 refresh_token_families 表插入新家族记录。
// providerID 为 0 时存 NULL（密码登录场景）。
func (s *Service) recordRefreshFamily(ctx context.Context, familyID string, userID int64, providerID int64, expiresAt time.Time) error {
	if s.db == nil {
		return nil
	}
	var pid *int64
	if providerID != 0 {
		pid = &providerID
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO refresh_token_families (family_id, user_id, provider_id, expires_at)
		VALUES ($1, $2, $3, $4)`,
		familyID, userID, pid, expiresAt)
	return err
}

// CleanupExpiredFamilies 清理过期 7 天的 refresh token 家族记录。
// 建议作为后台定时任务每日执行一次（由 cron_scheduler 调度或外部 cron 触发）。
func (s *Service) CleanupExpiredFamilies(ctx context.Context) (int64, error) {
	if s.db == nil {
		return 0, nil
	}
	tag, err := s.db.Exec(ctx,
		`DELETE FROM refresh_token_families WHERE expires_at < now() - interval '7 days'`)
	if err != nil {
		return 0, errs.ErrInternal.Wrap(err)
	}
	return tag.RowsAffected(), nil
}

func (s *Service) sign(c claims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.secret)
}

func (s *Service) parse(token string) (*claims, error) {
	var c claims
	_, err := jwt.ParseWithClaims(token, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	}, jwt.WithIssuer(s.issuer))
	return &c, err
}
