// Package persistence 测试：批量插入工具（bulk_insert.go）。
//
// 验证 COPY 数据源的迭代语义、列数校验、标识符安全包装。
package persistence

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// fakeCopyTx 记录 CopyFrom 调用参数，用于断言而不触达真实数据库。
type fakeCopyTx struct {
	tableName   pgx.Identifier
	columnNames []string
	rowSrc      pgx.CopyFromSource
	rows        [][]any
}

func (f *fakeCopyTx) CopyFrom(_ context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	f.tableName = tableName
	f.columnNames = columnNames
	f.rowSrc = rowSrc

	// 驱动数据源以捕获全部行。
	for rowSrc.Next() {
		vals, err := rowSrc.Values()
		if err != nil {
			return 0, err
		}
		f.rows = append(f.rows, vals)
	}
	return int64(len(f.rows)), rowSrc.Err()
}

func TestCopyRows_Iteration(t *testing.T) {
	t.Parallel()

	rows := [][]any{{"a", 1}, {"b", 2}, {"c", 3}}
	src := NewCopyRows(rows)

	var got [][]any
	for src.Next() {
		vals, err := src.Values()
		if err != nil {
			t.Fatalf("Values() error: %v", err)
		}
		got = append(got, vals)
	}
	if err := src.Err(); err != nil {
		t.Fatalf("Err() = %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("got %d rows, want 3", len(got))
	}
	for i, want := range rows {
		if len(got[i]) != len(want) {
			t.Errorf("row %d: got %d cols, want %d", i, len(got[i]), len(want))
		}
	}
}

func TestCopyRows_Empty(t *testing.T) {
	t.Parallel()

	src := NewCopyRows(nil)
	if src.Next() {
		t.Error("Next() = true for empty rows, want false")
	}
	if err := src.Err(); err != nil {
		t.Errorf("Err() = %v, want nil", err)
	}
}

func TestCopyRows_NextIsRepeatable(t *testing.T) {
	t.Parallel()

	src := NewCopyRows([][]any{{"a"}})
	// 重复调用 Next 不应 panic，且最终停在末尾。
	for i := 0; i < 5; i++ {
		src.Next()
	}
	if src.Next() {
		t.Error("Next() should return false after exhausting rows")
	}
}

func TestCopyFrom_PropagatesRows(t *testing.T) {
	t.Parallel()

	fake := &fakeCopyTx{}
	rows := [][]any{
		{"task-001", "First", "high", 1, 2},
		{"task-002", "Second", "medium", 1, 2},
	}
	n, err := CopyFrom(context.Background(), fake, "task",
		[]string{"identifier", "name", "priority", "workspace_id", "project_id"}, rows)
	if err != nil {
		t.Fatalf("CopyFrom() error: %v", err)
	}
	if n != 2 {
		t.Errorf("n = %d, want 2", n)
	}
	if len(fake.rows) != 2 {
		t.Errorf("captured %d rows, want 2", len(fake.rows))
	}
	if len(fake.tableName) != 1 || fake.tableName[0] != "task" {
		t.Errorf("tableName = %v, want [task]", fake.tableName)
	}
	if len(fake.columnNames) != 5 {
		t.Errorf("columnNames = %v, want 5 entries", fake.columnNames)
	}
}

func TestCopyFrom_EmptyRows(t *testing.T) {
	t.Parallel()

	fake := &fakeCopyTx{}
	n, err := CopyFrom(context.Background(), fake, "task", []string{"id"}, nil)
	if err != nil {
		t.Fatalf("CopyFrom() error: %v", err)
	}
	if n != 0 {
		t.Errorf("n = %d, want 0", n)
	}
	// 空数据不应触达数据库。
	if fake.rowSrc != nil {
		t.Error("CopyFrom should not call tx.CopyFrom for empty rows")
	}
}

func TestCopyFrom_EmptyColumnNames(t *testing.T) {
	t.Parallel()

	fake := &fakeCopyTx{}
	_, err := CopyFrom(context.Background(), fake, "task", nil, [][]any{{1}})
	if err == nil {
		t.Fatal("CopyFrom() should error on empty columnNames")
	}
	if !strings.Contains(err.Error(), "columnNames is empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCopyFrom_MismatchedColumnCount(t *testing.T) {
	t.Parallel()

	fake := &fakeCopyTx{}
	rows := [][]any{
		{"task-001", "First", "high", 1, 2},
		{"task-002", "Second"}, // 列数不足
	}
	_, err := CopyFrom(context.Background(), fake, "task",
		[]string{"identifier", "name", "priority", "workspace_id", "project_id"}, rows)
	if err == nil {
		t.Fatal("CopyFrom() should error on mismatched column count")
	}
	if !strings.Contains(err.Error(), "row 1") {
		t.Errorf("error should mention row 1: %v", err)
	}
	if fake.rowSrc != nil {
		t.Error("CopyFrom should validate before calling tx.CopyFrom")
	}
}

func TestIdentifier_String(t *testing.T) {
	t.Parallel()

	cases := []struct {
		ident string
		want  string
	}{
		{"task", `"task"`},
		{"public.task", `"public"."task"`},
		{"my_column", `"my_column"`},
	}
	for _, c := range cases {
		got := NewIdentifier(c.ident).String()
		if got != c.want {
			t.Errorf("Identifier(%q).String() = %q, want %q", c.ident, got, c.want)
		}
	}
}

func TestSplitIdent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  []string
	}{
		{"task", []string{"task"}},
		{"public.task", []string{"public", "task"}},
		{"a.b.c", []string{"a", "b", "c"}},
		{"", []string{""}},
	}
	for _, c := range cases {
		got := splitIdent(c.input)
		if len(got) != len(c.want) {
			t.Errorf("splitIdent(%q) = %v, want %v", c.input, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitIdent(%q)[%d] = %q, want %q", c.input, i, got[i], c.want[i])
			}
		}
	}
}
