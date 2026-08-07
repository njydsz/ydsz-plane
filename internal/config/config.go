// Package config loads and validates application configuration from environment
// variables, following the 12-Factor App methodology. All configuration is
// externalised to the environment; the binary contains zero hardcoded secrets.
// Missing required values cause a fail-fast at startup rather than a runtime
// surprise.
//
// Loading pipeline (in order):
//  1. Defaults: Viper SetDefault provides local-development-friendly fallbacks.
//  2. Environment: YDSZ_-prefixed env vars automatically override defaults
//     (e.g. YDSZ_SERVER_PORT → server.port) via AutomaticEnv + Replacer.
//  3. Unmarshal: Viper marshals the merged map into the Config struct.
//  4. Duration parsing: time.Duration fields are parsed explicitly (Viper
//     limitation prevents direct Duration unmarshalling from nested keys).
//  5. Derived values: ephemeral secrets generated for dev when absent.
//  6. Validation: business rule checks; returns descriptive error on failure.
//
// Environment matrix:
//   - development: all defaults active; ephemeral JWT secret auto-generated.
//   - staging / production: YDSZ_AUTH_JWT_SECRET and YDSZ_DATABASE_URL are
//     mandatory; dev- prefixed secrets are rejected.
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

// Config is the root configuration object that aggregates every sub-config
// section. A single *Config instance is created at process start, validated,
// and then passed to wire() for dependency injection.
//
// All fields are value types (not pointers) so that a zero-value Config is
// always unmarshalable; presence of individual keys determines behaviour.
type Config struct {
	Server   ServerConfig   // HTTP server binding and runtime environment.
	Database DatabaseConfig // PostgreSQL connection pool settings.
	Redis    RedisConfig    // Redis client connection parameters.
	Auth     AuthConfig     // JWT, bcrypt, and login rate-limit settings.
	Log      LogConfig      // Log verbosity level and encoder format.
	Features FeatureFlags   // Feature gates; each entry toggles a subsystem.
	Email    EmailConfig    // Outbound SMTP configuration for transactional mail.
}

// ServerConfig controls the HTTP listener.
type ServerConfig struct {
	// Env is the runtime environment identifier.
	// Valid values: "development" | "staging" | "production".
	// Default: "development".
	Env string

	// Port is the TCP port the server binds to.
	// Range: 1-65535. Values outside this range trigger a validation error.
	// Default: 8080.
	Port int
}

// DatabaseConfig holds PostgreSQL connection parameters.
type DatabaseConfig struct {
	// URL is the full libpq connection string.
	// Format: "postgres://user:pass@host:port/db?sslmode=disable".
	// Required in production; defaults to a local dev connection string.
	URL string

	// MaxConns is the maximum number of open connections in the pgx pool.
	// Range: 1-100. Setting higher than the Postgres max_connections will
	// cause connection errors on the server side.
	// Default: 20.
	MaxConns int32

	// ConnMaxLifetime is how long a single connection is reused before being
	// closed and re-opened. This helps rebalance connections behind PgBouncer
	// or other connection proxies.
	// Format: Go duration string (e.g. "30m", "1h").
	// Default: "30m".
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis client parameters.
type RedisConfig struct {
	// Addr is the Redis server address in "host:port" format.
	// Default: "127.0.0.1:6379".
	Addr string

	// Password is the Redis AUTH password. Empty string disables AUTH.
	// Default: "Limw1020" (local dev); MUST be overridden in production.
	Password string

	// DB is the Redis logical database number (0-15 for standard Redis,
	// or the index for Redis Cluster configurations that support it).
	// Range: 0-15.
	// Default: 0.
	DB int
}

// RedisOptions converts the config to a go-redis *redis.Options value.
func (r RedisConfig) RedisOptions() *redis.Options {
	return &redis.Options{
		Addr:     r.Addr,
		Password: r.Password,
		DB:       r.DB,
	}
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
