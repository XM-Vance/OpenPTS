// 数据脱敏工具：客户名、地址、电话、客户经理按规则替换字符。
// 用法：MaskName("广州陶瓷工业有限公司") → "广州陶***公司"
package masking

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// MaskName 保留前 2 字 + 末 2 字，中间用 *** 替代。
// 长度 ≤4 时全脱敏（保留首末各 1）。
func MaskName(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	n := len(r)
	if n <= 2 {
		return string(r[0]) + "*"
	}
	if n <= 4 {
		return string(r[0]) + "**" + string(r[n-1])
	}
	return string(r[:2]) + "***" + string(r[n-2:])
}

// MaskManager 经理：只显示姓氏 + ***。
func MaskManager(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	return string(r[0]) + "***"
}

// MaskPhone 11 位手机号脱中间 4 位。
var phoneRe = regexp.MustCompile(`^(1\d{2})\d{4}(\d{4})$`)

func MaskPhone(s string) string {
	s = strings.TrimSpace(s)
	if phoneRe.MatchString(s) {
		return phoneRe.ReplaceAllString(s, "$1****$2")
	}
	if len(s) >= 7 {
		return s[:3] + "****" + s[len(s)-2:]
	}
	return s
}

// MaskLocation 地址脱敏：保留省/市，其余 ***。
func MaskLocation(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	if len(r) <= 4 {
		return MaskName(s)
	}
	return string(r[:4]) + "***"
}

// MaskIDCard 身份证脱中间 8 位。
func MaskIDCard(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 18 {
		return s[:6] + "********" + s[14:]
	}
	if len(s) >= 8 {
		return s[:3] + strings.Repeat("*", len(s)-6) + s[len(s)-3:]
	}
	return s
}

// MaskEmail 邮箱本地部分保留首字符其余替换。
func MaskEmail(s string) string {
	s = strings.TrimSpace(s)
	at := strings.LastIndex(s, "@")
	if at <= 1 {
		return s
	}
	local := s[:at]
	domain := s[at:]
	if utf8.RuneCountInString(local) <= 2 {
		return local[:1] + "*" + domain
	}
	return local[:1] + strings.Repeat("*", utf8.RuneCountInString(local)-2) + local[len(local)-1:] + domain
}
