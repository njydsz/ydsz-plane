// Package storage — 存储客户端纯逻辑测试。
//
// 不连接 MinIO/S3，仅测试：
//   - newFSClient 分支（绕过网络）
//   - fsPath 路径拼合
//   - Bucket() / IsFSFallback() 访问器
//   - Bucket 名默认值验证
package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/njydsz/ydsz-plane/internal/config"
)

// TestNewFSClient_Fields 验证 newFSClient 正确设置回退字段。
// 不依赖外部 MinIO 实例（newFSClient 仅创建本地目录）。
func TestNewFSClient_Fields(t *testing.T) {
	tests := []struct {
		name   string
		bucket string
	}{
		{"ydsz-plane", "ydsz-plane"},
		{"test-bucket", "test-bucket"},
		{"edge-case.bucket", "edge-case.bucket"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.StorageConfig{
				Endpoint:  "127.0.0.1:9000",
				AccessKey: "test",
				SecretKey: "test",
				Bucket:    tt.bucket,
				UseSSL:    false,
				Region:    "us-east-1",
			}

			c, err := newFSClient(cfg)
			if err != nil {
				t.Fatalf("newFSClient: %v", err)
			}

			// 回退模式应激活
			if !c.IsFSFallback() {
				t.Fatal("expected IsFSFallback() = true")
			}

			// Bucket 名应保持一致
			if c.Bucket() != tt.bucket {
				t.Fatalf("Bucket() = %q, want %q", c.Bucket(), tt.bucket)
			}

			// mc 应为 nil（回退模式）
			if c.mc != nil {
				t.Fatal("expected mc to be nil in fs fallback mode")
			}

			// fsDir 应包含 bucket 名
			if !strings.Contains(c.fsDir, tt.bucket) {
				t.Fatalf("fsDir = %q, expected to contain %q", c.fsDir, tt.bucket)
			}

			// 清理测试产生的临时目录
			t.Cleanup(func() {
				_ = os.RemoveAll(c.fsDir)
			})
		})
	}
}

// TestFSPath 验证文件系统回退模式下的拼合路径逻辑。
func TestFSPath(t *testing.T) {
	cfg := config.StorageConfig{
		Endpoint:  "127.0.0.1:9000",
		AccessKey: "test",
		SecretKey: "test",
		Bucket:    "test-bucket",
	}
	c, err := newFSClient(cfg)
	if err != nil {
		t.Fatalf("newFSClient: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(c.fsDir) })

	tests := []struct {
		name       string
		storageKey string
		wantSuffix string
	}{
		{"simple key", "file.txt", "file.txt"},
		{"nested key", filepath.Join("2026", "08", "report.pdf"), filepath.Join("2026", "08", "report.pdf")},
		{"uuid key", filepath.Join("abc-123-def", "image.png"), filepath.Join("abc-123-def", "image.png")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.fsPath(tt.storageKey)
			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Fatalf("fsPath(%q) = %q, want suffix %q",
					tt.storageKey, got, tt.wantSuffix)
			}
			// 路径应在 fsDir 下
			if !strings.HasPrefix(got, c.fsDir) {
				t.Fatalf("fsPath(%q) = %q, want prefix %q",
					tt.storageKey, got, c.fsDir)
			}
		})
	}
}

// TestNewFSClient_DirCreated 验证回退目录被物理创建。
func TestNewFSClient_DirCreated(t *testing.T) {
	cfg := config.StorageConfig{
		Endpoint:  "127.0.0.1:9000",
		AccessKey: "test",
		SecretKey: "test",
		Bucket:    "dir-check",
	}
	c, err := newFSClient(cfg)
	if err != nil {
		t.Fatalf("newFSClient: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(c.fsDir) })

	info, err := os.Stat(c.fsDir)
	if err != nil {
		t.Fatalf("stat fsDir: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("fsDir is not a directory")
	}
}

// TestStorageConfig_Defaults 验证 StorageConfig 零值时的默认行为。
func TestStorageConfig_Defaults(t *testing.T) {
	cfg := config.StorageConfig{}

	// 未设置 UseSSL → false（默认）
	if cfg.UseSSL {
		t.Fatal("expected UseSSL = false (zero value)")
	}

	// 未设置 Region → ""
	if cfg.Region != "" {
		t.Fatalf("expected Region = \"\", got %q", cfg.Region)
	}

	// 未设置 PublicURL → ""
	if cfg.PublicURL != "" {
		t.Fatalf("expected PublicURL = \"\", got %q", cfg.PublicURL)
	}
}
