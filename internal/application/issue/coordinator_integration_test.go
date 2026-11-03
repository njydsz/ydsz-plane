//go:build integration

// Package issue — Coordinator 核心链路扩展集成测试（需要真实 PostgreSQL）。
//
// 在 core_chain_integration_test.go 的需求 CRUD 链路基础上，扩展覆盖：
//   - Update（乐观锁冲突 → 409）
//   - Transition（状态流转 + 非法流转拒绝）
//   - BatchUpdate（批量优先级/状态变更）
//   - Restore（软删除后恢复）
//
// 运行方式：
//
//	YDSZ_TEST_DATABASE_URL="postgres://ydsz:ydsz@localhost:5432/ydsz_plane_test" \
//		make migrate && \
//		go test -tags=integration ./internal/application/issue/...
//
// 未设置环境变量时自动跳过。
package issue

import (
	"context"
	"strings"
	"testing"
)

// TestCoordinatorUpdateOptimisticLock 验证 Update 的乐观锁语义。
//
// 场景：
//  1. 用正确的 version 更新 → 成功，version 递增；
//  2. 用旧的 version 再次更新 → ErrVersionConflict（409）。
func TestCoordinatorUpdateOptimisticLock(t *testing.T) {
	pool := integrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	wsID, projectID, stateIDs := seedWorkspaceProjectStates(t, pool.Pool, ctx)
	svc := NewService(pool.Pool)

	created, err := svc.Create(ctx, CreateIssueInput{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		TypeCode:    TypeRequirement,
		Name:        "乐观锁测试需求",
		Priority:    PriorityHigh,
		StateID:     stateIDs[0],
		CreatedBy:   1,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// 1) 正确 version 更新
	updated, err := svc.Update(ctx, wsID, created.ID, UpdateIssueInput{
		Name:     ptrStr("更新后名称"),
		Priority: priorityPtr(PriorityUrgent),
		Version:  created.Version,
	})
	if err != nil {
		t.Fatalf("Update with correct version: %v", err)
	}
	if updated.Name != "更新后名称" {
		t.Errorf("name = %q, want %q", updated.Name, "更新后名称")
	}
	if updated.Version != created.Version+1 {
		t.Errorf("version = %d, want %d", updated.Version, created.Version+1)
	}

	// 2) 旧 version 更新 → 冲突
	_, err = svc.Update(ctx, wsID, created.ID, UpdateIssueInput{
		Name:    ptrStr("冲突更新"),
		Version: created.Version,
	})
	if err == nil {
		t.Fatal("Update with stale version should fail")
	}
	if !isVersionConflict(err) {
		t.Errorf("error = %v, want version conflict", err)
	}
}

// TestCoordinatorTransition 验证状态流转与非法流转拒绝。
func TestCoordinatorTransition(t *testing.T) {
	pool := integrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	wsID, projectID, stateIDs := seedWorkspaceProjectStates(t, pool.Pool, ctx)
	svc := NewService(pool.Pool)

	created, err := svc.Create(ctx, CreateIssueInput{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		TypeCode:    TypeTask,
		Name:        "状态流转测试任务",
		Priority:    PriorityMedium,
		StateID:     stateIDs[0],
		CreatedBy:   1,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// 1) 合法流转：待办 → 进行中
	transitioned, err := svc.Transition(ctx, wsID, projectID, created.ID, stateIDs[1], 1)
	if err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if transitioned.StateID != stateIDs[1] {
		t.Errorf("state_id = %d, want %d", transitioned.StateID, stateIDs[1])
	}

	// 2) 非法流转：不存在的状态 → 应失败
	_, err = svc.Transition(ctx, wsID, projectID, created.ID, 999999, 1)
	if err == nil {
		t.Error("Transition to non-existent state should fail")
	}
}

// TestCoordinatorBatchUpdate 验证批量更新的部分成功语义。
func TestCoordinatorBatchUpdate(t *testing.T) {
	pool := integrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	wsID, projectID, stateIDs := seedWorkspaceProjectStates(t, pool.Pool, ctx)
	svc := NewService(pool.Pool)

	// 创建 3 个工作项
	ids := make([]int64, 0, 3)
	for i := 0; i < 3; i++ {
		created, err := svc.Create(ctx, CreateIssueInput{
			WorkspaceID: wsID,
			ProjectID:   projectID,
			TypeCode:    TypeRequirement,
			Name:        "批量测试需求",
			Priority:    PriorityLow,
			StateID:     stateIDs[0],
			CreatedBy:   1,
		})
		if err != nil {
			t.Fatalf("Create #%d: %v", i, err)
		}
		ids = append(ids, created.ID)
	}

	// 批量更新优先级（含一个不存在的 ID，验证部分成功）
	urgent := string(PriorityUrgent)
	batch := BatchUpdateInput{
		IDs:      append(ids, 999999),
		Priority: &urgent,
	}
	result, err := svc.BatchUpdate(ctx, wsID, projectID, 1, batch)
	if err != nil {
		t.Fatalf("BatchUpdate: %v", err)
	}
	if result.Succeeded != 3 {
		t.Errorf("succeeded = %d, want 3", result.Succeeded)
	}
	if result.Failed != 1 {
		t.Errorf("failed = %d, want 1", result.Failed)
	}
}

// TestCoordinatorRestore 验证软删除后恢复。
func TestCoordinatorRestore(t *testing.T) {
	pool := integrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	wsID, projectID, stateIDs := seedWorkspaceProjectStates(t, pool.Pool, ctx)
	svc := NewService(pool.Pool)

	created, err := svc.Create(ctx, CreateIssueInput{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		TypeCode:    TypeDefect,
		Name:        "恢复测试缺陷",
		Priority:    PriorityHigh,
		StateID:     stateIDs[0],
		CreatedBy:   1,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// 软删除
	if err := svc.SoftDelete(ctx, wsID, created.ID); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}

	// 删除后查询应返回 NotFound
	if _, err := svc.GetByID(ctx, wsID, created.ID); err == nil {
		t.Error("GetByID after soft delete should return error")
	}

	// 恢复
	if err := svc.Restore(ctx, wsID, created.ID); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	// 恢复后应可查询
	got, err := svc.GetByID(ctx, wsID, created.ID)
	if err != nil {
		t.Fatalf("GetByID after restore: %v", err)
	}
	if got.Name != "恢复测试缺陷" {
		t.Errorf("name = %q, want %q", got.Name, "恢复测试缺陷")
	}
}

// --- helpers ---

// isVersionConflict 判断错误是否为乐观锁冲突。
func isVersionConflict(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "version") || strings.Contains(msg, "conflict")
}
