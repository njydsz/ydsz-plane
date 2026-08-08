// Command migrate 执行数据库初始化/迁移。
// 用法: go run ./cmd/migrate [up|down N|version]
//
// 双模式（自动探测 sql/ 目录）:
//  1. 增量迁移模式：sql/ 下存在 NNNN_*.up.sql 时，使用 golang-migrate（历史兼容，
//     保留增量演进与回滚能力）。
//  2. 全量 dump 模式：sql/ 下仅有 ydsz-plane-init.sql 时，执行全量初始化脚本。
//     dump 模式通过 ydsz_dump_state 表记录初始化状态，重复执行幂等跳过；
//     整个 dump 在单事务内执行，失败全量回滚，不产生半初始化状态。
//
// 参考标准：互联网大厂数据库变更管理（增量迁移为主、初始化脚本幂等、变更可追溯）。
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/jackc/pgx/v5"

	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

// dumpVersion 是全量初始化脚本在状态表中的版本标识。
const dumpVersion = "0000_ydsz_plane_init"

// dumpFile 是全量初始化脚本文件名。
const dumpFile = "ydsz-plane-init.sql"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}

// run 解析命令并分派到对应模式。
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	// 自动探测：存在增量迁移文件则走 golang-migrate，否则走全量 dump。
	if hasIncrementalMigrations() {
		return runIncremental(cmd, cfg.Database.URL)
	}
	return runDump(cmd, cfg.Database.URL)
}

// hasIncrementalMigrations 检查 sql/ 目录是否存在 NNNN_*.up.sql 增量迁移。
func hasIncrementalMigrations() bool {
	matches, err := filepath.Glob("sql/*.up.sql")
	if err != nil {
		return false
	}
	return len(matches) > 0
}

// --- 增量迁移模式（golang-migrate，保持原行为） ---

// runIncremental 使用 golang-migrate 执行增量迁移。
func runIncremental(cmd, dbURL string) error {
	m, err := migrate.New("file://sql", dbURL)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()

	switch cmd {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	case "down":
		n := 1
		if len(os.Args) > 2 {
			if n, err = strconv.Atoi(os.Args[2]); err != nil {
				return fmt.Errorf("invalid down steps: %w", err)
			}
		}
		if err := m.Steps(-n); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			return err
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
		return nil
	default:
		return fmt.Errorf("unknown command %q (up|down N|version)", cmd)
	}

	v, dirty, _ := m.Version()
	fmt.Printf("migrated: version=%d dirty=%v\n", v, dirty)
	return nil
}

// --- 全量 dump 模式 ---

// runDump 在全量 dump 模式下执行初始化。
func runDump(cmd, dbURL string) error {
	switch cmd {
	case "up":
		return dumpUp(dbURL)
	case "down":
		return fmt.Errorf("dump 模式不支持增量回滚：sql/ 下仅有 %s 全量初始化脚本。"+
			"如需重建数据库，请先 DROP SCHEMA public CASCADE 后重新执行 migrate up", dumpFile)
	case "version":
		return showDumpVersion(dbURL)
	default:
		return fmt.Errorf("unknown command %q (up|down N|version)", cmd)
	}
}

// dumpUp 执行全量初始化脚本（幂等：已初始化则跳过）。
// 整个 dump 在单事务内执行：任一句失败即回滚，不产生半初始化状态。
func dumpUp(dbURL string) error {
	ctx := context.Background()
	pool, err := persistence.NewPool(ctx, dbURL, 4)
	if err != nil {
		return err
	}
	defer pool.Close()

	// 状态表：独立命名，避免与 golang-migrate 的 schema_migrations 冲突。
	if _, err := pool.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS ydsz_dump_state (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("ensure ydsz_dump_state: %w", err)
	}

	var appliedAt time.Time
	err = pool.Pool.QueryRow(ctx,
		"SELECT applied_at FROM ydsz_dump_state WHERE version = $1", dumpVersion).Scan(&appliedAt)
	if err == nil {
		fmt.Printf("database already initialized via %s (%s), skipping\n",
			dumpFile, appliedAt.Format(time.RFC3339))
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("query ydsz_dump_state: %w", err)
	}

	content, err := os.ReadFile(filepath.Join("sql", dumpFile))
	if err != nil {
		return fmt.Errorf("read %s: %w", dumpFile, err)
	}
	if strings.TrimSpace(string(content)) == "" {
		return fmt.Errorf("%s is empty", dumpFile)
	}

	start := time.Now()
	tx, err := pool.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 按顶层 ';' 分语句执行 dump（与 psql -f 行为一致，正确处理 $$ 美元引用
	// 与字符串字面量内的 ';'）。整文件单 Exec 在 simple query 协议下对 Navicat
	// 风格大文件偶发解析异常，分语句逐条执行更健壮且仍在同一事务内保持原子性。
	for _, stmt := range splitSQLStatements(string(content)) {
		if _, err := tx.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("execute %s: %w", dumpFile, err)
		}
	}
	if _, err := tx.Exec(ctx,
		"INSERT INTO ydsz_dump_state (version) VALUES ($1)", dumpVersion); err != nil {
		return fmt.Errorf("record ydsz_dump_state: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	fmt.Printf("initialized database via %s in %s\n",
		dumpFile, time.Since(start).Round(time.Millisecond))
	return nil
}

// showDumpVersion 查询 dump 初始化状态。
func showDumpVersion(dbURL string) error {
	ctx := context.Background()
	pool, err := persistence.NewPool(ctx, dbURL, 2)
	if err != nil {
		return err
	}
	defer pool.Close()

	var appliedAt time.Time
	err = pool.Pool.QueryRow(ctx,
		"SELECT applied_at FROM ydsz_dump_state WHERE version = $1", dumpVersion).
		Scan(&appliedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("version=uninitialized (full dump not applied)")
			return nil
		}
		return fmt.Errorf("query ydsz_dump_state: %w", err)
	}
	fmt.Printf("version=%s applied_at=%s\n", dumpVersion, appliedAt.Format(time.RFC3339))
	return nil
}

// splitSQLStatements 将整段 SQL 按顶层 ';' 拆分为多条独立语句。
//
// 与 psql -f 行为一致：
//   - 忽略 '$$' / '$$tag$$' 美元引用体内的 ';'（函数体、DO 块等）
//   - 忽略单引号 / 双引号字符串字面量内的 ';'
//   - 忽略 '--' 行注释与 '/* */' 块注释内的 ';'
//
// 返回的非空语句已去除首尾空白；空语句（仅注释/空白）被跳过。
func splitSQLStatements(sql string) []string {
	var stmts []string
	var b strings.Builder
	inDollar := false
	dollarTag := ""
	inSingle := false
	inDouble := false
	inLineComment := false
	inBlockComment := false

	runes := []rune(sql)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		prev := byte(0)
		if i > 0 {
			prev = byte(runes[i-1])
		}

		// 块注释开始/结束
		if !inDollar && !inSingle && !inDouble && !inLineComment {
			if !inBlockComment && i+1 < len(runes) && ch == '/' && runes[i+1] == '*' {
				inBlockComment = true
				b.WriteRune(ch)
				b.WriteRune(runes[i+1])
				i++
				continue
			}
			if inBlockComment && i+1 < len(runes) && ch == '*' && runes[i+1] == '/' {
				inBlockComment = false
				b.WriteRune(ch)
				b.WriteRune(runes[i+1])
				i++
				continue
			}
		}

		// 行注释（仅在非字符串/非美元引用/非块注释状态）
		if !inDollar && !inSingle && !inDouble && !inBlockComment && ch == '-' && i+1 < len(runes) && runes[i+1] == '-' {
			inLineComment = true
			b.WriteRune(ch)
			b.WriteRune(runes[i+1])
			i++
			continue
		}
		if inLineComment {
			b.WriteRune(ch)
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}

		// 块注释内字符原样写入
		if inBlockComment {
			b.WriteRune(ch)
			continue
		}

		// 美元引用切换：$$ 或 $tag$$
		if !inSingle && !inDouble {
			if ch == '$' {
				// 尝试读取完整美元标签：$$ 或 $name$$
				end := i
				for end < len(runes) && (runes[end] == '$' || isDollarTagChar(runes[end])) {
					end++
				}
				if end < len(runes) && runes[end] == '$' {
					tag := string(runes[i : end+1])
					if !inDollar {
						inDollar = true
						dollarTag = tag
						b.WriteString(tag)
						i = end
						continue
					}
					// 匹配结束标签
					if tag == dollarTag {
						inDollar = false
						dollarTag = ""
						b.WriteString(tag)
						i = end
						continue
					}
					// 不匹配的 $...$ 当作普通文本
					b.WriteRune(ch)
					continue
				}
			}
		}

		// 字符串字面量
		if !inDollar {
			if ch == '\'' && !inDouble {
				if inSingle {
					if prev != '\\' { // 简化：标准下连续 '' 为转义，这里按 pgx 默认 standard_conforming_strings 处理
						inSingle = false
					}
				} else {
					inSingle = true
				}
				b.WriteRune(ch)
				continue
			}
			if ch == '"' && !inSingle {
				inDouble = !inDouble
				b.WriteRune(ch)
				continue
			}
		}

		// 顶层语句分隔符
		if ch == ';' && !inDollar && !inSingle && !inDouble {
			stmt := strings.TrimSpace(b.String())
			if stmt != "" {
				stmts = append(stmts, stmt)
			}
			b.Reset()
			continue
		}

		b.WriteRune(ch)
	}

	// 末尾可能还有一条未以 ';' 结尾的语句
	if tail := strings.TrimSpace(b.String()); tail != "" {
		stmts = append(stmts, tail)
	}
	return stmts
}

// isDollarTagChar 判断美元引用标签中的字符（字母、数字、下划线）。
func isDollarTagChar(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
