// Command seed 向开发/演示环境插入确定性种子数据。
//
// 种子数据（幂等，可重复执行）：
//   - 管理员 / PM / 工程 / 设计 / 访客 5 个测试账号（详见 users 变量）。
//   - 3 个工作空间（核心产品 / 设计系统 / 基础设施），含成员关系。
//   - 审计日志演示条目。
//
// 参考: Linear / Plane / Asana 的 seed 策略——确定性数据便于复现 bug 与演示。
//
// 典型的 run() 顺序：
//   1. 加载配置
//   2. 建 DB 连接池
//   3. 构造 auth.Service（用于 HashPassword）
//   4. INSERT ... ON CONFLICT 幂等写入 users / workspaces / workspace_members / audit_logs
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

/* ------------------------------------------------------------------ */
/* seed data (deterministic)                                            */
/* ------------------------------------------------------------------ */

// seedUser 一个种子用户的基本信息。
type seedUser struct {
	Email       string // 邮箱，唯一标识。
	Password    string // 明文密码（seed 中仅用于生成 bcrypt hash，不落库）。
	DisplayName string // 显示名。
	Timezone    string // IANA 时区。
}

// seedWorkspace 一个种子工作空间的定义。
type seedWorkspace struct {
	Name    string            // 显示名。
	Slug    string            // URL 友好唯一标识。
	Members map[string]string // 成员映射：email → role（owner / admin / member / guest）。
}

var users = []seedUser{
	{"admin@ydsz.dev",    "Admin@123",   "系统管理员", "Asia/Shanghai"},
	{"pm@ydsz.dev",       "Pm@123456",   "李产品",     "Asia/Shanghai"},
	{"dev@ydsz.dev",      "Dev@123456",   "王工程",     "Asia/Shanghai"},
	{"designer@ydsz.dev", "Design@123",  "张设计",     "Asia/Shanghai"},
	{"viewer@ydsz.dev",   "Viewer@123",  "访客小赵",   "Asia/Shanghai"},
}

var workspaces = []seedWorkspace{
	{
		Name: "核心产品",
		Slug: "core",
		Members: map[string]string{
			"admin@ydsz.dev":    "owner",
			"pm@ydsz.dev":       "admin",
			"dev@ydsz.dev":      "member",
			"designer@ydsz.dev": "member",
			"viewer@ydsz.dev":   "guest",
		},
	},
	{
		Name: "设计系统",
		Slug: "design-system",
		Members: map[string]string{
			"admin@ydsz.dev":    "admin",
			"designer@ydsz.dev": "owner",
			"pm@ydsz.dev":       "member",
		},
	},
	{
		Name: "基础设施",
		Slug: "infra",
		Members: map[string]string{
			"admin@ydsz.dev": "owner",
			"dev@ydsz.dev":    "admin",
		},
	},
}

/* ------------------------------------------------------------------ */
/* main                                                                 */
/* ------------------------------------------------------------------ */

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
}

// run 执行种子数据写入流程。
//
// 顺序：
//  1. 加载配置并建立 DB 连接池；
//  2. 构造 auth.Service（用于 HashPassword 生成 bcrypt 散列）；
//  3. 幂等写入 users / workspaces / workspace_members / audit_logs；
//  4. 打印账号与工作空间摘要。
func run() error {
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

	if err := seedMembers(ctx, pool, wsSlugsToIDs, workspaces); err != nil {
		return err
	}

	if err := seedAuditLogs(ctx, pool, emailsToIDs); err != nil {
		return err
	}

	printSummary(emailsToIDs, wsSlugsToIDs)
	return nil
}

/* ------------------------------------------------------------------ */
/* helpers                                                              */
/* ------------------------------------------------------------------ */

// seedUsers 幂等写入种子用户，返回 email → 用户 ID 的映射。
// 已存在的用户按 email 冲突更新密码散列，保证可重复执行。
func seedUsers(ctx context.Context, pool *persistence.Pool, svc *auth.Service) (map[string]int64, error) {
	out := make(map[string]int64, len(users))
	for _, u := range users {
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

// seedWorkspaces 幂等写入种子工作空间，返回 slug → 工作空间 ID 的映射。
// 已归档的同 slug 记录会被跳过（WHERE status <> 'archived'），
// 激活状态的记录按 name 更新。
func seedWorkspaces(ctx context.Context, pool *persistence.Pool, userIDs map[string]int64) (map[string]int64, error) {
	ownerID := userIDs["admin@ydsz.dev"]
	out := make(map[string]int64, len(workspaces))
	for _, ws := range workspaces {
		var id int64
		// 兜底保护：若 Members 映射中恰好没有 admin，
		// 首个写入的成员也会成为 owner。
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
// 已存在的 (workspace_id, user_id) 组合按角色更新，保证可重复执行。
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

// seedAuditLogs 写入 5 条演示用审计日志，便于开发环境查看审计时间线。
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

// printSummary 打印种子数据执行结果（账号列表与工作空间成员数）。
func printSummary(userIDs map[string]int64, wsIDs map[string]int64) {
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  Ydsz Plane Seed — OK")
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("  Users:      %d\n", len(users))
	fmt.Printf("  Workspaces: %d\n", len(workspaces))
	fmt.Println("───────────────────────────────────────")
	fmt.Println(" 用户 / 密码:")
	for _, u := range users {
		fmt.Printf("   %-22s %s\n", u.Email, u.Password)
	}
	fmt.Println("───────────────────────────────────────")
	fmt.Println(" 工作空间:")
	for _, ws := range workspaces {
		fmt.Printf("   %-16s (%d 成员)\n", ws.Slug, len(ws.Members))
	}
	_ = userIDs
	_ = wsIDs
	fmt.Println("═══════════════════════════════════════")
}
