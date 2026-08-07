// Package auth — 密码重置服务。
//
// 设计参考: GitHub / Linear password-reset flow。
// - Token 15 分钟有效、一次性（used_at 标记）
// - 仅以 SHA-256 hash 落库；原始 token 只在邮件中出现
// - 不论邮箱是否存在都返回 202（防枚举）
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mail"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// PasswordResetService 管理密码重置流程。
type PasswordResetService struct {
	db         *pgxpool.Pool
	mailer     mail.EmailService
	appBaseURL string
	bcryptCost int
}

// NewPasswordResetService 创建密码重置服务。
func NewPasswordResetService(db *pgxpool.Pool, mailer mail.EmailService, appBaseURL string, bcryptCost int) *PasswordResetService {
	return &PasswordResetService{db: db, mailer: mailer, appBaseURL: appBaseURL, bcryptCost: bcryptCost}
}

// RequestReset 请求发送密码重置邮件。
// 返回 nil（始终 202，防枚举）；内部有匹配用户时异步发邮件。
func (s *PasswordResetService) RequestReset(ctx context.Context, email string) error {
	var (
		userID int64
		name   string
	)
	err := s.db.QueryRow(ctx, `SELECT id, display_name FROM users WHERE email = $1 AND is_active`, email).Scan(&userID, &name)
	if err != nil {
		// 用户不存在 → 不泄露，仍返回 nil
		return nil
	}

	// 生成 token
	rawToken, hash := generateResetToken()

	_, err = s.db.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, now() + interval '15 minutes')`, userID, hash)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}

	// 异步发送邮件（不增加接口延迟）
	go s.sendResetEmail(email, name, rawToken)
	return nil
}

// ValidateToken 校验 token 是否有效（未使用、未过期）。
func (s *PasswordResetService) ValidateToken(ctx context.Context, token string) (int64, error) {
	hash := sha256Hex(token)
	var userID int64
	err := s.db.QueryRow(ctx, `
		SELECT user_id FROM password_reset_tokens
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()`, hash).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errs.New("AUTH.RESET_TOKEN_INVALID", "重置链接无效或已过期", 400)
		}
		return 0, errs.ErrInternal.Wrap(err)
	}
	return userID, nil
}

// ResetPassword 使用 token 完成密码重置。
func (s *PasswordResetService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 8 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "new_password", Reason: "密码长度至少 8 位",
		})
	}
	hash := sha256Hex(token)

	var userID int64
	err := pgx.BeginTxFunc(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// 校验 token 有效性
		err := tx.QueryRow(ctx, `
			SELECT user_id FROM password_reset_tokens
			WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()
			FOR UPDATE`, hash).Scan(&userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errs.New("AUTH.RESET_TOKEN_INVALID", "重置链接无效或已过期", 400)
			}
			return errs.ErrInternal.Wrap(err)
		}

		// Hash 新密码
		newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.bcryptCost)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}

		// 更新密码
		_, err = tx.Exec(ctx, `UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2`, string(newHash), userID)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}

		// 标记 token 已使用
		_, err = tx.Exec(ctx, `UPDATE password_reset_tokens SET used_at = now() WHERE token_hash = $1`, hash)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}

		// 失效该用户其他未使用的 token（安全收紧）
		_, err = tx.Exec(ctx, `UPDATE password_reset_tokens SET used_at = now() WHERE user_id = $1 AND used_at IS NULL`, userID)
		return err
	})
	return err
}

// --- Internal ---

func (s *PasswordResetService) sendResetEmail(toEmail, name, rawToken string) {
	if s.mailer == nil {
		return
	}
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.appBaseURL, rawToken)
	msg := mail.RenderResetPassword(mail.ResetPasswordData{
		RecipientName: name,
		ResetURL:      resetURL,
		TTLMin:        15,
	})
	msg.To = toEmail
	_ = s.mailer.Send(msg)
}

func generateResetToken() (raw, hash string) {
	b := make([]byte, 32)
	rand.Read(b)
	raw = hex.EncodeToString(b)
	hash = sha256Hex(raw)
	return
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
