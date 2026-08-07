// 大规模造数脚本 — 用于 S7 性能基线压测。
//
// 目标：在已有种子数据基础上，批量生成 100 万工作项，供 k6 压测使用。
// 用法：
//   go run ./scripts/seed-scale -count=1000000 -project=1
//
// 特性：
//   - 批量 INSERT（每批 1000 行），降低网络往返
//   - 并发写入（可配置 worker 数），缩短造数时间
//   - 随机数据分布：type(epic/story/task/bug)、priority、state 均匀分布
//   - 可恢复：记录最后写入的 issue_id，支持断点续传
//   - 进度日志：每 10 万条输出耗时
//
// 运行前提：
//   - 数据库已执行 make migrate && make seed（至少 1 个项目存在）
//   - 项目状态模板已初始化（states 表非空）
package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

var (
	count   = flag.Int("count", 1000000, "生成的工作项总数")
	project = flag.Int64("project", 1, "目标项目 ID")
	batch   = flag.Int("batch", 1000, "每批 INSERT 行数")
	workers = flag.Int("workers", 8, "并发写入 worker 数")
	resume  = flag.Int64("resume", 0, "从指定 issue_id 续传（0=从头开始）")
)

// issueType 工作项类型。
var issueTypes = []string{"epic", "story", "task", "bug"}

// priorities 优先级列表。
var priorities = []string{"critical", "high", "medium", "low"}

// severities 严重级别（仅 bug 使用）。
var severities = []int{1, 2, 3, 4, 5}

func main() {
	flag.Parse()

	fmt.Printf("══════════════════════════════════════════════\n")
	fmt.Printf("  Ydsz Plane — 大规模造数脚本\n")
	fmt.Printf("  目标: %d 条工作项 | 项目: %d | 并发: %d\n", *count, *project, *workers)
	fmt.Printf("══════════════════════════════════════════════\n\n")

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seed-scale: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("加载配置: %w", err)
	}

	ctx := context.Background()
	pool, err := persistence.NewPool(ctx, cfg.Database.URL, cfg.Database.MaxConns)
	if err != nil {
		return fmt.Errorf("连接数据库: %w", err)
	}
	defer pool.Close()

	// 验证项目存在
	var projName string
	if err := pool.QueryRow(ctx,
		`SELECT name FROM projects WHERE id = $1`, *project).Scan(&projName); err != nil {
		return fmt.Errorf("项目 %d 不存在: %w", *project, err)
	}
	fmt.Printf("项目: %s (id=%d)\n\n", projName, *project)

	// 读取现有 states，确保状态字段有效
	states, err := loadStates(ctx, pool, *project)
	if err != nil {
		return fmt.Errorf("加载状态模板: %w", err)
	}
	if len(states) == 0 {
		return fmt.Errorf("项目 %d 无可用状态模板，请先执行 make seed", *project)
	}
	fmt.Printf("状态模板: %d 个\n\n", len(states))

	// 获取 creator_id
	var creatorID int64
	if err := pool.QueryRow(ctx,
		`SELECT id FROM users WHERE email = 'admin@ydsz.dev'`).Scan(&creatorID); err != nil {
		return fmt.Errorf("查找 admin 用户: %w", err)
	}

	start := time.Now()
	nextID := atomic.Int64{}
	nextID.Store(*resume)
	inserted := atomic.Int64{}
	lastLog := atomic.Int64{}

	// 分配任务到 worker
	total := int64(*count)
	var wg sync.WaitGroup

	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			buf := make([]string, 0, *batch)
			bufArgs := make([]interface{}, 0, *batch*10)

			for {
				id := nextID.Add(1)
				if id > total {
					break
				}

				// 随机生成一条工作项数据
				typ := issueTypes[rng.Intn(len(issueTypes))]
				name := fmt.Sprintf("[perf-%d] 压测工作项 %d", *project, id)
				priority := priorities[rng.Intn(len(priorities))]
				stateID := states[rng.Intn(len(states))]
				point := rng.Intn(13) + 1
				daysAgo := rng.Intn(90)
				createdAt := time.Now().AddDate(0, 0, -daysAgo)

				buf = append(buf, fmt.Sprintf(
					"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
					len(bufArgs)+1, len(bufArgs)+2, len(bufArgs)+3, len(bufArgs)+4,
					len(bufArgs)+5, len(bufArgs)+6, len(bufArgs)+7, len(bufArgs)+8,
					len(bufArgs)+9, len(bufArgs)+10, len(bufArgs)+11, len(bufArgs)+12,
					len(bufArgs)+13,
				))

				bufArgs = append(bufArgs,
					*project, typ, name, fmt.Sprintf("PROJ-%d", id),
					priority, nil, nil, nil, stateID,
					creatorID, createdAt, createdAt, point,
				)

				// 每批满时写入
				if len(buf) >= *batch {
					if err := batchInsert(ctx, pool, buf, bufArgs); err != nil {
						fmt.Fprintf(os.Stderr, "worker %d batch insert error: %v\n", workerID, err)
						return
					}
					n := int64(len(buf))
					inserted.Add(n)
					totalDone := inserted.Load()

					// 每 10 万条输出一次进度
					if totalDone-lastLog.Load() >= 100000 {
						lastLog.Store(totalDone)
						elapsed := time.Since(start).Seconds()
						rate := float64(totalDone) / elapsed
						fmt.Printf("  进度: %d / %d (%.1f%%) | %.0f 条/秒 | 耗时 %.1fs\n",
							totalDone, total, float64(totalDone)/float64(total)*100,
							rate, elapsed)
					}

					buf = make([]string, 0, *batch)
					bufArgs = make([]interface{}, 0, *batch*10)
				}
			}

			// 最后一批残余数据
			if len(buf) > 0 {
				if err := batchInsert(ctx, pool, buf, bufArgs); err != nil {
					fmt.Fprintf(os.Stderr, "worker %d final batch error: %v\n", workerID, err)
					return
				}
				inserted.Add(int64(len(buf)))
			}
		}(w)
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("\n══════════════════════════════════════════════\n")
	fmt.Printf("  造数完成: %d 条 | 耗时 %s | %.0f 条/秒\n",
		inserted.Load(), elapsed.Round(time.Millisecond),
		float64(inserted.Load())/elapsed.Seconds())
	fmt.Printf("══════════════════════════════════════════════\n")

	return nil
}

// loadStates 加载项目下所有状态。
func loadStates(ctx context.Context, pool *persistence.Pool, projectID int64) ([]int64, error) {
	rows, err := pool.Query(ctx,
		`SELECT id FROM states WHERE project_id = $1 ORDER BY sort_order`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// batchInsert 批量插入工作项。
func batchInsert(ctx context.Context, pool *persistence.Pool, vals []string, args []interface{}) error {
	query := fmt.Sprintf(`
		INSERT INTO issues (
			project_id, type_code, name, identifier,
			priority, severity, parent_id, epic_id, state_id,
			created_by, created_at, updated_at, story_points
		) VALUES %s`, join(vals, ","))
	_, err := pool.Exec(ctx, query, args...)
	return err
}

// join 连接字符串切片。
func join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
