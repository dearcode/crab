package str

import (
	"strings"
	"unicode"
)

// TrimSplit 按sep拆分，并去掉空字符.
func TrimSplit(raw, sep string) []string {
	var ss []string
	for _, val := range strings.Split(raw, sep) {
		if s := strings.TrimSpace(val); s != "" {
			ss = append(ss, s)
		}
	}
	return ss
}

// FieldEscape 转换为小写下划线分隔
func FieldEscape(k string) string {
	buf := []byte{}
	up := true
	for _, c := range k {
		if unicode.IsUpper(c) {
			if !up {
				buf = append(buf, '_')
			}
			c += 32
			up = true
		} else {
			up = false
		}

		buf = append(buf, byte(c))
	}
	return string(buf)
}
