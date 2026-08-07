// Package main 为 seed 工具集提供数据库连接与通用辅助函数。
//
// 提供：
//   - loadDSN：从环境变量或配置文件读取数据库连接串
//   - maskDSN：连接串安全脱敏打印（密码隐藏）
package main

import (
	"fmt"
	"os"

	"github.com/njydsz/ydsz-plane/internal/config"
)

// loadDSN 从环境变量或默认配置获取数据库连接串。
func loadDSN() (string, error) {
	// 先尝试直接环境变量
	if dsn := os.Getenv("YDSZ_DATABASE_URL"); dsn != "" {
		return dsn, nil
	}
	// 回退读取默认配置
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("load config: %w", err)
	}
	return cfg.Database.URL, nil
}

// maskDSN 安全化打印连接串（密码脱敏）。
func maskDSN(dsn string) string {
	// 简单处理：如果有 @ 就隐藏前面部分
	// postgres://user:pass@host → postgres://***@host
	at := -1
	for i := 0; i < len(dsn); i++ {
		if dsn[i] == '@' {
			at = i
			break
		}
	}
	if at > 0 {
		return "***" + dsn[at:]
	}
	return dsn
}
