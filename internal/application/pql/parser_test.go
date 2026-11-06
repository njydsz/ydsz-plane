package pql

import (
	"strings"
	"testing"
)

func TestParseSimple(t *testing.T) {
	q, err := Parse(`state = "in_progress"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Where == nil {
		t.Fatal("expected where clause")
	}
	cmp, ok := q.Where.(*ComparisonExpr)
	if !ok {
		t.Fatalf("expected ComparisonExpr, got %T", q.Where)
	}
	if cmp.Field != "state" || cmp.Op != TokenEq || cmp.Value != "in_progress" {
		t.Fatalf("unexpected comparison: %+v", cmp)
	}
}

func TestParseLogical(t *testing.T) {
	q, err := Parse(`state = "active" AND priority = "high"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	logic, ok := q.Where.(*LogicalExpr)
	if !ok {
		t.Fatalf("expected LogicalExpr, got %T", q.Where)
	}
	if logic.Op != TokenAnd {
		t.Fatalf("expected AND, got %s", logic.Op)
	}
}

func TestParseNotIn(t *testing.T) {
	q, err := Parse(`NOT state = "cancelled" AND label in ("a", "b")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Where == nil {
		t.Fatal("expected where clause")
	}
}

func TestParseLike(t *testing.T) {
	q, err := Parse(`title ~ "登录"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmp := q.Where.(*ComparisonExpr)
	if cmp.Op != TokenTilde {
		t.Fatalf("expected tilde, got %s", cmp.Op)
	}
}

func TestParseSortLimit(t *testing.T) {
	q, err := Parse(`type = "task" sort:created_at desc limit:20 offset:5`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.SortBy != "created_at" || q.SortDir != "desc" {
		t.Fatalf("unexpected sort: %s %s", q.SortBy, q.SortDir)
	}
	if q.Limit != 20 || q.Offset != 5 {
		t.Fatalf("unexpected limit/offset: %d/%d", q.Limit, q.Offset)
	}
}

func TestParseMeAndAt(t *testing.T) {
	q, err := Parse(`assignee = "me" OR assignee = @zhangsan`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Where == nil {
		t.Fatal("expected where clause")
	}
}

func TestParseInvalidField(t *testing.T) {
	// 解析层不校验字段名，编译层校验
	q, err := Parse(`evil_field = "x"`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	c := NewCompiler()
	if _, cerr := c.Compile(q, 1, nil); cerr == nil {
		t.Fatal("expected compile error for unknown field")
	}
}

func TestCompileSimple(t *testing.T) {
	q, err := Parse(`priority = "high" AND estimate >= 5`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	c := NewCompiler()
	compiled, err := c.Compile(q, 1, nil)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	if !strings.Contains(compiled.WhereSQL, "priority = $1") {
		t.Fatalf("expected priority = $1, got %s", compiled.WhereSQL)
	}
	if !strings.Contains(compiled.WhereSQL, "estimate_point >= $2") {
		t.Fatalf("expected estimate_point >= $2, got %s", compiled.WhereSQL)
	}
	if len(compiled.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(compiled.Args))
	}
}

func TestCompileMe(t *testing.T) {
	// 无引号的 me 被识别为 TokenMe
	q, err := Parse(`assignee = me`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	c := NewCompiler()
	compiled, err := c.Compile(q, 42, nil)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	if compiled.Args[0] != int64(42) {
		t.Fatalf("expected me=42, got %v", compiled.Args[0])
	}
}

func TestCompileAtUser(t *testing.T) {
	q, err := Parse(`assignee = @zhangsan`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	c := NewCompiler()
	compiled, err := c.Compile(q, 1, map[string]int64{"zhangsan": 7})
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	if compiled.Args[0] != int64(7) {
		t.Fatalf("expected @zhangsan=7, got %v", compiled.Args[0])
	}
}

func TestCompileInvalidEnumValue(t *testing.T) {
	q, err := Parse(`priority = "super-urgent"`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	c := NewCompiler()
	_, err = c.Compile(q, 1, nil)
	if err == nil {
		t.Fatal("expected error for invalid enum value")
	}
}

func TestCompileInjectionBlocked(t *testing.T) {
	// 字段名非法（不在白名单）应被拒绝
	q, err := Parse(`" = "x"`)
	if err == nil {
		// 解析可能成功但编译应拒绝
		c := NewCompiler()
		if _, cerr := c.Compile(q, 1, nil); cerr == nil {
			t.Fatal("expected compile error for unknown field")
		}
	}
}

func TestCompileSortWhitelist(t *testing.T) {
	q, err := Parse(`state = "active" sort:title asc`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	c := NewCompiler()
	compiled, err := c.Compile(q, 1, nil)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	if compiled.SortColumn != "title" {
		t.Fatalf("expected title, got %s", compiled.SortColumn)
	}
}

func TestCompileInList(t *testing.T) {
	q, err := Parse(`label in ("backend", "frontend", "urgent")`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	c := NewCompiler()
	compiled, err := c.Compile(q, 1, nil)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	if !strings.Contains(compiled.WhereSQL, "IN ($1, $2, $3)") {
		t.Fatalf("expected IN list, got %s", compiled.WhereSQL)
	}
}
