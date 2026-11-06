// Package godmode — HTTP handlers。
//
// 路由：/api/v1/admin/*（需实例管理员权限）。
package godmode

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
)

// HandlerDeps handler 依赖。
type HandlerDeps struct {
	Svc *Service
	// Telemetry 可选的遥测上报器。
	Telemetry *TelemetryReporter
}

// Handler Gin handler 集合。
type Handler struct {
	d *HandlerDeps
}

// NewHandler 构造 handler。
func NewHandler(d *HandlerDeps) *Handler { return &Handler{d: d} }

// Register 注册路由（调用方已挂载实例管理员中间件）。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/admin")
	g.GET("/settings", h.GetSettings)
	g.PUT("/settings", h.UpdateSettings)
	g.POST("/settings/reset", h.ResetSettings)
	g.GET("/license", h.GetLicense)
	g.GET("/status", h.GetStatus)
	g.GET("/usage", h.GetUsage)
	g.POST("/test-smtp", h.TestSMTP)
	g.POST("/telemetry/report", h.ReportTelemetry)
}

// IsInstanceAdmin 实例管理员判定。
//
// 判定规则（任一满足）：
//   - 请求头 X-Instance-Admin-Key 与环境变量 YDSZ_INSTANCE_ADMIN_KEY 匹配；
//   - 已登录用户在实例管理员列表中（YDSZ_INSTANCE_ADMINS，逗号分隔的用户 ID）。
func IsInstanceAdmin(c *gin.Context) bool {
	// 方式一：专用管理密钥
	if expected := os.Getenv("YDSZ_INSTANCE_ADMIN_KEY"); expected != "" {
		provided := c.GetHeader("X-Instance-Admin-Key")
		if provided != "" && provided == expected {
			return true
		}
	}
	// 方式二：用户 ID 白名单
	allowed := os.Getenv("YDSZ_INSTANCE_ADMINS")
	if allowed == "" {
		return false
	}
	uid := c.GetInt64(middleware.CtxUserID)
	if uid <= 0 {
		return false
	}
	for _, part := range strings.Split(allowed, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		var id int64
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				id = 0
				break
			}
			id = id*10 + int64(ch-'0')
		}
		if id == uid {
			return true
		}
	}
	return false
}

// --- handlers ---

// GetSettings godoc
//
//	@Summary		实例设置
//	@Tags			god-mode
//	@Produce		json
//	@Success		200	{object}	InstanceSettings
//	@Router			/admin/settings [get]
func (h *Handler) GetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, h.d.Svc.GetSettings())
}

// UpdateSettings godoc
//
//	@Summary		更新实例设置
//	@Tags			god-mode
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	InstanceSettings
//	@Router			/admin/settings [put]
func (h *Handler) UpdateSettings(c *gin.Context) {
	var body UpdateSettingsInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.UpdatedBy = c.GetInt64(middleware.CtxUserID)
	out, err := h.d.Svc.UpdateSettings(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// ResetSettings godoc
//
//	@Summary		重置实例设置
//	@Tags			god-mode
//	@Produce		json
//	@Success		200	{object}	InstanceSettings
//	@Router			/admin/settings/reset [post]
func (h *Handler) ResetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, h.d.Svc.ResetSettings())
}

// GetLicense godoc
//
//	@Summary		许可证信息
//	@Tags			god-mode
//	@Produce		json
//	@Success		200	{object}	LicenseInfo
//	@Router			/admin/license [get]
func (h *Handler) GetLicense(c *gin.Context) {
	c.JSON(http.StatusOK, h.d.Svc.GetLicense())
}

// GetStatus godoc
//
//	@Summary		系统状态
//	@Tags			god-mode
//	@Produce		json
//	@Success		200	{object}	SystemStatus
//	@Router			/admin/status [get]
func (h *Handler) GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, h.d.Svc.GetStatus())
}

// GetUsage godoc
//
//	@Summary		用量统计
//	@Tags			god-mode
//	@Produce		json
//	@Success		200	{object}	UsageStats
//	@Router			/admin/usage [get]
func (h *Handler) GetUsage(c *gin.Context) {
	stats, err := h.d.Svc.GetUsage(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// TestSMTP godoc
//
//	@Summary		测试 SMTP 配置
//	@Tags			god-mode
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/admin/test-smtp [post]
func (h *Handler) TestSMTP(c *gin.Context) {
	var body struct {
		To string `json:"to"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.To == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供收件人邮箱"})
		return
	}
	settings := h.d.Svc.GetSettings()
	if !settings.SMTP.Enabled || settings.SMTP.Host == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"ok": false, "message": "SMTP 未配置或未启用"})
		return
	}

	msg := MailMessage{
		From:    settings.SMTP.FromEmail,
		To:      []string{body.To},
		Subject: "[Ydsz Plane] SMTP 配置测试",
		Body:    "这是一封来自 Ydsz Plane 实例管理面板的测试邮件。\n\n如果你收到此邮件，说明 SMTP 配置正确。",
		HTML: "<p>这是一封来自 <strong>Ydsz Plane</strong> 实例管理面板的测试邮件。</p>" +
			"<p>如果你收到此邮件，说明 SMTP 配置正确。</p>",
	}
	if err := SendEmail(settings.SMTP, msg); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"ok":      false,
			"message": "SMTP 发送失败：" + err.Error(),
			"config": gin.H{
				"host":       settings.SMTP.Host,
				"port":       settings.SMTP.Port,
				"use_tls":    settings.SMTP.UseTLS,
				"from_email": settings.SMTP.FromEmail,
			},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "测试邮件已发送，请查收",
		"config": gin.H{
			"host":       settings.SMTP.Host,
			"port":       settings.SMTP.Port,
			"use_tls":    settings.SMTP.UseTLS,
			"from_email": settings.SMTP.FromEmail,
		},
	})
}

// ReportTelemetry godoc
//
//	@Summary		立即上报遥测
//	@Tags			god-mode
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/admin/telemetry/report [post]
func (h *Handler) ReportTelemetry(c *gin.Context) {
	settings := h.d.Svc.GetSettings()
	if !settings.Telemetry.Enabled {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"ok": false, "message": "遥测未启用"})
		return
	}
	if h.d.Telemetry == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"ok": false, "message": "遥测上报器未配置"})
		return
	}
	if err := h.d.Telemetry.Report(c.Request.Context()); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "message": "上报失败：" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "遥测已上报"})
}
