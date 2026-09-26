// Package health 提供外部依赖的连通性探测（ConnectivityChecker），
// 供 /readyz 端点聚合输出 Kubernetes-style 就绪探针响应。
//
// 设计要点：
//   - ConnectivityChecker 接口统一各依赖的探测行为
//   - 每个 checker 独立设置超时（PostgreSQL/Redis 3s，RabbitMQ/Elasticsearch/OIDC 5s）
//   - SMTP 未配置时跳过（Host 为空）
//   - OIDC 仅探测已启用 Provider 的 IssuerURL
//
// /readyz 响应格式遵循 Kubernetes probes 惯例：
//
//	200: {"status":"ok","checks":{"postgres":"ok","redis":"ok",...}}
//	503: {"status":"degraded","checks":{"postgres":"error: ..."}}
package health

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// ConnectivityChecker 定义外部依赖的连通性探测接口。
type ConnectivityChecker interface {
	// Check 执行连通性探测，ctx 携带超时。
	Check(ctx context.Context) error
	// Name 返回 checker 名称（用于 /readyz 响应 JSON key）。
	Name() string
}

// --- PostgreSQL ---

// PostgresChecker 通过 db.Ping 探测 PostgreSQL 连通性。
type PostgresChecker struct {
	db *pgxpool.Pool
}

// NewPostgresChecker 创建 PostgreSQL 连通性 checker。
func NewPostgresChecker(db *pgxpool.Pool) *PostgresChecker {
	return &PostgresChecker{db: db}
}

// Name 返回 checker 名称。
func (c *PostgresChecker) Name() string { return "postgres" }

// Check 执行 PostgreSQL Ping（超时由 ctx 控制，调用方设置 3s）。
func (c *PostgresChecker) Check(ctx context.Context) error {
	if c.db == nil {
		return fmt.Errorf("postgres: not initialized")
	}
	if err := c.db.Ping(ctx); err != nil {
		return fmt.Errorf("postgres: ping failed: %w", err)
	}
	return nil
}

// --- Redis ---

// RedisChecker 通过 client.Ping 探测 Redis 连通性。
type RedisChecker struct {
	client *redis.Client
}

// NewRedisChecker 创建 Redis 连通性 checker。
func NewRedisChecker(client *redis.Client) *RedisChecker {
	return &RedisChecker{client: client}
}

// Name 返回 checker 名称。
func (c *RedisChecker) Name() string { return "redis" }

// Check 执行 Redis Ping（超时由 ctx 控制，调用方设置 3s）。
func (c *RedisChecker) Check(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("redis: not initialized")
	}
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: ping failed: %w", err)
	}
	return nil
}

// --- RabbitMQ ---

// RabbitMQChecker 探测 RabbitMQ channel 是否处于打开状态。
type RabbitMQChecker struct {
	// HealthReporter 返回底层连接/channel 是否健康。
	HealthReporter func() bool
}

// NewRabbitMQChecker 创建 RabbitMQ 连通性 checker。
func NewRabbitMQChecker(healthy func() bool) *RabbitMQChecker {
	return &RabbitMQChecker{HealthReporter: healthy}
}

// Name 返回 checker 名称。
func (c *RabbitMQChecker) Name() string { return "rabbitmq" }

// Check 检查 RabbitMQ channel 状态（超时由 ctx 控制，调用方设置 5s）。
func (c *RabbitMQChecker) Check(ctx context.Context) error {
	if c.HealthReporter == nil {
		return fmt.Errorf("rabbitmq: not configured")
	}
	// 使用 channel 避免阻塞：在 goroutine 中检查并通过 select 应用超时。
	resultCh := make(chan bool, 1)
	go func() {
		resultCh <- c.HealthReporter()
	}()
	select {
	case healthy := <-resultCh:
		if !healthy {
			return fmt.Errorf("rabbitmq: channel is closed")
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("rabbitmq: health check timed out: %w", ctx.Err())
	}
}

// --- Elasticsearch ---

// ElasticsearchChecker 通过 HTTP 请求探测 ES 集群健康。
type ElasticsearchChecker struct {
	// DoRequest 执行 HTTP 请求并返回 response；path 形如 "/_cluster/health"。
	DoRequest func(ctx context.Context, method, path string) (*http.Response, error)
}

// NewElasticsearchChecker 创建 Elasticsearch 连通性 checker。
func NewElasticsearchChecker(doRequest func(ctx context.Context, method, path string) (*http.Response, error)) *ElasticsearchChecker {
	return &ElasticsearchChecker{DoRequest: doRequest}
}

// Name 返回 checker 名称。
func (c *ElasticsearchChecker) Name() string { return "elasticsearch" }

// Check 调用 ES _cluster/health 端点（超时由 ctx 控制，调用方设置 5s）。
func (c *ElasticsearchChecker) Check(ctx context.Context) error {
	if c.DoRequest == nil {
		return fmt.Errorf("elasticsearch: not configured")
	}
	resp, err := c.DoRequest(ctx, "GET", "/_cluster/health?timeout=3s")
	if err != nil {
		return fmt.Errorf("elasticsearch: health request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("elasticsearch: cluster returned status %d", resp.StatusCode)
	}
	return nil
}

// --- SMTP ---

// SmtpChecker 通过 DialTimeout + StartTLS 握手探测 SMTP 连通性。
type SmtpChecker struct {
	Host   string
	Port   int
	UseTLS bool
}

// NewSmtpChecker 创建 SMTP 连通性 checker。
// 当 Host 为空时，Check 返回 nil（跳过）。
func NewSmtpChecker(host string, port int, useTLS bool) *SmtpChecker {
	return &SmtpChecker{Host: host, Port: port, UseTLS: useTLS}
}

// Name 返回 checker 名称。
func (c *SmtpChecker) Name() string { return "smtp" }

// Check 尝试建立 SMTP 连接（超时 3s）。
// SMTP 未配置（Host 为空）时跳过。
func (c *SmtpChecker) Check(ctx context.Context) error {
	if c.Host == "" {
		return nil // SMTP 未配置，跳过
	}
	addr := net.JoinHostPort(c.Host, fmt.Sprintf("%d", c.Port))

	var conn net.Conn
	var err error
	if c.UseTLS {
		conn, err = tls.DialWithDialer(
			&net.Dialer{Timeout: 3 * time.Second},
			"tcp", addr,
			&tls.Config{ServerName: hostnameOf(addr)},
		)
	} else {
		conn, err = net.DialTimeout("tcp", addr, 3*time.Second)
	}
	if err != nil {
		return fmt.Errorf("smtp: dial %s failed: %w", addr, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return fmt.Errorf("smtp: new client failed: %w", err)
	}
	defer client.Close()

	if !c.UseTLS {
		if err := client.StartTLS(&tls.Config{ServerName: c.Host}); err != nil {
			return fmt.Errorf("smtp: STARTTLS failed: %w", err)
		}
	}
	return nil
}

// --- OIDC IdP ---

// OIDCChecker 探测 OIDC IdP 的 .well-known/openid-configuration 端点。
type OIDCChecker struct {
	// IssuerURLs 是要探测的 OIDC IssuerURL 列表。
	IssuerURLs []string
	// HTTPClient 用于发起探测请求；若为 nil 则使用默认 5s 超时 client。
	HTTPClient *http.Client
}

// NewOIDCChecker 创建 OIDC IdP 连通性 checker。
func NewOIDCChecker(issuerURLs []string) *OIDCChecker {
	return &OIDCChecker{
		IssuerURLs: issuerURLs,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// Name 返回 checker 名称。
func (c *OIDCChecker) Name() string { return "oidc" }

// GET 探测结果（单个 issuer）。
type oidcCheckResult struct {
	issuer string
	err    error
}

// Check 尝试 GET 每个已启用 Provider 的 /.well-known/openid-configuration。
// 任一可达即视为 ok（200 即过，不 care body）。
// 所有 Provider 不可达时返回聚合错误。
func (c *OIDCChecker) Check(ctx context.Context) error {
	if len(c.IssuerURLs) == 0 {
		return nil // 无可检查则跳过
	}

	resultCh := make(chan oidcCheckResult, len(c.IssuerURLs))

	for _, issuer := range c.IssuerURLs {
		go func(iss string) {
			err := c.checkOne(ctx, iss)
			resultCh <- oidcCheckResult{issuer: iss, err: err}
		}(issuer)
	}

	var errs []error
	for i := 0; i < len(c.IssuerURLs); i++ {
		select {
		case r := <-resultCh:
			if r.err == nil {
				return nil // 任一可达即通过
			}
			errs = append(errs, fmt.Errorf("%s: %w", r.issuer, r.err))
		case <-ctx.Done():
			return fmt.Errorf("oidc: check timed out: %w", ctx.Err())
		}
	}
	return fmt.Errorf("oidc: all providers unreachable: %v", errs)
}

// checkOne 探测单个 issuer 的 .well-known/openid-configuration。
func (c *OIDCChecker) checkOne(ctx context.Context, issuer string) error {
	discoveryURL, err := url.JoinPath(issuer, ".well-known/openid-configuration")
	if err != nil {
		return fmt.Errorf("invalid issuer url %s: %w", issuer, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s failed: %w", discoveryURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s returned status %d", discoveryURL, resp.StatusCode)
	}
	return nil
}

// --- ReadyzHandler ---

// ReadyzHandler 聚合多个 ConnectivityChecker 的结果，输出 Kubernetes-style 响应。
// 内置 TTL 缓存避免探测过于频繁（默认 30s）。
type ReadyzHandler struct {
	checkers []ConnectivityChecker
	ttl      time.Duration

	mu        sync.Mutex
	cache     map[string]string // name -> "ok" or "error: ..."
	cachedAt  time.Time
	statusOK  bool
}

// NewReadyzHandler 创建带缓存的就绪探针 handler。
//
// 参数：
//   - checkers：连通性探测器列表；若 nil 则跳过该检查。
//   - ttl：缓存有效期；<=0 时使用默认 30s。
func NewReadyzHandler(checkers []ConnectivityChecker, ttl time.Duration) *ReadyzHandler {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return &ReadyzHandler{
		checkers: checkers,
		ttl:      ttl,
		cache:    make(map[string]string),
	}
}

// RunChecks 执行所有 checker 并缓存结果。
// 返回 checks map 和整体健康状态。
func (h *ReadyzHandler) RunChecks(ctx context.Context) (map[string]string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 缓存有效期内直接返回
	if len(h.cache) > 0 && time.Since(h.cachedAt) < h.ttl {
		return h.cache, h.statusOK
	}

	checks := make(map[string]string)
	allOK := true

	for _, checker := range h.checkers {
		if checker == nil {
			continue
		}
		checkCtx, cancel := context.WithTimeout(ctx, defaultTimeout(checker))
		err := checker.Check(checkCtx)
		cancel()

		if err != nil {
			checks[checker.Name()] = "error: " + err.Error()
			allOK = false
		} else {
			checks[checker.Name()] = "ok"
		}
	}

	h.cache = checks
	h.cachedAt = time.Now()
	h.statusOK = allOK
	return checks, allOK
}

// ResetCache 清除下次请求将重新探测的缓存。
func (h *ReadyzHandler) ResetCache() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cache = make(map[string]string)
	h.cachedAt = time.Time{}
	h.statusOK = false
}

// LiveResponse 是 /livez 端点的响应格式（极简，仅检查进程存活）。
type LiveResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Uptime    string `json:"uptime"`
}

// HealthResponse 是 /healthz 端点的深度健康检测（含错误率 / 连接池 / Outbox 堆积）。
type HealthResponse struct {
	Status      string            `json:"status"`
	Timestamp   string            `json:"timestamp"`
	Uptime      string            `json:"uptime"`
	ErrorRate   float64           `json:"error_rate_5m,omitempty"`
	Checks      map[string]string `json:"checks,omitempty"`
	PoolUtilization map[string]float64 `json:"pool_utilization,omitempty"`
}

// defaultTimeout 根据 checker 名称返回推荐超时。
func defaultTimeout(c ConnectivityChecker) time.Duration {
	switch c.Name() {
	case "postgres", "redis", "smtp":
		return 3 * time.Second
	case "rabbitmq", "elasticsearch", "oidc":
		return 5 * time.Second
	default:
		return 3 * time.Second
	}
}

func hostnameOf(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
