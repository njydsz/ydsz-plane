// Command reindex 全量重建搜索索引（search_documents）并清理孤儿文档。
//
// 用途（修复 Makefile reindex target 悬空；S12 收口）：
//   1) 遍历全部工作空间，将 issues / sprints / versions 中未软删的记录
//      逐条经 search.Indexer upsert 进 search_documents（全量重建）；
//   2) 清理索引中存在但源表已软删/不存在的孤儿文档。
//
// 与 reconcile-search 的区别：
//   - reconcile-search 只处理"漂移"（差异增量修复，常用于日常巡检）；
//   - reindex 全量重建（忽略现有索引，逐条重写），常用于 schema 变更、
//     索引逻辑升级、误删恢复后的整体重建。
//
// 用法：
//   go run ./scripts/reindex                          # 全部对象类型重建
//   go run ./scripts/reindex --doc-type issue         # 仅重建工作项索引
//   go run ./scripts/reindex --db-dsn "postgres://..." --no-clean-orphans
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/application/search"
	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

var (
	flagDocType   = flag.String("doc-type", "", "仅重建指定类型: issue|sprint|version（空=全部）")
	flagDBDSN     = flag.String("db-dsn", "", "数据库连接串；空则读取 YDSZ_DATABASE_URL")
	flagClean     = flag.Bool("clean-orphans", true, "重建后清理孤儿文档（索引存在但源表已删除）")
	flagBatchSize = flag.Int("batch", 500, "每个工作空间每类文档的单批重建数量")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "reindex:", err)
		os.Exit(2)
	}
	dsn := *flagDBDSN
	if dsn == "" {
		dsn = cfg.Database.URL
	}

	pool, err := persistence.NewPool(ctx, dsn, cfg.Database.MaxConns)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reindex: connect:", err)
		os.Exit(2)
	}
	defer pool.Close()

	docTypes := []string{"issue", "sprint", "version"}
	if *flagDocType != "" {
		if !validDocType(*flagDocType) {
			fmt.Fprintf(os.Stderr, "reindex: 未知 doc-type %q（支持 issue|sprint|version）\n", *flagDocType)
			os.Exit(2)
		}
		docTypes = []string{*flagDocType}
	}

	wsIDs, err := listWorkspaces(ctx, pool.Pool)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reindex:", err)
		os.Exit(2)
	}

	indexer := search.NewIndexer(pool.Pool)
	start := time.Now()
	totalUpserted, totalOrphans := 0, 0

	fmt.Printf("全量重建搜索索引: 工作空间 %d 个, 类型 %v, 清理孤儿=%v\n\n",
		len(wsIDs), docTypes, *flagClean)

	for _, wsID := range wsIDs {
		for _, docType := range docTypes {
			n, err := rebuildWorkspace(ctx, pool.Pool, indexer, wsID, docType)
			if err != nil {
				fmt.Fprintf(os.Stderr, "reindex: ws=%d type=%s: %v\n", wsID, docType, err)
				os.Exit(2)
			}
			totalUpserted += n
			fmt.Printf("  ws=%-6d %-8s 重建 %d 条\n", wsID, docType, n)

			if *flagClean {
				orphans, err := removeOrphans(ctx, pool.Pool, indexer, wsID, docType)
				if err != nil {
					fmt.Fprintf(os.Stderr, "reindex: ws=%d type=%s cleanup: %v\n", wsID, docType, err)
					os.Exit(2)
				}
				totalOrphans += orphans
				if orphans > 0 {
					fmt.Printf("  ws=%-6d %-8s 清理孤儿 %d 条\n", wsID, docType, orphans)
				}
			}
		}
	}

	fmt.Printf("\n汇总: upsert %d 条, 清理孤儿 %d 条, 耗时 %s ✅\n",
		totalUpserted, totalOrphans, time.Since(start).Round(time.Millisecond))
}

func validDocType(t string) bool {
	switch t {
	case "issue", "sprint", "version":
		return true
	}
	return false
}

func listWorkspaces(ctx context.Context, db *pgxpool.Pool) ([]int64, error) {
	rows, err := db.Query(ctx, `
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

// rebuildWorkspace 在租户上下文事务内取源表全部 ID，逐条 upsert 索引。
func rebuildWorkspace(ctx context.Context, db *pgxpool.Pool, indexer *search.Indexer, wsID int64, docType string) (int, error) {
	table := sourceTableFor(docType)

	// 先取全量 ID（不设事务，仅查询）
	rows, err := db.Query(ctx, fmt.Sprintf(
		"SELECT id FROM %s WHERE workspace_id = $1 AND deleted_at IS NULL ORDER BY id", table), wsID)
	if err != nil {
		return 0, err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	count := 0
	for _, id := range ids {
		var syncErr error
		switch docType {
		case "issue":
			syncErr = indexer.SyncIssue(ctx, id)
		case "sprint":
			syncErr = indexer.SyncSprint(ctx, id)
		case "version":
			syncErr = indexer.SyncVersion(ctx, id)
		}
		if syncErr != nil {
			return count, fmt.Errorf("sync %s %d: %w", docType, id, syncErr)
		}
		count++
	}
	return count, nil
}

// removeOrphans 删除索引中存在但源表已软删/不存在的文档。
func removeOrphans(ctx context.Context, db *pgxpool.Pool, indexer *search.Indexer, wsID int64, docType string) (int, error) {
	table := sourceTableFor(docType)
	rows, err := db.Query(ctx, fmt.Sprintf(
		"SELECT d.doc_id FROM search_documents d WHERE d.workspace_id = $1 AND d.doc_type = $2 "+
			"AND NOT EXISTS (SELECT 1 FROM %s src WHERE src.id = d.doc_id "+
			"AND src.workspace_id = $1 AND src.deleted_at IS NULL) ORDER BY d.doc_id",
		table), wsID, docType)
	if err != nil {
		return 0, err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	for _, id := range ids {
		if err := indexer.RemoveDocument(ctx, docType, wsID, id); err != nil {
			return 0, fmt.Errorf("remove %s %d: %w", docType, id, err)
		}
	}
	return len(ids), nil
}

func sourceTableFor(docType string) string {
	switch docType {
	case "sprint":
		return "sprints"
	case "version":
		return "versions"
	default:
		return "issues"
	}
}

// 保持对 pgx 的显式引用（确保依赖不漂移）。
var _ = pgx.ErrNoRows
