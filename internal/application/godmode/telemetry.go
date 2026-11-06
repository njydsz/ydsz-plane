// Package godmode — 遥测上报。
//
// 在启用时定期上报匿名实例指标（版本、工作空间数、用户数、功能开关），
// 帮助改进产品。上报内容不含任何业务数据。
package godmode

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// TelemetryReporter 遥测上报器。
type TelemetryReporter struct {
	svc      *Service
	endpoint string
	interval time.Duration
	client   *http.Client
	stopCh   chan struct{}
}

// NewTelemetryReporter 构造上报器。
func NewTelemetryReporter(svc *Service, endpoint string, interval time.Duration) *TelemetryReporter {
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	return &TelemetryReporter{
		svc:      svc,
		endpoint: endpoint,
		interval: interval,
		client:   &http.Client{Timeout: 15 * time.Second},
		stopCh:   make(chan struct{}),
	}
}

// Start 启动定期上报。
func (r *TelemetryReporter) Start(ctx context.Context) {
	if r.endpoint == "" {
		return
	}
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-r.stopCh:
				return
			case <-ticker.C:
				_ = r.Report(ctx)
			}
		}
	}()
}

// Stop 停止上报。
func (r *TelemetryReporter) Stop() { close(r.stopCh) }

// Report 立即上报一次。
func (r *TelemetryReporter) Report(ctx context.Context) error {
	if r.endpoint == "" {
		return nil
	}
	settings := r.svc.GetSettings()
	if !settings.Telemetry.Enabled {
		return nil
	}

	payload := map[string]any{
		"instance_id": settings.Telemetry.InstanceID,
		"version":     r.svc.GetStatus().Version,
		"edition":     r.svc.GetStatus().Edition,
		"features":    r.svc.GetStatus().Features,
		"reported_at": time.Now().UTC().Format(time.RFC3339),
	}

	// 附加用量（若可获取）
	if stats, err := r.svc.GetUsage(ctx); err == nil {
		payload["usage"] = stats
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ydsz-plane-telemetry")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(io.LimitReader(resp.Body, 64<<10))

	if resp.StatusCode >= 400 {
		return &TelemetryError{StatusCode: resp.StatusCode}
	}
	return nil
}

// TelemetryError 上报失败。
type TelemetryError struct {
	StatusCode int
}

func (e *TelemetryError) Error() string {
	return "telemetry: server returned " + itoa(e.StatusCode)
}
