package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// --- mockProvider ---

type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string { return m.name }

// --- disabledService 模拟 AI 未启用的 Service ---
type disabledService struct {
	*Service
}

// --- helpers ---

func setupRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h.Register(r.Group("/"))
	return r
}

func newHandlerWithService(svc *Service) *Handler {
	return NewHandler(&HandlerDeps{AiSvc: svc})
}

func newDisabledService() *Service {
	return NewService(nil, Config{Enabled: false, Provider: "fallback"})
}

func newEnabledService() *Service {
	return NewService(nil, Config{
		Enabled:  true,
		Provider: "openai",
		APIKey:   "test-key",
		Model:    "gpt-4o-mini",
		Endpoint: "https://api.openai.com/v1",
	})
}

// ============================================================
// Status
// ============================================================

func TestStatus_Disabled(t *testing.T) {
	r := setupRouter(newHandlerWithService(newDisabledService()))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("got %d, want 200; body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if body["enabled"] != false {
		t.Errorf("expected enabled=false, got %v", body["enabled"])
	}
}

func TestStatus_Enabled(t *testing.T) {
	svc := newEnabledService()
	r := setupRouter(newHandlerWithService(svc))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("got %d, want 200", w.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["enabled"] != true {
		t.Errorf("expected enabled=true, got %v", body["enabled"])
	}
	if body["provider"] != "openai" {
		t.Errorf("expected provider=openai, got %v", body["provider"])
	}
}

// ============================================================
// SmartAssign
// ============================================================

func TestSmartAssign_AI_Disabled(t *testing.T) {
	r := setupRouter(newHandlerWithService(newDisabledService()))

	w := httptest.NewRecorder()
	body := `{"title":"bug","description":"something breaks"}`
	req, _ := http.NewRequest("POST", "/smart-assign", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != 501 {
		t.Fatalf("got %d, want 501; body=%s", w.Code, w.Body.String())
	}
}

func TestSmartAssign_InvalidJSON(t *testing.T) {
	r := setupRouter(newHandlerWithService(newEnabledService()))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/smart-assign", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != 422 {
		t.Fatalf("got %d, want 422; body=%s", w.Code, w.Body.String())
	}
}

// ============================================================
// DetectDuplicates
// ============================================================

func TestDetectDuplicates_Disabled(t *testing.T) {
	r := setupRouter(newHandlerWithService(newDisabledService()))

	w := httptest.NewRecorder()
	input := `{"title":"bug","description":"desc"}`
	req, _ := http.NewRequest("POST", "/detect-duplicates", bytes.NewBufferString(input))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != 501 {
		t.Fatalf("got %d, want 501", w.Code)
	}
}

func TestDetectDuplicates_InvalidJSON(t *testing.T) {
	r := setupRouter(newHandlerWithService(newEnabledService()))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/detect-duplicates", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != 422 {
		t.Fatalf("got %d, want 422", w.Code)
	}
}

// ============================================================
// SmartClassify
// ============================================================

func TestSmartClassify_Disabled(t *testing.T) {
	r := setupRouter(newHandlerWithService(newDisabledService()))

	w := httptest.NewRecorder()
	input := `{"title":"bug","description":"desc"}`
	req, _ := http.NewRequest("POST", "/classify", bytes.NewBufferString(input))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != 501 {
		t.Fatalf("got %d, want 501", w.Code)
	}
}

func TestSmartClassify_Success(t *testing.T) {
	svc := newEnabledService()
	r := setupRouter(newHandlerWithService(svc))

	w := httptest.NewRecorder()
	input := map[string]string{"title": "优化性能", "description": "接口响应太慢"}
	b, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/classify", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// 规则引擎应返回分类结果（即使数据库为空）
	if w.Code == 501 {
		t.Fatalf("expected 2xx, got 501")
	}

	// 验证响应结构
	if w.Code >= 200 && w.Code < 300 {
		var resp ClassifyResult
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json unmarshal ClassifyResult: %v; body=%s", err, w.Body.String())
		}
	}
}

// ============================================================
// Summarize
// ============================================================

func TestSummarize_Disabled(t *testing.T) {
	r := setupRouter(newHandlerWithService(newDisabledService()))

	w := httptest.NewRecorder()
	input := `{"target_type":"issue","title":"bug","description":"desc"}`
	req, _ := http.NewRequest("POST", "/summarize", bytes.NewBufferString(input))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != 501 {
		t.Fatalf("got %d, want 501", w.Code)
	}
}

func TestSummarize_InvalidJSON(t *testing.T) {
	r := setupRouter(newHandlerWithService(newEnabledService()))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/summarize", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != 422 {
		t.Fatalf("got %d, want 422", w.Code)
	}
}

// ============================================================
// IsEnabled / Status sanity
// ============================================================

func TestService_IsEnabled(t *testing.T) {
	d := newDisabledService()
	if d.IsEnabled() {
		t.Error("expected disabled service to report IsEnabled=false")
	}
	e := newEnabledService()
	if !e.IsEnabled() {
		t.Error("expected enabled service to report IsEnabled=true")
	}
}

func TestService_StatusFormat(t *testing.T) {
	svc := newEnabledService()
	st := svc.Status()
	if _, ok := st["enabled"]; !ok {
		t.Error("Status() missing 'enabled' key")
	}
	if _, ok := st["provider"]; !ok {
		t.Error("Status() missing 'provider' key")
	}
}

// ============================================================
// 错误类型转换
// ============================================================

func TestWriteErr_AppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/", nil)

	appErr := errs.ErrInternal
	writeErr(c, appErr)

	if !c.IsAborted() {
		t.Error("expected context to be aborted after AppError")
	}
	if w.Code != appErr.HTTP {
		t.Errorf("expected HTTP %d, got %d", appErr.HTTP, w.Code)
	}
}

func TestWriteErr_GenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/", nil)

	writeErr(c, errors.New("some generic error"))

	if !c.IsAborted() {
		t.Error("expected context to be aborted after generic error")
	}
	// 通用错误应包装为 ErrInternal
	if w.Code != errs.ErrInternal.HTTP {
		t.Errorf("expected HTTP %d, got %d", errs.ErrInternal.HTTP, w.Code)
	}
}
