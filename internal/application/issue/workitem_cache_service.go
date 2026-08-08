package issue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// WorkitemCacheService 工作项缓存服务
// 实现详情缓存、列表缓存失效逻辑，提升查询性能
type WorkitemCacheService struct {
	redisClient *redis.Client
	detailTTL   time.Duration // 详情缓存默认TTL
	listTTL     time.Duration // 列表缓存默认TTL
}

// NewWorkitemCacheService 创建缓存服务
func NewWorkitemCacheService(redisClient *redis.Client) *WorkitemCacheService {
	return &WorkitemCacheService{
		redisClient: redisClient,
		detailTTL:   10 * time.Minute, // 详情缓存10分钟，对齐项目现有设计
		listTTL:     5 * time.Minute,  // 列表缓存5分钟
	}
}

// GetDetail 获取工作项详情缓存
func (s *WorkitemCacheService) GetDetail(ctx context.Context, entityType IssueTypeCode, entityID int64, dest any) (bool, error) {
	key := s.getDetailCacheKey(entityType, entityID)
	data, err := s.redisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(data, dest)
	if err != nil {
		return false, err
	}
	return true, nil
}

// SetDetail 设置工作项详情缓存
func (s *WorkitemCacheService) SetDetail(ctx context.Context, entityType IssueTypeCode, entityID int64, src any) error {
	key := s.getDetailCacheKey(entityType, entityID)
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return s.redisClient.Set(ctx, key, data, s.detailTTL).Err()
}

// InvalidateDetail 失效工作项详情缓存（更新、删除时调用）
func (s *WorkitemCacheService) InvalidateDetail(ctx context.Context, entityType IssueTypeCode, entityID int64) error {
	key := s.getDetailCacheKey(entityType, entityID)
	return s.redisClient.Del(ctx, key).Err()
}

// InvalidateListCache 失效工作项列表缓存（用于筛选条件变化时）
func (s *WorkitemCacheService) InvalidateListCache(ctx context.Context, wsID, projectID int64) error {
	// 用通配符删除该项目下的所有列表缓存
	pattern := fmt.Sprintf("workitem:list:%d:%d:*", wsID, projectID)
	iter := s.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		_ = s.redisClient.Del(ctx, iter.Val())
	}
	return iter.Err()
}

// getDetailCacheKey 生成详情缓存的key
func (s *WorkitemCacheService) getDetailCacheKey(entityType IssueTypeCode, entityID int64) string {
	return fmt.Sprintf("workitem:detail:%s:%d", entityType, entityID)
}
