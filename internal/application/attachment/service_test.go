// Package attachment 附件文件名净化纯函数测试。
package attachment

import (
	"testing"
)

// TestSanitizeFilename 验证文件名净化：
// 先取路径 base（filepath.Base，防路径穿越），再替换特殊字符。
func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"report.pdf", "report.pdf"},
		{"a/b/c.pdf", "c.pdf"},            // 取 base 防目录穿越
		{"..\\..\\evil.sh", "evil.sh"},     // 反斜杠路径取 base
		{"../etc/passwd", "passwd"},        // base 后无特殊字符
		{"file name.txt", "file_name.txt"}, // 空格转下划线
		{"quote\"star*question?pipe|angle<>", "quote_star_question_pipe_angle__"},
		{"", "."}, // filepath.Base("") 返回 "."
	}
	for _, tc := range tests {
		if got := sanitizeFilename(tc.in); got != tc.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestSanitizeFilename_LengthLimit 验证文件名超过 200 字节时被截断。
func TestSanitizeFilename_LengthLimit(t *testing.T) {
	long := ""
	for i := 0; i < 250; i++ {
		long += "a"
	}
	got := sanitizeFilename(long)
	if len(got) != 200 {
		t.Errorf("length = %d, want 200", len(got))
	}
}
