package issue

import (
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// commentSanitizer 评论富文本 HTML 白名单清洗器（服务端安全防线）。
//
// 对齐前端 TipTap 编辑器输出能力，同时遵循最小授权原则：
//   - 仅允许排版与行内样式标签（p/strong/em/u/s/a/ul/ol/li/code/pre/blockquote/br/h1-h6）
//   - 禁止 script / iframe / img-src 外链 / 事件属性（onclick 等）与 style 属性
//   - 链接仅允许 http/https/mailto，并强制 rel="noopener noreferrer" + target="_blank"
//
// 说明：content_json（TipTap 文档）仍原样存储用于二次编辑；content_html 是渲染源，
// 必须经此清洗后才落库，防止存储型 XSS（CommentItem 使用 v-html 渲染该字段）。
var commentSanitizer *bluemonday.Policy

func init() {
	// 用 StrictPolicy 作为基底（默认剥离所有标签/属性/URL），再按需放行，
	// 确保 img/iframe/script/style 等一律不可用（图片走独立附件机制）。
	p := bluemonday.StrictPolicy()

	// 富文本排版（含 TipTap 常用节点）
	p.AllowElements(
		"p", "br", "hr",
		"strong", "b", "em", "i", "u", "s", "strike",
		"a",
		"ul", "ol", "li",
		"code", "pre",
		"blockquote",
		"h1", "h2", "h3", "h4", "h5", "h6",
	)
	// 链接策略：仅 http/https/mailto；外链强制 nofollow + noreferrer（防 window.opener 漏洞）
	p.AllowAttrs("href").OnElements("a")
	p.AllowURLSchemes("http", "https", "mailto")
	p.AllowAttrs("target", "rel").Matching(regexp.MustCompile(`^(_blank|noopener|noreferrer|nofollow)$`)).OnElements("a")
	p.RequireNoFollowOnLinks(true)
	p.RequireNoReferrerOnLinks(true)

	// 有序列表 / 代码块可用属性（对齐 TipTap）
	p.AllowAttrs("start", "type").OnElements("ol")
	p.AllowAttrs("class").OnElements("code", "pre")

	commentSanitizer = p
}

// sanitizeCommentHTML 清洗评论 HTML；空串返回空串。
func sanitizeCommentHTML(html string) string {
	if strings.TrimSpace(html) == "" {
		return ""
	}
	return commentSanitizer.Sanitize(html)
}
