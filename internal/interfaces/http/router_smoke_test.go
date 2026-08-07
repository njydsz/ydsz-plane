// Package httpapi 测试：对引擎装配、安全头、指标端点与 Sprint 路由做冒烟验证。
package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/application/sprint"
	"github.com/njydsz/ydsz-plane/internal/application/version"
	"github.com/njydsz/ydsz-plane/internal/config"
)

// stubDeps 构造最小可用的 Deps（开发环境 + 无操作日志）。
func stubDeps() *Deps {
	return &Deps{
		Cfg: &config.Config{
			Server:  config.ServerConfig{Env: "development", Port: 8080},
			Auth:     config.AuthConfig{LoginRateLimitPer: 100},
			Features: config.FeatureFlags{RegistrationOpen: true},
		},
		Log: zap.NewNop(),
	}
}

// TestSecurityHeaders 验证中间件在每个响应上设置预期的纵深防御头。
func TestSecurityHeaders(t *testing.T) {
	r := NewEngine(stubDeps())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/healthz", nil)
	r.ServeHTTP(w, req)

	wantHeaders := map[string]string{
		"X-Content-Type-Options":      "nosniff",
		"X-Frame-Options":             "SAMEORIGIN",
		"Referrer-Policy":             "strict-origin-when-cross-origin",
		"Cross-Origin-Opener-Policy":  "same-origin",
		"Cross-Origin-Resource-Policy": "same-origin",
	}
	for hdr, want := range wantHeaders {
		if got := w.Header().Get(hdr); got != want {
			t.Errorf("header %q = %q, want %q", hdr, got, want)
		}
	}
	for _, hdr := range []string{"Content-Security-Policy", "Permissions-Policy"} {
		if w.Header().Get(hdr) == "" {
			t.Errorf("missing header %q", hdr)
		}
	}
}

// TestMetricsEndpoint 验证 /metrics 返回 200 且导出自定义 HTTP RED 指标，
// 不依赖任何外部依赖。
func TestMetricsEndpoint(t *testing.T) {
	r := NewEngine(stubDeps())

	// 先产生至少一个请求指标样本。
	r.ServeHTTP(httptest.NewRecorder(), mustReq("GET", "/healthz", ""))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, mustReq("GET", "/metrics", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want 200: %s", w.Code, w.Body.String())
	}
	// 活跃记录的指标（HTTP 计数器）在至少一个样本后出现。
	for _, needle := range []string{
		"ydsz_http_request_total",
		"ydsz_http_request_duration_ms",
	} {
		if !strings.Contains(w.Body.String(), needle) {
			t.Errorf("metrics body missing %q", needle)
		}
	}
	// promauto 收集器在首个标签设置后暴露 HELP/TYPE；对未使用的收集器，
	// 退而验证进程级收集器已导出。
	for _, needle := range []string{
		"go_goroutines",
		"process_resident_memory_bytes",
	} {
		if !strings.Contains(w.Body.String(), needle) {
			t.Errorf("metrics body missing %q", needle)
		}
	}
}

func mustReq(method, path, body string) *http.Request {
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// ==========================================================================
// Sprint 路由冒烟测试（Sprint 5.11 出口检查）
// ==========================================================================

// stubSprintDeps 返回装配了 SprintHandler 的 Deps。
func stubSprintDeps() *Deps {
	d := stubDeps()
	d.SprintHandler = sprint.NewHandler(nil)
	return d
}

// TestSprintRouteMounting 验证 Sprint 路由已挂载，未认证请求返回预期状态码（401）。
func TestSprintRouteMounting(t *testing.T) {
	r := NewEngine(stubSprintDeps())
	RegisterSprintRoutes(r, stubSprintDeps())

	tests := []struct {
		method string
		path   string
		want   int
	}{
		// 集合
		{"GET", "/api/v1/workspaces/1/projects/1/sprints", 401},
		{"POST", "/api/v1/workspaces/1/projects/1/sprints", 401},
		{"GET", "/api/v1/workspaces/1/projects/1/sprints/backlog", 401},
		{"GET", "/api/v1/workspaces/1/projects/1/sprints/suggest-capacity", 401},
		// 单资源
		{"GET", "/api/v1/workspaces/1/projects/1/sprints/1", 401},
		{"PATCH", "/api/v1/workspaces/1/projects/1/sprints/1", 401},
		{"DELETE", "/api/v1/workspaces/1/projects/1/sprints/1", 401},
		// 生命周期
		{"POST", "/api/v1/workspaces/1/projects/1/sprints/1/start", 401},
		{"POST", "/api/v1/workspaces/1/projects/1/sprints/1/complete", 401},
		// 进度 / 规划
		{"GET", "/api/v1/workspaces/1/projects/1/sprints/1/progress", 401},
		{"GET", "/api/v1/workspaces/1/projects/1/sprints/1/issues", 401},
		{"POST", "/api/v1/workspaces/1/projects/1/sprints/1/issues", 401},
		{"DELETE", "/api/v1/workspaces/1/projects/1/sprints/1/issues/1", 401},
		// 燃尽图 / 复盘
		{"GET", "/api/v1/workspaces/1/projects/1/sprints/1/burndown", 401},
		{"GET", "/api/v1/workspaces/1/projects/1/sprints/1/review", 401},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := mustReq(tc.method, tc.path, "")
			r.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Errorf("status = %d, want %d", w.Code, tc.want)
			}
		})
	}
}

// TestSprintNotWiredWithoutHandler 验证未设置 SprintHandler 时，
// Sprint 路由不应被挂载（RegisterSprintRoutes 提前返回）。
func TestSprintNotWiredWithoutHandler(t *testing.T) {
	r := NewEngine(stubDeps())
	// 未设置 SprintHandler
	RegisterSprintRoutes(r, stubDeps())

	w := httptest.NewRecorder()
	req := mustReq("GET", "/api/v1/workspaces/1/projects/1/sprints", "")
	r.ServeHTTP(w, req)
	// 没有 SprintHandler 时，/sprints 应返回 404（而非 401），
	// 因为路由组从未注册。
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (routes not mounted)", w.Code)
	}
}

// ==========================================================================
// Version 路由冒烟测试（Sprint 6 出口检查）
// ==========================================================================

// stubVersionDeps 返回装配了 VersionHandler 的 Deps。
func stubVersionDeps() *Deps {
	d := stubDeps()
	d.VersionHandler = version.NewHandler(nil)
	return d
}

// TestVersionRouteMounting 验证 Version 路由已挂载，未认证请求返回 401。
func TestVersionRouteMounting(t *testing.T) {
	r := NewEngine(stubVersionDeps())
	RegisterVersionRoutes(r, stubVersionDeps())

	tests := []struct {
		method string
		path   string
		want   int
	}{
		// 集合
		{"GET", "/api/v1/workspaces/1/projects/1/versions", 401},
		{"POST", "/api/v1/workspaces/1/projects/1/versions", 401},
		{"GET", "/api/v1/workspaces/1/projects/1/versions/defects", 401},
		// 单资源
		{"GET", "/api/v1/workspaces/1/projects/1/versions/1", 401},
		{"PATCH", "/api/v1/workspaces/1/projects/1/versions/1", 401},
		{"DELETE", "/api/v1/workspaces/1/projects/1/versions/1", 401},
		// 生命周期
		{"POST", "/api/v1/workspaces/1/projects/1/versions/1/activate", 401},
		{"POST", "/api/v1/workspaces/1/projects/1/versions/1/release", 401},
		{"POST", "/api/v1/workspaces/1/projects/1/versions/1/archive", 401},
		// 进度 / 质量 / 交付
		{"GET", "/api/v1/workspaces/1/projects/1/versions/1/progress", 401},
		{"GET", "/api/v1/workspaces/1/projects/1/versions/1/quality", 401},
		{"GET", "/api/v1/workspaces/1/projects/1/versions/1/delivery-report", 401},
		// Release Notes
		{"GET", "/api/v1/workspaces/1/projects/1/versions/1/release-notes", 401},
		{"POST", "/api/v1/workspaces/1/projects/1/versions/1/release-notes/regenerate", 401},
		// 缺陷面板
		{"GET", "/api/v1/workspaces/1/projects/1/versions/1/defects", 401},
		// 迭代聚合
		{"GET", "/api/v1/workspaces/1/projects/1/versions/1/sprints", 401},
		{"POST", "/api/v1/workspaces/1/projects/1/versions/1/sprints", 401},
		{"DELETE", "/api/v1/workspaces/1/projects/1/versions/1/sprints/1", 401},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := mustReq(tc.method, tc.path, "")
			r.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Errorf("status = %d, want %d", w.Code, tc.want)
			}
		})
	}
}

// TestVersionNotWiredWithoutHandler 验证未设置 VersionHandler 时路由不会被注册。
func TestVersionNotWiredWithoutHandler(t *testing.T) {
	r := NewEngine(stubDeps())
	RegisterVersionRoutes(r, stubDeps())

	w := httptest.NewRecorder()
	req := mustReq("GET", "/api/v1/workspaces/1/projects/1/versions", "")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (routes not mounted)", w.Code)
	}
}

// TestVersionLifecycleRoutePatterns 验证全生命周期端点路径模式正确。
func TestVersionLifecycleRoutePatterns(t *testing.T) {
	// 状态流转可用端点
	lifecyclePaths := []string{
		"/api/v1/workspaces/1/projects/1/versions/1/activate",
		"/api/v1/workspaces/1/projects/1/versions/1/release",
		"/api/v1/workspaces/1/projects/1/versions/1/archive",
	}

	r := NewEngine(stubVersionDeps())
	RegisterVersionRoutes(r, stubVersionDeps())

	for _, path := range lifecyclePaths {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := mustReq("POST", path, "")
			r.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s: status = %d, want 401 (auth required)", path, w.Code)
			}
		})
	}
}

// TestVersionQueryEndpoints 验证所有查询端点均需要认证。
func TestVersionQueryEndpoints(t *testing.T) {
	r := NewEngine(stubVersionDeps())
	RegisterVersionRoutes(r, stubVersionDeps())

	queryPaths := []string{
		"/api/v1/workspaces/1/projects/1/versions/1/progress",
		"/api/v1/workspaces/1/projects/1/versions/1/quality",
		"/api/v1/workspaces/1/projects/1/versions/1/delivery-report",
		"/api/v1/workspaces/1/projects/1/versions/1/release-notes",
		"/api/v1/workspaces/1/projects/1/versions/1/defects",
		"/api/v1/workspaces/1/projects/1/versions/1/sprints",
	}

	for _, path := range queryPaths {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := mustReq("GET", path, "")
			r.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s: status = %d, want 401", path, w.Code)
			}
		})
	}
}
