// Package auth 实现认证用例：密码登录、token 签发/刷新与 token 解析。
// 密码使用 bcrypt 散列；token 使用 JWT（MVP 阶段 HS256，Phase 3 切换
// 非对称密钥对 RS256）。
package auth

import (
	"context"
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
	Kind string `json:"kind"` // access | refresh
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

// Refresh 将刷新令牌轮换为一对新令牌。旧刷新令牌的复用会被
// 类型/有效期检查拒绝（轮换存储将在 S2 落地）。
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
	return s.issuePair(uid, email, displayName, avatarURL)
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
	now := time.Now()
	accessExp := now.Add(s.accessTTL)
	refreshExp := now.Add(s.refreshTTL)

	access, err := s.sign(claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: s.issuer, Subject: fmtInt(userID),
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(accessExp),
		},
		Kind: "access",
	})
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	refresh, err := s.sign(claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: s.issuer, Subject: fmtInt(userID),
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(refreshExp),
		},
		Kind: "refresh",
	})
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresAt:    accessExp,
		User:         UserBrief{ID: userID, Email: email, DisplayName: displayName, AvatarURL: avatarURL},
	}, nil
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
