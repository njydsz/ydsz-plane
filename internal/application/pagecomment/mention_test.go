// Package pagecomment — 纯逻辑单元测试。
//
// 覆盖：@提及 姓名提取（去重、去尾点、空内容）。
package pagecomment

import (
	"reflect"
	"testing"
)

func TestExtractMentionNames(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		content string
		want    []string
	}{
		{"empty", "", nil},
		{"whitespace only", "   \n\t", nil},
		{"no mention", "hello world", nil},
		{"single mention", "ping @Alice please", []string{"Alice"}},
		{"multiple mentions", "@Alice and @Bob please review", []string{"Alice", "Bob"}},
		{"duplicate mentions", "@Alice @Alice @Bob", []string{"Alice", "Bob"}},
		{"trailing dot trimmed", "cc @Alice. thanks", []string{"Alice"}},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := ExtractMentionNames(c.content)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("ExtractMentionNames(%q) = %v, want %v", c.content, got, c.want)
			}
		})
	}
}

func TestExtractMentionNames_PreservesOrder(t *testing.T) {
	t.Parallel()

	got := ExtractMentionNames("@Carol @Alice @Bob")
	want := []string{"Carol", "Alice", "Bob"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("order not preserved: got %v, want %v", got, want)
	}
}
