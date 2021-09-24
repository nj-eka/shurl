package strutils

import "unicode/utf8"

func Truncate(s string, n int, suffix string) string {
	if len(s) <= n {
		return s
	}
	for !utf8.ValidString(s[:n]) {
		n--
	}
	return s[:n] + suffix
}
