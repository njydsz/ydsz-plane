// Package searchql 解析器单元测试。
package searchql

import (
	"testing"
	"time"
)

// TestParseBasic 验证基本解析场景。
func TestParseBasic(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantText  string
		wantClauses int
		wantDegraded bool
	}{
		{
			name:    "纯文本搜索",
			input:   "登录页 闪退",
			wantText: "登录页 闪退",
			wantClauses: 0,
		},
		{
			name:    "空输入",
			input:   "",
			wantText: "",
			wantClauses: 0,
		},
		{
			name:      "单字段过滤",
			input:     "project:YD",
			wantText:  "",
			wantClauses: 1,
		},
		{
			name:      "字段+文本混合",
			input:     "登录页 project:YD type:defect",
			wantText:  "登录页",
			wantClauses: 2,
		},
		{
			name:      "引号短语",
			input:     `"支付回调" AND module:支付`,
			wantText:  `"支付回调"`,
			wantClauses: 1,
		},
		{
			name:      "me()函数",
			input:     "assignee:me()",
			wantText:  "",
			wantClauses: 1,
		},
		{
			name:      "now()函数",
			input:     "created>now(-7d)",
			wantText:  "",
			wantClauses: 1,
		},
		{
			name:      "比较运算符",
			input:     "severity>=3 priority!=low",
			wantText:  "",
			wantClauses: 2,
		},
		{
			name:      "in列表",
			input:     "status in (todo, doing, review)",
			wantText:  "",
			wantClauses: 1,
		},
		{
			name:      "括号分组",
			input:     "(project:YD OR project:TG) status:todo",
			wantText:  "",
			wantClauses: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := Parse(tt.input)
			if q.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", q.Text, tt.wantText)
			}
			if len(q.Clauses) != tt.wantClauses {
				t.Errorf("Clauses len = %d, want %d", len(q.Clauses), tt.wantClauses)
			}
			if q.IsDegraded != tt.wantDegraded {
				t.Errorf("IsDegraded = %v, want %v", q.IsDegraded, tt.wantDegraded)
			}
		})
	}
}

// TestParseClauseValues 验证字段值的正确解析。
func TestParseClauseValues(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantField string
		wantOp    string
		wantVal   any
	}{
		{
			name:    "冒号操作符",
			input:   "project:YD",
			wantField: "project",
			wantOp:    ":",
			wantVal:   "YD",
		},
		{
			name:    "等于操作符",
			input:   "type=defect",
			wantField: "type",
			wantOp:    "=",
			wantVal:   "defect",
		},
		{
			name:    "不等于操作符",
			input:   "priority!=low",
			wantField: "priority",
			wantOp:    "!=",
			wantVal:   "low",
		},
		{
			name:    "me()值",
			input:   "assignee:me()",
			wantField: "assignee",
			wantOp:    ":",
			wantVal:   "__CURRENT_USER__",
		},
		{
			name:    "currentUser()值",
			input:   "assignee:currentUser()",
			wantField: "assignee",
			wantOp:    ":",
			wantVal:   "__CURRENT_USER__",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := Parse(tt.input)
			if len(q.Clauses) != 1 {
				t.Fatalf("expected 1 clause, got %d", len(q.Clauses))
			}
			c := q.Clauses[0]
			if c.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", c.Field, tt.wantField)
			}
			if c.Operator != tt.wantOp {
				t.Errorf("Operator = %q, want %q", c.Operator, tt.wantOp)
			}
			if c.Value != tt.wantVal {
				t.Errorf("Value = %v, want %v", c.Value, tt.wantVal)
			}
		})
	}
}

// TestParseNowOffset 验证时间偏移解析。
func TestParseNowOffset(t *testing.T) {
	now := time.Now().Truncate(24 * time.Hour)
	tests := []struct {
		offset string
		want   time.Time
	}{
		{"", now},
		{"-7d", now.Add(-7 * 24 * time.Hour)},
		{"+1w", now.Add(7 * 24 * time.Hour)},
		{"-3h", now.Add(-3 * time.Hour)},
	}

	for _, tt := range tests {
		got := parseNowOffset(tt.offset)
		gotTime, ok := got.(time.Time)
		if !ok {
			t.Errorf("parseNowOffset(%q) returned non-time value: %v", tt.offset, got)
			continue
		}
		if !gotTime.Equal(tt.want) {
			t.Errorf("parseNowOffset(%q) = %v, want %v", tt.offset, gotTime, tt.want)
		}
	}
}

// TestKnownFields 验证字段列表。
func TestKnownFields(t *testing.T) {
	fields := KnownFields()
	if len(fields) == 0 {
		t.Error("KnownFields should not be empty")
	}
	for _, f := range fields {
		if !IsValidField(f) {
			t.Errorf("IsValidField(%q) = false, want true", f)
		}
	}
	if IsValidField("nonexistent") {
		t.Error("IsValidField('nonexistent') = true, want false")
	}
}

// TestParseEdgeCases 边界情况测试。
func TestParseEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantDegraded bool
	}{
		{
			name:    "只有空白",
			input:   "   \t  \n ",
			wantDegraded: false,
		},
		{
			name:    "特殊字符",
			input:   "测试@#$%^&*",
			wantDegraded: false,
		},
		{
			name:    "超长输入",
			input:   "这是一个非常长的搜索查询字符串用来测试解析器在大输入下的性能和正确性" +
				"它包含了很多中文字符以及一些英文单词mixed together来模拟真实的搜索场景" +
				"project:YD type:task priority:high assignee:me()",
			wantDegraded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := Parse(tt.input)
			if q.IsDegraded != tt.wantDegraded {
				t.Errorf("IsDegraded = %v, want %v", q.IsDegraded, tt.wantDegraded)
			}
			if q.Raw != tt.input {
				t.Errorf("Raw = %q, want %q", q.Raw, tt.input)
			}
		})
	}
}

// BenchmarkParse 基准测试解析性能（目标：<100µs per query）。
func BenchmarkParse(b *testing.B) {
	inputs := []string{
		"登录页 闪退",
		"project:YD status:todo assignee:me()",
		"type:defect severity>=3 created>now(-7d)",
		`"支付回调" AND module:支付`,
		"project:YD type:task priority:high status in (todo, doing) assignee:me() created>now(-7d)",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parse(inputs[i%len(inputs)])
	}
}
