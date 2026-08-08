// Package automation — automation_templates 种子数据写入。
//
// 让 go run ./scripts/seed 也能写入模板，避免仅靠 SQL 迁移。
// 幂等：ON CONFLICT (slug) DO UPDATE。
package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedAutomationTemplates 写入全部 BuiltInTemplates() 到 automation_templates 表。
// 幂等：重复执行不报错，已存在的条目按 slug 更新。
//
// 需要 automation_templates 表已存在，且 slug 列有唯一约束。
func SeedAutomationTemplates(ctx context.Context, db *pgxpool.Pool) (int, error) {
	templates := BuiltInTemplates()
	inserted := 0

	for _, t := range templates {
		dslJSON, err := json.Marshal(t.DSLTemplate)
		if err != nil {
			return inserted, fmt.Errorf("template %s marshal: %w", t.Slug, err)
		}

		tag, err := db.Exec(ctx, `
			INSERT INTO automation_templates
				(slug, name, description, category, dsl_template, icon, sort_order, is_recommended, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
			ON CONFLICT (slug) DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				category = EXCLUDED.category,
				dsl_template = EXCLUDED.dsl_template,
				icon = EXCLUDED.icon,
				sort_order = EXCLUDED.sort_order,
				is_recommended = EXCLUDED.is_recommended`,
			t.Slug, t.Name, t.Description, t.Category,
			dslJSON, t.Icon, t.SortOrder, t.IsRecommended,
		)
		if err != nil {
			// 表不存在时优雅跳过（避免 seed 脚本强依赖某些表）
			if isUndefinedTable(err) {
				return 0, nil
			}
			return inserted, fmt.Errorf("seed template %s: %w", t.Slug, err)
		}
		if tag.RowsAffected() > 0 {
			inserted++
		}
	}
	return inserted, nil
}

// isUndefinedTable 检查错误是否为 "undefined_table" (42P01)。
func isUndefinedTable(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "42P01")
}
