// Package errs 定义了统一的领域错误类型 AppError 和全局错误码注册机制。
//
// ## 错误码命名规范
//
// 错误码遵循 DOMAIN.SNAKE_CASE 格式，由两部分组成，以点号分隔：
//
//   - DOMAIN（前缀）：大写字母，标识业务域或错误大类。
//     - 通用错误：VALIDATION / AUTH / RBAC / RATE_LIMIT / INTERNAL / RESOURCE
//     - 业务域：ISSUE（需求和缺陷域，代号 S3+）、SPRINT（迭代域，代号 S5）
//       后续新增域按项目管理代号扩展（如 TESTCASE、RELEASE 等）。
//   - SNAKE_CASE（后缀）：大写蛇形命名，描述具体错误场景。
//
// 命名示例：RESOURCE.NOT_FOUND、AUTH.TOKEN_EXPIRED、ISSUE.VERSION_CONFLICT。
//
// ## 设计目标
//
//   1. 单一错误类型跨越所有分层边界（handler → service → repository），统一序列化为错误信封响应。
//   2. 支持 Go 1.13+ errors.Is / errors.As 错误链追责，便于在中间层溯源根因。
//   3. 错误码全局唯一、枚举化管理，配合 docs/architecture/05 文档对外暴露给前端做国际化映射。
//
// ## 使用约定
//
//  - 错误码变量（如 ErrNotFound）是单例，业务代码中直接引用，不做值修改。
//  - 链式调用 From → WithCodeMessage → Details 构造错误实例，避免误改全局变量。
//  - 在中间层通过 errors.As(err, &appErr) 提取 AppError 进行日志埋点或链路追踪。
//  - HTTP 状态码与错误码一一对应，handler 层必须读取 AppError.HTTP 写入响应头。
//
// 参考模式：Google Cloud Error Model、Uber Go Style Guide 中的错误处理章节。
package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError 是跨越分层边界的唯一错误类型。
//
// handler 层负责将其序列化为统一错误信封：
//
//	{
//	  "code":    "RESOURCE.NOT_FOUND",
//	  "message": "资源不存在",
//	  "details": [{"field": "id", "reason": "不存在该 ID"}]
//	}
//
// 所有字段在 JSON 序列化时遵循 `json` 标签规则。
// HTTP 字段标记为 `json:"-"`，仅用于 handler 层写入 HTTP Status Code，不暴露给前端。
type AppError struct {
	// Code 为唯一错误码，由 "DOMAIN.SNAKE_CASE" 格式的 ASCII 字符串构成。
	// 是前端做国际化映射（i18n key）和告警分组的依据。
	Code string `json:"code"`

	// Message 是面向终端用户的中文提示文案，可直接渲染到 UI。
	// 要求：简洁、可理解、无敏感堆栈信息。前端根据此字段作为 fallback 文案展示。
	Message string `json:"message"`

	// HTTP 为该错误对应的 HTTP 响应状态码（如 400、404、500）。
	// handler 层序列化时必须据此设置 http.ResponseWriter.WriteHeader。
	HTTP int `json:"-"`

	// Details 携带字段级校验错误明细，通常用于请求体（DTO）参数校验失败场景。
	// 如果无字段级错误，该 slice 应为 nil（JSON 序列化时被 omitempty 省略）。
	Details []FieldDetail `json:"details,omitempty"`

	// cause 记录底层被包装的根因错误（如 sql.ErrNoRows、redis 连接超时等）。
	// 仅用于内部日志和链路追踪，不参与 JSON 序列化。通过 errors.Is / errors.As 即可提取。
	cause error
}

// FieldDetail 描述单个字段级别的校验失败信息。
//
// 典型使用场景：请求参数 bind 失败或 DTO validator 校验未通过时，
// 一个 AppError 可携带多个 FieldDetail，对应表单中多个字段的错误原因。
type FieldDetail struct {
	// Field 为请求参数名或 JSON path（如 "title"、"assignee.id"），
	// 前端可根据此字段将错误文案定位到对应表单控件。
	Field string `json:"field"`

	// Reason 为该字段的具体错误原因（中文），如 "不能超过 255 个字符"、"格式不符合邮箱规范"。
	// 要求：简洁、无敏感信息、可直接渲染。
	Reason string `json:"reason"`
}

// Error 实现 error 接口，返回格式为 "CODE: Message (cause)"。
// 当 cause 为 nil 时，括号部分省略。
//
// 设计的权衡：不要在 handler 层直接暴露此字符串给前端，应使用 JSON 信封结构；
// 此方法仅用于日志打印和内部调试。
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 实现错误链解包，使 errors.Is / errors.As 能正确遍历到 cause。
//
// 典型用法：
//
//	var appErr *errs.AppError
//	if errors.As(wrappedErr, &appErr) {
//	    slog.Error("handler error", "code", appErr.Code, "cause", appErr.Unwrap())
//	}
func (e *AppError) Unwrap() error { return e.cause }

// New 创建一个新的 AppError 实例（包级构造函数）。
//
// 用于构建一次性错误（如根据数据库约束动态构造错误码），
// 也用于初始化全局错误常量（见文件下方 var 块）。
//
// 示例：
//
//	return nil, errs.New("WORKSPACE.SLUG_TAKEN", "该链接标识已被使用", 409)
func New(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTP: httpStatus}
}

// From 从错误码单例（如 ErrNotFound）克隆一个新的 AppError 实例。
//
// 推荐在 service 层使用：它返回错误码变量的拷贝，确保后续链式调用
//（如 Details）不会修改全局共享的错误码单例。
//
// 使用场景：
//
//	if row == nil {
//	    return errs.ErrNotFound.From()
//	}
func (e *AppError) From() *AppError {
	clone := *e
	return &clone
}

// Wrap 包装一个底层错误作为 cause，返回新实例。
//
// 这是 errs.ErrInternal.Wrap(pgErr) 的快捷等价，供 service 层
// 在向上抛出 INTERNAL 错误时携带底层错误用于日志排查。
//
// 建议优先使用 From() 链式调用：
//
//	return errs.ErrInternal.From()  // 不带 cause
//	return errs.ErrInternal.Wrap(txErr)  // 带 cause
func (e *AppError) Wrap(err error) *AppError {
	clone := *e
	clone.cause = err
	return &clone
}

// WithCodeMessage 覆盖错误码和文案，用于同一场景在不同 endpoint 下
// 需要区分错误码（如 domainXxx 与 domainCreate 场景）的情况。
//
// 注意：应谨慎使用，仅在确实需要区分错误码时调用，避免破坏错误码全局单一职责。
func (e *AppError) WithCodeMessage(code, message string) *AppError {
	clone := *e
	clone.Code = code
	clone.Message = message
	return &clone
}

// Details 附加字段级错误明细。每次调用返回新实例（不变性）。
//
// 典型 service 层用法：
//
//	return errs.ErrValidation.From().Details(
//	    errs.FieldDetail{Field: "title", Reason: "不能为空"},
//	    errs.FieldDetail{Field: "assigneeId", Reason: "用户不存在"},
//	)
// WithDetails 附加字段级错误明细。每次调用返回新实例（不变性）。
func (e *AppError) WithDetails(details ...FieldDetail) *AppError {
	if len(details) == 0 {
		return e
	}
	clone := *e
	clone.Details = details
	return &clone
}

// ---------------------------------------------------------------------------
// 全局错误码注册表
//
// 所有错误码变量在此集中注册，杜绝散落在业务文件中的魔法字符串。
// 每个常量必须附带完整注释，说明：触发场景、HTTP 码含义、用户文案的设计考虑。
// ---------------------------------------------------------------------------

var (
	// ==========================================================================
	// 通用错误（Domain: VALIDATION / AUTH / RBAC / RESOURCE / RATE_LIMIT / INTERNAL）
	//
	// 这些错误不归属于某个具体业务域，适用于所有 endpoint。
	// ==========================================================================

	// ErrValidation 请求参数校验失败（HTTP 422 Unprocessable Entity）。
	//
	// 触发场景：handler 层 DTO bind 失败、validator tag 校验未通过、自定义参数规则（如文件大小、枚举范围）不满足。
	// HTTP 含义：请求格式正确但语义错误（不同于 400 Bad Request 的格式层面错误）。
	// 用户文案考虑：文案统一为 "参数校验失败"，具体字段的错误原因由 Details 数组提供，前端绑定到对应表单控件。
	ErrValidation = New("VALIDATION.FAILED", "参数校验失败", http.StatusUnprocessableEntity)

	// ErrNotImplemented 功能尚未实现（HTTP 501 Not Implemented）。
	//
	// 触发场景：模块处于规划/占位阶段（如 AI 能力未接入时返回占位错误），
	// 前端据此展示 "功能开发中" 状态而非笼统的 500。
	ErrNotImplemented = New("NOT_IMPLEMENTED", "功能尚未实现", http.StatusNotImplemented)

	// ErrUnauthorized 未认证或凭证已失效（HTTP 401 Unauthorized）。
	//
	// 触发场景：请求未携带 Token、Token 格式错误、Token 中间件反序列化失败。
	// HTTP 含义：RFC 7235 定义的身份认证缺失或无效。
	// 用户文案考虑：告知用户认证状态异常，引导重新登录。
	ErrUnauthorized = New("AUTH.UNAUTHORIZED", "未认证或凭证已失效", http.StatusUnauthorized)

	// ErrForbidden 权限不足（HTTP 403 Forbidden）。
	//
	// 触发场景：已认证用户尝试访问未授权的资源、RBAC 策略判定无操作权限、工作空间角色不匹配。
	// HTTP 含义：服务器理解请求但拒绝授权（与 401 不同，401 是未认证）。
	// 用户文案考虑：提示权限不足，不应泄露过多资源信息（防止越权探测）。
	ErrForbidden = New("RBAC.FORBIDDEN", "没有执行该操作的权限", http.StatusForbidden)

	// ErrNotFound 资源不存在（HTTP 404 Not Found）。
	//
	// 触发场景：根据 ID 查询数据库记录为 nil、关联资源已被删除、URL 路径资源路径不存在（仅限 REST 语义）。
	// HTTP 含义：服务器未找到目标资源。
	// 用户文案考虑：统一返回 "资源不存在"，不暴露数据库表名或内部 ID。
	ErrNotFound = New("RESOURCE.NOT_FOUND", "资源不存在", http.StatusNotFound)

	// ErrRateLimited 请求频率超限（HTTP 429 Too Many Requests）。
	//
	// 触发场景：API 网关或中间件层（redis/sliding-window）判定某用户/IP 的请求速率超过配额。
	// HTTP 含义：用户在给定时间内发送了过多请求。
	// 用户文案考虑：提示等待重试，对应响应应配合 Retry-After 头部使用。
	ErrRateLimited = New("RATE_LIMIT.EXCEEDED", "请求过于频繁，请稍后再试", http.StatusTooManyRequests)

	// ErrInternal 服务内部错误（HTTP 500 Internal Server Error）。
	//
	// 触发场景：未被任何其他错误码兜底的异常、第三方服务调用超时、数据库连接失败、panic 恢复后的兜底上抛。
	// HTTP 含义：服务器遇到意外情况无法完成请求。
	// 用户文案考虑：不返回异常堆栈，仅给通用提示；详细原因应记录到日志和链路追踪系统（如 slog / Sentry）。
	// 安全要点：Handler 层必须确保 ErrInternal 的 Details 字段绝不包含堆栈、SQL、内部 IP。
	ErrInternal = New("INTERNAL.ERROR", "服务内部错误", http.StatusInternalServerError)

	// ==========================================================================
	// 认证与账号域（Domain: AUTH）
	// ==========================================================================

	// ErrInvalidCredentials 账号或密码认证失败（HTTP 401 Unauthorized）。
	//
	// 触发场景：登录接口密码比对失败、OAuth/SSO 第三方回调校验不通过、多因素认证错误。
	// 用户文案考虑：统一模糊提示 "邮箱或密码错误"，不区分是邮箱不存在还是密码错误（避免用户枚举漏洞）。
	ErrInvalidCredentials = New("AUTH.INVALID_CREDENTIALS", "邮箱或密码错误", http.StatusUnauthorized)

	// ErrTokenExpired Token 已过期（HTTP 401 Unauthorized）。
	//
	// 触发场景：Access Token 超过 TTL、Refresh Token 失效、JWT exp claim 被中间件验证为过期。
	// 用户文案考虑：提示过期并引导重新登录；前端收到此错误码应跳登录页而非简单刷新。
	ErrTokenExpired = New("AUTH.TOKEN_EXPIRED", "登录已过期，请重新登录", http.StatusUnauthorized)

	// ==========================================================================
	// 安全域（Domain: SECURITY）
	// ==========================================================================

	// ErrCSRFTokenInvalid CSRF 令牌校验失败（HTTP 403 Forbidden）。
	//
	// 触发场景：浏览器 SPA 携带会话 Cookie 发起状态变更请求（POST/PUT/PATCH/DELETE）时，
	// 缺失 X-CSRF-Token 请求头、或请求头中的令牌值与会话绑定的预期令牌不匹配。
	// 设计说明：
	//   - 仅对携带 ydsz_access 会话 Cookie 的请求生效（浏览器 SPA）。
	//   - 纯 API 客户端（Authorization: Bearer / X-Api-Key，无会话 Cookie）不受影响，
	//     因为它们不属于 CSRF 攻击的载体场景。
	//   - 安全方法（GET/HEAD/OPTIONS）与 WebSocket 升级请求不受 CSRF 约束。
	// 用户文案考虑：文案模糊处理，不暴露校验细节，避免为攻击者构造绕过提供帮助。
	ErrCSRFTokenInvalid = New("SECURITY.CSRF_INVALID", "请求验证失败，请刷新页面后重试", http.StatusForbidden)

	// ==========================================================================
	// 需求/缺陷域（Domain: ISSUE，代号 S3+）
	// ==========================================================================

	// ErrVersionConflict 数据并发修改冲突（HTTP 409 Conflict）。
	//
	// 触发场景：乐观锁 version 字段比对失败（CAS 更新返回 affectedRows=0）、编辑他人已修改的工作项。
	// HTTP 含义：请求与目标资源的当前状态冲突。
	// 用户文案考虑：提示数据已被他人修改，要求前端刷新后重新提交，而非直接重试覆盖。
	ErrVersionConflict = New("ISSUE.VERSION_CONFLICT", "数据已被他人修改，请刷新后重试", http.StatusConflict)

	// ErrInvalidTransition 状态流转非法（HTTP 422 Unprocessable Entity）。
	//
	// 触发场景：工作项状态机不合法流转（如直接从 DRAFT 转为 DONE）、审批流程中禁止跳跃。
	// 用户文案考虑：提示当前状态不允许操作；前端接收到此错误时应调接口重新拉取状态机图以更新可执行操作列表。
	ErrInvalidTransition = New("ISSUE.INVALID_STATE_TRANSITION", "当前状态不允许该流转", http.StatusUnprocessableEntity)

	// ErrWBSDepthExceeded 工作项层级超限（HTTP 422 Unprocessable Entity）。
	//
	// 触发场景：工作项挂载为子级时检测到递归深度超过业务最大层级（如三级 WBS）。
	// 用户文案考虑：硬编码最大深度提示（与业务约定一致），前端可禁用过深的子级挂载按钮。
	ErrWBSDepthExceeded = New("ISSUE.WBS_DEPTH_EXCEEDED", "工作项层级最多支持三级", http.StatusUnprocessableEntity)

	// ErrCircularParent 子级挂载形成环路（HTTP 422 Unprocessable Entity）。
	//
	// 触发场景：尝试将工作项设置为自己的祖先节点，导致树结构出现环。
	// 用户文案考虑：提示用户挂载操作会导致环路，前端应在用户拖拽/选择时拦截候选节点。
	ErrCircularParent = New("ISSUE.CIRCULAR_PARENT", "不能将工作项挂载到自己或其子级之下", http.StatusUnprocessableEntity)

	// ==========================================================================
	// 迭代域（Domain: SPRINT，代号 S5）
	// ==========================================================================

	// ErrSprintConflict 迭代数据冲突（HTTP 409 Conflict）。
	//
	// 触发场景：并发操作同一迭代、迭代状态与他人提交的变更冲突。
	// 用户文案考虑：提示迭代状态冲突，前端应重新拉取迭代详情。
	ErrSprintConflict = New("SPRINT.CONFLICT", "迭代状态冲突", http.StatusConflict)

	// ErrSprintInvalidLifecycle 迭代生命周期操作非法（HTTP 422 Unprocessable Entity）。
	//
	// 触发场景：尝试对已归档的迭代执行编辑操作、在 CLOSED 状态下开始迭代。
	// 用户文案考虑：明确告知当前迭代状态不允许该操作，前端应根据迭代实际状态渲染/隐藏操作按钮。
	ErrSprintInvalidLifecycle = New("SPRINT.INVALID_LIFECYCLE", "当前迭代状态不允许该操作", http.StatusUnprocessableEntity)

	// ErrSprintCapacityExceeded 迭代容量超限（HTTP 422 Unprocessable Entity）。
	//
	// 触发场景：向迭代添加工作项时总 story point 或工作量超出迭代预设容量。
	// 用户文案考虑：提示容量超限，前端可增加容量校验以提前拦截并展示具体数值。
	ErrSprintCapacityExceeded = New("SPRINT.CAPACITY_EXCEEDED", "迭代容量已超出设定值", http.StatusUnprocessableEntity)

	// ==========================================================================
// 版本域（Domain: VERSION，代号 S6）
// ==========================================================================

// ErrVersionDataConflict 版本数据冲突（HTTP 409 Conflict）。
//
// 触发场景：并发修改同一版本、semver 唯一性冲突、乐观锁 version 字段比对失败。
ErrVersionDataConflict = New("VERSION.CONFLICT", "版本状态冲突或版本号已被占用", http.StatusConflict)

// ErrVersionInvalidLifecycle 版本生命周期非法（HTTP 422 Unprocessable Entity）。
//
// 触发场景：对已归档/已发布的版本执行编辑、尝试回退到 planning。
ErrVersionInvalidLifecycle = New("VERSION.INVALID_LIFECYCLE", "当前版本状态不允许该操作", http.StatusUnprocessableEntity)

	// ErrVersionSemverInvalid 语义版本号非法（HTTP 422 Unprocessable Entity）。
	//
	// 触发场景：Semver 解析失败、leading zero、patch 非整数。
	ErrVersionSemverInvalid = New("VERSION.SEMVER_INVALID", "版本号不符合语义版本规范", http.StatusUnprocessableEntity)

	// ErrVersionNotQualityGate 质量门禁未通过（HTTP 422 Unprocessable Entity）。
	//
	// 触发场景：发布前存在致命/严重未关闭缺陷且未配置 forceRelease。
	ErrVersionNotQualityGate = New("VERSION.QUALITY_GATE_BLOCKED", "质量门禁未通过：存在未关闭的致命/严重缺陷", http.StatusUnprocessableEntity)

	// ErrVersionChecklistIncomplete 发布清单未完成（HTTP 422 Unprocessable Entity）。
	//
	// 触发场景：必要检查项未全部勾选试图发布。
	ErrVersionChecklistIncomplete = New("VERSION.CHECKLIST_INCOMPLETE", "发布检查清单还有未完成的必要项", http.StatusUnprocessableEntity)

// ErrVersionNotFound 版本不存在（HTTP 404 Not Found）。
ErrVersionNotFound = New("VERSION.NOT_FOUND", "版本不存在", http.StatusNotFound)
)

// As 包装标准库 errors.As，方便调用方直接 errs.As(err, &appErr) 而无需额外导入 errors 包。
//
// 典型用法：在 handler 中间件中统一提取 AppError 写入日志和链路追踪。
//
//	var appErr *errs.AppError
//	if errs.As(recoveredErr, &appErr) {
//	    slog.Error("handler error", "code", appErr.Code, "http", appErr.HTTP)
//	} else {
//	    slog.Error("unknown error", "cause", recoveredErr)
//	}
func As(err error, target any) bool { return errors.As(err, target) }

// --- 便捷构造函数 ---

// NotFound 快速创建 RESOURCE.NOT_FOUND 错误。
func NotFound(code, message string) *AppError {
	return &AppError{Code: code, Message: message, HTTP: http.StatusNotFound}
}

// Validation 快速创建 VALIDATION 错误。
func Validation(code, message string) *AppError {
	return &AppError{Code: code, Message: message, HTTP: http.StatusUnprocessableEntity}
}

// Conflict 快速创建 CONFLICT 错误。
func Conflict(code, message string) *AppError {
	return &AppError{Code: code, Message: message, HTTP: http.StatusConflict}
}

// Forbidden 快速创建 FORBIDDEN 错误。
func Forbidden(code, message string) *AppError {
	return &AppError{Code: code, Message: message, HTTP: http.StatusForbidden}
}
