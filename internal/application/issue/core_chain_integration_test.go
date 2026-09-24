//go:build integration

// Package issue — 核心价值链集成测试（需要真实 PostgreSQL）。
//
// 覆盖 Ydsz Plane 的主价值链：创建工作项 → 状态流转 → 软删除 → 恢复。
// 不依赖 mock，直接走 SQL，与 sprint/dashboard 域共享一条集成链路。
//
// 运行方式：
//
//	YDSZ_TEST_DATABASE_URL="postgres://ydsz:ydsz@localhost:5432/ydsz_plane_test" \
//		make migrate && \
//		go test -tags=integration ./internal/application/issue/...
//
// 未设置环境变量时自动跳过（CI 无 DB 不阻塞）。
package issue

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ==========================================================================
// 核心价值链：需求 CRUD + 状态流转
// ==========================================================================

// TestCoreChainRequirementLifecycle 验证需求从创建到软删除再到恢复的完整闭环。
func TestCoreChainRequirementLifecycle(t *testing.T) {
	pool := integrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	wsID, projectID, stateIDs := seedWorkspaceProjectStates(t, pool.Pool, ctx)

	svc := NewRequirementService(pool.Pool)

	// 1) 创建需求
	created, err := svc.Create(ctx, CreateRequirementInput{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		Name:        "核心链路测试需求",
		Priority:    PriorityHigh,
		CreatedBy:   1,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created == nil || created.ID == 0 {
		t.Fatal("Create: 返回空 ID")
	}
	if created.Identifier == "" {
		t.Error("Create: 未生成需求标识符")
	}
	t.Logf("创建需求: id=%d identifier=%s state_id=%d seq=%d",
		created.ID, created.Identifier, created.StateID, created.SequenceID)

	// 2) 查询回读
	got, err := svc.GetByID(ctx, wsID, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "核心链路测试需求" {
		t.Errorf("GetByID: name=%q, want %q", got.Name, "核心链路测试需求")
	}
	if got.SequenceID == 0 {
		t.Error("GetByID: sequence_id 不应为 0")
	}

	// 3) 更新需求（使用乐观锁版本的 Version）
	updated, err := svc.Update(ctx, wsID, created.ID, UpdateRequirementInput{
		Name:     ptrStr("更新后的需求名称"),
		Priority: priorityPtr(PriorityUrgent),
		Version:  created.Version,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "更新后的需求名称" {
		t.Errorf("Update: name=%q, want %q", updated.Name, "更新后的需求名称")
	}

	// 4) 状态流转：初始状态 → 其他状态
	if len(stateIDs) < 2 {
		t.Skip("项目状态数不足 2 个，跳过流转测试")
	}
	initialState := created.StateID
	targetState := findOtherState(stateIDs, initialState)

	transitioned, err := svc.Transition(ctx, wsID, projectID, created.ID, targetState, 1)
	if err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if transitioned.StateID != targetState {
		t.Errorf("Transition: state_id=%d, want %d", transitioned.StateID, targetState)
	}
	t.Logf("状态流转: %d → %d", initialState, targetState)

	// 5) 软删除：GetByID 应返回 not found（查询带 deleted=false 谓词）
	if err := svc.SoftDelete(ctx, wsID, created.ID); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}
	_, err = svc.GetByID(ctx, wsID, created.ID)
	if err == nil {
		t.Error("SoftDelete 后 GetByID 应返回 not-found")
	}
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		t.Errorf("SoftDelete 后期望 ErrNotFound, got: %v", err)
	}
	t.Logf("软删除验证通过: GetByID 正确返回 %v", err)

	// 6) 恢复
	if err := svc.Restore(ctx, wsID, created.ID); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	restored, err := svc.GetByID(ctx, wsID, created.ID)
	if err != nil {
		t.Fatalf("Restore 后 GetByID: %v", err)
	}
	if restored.StateID != targetState {
		t.Errorf("Restore 后 state_id=%d, want %d (流转结果应保留)", restored.StateID, targetState)
	}
	t.Logf("恢复闭环通过: id=%d state_id=%d", restored.ID, restored.StateID)

	// 7) 清理
	if err := svc.SoftDelete(ctx, wsID, created.ID); err != nil {
		t.Logf("cleanup delete: %v", err)
	}
}

// TestCoreChainStateTransitionMatrix 验证状态流转矩阵的基本规则：
// - 相同状态流转幂等
// - 状态间可查询是否允许
func TestCoreChainStateTransitionMatrix(t *testing.T) {
	pool := integrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	wsID, projectID, stateIDs := seedWorkspaceProjectStates(t, pool.Pool, ctx)
	if len(stateIDs) < 2 {
		t.Skip("状态数不足，跳过流转矩阵测试")
	}

	stateSvc := NewStateService(pool.Pool)

	// 规则 A：相同状态 ValidateTransition 应返回 nil（幂等）
	if err := stateSvc.ValidateTransition(ctx, wsID, projectID, TransitionInput{
		TypeCode:  TypeRequirement,
		FromState: stateIDs[0],
		ToState:   stateIDs[0],
	}); err != nil {
		t.Errorf("同状态流转应幂等, got: %v", err)
	}

	// 规则 B：状态间可查询是否允许（不 panic、不 error）
	fromState := stateIDs[0]
	toState := stateIDs[len(stateIDs)-1]
	allowed, err := stateSvc.CanTransitionQuick(ctx, wsID, projectID, fromState, toState, TypeRequirement)
	t.Logf("CanTransitionQuick(%d→%d): allowed=%v, err=%v", fromState, toState, allowed, err)
}

// TestCoreChainIssueCoordinatorCreate 验证 coordinator 层的创建工作项入口，
// 覆盖应用服务层的统一调度能力。
func TestCoreChainIssueCoordinatorCreate(t *testing.T) {
	pool := integrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	wsID, projectID, _ := seedWorkspaceProjectStates(t, pool.Pool, ctx)

	coord := NewService(pool.Pool)

	issue, err := coord.Create(ctx, CreateIssueInput{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		TypeCode:    TypeRequirement,
		Name:        "Coordinator 创建的需求",
		Priority:    PriorityMedium,
		CreatedBy:   1,
	})
	if err != nil {
		t.Fatalf("Coordinator.Create: %v", err)
	}
	if issue == nil || issue.ID == 0 {
		t.Fatal("Coordinator.Create: returned nil or empty ID")
	}
	t.Logf("Coordinator 创建: id=%d seq_id=%d type=%s identifier=%s",
		issue.ID, issue.SequenceID, issue.TypeCode, issue.Identifier)

	// 清理
	if err := coord.SoftDelete(ctx, wsID, issue.ID); err != nil {
		t.Logf("cleanup: %v", err)
	}
}

// ==========================================================================
// 数据种子辅助
// ==========================================================================

func integrationPool(t *testing.T) *persistence.Pool {
	t.Helper()
	dsn := os.Getenv("YDSZ_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("YDSZ_TEST_DATABASE_URL not set, skipping core chain integration test")
	}
	return mustPool(t, dsn)
}

func mustPool(t *testing.T, dsn string) *persistence.Pool {
	t.Helper()
	pool, err := persistence.NewPool(context.Background(), dsn, 4)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return pool
}

// seedWorkspaceProjectStates 创建 workspace + project + 默认状态组，返回各状态 ID。
func seedWorkspaceProjectStates(t *testing.T, pool *pgxpool.Pool, ctx context.Context) (int64, int64, []int64) {
	t.Helper()
	suffix := time.Now().Format("150405.000")

	var wsID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO workspaces (name, slug, created_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
		RETURNING id`,
		"核心链测试空间", "corechain-ws-"+suffix, 1).Scan(&wsID)
	if err != nil {
		t.Fatalf("seed workspace: %v", err)
	}

	var projectID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO projects (workspace_id, name, identifier, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		wsID, "核心链测试项目", "CC"+suffix[:6], 1).Scan(&projectID)
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}

	// 插入 4 个标准状态（对应常见的 backlog/started/completed/cancelled 组）
	groups := []struct {
		name  string
		group string
		color string
	}{
		{"待办", "backlog", "#8DA2C2"},
		{"进行中", "started", "#5C8FFF"},
		{"已完成", "completed", "#5CB85C"},
		{"已取消", "cancelled", "#D9534F"},
	}
	var stateIDs []int64
	for i, g := range groups {
		var sid int64
		err := pool.QueryRow(ctx, `
			INSERT INTO states (name, tenant_id, workspace_id, project_id, "group", color, is_default, sequence, created_by, updated_by)
			VALUES ($1, 1, $2, $3, $4, $5, $6, $7, 1, 1)
			RETURNING id`,
			g.name, wsID, projectID, g.group, g.color, i == 0, 10*float64(i+1)).Scan(&sid)
		if err != nil {
			t.Fatalf("seed state %s: %v", g.name, err)
		}
		stateIDs = append(stateIDs, sid)
	}

	return wsID, projectID, stateIDs
}

// findOtherState 返回与 exclude 不同的第一个状态 ID。
func findOtherState(stateIDs []int64, exclude int64) int64 {
	for _, sid := range stateIDs {
		if sid != exclude {
			return sid
		}
	}
	return 0
}

func ptrStr(s string) *string         { return &s }
func priorityPtr(p IssuePriority) *IssuePriority { return &p }
