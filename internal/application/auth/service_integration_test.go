//go:build integration

// Package auth — 认证核心链路集成测试（需要真实 PostgreSQL）。
//
// 覆盖：Register → Login → ParseAccess → Refresh → Logout。
//
// 运行方式：
//
//	YDSZ_TEST_DATABASE_URL="postgres://ydsz:ydsz@localhost:5432/ydsz_plane_test" \
//		make migrate && \
//		go test -tags=integration ./internal/application/auth/...
//
// 未设置环境变量时自动跳过。
package auth

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

func authIntegrationPool(t *testing.T) *persistence.Pool {
	t.Helper()
	dsn := os.Getenv("YDSZ_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("YDSZ_TEST_DATABASE_URL not set, skipping auth integration test")
	}
	pool, err := persistence.NewPool(context.Background(), dsn, 4)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return pool
}

func newTestService(pool *persistence.Pool) *Service {
	return NewService(
		pool.Pool,
		"integration-test-secret-key-must-be-32bytes!",
		"ydsz-plane-integration",
		15*time.Minute,
		7*24*time.Hour,
		4, // bcrypt cost 调低以加速测试
		true,
	)
}

// TestAuthRegisterLoginRefresh 验证注册 → 登录 → 刷新 → 登出的完整闭环。
func TestAuthRegisterLoginRefresh(t *testing.T) {
	pool := authIntegrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	svc := newTestService(pool)
	suffix := time.Now().Format("150405.000")
	email := "integration_" + suffix + "@example.com"
	password := "Integration-Passw0rd!"

	// 1) 注册
	pair, err := svc.Register(ctx, RegisterInput{
		Email:       email,
		Password:    password,
		DisplayName: "集成测试用户",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("Register: token 为空")
	}

	// 2) ParseAccess 解析 access token
	userID, err := svc.ParseAccess(pair.AccessToken)
	if err != nil {
		t.Fatalf("ParseAccess: %v", err)
	}
	if userID == 0 {
		t.Error("ParseAccess: userID 为 0")
	}

	// 3) Login 使用同一凭据
	loginPair, err := svc.Login(ctx, email, password)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if loginPair.AccessToken == "" {
		t.Error("Login: access token 为空")
	}

	// 4) Login 错误密码 → 应失败
	if _, err := svc.Login(ctx, email, "wrong-password"); err == nil {
		t.Error("Login with wrong password should fail")
	}

	// 5) Refresh 刷新 token
	refreshed, err := svc.Refresh(ctx, loginPair.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if refreshed.AccessToken == "" {
		t.Error("Refresh: access token 为空")
	}

	// 6) Logout 登出（使 refresh token 失效）
	if err := svc.Logout(ctx, refreshed.RefreshToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	// 7) 登出后旧 refresh token 应失效
	if _, err := svc.Refresh(ctx, refreshed.RefreshToken); err == nil {
		t.Error("Refresh with revoked token should fail")
	}
}

// TestAuthHashPassword 验证密码哈希的不可逆与可验证性。
func TestAuthHashPassword(t *testing.T) {
	pool := authIntegrationPool(t)
	defer pool.Close()

	svc := newTestService(pool)

	const password = "Hash-Test-Password-123"

	hash, err := svc.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword: 返回空哈希")
	}
	if hash == password {
		t.Error("HashPassword: 哈希不应等于明文")
	}

	// 相同明文两次哈希结果应不同（bcrypt 自动加盐）。
	hash2, err := svc.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword (2nd): %v", err)
	}
	if hash == hash2 {
		t.Error("HashPassword: 相同明文应产生不同哈希（加盐）")
	}
}
