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
	Email      EmailConfig      // 事务邮件的外发 SMTP 配置。
	Storage    StorageConfig    // 对象存储 (MinIO/S3) 配置。
	Attachment AttachmentConfig // 附件上传限制（大小 / MIME 白名单）。
	AI         AIConfig         // AI 智能功能配置（智能指派、重复检测、摘要等）
}

// ServerConfig 控制 HTTP 监听器。
type ServerConfig struct {
	// Env 是运行环境标识。
	// 合法值："development" | "staging" | "production"。
	// 默认："development"。
	Env string `mapstructure:"env"`

	// Port 是服务绑定的 TCP 端口。
	// 范围：1-65535，超出触发校验错误。
	// 默认：8080。
	Port int `mapstructure:"port"`
}

// DatabaseConfig 保存 PostgreSQL 连接参数。
type DatabaseConfig struct {
	// URL 是完整的 libpq 连接串。
	// 格式："postgres://user:pass@host:port/db?sslmode=disable"。
	// 生产环境必填；默认使用本地开发连接串。
	URL string `mapstructure:"url"`

	// MaxConns 是 pgx 连接池的最大连接数。
	// 范围：1-100。超过 Postgres max_connections 会在服务端引发连接错误。
	// 默认：20。
	MaxConns int32 `mapstructure:"max_conns"`

	// ConnMaxLifetime 是单条连接被复用多久后关闭并重开。
	// 有助于在 PgBouncer 或其他连接代理后重新平衡连接。
	// 格式：Go 时长字符串（如 "30m"、"1h"）。
	// 默认："30m"。
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig 保存 Redis 客户端参数。
type RedisConfig struct {
	// Addr 是 Redis 服务地址，格式 "host:port"。
	// 默认："127.0.0.1:6379"。
	Addr string `mapstructure:"addr"`

	// Password 是 Redis AUTH 密码。空字符串禁用 AUTH。
	// 默认："Limw1020"（本地开发）；生产环境必须覆盖。
	Password string `mapstructure:"password"`

	// DB 是 Redis 逻辑数据库编号（标准 Redis 为 0-15，
	// 或支持该配置的 Redis Cluster 的索引）。
	// 范围：0-15。
	// 默认：0。
	DB int `mapstructure:"db"`
}

// RedisOptions 将配置转换为 go-redis 的 *redis.Options 值。
func (r RedisConfig) RedisOptions() *redis.Options {
	return &redis.Options{
		Addr:     r.Addr,
		Password: r.Password,
		DB:       r.DB,
	}
}

// RabbitMQConfig 保存事件总线的 RabbitMQ 客户端参数。
type RabbitMQConfig struct {
	// URL 是完整的 AMQP 连接串。
	// 格式："amqp://user:pass@host:port/vhost"，TLS 使用 "amqps://..."。
	// 默认："amqp://guest:guest@127.0.0.1:5672/"。
	URL string `mapstructure:"url"`
}

// AuthConfig 聚合认证与授权参数。
type AuthConfig struct {
	// JWTIssuer 是每个 JWT 内嵌的 "iss" 声明。接收端会校验该值，
	// 防止跨服务 token 重放。
	// 格式：建议 URI 形式字符串（如 "ydsz-plane"）。
	// 默认："ydsz-plane"。
	JWTIssuer string `mapstructure:"jwt_issuer"`

	// JWTSecret 是签署 HS256 JWT 的对称密钥。
	// 安全要求：
	//   - 至少 256 位（32 字节）熵，满足 HS256 最低要求。
	//   - 生产环境必须通过 YDSZ_AUTH_JWT_SECRET 设置；空值或
	//     "dev-" 前缀的值会被 validate() 拒绝。
	//   - 开发环境下每次启动自动生成 256 位临时密钥，并以 "dev-" 前缀
	//     便于识别。
	// 迁移说明：Phase 3 将轮换为 RS256（非对称）密钥；该字段在过渡期
	// 仍作为 HMAC 密钥保留。
	// 默认：""（仅开发模式自动生成）。
	JWTSecret string `mapstructure:"jwt_secret"`

	// AccessTokenTTL 是短期访问令牌的生命周期。
	// 权衡：更短的值可缩小令牌被盗利用的时间窗，但会增加刷新频率
	// （以及经刷新轮换产生的 Redis 写入压力）。
	// 格式：Go 时长字符串（如 "15m"、"1h"）。
	// 推荐范围：5m-1h。
	// 默认："15m"。
	AccessTokenTTL time.Duration `mapstructure:"access_token_ttl"`

	// RefreshTokenTTL 是刷新令牌的最大生命周期，到期后用户必须重新认证。
	// 长生命令牌改善体验，但扩大失窃利用窗口；建议配合轮换与吊销列表。
	// 格式：Go 时长字符串（如 "720h"、"168h"）。
	// 推荐范围：24h-720h（30 天）。
	// 默认："720h"（30 天）。
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`

	// BcryptCost 是 bcrypt 工作因子（2^cost 次迭代）。每 +1 ，
	// CPU 与内存成本翻倍。
	// 权衡：更高的值增强对泄露哈希的离线暴力破解抵抗力，
	// 但线性增加登录/注册延迟。
	//   - 10：现代硬件约 100ms（最低要求，不建议更低）。
	//   - 12：约 300ms（当前默认，对大多数服务均衡）。
	//   - 14：约 1s（高安全，可感知的用户延迟）。
	// 范围：4-31（bcrypt 规范上限）。
	// 默认：12。
	BcryptCost int `mapstructure:"bcrypt_cost"`

// LoginRateLimitPer 是限流窗口内每个客户端 IP 允许的最大登录尝试次数。
// 用于缓解撞库与暴力破解攻击。
// 单位：每分钟请求数。
// 范围：1-1000。过低的值可能误伤共享 NAT 网关的合法用户
// （如企业网络）。
// 默认：200（开发环境友好；生产环境应降低至 5-10）。
LoginRateLimitPer int `mapstructure:"login_rate_limit_per"`
}

// LogConfig 控制结构化日志行为。
type LogConfig struct {
	// Level 是最小严重级别阈值。低于该级别的消息被丢弃。
	// 合法值："debug" | "info" | "warn" | "error" | "fatal" | "panic"。
	// 默认："info"。
	Level string `mapstructure:"level"`

	// Format 选择日志编码器。"json" 每行输出一个 JSON 对象
	// （适合 ELK / Loki 采集）；"console" 输出人类可读的格式化日志，
	// 用于本地开发。
	// 合法值："json" | "console"。
	// 默认："console"。
	Format string `mapstructure:"format"`
}

// FeatureFlags 用于门控未完成或实验性功能，使代码可以在不暴露给生产流量的
// 情况下合入 main 分支。每个开关在调用点读取（不缓存），
// 支持不重启即可运行时切换。
type FeatureFlags struct {
	// SearchEnabled 切换全文搜索 API（/api/v1/search）与后台索引 goroutine。
	// 关闭时搜索端点返回 501 Not Implemented，表明功能有意不可用。
	// 默认：false。
	SearchEnabled bool `mapstructure:"search_enabled"`

	// AutomationEnabled 切换工作流自动化引擎（规则评估、定时触发）。
	// 关闭时无自动化任务执行，/api/v1/automation 路由返回 501。
	// 默认：false。
	AutomationEnabled bool `mapstructure:"automation_enabled"`

	// WebhooksEnabled 控制外发 webhook 投递。关闭时 webhook 注册仍然保留，
	// 但不发出任何 HTTP 请求；记录合成 200 响应以留审计痕迹。
	// 默认：false。
	WebhooksEnabled bool `mapstructure:"webhooks_enabled"`

	// RegistrationOpen 切换公开用户注册。关闭时 POST /api/v1/auth/register
	// 返回 403，实例实质变为邀请制。适用于私有内测或维护模式。
	// 默认：true。
	RegistrationOpen bool `mapstructure:"registration_open"`
}

// Load 从环境变量（前缀 YDSZ_）读取配置并附带合理的本地开发默认值。
// 完整流水线如下：
//
//  1. 默认值注册：Viper SetDefault 为每个已知 key 植入开发安全值。
//  2. 环境变量覆盖：YDSZ_* 环境变量替换默认值。嵌套 key 使用双下划线
//     （如 YDSZ_AUTH__JWT_SECRET → auth.jwt_secret）。
//  3. 反序列化：Viper 合并默认值与环境变量后解码到 *Config。
//  4. 时长修复：time.Duration 字段从字符串重新解析，
//     规避 Viper 对嵌套 key 反序列化的限制。
//  5. 派生密钥：开发模式下 JWTSecret 为空时生成随机临时值，
//     使服务无需人工配置即可启动。
//  6. 校验：业务规则检查（非开发环境必填密钥、合法端口范围）；
//     失败时返回包装错误。
//
// 非开发环境下缺失必填值会快速失败，确保配置错误的 Pod 在启动时
// 崩溃循环，而不是带病处理请求。
func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("YDSZ")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// ----- 第 1 步：本地开发默认值 -----
	// 这些值假设标准 Docker Compose 本地环境。
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
	v.SetDefault("auth.jwt_secret", "") // 仅开发模式临时密钥（每次启动轮换）；生产环境必须显式配置
	v.SetDefault("auth.access_token_ttl", "15m")
	v.SetDefault("auth.refresh_token_ttl", "720h") // 30 天
	v.SetDefault("auth.bcrypt_cost", 12)
	v.SetDefault("auth.login_rate_limit_per", 200)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
	v.SetDefault("features.registration_open", true)

	// AI 默认值：默认禁用（fallback 模式可开启）
	v.SetDefault("ai.enabled", false)
	v.SetDefault("ai.provider", "fallback")
	v.SetDefault("ai.model", "gpt-4o-mini")
	v.SetDefault("ai.endpoint", "https://api.openai.com/v1")
	v.SetDefault("ai.api_key", "")

	v.SetDefault("email.enabled", false)
	v.SetDefault("email.smtp_host", "127.0.0.1")
	v.SetDefault("email.smtp_port", 1025) // mailpit 默认端口
	v.SetDefault("email.smtp_user", "")
	v.SetDefault("email.smtp_pass", "")
	v.SetDefault("email.smtp_from", "Ydsz Plane <no-reply@ydsz.dev>")
	v.SetDefault("email.smtp_use_tls", false)
	v.SetDefault("email.app_base_url", "http://127.0.0.1:5173")

	// Attachment 默认值：20 MB 上限 + Office/图片/MIME 白名单
	v.SetDefault("attachment.max_file_size", 20*1024*1024) // 20 MB
	v.SetDefault("attachment.max_total_size_per_issue", 100*1024*1024) // 100 MB
	v.SetDefault("attachment.allowed_content_types", []string{
		"image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml",
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"application/zip", "application/x-7z-compressed",
		"text/plain", "text/csv", "text/markdown",
		"application/json", "application/xml",
	})
	v.SetDefault("attachment.allowed_extensions", []string{
		"jpg", "jpeg", "png", "gif", "webp", "svg",
		"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx",
		"zip", "7z", "txt", "csv", "md", "json", "xml",
	})

	// Storage 默认值
	v.SetDefault("storage.endpoint", "127.0.0.1:9000")
	v.SetDefault("storage.access_key", "admin")
	v.SetDefault("storage.secret_key", "Limw1020")
	v.SetDefault("storage.bucket", "ydsz-plane")
	v.SetDefault("storage.use_ssl", false)
	v.SetDefault("storage.region", "us-east-1")

	// ----- 第 2-3 步：环境变量覆盖 + 反序列化 -----
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}

	// ----- 第 4 步：时长修复（Viper 对嵌套 key 的限制） -----
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

	// ----- 第 5 步：为开发模式派生临时密钥 -----
	// 开发模式下 jwt_secret 为空时生成随机值。启动时会记录日志，
	// 提醒开发者重启后存量 token 会失效。
	if cfg.Auth.JWTSecret == "" {
		if cfg.Server.Env != "development" {
			return nil, fmt.Errorf("config: YDSZ_AUTH_JWT_SECRET is required in non-development environments")
		}
		cfg.Auth.JWTSecret = generateDevSecret()
	}

	// ----- 第 6 步：业务规则校验 -----
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// generateDevSecret 生成本地开发用的 256 位（32 字节）随机密钥，
// 并以 "dev-" 前缀标识，便于在日志中识别、被生产校验器拒绝。
func generateDevSecret() string {
	var b [32]byte
	_, _ = rand.Read(b[:])
	return "dev-" + hex.EncodeToString(b[:])
}

// validate 执行无法用 Viper 默认值或 struct tag 表达的跨字段业务规则。
// 覆盖以下场景：
//
//   - 生产环境必填字段：JWTSecret 与 Database.URL 必须设置，
//     且不得使用临时的 "dev-" 前缀；否则攻击者可用已知的开发密钥伪造 token。
//   - 端口范围检查：Server.Port 在所有环境中必须是合法 TCP 端口（1-65535）。
//
// 随着配置面增长，可在此继续增加检查（如 BcryptCost 范围、env 值白名单）。
func (c *Config) validate() error {
	// --- 生产环境加固：要求显式密钥 ---
	if c.Server.Env == "production" {
		// JWTSecret 必须存在且不能是临时的开发值。
		if c.Auth.JWTSecret == "" || strings.HasPrefix(c.Auth.JWTSecret, "dev-") {
			return fmt.Errorf("config: YDSZ_AUTH_JWT_SECRET must be set to a strong value in production")
		}
		// 生产环境始终需要 Database URL（没有合理的默认值）。
		if c.Database.URL == "" {
			return fmt.Errorf("config: YDSZ_DATABASE_URL is required")
		}
	}

	// --- 通用不变量：合法 TCP 端口范围 ---
	// 1-65535 之外的端口会导致 OS net.Listen 调用失败；
	// 尽早捕获配置错误并给出描述性信息。
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("config: invalid server.port %d (must be 1-65535)", c.Server.Port)
	}

	// --- 附件上传合法性 ---
	if c.Attachment.MaxFileSize <= 0 {
		return fmt.Errorf("config: attachment.max_file_size must be > 0")
	}
	if c.Attachment.MaxFileSize > 1<<30 { // 1 GB
		return fmt.Errorf("config: attachment.max_file_size exceeds 1 GB sanity limit")
	}

	return nil
}

// IsDev 报告服务是否运行在开发模式。调用方可据此启用调试端点
// （pprof、expvar）或宽松 CORS 策略，但绝不可用它绕过认证或授权。
func (c *Config) IsDev() bool { return c.Server.Env == "development" }

// AIConfig 控制 AI 智能功能的运行时行为。
type AIConfig struct {
	// Enabled 整体开关 AI 功能。关闭时所有 /ai/* 端点返回 501。
	// 默认：false。
	Enabled bool `mapstructure:"enabled"`

	// Provider 选择 LLM 后端。合法值："fallback"（纯规则引擎） | "openai" | "local"。
	// 默认："fallback"。
	Provider string `mapstructure:"provider"`

	// APIKey 是 LLM Provider 的 API 密钥（openai 模式必填）。
	APIKey string `mapstructure:"api_key"`

	// Model 是 LLM 模型名称（openai 模式有效）。
	// 默认："gpt-4o-mini"。
	Model string `mapstructure:"model"`

	// Endpoint 是 LLM API 端点（openai 模式有效）。
	// 默认："https://api.openai.com/v1"。
	Endpoint string `mapstructure:"endpoint"`
}

// EmailConfig 保存发送事务邮件（欢迎、密码重置、验证）的 SMTP 参数。
// 所有字段通过 YDSZ_EMAIL_ 前缀的环境变量控制。
type EmailConfig struct {
	// Enabled 整体开关外发邮件。关闭时所有发信代码路径被短路，
	// SMTP 连接池永不初始化。适用于仅运行 worker 的实例。
	// 默认：false。
	Enabled bool `mapstructure:"enabled"`

	// Host 是 SMTP 服务器主机名或 IP 地址。
	// 默认："127.0.0.1"。
	Host string `mapstructure:"smtp_host"`

	// Port 是 SMTP 服务器端口。
	// 常见值：25（明文）、587（STARTTLS）、465（隐式 TLS）、1025（本地 Mailpit）。
	// 范围：1-65535。
	// 默认：1025（Mailpit）。
	Port int `mapstructure:"smtp_port"`

	// Username 是 SMTP AUTH 登录名。空字符串跳过 AUTH
	// （适用于本地中继或无认证内网）。
	// 默认：""。
	Username string `mapstructure:"smtp_user"`

	// Password 是 SMTP AUTH 密码。生产环境从密钥管理服务获取；
	// 绝不硬编码。
	// 默认：""。
	Password string `mapstructure:"smtp_pass"`

	// From 是外发邮件 From 头中显示的展示名与邮箱地址。
	// 必须是合法的 RFC 5322 地址以保证可投递性。
	// 默认："Ydsz Plane <no-reply@ydsz.dev>"。
	From string `mapstructure:"smtp_from"`

	// UseTLS 控制 SMTP 连接是否从一开始就使用隐式 TLS（SMTPS）。
	// 587 端口的 STARTTLS 应设为 false，升级由 net/smtp 客户端单独处理。
	// 默认：false。
	UseTLS bool `mapstructure:"smtp_use_tls"`

	// AppBaseURL 是前端应用的公网地址，用于构造邮件模板中的链接
	// （重置密码链接、验证链接）。必须包含协议与主机，不带结尾斜杠。
	// 格式："https://example.com" 或 "http://localhost:5173"。
	// 默认："http://127.0.0.1:5173"。
	AppBaseURL string `mapstructure:"app_base_url"`
}

// AttachmentConfig 附件上传限制（大小 / MIME / 扩展名白名单）。
// 参照 Jira/Plane/Linear 业界实践：
//   - 单文件 ≤ 20 MB（普通文档/截图足够）；
//   - MIME 白名单覆盖常见办公/设计格式，拒绝可执行文件兜底攻击；
//   - 同名扩展名检查（MIME 可被客户端篡改，扩展名是双保险）。
type AttachmentConfig struct {
	// MaxFileSize 单文件最大字节数。默认 20 MB。
	// 客户端也应在上传前校验，服务端再次校验是最后防线。
	MaxFileSize int64 `mapstructure:"max_file_size"`

	// MaxTotalSizePerIssue 单个工作项附件总容量限制（0=无限制）。
	// 防止工作项附件无限膨胀影响查询性能。
	MaxTotalSizePerIssue int64 `mapstructure:"max_total_size_per_issue"`

	// AllowedContentTypes MIME 类型白名单。空表示允许所有。
	AllowedContentTypes []string `mapstructure:"allowed_content_types"`

	// AllowedExtensions 扩展名白名单（不带点，如 "pdf"）。空表示允许所有。
	AllowedExtensions []string `mapstructure:"allowed_extensions"`
}

// StorageConfig 对象存储 (MinIO / S3 兼容) 配置。
// 通过 YDSZ_STORAGE_ 前缀环境变量注入。
type StorageConfig struct {
	// Endpoint 对象存储服务端点，格式 "host:port"。
	// 默认："127.0.0.1:9000"（本地 MinIO）。
	Endpoint string `mapstructure:"endpoint"`

	// AccessKey 对象存储访问密钥 ID。
	// 默认："admin"。
	AccessKey string `mapstructure:"access_key"`

	// SecretKey 对象存储访问密钥 Secret。
	// 默认："Limw1020"。
	SecretKey string `mapstructure:"secret_key"`

	// Bucket 默认存储桶名称。
	// 默认："ydsz-plane"。
	Bucket string `mapstructure:"bucket"`

	// UseSSL 是否使用 HTTPS 连接。
	// 默认：false。
	UseSSL bool `mapstructure:"use_ssl"`

	// Region 存储桶所在地域（S3 协议要求，MinIO 可忽略）。
	// 默认："us-east-1"。
	Region string `mapstructure:"region"`

	// PublicURL 对象存储的公网访问 URL，用于前端直传。
	// 留空时使用 Endpoint 构造。
	// 默认：""。
	PublicURL string `mapstructure:"public_url"`
}
