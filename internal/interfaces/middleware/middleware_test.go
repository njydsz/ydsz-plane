// Package middleware 中间件纯逻辑与安全头测试。
package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/config"
)

// TestParseBigInt 验证 parseBigInt 对合法数字、空串、非法字符与溢出的处理。
func TestParseBigInt(t *testing.T) {
	tests := []struct {
		in      string
		wantVal int64
		wantOK  bool
	}{
		{"0", 0, true},
		{"42", 42, true},
		{"9223372036854775807", 9223372036854775807, true}, // MaxInt64
		{"", 0, false},
		{"abc", 0, false},
		{"12a", 0, false},
		{"-1", 0, false},
		{"9223372036854775808", 0, false}, // MaxInt64+1 溢出
		{"99999999999999999999", 0, false},
	}
	for _, tc := range tests {
		gotVal, gotOK := parseBigInt(tc.in)
		if gotVal != tc.wantVal || gotOK != tc.wantOK {
			t.Errorf("parseBigInt(%q) = (%d, %v), want (%d, %v)",
				tc.in, gotVal, gotOK, tc.wantVal, tc.wantOK)
		}
	}
}

// TestBearerToken 验证凭证提取优先级：Bearer 头 > X-Api-Key > Cookie。
func TestBearerToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
		apiKey string
		cookie string
		want   string
	}{
		{"bearer header", "Bearer tok1", "", "", "tok1"},
		{"lowercase not matched", "bearer tok", "", "", ""}, // 区分大小写，返回空
		{"api key", "", "key2", "", "key2"},
		{"cookie", "", "", "ck3", "ck3"},
		{"priority bearer over api key", "Bearer tokA", "keyB", "", "tokA"},
		{"priority api key over cookie", "", "keyC", "ckD", "keyC"},
		{"empty", "", "", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.GET("/", func(c *gin.Context) {
				c.String(http.StatusOK, bearerToken(c))
			})
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			if tc.apiKey != "" {
				req.Header.Set("X-Api-Key", tc.apiKey)
			}
			if tc.cookie != "" {
				req.AddCookie(&http.Cookie{Name: "ydsz_access", Value: tc.cookie})
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Body.String() != tc.want {
				t.Errorf("bearerToken = %q, want %q", w.Body.String(), tc.want)
			}
		})
	}
}

// TestSecurityHeaders 验证安全响应头均被设置。
func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders(&config.SecurityConfig{TLSEnabled: false}))
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	for _, header := range []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
		"Permissions-Policy",
		"Cross-Origin-Opener-Policy",
		"Cross-Origin-Resource-Policy",
		"Content-Security-Policy",
	} {
		if w.Header().Get(header) == "" {
			t.Errorf("missing security header %s", header)
		}
	}
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("X-Content-Type-Options = %s", w.Header().Get("X-Content-Type-Options"))
	}
	// HSTS 默认不启用。
	if w.Header().Get("Strict-Transport-Security") != "" {
		t.Errorf("HSTS should not be set when tls_enabled=false")
	}
}

// TestSecurityHeaders_HSTS 验证 TLSEnabled=true 时 HSTS 头设置。
func TestSecurityHeaders_HSTS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders(&config.SecurityConfig{TLSEnabled: true}))
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	want := "max-age=63072000; includeSubDomains; preload"
	if got := w.Header().Get("Strict-Transport-Security"); got != want {
		t.Errorf("Strict-Transport-Security = %q, want %q", got, want)
	}
}

// TestSecurityHeaders_OAC 验证 Origin-Agent-Cluster 头存在。
func TestSecurityHeaders_OAC(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders(&config.SecurityConfig{}))
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Origin-Agent-Cluster"); got != "?1" {
		t.Errorf("Origin-Agent-Cluster = %q, want \"?1\"", got)
	}
}

// TestCSPNonce 验证 CSP 头里的 script-src nonce 片段与 ctx 中的 nonce。
func TestCSPNonce(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders(&config.SecurityConfig{}))
	r.Use(CSPNonce())
	r.GET("/", func(c *gin.Context) {
		nonce, ok := c.Get(CSPNonceKey)
		if !ok || nonce.(string) == "" {
			t.Errorf("csp-nonce not set in ctx")
		}
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	csp := w.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "script-src 'nonce-") {
		t.Errorf("CSP missing nonce: %q", csp)
	}
	if !strings.Contains(csp, "'strict-dynamic'") {
		t.Errorf("CSP missing strict-dynamic: %q", csp)
	}
}

// TestCORS_DefaultOrigins 验证：dev 默认白名单 + 非匹配 origin 不会返回 ACAO。
func TestCORS_DefaultOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	log := zap.NewNop()
	def := []string{"http://localhost:5173"}
	r.Use(CORS(def, config.CORSConfig{AllowCredentials: true}, log))
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	// 匹配的 origin
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("ACAO = %q, want matched origin", got)
	}
	if got := w.Header().Get("Vary"); !strings.Contains(got, "Origin") {
		t.Errorf("Vary = %q, want contains Origin", got)
	}

	// 不匹配的 origin
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Origin", "https://evil.example.com")
	r.ServeHTTP(w2, req2)

	if got := w2.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("non-matched origin should not receive ACAO, got %q", got)
	}
}

// TestCORS_WildcardDisablesCredentials 验证 wildcard + credentials 时，运行时降级禁用 credentials。
func TestCORS_WildcardDisablesCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// 配置层面已拒绝 wildcard+credentials，此处模拟运行期再保险路径：
	// 传入 wildcard origin + credentials=true，应自动降级 credentials=false。
	r := gin.New()
	r.Use(CORS([]string{}, config.CORSConfig{
		AllowedOrigins:   "*",
		AllowCredentials: true,
	}, zap.NewNop()))
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://any.example.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("wildcard CORS should never send ACA-Credentials, got %q", got)
	}
}

// TestTrustedHost 验证未列出的 host 返回 400。
func TestTrustedHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TrustedHost([]string{"plane.example.com", "www.example.com"}, zap.NewNop()))
	// 注册 /healthz 通用处理器，以便测试 healthz 跳过 host 校验的行为。
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.NoRoute(func(c *gin.Context) { c.Status(http.StatusNotFound) })
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	// 匹配的 host
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "plane.example.com"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("matched host: status = %d, want 200", w.Code)
	}

	// 不匹配的 host
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Host = "evil.example.com"
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Errorf("non-matched host: status = %d, want 400", w2.Code)
	}

	// healthz 跳过
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req3.Host = "evil.example.com"
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Errorf("healthz should always pass: status = %d", w3.Code)
	}
}

// TestTrustedHost_SkipWhenEmpty 验证配置为空时跳过校验。
func TestTrustedHost_SkipWhenEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TrustedHost(nil, zap.NewNop()))
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "anything"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("empty allowlist should skip: status = %d", w.Code)
	}
}

// TestRequireWorkspaceParam 验证 workspace_id 参数解析成功/失败行为。
func TestRequireWorkspaceParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid id sets ctx", func(t *testing.T) {
		r := gin.New()
		r.GET("/:workspace_id", RequireWorkspaceParam(), func(c *gin.Context) {
			id := c.GetInt64(CtxWorkspaceID)
			c.String(http.StatusOK, "id=%d", id)
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/123", nil))
		if w.Code != http.StatusOK || w.Body.String() != "id=123" {
			t.Errorf("status=%d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid id returns 422", func(t *testing.T) {
		r := gin.New()
		r.GET("/:workspace_id", RequireWorkspaceParam(), func(c *gin.Context) {
			c.String(http.StatusOK, "unreachable")
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/abc", nil))
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want 422", w.Code)
		}
	})
}
