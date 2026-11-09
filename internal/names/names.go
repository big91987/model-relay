package names

import (
	"strings"
)

func Valid(value string) bool {
	return len(value) > 0 && len(value) <= 128 && !strings.ContainsAny(value, "\r\n\x00")
}

func Alias(value string) bool {
	if !Valid(value) {
		return false
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || strings.ContainsRune("._-", character)) {
			return false
		}
	}
	return true
}
