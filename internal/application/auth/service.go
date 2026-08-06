// Package auth implements authentication use cases: password login, token
// issue/refresh, and token parsing. Passwords are bcrypt-hashed; tokens are
// JWT (HS256 for MVP, RS256 via key pair in Phase 3).
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/ydszopen/ydsz-plane/pkg/errs"
)

// Service provides auth use cases.
type Service struct {
	db              *pgxpool.Pool
	secret          []byte
	issuer          string
	accessTTL       time.Duration
	refreshTTL      time.Duration
	bcryptCost      int
	registrationOpen bool
}

// NewService constructs the auth service.
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

// TokenPair is the login/refresh response payload.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"` // Bearer
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserBrief `json:"user"`
}

// UserBrief is the embedded user summary in auth responses.
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

// Login authenticates by email+password and issues a token pair.
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
			return nil, errs.ErrInvalidCredentials // 模糊化，不区分账号/密码
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	if !isActive || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, errs.ErrInvalidCredentials
	}
	return s.issuePair(id, email, displayName, avatarURL)
}

// Refresh rotates a refresh token into a new pair. Reuse of an old refresh
// token is rejected by kind/expiry checks (rotation storage lands in S2).
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

// ParseAccess validates an access token and returns the user id.
func (s *Service) ParseAccess(token string) (int64, error) {
	c, err := s.parse(token)
	if err != nil || c.Kind != "access" {
		return 0, fmt.Errorf("invalid access token")
	}
	return parseSubject(c.Subject)
}

// HashPassword hashes a plaintext password (used by registration/seed).
func (s *Service) HashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), s.bcryptCost)
	return string(h), err
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
