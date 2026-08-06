// Package config loads and validates application configuration from environment
// variables, following the 12-Factor App methodology. Missing required values
// cause a fail-fast at startup.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the root configuration object.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	NATS     NATSConfig
	Auth     AuthConfig
	Log      LogConfig
	Features FeatureFlags
}

type ServerConfig struct {
	Env  string // development | staging | production
	Port int
}

type DatabaseConfig struct {
	URL             string
	MaxConns        int32
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type NATSConfig struct {
	URL string
}

type AuthConfig struct {
	JWTIssuer         string
	JWTSecret         string // HS256 for MVP; RS256 keys in Phase 3
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	BcryptCost        int
	LoginRateLimitPer int // requests per minute per IP
}

type LogConfig struct {
	Level  string
	Format string // json | console
}

// FeatureFlags gate unfinished features so code can merge to main safely.
type FeatureFlags struct {
	SearchEnabled      bool
	AutomationEnabled  bool
	WebhooksEnabled    bool
	RegistrationOpen   bool
}

// Load reads configuration from environment variables (prefix YDSZ_) with
// sensible local-development defaults. It fails fast when required values
// are missing in non-development environments.
func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("YDSZ")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// defaults (local development friendly)
	v.SetDefault("server.env", "development")
	v.SetDefault("server.port", 8080)
	v.SetDefault("database.url", "postgres://ydsz:ydsz@localhost:5432/ydsz?sslmode=disable")
	v.SetDefault("database.max_conns", 20)
	v.SetDefault("database.conn_max_lifetime", "30m")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("auth.jwt_issuer", "ydsz-plane")
	v.SetDefault("auth.jwt_secret", "dev-only-secret-change-me")
	v.SetDefault("auth.access_token_ttl", "15m")
	v.SetDefault("auth.refresh_token_ttl", "720h") // 30d
	v.SetDefault("auth.bcrypt_cost", 12)
	v.SetDefault("auth.login_rate_limit_per", 10)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
	v.SetDefault("features.registration_open", true)

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}

	// viper does not natively parse time.Duration strings on Unmarshal for
	// nested structs in all versions; normalize explicitly.
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

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Env == "production" {
		if c.Auth.JWTSecret == "" || c.Auth.JWTSecret == "dev-only-secret-change-me" {
			return fmt.Errorf("config: YDSZ_AUTH_JWT_SECRET must be set to a strong value in production")
		}
		if c.Database.URL == "" {
			return fmt.Errorf("config: YDSZ_DATABASE_URL is required")
		}
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("config: invalid server.port %d", c.Server.Port)
	}
	return nil
}

// IsDev reports whether the server runs in development mode.
func (c *Config) IsDev() bool { return c.Server.Env == "development" }
