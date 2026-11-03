// Command migrate — 版本化迁移模式的单元测试。
//
// 覆盖：文件扫描与版本排序、版本号解析、SQL 语句切分的幂等性。
// 不依赖真实数据库（DB 交互由集成测试覆盖）。
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListVersionedFiles_SortsByVersion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	sqlDir := filepath.Join(dir, "sql")
	if err := os.MkdirAll(sqlDir, 0o755); err != nil {
		t.Fatalf("mkdir sql: %v", err)
	}
	names := []string{
		"V26.11.03__zzz_last.sql",
		"V26.09.24__aaa_first.sql",
		"V26.10.03__mmm_middle.sql",
	}
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(sqlDir, n), []byte("-- test\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", n, err)
		}
	}

	files, err := listVersionedFiles(sqlDir)
	if err != nil {
		t.Fatalf("listVersionedFiles: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("got %d files, want 3", len(files))
	}

	wantOrder := []string{"V26.09.24", "V26.10.03", "V26.11.03"}
	for i, want := range wantOrder {
		if files[i].Version != want {
			t.Errorf("files[%d].Version = %q, want %q", i, files[i].Version, want)
		}
	}

	// 描述解析正确。
	if files[0].Desc != "aaa_first" {
		t.Errorf("files[0].Desc = %q, want %q", files[0].Desc, "aaa_first")
	}
}

func TestListVersionedFiles_IgnoresNonMatching(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	sqlDir := filepath.Join(dir, "sql")
	if err := os.MkdirAll(sqlDir, 0o755); err != nil {
		t.Fatalf("mkdir sql: %v", err)
	}
	files := map[string]string{
		"ydsz-plane-init.sql": "-- dump\n",
		"V26.11.03__ok.sql":    "-- ok\n",
		"README.md":            "not sql\n",
		"V_missing_sep.sql":    "-- no double underscore\n",
	}
	for n, c := range files {
		if err := os.WriteFile(filepath.Join(sqlDir, n), []byte(c), 0o644); err != nil {
			t.Fatalf("write %s: %v", n, err)
		}
	}

	got, err := listVersionedFiles(sqlDir)
	if err != nil {
		t.Fatalf("listVersionedFiles: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d files, want 1 (only V26.11.03__ok.sql)", len(got))
	}
	if got[0].Version != "V26.11.03" {
		t.Errorf("Version = %q, want V26.11.03", got[0].Version)
	}
}

func TestSplitSQLStatements_Idempotent(t *testing.T) {
	t.Parallel()

	input := `-- comment
CREATE TABLE a (id BIGINT);
INSERT INTO a VALUES (1);
`
	stmts := splitSQLStatements(input)
	if len(stmts) != 2 {
		t.Fatalf("got %d statements, want 2: %v", len(stmts), stmts)
	}

	// 重复切分结果一致（幂等）。
	again := splitSQLStatements(input)
	if len(again) != len(stmts) {
		t.Errorf("splitSQLStatements not idempotent: %d vs %d", len(again), len(stmts))
	}
}

func TestSplitSQLStatements_EmptyInput(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"", "   ", "\n\n", "-- only a comment\n"} {
		if got := splitSQLStatements(in); len(got) != 0 {
			t.Errorf("splitSQLStatements(%q) = %v, want empty", in, got)
		}
	}
}
