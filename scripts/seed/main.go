// Command seed 向数据库插入种子数据，支持两种模式：
//
// 1) 确定性种子模式（默认）：插入测试账号、工作空间、成员关系、审计日志（幂等）。
// 2) 性能压测造数模式（--count N）：批量生成 N 条工作项，用于负载测试基线建立。
//
// 使用示例：
//   # 确定性种子
//   go run ./scripts/seed
//
//   # 造 10 万工作项
//   go run ./scripts/seed --count 100000
//
//   # 指定数据库连接
//   go run ./scripts/seed --count 100000 --db-dsn "postgres://..." --batch-size 2000
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

// ----- 命令行参数 -----
var (
	flagCount     = flag.Int("count", 0, "批量造工作项数量；0 表示运行确定性种子模式")
	flagBatchSize = flag.Int("batch-size", 1000, "每批 INSERT 数量（性能模式）")
	flagDBDSN     = flag.String("db-dsn", "", "数据库连接串；空则读取 YDSZ_DATABASE_URL 或默认值")
)

/* ------------------------------------------------------------------ */
/* 确定性种子数据                                                       */
/* ------------------------------------------------------------------ */

type seedUser struct {
	Email       string
	Password    string
	DisplayName string
	Timezone    string
}

type seedWorkspace struct {
	Name    string
	Slug    string
	Members map[string]string // email → role
}

var deterministicUsers = []seedUser{
	{"admin@ydsz.dev", "Admin@123", "系统管理员", "Asia/Shanghai"},
	{"pm@ydsz.dev", "Pm@123456", "李产品", "Asia/Shanghai"},
	{"dev@ydsz.dev", "Dev@123456", "王工程", "Asia/Shanghai"},
	{"designer@ydsz.dev", "Design@123", "张设计", "Asia/Shanghai"},
	{"viewer@ydsz.dev", "Viewer@123", "访客小赵", "Asia/Shanghai"},
}

var deterministicWorkspaces = []seedWorkspace{
	{
		Name: "核心产品", Slug: "core",
		Members: map[string]string{
			"admin@ydsz.dev": "owner", "pm@ydsz.dev": "admin",
			"dev@ydsz.dev": "member", "designer@ydsz.dev": "member", "viewer@ydsz.dev": "guest",
		},
	},
	{
		Name: "设计系统", Slug: "design-system",
		Members: map[string]string{
			"admin@ydsz.dev": "admin", "designer@ydsz.dev": "owner", "pm@ydsz.dev": "member",
		},
	},
	{
		Name: "基础设施", Slug: "infra",
		Members: map[string]string{
			"admin@ydsz.dev": "owner", "dev@ydsz.dev": "admin",
		},
	},
}

/* ------------------------------------------------------------------ */
/* main                                                                */
/* ------------------------------------------------------------------ */

func main() {
	flag.Parse()

	var err error
	if *flagCount > 0 {
		err = runBulk(*flagCount, *flagBatchSize, *flagDBDSN)
	} else {
		err = runDeterministic()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
}

/* ================================================================== */
/* 模式 1：确定性种子                                                    */
/* ================================================================== */

func runDeterministic() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()

	pool, err := persistence.NewPool(ctx, cfg.Database.URL, cfg.Database.MaxConns)
	if err != nil {
		return err
	}
	defer pool.Close()

	authSvc := auth.NewService(pool.Pool, cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer,
		cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL, cfg.Auth.BcryptCost, true)

	emailsToIDs, err := seedUsers(ctx, pool, authSvc)
	if err != nil {
		return err
	}
	wsSlugsToIDs, err := seedWorkspaces(ctx, pool, emailsToIDs)
	if err != nil {
		return err
	}
	if err := seedMembers(ctx, pool, wsSlugsToIDs, deterministicWorkspaces); err != nil {
		return err
	}
	if err := seedAuditLogs(ctx, pool, emailsToIDs); err != nil {
		return err
	}
	printSummary(emailsToIDs, wsSlugsToIDs)
	return nil
}

// seedUsers 幂等写入确定性种子用户。
func seedUsers(ctx context.Context, pool *persistence.Pool, svc *auth.Service) (map[string]int64, error) {
	out := make(map[string]int64, len(deterministicUsers))
	for _, u := range deterministicUsers {
		hash, err := svc.HashPassword(u.Password)
		if err != nil {
			return nil, fmt.Errorf("hash %s: %w", u.Email, err)
		}
		var id int64
		err = pool.QueryRow(ctx, `
			INSERT INTO users (email, password_hash, display_name, timezone)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash
			RETURNING id`, u.Email, hash, u.DisplayName, u.Timezone).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("seed user %s: %w", u.Email, err)
		}
		out[u.Email] = id
	}
	return out, nil
}

// seedWorkspaces 幂等写入工作空间。
func seedWorkspaces(ctx context.Context, pool *persistence.Pool, userIDs map[string]int64) (map[string]int64, error) {
	ownerID := userIDs["admin@ydsz.dev"]
	out := make(map[string]int64, len(deterministicWorkspaces))
	for _, ws := range deterministicWorkspaces {
		var id int64
		err := pool.QueryRow(ctx, `
			INSERT INTO workspaces (name, slug, owner_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (slug) WHERE status <> 'archived'
			DO UPDATE SET name = EXCLUDED.name
			RETURNING id`, ws.Name, ws.Slug, ownerID).Scan(&id)
		if err != nil {
			if err.Error() == "no rows in result set" {
				err = pool.QueryRow(ctx,
					`SELECT id FROM workspaces WHERE slug = $1 AND status = 'active'`, ws.Slug).Scan(&id)
				if err != nil {
					return nil, fmt.Errorf("seed ws lookup %s: %w", ws.Slug, err)
				}
			} else {
				return nil, fmt.Errorf("seed workspace %s: %w", ws.Slug, err)
			}
		}
		out[ws.Slug] = id
	}
	return out, nil
}

// seedMembers 幂等写入工作空间成员关系。
func seedMembers(ctx context.Context, pool *persistence.Pool, wsIDs map[string]int64, wsList []seedWorkspace) error {
	for _, ws := range wsList {
		wsID, ok := wsIDs[ws.Slug]
		if !ok {
			continue
		}
		for email, role := range ws.Members {
			var userID int64
			if err := pool.QueryRow(ctx,
				`SELECT id FROM users WHERE email = $1 AND is_active`, email).Scan(&userID); err != nil {
				return fmt.Errorf("lookup user %s: %w", email, err)
			}
			if _, err := pool.Exec(ctx, `
				INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
				wsID, userID, role, time.Now().Add(-30*24*time.Hour)); err != nil {
				return fmt.Errorf("seed member %s@%s: %w", email, ws.Slug, err)
			}
		}
	}
	return nil
}

// seedAuditLogs 写入演示审计日志。
func seedAuditLogs(ctx context.Context, pool *persistence.Pool, userIDs map[string]int64) error {
	admin := userIDs["admin@ydsz.dev"]
	for i := 0; i < 5; i++ {
		if _, err := pool.Exec(ctx, `
			INSERT INTO audit_logs (actor_id, action, target, detail)
			VALUES ($1, $2, $3, $4)`,
			admin,
			fmt.Sprintf("seed.demo_event_%d", i),
			fmt.Sprintf("demo-target-%d", i),
			fmt.Sprintf(`{"index":%d,"note":"seeded for demo"}`, i)); err != nil {
			return fmt.Errorf("seed audit %d: %w", i, err)
		}
	}
	return nil
}

// printSummary 打印确定性种子执行结果。
func printSummary(userIDs map[string]int64, wsIDs map[string]int64) {
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  Ydsz Plane Seed — OK")
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("  Users:      %d\n", len(deterministicUsers))
	fmt.Printf("  Workspaces: %d\n", len(deterministicWorkspaces))
	fmt.Println("───────────────────────────────────────")
	fmt.Println(" 用户 / 密码:")
	for _, u := range deterministicUsers {
		fmt.Printf("   %-22s %s\n", u.Email, u.Password)
	}
	fmt.Println("───────────────────────────────────────")
	fmt.Println(" 工作空间:")
	for _, ws := range deterministicWorkspaces {
		fmt.Printf("   %-16s (%d 成员)\n", ws.Slug, len(ws.Members))
	}
	_ = userIDs
	_ = wsIDs
	fmt.Println("═══════════════════════════════════════")
}
