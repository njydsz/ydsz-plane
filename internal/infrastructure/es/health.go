package es

import (
	"context"
	"net/http"
)

// HealthCheckHTTP 执行 ES 集群健康检查，返回原始 HTTP response。
// 供 health 包的 ElasticsearchChecker 使用，以适配统一的连通性探测接口。
//
// 调用方负责 close response body。
func (c *Client) HealthCheckHTTP(ctx context.Context) (*http.Response, error) {
	return c.doRequest(ctx, "GET", "/_cluster/health?timeout=3s", nil)
}
