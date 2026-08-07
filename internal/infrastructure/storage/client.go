// Package storage 提供对象存储（MinIO / S3 兼容）抽象层。
package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/njydsz/ydsz-plane/internal/config"
)

// Client 封装 MinIO 客户端，提供预签名上传/下载能力。
type Client struct {
	mc      *minio.Client
	bucket  string
	cfg     config.StorageConfig
}

// New 根据配置创建 MinIO 客户端并确保 Bucket 存在。
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
		return nil, fmt.Errorf("storage: check bucket %s: %w", cfg.Bucket, err)
	}
	if !exists {
		if err := mc.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{
			Region: cfg.Region,
		}); err != nil {
			return nil, fmt.Errorf("storage: create bucket %s: %w", cfg.Bucket, err)
		}
	}

	return &Client{mc: mc, bucket: cfg.Bucket, cfg: cfg}, nil
}

// Upload 上传文件到对象存储。
// storageKey 是对象在桶中的唯一标识；reader 是文件内容；size 是文件大小；
// contentType 是 MIME 类型。
func (c *Client) Upload(ctx context.Context, storageKey string, reader io.Reader, size int64, contentType string) error {
	_, err := c.mc.PutObject(ctx, c.bucket, storageKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("storage: upload %s: %w", storageKey, err)
	}
	return nil
}

// PresignedUploadURL 生成预签名上传 URL。
// 客户端可用此 URL 直传文件（PUT 方式），无需经过 API 服务器中转。
// expiry 是 URL 的有效期。
func (c *Client) PresignedUploadURL(ctx context.Context, storageKey string, expiry time.Duration, contentType string) (string, error) {
	u, err := c.mc.PresignedPutObject(ctx, c.bucket, storageKey, expiry)
	if err != nil {
		return "", fmt.Errorf("storage: presigned upload %s: %w", storageKey, err)
	}
	return u.String(), nil
}

// PresignedDownloadURL 生成预签名下载 URL。
// 生成的 URL 在 expiry 时间内有效，可直接用于浏览器下载/预览。
func (c *Client) PresignedDownloadURL(ctx context.Context, storageKey string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	u, err := c.mc.PresignedGetObject(ctx, c.bucket, storageKey, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("storage: presigned download %s: %w", storageKey, err)
	}
	return u.String(), nil
}

// Delete 从对象存储中删除文件。
func (c *Client) Delete(ctx context.Context, storageKey string) error {
	err := c.mc.RemoveObject(ctx, c.bucket, storageKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("storage: delete %s: %w", storageKey, err)
	}
	return nil
}

// Exists 检查对象是否存在。
func (c *Client) Exists(ctx context.Context, storageKey string) (bool, error) {
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

// Bucket 返回当前使用的存储桶名称。
func (c *Client) Bucket() string { return c.bucket }
