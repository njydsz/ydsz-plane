// Package issue — 评论 HTML XSS 防护。
//
// 使用 bluemonday（GitHub 广泛采用的 HTML 净化器）对客户端传入的
// content_html 做服务端二次净化，阻断以下攻击向量：
//
//   - 内联脚本注入：<script>alert(1)</script>
//   - 事件处理器注入：<img onerror="steal()">
//   - 无值属性注入：<a href="javascript:alert(1)">
//   - 表达式/CSS 注入：<div style="background-url:evil">
//
// 防御纵深：
//   客户端（ProseMirror 渲染）→ 服务端二次净化（本模块）
//   → 前端渲染（v-html + trusted sanitizer）
//
// 由于 content_html 是从客户端富文本编辑器接受的，我们假设它是不可信的。
package issue

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

// safeHrefPattern 匹配安全的 http/https 链接或锚点；拒绝 javascript:、data: 等。
var safeHrefPattern = regexp.MustCompile(`^(https?:|#)`)

// SanitizeHTML 净化富文本 HTML，移除潜在的 XSS 载体。
// 传入空字符串时直接返回空字符串。
//
// 安全：每次调用都会新建 Policy 实例（bluemonday 实例的 regexp 不是并发安全的）。
func SanitizeHTML(input string) string {
	if input == "" {
		return ""
	}

	// 新 Policy 实例 — 仅允许安全的富文本标签
	p := bluemonday.NewPolicy()

	// 文本格式
	p.AllowElements(
		"p", "br", "strong", "b", "em", "i", "u", "s", "strike", "del",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "li", "blockquote", "pre", "code",
		"span", "sub", "sup",
	)

	// 链接 — 仅允许 http/https/# 开头的 href，屏蔽 javascript:chrome: 等伪协议
	p.AllowAttrs("href").Matching(safeHrefPattern).OnElements("a")
	p.AllowAttrs("rel").Globally()
	p.AllowAttrs("target").Globally()
	p.AllowAttrs("title").Globally()

	// 图片 — 仅允许 http/https 来源
	p.AllowAttrs("src").Matching(safeHrefPattern).OnElements("img")
	p.AllowAttrs("alt", "width", "height").Globally()

	// 表格
	p.AllowElements("table", "thead", "tbody", "tr", "th", "td")

	// code 语言标注
	p.AllowAttrs("class").OnElements("code", "pre")

	return p.Sanitize(input)
}

// StripHTML 剥除所有 HTML 标签，返回纯文本内容。
// 用于 content_stripped 字段（确保即使前端没剥干净，服务端也能兜底）。
func StripHTML(input string) string {
	if input == "" {
		return ""
	}
	return bluemonday.StrictPolicy().Sanitize(input)
}
