//go:build dockertest

// Package dockertesthelper 提供基于 testcontainers-go 的集成测试容器管理。
//
// 通过构建标签 dockertest 控制：默认不引入 testcontainers 依赖，
// 仅在 `go test -tags dockertest` 时激活完整容器化测试。
//
// 依赖说明：本文件引用 github.com/testcontainers/testcontainers-go/modules/posts，
// 需要在 go.mod 中显式添加后方可编译。详见 dockertest_remarks.md。
package dockertesthelper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// DBConfig 是测试数据库配置。
type DBConfig struct {
	DriverName string
	DSN        string
}

// StartPostgres 启动一个 postgres 容器并返回其 DSN。
//
// 实现特性：
//   - 使用 testcontainers-go/modules/postgres 模块
//   - 镜像 postgres:16-alpine
//   - 自动执行 sql/ 目录下的迁移脚本
//   - 注册 t.Cleanup 自动销毁容器
//   - 支持 DOCKERTEST_REUSE=true 复用已存在容器（环境变量驱动）
//
// 调用方通过返回的 DBConfig.DSN 创建 persistence.Pool。
func StartPostgres(ctx context.Context, t testing.TB) DBConfig {
	t.Helper()

	image := "postgres:16-alpine"
	if v := os.Getenv("DOCKERTEST_POSTGRES_IMAGE"); v != "" {
		image = v
	}

	// 容器数据库默认配置
	const (
		dbUser     = "test"
		dbPassword = "test"
		dbName     = "ydsz_plane_test"
	)

	container, err := tcpostgres.Run(
		ctx,
		image,
		tcpostgres.WithDatabase(dbName),
		tcpostgres.WithUsername(dbUser),
		tcpostgres.WithPassword(dbPassword),
		testcontainers.WithReuse(os.Getenv("DOCKERTEST_REUSE") == "true"),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_INITDB_ARGS": "--encoding=UTF-8",
		}),
	)
	if err != nil {
		t.Fatalf("dockertest: start postgres container: %v", err)
	}

	dsn, err := container.ConnectionString(ctx,
		"sslmode=disable",
		"application_name=dockertest",
	)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("dockertest: get connection string: %v", err)
	}

	// 注册清理：无论测试 panic/t.FailNow/正常退出，均销毁容器
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("dockertest: terminate container warning: %v", err)
		}
	})

	cfg := DBConfig{
		DriverName: "pgx",
		DSN:        dsn,
	}

	// 运行迁移脚本
	runMigrations(ctx, t, dsn)

	return cfg
}

// runMigrations 自动执行 sql/ 目录下的所有 .sql 文件。
// 文件按字典序执行（与 golang-migrate 版本号命名约定兼容）。
// 出错时 t.Fatalf 终止——迁移失败说明容器或脚本有问题。
func runMigrations(ctx context.Context, t testing.TB, dsn string) {
	t.Helper()

	sqlDir := findSQLDir(t)
	if sqlDir == "" {
		t.Log("dockertest: sql/ directory not found, skipping migrations")
		return
	}

	entries, err := os.ReadDir(sqlDir)
	if err != nil {
		t.Fatalf("dockertest: read sql dir: %v", err)
	}

	// 按文件名排序（ReadDir 已按文件名排）
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		path := filepath.Join(sqlDir, entry.Name())
		content, rerr := os.ReadFile(path)
		if rerr != nil {
			t.Fatalf("dockertest: read %s: %v", entry.Name(), rerr)
		}

		// pgxpool 不支持多语句 Exec（无.simple protocol batch），
		// 分号切分后逐条执行（跳过空语句）。
		statements := splitSQLStatements(string(content))
		for i, stmt := range statements {
			if strings.TrimSpace(stmt) == "" {
				continue
			}
			if _, perr := executeStatement(ctx, dsn, stmt); perr != nil {
				t.Fatalf("dockertest: migrate %s (stmt#%d): %v", entry.Name(), i, perr)
			}
		}
	}

	t.Logf("dockertest: migrations applied from %s", sqlDir)
}

// executeStatement 通过独立连接执行单条 SQL。
// 迁移脚本可能包含 CREATE EXTENSION 等需要超级用户的命令，
// 因此每次新建连接而非用连接池（避免 pool 的副作用）。
func executeStatement(ctx context.Context, dsn, sql string) (commandTag string, err error) {
	pool, perr := pgxpool.New(ctx, dsn)
	if perr != nil {
		return "", perr
	}
	defer pool.Close()

	ct, cerr := pool.Exec(ctx, sql)
	if cerr != nil {
		return "", cerr
	}
	return ct.String(), nil
}

// splitSQLStatements 按分号切分 SQL（不处理 $$  dollar-quoting，
// 适用于本项目的 ydsz-plane-init.sql 风格——无 dollar-quoting）。
// 如有 PL/pgSQL 函数需要 dollar-quoting，可切换到 golang-migrate 执行。
func splitSQLStatements(sql string) []string {
	// 粗略切分：不在字符串内、不在注释内的分号才作为分隔符。
	// 对 init.sql 中包含 $$ ... $$ 的函数定义，这里做一个简化处理：
	// 遇到 $$ 视为一个整体不切分。
	var stmts []string
	var current strings.Builder
	inDollarQuote := false
	dollarTag := ""
	inSingleQuote := false
	inDoubleQuote := false
	inLineComment := false
	inBlockComment := false

	for i := 0; i < len(sql); i++ {
		ch := sql[i]

		// 行注释
		if inLineComment {
			current.WriteByte(ch)
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}
		// 块注释
		if inBlockComment {
			current.WriteByte(ch)
			if i+1 < len(sql) && ch == '*' && sql[i+1] == '/' {
				current.WriteByte(sql[i+1])
				inBlockComment = false
				i++
			}
			continue
		}
		// 单引号字符串
		if inSingleQuote && !inDollarQuote && !inDoubleQuote {
			current.WriteByte(ch)
			if ch == '\'' {
				// '' 转义
				if i+1 < len(sql) && sql[i+1] == '\'' {
					current.WriteByte(sql[i+1])
					i++
				} else {
					inSingleQuote = false
				}
			}
			continue
		}
		// 双引号标识符
		if inDoubleQuote {
			current.WriteByte(ch)
			if ch == '"' {
				inDoubleQuote = false
			}
			continue
		}
		// dollar quoting
		if inDollarQuote {
			current.WriteByte(ch)
			tagLen := len(dollarTag)
			if i+1 >= tagLen {
				// 检查是否到达结束 $$
				possible := sql[i+1-tagLen : i+1]
				if possible == dollarTag {
					// 检查后续字符
					inDollarQuote = false
				}
			}
			continue
		}

		// 起始状态检测
		if ch == '\'' {
			inSingleQuote = true
			current.WriteByte(ch)
			continue
		}
		if ch == '"' {
			inDoubleQuote = true
			current.WriteByte(ch)
			continue
		}
		if ch == '-' && i+1 < len(sql) && sql[i+1] == '-' {
			inLineComment = true
			current.WriteByte(ch)
			current.WriteByte(sql[i+1])
			i++
			continue
		}
		if ch == '/' && i+1 < len(sql) && sql[i+1] == '*' {
			inBlockComment = true
			current.WriteByte(ch)
			current.WriteByte(sql[i+1])
			i++
			continue
		}
		// dollar quoting 起始: $tag$
		if ch == '$' {
			// 向后查找 $tag$ 结尾
			if tag := findDollarTag(sql, i); tag != "" {
				inDollarQuote = true
				dollarTag = tag
				current.WriteString(tag)
				i += len(tag) - 1 // -1 because loop will i++
				continue
			}
		}

		if ch == ';' {
			stmts = append(stmts, current.String())
			current.Reset()
			continue
		}

		current.WriteByte(ch)
	}

	if current.Len() > 0 {
		stmts = append(stmts, current.String())
	}
	return stmts
}

// findDollarTag 检查从位置 i 开始的 $tag$ 模式，返回完整标签如 "$$" 或 "$func$"。
// 如果不是有效的 dollar-quote 起始，返回 ""。
func findDollarTag(s string, i int) string {
	if s[i] != '$' {
		return ""
	}
	// 找到下一个 '$'
	for j := i + 1; j < len(s); j++ {
		if s[j] == '$' {
			return s[i : j+1]
		}
	}
	return ""
}

// findSQLDir 从当前工作目录向上查找 sql/ 目录。
func findSQLDir(t testing.TB) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := cwd
	for {
		candidate := filepath.Join(dir, "sql")
		if info, serr := os.Stat(candidate); serr == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// NewPool 是 StartPostgres 的便捷组合：启动容器 + 创建 persistence.Pool。
func NewPool(ctx context.Context, t testing.TB) (*persistence.Pool, DBConfig) {
	t.Helper()
	cfg := StartPostgres(ctx, t)
	pool, err := persistence.NewPool(ctx, cfg.DSN, 2)
	if err != nil {
		t.Fatalf("dockertest: new pool: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool, cfg
}

// ParallelPool 用于需要独立容器隔离的并行测试子用例。
// 每个子测试启动自己的容器，通过 t.Parallel() 并行执行。
//
// 用法：
//
//	t.Run("sub", func(t *testing.T) {
//	    t.Parallel()
//	    pool, _ := dockertesthelper.ParallelPool(ctx, t)
//	    // ... 使用 pool
//	})
func ParallelPool(ctx context.Context, t testing.TB) *persistence.Pool {
	t.Helper()
	pool, _ := NewPool(ctx, t)
	return pool
}

