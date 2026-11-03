//go:build integration

// Package intake — 收件箱核心链路集成测试（需要真实 PostgreSQL）。
//
// 覆盖：CreateChannel → GetChannel → UpdateChannel → ListChannels → DeleteChannel。
//
// 运行方式：
//
//	YDSZ_TEST_DATABASE_URL="postgres://ydsz:ydsz@localhost:5432/ydsz_plane_test" \
//		make migrate && \
//		go test -tags=integration ./internal/application/intake/...
//
// 未设置环境变量时自动跳过。
package intake

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

func intakeIntegrationPool(t *testing.T) *persistence.Pool {
	t.Helper()
	dsn := os.Getenv("YDSZ_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("YDSZ_TEST_DATABASE_URL not set, skipping intake integration test")
	}
	pool, err := persistence.NewPool(context.Background(), dsn, 4)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return pool
}

// TestIntakeChannelLifecycle 验证渠道从创建到删除的完整闭环。
func TestIntakeChannelLifecycle(t *testing.T) {
	pool := intakeIntegrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	svc := NewService(pool.Pool)
	const wsID int64 = 1
	suffix := time.Now().Format("150405.000")

	// 1) 创建渠道
	created, err := svc.CreateChannel(ctx, CreateChannelInput{
		WorkspaceID: wsID,
		Name:        "测试提报渠道",
		Slug:        "test-channel-" + suffix,
		Description: "集成测试用渠道",
	})
	if err != nil {
		t.Fatalf("CreateChannel: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("CreateChannel: 返回空 ID")
	}
	if created.Status != string(ChannelActive) {
		t.Errorf("status = %q, want %q", created.Status, ChannelActive)
	}

	// 2) 回读
	got, err := svc.GetChannel(ctx, wsID, created.ID)
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	if got.Name != "测试提报渠道" {
		t.Errorf("name = %q, want %q", got.Name, "测试提报渠道")
	}

	// 3) 按 slug 查询
	bySlug, err := svc.GetChannelBySlug(ctx, created.Slug)
	if err != nil {
		t.Fatalf("GetChannelBySlug: %v", err)
	}
	if bySlug.ID != created.ID {
		t.Errorf("id = %d, want %d", bySlug.ID, created.ID)
	}

	// 4) 更新
	newName := "更新后的渠道名称"
	newDesc := "更新后的描述"
	updated, err := svc.UpdateChannel(ctx, UpdateChannelInput{
		ID:          created.ID,
		WorkspaceID: wsID,
		Name:        &newName,
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("UpdateChannel: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("name = %q, want %q", updated.Name, newName)
	}

	// 5) 列表
	list, err := svc.ListChannels(ctx, wsID, nil, false)
	if err != nil {
		t.Fatalf("ListChannels: %v", err)
	}
	found := false
	for _, c := range list {
		if c.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ListChannels: 未找到刚创建的渠道 id=%d", created.ID)
	}

	// 6) 删除
	if err := svc.DeleteChannel(ctx, wsID, created.ID); err != nil {
		t.Fatalf("DeleteChannel: %v", err)
	}
}

// TestIntakeChannelDuplicateSlug 验证同工作空间 slug 唯一约束。
func TestIntakeChannelDuplicateSlug(t *testing.T) {
	pool := intakeIntegrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	svc := NewService(pool.Pool)
	const wsID int64 = 1
	suffix := time.Now().Format("150405.000")
	slug := "dup-slug-" + suffix

	if _, err := svc.CreateChannel(ctx, CreateChannelInput{
		WorkspaceID: wsID,
		Name:        "首个渠道",
		Slug:        slug,
	}); err != nil {
		t.Fatalf("first CreateChannel: %v", err)
	}

	// 同 slug 第二次创建应失败（唯一约束）。
	if _, err := svc.CreateChannel(ctx, CreateChannelInput{
		WorkspaceID: wsID,
		Name:        "重复渠道",
		Slug:        slug,
	}); err == nil {
		t.Error("second CreateChannel with duplicate slug should fail")
	}
}
