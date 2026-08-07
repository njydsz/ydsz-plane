//go:build integration

// Package sprint — Sprint 域集成测试（需要真实 PostgreSQL）。
//
// 运行方式：
//
//	# 启动 PostgreSQL 后执行迁移与种子数据，然后：
//	YDSZ_TEST_DB_URL="postgres://ydsz:ydsz@localhost:5432/ydsz_plane_test" \
//		go test -tags=integration ./internal/application/sprint/...
//
// 未设置 YDSZ_TEST_DB_URL 时自动跳过（CI 无 DB 不阻塞）。
package sprint

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("YDSZ_TEST_DB_URL")
	if url == "" {
		t.Skip("YDSZ_TEST_DB_URL 未设置，跳过集成测试")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("连接测试库失败: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// TestIntegrationSprintLifecycle 覆盖迭代全生命周期：
// Create → AddIssue → Start → Complete → SoftDelete 的真实 DB 流转。
func TestIntegrationSprintLifecycle(t *testing.T) {
	pool := integrationPool(t)
	svc := NewService(pool)
	ctx := context.Background()

	// 准备：插入一个最小可用的项目（依赖 projects 表与 workspace）
	wsID, projectID := seedProject(t, pool, ctx)

	// 1) 创建迭代
	sp, err := svc.Create(ctx, CreateSprintInput{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		Name:        "集成测试迭代",
		StartDate:   ptrTime(time.Now().AddDate(0, 0, -1)),
		EndDate:     ptrTime(time.Now().AddDate(0, 0, 7)),
		Capacity:    ptrFloat(10),
		CreatedBy:   1,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if sp.Status != SprintPlanned {
		t.Fatalf("创建后状态应为 planned, got %s", sp.Status)
	}

	// 2) 创建工作项（依赖 issues 表 + 状态）
	issueID, point := seedIssue(t, pool, ctx, wsID, projectID, 5)

	// 3) 加入迭代
	if err := svc.AddIssue(ctx, wsID, AddIssueInput{SprintID: sp.ID, IssueID: issueID, SortOrder: 1, AddedBy: 1}); err != nil {
		t.Fatalf("AddIssue: %v", err)
	}

	// 4) 容量校验：点数为 5，容量 10，再加一个 8 点工作项应触发超限
	bigIssue, _ := seedIssue(t, pool, ctx, wsID, projectID, 8)
	err = svc.AddIssue(ctx, wsID, AddIssueInput{SprintID: sp.ID, IssueID: bigIssue, SortOrder: 2, AddedBy: 1})
	if err == nil || !errors.Is(err, errs.ErrSprintCapacityExceeded) {
		t.Fatalf("容量超限应返回 ErrSprintCapacityExceeded, got %v", err)
	}

	// 5) 启动迭代
	started, err := svc.Start(ctx, wsID, sp.ID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if started.Status != SprintActive {
		t.Fatalf("启动后状态应为 active, got %s", started.Status)
	}

	// 6) 中途加项标记
	midIssue, _ := seedIssue(t, pool, ctx, wsID, projectID, 3)
	if err := svc.AddIssue(ctx, wsID, AddIssueInput{SprintID: sp.ID, IssueID: midIssue, SortOrder: 3, AddedBy: 1}); err != nil {
		t.Fatalf("中途 AddIssue: %v", err)
	}

	// 7) 结束迭代（退回 Backlog 策略）
	completed, err := svc.Complete(ctx, wsID, sp.ID, CompleteSprintInput{Strategy: UnfinishedBacklog})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if completed.Status != SprintCompleted {
		t.Fatalf("结束后状态应为 completed, got %s", completed.Status)
	}
	if completed.ReviewSnapshot == nil {
		t.Fatal("结束后应生成复盘快照")
	}
	if completed.ReviewSnapshot.JoinedIssues < 1 {
		t.Fatalf("中途加入的工作项应被记录, joined=%d", completed.ReviewSnapshot.JoinedIssues)
	}

	// 8) 归档（SoftDelete）
	if err := svc.SoftDelete(ctx, wsID, sp.ID); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}
	_ = point
}

// TestIntegrationSprintBurndown 验证启动快照 + 手动快照的燃尽图数据可读。
func TestIntegrationSprintBurndown(t *testing.T) {
	pool := integrationPool(t)
	svc := NewService(pool)
	ctx := context.Background()

	wsID, projectID := seedProject(t, pool, ctx)
	sp, err := svc.Create(ctx, CreateSprintInput{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		Name:        "燃尽图集成测试",
		StartDate:   ptrTime(time.Now().AddDate(0, 0, -2)),
		EndDate:     ptrTime(time.Now().AddDate(0, 0, 5)),
		CreatedBy:   1,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	issueID, _ := seedIssue(t, pool, ctx, wsID, projectID, 4)
	if err := svc.AddIssue(ctx, wsID, AddIssueInput{SprintID: sp.ID, IssueID: issueID, SortOrder: 1, AddedBy: 1}); err != nil {
		t.Fatalf("AddIssue: %v", err)
	}
	if _, err := svc.Start(ctx, wsID, sp.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// 手动触发当日快照
	if _, count := svc.SnapshotAllActive(ctx); count == 0 {
		t.Fatal("SnapshotAllActive 应处理至少 1 个 active 迭代")
	}

	_, points, err := svc.BurndownData(ctx, wsID, sp.ID)
	if err != nil {
		t.Fatalf("BurndownData: %v", err)
	}
	if len(points) == 0 {
		t.Fatal("启动后应有至少 1 个快照点")
	}
}

/* ------------------------------------------------------------------ */
/* 测试数据种子辅助                                                      */
/* ------------------------------------------------------------------ */

func seedProject(t *testing.T, pool *pgxpool.Pool, ctx context.Context) (int64, int64) {
	t.Helper()
	var wsID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO workspaces (name, slug, created_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
		RETURNING id`,
		"集成测试空间", "it-ws-"+time.Now().Format("150405.000"), 1).Scan(&wsID)
	if err != nil {
		t.Fatalf("插入 workspace: %v", err)
	}

	var projectID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO projects (workspace_id, name, identifier, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		wsID, "集成测试项目", "IT"+time.Now().Format("1504"), 1).Scan(&projectID)
	if err != nil {
		t.Fatalf("插入 project: %v", err)
	}
	return wsID, projectID
}

func seedIssue(t *testing.T, pool *pgxpool.Pool, ctx context.Context, wsID, projectID int64, point int) (int64, int) {
	t.Helper()
	// 使用项目内第一个状态作为初始状态
	var stateID int64
	err := pool.QueryRow(ctx,
		`SELECT id FROM states WHERE project_id = $1 ORDER BY id LIMIT 1`, projectID).Scan(&stateID)
	if err != nil {
		t.Fatalf("查询状态: %v", err)
	}

	var issueID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO issues (workspace_id, project_id, name, type_code, state_id, point, created_by)
		VALUES ($1, $2, $3, 'task', $4, $5, $6)
		RETURNING id`,
		wsID, projectID, "集成测试工作项"+time.Now().Format("150405.000"), stateID, point, 1).Scan(&issueID)
	if err != nil {
		t.Fatalf("插入 issue: %v", err)
	}
	return issueID, point
}

func ptrTime(t time.Time) *time.Time { return &t }
func ptrFloat(f float64) *float64    { return &f }
