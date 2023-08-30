package monitor

import "strings"

// spliceStr 拼接字符串
func spliceStr(p ...string) string {
	var b strings.Builder
	l := len(p)
	for i := 0; i < l; i++ {
		b.WriteString(p[i])
	}
	return b.String()
}
