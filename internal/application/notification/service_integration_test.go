//go:build integration

// Package notification — 通知核心链路集成测试（需要真实 PostgreSQL）。
//
// 覆盖：Create → List → MarkRead → UnreadCount → Archive。
//
// 运行方式：
//
//	YDSZ_TEST_DATABASE_URL="postgres://ydsz:ydsz@localhost:5432/ydsz_plane_test" \
//		make migrate && \
//		go test -tags=integration ./internal/application/notification/...
//
// 未设置环境变量时自动跳过。
package notification

import (
	"context"
	"os"
	"testing"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

func notifIntegrationPool(t *testing.T) *persistence.Pool {
	t.Helper()
	dsn := os.Getenv("YDSZ_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("YDSZ_TEST_DATABASE_URL not set, skipping notification integration test")
	}
	pool, err := persistence.NewPool(context.Background(), dsn, 4)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return pool
}

// TestNotificationLifecycle 验证通知从创建到归档的完整闭环。
func TestNotificationLifecycle(t *testing.T) {
	pool := notifIntegrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	svc := NewService(pool.Pool)
	const wsID int64 = 1
	const recipientID int64 = 99001

	// 1) 创建通知
	created, err := svc.Create(ctx, CreateNotificationInput{
		WorkspaceID: wsID,
		RecipientID: recipientID,
		EventType:   EventIssueAssigned,
		EntityType:  EntityIssue,
		EntityID:    12345,
		Title:       "测试通知标题",
		Body:        "测试通知正文",
		ActionURL:   "/issues/12345",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("Create: 返回空 ID")
	}
	if created.IsRead {
		t.Error("Create: 新通知不应为已读")
	}
	if created.IsArchived {
		t.Error("Create: 新通知不应为已归档")
	}

	// 2) 列表查询（应包含刚创建的通知）
	list, err := svc.List(ctx, ListInput{
		WorkspaceID: wsID,
		RecipientID: recipientID,
		Limit:       50,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, n := range list.Items {
		if n.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("List: 未找到刚创建的通知 id=%d", created.ID)
	}

	// 3) 未读数应为 1
	count, err := svc.UnreadCount(ctx, wsID, recipientID)
	if err != nil {
		t.Fatalf("UnreadCount: %v", err)
	}
	if count != 1 {
		t.Errorf("unread count = %d, want 1", count)
	}

	// 4) 标记已读
	if err := svc.MarkRead(ctx, created.ID, recipientID); err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
	count, err = svc.UnreadCount(ctx, wsID, recipientID)
	if err != nil {
		t.Fatalf("UnreadCount after read: %v", err)
	}
	if count != 0 {
		t.Errorf("unread count after read = %d, want 0", count)
	}

	// 5) 归档
	if err := svc.Archive(ctx, created.ID, recipientID); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	// 6) 归档后默认列表不应再返回该通知
	listAfter, err := svc.List(ctx, ListInput{
		WorkspaceID: wsID,
		RecipientID: recipientID,
		Limit:       50,
	})
	if err != nil {
		t.Fatalf("List after archive: %v", err)
	}
	for _, n := range listAfter.Items {
		if n.ID == created.ID {
			t.Error("List after archive: 已归档通知不应出现在默认列表中")
		}
	}
}

// TestNotificationCreateBatch 验证批量创建的部分成功语义。
func TestNotificationCreateBatch(t *testing.T) {
	pool := notifIntegrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	svc := NewService(pool.Pool)
	const wsID int64 = 1

	inputs := []CreateNotificationInput{
		{
			WorkspaceID: wsID, RecipientID: 1, EventType: EventIssueCreated,
			EntityType: EntityIssue, EntityID: 1, Title: "批量 1", Body: "b1",
		},
		{
			WorkspaceID: wsID, RecipientID: 2, EventType: EventIssueCreated,
			EntityType: EntityIssue, EntityID: 2, Title: "批量 2", Body: "b2",
		},
	}

	n, err := svc.CreateBatch(ctx, inputs)
	if err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}
	if n != 2 {
		t.Errorf("created = %d, want 2", n)
	}
}

// TestNotificationMarkAllRead 验证一键已读。
func TestNotificationMarkAllRead(t *testing.T) {
	pool := notifIntegrationPool(t)
	defer pool.Close()
	ctx := context.Background()

	svc := NewService(pool.Pool)
	const wsID int64 = 1
	const recipientID int64 = 99002

	for i := 0; i < 3; i++ {
		_, err := svc.Create(ctx, CreateNotificationInput{
			WorkspaceID: wsID, RecipientID: recipientID, EventType: EventCommentCreated,
			EntityType: EntityComment, EntityID: int64(i + 1),
			Title: "评论通知", Body: "body",
		})
		if err != nil {
			t.Fatalf("Create #%d: %v", i, err)
		}
	}

	affected, err := svc.MarkAllRead(ctx, wsID, recipientID)
	if err != nil {
		t.Fatalf("MarkAllRead: %v", err)
	}
	if affected != 3 {
		t.Errorf("affected = %d, want 3", affected)
	}

	count, err := svc.UnreadCount(ctx, wsID, recipientID)
	if err != nil {
		t.Fatalf("UnreadCount: %v", err)
	}
	if count != 0 {
		t.Errorf("unread count = %d, want 0", count)
	}
}
