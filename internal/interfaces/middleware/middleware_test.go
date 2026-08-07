// Package middleware 中间件纯逻辑与安全头测试。
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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
	r.Use(SecurityHeaders())
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
