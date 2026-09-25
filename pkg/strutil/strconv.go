// Package strutil 提供字符串工具函数。
package strutil

import "strconv"

// FastItoa 将 int 转换为其十进制字符串表示。
// 使用 strconv.FormatInt 实现，作为全局统一的整数→字符串工具函数，
// 消除各包内重复的 itoa 定义。
func FastItoa(n int) string {
	return strconv.FormatInt(int64(n), 10)
}

// FastItoaInt64 将 int64 转换为其十进制字符串表示。
func FastItoaInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}
