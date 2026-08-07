

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/njydsz/ydsz-plane/internal/application/issue"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

func main() {
	dsn := os.Getenv("YDSZ_TEST_DATABASE_URL")
	ctx := context.Background()
	pool, err := persistence.NewPool(ctx, dsn, 4)
	if err != nil {
		fmt.Println("pool err:", err)
		return
	}
	defer pool.Close()

	svc := issue.NewService(pool.Pool)
	name := "probe-go"
	sev := 3
	phase := "integration"
	_, err = svc.Create(ctx, issue.CreateIssueInput{
		WorkspaceID:     5,
		ProjectID:       10,
		TypeCode:        issue.TypeDefect,
		Name:            name,
		DescriptionHTML: "<p>x</p>",
		Severity:        &sev,
		FoundPhase:      &phase,
		Assignees:       []int64{},
		Labels:          []int64{},
		Modules:         []int64{},
		CreatedBy:       1,
	})
	if err != nil {
		fmt.Printf("CREATE ERROR: %v\n", err)
		return
	}
	fmt.Println("CREATE OK")
}
