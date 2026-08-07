// Command migrate 使用 golang-migrate 执行数据库迁移。
// 用法: go run ./cmd/migrate [up|down N|version]
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/njydsz/ydsz-plane/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}

// run 解析命令行参数并执行迁移命令。
//
// 支持的子命令：
//   - up：应用所有待执行的迁移（默认，无参数时生效）。
//   - down N：回退 N 个迁移，默认回退 1 个。
//   - version：仅打印当前数据库迁移版本与脏状态，不执行任何变更。
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	m, err := migrate.New("file://sql", cfg.Database.URL)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	case "down":
		n := 1
		if len(os.Args) > 2 {
			if n, err = strconv.Atoi(os.Args[2]); err != nil {
				return fmt.Errorf("invalid down steps: %w", err)
			}
		}
		if err := m.Steps(-n); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			return err
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
		return nil
	default:
		return fmt.Errorf("unknown command %q (up|down N|version)", cmd)
	}

	v, dirty, _ := m.Version()
	fmt.Printf("migrated: version=%d dirty=%v\n", v, dirty)
	return nil
}
