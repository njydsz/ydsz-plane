// Package auth 认证域的小型辅助函数集合：错误归一化、
// 格式化与 JWT subject 解析，供同包其他文件复用。
package auth

import (
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// pgxErrNoRows 返回 pgx 的"无记录"错误，供调用方统一比较。
func pgxErrNoRows() error { return pgx.ErrNoRows }

// fmtInt 将 int64 格式化为十进制字符串。
func fmtInt(v int64) string { return strconv.FormatInt(v, 10) }

// parseSubject 解析 JWT subject（用户 ID 字符串）为 int64。
// 解析失败返回带上下文的错误。
func parseSubject(sub string) (int64, error) {
	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("auth: bad subject: %w", err)
	}
	return id, nil
}
