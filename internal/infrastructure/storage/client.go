// Package storage 提供对象存储（MinIO / S3 兼容）抽象层。
package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/njydsz/ydsz-plane/internal/config"
)

// Client 封装 MinIO 客户端，提供预签名上传/下载能力。
type Client struct {
	mc     *minio.Client
	bucket string
	cfg    config.StorageConfig
	// fsFallback 为 true 时使用本地文件系统模式（无需 MinIO）
	fsFallback bool
	fsDir      string
}

// New 根据配置创建 MinIO 客户端并确保 Bucket 存在。
// 开发模式下若 MinIO 不可达，自动降级为本地文件系统存储。
func New(cfg config.StorageConfig) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: connect to %s: %w", cfg.Endpoint, err)
	}

	exists, err := mc.BucketExists(context.Background(), cfg.Bucket)
	if err != nil {
		// MinIO 不可达，启用文件系统回退
		return newFSClient(cfg)
	}
	if !exists {
		if err := mc.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{
			Region: cfg.Region,
		}); err != nil {
			// 创建桶失败，启用文件系统回退
			return newFSClient(cfg)
		}
	}

	return &Client{mc: mc, bucket: cfg.Bucket, cfg: cfg}, nil
}

// newFSClient 创建本地文件系统存储客户端（开发模式回退）。
func newFSClient(cfg config.StorageConfig) (*Client, error) {
	fsDir := filepath.Join(os.TempDir(), "ydsz-plane-storage", cfg.Bucket)
	if err := os.MkdirAll(fsDir, 0o755); err != nil {
		return nil, fmt.Errorf("storage: create fallback dir %s: %w", fsDir, err)
	}
	return &Client{
		bucket:     cfg.Bucket,
		cfg:        cfg,
		fsFallback: true,
		fsDir:      fsDir,
	}, nil
}

// Upload 上传文件到对象存储。
func (c *Client) Upload(ctx context.Context, storageKey string, reader io.Reader, size int64, contentType string) error {
	if c.fsFallback {
		return c.fsUpload(storageKey, reader)
	}
	_, err := c.mc.PutObject(ctx, c.bucket, storageKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("storage: upload %s: %w", storageKey, err)
	}
	return nil
}

// PresignedUploadURL 生成预签名上传 URL。
func (c *Client) PresignedUploadURL(ctx context.Context, storageKey string, expiry time.Duration, contentType string) (string, error) {
	if c.fsFallback {
		return fmt.Sprintf("http://127.0.0.1:8080/_local_upload/%s", storageKey), nil
	}
	u, err := c.mc.PresignedPutObject(ctx, c.bucket, storageKey, expiry)
	if err != nil {
		return "", fmt.Errorf("storage: presigned upload %s: %w", storageKey, err)
	}
	return u.String(), nil
}

// PresignedDownloadURL 生成预签名下载 URL。
func (c *Client) PresignedDownloadURL(ctx context.Context, storageKey string, expiry time.Duration) (string, error) {
	if c.fsFallback {
		return fmt.Sprintf("http://127.0.0.1:8080/_local_download/%s", storageKey), nil
	}
	reqParams := make(url.Values)
	u, err := c.mc.PresignedGetObject(ctx, c.bucket, storageKey, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("storage: presigned download %s: %w", storageKey, err)
	}
	return u.String(), nil
}

// Delete 从对象存储中删除文件。
func (c *Client) Delete(ctx context.Context, storageKey string) error {
	if c.fsFallback {
		return c.fsDelete(storageKey)
	}
	err := c.mc.RemoveObject(ctx, c.bucket, storageKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("storage: delete %s: %w", storageKey, err)
	}
	return nil
}

// Exists 检查对象是否存在。
func (c *Client) Exists(ctx context.Context, storageKey string) (bool, error) {
	if c.fsFallback {
		return c.fsExists(storageKey)
	}
	_, err := c.mc.StatObject(ctx, c.bucket, storageKey, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("storage: stat %s: %w", storageKey, err)
	}
	return true, nil
}

// Size 返回已上传对象的字节大小。
func (c *Client) Size(ctx context.Context, storageKey string) (int64, error) {
	if c.fsFallback {
		return c.fsSize(storageKey)
	}
	info, err := c.mc.StatObject(ctx, c.bucket, storageKey, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return 0, nil
		}
		return 0, fmt.Errorf("storage: stat %s: %w", storageKey, err)
	}
	return info.Size, nil
}

// Bucket 返回当前使用的存储桶名称。
func (c *Client) Bucket() string { return c.bucket }

// IsFSFallback 返回是否正在使用本地文件系统回退存储。
func (c *Client) IsFSFallback() bool { return c.fsFallback }

// --- 文件系统回退实现 ---

func (c *Client) fsPath(storageKey string) string {
	return filepath.Join(c.fsDir, storageKey)
}

func (c *Client) fsUpload(storageKey string, reader io.Reader) error {
	path := c.fsPath(storageKey)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("storage: create dir %s: %w", dir, err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("storage: read upload data: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("storage: write %s: %w", storageKey, err)
	}
	return nil
}

func (c *Client) fsDelete(storageKey string) error {
	path := c.fsPath(storageKey)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: delete %s: %w", storageKey, err)
	}
	return nil
}

func (c *Client) fsExists(storageKey string) (bool, error) {
	path := c.fsPath(storageKey)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("storage: stat %s: %w", storageKey, err)
	}
	return true, nil
}

func (c *Client) fsSize(storageKey string) (int64, error) {
	path := c.fsPath(storageKey)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return -1, fmt.Errorf("storage: stat %s: %w", storageKey, err)
	}
	return info.Size(), nil
}
