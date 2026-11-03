//go:build integration

// Package preference — 视图偏好核心链路集成测试（需要真实 PostgreSQL）。
//
// 覆盖：Save → Get → ListByUser，含默认值填充与幂等 upsert。
//
// 运行方式：
//
//	YDSZ_TEST_DATABASE_URL="postgres://ydsz:ydsz@localhost:5432/ydsz_plane_test" \
//		make migrate && \
//		go test -tags=integration ./internal/application/preference/...
//
// 未设置环境变量时自动跳过。
package preference

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

func prefIntegrationPool(t *testing.T) *persistence.Pool {
	t.Helper()
	dsn := os.Getenv("YDSZ_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("YDSZ_TEST_DATABASE_URL not set, skipping preference integration test")
	}
	pool, err := persistence.NewPool(context.Background(), dsn, 4)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return pool
}

// TestViewPreferenceUpsert 验证保存的幂等 upsert 语义。
func TestViewPreferenceUpsert(t *testing.T) {
	pool := prefIntegrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	svc := NewService(pool.Pool)
	const (
		wsID     int64 = 1
		projectID int64 = 0
		userID   int64 = 99101
	)

	// 首次保存（ViewType 为空应回退为默认值）
	first, err := svc.Save(ctx, wsID, projectID, userID, &ViewPreference{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		UserID:      userID,
		ViewType:    "",
		Layout:      "board",
		Columns:     json.RawMessage(`["id","name","status"]`),
	})
	if err != nil {
		t.Fatalf("first Save: %v", err)
	}
	if first.ID == 0 {
		t.Fatal("first Save: 返回空 ID")
	}
	if first.ViewType == "" {
		t.Error("first Save: ViewType 应被填充默认值")
	}
	t.Logf("默认 ViewType = %q", first.ViewType)

	// 第二次保存同一视图类型 → 应更新而非新建
	updated, err := svc.Save(ctx, wsID, projectID, userID, &ViewPreference{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		UserID:      userID,
		ViewType:    first.ViewType,
		Layout:      "list",
	})
	if err != nil {
		t.Fatalf("second Save: %v", err)
	}
	if updated.ID != first.ID {
		t.Errorf("upsert 应复用同一行：first=%d second=%d", first.ID, updated.ID)
	}
	if updated.Layout != "list" {
		t.Errorf("layout = %q, want %q", updated.Layout, "list")
	}
}

// TestViewPreferenceGet 验证查询。
func TestViewPreferenceGet(t *testing.T) {
	pool := prefIntegrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	svc := NewService(pool.Pool)
	const (
		wsID   int64 = 1
		userID int64 = 99102
	)

	_, err := svc.Save(ctx, wsID, 0, userID, &ViewPreference{
		WorkspaceID: wsID,
		UserID:      userID,
		ViewType:    ViewList,
		Layout:      "table",
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := svc.Get(ctx, wsID, 0, userID, ViewList)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Layout != "table" {
		t.Errorf("layout = %q, want %q", got.Layout, "table")
	}

	list, err := svc.ListByUser(ctx, wsID, 0, userID)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(list) == 0 {
		t.Error("ListByUser: 应至少返回一条记录")
	}
}

// TestViewPreferenceDifferentViewTypesAreDistinct 验证不同视图类型互不干扰。
func TestViewPreferenceDifferentViewTypesAreDistinct(t *testing.T) {
	pool := prefIntegrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	svc := NewService(pool.Pool)
	const (
		wsID   int64 = 1
		userID int64 = 99103
	)

	kanban, err := svc.Save(ctx, wsID, 0, userID, &ViewPreference{
		WorkspaceID: wsID, UserID: userID, ViewType: ViewKanban, Layout: "board",
	})
	if err != nil {
		t.Fatalf("Save kanban: %v", err)
	}
	list, err := svc.Save(ctx, wsID, 0, userID, &ViewPreference{
		WorkspaceID: wsID, UserID: userID, ViewType: ViewList, Layout: "table",
	})
	if err != nil {
		t.Fatalf("Save list: %v", err)
	}
	if kanban.ID == list.ID {
		t.Errorf("不同视图类型不应共享同一行：%d == %d", kanban.ID, list.ID)
	}
}
