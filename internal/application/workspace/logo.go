// Package workspace — 工作空间 Logo 上传服务。
//
// Logo 文件存储在对象存储 (MinIO/S3) 的 workspaces/{wsID}/ 路径下，
// 返回预签名下载 URL 写入 workspaces.logo_url 列。
package workspace

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/storage"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// WorkspaceLogoMaxSize 是 Logo 文件的最大尺寸（5 MB）。
const WorkspaceLogoMaxSize = 5 * 1024 * 1024

// allowedLogoContentTypes 是 MIME 类型白名单（仅图片）。
var allowedLogoContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
	"image/svg":  ".svg",
	"image/svg+xml": ".svg",
}

// LogoService 提供 Logo 上传与删除能力。
type LogoService struct {
	svc *Service
	st  *storage.Client
}

// NewLogoService 构造 LogoService。
func NewLogoService(svc *Service, st *storage.Client) *LogoService {
	return &LogoService{svc: svc, st: st}
}

// SaveLogo 上传 Logo 图片并更新工作空间的 logo_url 字段。
//
// 流程：
//  1. 校验 MIME 类型白名单（仅 image/*）；
//  2. 上传到对象存储 workspaces/{wsID}/logo-{timestamp}{ext}；
//  3. 生成预签名下载 URL（7 天有效期）；
//  4. 更新 workspaces.logo_url。
func (s *LogoService) SaveLogo(ctx context.Context, wsID int64, file io.Reader, size int64, contentType string) (string, error) {
	// --- MIME 类型校验 ---
	ext, ok := allowedLogoContentTypes[contentType]
	if !ok {
		return "", errs.Validation("WORKSPACE.LOGO_INVALID_TYPE",
			fmt.Sprintf("不支持的图片类型 %s，仅支持 JPEG / PNG / GIF / WebP / SVG", contentType))
	}

	// --- 大小校验 ---
	if size <= 0 {
		return "", errs.Validation("WORKSPACE.LOGO_EMPTY", "文件为空")
	}
	if size > WorkspaceLogoMaxSize {
		return "", errs.Validation("WORKSPACE.LOGO_TOO_LARGE",
			fmt.Sprintf("图片不能超过 %d MB", WorkspaceLogoMaxSize/1024/1024))
	}

	// --- 上传到对象存储 ---
	storageKey := fmt.Sprintf("workspaces/%d/logo-%d%s", wsID, time.Now().UnixNano(), ext)
	if err := s.st.Upload(ctx, storageKey, file, size, contentType); err != nil {
		return "", errs.ErrInternal.Wrap(err)
	}

	// --- 生成预签名下载 URL（7 天有效期） ---
	presignedURL, err := s.st.PresignedDownloadURL(ctx, storageKey, 7*24*time.Hour)
	if err != nil {
		// URL 生成失败不阻断上传成功，返回空 URL；上层可选择重试
		presignedURL = ""
	}

	// --- 更新 workspaces.logo_url ---
	_, err = s.svc.Update(ctx, wsID, UpdateInput{LogoURL: &presignedURL})
	if err != nil {
		// DB 更新失败，尝试清理对象存储（避免孤儿文件）
		_ = s.st.Delete(ctx, storageKey)
		return "", err
	}

	return presignedURL, nil
}

// RemoveLogo 清除工作空间 Logo（logo_url 置空，对象存储文件保留为软删除）。
func (s *LogoService) RemoveLogo(ctx context.Context, wsID int64) error {
	empty := ""
	_, err := s.svc.Update(ctx, wsID, UpdateInput{LogoURL: &empty})
	return err
}
