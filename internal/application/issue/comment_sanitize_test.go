package issue

import (
	"strings"
	"testing"
)

func TestSanitizeCommentHTML_StripsScriptAndEvents(t *testing.T) {
	input := `<p onclick="alert(1)">你好<script>alert('xss')</script></p><img src=x onerror=alert(1)>`
	out := sanitizeCommentHTML(input)
	if strings.Contains(out, "<script") {
		t.Errorf("script 未被剥离: %s", out)
	}
	if strings.Contains(out, "onclick") || strings.Contains(out, "onerror") {
		t.Errorf("事件属性未被剥离: %s", out)
	}
	if strings.Contains(out, "<img") {
		t.Errorf("img 标签未被剥离: %s", out)
	}
	if !strings.Contains(out, "你好") {
		t.Errorf("正文内容丢失: %s", out)
	}
}

func TestSanitizeCommentHTML_KeepsRichText(t *testing.T) {
	input := `<p><strong>加粗</strong>与<em>斜体</em>，<a href="https://example.com" target="_blank">链接</a></p><ul><li>项</li></ul><pre><code>code</code></pre>`
	out := sanitizeCommentHTML(input)
	for _, want := range []string{"<strong>", "<em>", "<ul>", "<li>", "<pre>", "<code>", "https://example.com"} {
		if !strings.Contains(out, want) {
			t.Errorf("富文本标签 %s 被误删: %s", want, out)
		}
	}
	if strings.Contains(out, "<script") {
		t.Errorf("不应出现 script")
	}
}

func TestSanitizeCommentHTML_LinkAttrs(t *testing.T) {
	// 链接必须强制 nofollow + noreferrer，防止 window.opener 漏洞
	input := `<a href="https://evil.example.com" target="_blank" rel="x">bad</a>`
	out := sanitizeCommentHTML(input)
	if !strings.Contains(out, "nofollow") || !strings.Contains(out, "noreferrer") {
		t.Errorf("链接缺少安全 rel 属性: %s", out)
	}
	if strings.Contains(out, `rel="x"`) {
		t.Errorf("非法 rel 未被覆盖: %s", out)
	}
}

func TestSanitizeCommentHTML_Empty(t *testing.T) {
	if got := sanitizeCommentHTML(""); got != "" {
		t.Errorf("空串应返回空串, got %q", got)
	}
	if got := sanitizeCommentHTML("   "); got != "" {
		t.Errorf("空白串应返回空串, got %q", got)
	}
}
