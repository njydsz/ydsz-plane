// Package errs defines the unified application error type and the error-code
// registry conventions (DOMAIN.SNAKE_CASE), per docs/architecture/05.
package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is the single error type crossing layer boundaries. Handlers map
// it to the unified error envelope.
type AppError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"` // user-facing (zh)
	HTTP    int            `json:"-"`
	Details []FieldDetail  `json:"details,omitempty"`
	cause   error
}

// FieldDetail describes a single field-level validation failure.
type FieldDetail struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap exposes the underlying cause for errors.Is/As chains.
func (e *AppError) Unwrap() error { return e.cause }

// New creates an AppError.
func New(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTP: httpStatus}
}

// Wrap attaches an underlying cause.
func (e *AppError) Wrap(err error) *AppError {
	clone := *e
	clone.cause = err
	return &clone
}

// WithDetails attaches field-level details.
func (e *AppError) WithDetails(details ...FieldDetail) *AppError {
	clone := *e
	clone.Details = details
	return &clone
}

// --- Error-code registry (increment as domains are implemented) ---

var (
	// generic
	ErrValidation    = New("VALIDATION.FAILED", "参数校验失败", http.StatusUnprocessableEntity)
	ErrUnauthorized  = New("AUTH.UNAUTHORIZED", "未认证或凭证已失效", http.StatusUnauthorized)
	ErrForbidden     = New("RBAC.FORBIDDEN", "没有执行该操作的权限", http.StatusForbidden)
	ErrNotFound      = New("RESOURCE.NOT_FOUND", "资源不存在", http.StatusNotFound)
	ErrRateLimited   = New("RATE_LIMIT.EXCEEDED", "请求过于频繁，请稍后再试", http.StatusTooManyRequests)
	ErrInternal      = New("INTERNAL.ERROR", "服务内部错误", http.StatusInternalServerError)

	// auth
	ErrInvalidCredentials = New("AUTH.INVALID_CREDENTIALS", "邮箱或密码错误", http.StatusUnauthorized)
	ErrTokenExpired       = New("AUTH.TOKEN_EXPIRED", "登录已过期，请重新登录", http.StatusUnauthorized)

	// issue domain (S3+)
	ErrVersionConflict    = New("ISSUE.VERSION_CONFLICT", "数据已被他人修改，请刷新后重试", http.StatusConflict)
	ErrInvalidTransition  = New("ISSUE.INVALID_STATE_TRANSITION", "当前状态不允许该流转", http.StatusUnprocessableEntity)
	ErrWBSDepthExceeded   = New("ISSUE.WBS_DEPTH_EXCEEDED", "工作项层级最多支持三级", http.StatusUnprocessableEntity)
	ErrCircularParent     = New("ISSUE.CIRCULAR_PARENT", "不能将工作项挂载到自己或其子级之下", http.StatusUnprocessableEntity)
)

// As is a convenience re-export so callers can do errs.As(err, &appErr).
func As(err error, target any) bool { return errors.As(err, target) }
