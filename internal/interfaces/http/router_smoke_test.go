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
