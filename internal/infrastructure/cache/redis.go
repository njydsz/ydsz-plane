// Package cache 提供 Redis 客户端，用于缓存、限流、分布式锁与 WebSocket 扇出。
package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// NewClient 创建 Redis 客户端并验证连通性。
//
// 参数：
//   - addr：Redis 服务地址（host:port）。
//   - password：AUTH 密码；空字符串表示不启用认证。
//   - db：逻辑数据库编号。
//
// 返回连接失败时错误；Ping 失败会关闭客户端避免句柄泄漏。
func NewClient(ctx context.Context, addr, password string, db int) (*redis.Client, error) {
	cli := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	if err := cli.Ping(ctx).Err(); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("cache: ping redis: %w", err)
	}
	return cli, nil
}
