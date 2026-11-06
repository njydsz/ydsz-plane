// Package pagecomment — 页面评论与行内评论领域。
//
// 对标 Plane 的 Page Comments / Inline Comments：
//   - 页面级评论：整篇文档的讨论线程；
//   - 行内评论：锚定到文档块（block_id）并引用原文（quote_text）；
//   - 线程回复：parent_id 形成树状讨论；
//   - 表情回应：对评论附加 emoji 反应；
//   - 线程解决：resolved_at / resolved_by 标记讨论已闭环。
//
// 数据表：
//   - page_comments         （id 由应用层 genID 生成）
//   - page_comment_reactions（id 由应用层 genID 生成）
package pagecomment

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// PageType 评论归属的页面类型。
type PageType string

const (
	// PageTypeProject 项目文档页面（pages 表）。
	PageTypeProject PageType = "project"
	// PageTypeKnowledge 知识库文档页面（knowledge_pages 表）。
	PageTypeKnowledge PageType = "knowledge"
)

// IsValid 校验页面类型是否合法。
func (t PageType) IsValid() bool {
	return t == PageTypeProject || t == PageTypeKnowledge
}

// PageComment 页面评论（支持页面级与行内评论）。
type PageComment struct {
	ID          int64  `json:"id"`
	WorkspaceID int64  `json:"workspace_id"`
	PageID      int64  `json:"page_id"`             // 关联页面
	PageType    string `json:"page_type"`           // "project" | "knowledge"
	ParentID    *int64 `json:"parent_id,omitempty"` // 回复的评论 ID（线程）
	AuthorID    int64  `json:"author_id"`
	Content     string `json:"content"` // 纯文本内容
	// 行内评论定位（可选）
	BlockID   string `json:"block_id,omitempty"`   // 文档块标识
	QuoteText string `json:"quote_text,omitempty"` // 被引用的原文
	// 线程解决状态
	ResolvedAt *time.Time `json:"resolved_at,omitempty"` // 线程已解决
	ResolvedBy int64      `json:"resolved_by,omitempty"`
	// 审核状态
	Hidden           bool       `json:"hidden"`                     // 是否被审核隐藏
	ModeratedAt      *time.Time `json:"moderated_at,omitempty"`      // 最近审核时间
	ModeratedBy      int64      `json:"moderated_by,omitempty"`      // 最近审核人
	ModerationReason string     `json:"moderation_reason,omitempty"` // 审核原因
	// 提及
	Mentions     []int64    `json:"mentions,omitempty"`
	IsEdited     bool       `json:"is_edited"`
	AuthorName   string     `json:"author_name,omitempty"`
	AuthorAvatar string     `json:"author_avatar,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// PageCommentReaction 评论表情回应。
type PageCommentReaction struct {
	ID        int64     `json:"id"`
	CommentID int64     `json:"comment_id"`
	UserID    int64     `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

// --- Input Types ---

// CreateCommentInput 创建评论入参。
type CreateCommentInput struct {
	WorkspaceID int64
	PageID      int64
	PageType    PageType
	ParentID    *int64
	AuthorID    int64
	Content     string
	BlockID     string
	QuoteText   string
	Mentions    []int64
}

// UpdateCommentInput 更新评论入参（指针字段为 nil 时不更新）。
type UpdateCommentInput struct {
	Content  *string
	Mentions []int64
}

// ListCommentsOptions 评论列表查询选项。
type ListCommentsOptions struct {
	WorkspaceID int64
	PageID      int64
	PageType    PageType
	Limit       int
	Offset      int
	// IncludeHidden=true 时包含被审核隐藏的评论（管理视图）。
	IncludeHidden bool
	// HiddenOnly=true 时仅返回被隐藏的评论（审核队列）。
	HiddenOnly bool
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID（与其它应用层包策略一致）。
//
// 结构：高 42 位为毫秒时间戳，低 22 位为密码学随机数。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
