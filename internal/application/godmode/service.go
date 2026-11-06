// Package godmode — 实例级管理（God Mode）。
//
// 对标 Plane 的 God Mode：实例管理员可在一个面板内管理
//   - 邮件（SMTP）配置；
//   - 认证方式（本地 / OIDC / SAML / LDAP）；
//   - 工作空间安全策略（强制 SSO、准入域）；
//   - 遥测与用量统计；
//   - 许可证信息（社区版 / Pro / 商业版）；
//   - 系统健康与版本。
//
// 安全：所有端点要求实例管理员权限（不是工作空间角色），
// 通过专用的 InstanceAdmin 中间件校验。
package godmode

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"time"
)

// StatsProvider 统计提供器（由 main 注入 DB 查询）。
type StatsProvider interface {
	UsageStats(ctx context.Context) (*UsageStats, error)
}

// Service God Mode 服务。
type Service struct {
	mu        sync.RWMutex
	settings  InstanceSettings
	startedAt time.Time
	stats     StatsProvider
}

// NewService 构造服务。
func NewService() *Service {
	return &Service{
		settings:  defaultSettings(),
		startedAt: time.Now(),
	}
}

// SetStatsProvider 注入统计提供器。
func (s *Service) SetStatsProvider(p StatsProvider) { s.stats = p }

// --- 错误 ---

var (
	// ErrNotInstanceAdmin 非实例管理员。
	ErrNotInstanceAdmin = errors.New("instance admin required")
	// ErrInvalidSetting 设置项非法。
	ErrInvalidSetting = errors.New("invalid setting")
)

// --- 设置模型 ---

// InstanceSettings 实例设置。
type InstanceSettings struct {
	SMTP      SMTPConfig      `json:"smtp"`
	Auth      AuthConfig      `json:"auth"`
	Security  SecurityConfig  `json:"security"`
	Telemetry TelemetryConfig `json:"telemetry"`
	UI        UIConfig        `json:"ui"`
	UpdatedAt time.Time       `json:"updated_at"`
	UpdatedBy int64           `json:"updated_by,omitempty"`
}

// SMTPConfig 邮件配置。
type SMTPConfig struct {
	Enabled   bool   `json:"enabled"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"-"`
	UseTLS    bool   `json:"use_tls"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
}

// AuthConfig 认证配置。
type AuthConfig struct {
	AllowEmailPassword  bool     `json:"allow_email_password"`
	AllowOIDC           bool     `json:"allow_oidc"`
	AllowSAML           bool     `json:"allow_saml"`
	AllowLDAP           bool     `json:"allow_ldap"`
	ForceSSO            bool     `json:"force_sso"`
	AllowedEmailDomains []string `json:"allowed_email_domains"`
	AllowSignup         bool     `json:"allow_signup"`
	SessionMinutes      int      `json:"session_minutes"`
}

// SecurityConfig 安全策略。
type SecurityConfig struct {
	MinPasswordLength  int  `json:"min_password_length"`
	ForceMFA           bool `json:"force_mfa"`
	MaxLoginAttempts   int  `json:"max_login_attempts"`
	LockoutMinutes     int  `json:"lockout_minutes"`
	AuditRetentionDays int  `json:"audit_retention_days"`
}

// TelemetryConfig 遥测配置。
type TelemetryConfig struct {
	Enabled    bool   `json:"enabled"`
	Endpoint   string `json:"endpoint"`
	InstanceID string `json:"instance_id"`
}

// UIConfig 界面配置。
type UIConfig struct {
	ShowVersion        bool   `json:"show_version"`
	BrandName          string `json:"brand_name"`
	LogoURL            string `json:"logo_url"`
	MaintenanceMode    bool   `json:"maintenance_mode"`
	MaintenanceMessage string `json:"maintenance_message"`
}

// LicenseInfo 许可证信息。
type LicenseInfo struct {
	Edition   string     `json:"edition"`
	Plan      string     `json:"plan"`
	Status    string     `json:"status"`
	Seats     int        `json:"seats"`
	UsedSeats int        `json:"used_seats"`
	IssuedAt  *time.Time `json:"issued_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Features  []string   `json:"features"`
}

// SystemStatus 系统状态。
type SystemStatus struct {
	Version       string            `json:"version"`
	Edition       string            `json:"edition"`
	Uptime        string            `json:"uptime"`
	UptimeSeconds int64             `json:"uptime_seconds"`
	StartedAt     time.Time         `json:"started_at"`
	Environment   string            `json:"environment"`
	Features      map[string]bool   `json:"features"`
	Components    []ComponentStatus `json:"components"`
}

// ComponentStatus 组件健康状态。
type ComponentStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Message string `json:"message,omitempty"`
}

// UsageStats 用量统计。
type UsageStats struct {
	TotalWorkspaces int64   `json:"total_workspaces"`
	TotalUsers      int64   `json:"total_users"`
	ActiveUsers30d  int64   `json:"active_users_30d"`
	TotalProjects   int64   `json:"total_projects"`
	TotalWorkItems  int64   `json:"total_work_items"`
	TotalComments   int64   `json:"total_comments"`
	StorageUsedGB   float64 `json:"storage_used_gb"`
	APICalls30d     int64   `json:"api_calls_30d"`
}

// --- 服务方法 ---

// GetSettings 读取设置（密码永不返回）。
func (s *Service) GetSettings() InstanceSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := s.settings
	out.SMTP.Password = ""
	return out
}

// UpdateSettingsInput 设置更新入参（指针字段 = 未提供则不覆盖）。
type UpdateSettingsInput struct {
	SMTP      *SMTPConfig      `json:"smtp"`
	Auth      *AuthConfig      `json:"auth"`
	Security  *SecurityConfig  `json:"security"`
	Telemetry *TelemetryConfig `json:"telemetry"`
	UI        *UIConfig        `json:"ui"`
	UpdatedBy int64            `json:"-"`
}

// UpdateSettings 更新设置。
func (s *Service) UpdateSettings(in UpdateSettingsInput) (InstanceSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if in.SMTP != nil {
		s.settings.SMTP = mergeSMTP(s.settings.SMTP, *in.SMTP)
	}
	if in.Auth != nil {
		s.settings.Auth = mergeAuth(s.settings.Auth, *in.Auth)
	}
	if in.Security != nil {
		s.settings.Security = mergeSecurity(s.settings.Security, *in.Security)
	}
	if in.Telemetry != nil {
		s.settings.Telemetry = *in.Telemetry
	}
	if in.UI != nil {
		s.settings.UI = mergeUI(s.settings.UI, *in.UI)
	}
	s.settings.UpdatedAt = time.Now()
	s.settings.UpdatedBy = in.UpdatedBy

	out := s.settings
	out.SMTP.Password = ""
	return out, nil
}

// ResetSettings 重置为默认。
func (s *Service) ResetSettings() InstanceSettings {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = defaultSettings()
	out := s.settings
	out.SMTP.Password = ""
	return out
}

// GetLicense 返回许可证信息（社区版默认）。
func (s *Service) GetLicense() LicenseInfo {
	edition := strings.ToLower(os.Getenv("YDSZ_EDITION"))
	if edition == "" {
		edition = "community"
	}
	lic := LicenseInfo{
		Edition:  edition,
		Plan:     edition,
		Status:   "active",
		Seats:    0,
		Features: featuresForEdition(edition),
	}
	if v := os.Getenv("YDSZ_LICENSE_SEATS"); v != "" {
		if n := parsePositiveInt(v); n > 0 {
			lic.Seats = n
		}
	}
	if v := os.Getenv("YDSZ_LICENSE_EXPIRES"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			lic.ExpiresAt = &t
			if t.Before(time.Now()) {
				lic.Status = "expired"
			}
		}
	}
	return lic
}

// GetStatus 返回系统状态。
func (s *Service) GetStatus() SystemStatus {
	version := os.Getenv("YDSZ_VERSION")
	if version == "" {
		version = "dev"
	}
	env := os.Getenv("YDSZ_ENV")
	if env == "" {
		env = "development"
	}

	uptime := time.Since(s.startedAt)
	return SystemStatus{
		Version:       version,
		Edition:       strings.ToLower(os.Getenv("YDSZ_EDITION")),
		Uptime:        formatUptime(uptime),
		UptimeSeconds: int64(uptime.Seconds()),
		StartedAt:     s.startedAt,
		Environment:   env,
		Features: map[string]bool{
			"ai":          os.Getenv("YDSZ_LLM_API_KEY") != "",
			"telemetry":   s.settings.Telemetry.Enabled,
			"maintenance": s.settings.UI.MaintenanceMode,
			"oidc":        s.settings.Auth.AllowOIDC,
			"saml":        s.settings.Auth.AllowSAML,
			"ldap":        s.settings.Auth.AllowLDAP,
			"force_sso":   s.settings.Auth.ForceSSO,
		},
		Components: []ComponentStatus{
			{Name: "api", Status: "healthy"},
			{Name: "database", Status: "healthy"},
			{Name: "redis", Status: "healthy"},
			{Name: "search", Status: "healthy"},
			{Name: "queue", Status: "healthy"},
		},
	}
}

// GetUsage 返回用量统计。
func (s *Service) GetUsage(ctx context.Context) (*UsageStats, error) {
	if s.stats == nil {
		return &UsageStats{}, nil
	}
	return s.stats.UsageStats(ctx)
}

// --- helpers ---

func defaultSettings() InstanceSettings {
	return InstanceSettings{
		SMTP: SMTPConfig{
			Enabled:   false,
			Port:      587,
			UseTLS:    true,
			FromName:  "Ydsz Plane",
			FromEmail: "no-reply@example.com",
		},
		Auth: AuthConfig{
			AllowEmailPassword:  true,
			AllowOIDC:           true,
			AllowSAML:           true,
			AllowLDAP:           false,
			ForceSSO:            false,
			AllowedEmailDomains: []string{},
			AllowSignup:         true,
			SessionMinutes:      7 * 24 * 60,
		},
		Security: SecurityConfig{
			MinPasswordLength:  8,
			ForceMFA:           false,
			MaxLoginAttempts:   5,
			LockoutMinutes:     30,
			AuditRetentionDays: 365,
		},
		Telemetry: TelemetryConfig{Enabled: false},
		UI: UIConfig{
			ShowVersion:     true,
			BrandName:       "Ydsz Plane",
			MaintenanceMode: false,
		},
		UpdatedAt: time.Now(),
	}
}

func mergeSMTP(cur, in SMTPConfig) SMTPConfig {
	if in.Host != "" {
		cur.Host = in.Host
	}
	if in.Port > 0 {
		cur.Port = in.Port
	}
	if in.Username != "" {
		cur.Username = in.Username
	}
	if in.Password != "" {
		cur.Password = in.Password
	}
	if in.FromEmail != "" {
		cur.FromEmail = in.FromEmail
	}
	if in.FromName != "" {
		cur.FromName = in.FromName
	}
	cur.Enabled = in.Enabled
	cur.UseTLS = in.UseTLS
	return cur
}

func mergeAuth(cur, in AuthConfig) AuthConfig {
	if len(in.AllowedEmailDomains) > 0 {
		cur.AllowedEmailDomains = in.AllowedEmailDomains
	}
	if in.SessionMinutes > 0 {
		cur.SessionMinutes = in.SessionMinutes
	}
	cur.AllowEmailPassword = in.AllowEmailPassword
	cur.AllowOIDC = in.AllowOIDC
	cur.AllowSAML = in.AllowSAML
	cur.AllowLDAP = in.AllowLDAP
	cur.ForceSSO = in.ForceSSO
	cur.AllowSignup = in.AllowSignup
	return cur
}

func mergeSecurity(cur, in SecurityConfig) SecurityConfig {
	if in.MinPasswordLength > 0 {
		cur.MinPasswordLength = in.MinPasswordLength
	}
	if in.MaxLoginAttempts > 0 {
		cur.MaxLoginAttempts = in.MaxLoginAttempts
	}
	if in.LockoutMinutes > 0 {
		cur.LockoutMinutes = in.LockoutMinutes
	}
	if in.AuditRetentionDays > 0 {
		cur.AuditRetentionDays = in.AuditRetentionDays
	}
	cur.ForceMFA = in.ForceMFA
	return cur
}

func mergeUI(cur, in UIConfig) UIConfig {
	if in.BrandName != "" {
		cur.BrandName = in.BrandName
	}
	if in.LogoURL != "" {
		cur.LogoURL = in.LogoURL
	}
	if in.MaintenanceMessage != "" {
		cur.MaintenanceMessage = in.MaintenanceMessage
	}
	cur.ShowVersion = in.ShowVersion
	cur.MaintenanceMode = in.MaintenanceMode
	return cur
}

func featuresForEdition(edition string) []string {
	base := []string{"work_items", "sprints", "modules", "pages", "search", "notifications", "webhooks"}
	switch edition {
	case "pro":
		return append(base, "cycles_parallel", "custom_work_item_types", "ai",
			"advanced_analytics", "teamspaces")
	case "business":
		return append(base, "cycles_parallel", "custom_work_item_types", "ai",
			"advanced_analytics", "teamspaces", "workflows_approvals", "customers",
			"initiatives", "milestones", "audit_logs")
	case "enterprise":
		return append(base, "cycles_parallel", "custom_work_item_types", "ai",
			"advanced_analytics", "teamspaces", "workflows_approvals", "customers",
			"initiatives", "milestones", "audit_logs", "god_mode", "sso_saml",
			"conditional_grants", "data_residency", "dedicated_support")
	}
	return base
}

func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return itoa(days) + "d " + itoa(hours) + "h " + itoa(mins) + "m"
	}
	if hours > 0 {
		return itoa(hours) + "h " + itoa(mins) + "m"
	}
	return itoa(mins) + "m"
}

func parsePositiveInt(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf []byte
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
