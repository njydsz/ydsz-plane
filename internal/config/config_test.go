// Package config — 配置加载纯逻辑单元测试（无需外部依赖）。
package config

import (
	"strings"
	"testing"
)

// TestValidate_PortRange 校验端口范围不变量。
func TestValidate_PortRange(t *testing.T) {
	cases := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"zero", 0, true},
		{"negative", -1, true},
		{"too-large", 65536, true},
		{"valid-min", 1, false},
		{"valid-max", 65535, false},
		{"default", 8080, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{
				Server:   ServerConfig{Port: tc.port, Env: "production"},
				Auth:     AuthConfig{JWTSecret: "x"},
				Database: DatabaseConfig{URL: "postgres://"},
				Attachment: AttachmentConfig{MaxFileSize: 1024},
			}
			err := cfg.validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for port %d, got nil", tc.port)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error for port %d: %v", tc.port, err)
			}
		})
	}
}

// TestValidate_ProductionRequiresSecret 校验生产环境必须有显式 JWT 密钥。
func TestValidate_ProductionRequiresSecret(t *testing.T) {
	// 空密钥 → 失败
	cfg := &Config{
		Server:   ServerConfig{Env: "production", Port: 8080},
		Auth:     AuthConfig{JWTSecret: ""},
		Database: DatabaseConfig{URL: "postgres://localhost/db"},
		Attachment: AttachmentConfig{MaxFileSize: 1024},
	}
	if err := cfg.validate(); err == nil {
		t.Fatal("expected error for empty JWT secret in production")
	}

	// dev- 前缀的密钥 → 失败
	cfg.Auth.JWTSecret = "dev-temp"
	if err := cfg.validate(); err == nil {
		t.Fatal("expected error for dev-prefixed secret in production")
		if !strings.Contains(err.Error(), "strong") {
			t.Fatalf("error message should mention 'strong', got: %v", err)
		}
	}

	// 有效生产密钥 → 通过
	cfg.Auth.JWTSecret = "prod-secret-key-at-least-32-bytes-long!!"
	if err := cfg.validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestValidate_AttachmentLimits 校验附件上传限制不变量。
func TestValidate_AttachmentLimits(t *testing.T) {
	cases := []struct {
		name string
		size int64
		want bool // true = expect error
	}{
		{"zero", 0, true},
		{"negative", -1, true},
		{"one-byte", 1, false},
		{"20mb", 20 * 1024 * 1024, false},
		{"1gb-limit", 1 << 30, false},
		{"over-1gb", 1<<30 + 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{
				Server:     ServerConfig{Env: "development", Port: 8080},
				Auth:       AuthConfig{JWTSecret: "dev-key"},
				Attachment: AttachmentConfig{MaxFileSize: tc.size},
			}
			err := cfg.validate()
			if tc.want && err == nil {
				t.Fatalf("expected error for size %d", tc.size)
			}
			if !tc.want && err != nil {
				t.Fatalf("unexpected error for size %d: %v", tc.size, err)
			}
		})
	}
}

// TestGenerateDevSecret 校验开发密钥熵与前缀。
func TestGenerateDevSecret(t *testing.T) {
	s := generateDevSecret()
	if !strings.HasPrefix(s, "dev-") {
		t.Fatalf("expected 'dev-' prefix, got: %s", s)
	}
	// "dev-" + 64 hex chars = 68
	if len(s) != 68 {
		t.Fatalf("expected length 68, got %d", len(s))
	}
	// 两次调用结果应不同（高熵）
	s2 := generateDevSecret()
	if s == s2 {
		t.Fatal("two calls returned identical secrets")
	}
}

// TestIsDev 校验环境判断。
func TestIsDev(t *testing.T) {
	dev := &Config{Server: ServerConfig{Env: "development"}}
	if !dev.IsDev() {
		t.Fatal("expected IsDev() == true for development")
	}
	prod := &Config{Server: ServerConfig{Env: "production"}}
	if prod.IsDev() {
		t.Fatal("expected IsDev() == false for production")
	}
}

// TestRedisOptions 校验 Redis 配置转换。
func TestRedisOptions(t *testing.T) {
	r := RedisConfig{Addr: "localhost:6379", Password: "secret", DB: 1}
	opts := r.RedisOptions()
	if opts.Addr != r.Addr || opts.Password != r.Password || opts.DB != r.DB {
		t.Fatalf("RedisOptions mismatch: got %+v", opts)
	}
}
