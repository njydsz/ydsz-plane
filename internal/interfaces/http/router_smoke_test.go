package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/application/sprint"
	"github.com/njydsz/ydsz-plane/internal/config"
)

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

// TestSecurityHeaders verifies that the middleware sets the expected
// defense-in-depth headers on every response.
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

// TestMetricsEndpoint proves that /metrics responds 200 and exports our
// custom HTTP RED metrics without needing any external dependency.
func TestMetricsEndpoint(t *testing.T) {
	r := NewEngine(stubDeps())

	// Generate at least one request metric sample.
	r.ServeHTTP(httptest.NewRecorder(), mustReq("GET", "/healthz", ""))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, mustReq("GET", "/metrics", ""))

	if w.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want 200: %s", w.Code, w.Body.String())
	}
	// Actively-recorded metrics (HTTP counters) show after at least one sample.
	for _, needle := range []string{
		"ydsz_http_request_total",
		"ydsz_http_request_duration_ms",
	} {
		if !strings.Contains(w.Body.String(), needle) {
			t.Errorf("metrics body missing %q", needle)
		}
	}
	// promauto collectors expose HELP/TYPE once first label is set; for unused
	// ones we fall back to verifying process-level collectors are exported.
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

// stubSprintDeps returns a Deps with SprintHandler wired.
func stubSprintDeps() *Deps {
	d := stubDeps()
	d.SprintHandler = sprint.NewHandler(nil)
	return d
}

// TestSprintRouteMounting verifies Sprint routes are mounted and return
// expected HTTP status codes for unauthenticated requests (401).
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

// TestSprintNotWiredWithoutHandler verifies that without SprintHandler,
// sprint routes should NOT be mounted (RegisterSprintRoutes returns early).
func TestSprintNotWiredWithoutHandler(t *testing.T) {
	r := NewEngine(stubDeps())
	// No SprintHandler set
	RegisterSprintRoutes(r, stubDeps())

	w := httptest.NewRecorder()
	req := mustReq("GET", "/api/v1/workspaces/1/projects/1/sprints", "")
	r.ServeHTTP(w, req)
	// Without SprintHandler, /sprints should 404 (not 401), because the route
	// group was never registered.
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (routes not mounted)", w.Code)
	}
}
