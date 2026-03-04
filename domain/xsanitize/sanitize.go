package xsanitize

import "strings"

// SanitizeEscapeCharacters replaces escape characters in s with their
// printable equivalents to prevent log injection attacks.
// \n → \n, \r → \r, \t → \t
func SanitizeEscapeCharacters(s string) string {
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}
