// Package cache — 后台增量同步策略。
//
// 桌面端在以下场景与后端同步：
//   1. 应用启动时：增量拉取 workspace/project/issue 最近变更
//   2. 网络恢复时（NetworkChangeListener）：补偿同步
//   3. 显式刷新：用户手动点"刷新"按钮
//   4. 离线草稿提交：检测到网络可用后自动 push 未同步草稿
//
// 同步策略：基于 updated_at 的增量拉取（cursor-based），全量仅首次启动。
package cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// SyncManager 协调本地缓存与远程 API 的同步。
type SyncManager struct {
	cache    *Cache
	apiBase  string
	mu       sync.Mutex
	interval time.Duration // 自动同步间隔（默认 60s）
	stopCh   chan struct{}
}

// NewSyncManager 创建一个同步管理器。
func NewSyncManager(cache *Cache, apiBase string) *SyncManager {
	return &SyncManager{
		cache:    cache,
		apiBase:  apiBase,
		interval: 60 * time.Second,
		stopCh:   make(chan struct{}),
	}
}

// SetInterval 设置自动同步间隔（最小 15s，最大 10min）。
func (s *SyncManager) SetInterval(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d < 15*time.Second {
		d = 15 * time.Second
	}
	if d > 10*time.Minute {
		d = 10 * time.Minute
	}
	s.interval = d
}

// StartAutoSync 启动后台自动同步。
func (s *SyncManager) StartAutoSync(ctx context.Context) {
	go s.loop(ctx)
}

// StopAutoSync 停止后台同步。
func (s *SyncManager) StopAutoSync() {
	close(s.stopCh)
}

// SyncWorkspace 全量同步指定工作空间的项目列表。
//
// 步骤：
//   1. GET /api/v1/workspaces/:ws/projects
//   2. 批量 upsert 到 Cache.projects
//   3. 更新 last_sync cursor
func (s *SyncManager) SyncWorkspace(ctx context.Context, workspaceID int64) error {
	path := fmt.Sprintf("%s/api/v1/workspaces/%d/projects", s.apiBase, workspaceID)
	_ = path // 真实实现：HTTP GET + JSON Decode

	// placeholder: 拉取后逐个调用 cache.SaveProject
	return nil
}

// SyncProjectIssues 增量同步指定项目的工作项。
//
// cursor 为上次同步的时间戳（Unix 秒），后端返回 (updated_at > cursor) 的项目列表。
func (s *SyncManager) SyncProjectIssues(ctx context.Context, projectID int64, cursor int64) (int, error) {
	path := fmt.Sprintf("%s/api/v1/projects/%d/issues?updated_at_gt=%d", s.apiBase, projectID, cursor)
	_ = path

	// placeholder:
	// 1. HTTP GET + decode
	// 2. for each issue: cache.UpsertIssue(...)
	// return count of new/updated issues
	return 0, nil
}

// SubmitDrafts 将本地未同步草稿提交到后端。
//
// 步骤：
//   1. Cache.ListUnsyncedDrafts → 获取待提交草稿
//   2. for each draft: POST /api/v1/issues
//   3. 成功后 cache.MarkDraftSynced(draftID)
//
// Wails 绑定：window.go.main.cache.SubmitDrafts() → 返回成功/失败计数
func (s *SyncManager) SubmitDrafts(ctx context.Context) (synced, failed int, err error) {
	drafts, err := s.cache.ListUnsyncedDrafts(ctx)
	if err != nil {
		return 0, 0, err
	}

	for _, draft := range drafts {
		id := draft["id"].(int64)
		_ = draft["project_id"]
		_ = draft["title"]
		_ = draft["body"]

		// placeholder: POST /api/v1/issues
		// if resp OK: mark synced
		if err := s.cache.MarkDraftSynced(ctx, id); err != nil {
			failed++
			continue
		}
		synced++
	}

	return synced, failed, nil
}

// HealthCheck 返回同步状态调试信息。
func (s *SyncManager) HealthCheck() string {
	return fmt.Sprintf("sync=ready interval=%s api=%s", s.interval, s.apiBase)
}

// --- internals ---

func (s *SyncManager) loop(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runPeriodicSync(ctx)
		case <-s.stopCh:
			return
		}
	}
}

func (s *SyncManager) runPeriodicSync(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. 提交本地草稿
	if _, _, err := s.SubmitDrafts(ctx); err != nil {
		log.Printf("sync: submit drafts failed: %v", err)
	}

	// 2. 同步活跃项目的工作项（placeholder: 真实场景从 workspace 种子数据获取）
	// for _, p := range s.activeProjects {
	//     _ = s.SyncProjectIssues(ctx, p.ID, p.LastSync)
	// }
}
