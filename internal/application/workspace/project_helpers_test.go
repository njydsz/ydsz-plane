// Package workspace 项目 slug 与 identifier 规范化纯函数测试。
package workspace

import (
	"testing"
)

// TestNormalizeProjectSlug 验证 slug 规范化：小写、ASCII 非字母数字转连字符、去首尾连字符。
// 注意：中文字符不在 a-z/0-9 白名单内，会全部转为连字符，最终可能被 Trim 为空。
func TestNormalizeProjectSlug(t *testing.T) {
	tests := []struct {
		in, fallback, want string
	}{
		{"My Project", "x", "my-project"},
		{"  hello   world  ", "x", "hello-world"},
		{"ABC_123", "x", "abc-123"},
		{"--leading--trailing--", "x", "leading-trailing"},
		{"nochange", "x", "nochange"},
		{"", "Fallback Name", "fallback-name"},
		{"平台 开发", "x", ""}, // 纯中文 → 全连字符 → Trim 后为空
	}
	for _, tc := range tests {
		if got := normalizeProjectSlug(tc.in, tc.fallback); got != tc.want {
			t.Errorf("normalizeProjectSlug(%q, %q) = %q, want %q", tc.in, tc.fallback, got, tc.want)
		}
	}
}

// TestNormalizeIdentifier 验证 identifier 生成。
// 语义：s 非空时直接大写并截断到 6 字符；s 为空时从 fallback 拆词生成。
func TestNormalizeIdentifier(t *testing.T) {
	tests := []struct {
		in, fallback, want string
	}{
		// s 非空：直接 ToUpper，截断 6 字符（含空格等原样保留）
		{"web", "x", "WEB"},
		{"my project", "x", "MY PRO"},      // 不拆词，直接截 6 字符
		{"toolongidentifier", "x", "TOOLON"}, // 截断 6
		{"a b c d e f g", "x", "A B C "},    // 截 6 字符含空格

		// s 为空：从 fallback 拆词
		{"", "core-platform", "CP"},   // 多词取首字母
		{"", "singleword", "SINGLE"},  // 单词取全称截 6
		{"", "一长串中文默认名", "一长"}, // 中文按字节截断（UTF-8 3字节/字 → 前 2 字）
	}
	for _, tc := range tests {
		if got := normalizeIdentifier(tc.in, tc.fallback); got != tc.want {
			t.Errorf("normalizeIdentifier(%q, %q) = %q, want %q", tc.in, tc.fallback, got, tc.want)
		}
	}
}
