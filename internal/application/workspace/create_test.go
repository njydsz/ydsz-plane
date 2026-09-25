package workspace

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateWorkspace_Sequence(t *testing.T) {
	dsn := "postgres://postgres:Limw1020@127.0.0.1:5432/ydsz-plane?sslmode=disable"
	db, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)
	defer db.Close()

	svc := NewService(db)

	// Test create
	ws, err := svc.Create(context.Background(), CreateInput{
		Name:     "test-seq-create-1111",
		Slug:     "",
		OwnerID:  1,
		Timezone: "Asia/Shanghai",
	})

	if err != nil {
		t.Logf("Create error: %v", err)
		t.Fatalf("Create failed: %v", err)
	}

	require.NotNil(t, ws)
	assert.NotZero(t, ws.ID)
	assert.Equal(t, "test-seq-create-1111", ws.Name)
	t.Logf("Created workspace id=%d, slug=%s", ws.ID, ws.Slug)

	// Cleanup
	_, _ = db.Exec(context.Background(), "DELETE FROM workspace_members WHERE workspace_id = $1", ws.ID)
	_, _ = db.Exec(context.Background(), "DELETE FROM workspaces WHERE id = $1", ws.ID)
}
