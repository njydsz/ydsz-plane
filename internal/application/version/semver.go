// 本文件属于 Package version —— SemVer 2.0 语义版本号解析与比较。
//
// 支持完整 SemVer 2.0 语法：major.minor.patch[-prerelease][+build]，
// 校验前导零、非法字符，并提供规范化输出与优先级比较。
package version

import (
	"fmt"
	"strconv"
	"strings"
)

// SemVer 解析后的语义版本。
type SemVer struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Build      string
	Raw        string
}

// ParseSemVer 解析并校验字符串为合法 SemVer 2.0。
// 不允许 leading zeros（如 01.2.3 非法）。
// 成功时返回 (nil, 版本)；失败时返回 (错误信息, nil)。
func ParseSemVer(raw string) (*SemErr, *SemVer) {
	main := raw
	if i := strings.Index(raw, "+"); i >= 0 {
		main = raw[:i]
	}
	mainPart := main
	var pre string
	if i := strings.Index(main, "-"); i >= 0 {
		mainPart = main[:i]
		pre = main[i+1:]
	}
	parts := strings.Split(mainPart, ".")
	if len(parts) != 3 {
		return &SemErr{Reason: "版本号格式必须为 major.minor.patch", Value: raw}, nil
	}
	nums := make([]int, 3)
	for i, p := range parts {
		if p == "" {
			return &SemErr{Reason: "版本号各段不能为空", Value: raw}, nil
		}
		if len(p) > 1 && p[0] == '0' {
			return &SemErr{Reason: fmt.Sprintf("版本号第 %d 段不允许前导零", i+1), Value: raw}, nil
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return &SemErr{Reason: fmt.Sprintf("版本号第 %d 段必须是数字", i+1), Value: raw}, nil
		}
		nums[i] = n
	}
	var build string
	if i := strings.Index(raw, "+"); i >= 0 {
		build = raw[i+1:]
		if build == "" {
			return &SemErr{Reason: "build 元数据不能为空", Value: raw}, nil
		}
	}
	if pre != "" && !isValidSemVerIdentifiers(pre) {
		return &SemErr{Reason: "pre-release 标识符仅允许 [0-9A-Za-z-]", Value: raw}, nil
	}
	if build != "" && !isValidSemVerIdentifiers(build) {
		return &SemErr{Reason: "build 元数据仅允许 [0-9A-Za-z-]", Value: raw}, nil
	}
	return nil, &SemVer{
		Major:      nums[0],
		Minor:      nums[1],
		Patch:      nums[2],
		PreRelease: pre,
		Build:      build,
		Raw:        raw,
	}
}

// isValidSemVerIdentifiers 校验 pre-release / build 标识符仅含
// [0-9A-Za-z-.] 字符。
func isValidSemVerIdentifiers(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

// SemErr 校验失败可读信息。
type SemErr struct {
	Reason string
	Value  string
}

// Error 实现 error 接口；超长输入截断到 64 字符便于日志展示。
func (e *SemErr) Error() string {
	v := e.Value
	if len(v) > 64 {
		v = v[:64] + "…"
	}
	return fmt.Sprintf("%s (输入: %q)", e.Reason, v)
}

// String 返回规范化表达 (不含 build metadata per SemVer spec)。
func (v *SemVer) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		s += "-" + v.PreRelease
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}

// Compare 比较两个 SemVer。返回 -1 / 0 / 1。
func (v *SemVer) Compare(o *SemVer) int {
	if v.Major != o.Major {
		return cmpInt(v.Major, o.Major)
	}
	if v.Minor != o.Minor {
		return cmpInt(v.Minor, o.Minor)
	}
	if v.Patch != o.Patch {
		return cmpInt(v.Patch, o.Patch)
	}
	if v.PreRelease == "" && o.PreRelease == "" {
		return 0
	}
	if v.PreRelease == "" {
		return 1
	}
	if o.PreRelease == "" {
		return -1
	}
	p1 := strings.Split(v.PreRelease, ".")
	p2 := strings.Split(o.PreRelease, ".")
	for i := 0; i < len(p1) && i < len(p2); i++ {
		c := comparePreReleaseID(p1[i], p2[i])
		if c != 0 {
			return c
		}
	}
	return cmpInt(len(p1), len(p2))
}

// cmpInt 返回 a 与 b 的大小关系（-1 / 0 / 1）。
func cmpInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// comparePreReleaseID 按 SemVer 规则比较两个 pre-release 标识符：
// 纯数字按数值比较；数字标识符小于字母标识符；否则按 ASCII 字典序。
func comparePreReleaseID(a, b string) int {
	an, aErr := strconv.Atoi(a)
	bn, bErr := strconv.Atoi(b)
	if aErr == nil && bErr == nil {
		return cmpInt(an, bn)
	}
	if aErr == nil {
		return -1
	}
	if bErr == nil {
		return 1
	}
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
