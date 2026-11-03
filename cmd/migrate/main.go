// Command migrate 执行数据库初始化/迁移。
// 用法: go run ./cmd/migrate [up|down N|version]
//
// 三模式（自动探测 sql/ 目录）:
//  1. golang-migrate 增量模式：sql/ 下存在 NNNN_*.up.sql 时，使用 golang-migrate
//     （历史兼容，保留增量演进与回滚能力）。
//  2. 版本化增量模式：sql/ 下存在 V<ver>__<desc>.sql（如 V26.11.03__xxx.sql）时，
//     按版本号顺序执行；通过 ydsz_schema_version 表记录已执行版本，重复执行幂等跳过。
//     该模式是本仓库的实际约定（历史迁移均为此命名），修复了此前「文件存在但从未执行」
//     的断链问题。
//  3. 全量 dump 模式：sql/ 下仅有 ydsz-plane-init.sql 时，执行全量初始化脚本。
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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

// dumpVersion 是全量初始化脚本在状态表中的版本标识。
const dumpVersion = "0000_ydsz_plane_init"

// dumpFile 是全量初始化脚本文件名。
const dumpFile = "ydsz-plane-init.sql"

// sqlDir 是迁移脚本所在目录（相对当前工作目录）。
const sqlDir = "sql"

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

	// 自动探测：优先 golang-migrate 命名，其次版本化命名，最后全量 dump。
	if hasIncrementalMigrations() {
		return runIncremental(cmd, cfg.Database.URL)
	}
	if hasVersionedMigrations() {
		return runVersioned(cmd, cfg.Database.URL)
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

// hasVersionedMigrations 检查 sql/ 目录是否存在 V<ver>__<desc>.sql 版本化迁移。
//
// 命名示例：V26.11.03__drop_unused_tables.sql
// 该命名是本仓库的历史约定，此前未被任何执行路径识别（导致迁移从未生效）。
func hasVersionedMigrations() bool {
	matches, err := filepath.Glob("sql/V*__*.sql")
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

// --- 版本化增量迁移模式（V<ver>__<desc>.sql） ---

// versionStateTable 记录已执行的版本化迁移（幂等去重的依据）。
const versionStateTable = "ydsz_schema_version"

// versionedFile 描述一个版本化迁移文件。
type versionedFile struct {
	Path    string // 文件完整路径
	Version string // 版本号，如 "V26.11.03"
	Desc    string // 描述，如 "drop_unused_tables"
}

// runVersioned 执行 V<ver>__<desc>.sql 命名的版本化迁移。
func runVersioned(cmd, dbURL string) error {
	switch cmd {
	case "up":
		return versionedUp(dbURL)
	case "down":
		return fmt.Errorf("版本化迁移模式不支持自动回滚：请手工执行对应的逆向 DDL。已执行版本可用 `migrate version` 查看")
	case "version":
		return showVersionedVersion(dbURL)
	default:
		return fmt.Errorf("unknown command %q (up|down|version)", cmd)
	}
}

// versionedUp 按版本号顺序执行所有未应用的迁移。
//
// 幂等语义：每个版本在 ydsz_schema_version 表中有且仅有一条记录，
// 重复执行自动跳过；同一事务内完成 DDL + 版本记录，保证原子性。
func versionedUp(dbURL string) error {
	ctx := context.Background()
	pool, err := persistence.NewPool(ctx, dbURL, 4)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	files, err := listVersionedFiles(sqlDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Println("no versioned migrations found")
		return nil
	}

	applied, err := loadAppliedVersions(ctx, pool.Pool)
	if err != nil {
		return err
	}

	pending := make([]versionedFile, 0, len(files))
	for _, f := range files {
		if _, done := applied[f.Version]; !done {
			pending = append(pending, f)
		}
	}
	if len(pending) == 0 {
		fmt.Printf("all %d versioned migrations already applied\n", len(files))
		return nil
	}

	fmt.Printf("applying %d/%d pending migration(s)\n", len(pending), len(files))
	for i, f := range pending {
		fmt.Printf("  [%d/%d] %s (%s)\n", i+1, len(pending), f.Version, f.Desc)

		sqlBytes, err := os.ReadFile(f.Path)
		if err != nil {
			return fmt.Errorf("read %s: %w", f.Path, err)
		}
		stmts := splitSQLStatements(string(sqlBytes))
		if len(stmts) == 0 {
			// 空文件也记录版本，避免重复扫描。
			stmts = nil
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", f.Version, err)
		}

		for _, stmt := range stmts {
			if _, err := tx.Exec(ctx, stmt); err != nil {
				_ = tx.Rollback(ctx)
				return fmt.Errorf("apply %s: %w", f.Version, err)
			}
		}
		if _, err := tx.Exec(ctx,
			"INSERT INTO "+versionStateTable+" (version, description, applied_at) VALUES ($1, $2, now())",
			f.Version, f.Desc); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record %s: %w", f.Version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", f.Version, err)
		}
	}

	fmt.Printf("migrated: applied=%d\n", len(pending))
	return nil
}

// showVersionedVersion 输出已执行的版本化迁移。
func showVersionedVersion(dbURL string) error {
	ctx := context.Background()
	pool, err := persistence.NewPool(ctx, dbURL, 4)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	applied, err := loadAppliedVersions(ctx, pool.Pool)
	if err != nil {
		return err
	}
	if len(applied) == 0 {
		fmt.Println("version=none (no versioned migrations applied)")
		return nil
	}
	versions := make([]string, 0, len(applied))
	for v := range applied {
		versions = append(versions, v)
	}
	sort.Strings(versions)
	fmt.Printf("version=%s applied=%d\n", versions[len(versions)-1], len(versions))
	return nil
}

// loadAppliedVersions 读取已执行的版本集合。
// 若状态表不存在则自动创建（幂等），返回空集合。
func loadAppliedVersions(ctx context.Context, db *pgxpool.Pool) (map[string]string, error) {
	_, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS `+versionStateTable+` (
		version     VARCHAR(64) PRIMARY KEY,
		description VARCHAR(255),
		applied_at  TIMESTAMPTZ DEFAULT now()
	)`)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", versionStateTable, err)
	}

	rows, err := db.Query(ctx, `SELECT version, COALESCE(description, '') FROM `+versionStateTable)
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", versionStateTable, err)
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var v, d string
		if err := rows.Scan(&v, &d); err != nil {
			return nil, fmt.Errorf("scan %s: %w", versionStateTable, err)
		}
		out[v] = d
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s: %w", versionStateTable, err)
	}
	return out, nil
}

// listVersionedFiles 扫描 dir 目录下并按版本号排序 V<ver>__<desc>.sql 文件。
//
// dir 作为参数传入（而非固定 "sql/"），使单元测试可在临时目录验证，
// 避免 os.Chdir 带来的进程级副作用（go test 并行执行时会互相干扰）。
func listVersionedFiles(dir string) ([]versionedFile, error) {
	pattern := filepath.Join(dir, "V*__*.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	out := make([]versionedFile, 0, len(matches))
	for _, m := range matches {
		base := filepath.Base(m)
		name := strings.TrimSuffix(base, ".sql")
		parts := strings.SplitN(name, "__", 2)
		version := parts[0]
		desc := ""
		if len(parts) == 2 {
			desc = parts[1]
		}
		out = append(out, versionedFile{Path: m, Version: version, Desc: desc})
	}
	// 版本号格式统一（V26.11.03），字符串排序即等价于时间排序。
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
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
			if stmt != "" && !isCommentOnlyStatement(stmt) {
				stmts = append(stmts, stmt)
			}
			b.Reset()
			continue
		}

		b.WriteRune(ch)
	}

	// 末尾可能还有一条未以 ';' 结尾的语句
	if tail := strings.TrimSpace(b.String()); tail != "" && !isCommentOnlyStatement(tail) {
		stmts = append(stmts, tail)
	}
	return stmts
}

// isCommentOnlyStatement 判断语句是否为纯注释（无实际 SQL 内容）。
//
// 纯注释语句提交给 PostgreSQL 会报错（"cannot insert multiple commands into a prepared statement"
// 或空语句错误），需在切分时过滤。
// 判定：剥除所有 -- 注释行与空白后无剩余内容。
func isCommentOnlyStatement(stmt string) bool {
	lines := strings.Split(stmt, "\n")
	for _, ln := range lines {
		s := strings.TrimSpace(ln)
		if s == "" || strings.HasPrefix(s, "--") {
			continue
		}
		return false
	}
	return true
}

// isDollarTagChar 判断美元引用标签中的字符（字母、数字、下划线）。
func isDollarTagChar(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
