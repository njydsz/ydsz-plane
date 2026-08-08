// Package dashboard 仪表盘集成测试。
//
// 覆盖 dashboard.Service 各 widget 查询的 SQL 正确性（9.2 P0 五类卡片 + 扩展卡片）。
// 这类方法全部依赖数据库，采用"环境变量 gate"集成测试模式（大厂标准）：
//   - 设置 YDSZ_TEST_DATABASE_URL 时真实运行（CI 提供测试库）
//   - 未设置时自动 Skip（本地无库不阻塞）
//
// 测试自建租户数据（workspace/project/state/issues），与既有数据隔离。
package dashboard

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

// TestWidgetQueriesIntegration 验证全部 widget 数据查询无 SQL 错误并返回非空结构。
func TestWidgetQueriesIntegration(t *testing.T) {
	dsn := os.Getenv("YDSZ_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("YDSZ_TEST_DATABASE_URL not set; skipping dashboard integration test")
	}
	ctx := context.Background()

	pool, err := persistence.NewPool(ctx, dsn, 4)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	svc := NewService(pool.Pool)
	wsID, projID := seedDashboardData(t, ctx, pool)
	// LIFO：先清理测试数据，最后关闭连接池，避免 "closed pool" 警告。
	t.Cleanup(func() { cleanupDashboardData(t, ctx, pool, wsID, projID) })
	t.Cleanup(func() { pool.Close() })

	widgets := []struct {
		name string
		fn   func(context.Context, int64) (any, error)
	}{
		{"progress_overview", svc.getProgressOverview},
		{"priority_split", svc.getPrioritySplit},
		{"state_distribution", svc.getStateDistribution},
		{"overdue_list", svc.getOverdueList},
		{"blocked_list", svc.getBlockedList},
		{"active_sprint_burndown", svc.getActiveSprintBurndown},
		{"team_workload", svc.getTeamWorkload},
		{"recent_activity", svc.getRecentActivity},
		{"velocity", svc.getVelocity},
		{"active_alerts", func(ctx context.Context, projectID int64) (any, error) {
			return svc.getActiveAlerts(ctx, projectID)
		}},
	}
	for _, w := range widgets {
		t.Run(w.name, func(t *testing.T) {
			data, err := w.fn(ctx, projID)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", w.name, err)
			}
			if data == nil {
				t.Fatalf("%s: nil data returned", w.name)
			}
		})
	}
}

// TestGetDashboardIntegration 验证 GetDashboard 聚合入口（widget 配置 + 快照 + 告警）。
func TestGetDashboardIntegration(t *testing.T) {
	dsn := os.Getenv("YDSZ_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("YDSZ_TEST_DATABASE_URL not set; skipping dashboard integration test")
	}
	ctx := context.Background()

	pool, err := persistence.NewPool(ctx, dsn, 4)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	svc := NewService(pool.Pool)
	wsID, projID := seedDashboardData(t, ctx, pool)
	// LIFO：先清理测试数据，最后关闭连接池，避免 "closed pool" 警告。
	t.Cleanup(func() { cleanupDashboardData(t, ctx, pool, wsID, projID) })
	t.Cleanup(func() { pool.Close() })

	data, err := svc.GetDashboard(ctx, wsID, projID)
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	if data == nil {
		t.Fatal("GetDashboard: nil result")
	}
}

// --- 测试数据辅助 ---

// seedDashboardData 插入独立租户的 project/state/issues。
func seedDashboardData(t *testing.T, ctx context.Context, pool *persistence.Pool) (int64, int64) {
	t.Helper()
	// slug 唯一：避免并发/重复运行冲突
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	// 确保 FK 引用的测试用户（id=1）存在；GENERATED ALWAYS 需 OVERRIDING SYSTEM VALUE。
	if _, err := pool.Pool.Exec(ctx, `
		INSERT INTO users (id, public_id, email, password_hash, display_name, is_active, timezone, created_at, updated_at)
		OVERRIDING SYSTEM VALUE
		VALUES (1, gen_random_uuid(), 'dash-int-test-user@ydsz.dev', 'seed', 'Dash Test User', true, 'Asia/Shanghai', now(), now())
		ON CONFLICT (id) DO NOTHING`); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	var wsID, projID, stateID int64
	if err := pool.Pool.QueryRow(ctx, `
		INSERT INTO workspaces (name, slug, owner_id) VALUES ('dash-int-test', $1, 1)
		RETURNING id`, "dash-int-"+suffix).Scan(&wsID); err != nil {
		t.Fatalf("seed workspace: %v", err)
	}
	if err := pool.Pool.QueryRow(ctx, `
		INSERT INTO projects (workspace_id, name, slug, identifier, created_by)
		VALUES ($1, 'Dash Int Proj', $2, $3, 1) RETURNING id`,
		wsID, "dash-int-proj-"+suffix, "DIP"+suffix[len(suffix)-4:]).Scan(&projID); err != nil {
		t.Fatalf("seed project: %v", err)
	}
	if err := pool.Pool.QueryRow(ctx, `
		INSERT INTO states (workspace_id, project_id, name, "group") VALUES ($1, $2, '待办', 'backlog')
		RETURNING id`, wsID, projID).Scan(&stateID); err != nil {
		t.Fatalf("seed state: %v", err)
	}
	// 三条 issue：不同优先级/状态（type_code 需满足 issues_type_code_check；defect 需 severity+found_phase）
	types := []string{"requirement", "task", "defect"}
	priorities := []string{"urgent", "high", "medium"}
	issueIDs := make([]int64, 0, 3)
	for i, typ := range types {
		var id int64
		err := pool.Pool.QueryRow(ctx, `
			INSERT INTO issues (workspace_id, project_id, sequence_id, type_code, name, state_id, priority,
			                    severity, found_phase, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7,
			        CASE WHEN $4 = 'defect' THEN 2 ELSE NULL END,
			        CASE WHEN $4 = 'defect' THEN 'integration'::text ELSE NULL END, 1)
			RETURNING id`,
			wsID, projID, i+1, typ, "Dash Issue "+typ, stateID, priorities[i]).Scan(&id)
		if err != nil {
			t.Fatalf("seed issue %d: %v", i, err)
		}
		issueIDs = append(issueIDs, id)
	}
	// 迭代数据：active + completed 各一个（驱动 burndown/velocity）
	var activeSprintID, doneSprintID int64
	if err := pool.Pool.QueryRow(ctx, `
		INSERT INTO sprints (workspace_id, project_id, name, status, created_by)
		VALUES ($1, $2, 'Active Sprint', 'active', 1) RETURNING id`, wsID, projID).Scan(&activeSprintID); err != nil {
		t.Fatalf("seed active sprint: %v", err)
	}
	if err := pool.Pool.QueryRow(ctx, `
		INSERT INTO sprints (workspace_id, project_id, name, status, created_by, completed_at)
		VALUES ($1, $2, 'Done Sprint', 'completed', 1, now()) RETURNING id`, wsID, projID).Scan(&doneSprintID); err != nil {
		t.Fatalf("seed done sprint: %v", err)
	}
	// sprint_issues：全部 issue 加入两个迭代（驱动 burndown/velocity 计数）
	for _, id := range issueIDs {
		if _, err := pool.Pool.Exec(ctx, `
			INSERT INTO sprint_issues (sprint_id, issue_id) VALUES ($1, $2), ($3, $4)`,
			activeSprintID, id, doneSprintID, id); err != nil {
			t.Fatalf("seed sprint_issues %d: %v", id, err)
		}
	}
	// 指派：用户 1 负责第一条 issue（驱动 team_workload）
	if _, err := pool.Pool.Exec(ctx, `
		INSERT INTO issue_assignees (issue_id, user_id, assigned_by) VALUES ($1, 1, 1)`,
		issueIDs[0]); err != nil {
		t.Fatalf("seed assignee: %v", err)
	}
	// 活动记录（驱动 recent_activity）
	if _, err := pool.Pool.Exec(ctx, `
		INSERT INTO issue_activities (workspace_id, project_id, issue_id, verb, actor_id, actor_name)
		VALUES ($1, $2, $3, 'created', 1, 'Admin')`,
		wsID, projID, issueIDs[0]); err != nil {
		t.Fatalf("seed activity: %v", err)
	}
	return wsID, projID
}

// cleanupDashboardData 删除测试租户数据（按 workspace 级联清理）。
func cleanupDashboardData(t *testing.T, ctx context.Context, pool *persistence.Pool, wsID, projID int64) {
	t.Helper()
	if _, err := pool.Pool.Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, wsID); err != nil {
		t.Logf("cleanup workspace %d: %v", wsID, err)
	}
}
