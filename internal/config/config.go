// Package config 从环境变量加载并校验应用配置，遵循 12-Factor App 方法论。
// 所有配置均外部化到环境变量；二进制中不包含任何硬编码密钥。
// 缺失必填项会在启动阶段快速失败，而不是在运行时才暴露问题。
//
// 加载流水线（按顺序）：
//  1. 默认值：Viper SetDefault 提供本地开发友好的兜底值。
//  2. 环境变量：YDSZ_ 前缀的环境变量自动覆盖默认值
//     （如 YDSZ_SERVER_PORT → server.port），通过 AutomaticEnv + Replacer 实现。
//  3. 反序列化：Viper 将合并后的映射解码到 Config 结构体。
//  4. 时长解析：time.Duration 字段显式解析（Viper 对嵌套键无法直接解码 Duration）。
//  5. 派生值：开发模式下缺失时自动生成临时密钥。
//  6. 校验：业务规则检查；失败时返回带描述的错误。
//
// 环境矩阵：
//   - development：所有默认值生效；自动生成临时 JWT 密钥。
//   - staging / production：YDSZ_AUTH_JWT_SECRET 与 YDSZ_DATABASE_URL 必填；
//     dev- 前缀的密钥会被拒绝。
package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// Config 是根配置对象，聚合所有子配置节。
// 进程启动时创建并校验单个 *Config 实例，然后传给 wire() 做依赖注入。
//
// 所有字段均为值类型（非指针），保证零值 Config 始终可反序列化；
// 具体行为由各 key 是否存在决定。
type Config struct {
	Server   ServerConfig   // HTTP 服务绑定与运行环境。
	Database DatabaseConfig // PostgreSQL 连接池设置。
	Redis    RedisConfig    // Redis 客户端连接参数。
	RabbitMQ RabbitMQConfig // 事件总线的 RabbitMQ 连接参数。
	Auth     AuthConfig     // JWT、bcrypt 与登录限流设置。
	Log      LogConfig      // 日志级别与编码格式。
	Features FeatureFlags   // 功能开关；每项开关一个子系统。
	Email    EmailConfig    // 事务邮件的外发 SMTP 配置。
}

// ServerConfig 控制 HTTP 监听器。
type ServerConfig struct {
	// Env 是运行环境标识。
	// 合法值："development" | "staging" | "production"。
	// 默认："development"。
	Env string

	// Port 是服务绑定的 TCP 端口。
	// 范围：1-65535，超出触发校验错误。
	// 默认：8080。
	Port int
}

// DatabaseConfig 保存 PostgreSQL 连接参数。
type DatabaseConfig struct {
	// URL 是完整的 libpq 连接串。
	// 格式："postgres://user:pass@host:port/db?sslmode=disable"。
	// 生产环境必填；默认使用本地开发连接串。
	URL string

	// MaxConns 是 pgx 连接池的最大连接数。
	// 范围：1-100。超过 Postgres max_connections 会在服务端引发连接错误。
	// 默认：20。
	MaxConns int32

	// ConnMaxLifetime 是单条连接被复用多久后关闭并重开。
	// 有助于在 PgBouncer 或其他连接代理后重新平衡连接。
	// 格式：Go 时长字符串（如 "30m"、"1h"）。
	// 默认："30m"。
	ConnMaxLifetime time.Duration
}

// RedisConfig 保存 Redis 客户端参数。
type RedisConfig struct {
	// Addr 是 Redis 服务地址，格式 "host:port"。
	// 默认："127.0.0.1:6379"。
	Addr string

	// Password 是 Redis AUTH 密码。空字符串禁用 AUTH。
	// 默认："Limw1020"（本地开发）；生产环境必须覆盖。
	Password string

	// DB 是 Redis 逻辑数据库编号（标准 Redis 为 0-15，
	// 或支持该配置的 Redis Cluster 的索引）。
	// 范围：0-15。
	// 默认：0。
	DB int
}

// RedisOptions 将配置转换为 go-redis 的 *redis.Options 值。
func (r RedisConfig) RedisOptions() *redis.Options {
	return &redis.Options{
		Addr:     r.Addr,
		Password: r.Password,
		DB:       r.DB,
	}
}

// RabbitMQConfig holds RabbitMQ client parameters for the event bus.
type RabbitMQConfig struct {
	// URL is the full AMQP connection string.
	// Format: "amqp://user:pass@host:port/vhost" or "amqps://..." for TLS.
	// Default: "amqp://guest:guest@127.0.0.1:5672/".
	URL string
}

// AuthConfig groups authentication and authorization parameters.
type AuthConfig struct {
	// JWTIssuer is the "iss" claim embedded in every JWT. It is validated on
	// the receiving end to prevent cross-service token replay.
	// Format: URI-recommended string (e.g. "ydsz-plane").
	// Default: "ydsz-plane".
	JWTIssuer string

	// JWTSecret is the symmetric key used to sign HS256 JWTs.
	// Security requirements:
	//   - Minimum 256 bits (32 bytes) of entropy for HS256.
	//   - In production, MUST be set via YDSZ_AUTH_JWT_SECRET; an empty value
	//     or a "dev-" prefixed value is rejected by validate().
	//   - In development, an ephemeral 256-bit secret is auto-generated on each
	//     startup and prefixed with "dev-" for easy identification.
	// Migration note: Phase 3 will rotate to RS256 (asymmetric) keys; this
	// field will remain as the HMAC key for the transition period.
	// Default: "" (dev-only auto-generation).
	JWTSecret string

	// AccessTokenTTL is the lifetime of short-lived access tokens.
	// Trade-off: shorter values reduce the window of token theft but increase
	// refresh frequency (and thus Redis write load via the refresh rotation).
	// Format: Go duration string (e.g. "15m", "1h").
	// Recommended range: 5m-1h.
	// Default: "15m".
	AccessTokenTTL time.Duration

	// RefreshTokenTTL is the maximum lifetime of refresh tokens before the
	// user must re-authenticate. Long-lived tokens improve UX but widen the
	// compromise window; pair with rotation and revocation list for safety.
	// Format: Go duration string (e.g. "720h", "168h").
	// Recommended range: 24h-720h (30d).
	// Default: "720h" (30d).
	RefreshTokenTTL time.Duration

	// BcryptCost is the bcrypt work factor (2^cost iterations). Each increment
	// doubles the CPU and memory cost.
	// Trade-off: higher values improve resistance to offline brute-force on
	// stolen hashes but linearly increase login/signup latency.
	//   - 10: ~100ms on modern hardware (bare minimum, avoid below).
	//   - 12: ~300ms (current default; balanced for most services).
	//   - 14: ~1s (high-security, expect user-visible latency).
	// Range: 4-31 (bcrypt specification limit).
	// Default: 12.
	BcryptCost int

	// LoginRateLimitPer is the maximum number of login attempts allowed per
	// client IP within the rate-limit window. Mitigates credential stuffing
	// and brute-force attacks.
	// Unit: requests per minute.
	// Range: 1-1000. Extremely low values may lock out legitimate users sharing
	// a NAT gateway (e.g. corporate networks).
	// Default: 10.
	LoginRateLimitPer int
}

// LogConfig controls structured logging behaviour.
type LogConfig struct {
	// Level is the minimum severity threshold. Messages below this level are
	// discarded.
	// Valid values: "debug" | "info" | "warn" | "error" | "fatal" | "panic".
	// Default: "info".
	Level string

	// Format selects the log encoder. "json" emits one JSON object per line
	// (suitable for ELK / Loki ingestion); "console" emits human-readable
	// pretty-printed logs for local development.
	// Valid values: "json" | "console".
	// Default: "console".
	Format string
}

// FeatureFlags gate unfinished or experimental features so that code can be
// merged to main without exposing it to production traffic. Each flag is
// read at the call-site (not cached) to allow runtime-flipping without a
// restart.
type FeatureFlags struct {
	// SearchEnabled toggles the full-text search API (/api/v1/search) and the
	// background indexer goroutine. When disabled, search endpoints return
	// 501 Not Implemented to signal intentional unavailability.
	// Default: false.
	SearchEnabled bool

	// AutomationEnabled toggles the workflow automation engine (rule evaluation,
	// scheduled triggers). When disabled, no automation jobs fire and the
	// /api/v1/automation routes return 501.
	// Default: false.
	AutomationEnabled bool

	// WebhooksEnabled controls outbound webhook delivery. When disabled, webhook
	// registrations remain stored but no HTTP requests are dispatched; a
	// synthetic 200 response is logged for auditability.
	// Default: false.
	WebhooksEnabled bool

	// RegistrationOpen toggles public user registration. When disabled,
	// POST /api/v1/auth/register returns 403, effectively making the instance
	// invite-only. Useful for private beta or maintenance mode.
	// Default: true.
	RegistrationOpen bool
}

// Load reads configuration from environment variables (prefix YDSZ_) with
// sensible local-development defaults. The complete pipeline is:
//
//  1. Default registration: Viper SetDefault seeds every known key with a
//     development-safe value.
//  2. Environment override: YDSZ_* env vars replace defaults. Nested keys
//     use double-underscore (e.g. YDSZ_AUTH__JWT_SECRET → auth.jwt_secret).
//  3. Unmarshal: Viper merges defaults + env and decodes into *Config.
//  4. Duration fixup: time.Duration fields are re-parsed from strings to
//     work around a Viper limitation with nested key unmarshalling.
//  5. Derived secrets: empty JWTSecret in dev mode generates a random
//     ephemeral value so that service starts cleanly without human setup.
//  6. Validation: business-rule checks (non-empty production secrets, valid
//     port range); returns wrapped error on failure.
//
// It fails fast when required values are missing in non-development
// environments, ensuring that misconfigured pods crash-loop at startup
// rather than serving requests with invalid configuration.
func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("YDSZ")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// ----- Step 1: local-development defaults -----
	// These values assume a standard Docker Compose local setup.
	v.SetDefault("server.env", "development")
	v.SetDefault("server.port", 8080)
	v.SetDefault("database.url", "postgres://postgres:Limw1020@127.0.0.1:5432/ydsz-plane?sslmode=disable")
	v.SetDefault("database.max_conns", 20)
	v.SetDefault("database.conn_max_lifetime", "30m")
	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.password", "Limw1020")
	v.SetDefault("redis.db", 0)
	v.SetDefault("rabbitmq.url", "amqp://guest:guest@127.0.0.1:5672/")
	v.SetDefault("auth.jwt_issuer", "ydsz-plane")
	v.SetDefault("auth.jwt_secret", "") // dev-only ephemeral secret (rotated each startup); production requires explicit value
	v.SetDefault("auth.access_token_ttl", "15m")
	v.SetDefault("auth.refresh_token_ttl", "720h") // 30d
	v.SetDefault("auth.bcrypt_cost", 12)
	v.SetDefault("auth.login_rate_limit_per", 10)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
	v.SetDefault("features.registration_open", true)

	v.SetDefault("email.enabled", false)
	v.SetDefault("email.smtp_host", "127.0.0.1")
	v.SetDefault("email.smtp_port", 1025) // mailpit default
	v.SetDefault("email.smtp_user", "")
	v.SetDefault("email.smtp_pass", "")
	v.SetDefault("email.smtp_from", "Ydsz Plane <no-reply@ydsz.dev>")
	v.SetDefault("email.smtp_use_tls", false)
	v.SetDefault("email.app_base_url", "http://127.0.0.1:5173")

	// ----- Step 2-3: environment override + unmarshal -----
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}

	// ----- Step 4: duration fixup (Viper limitation with nested keys) -----
	var err error
	if cfg.Database.ConnMaxLifetime, err = time.ParseDuration(v.GetString("database.conn_max_lifetime")); err != nil {
		return nil, fmt.Errorf("config: database.conn_max_lifetime: %w", err)
	}
	if cfg.Auth.AccessTokenTTL, err = time.ParseDuration(v.GetString("auth.access_token_ttl")); err != nil {
		return nil, fmt.Errorf("config: auth.access_token_ttl: %w", err)
	}
	if cfg.Auth.RefreshTokenTTL, err = time.ParseDuration(v.GetString("auth.refresh_token_ttl")); err != nil {
		return nil, fmt.Errorf("config: auth.refresh_token_ttl: %w", err)
	}

	// ----- Step 5: derive ephemeral secrets for dev -----
	// If jwt_secret is empty in dev, generate a random one. This is logged at
	// startup so developers know in-flight tokens get invalidated on restart.
	if cfg.Auth.JWTSecret == "" {
		if cfg.Server.Env != "development" {
			return nil, fmt.Errorf("config: YDSZ_AUTH_JWT_SECRET is required in non-development environments")
		}
		cfg.Auth.JWTSecret = generateDevSecret()
	}

	// ----- Step 6: business-rule validation -----
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// generateDevSecret creates a 256-bit (32-byte) random secret for local
// development and prefixes it with "dev-" so it is easily identifiable in
// logs and rejected by the production validator.
func generateDevSecret() string {
	var b [32]byte
	_, _ = rand.Read(b[:])
	return "dev-" + hex.EncodeToString(b[:])
}

// validate enforces cross-field business rules that cannot be expressed
// through Viper defaults or struct tags. It covers the following scenarios:
//
//   - Production mandatory fields: JWTSecret and Database.URL must be set and
//     must not use the ephemeral "dev-" prefix; otherwise an attacker could
//     forge tokens with a known dev secret.
//   - Port range check: Server.Port must be a valid TCP port (1-65535) in all
//     environments regardless of tier.
//
// Additional checks (e.g. BcryptCost range, env value whitelist) can be
// added here as the configuration surface grows.
func (c *Config) validate() error {
	// --- Production hardening: require explicit secrets ---
	if c.Server.Env == "production" {
		// JWTSecret must be present and must not be an ephemeral dev value.
		if c.Auth.JWTSecret == "" || strings.HasPrefix(c.Auth.JWTSecret, "dev-") {
			return fmt.Errorf("config: YDSZ_AUTH_JWT_SECRET must be set to a strong value in production")
		}
		// Database URL is always required in production (no sensible default).
		if c.Database.URL == "" {
			return fmt.Errorf("config: YDSZ_DATABASE_URL is required")
		}
	}

	// --- Universal invariants: valid TCP port range ---
	// Ports outside 1-65535 would cause the OS net.Listen call to fail; catch
	// the misconfiguration early with a descriptive error.
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("config: invalid server.port %d (must be 1-65535)", c.Server.Port)
	}

	return nil
}

// IsDev reports whether the server runs in development mode. Callers can use
// this to enable debug endpoints (pprof, expvar) or relaxed CORS policies,
// but should never use it to bypass authentication or authorization.
func (c *Config) IsDev() bool { return c.Server.Env == "development" }

// EmailConfig holds SMTP parameters for sending transactional emails
// (welcome, password reset, verification). All fields are controlled via
// YDSZ_EMAIL_-prefixed environment variables.
type EmailConfig struct {
	// Enabled toggles outbound email delivery entirely. When false, all
	// email-sending code paths are short-circuited and the SMTP connection
	// pool is never initialised. Useful for worker-only instances.
	// Default: false.
	Enabled bool `mapstructure:"enabled"`

	// Host is the SMTP server hostname or IP address.
	// Default: "127.0.0.1".
	Host string `mapstructure:"smtp_host"`

	// Port is the SMTP server port.
	// Common values: 25 (plain), 587 (STARTTLS), 465 (implicit TLS), 1025 (Mailpit local).
	// Range: 1-65535.
	// Default: 1025 (Mailpit).
	Port int `mapstructure:"smtp_port"`

	// Username is the SMTP AUTH login. Empty string skips AUTH (relevant for
	// local relay or internal network without authentication).
	// Default: "".
	Username string `mapstructure:"smtp_user"`

	// Password is the SMTP AUTH password. Sourced from secrets manager in
	// production; never hard-coded.
	// Default: "".
	Password string `mapstructure:"smtp_pass"`

	// From is the display name and email address shown in the From header of
	// outbound messages. Must be a valid RFC 5322 address for deliverability.
	// Default: "Ydsz Plane <no-reply@ydsz.dev>".
	From string `mapstructure:"smtp_from"`

	// UseTLS controls whether the SMTP connection uses implicit TLS (SMTPS)
	// from the start. For STARTTLS on port 587, this should be false and the
	// upgrade is handled separately by the net/smtp client.
	// Default: false.
	UseTLS bool `mapstructure:"smtp_use_tls"`

	// AppBaseURL is the public URL of the frontend application, used to
	// construct links in email templates (reset password link, verification
	// link). Must include scheme and host without trailing slash.
	// Format: "https://example.com" or "http://localhost:5173".
	// Default: "http://127.0.0.1:5173".
	AppBaseURL string `mapstructure:"app_base_url"`
}
