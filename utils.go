package jack_compiler

import (
	"regexp"
	"strings"
)

func ContainsInt(slice []int, n int) bool {
	for _, v := range slice {
		if v == n {
			return true
		}
	}
	return false
}

func ContainsString(slice []string, n string) bool {
	for _, v := range slice {
		if v == n {
			return true
		}
	}
	return false
}

func EscapeXml(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "&", "&amp;"), ">", "&gt;"), "<", "&lt;")
}

var stringReg = regexp.MustCompile(`^".*"$`)
var identifierReg = regexp.MustCompile(`^[a-zA-Z]\w*$`)

func isWord(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func isNumber(r rune) bool {
	return r >= '0' && r <= '9'
}

var commentReg = regexp.MustCompile(`(/\*([^*]|[\r\n]|(\*+([^*/]|[\r\n])))*\*+/)|(//.*)`)

func removeComments(line string) string {
	return strings.TrimSpace(commentReg.ReplaceAllString(line, ""))
}
