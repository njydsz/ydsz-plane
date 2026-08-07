// Package issue — XSS 防护单元测试。
//
// 覆盖：
//   - <script> 标签应被完全剥离
//   - onerror/onclick 事件处理器应被移除
//   - javascript: 伪协议 href 应被清理
//   - data: URI 应被屏蔽
//   - 合法标签（<strong>/<a>/<img>）应保留
//   - <a> 内文字符正确编码
package issue

import (
	"strings"
	"testing"
)

// TestSanitizeHTML_StripScript 验证 <script> 标签被完全剥离。
func TestSanitizeHTML_StripScript(t *testing.T) {
	input := `<p>Before</p><script>alert(1)</script><p>After</p>`
	got := SanitizeHTML(input)
	if strings.Contains(strings.ToLower(got), "<script") {
		t.Errorf("script tag should be stripped, got: %s", got)
	}
	if !strings.Contains(got, "Before") || !strings.Contains(got, "After") {
		t.Errorf("surrounding text should be preserved, got: %s", got)
	}
}

// TestSanitizeHTML_StripEventHandlers 验证 on* 事件处理器被移除。
func TestSanitizeHTML_StripEventHandlers(t *testing.T) {
	input := `<img src="https://example.com/x.png" onerror="steal.cookie()">`
	got := SanitizeHTML(input)
	if strings.Contains(strings.ToLower(got), "onerror") {
		t.Errorf("onerror should be stripped, got: %s", got)
	}
	// src 应该保留
	if !strings.Contains(got, "https://example.com/x.png") {
		t.Errorf("img src should be preserved, got: %s", got)
	}
}

// TestSanitizeHTML_BlockJavascriptHref 验证 javascript: 伪协议被清除。
func TestSanitizeHTML_BlockJavascriptHref(t *testing.T) {
	input := `<a href="javascript:alert(1)">click me</a>`
	got := SanitizeHTML(input)
	if strings.Contains(strings.ToLower(got), "javascript:") {
		t.Errorf("javascript: href should be blocked, got: %s", got)
	}
	// <a> 标签和文字应保留
	if !strings.Contains(got, "click me") {
		t.Errorf("link text should be preserved, got: %s", got)
	}
}

// TestSanitizeHTML_BlockDataURI 验证 data: URI 被屏蔽。
func TestSanitizeHTML_BlockDataURI(t *testing.T) {
	input := `<img src="data:text/html,<script>alert(1)</script>">`
	got := SanitizeHTML(input)
	if strings.Contains(strings.ToLower(got), "data:") {
		t.Errorf("data: URI should be blocked, got: %s", got)
	}
}

// TestSanitizeHTML_AllowSafeTags 验证合法富文本标签保留。
func TestSanitizeHTML_AllowSafeTags(t *testing.T) {
	input := `<p>Paragraph <strong>bold</strong> <em>italic</em> <code>code</code></p><ul><li>item</li></ul>`
	got := SanitizeHTML(input)
	for _, tag := range []string{"<p>", "<strong>", "<em>", "<code>", "<ul>", "<li>"} {
		if !strings.Contains(got, tag) {
			t.Errorf("safe tag %s should be preserved, got: %s", tag, got)
		}
	}
}

// TestSanitizeHTML_AllowSafeLink 验证 http/https 链接保留。
func TestSanitizeHTML_AllowSafeLink(t *testing.T) {
	input := `<a href="https://example.com/path?q=1">link</a>`
	got := SanitizeHTML(input)
	if !strings.Contains(got, `href="https://example.com/path?q=1"`) {
		t.Errorf("safe link should be preserved, got: %s", got)
	}
}

// TestSanitizeHTML_AllowAnchor 验证锚点链接保留。
func TestSanitizeHTML_AllowAnchor(t *testing.T) {
	input := `<a href="#section-1">jump</a>`
	got := SanitizeHTML(input)
	if !strings.Contains(got, `href="#section-1"`) {
		t.Errorf("anchor link should be preserved, got: %s", got)
	}
}

// TestSanitizeHTML_Empty 验证空字符串原样返回。
func TestSanitizeHTML_Empty(t *testing.T) {
	if got := SanitizeHTML(""); got != "" {
		t.Errorf("empty input should return empty, got: %s", got)
	}
}

// TestStripHTML_AllTagsRemoved 验证剥除所有 HTML 标签。
func TestStripHTML_AllTagsRemoved(t *testing.T) {
	input := `<p>Hello <strong>world</strong> <script>alert(1)</script></p>`
	got := StripHTML(input)
	if strings.Contains(got, "<") {
		t.Errorf("all tags should be stripped, got: %s", got)
	}
	if !strings.Contains(got, "Hello") || !strings.Contains(got, "world") {
		t.Errorf("text content should be preserved, got: %s", got)
	}
}

// TestStripHTML_Empty 验证空输入返回空。
func TestStripHTML_Empty(t *testing.T) {
	if got := StripHTML(""); got != "" {
		t.Errorf("empty input should return empty, got: %s", got)
	}
}
