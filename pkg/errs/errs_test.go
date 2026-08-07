// Package errs 测试：验证 AppError 的包装、解包与不变性语义。
package errs

import (
	"errors"
	"testing"
)

// TestWrapAndAs 验证 Wrap 包装 cause 后 errors.As 可解析为 *AppError，
// 且 errors.Is 能穿透错误链发现底层 cause。
func TestWrapAndAs(t *testing.T) {
	cause := errors.New("db down")
	err := ErrInternal.Wrap(cause)

	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatal("errors.As must resolve *AppError")
	}
	if appErr.Code != "INTERNAL.ERROR" {
		t.Fatalf("code = %s", appErr.Code)
	}
	if !errors.Is(err, cause) {
		t.Fatal("wrapped cause must be discoverable")
	}
}

// TestWithDetailsDoesNotMutateOriginal 验证 WithDetails 返回新实例，
// 全局注册的错误码单例不会被污染（不可变性约束）。
func TestWithDetailsDoesNotMutateOriginal(t *testing.T) {
	detailed := ErrValidation.WithDetails(FieldDetail{Field: "email", Reason: "invalid"})
	if len(ErrValidation.Details) != 0 {
		t.Fatal("registry error must stay pristine")
	}
	if len(detailed.Details) != 1 || detailed.Details[0].Field != "email" {
		t.Fatalf("details = %+v", detailed.Details)
	}
}
