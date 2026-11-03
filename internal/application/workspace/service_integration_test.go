//go:build integration

// Package workspace — 工作空间核心链路集成测试（需要真实 PostgreSQL）。
//
// 覆盖：Create → Get → Update → ListByUser → Archive。
//
// 运行方式：
//
//	YDSZ_TEST_DATABASE_URL="postgres://ydsz:ydsz@localhost:5432/ydsz_plane_test" \
//		make migrate && \
//		go test -tags=integration ./internal/application/workspace/...
//
// 未设置环境变量时自动跳过。
package workspace

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

func wsIntegrationPool(t *testing.T) *persistence.Pool {
	t.Helper()
	dsn := os.Getenv("YDSZ_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("YDSZ_TEST_DATABASE_URL not set, skipping workspace integration test")
	}
	pool, err := persistence.NewPool(context.Background(), dsn, 4)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return pool
}

// TestWorkspaceLifecycle 验证工作空间从创建到归档的完整闭环。
func TestWorkspaceLifecycle(t *testing.T) {
	pool := wsIntegrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	svc := NewService(pool.Pool)
	suffix := time.Now().Format("150405.000")

	// 1) 创建
	created, err := svc.Create(ctx, CreateInput{
		Name:     "集成测试空间",
		Slug:     "ws-integration-" + suffix,
		Timezone: "Asia/Shanghai",
		Language: "zh-CN",
		OwnerID:  1,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("Create: 返回空 ID")
	}
	if created.Name != "集成测试空间" {
		t.Errorf("name = %q, want %q", created.Name, "集成测试空间")
	}
	t.Logf("创建工作空间: id=%d slug=%s", created.ID, created.Slug)

	// 2) 回读
	got, err := svc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Slug != created.Slug {
		t.Errorf("slug = %q, want %q", got.Slug, created.Slug)
	}

	// 3) 更新
	newName := "更新后的空间名称"
	newTZ := "UTC"
	updated, err := svc.Update(ctx, created.ID, UpdateInput{
		Name:     &newName,
		Timezone: &newTZ,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("name = %q, want %q", updated.Name, newName)
	}
	if updated.Timezone != newTZ {
		t.Errorf("timezone = %q, want %q", updated.Timezone, newTZ)
	}

	// 4) ListByUser
	list, err := svc.ListByUser(ctx, created.OwnerID)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	found := false
	for _, w := range list {
		if w.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ListByUser: 未找到刚创建的空间 id=%d", created.ID)
	}

	// 5) Archive
	if err := svc.Archive(ctx, created.ID); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	archived, err := svc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get after archive: %v", err)
	}
	if archived.Status != "archived" {
		t.Errorf("status = %q, want %q", archived.Status, "archived")
	}
}

// TestWorkspaceGetBySlug 验证按 slug 查询。
func TestWorkspaceGetBySlug(t *testing.T) {
	pool := wsIntegrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	svc := NewService(pool.Pool)
	suffix := time.Now().Format("150405.000")

	created, err := svc.Create(ctx, CreateInput{
		Name:    "Slug 测试空间",
		Slug:    "ws-slug-" + suffix,
		OwnerID: 1,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := svc.GetBySlug(ctx, created.Slug)
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("id = %d, want %d", got.ID, created.ID)
	}
}
