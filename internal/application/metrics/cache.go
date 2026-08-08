// Package metrics — 效能度量的 Redis 读穿透缓存层。
//
// 对标参考:
//   - 美团 Raptor 多级缓存架构（本地 Caffeine + Redis 分布式）
//   - 字节跳动 Metric Cache（TTL 按指标粒度差异化）
//   - 阿里 ARMS 时效性分层（实时 vs 准实时）
//
// 设计原则:
//   - 实时类（WIP、今日吞吐量）→ 30s TTL，走防击穿 singleflight
//   - 日级聚合（Velocity、Lead Time P85）→ 5min TTL，后台每日快照兜底
//   - DORA 类 → 1h TTL，30 天窗口聚合开销大
//   - 缓存击穿防护：singleflight + 空值缓存（1min）
package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheTTL 按指标类型差异化的 TTL 配置。
type CacheTTL struct {
	Realtime    time.Duration // 实时类：30s
	DailyAgg    time.Duration // 日聚合类：5min
	DORA        time.Duration // DORA 类：1hour
	Snapshot    time.Duration // 快照历史：1hour
	EmptyMarker time.Duration // 空值标记：1min（防穿透）
}

// DefaultCacheTTL 默认 TTL 配置。
var DefaultCacheTTL = CacheTTL{
	Realtime:    30 * time.Second,
	DailyAgg:    5 * time.Minute,
	DORA:        1 * time.Hour,
	Snapshot:    1 * time.Hour,
	EmptyMarker: 1 * time.Minute,
}

// MetricCache 封装效能度量的 Redis 缓存操作。
type MetricCache struct {
	cli *redis.Client
	ttl CacheTTL
}

// NewMetricCache 创建效能度量缓存层。
func NewMetricCache(cli *redis.Client) *MetricCache {
	return &MetricCache{
		cli: cli,
		ttl: DefaultCacheTTL,
	}
}

// WithTTL 自定义 TTL 配置（可选链式调用）。
func (c *MetricCache) WithTTL(ttl CacheTTL) *MetricCache {
	c.ttl = ttl
	return c
}

// --- Key Builders ---

func (c *MetricCache) keyVelocity(wsID, projectID int64, lastN int) string {
	return fmt.Sprintf("mtr:vel:%d:%d:%d", wsID, projectID, lastN)
}

func (c *MetricCache) keyLeadTime(wsID, projectID int64, days int) string {
	return fmt.Sprintf("mtr:lt:%d:%d:%d", wsID, projectID, days)
}

func (c *MetricCache) keyQuality(wsID, projectID int64) string {
	return fmt.Sprintf("mtr:qly:%d:%d", wsID, projectID)
}

func (c *MetricCache) keyDORA(wsID, projectID int64) string {
	return fmt.Sprintf("mtr:dora:%d:%d", wsID, projectID)
}

func (c *MetricCache) keyResourceLoad(wsID, projectID int64) string {
	return fmt.Sprintf("mtr:rl:%d:%d", wsID, projectID)
}

func (c *MetricCache) keySnapshot(wsID, projectID int64, metric string) string {
	return fmt.Sprintf("mtr:snap:%d:%d:%s", wsID, projectID, metric)
}

// --- Generic Helpers ---

const emptyMarker = "__EMPTY__"

// getJSON 从缓存反序列化。未命中返回（nil, nil）。
func (c *MetricCache) getJSON(ctx context.Context, key string, dest any) (bool, error) {
	raw, err := c.cli.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // 未命中
	}
	if err != nil {
		return false, err // Redis 错误，让调用方决定是否降级
	}
	if raw == emptyMarker {
		return true, nil // 命中空标记（上次查无数据）
	}
	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		return false, fmt.Errorf("unmarshal cache: %w", err)
	}
	return true, nil
}

// setJSON 序列化后写入缓存。
func (c *MetricCache) setJSON(ctx context.Context, key string, val any, ttl time.Duration) error {
	if val == nil {
		// 空值标记（防穿透）
		return c.cli.Set(ctx, key, emptyMarker, c.ttl.EmptyMarker).Err()
	}
	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}
	return c.cli.Set(ctx, key, data, ttl).Err()
}

// InvalidateProject 失效某个项目的全部指标缓存（工作项/迭代变更后调用）。
func (c *MetricCache) InvalidateProject(ctx context.Context, wsID, projectID int64) error {
	pattern := fmt.Sprintf("mtr:*:%d:%d:*", wsID, projectID)
	// SCAN 避免 KEYS 阻塞
	iter := c.cli.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.cli.Del(ctx, keys...).Err()
	}
	return nil
}

// --- Typed Cache Ops ---

// GetVelocity 从缓存读取 VelocityResult。
func (c *MetricCache) GetVelocity(ctx context.Context, wsID, projectID int64, lastN int) (*VelocityResult, bool) {
	var r VelocityResult
	hit, err := c.getJSON(ctx, c.keyVelocity(wsID, projectID, lastN), &r)
	return &r, hit && err == nil
}

// SetVelocity 写入 VelocityResult 缓存。
func (c *MetricCache) SetVelocity(ctx context.Context, wsID, projectID int64, lastN int, r *VelocityResult) {
	_ = c.setJSON(ctx, c.keyVelocity(wsID, projectID, lastN), r, c.ttl.DailyAgg)
}

// GetLeadTime 从缓存读取 LeadTimeResult。
func (c *MetricCache) GetLeadTime(ctx context.Context, wsID, projectID int64, days int) (*LeadTimeResult, bool) {
	var r LeadTimeResult
	hit, err := c.getJSON(ctx, c.keyLeadTime(wsID, projectID, days), &r)
	return &r, hit && err == nil
}

// SetLeadTime 写入 LeadTimeResult 缓存。
func (c *MetricCache) SetLeadTime(ctx context.Context, wsID, projectID int64, days int, r *LeadTimeResult) {
	_ = c.setJSON(ctx, c.keyLeadTime(wsID, projectID, days), r, c.ttl.DailyAgg)
}

// GetQualityMetrics 从缓存读取 QualityMetrics。
func (c *MetricCache) GetQualityMetrics(ctx context.Context, wsID, projectID int64) (*QualityMetrics, bool) {
	var r QualityMetrics
	hit, err := c.getJSON(ctx, c.keyQuality(wsID, projectID), &r)
	return &r, hit && err == nil
}

// SetQualityMetrics 写入 QualityMetrics 缓存。
func (c *MetricCache) SetQualityMetrics(ctx context.Context, wsID, projectID int64, r *QualityMetrics) {
	_ = c.setJSON(ctx, c.keyQuality(wsID, projectID), r, c.ttl.DailyAgg)
}

// GetDORA 从缓存读取 DORAResult。
func (c *MetricCache) GetDORA(ctx context.Context, wsID, projectID int64) (*DORAResult, bool) {
	var r DORAResult
	hit, err := c.getJSON(ctx, c.keyDORA(wsID, projectID), &r)
	return &r, hit && err == nil
}

// SetDORA 写入 DORAResult 缓存。
func (c *MetricCache) SetDORA(ctx context.Context, wsID, projectID int64, r *DORAResult) {
	_ = c.setJSON(ctx, c.keyDORA(wsID, projectID), r, c.ttl.DORA)
}

// GetResourceLoad 从缓存读取 resourceLoadResponse。
func (c *MetricCache) GetResourceLoad(ctx context.Context, wsID, projectID int64) (*resourceLoadResponse, bool) {
	var r resourceLoadResponse
	hit, err := c.getJSON(ctx, c.keyResourceLoad(wsID, projectID), &r)
	return &r, hit && err == nil
}

// SetResourceLoad 写入 resourceLoadResponse 缓存。
func (c *MetricCache) SetResourceLoad(ctx context.Context, wsID, projectID int64, r *resourceLoadResponse) {
	_ = c.setJSON(ctx, c.keyResourceLoad(wsID, projectID), r, c.ttl.Realtime)
}

// GetSnapshots 从缓存读取 []map[string]any。
func (c *MetricCache) GetSnapshots(ctx context.Context, wsID, projectID int64, metric string) ([]map[string]any, bool) {
	var r []map[string]any
	hit, err := c.getJSON(ctx, c.keySnapshot(wsID, projectID, metric), &r)
	return r, hit && err == nil
}

// SetSnapshots 写入 []map[string]any 缓存。
func (c *MetricCache) SetSnapshots(ctx context.Context, wsID, projectID int64, metric string, r []map[string]any) {
	_ = c.setJSON(ctx, c.keySnapshot(wsID, projectID, metric), r, c.ttl.Snapshot)
}
