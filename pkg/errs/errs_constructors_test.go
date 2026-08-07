// Package errs 便捷构造函数与 JSON 序列化测试。
package errs

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestConvenienceConstructors 验证 NotFound/Validation/Conflict/Forbidden
// 四个便捷构造函数正确设置错误码与 HTTP 状态码。
func TestConvenienceConstructors(t *testing.T) {
	tests := []struct {
		name string
		got  *AppError
		wantCode string
		wantHTTP int
	}{
		{"NotFound", NotFound("X.NOT_FOUND", "缺失"), "X.NOT_FOUND", http.StatusNotFound},
		{"Validation", Validation("X.INVALID", "非法"), "X.INVALID", http.StatusUnprocessableEntity},
		{"Conflict", Conflict("X.CONFLICT", "冲突"), "X.CONFLICT", http.StatusConflict},
		{"Forbidden", Forbidden("X.FORBIDDEN", "拒绝"), "X.FORBIDDEN", http.StatusForbidden},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got.Code != tc.wantCode {
				t.Errorf("code = %s, want %s", tc.got.Code, tc.wantCode)
			}
			if tc.got.HTTP != tc.wantHTTP {
				t.Errorf("http = %d, want %d", tc.got.HTTP, tc.wantHTTP)
			}
		})
	}
}

// TestErrorStringWithAndWithoutCause 验证 Error() 在有/无 cause 时的输出格式。
func TestErrorStringWithAndWithoutCause(t *testing.T) {
	noCause := New("A.B", "提示", http.StatusBadRequest)
	if noCause.Error() != "A.B: 提示" {
		t.Errorf("no-cause format = %q", noCause.Error())
	}

	withCause := noCause.Wrap(ErrInternal)
	if withCause.Error() != "A.B: 提示 (INTERNAL.ERROR: 服务内部错误)" {
		t.Errorf("with-cause format = %q", withCause.Error())
	}
}

// TestJSONSerializationExcludesHTTP 验证 HTTP 字段不参与 JSON 序列化，
// 且 details 为 nil 时被 omitempty 省略。
func TestJSONSerializationExcludesHTTP(t *testing.T) {
	e := New("TEST.CODE", "测试", http.StatusTeapot)
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["http"]; ok {
		t.Error("http field must not be serialized")
	}
	if _, ok := m["details"]; ok {
		t.Error("empty details must be omitted")
	}
	if m["code"] != "TEST.CODE" {
		t.Errorf("code = %v", m["code"])
	}
}

// TestWithCodeMessage 验证 WithCodeMessage 返回新实例且不污染原错误。
func TestWithCodeMessage(t *testing.T) {
	base := ErrValidation
	derived := base.WithCodeMessage("VALIDATION.SPECIAL", "特殊校验")
	if base.Code != "VALIDATION.FAILED" {
		t.Error("base error must stay pristine")
	}
	if derived.Code != "VALIDATION.SPECIAL" || derived.Message != "特殊校验" {
		t.Errorf("derived = %+v", derived)
	}
}
