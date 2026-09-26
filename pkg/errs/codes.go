// Package errs 错误码中心注册表。
//
// 本文件定义了标准化的业务错误码类型 Code 及其全局常量集合，
// 用于前端 i18n 自动匹配翻译、告警分类和服务间错误契约。
//
// 分组约定：
//   - 1xxx 通用
//   - 2xxx 工作项
//   - 3xxx 空间
//   - 4xxx 邀请
//   - 5xxx 认证
//   - 6xxx 上传
//   - 7xxx 搜索
//   - 8xxx 通知
//   - 9xxx 自动化
//
// 注意：本文件只增不改——已发布的错误码生命周期内不可变更含义，
// 如需废弃，仅在使用侧停止引用，不在本文件删除常量。
package errs

// Code 是标准化的业务错误码（字符串枚举）。
//
// 前端根据此 code 字段匹配 i18n key，实现错误消息的国际化。
// Sentry / 告警系统也根据 Code 做分组聚合。
//
// Code 底层类型为 string，可直接向下转型为 string 用于日志输出、
// JSON 序列化或 context 传递（如 gin.Context.Set），不影响存量调用方。
type Code string

const (
	// --------------------------------------------------------------------------
	// 通用 1xxx
	// --------------------------------------------------------------------------

	// CodeValidation 请求参数校验失败。
	CodeValidation Code = "VALIDATION_ERROR"

	// CodeNotFound 资源不存在。
	CodeNotFound Code = "NOT_FOUND"

	// CodeAuth 认证失败（未登录或凭证无效）。
	CodeAuth Code = "AUTH_ERROR"

	// CodeForbidden 权限不足。
	CodeForbidden Code = "FORBIDDEN"

	// CodeConflict 数据冲突（并发修改、唯一约束）。
	CodeConflict Code = "CONFLICT"

	// CodeRateLimit 请求频率超限。
	CodeRateLimit Code = "RATE_LIMIT"

	// CodeInternal 服务内部错误。
	CodeInternal Code = "INTERNAL_ERROR"

	// --------------------------------------------------------------------------
	// 工作项 2xxx
	// --------------------------------------------------------------------------

	// CodeIssueNotFound 工作项不存在。
	CodeIssueNotFound Code = "ISSUE_NOT_FOUND"

	// CodeIssueInvalidState 工作项状态不允许当前操作。
	CodeIssueInvalidState Code = "ISSUE_INVALID_STATE"

	// CodeIssueMoveUnsupported 工作项拖拽/移动场景不支持。
	CodeIssueMoveUnsupported Code = "ISSUE_MOVE_UNSUPPORTED"

	// CodeIssueDepthLimit 工作项层级超限。
	CodeIssueDepthLimit Code = "ISSUE_DEPTH_LIMIT_EXCEEDED"

	// CodeIssueVersion 工作项版本冲突（乐观锁）。
	CodeIssueVersion Code = "ISSUE_VERSION_CONFLICT"

	// CodeIssueParentTypeMismatch 工作项父级类型不匹配（如将缺陷挂载到迭代父级下）。
	CodeIssueParentTypeMismatch Code = "ISSUE_PARENT_TYPE_MISMATCH"

	// --------------------------------------------------------------------------
	// 空间 3xxx
	// --------------------------------------------------------------------------

	// CodeWorkspaceNotFound 工作空间不存在。
	CodeWorkspaceNotFound Code = "WORKSPACE_NOT_FOUND"

	// CodeWorkspaceMemberNotFound 空间成员不存在。
	CodeWorkspaceMemberNotFound Code = "WORKSPACE_MEMBER_NOT_FOUND"

	// CodeWorkspacePermissionDenied 空间层面权限不足。
	CodeWorkspacePermissionDenied Code = "WORKSPACE_PERMISSION_DENIED"

	// CodeWorkspaceSlugConflict 空间链接标识（slug）已被占用。
	CodeWorkspaceSlugConflict Code = "WORKSPACE_SLUG_CONFLICT"

	// CodeWorkspaceLimitReached 空间数量或资源已达上限。
	CodeWorkspaceLimitReached Code = "WORKSPACE_LIMIT_REACHED"

	// --------------------------------------------------------------------------
	// 邀请 4xxx
	// --------------------------------------------------------------------------

	// CodeInvitationCodeInvalid 邀请码无效。
	CodeInvitationCodeInvalid Code = "INVITATION_CODE_INVALID"

	// CodeInvitationEmailConflict 邀请邮箱已有待处理邀请。
	CodeInvitationEmailConflict Code = "INVITATION_EMAIL_CONFLICT"

	// CodeInvitationExpired 邀请已过期。
	CodeInvitationExpired Code = "INVITATION_EXPIRED"

	// CodeInvitationRoleEscalation 邀请尝试赋予超出邀请者自身权限的角色。
	CodeInvitationRoleEscalation Code = "INVITATION_ROLE_ESCALATION"

	// --------------------------------------------------------------------------
	// 认证 5xxx
	// --------------------------------------------------------------------------

	// CodeAuthTokenExpired 认证 Token 已过期。
	CodeAuthTokenExpired Code = "AUTH_TOKEN_EXPIRED"

	// CodeAuthTokenInvalid 认证 Token 无效（格式错误、签名失败）。
	CodeAuthTokenInvalid Code = "AUTH_TOKEN_INVALID"

	// CodeAuthSSOCallbackInvalid SSO 回调校验失败。
	CodeAuthSSOCallbackInvalid Code = "AUTH_SSO_CALLBACK_INVALID"

	// --------------------------------------------------------------------------
	// 上传 6xxx
	// --------------------------------------------------------------------------

	// CodeUploadSizeExceeded 上传文件大小超限。
	CodeUploadSizeExceeded Code = "UPLOAD_SIZE_EXCEEDED"

	// CodeUploadMimeNotAllowed 上传文件 MIME 类型不在白名单。
	CodeUploadMimeNotAllowed Code = "UPLOAD_MIME_NOT_ALLOWED"

	// CodeUploadFileTypeNotAllowed 上传文件扩展名/类型不允许。
	CodeUploadFileTypeNotAllowed Code = "UPLOAD_FILE_TYPE_NOT_ALLOWED"

	// --------------------------------------------------------------------------
	// 搜索 7xxx
	// --------------------------------------------------------------------------

	// CodeSearchIndexUnavailable 搜索索引不可用（如 Meilisearch/OpenSearch 故障）。
	CodeSearchIndexUnavailable Code = "SEARCH_INDEX_UNAVAILABLE"

	// CodeSearchJQLParse JQL 查询语法解析失败。
	CodeSearchJQLParse Code = "SEARCH_JQL_PARSE_ERROR"

	// --------------------------------------------------------------------------
	// 通知 8xxx
	// --------------------------------------------------------------------------

	// CodeNotificationNotFound 通知不存在。
	CodeNotificationNotFound Code = "NOTIFICATION_NOT_FOUND"

	// --------------------------------------------------------------------------
	// 自动化 9xxx
	// --------------------------------------------------------------------------

	// CodeAutomationRuleNotFound 自动化规则不存在。
	CodeAutomationRuleNotFound Code = "AUTOMATION_RULE_NOT_FOUND"

	// CodeAutomationTriggerNotFound 自动化触发器不存在。
	CodeAutomationTriggerNotFound Code = "AUTOMATION_TRIGGER_NOT_FOUND"
)

// codeToString 将 Code 转为底层 string，供 AppError 内部使用。
// 这是一个零开销的纯类型转换（Code 底层即 string）。
func codeToString(c Code) string { return string(c) }
