// Command reconcile-search 对账 search_documents 索引与源表（issues/sprints/versions）的一致性。
//
// 用途（S8 出口 8.10）：
//   检测并修复索引漂移——源表存在但索引缺失（漏索引），或索引存在但源表已软删/不存在（孤儿）。
//   漂移来源：触发器遗漏、历史数据未回填、手工改库绕过触发器。
//
// 用法：
//   # 只读对账（默认）：输出三类差异报告，发现漂移时退出码为 1（便于 CI 告警）
//   go run ./scripts/reconcile-search
//
//   # 执行修复：回填缺失文档 + 清理孤儿文档（幂等，可重复执行）
//   go run ./scripts/reconcile-search --fix
//
//   # 指定数据库连接
//   go run ./scripts/reconcile-search --fix --db-dsn "postgres://..."
//
// 兼容性：
//   - 复用 search.Indexer（SyncIssue/SyncSprint/SyncVersion/RemoveDocument），
//     与 worker 增量同步共用同一套投影逻辑，保证修复结果与运行时索引一致。
//   - 对账查询在设置了 app.workspace_id 的事务内执行，RLS 启用后仍可工作。
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/application/search"
	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

var (
	flagFix   = flag.Bool("fix", false, "执行修复：回填缺失文档并清理孤儿文档")
	flagDBDSN = flag.String("db-dsn", "", "数据库连接串；空则读取 YDSZ_DATABASE_URL")
)

// driftReport 单个 (workspace, doc_type) 的漂移统计。
type driftReport struct {
	DocType string
	Missing []int64 // 源表存在但索引缺失
	Orphans []int64 // 索引存在但源表已软删/不存在
}

func main() {
	flag.Parse()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "reconcile-search:", err)
		os.Exit(2)
	}
	dsn := *flagDBDSN
	if dsn == "" {
		dsn = cfg.Database.URL
	}
	pool, err := persistence.NewPool(ctx, dsn, cfg.Database.MaxConns)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reconcile-search: connect:", err)
		os.Exit(2)
	}
	defer pool.Close()

	wsIDs, err := listWorkspaces(ctx, pool)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reconcile-search:", err)
		os.Exit(2)
	}

	indexer := search.NewIndexer(pool.Pool)
	var totalMissing, totalOrphans int
	var anyDrift bool

	fmt.Println("对账目标: search_documents vs (issues, sprints, versions)")
	fmt.Printf("模式: %s\n\n", map[bool]string{true: "FIX（执行修复）", false: "只读（默认）"}[*flagFix])

	for _, wsID := range wsIDs {
		for _, docType := range []string{"issue", "sprint", "version"} {
			rep, err := diffWorkspace(ctx, pool.Pool, wsID, docType)
			if err != nil {
				fmt.Fprintf(os.Stderr, "reconcile-search: ws=%d type=%s: %v\n", wsID, docType, err)
				os.Exit(2)
			}
			if len(rep.Missing) == 0 && len(rep.Orphans) == 0 {
				continue
			}
			anyDrift = true
			totalMissing += len(rep.Missing)
			totalOrphans += len(rep.Orphans)
			printReport(rep, wsID)
			if *flagFix {
				if err := fixDrift(ctx, indexer, rep, wsID); err != nil {
					fmt.Fprintf(os.Stderr, "reconcile-search: fix ws=%d type=%s: %v\n", wsID, docType, err)
					os.Exit(2)
				}
			}
		}
	}

	fmt.Printf("\n汇总: 缺失文档 %d 条, 孤儿文档 %d 条\n", totalMissing, totalOrphans)
	if anyDrift {
		if *flagFix {
			fmt.Println("结果: 漂移已修复 ✅")
			// fix 模式修复成功 → 最终状态一致，退出码 0
		} else {
			fmt.Println("结果: 发现漂移 ⚠️（使用 --fix 修复）")
			os.Exit(1) // CI 语义：只读模式发现漂移即失败
		}
	} else {
		fmt.Println("结果: 索引一致 ✅")
	}
}

// listWorkspaces 列出所有源表涉及的 workspace_id。
func listWorkspaces(ctx context.Context, pool *persistence.Pool) ([]int64, error) {
	rows, err := pool.Pool.Query(ctx, `
		SELECT DISTINCT workspace_id FROM (
			SELECT workspace_id FROM issues WHERE deleted_at IS NULL
			UNION SELECT workspace_id FROM sprints WHERE deleted_at IS NULL
			UNION SELECT workspace_id FROM versions WHERE deleted_at IS NULL
			UNION SELECT DISTINCT workspace_id FROM search_documents
		) t ORDER BY 1`)
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// diffWorkspace 在设置了租户上下文的事务内查询单个 (workspace, doc_type) 的漂移。
func diffWorkspace(ctx context.Context, db *pgxpool.Pool, wsID int64, docType string) (*driftReport, error) {
	table := sourceTableFor(docType)
	rep := &driftReport{DocType: docType}

	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 租户上下文（RLS 兼容；无 RLS 时无副作用）
	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)",
		strconv.FormatInt(wsID, 10)); err != nil {
		return nil, fmt.Errorf("set tenant context: %w", err)
	}

	// 缺失：源表存在（未软删）但索引无
	missingSQL := fmt.Sprintf(
		"SELECT src.id FROM %s src WHERE src.workspace_id = $1 AND src.deleted_at IS NULL "+
			"AND NOT EXISTS (SELECT 1 FROM search_documents d "+
			"WHERE d.workspace_id = $1 AND d.doc_type = $2 AND d.doc_id = src.id) ORDER BY src.id",
		table)
	rows, err := tx.Query(ctx, missingSQL, wsID, docType)
	if err != nil {
		return nil, fmt.Errorf("query missing: %w", err)
	}
	rep.Missing, err = scanIDs(rows)
	rows.Close()
	if err != nil {
		return nil, fmt.Errorf("scan missing: %w", err)
	}

	// 孤儿：索引有但源表无/已软删
	orphanSQL := fmt.Sprintf(
		"SELECT d.doc_id FROM search_documents d WHERE d.workspace_id = $1 AND d.doc_type = $2 "+
			"AND NOT EXISTS (SELECT 1 FROM %s src "+
			"WHERE src.id = d.doc_id AND src.workspace_id = $1 AND src.deleted_at IS NULL) ORDER BY d.doc_id",
		table)
	rows, err = tx.Query(ctx, orphanSQL, wsID, docType)
	if err != nil {
		return nil, fmt.Errorf("query orphans: %w", err)
	}
	rep.Orphans, err = scanIDs(rows)
	rows.Close()
	if err != nil {
		return nil, fmt.Errorf("scan orphans: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return rep, nil
}

// fixDrift 修复漂移：缺失文档逐条 upsert（幂等），孤儿文档逐条删除。
func fixDrift(ctx context.Context, indexer *search.Indexer, rep *driftReport, wsID int64) error {
	for _, id := range rep.Missing {
		var err error
		switch rep.DocType {
		case "issue":
			err = indexer.SyncIssue(ctx, id)
		case "sprint":
			err = indexer.SyncSprint(ctx, id)
		case "version":
			err = indexer.SyncVersion(ctx, id)
		}
		if err != nil {
			return fmt.Errorf("sync %s %d: %w", rep.DocType, id, err)
		}
	}
	for _, id := range rep.Orphans {
		if err := indexer.RemoveDocument(ctx, rep.DocType, wsID, id); err != nil {
			return fmt.Errorf("remove %s %d: %w", rep.DocType, id, err)
		}
	}
	return nil
}

// printReport 打印单个 workspace 的漂移明细。
func printReport(rep *driftReport, wsID int64) {
	fmt.Printf("workspace=%d doc_type=%s\n", wsID, rep.DocType)
	if len(rep.Missing) > 0 {
		fmt.Printf("  缺失(源有索引无): %d 条 -> %v\n", len(rep.Missing), rep.Missing)
	}
	if len(rep.Orphans) > 0 {
		fmt.Printf("  孤儿(索引有源无): %d 条 -> %v\n", len(rep.Orphans), rep.Orphans)
	}
}

// sourceTableFor 将 doc_type 映射为源表名。
func sourceTableFor(docType string) string {
	switch docType {
	case "issue":
		return "issues"
	case "sprint":
		return "sprints"
	case "version":
		return "versions"
	}
	return ""
}

// scanIDs 将单列 bigint 结果集读为 []int64。
func scanIDs(rows pgx.Rows) ([]int64, error) {
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
