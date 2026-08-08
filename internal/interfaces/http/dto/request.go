// Package dto — HTTP API 请求 DTO。
//
// 设计原则：
//   - DTO 与域模型（domain model）解耦：handler 层负责 DTO → domain 入参转换。
//   - 所有 Update 请求使用指针字段区分「未传」(nil) 与「显式置零」("")。
//   - binding tag 做第一轮结构校验，复杂业务校验在 application 层完成。
//
// 校验规则遵循 Gin binding 约定：required / min / max / oneof / email / url / hexcolor。
package dto

// --- Workspace ---

// CreateWorkspaceRequest 创建工作空间的请求体。
type CreateWorkspaceRequest struct {
	// Name 工作空间显示名，1-80 字符。
	Name string `json:"name" binding:"required,min=1,max=80"`
	// Slug 友好 URL 标识符，留空则由后端根据 Name 自动生成。格式：^[a-z0-9-]+$。
	Slug string `json:"slug" binding:"omitempty,max=60"`
	// Timezone IANA 时区名（如 Asia/Shanghai），默认 UTC。
	Timezone string `json:"timezone" binding:"omitempty"`
	// Language 界面语言代码，支持 zh-CN / en-US / en。
	Language string `json:"language" binding:"omitempty,oneof=zh-CN en-US en"`
}

// UpdateWorkspaceRequest 更新工作空间的请求体（指针字段 = 可选更新）。
type UpdateWorkspaceRequest struct {
	// Name 新的显示名；nil 表示不更新。
	Name *string `json:"name,omitempty" binding:"omitempty,min=1,max=80"`
	// Timezone 新的时区；nil 表示不更新。
	Timezone *string `json:"timezone,omitempty"`
	// Language 新的语言代码；nil 表示不更新。
	Language *string `json:"language,omitempty" binding:"omitempty,oneof=zh-CN en-US en"`
	// LogoURL 工作空间 Logo 的公开 URL；nil 表示不更新，传空字符串表示清除。
	LogoURL *string `json:"logo_url,omitempty" binding:"omitempty,url,max=500"`
}

// --- Member ---

// ChangeRoleRequest 变更成员角色的请求体。
// 工作空间级角色: admin / member / guest / pm / po / techlead / qalead / dev / owner
// 项目级角色: admin / member（后端按路由校验）。
type ChangeRoleRequest struct {
	// Role 目标角色（oneof 在后端按上下文校验，前端根据路由传入）。
	Role string `json:"role" binding:"required"`
}

// --- Invitation ---

// SendInvitationRequest 发送成员邀请的请求体。
type SendInvitationRequest struct {
	// Email 被邀请人的邮箱（唯一标识，用于账户匹配）。
	Email string `json:"email" binding:"required,email"`
	// Role 邀请时预设的角色；被邀请人接受后直接获得该角色。
	Role string `json:"role" binding:"required,oneof=admin member guest"`
	// Message 附赠的邀请说明，显示在邀请邮件中。
	Message string `json:"message" binding:"omitempty,max=500"`
}

// AcceptInvitationRequest 接受邀请的请求体。
type AcceptInvitationRequest struct {
	// Token 邀请 URL 中携带的随机 token（原始值，非 hash）。
	Token string `json:"token" binding:"required"`
}

// --- API Token ---

// CreateApiTokenRequest 创建个人 API Token 的请求体。
type CreateApiTokenRequest struct {
	// Name 令牌名称（用途备注），1-80 字符。
	Name string `json:"name" binding:"required,min=1,max=80"`
	// Scopes 权限范围白名单（至少 1 个，全部须命中 apitoken 白名单）。
	// 可选值：read:workspace write:workspace read:issues write:issues
	// read:sprints write:sprints read:versions write:versions read:audit *
	Scopes []string `json:"scopes" binding:"required,min=1,max=20,dive,oneof=read:workspace write:workspace read:issues write:issues read:sprints write:sprints read:versions write:versions read:audit *"`
	// ExpiresInSeconds 有效期（秒），60 ~ 31536000（365 天）；缺省表示永不过期。
	ExpiresInSeconds *int64 `json:"expires_in_seconds,omitempty" binding:"omitempty,min=60,max=31536000"`
}

// --- Project ---

// CreateProjectRequest 创建项目的请求体。
type CreateProjectRequest struct {
	// Name 项目名称，1-80 字符。
	Name string `json:"name" binding:"required,min=1,max=80"`
	// Slug 友好 URL 片段，留空则由 Name 自动生成。
	Slug string `json:"slug" binding:"omitempty,max=60"`
	// Identifier 项目前缀（2-6 位大写字母），用于工作项标识（如 `PROJ-12`），留空则自动生成。
	Identifier string `json:"identifier" binding:"omitempty,max=6"`
	// Description 项目简介，最长 500 字符。
	Description string `json:"description" binding:"omitempty,max=500"`
	// Network 可见性：public（空间内默认可见） / private（仅成员可见） / internal（仅项目成员可见）。默认 public。
	Network string `json:"network" binding:"omitempty,oneof=public private internal"`
	// Icon Emoji 或图标标识，最长 32 字符。
	Icon string `json:"icon" binding:"omitempty,max=32"`
	// Color Hex 主题色（如 `#2563eb`），7 字符定长。
	Color string `json:"color" binding:"omitempty,hexcolor,len=7"`
	// Template 项目模板代码：agile（敏捷） / waterfall（瀑布） / generic（通用看板），默认 generic。
	Template string `json:"template" binding:"omitempty,oneof=agile waterfall generic"`
	// CoverImageUrl 封面图片 URL（可选）；传空字符串表示清除。
	CoverImageUrl string `json:"cover_image_url" binding:"omitempty,url,max=500"`
	// Modules 功能模块开关；null 表示全部启用。
	Modules *struct {
		Intake   *bool `json:"intake,omitempty"`
		Sprint   *bool `json:"sprint,omitempty"`
		Version  *bool `json:"version,omitempty"`
		Estimate *bool `json:"estimate,omitempty"`
	} `json:"modules,omitempty"`
}

// UpdateProjectRequest 更新项目的请求体（指针字段 = 可选更新）。
type UpdateProjectRequest struct {
	// Name 新名称；nil 表示不更新。
	Name *string `json:"name,omitempty" binding:"omitempty,min=1,max=80"`
	// Slug 新 URL 片段；nil 表示不更新。
	Slug *string `json:"slug,omitempty"`
	// Description 新简介；nil 表示不更新。
	Description *string `json:"description,omitempty"`
	// Network 新可见性；nil 表示不更新。
	Network *string `json:"network,omitempty" binding:"omitempty,oneof=public private internal"`
	// Icon 新图标；nil 表示不更新。
	Icon *string `json:"icon,omitempty"`
	// Color 新主题色；nil 表示不更新。
	Color *string `json:"color,omitempty" binding:"omitempty,hexcolor,len=7"`
	// CoverImageUrl 新封面图片 URL；nil 表示不更新；传空字符串表示清除。
	CoverImageUrl *string `json:"cover_image_url,omitempty" binding:"omitempty,url,max=500"`
	// Modules 功能模块开关；null 表示不更新（非 nil 则整体替换）。
	Modules *struct {
		Intake   *bool `json:"intake,omitempty"`
		Sprint   *bool `json:"sprint,omitempty"`
		Version  *bool `json:"version,omitempty"`
		Estimate *bool `json:"estimate,omitempty"`
	} `json:"modules,omitempty"`
}

// --- Project Member ---

// AddProjectMemberRequest 添加项目成员的请求体。
type AddProjectMemberRequest struct {
	// UserID 被添加的用户 ID（必须是同 workspace 成员）。
	UserID int64 `json:"user_id" binding:"required,min=1"`
	// Role 在项目中的角色：admin / member。
	Role string `json:"role" binding:"required,oneof=admin member"`
}
