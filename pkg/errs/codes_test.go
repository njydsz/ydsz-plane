package errs

import (
	"fmt"
	"strings"
	"testing"
)

// TestCodeUniqueness 验证所有 Code 常量字符串值互不相同。
// 重复的错误码会导致前端 i18n 映射歧义、Sentry 分组混乱。
func TestCodeUniqueness(t *testing.T) {
	codes := []Code{
		// 通用 1xxx
		CodeValidation,
		CodeNotFound,
		CodeAuth,
		CodeForbidden,
		CodeConflict,
		CodeRateLimit,
		CodeInternal,

		// 工作项 2xxx
		CodeIssueNotFound,
		CodeIssueInvalidState,
		CodeIssueMoveUnsupported,
		CodeIssueDepthLimit,
		CodeIssueVersion,
		CodeIssueParentTypeMismatch,

		// 空间 3xxx
		CodeWorkspaceNotFound,
		CodeWorkspaceMemberNotFound,
		CodeWorkspacePermissionDenied,
		CodeWorkspaceSlugConflict,
		CodeWorkspaceLimitReached,

		// 邀请 4xxx
		CodeInvitationCodeInvalid,
		CodeInvitationEmailConflict,
		CodeInvitationExpired,
		CodeInvitationRoleEscalation,

		// 认证 5xxx
		CodeAuthTokenExpired,
		CodeAuthTokenInvalid,
		CodeAuthSSOCallbackInvalid,

		// 上传 6xxx
		CodeUploadSizeExceeded,
		CodeUploadMimeNotAllowed,
		CodeUploadFileTypeNotAllowed,

		// 搜索 7xxx
		CodeSearchIndexUnavailable,
		CodeSearchJQLParse,

		// 通知 8xxx
		CodeNotificationNotFound,

		// 自动化 9xxx
		CodeAutomationRuleNotFound,
		CodeAutomationTriggerNotFound,
	}

	seen := make(map[string]Code, len(codes))
	for _, c := range codes {
		s := string(c)
		if prev, exists := seen[s]; exists {
			t.Errorf("duplicate code %q: %s and %s", s, prev, c)
		}
		seen[s] = c
	}
}

// TestCodeNonEmpty 确保所有 Code 常量值非空。
// 空值会导致 JSON 响应中 code 字段为空字符串，前端 i18n 无法匹配。
func TestCodeNonEmpty(t *testing.T) {
	codes := map[string]Code{
		"CodeValidation":              CodeValidation,
		"CodeNotFound":                CodeNotFound,
		"CodeAuth":                    CodeAuth,
		"CodeForbidden":               CodeForbidden,
		"CodeConflict":                CodeConflict,
		"CodeRateLimit":               CodeRateLimit,
		"CodeInternal":                CodeInternal,
		"CodeIssueNotFound":           CodeIssueNotFound,
		"CodeIssueInvalidState":       CodeIssueInvalidState,
		"CodeIssueMoveUnsupported":    CodeIssueMoveUnsupported,
		"CodeIssueDepthLimit":         CodeIssueDepthLimit,
		"CodeIssueVersion":            CodeIssueVersion,
		"CodeIssueParentTypeMismatch": CodeIssueParentTypeMismatch,
		"CodeWorkspaceNotFound":       CodeWorkspaceNotFound,
		"CodeWorkspaceMemberNotFound": CodeWorkspaceMemberNotFound,
		"CodeWorkspacePermissionDenied": CodeWorkspacePermissionDenied,
		"CodeWorkspaceSlugConflict":   CodeWorkspaceSlugConflict,
		"CodeWorkspaceLimitReached":   CodeWorkspaceLimitReached,
		"CodeInvitationCodeInvalid":   CodeInvitationCodeInvalid,
		"CodeInvitationEmailConflict": CodeInvitationEmailConflict,
		"CodeInvitationExpired":       CodeInvitationExpired,
		"CodeInvitationRoleEscalation": CodeInvitationRoleEscalation,
		"CodeAuthTokenExpired":        CodeAuthTokenExpired,
		"CodeAuthTokenInvalid":        CodeAuthTokenInvalid,
		"CodeAuthSSOCallbackInvalid":  CodeAuthSSOCallbackInvalid,
		"CodeUploadSizeExceeded":      CodeUploadSizeExceeded,
		"CodeUploadMimeNotAllowed":    CodeUploadMimeNotAllowed,
		"CodeUploadFileTypeNotAllowed": CodeUploadFileTypeNotAllowed,
		"CodeSearchIndexUnavailable":  CodeSearchIndexUnavailable,
		"CodeSearchJQLParse":          CodeSearchJQLParse,
		"CodeNotificationNotFound":    CodeNotificationNotFound,
		"CodeAutomationRuleNotFound":  CodeAutomationRuleNotFound,
		"CodeAutomationTriggerNotFound": CodeAutomationTriggerNotFound,
	}
	for name, c := range codes {
		if c == "" {
			t.Errorf("%s is empty", name)
		}
	}
}

// TestWithCodeSetsCode 验证 WithCode 正确设置错误码且不污染原错误。
func TestWithCodeSetsCode(t *testing.T) {
	original := ErrValidation.From()
	if original.Code != "VALIDATION_ERROR" {
		t.Fatalf("setup: original code = %s", original.Code)
	}

	derived := original.WithCode(CodeValidation)
	if derived.Code != "VALIDATION_ERROR" {
		t.Errorf("WithCode: derived code = %s, want VALIDATION_ERROR", derived.Code)
	}
	// 原始实例未改变
	if original.Code != "VALIDATION_ERROR" {
		t.Errorf("WithCode: original mutated to %s", original.Code)
	}
}

// TestSetCodeMutatesInstance 验证 SetCode 直接修改当前实例。
func TestSetCodeMutatesInstance(t *testing.T) {
	e := &AppError{Code: "OLD.CODE", Message: "msg", HTTP: 400}
	result := e.SetCode(CodeIssueNotFound)
	if result != e {
		t.Error("SetCode must return the same pointer")
	}
	if e.Code != "ISSUE_NOT_FOUND" {
		t.Errorf("SetCode: code = %s, want ISSUE_NOT_FOUND", e.Code)
	}
}

// TestCodeStringConversion 验证 Code 到 string 的隐式转换不会破坏
// gin.Context.Set 的使用模式（基于 interface{} 的类型断言）。
func TestCodeStringConversion(t *testing.T) {
	// 模拟 gin.Context.Set / Get 使用模式：
	//   ctx.Set("error_code", e.Code)
	//   raw, _ := ctx.Get("error_code")
	//   code := raw.(string)  // 必须可以断言为 string
	var ctxMock interface{} = string(CodeIssueNotFound)
	raw, ok := ctxMock.(string)
	if !ok {
		t.Fatal("Code must be assertable as string through interface{}")
	}
	if raw != "ISSUE_NOT_FOUND" {
		t.Errorf("string assertion = %s", raw)
	}
}

// TestWithCodeCompatibleWithExistingMethods 验证 WithCode 可与已有方法链式组合，
// 且最终 JSON 输出 code 字段为 string 类型。
func TestWithCodeCompatibleWithExistingMethods(t *testing.T) {
	err := ErrInternal.From().
		WithCode(CodeInternal).
		WithCodeMessage("CUSTOM.MSG", "自定义").
		WithDetails(FieldDetail{Field: "f", Reason: "r"})

	if err.Code != "CUSTOM.MSG" {
		t.Errorf("code = %s, want CUSTOM.MSG", err.Code)
	}
	if len(err.Details) != 1 || err.Details[0].Field != "f" {
		t.Errorf("details = %+v", err.Details)
	}

	// 验证 Error() 方法仍然正常工作
	errStr := err.Error()
	if !strings.Contains(errStr, "CUSTOM.MSG") || !strings.Contains(errStr, "自定义") {
		t.Errorf("Error() = %q, missing expected parts", errStr)
	}

	// codeToString 转换结果应为 ASCII
	codeStr := string(CodeIssueNotFound)
	for _, r := range codeStr {
		if r > 127 {
			t.Errorf("code %q contains non-ASCII rune: %U", codeStr, r)
		}
	}
}

// TestCodeToStringZeroAlloc 验证 codeToString 是零分配的类型转换。
func TestCodeToStringZeroAlloc(t *testing.T) {
	c := Code("TEST.CODE")
	s := codeToString(c)
	if s != "TEST.CODE" {
		t.Errorf("codeToString = %q", s)
	}
	// 验证 string(c) 与 codeToString(c) 等价
	if string(c) != fmt.Sprintf("%s", c) {
		t.Errorf("underlying string mismatch")
	}
}
