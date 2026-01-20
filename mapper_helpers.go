package errors

import "strings"

func normalizeErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(err.Error()))
}

func containsAny(msg string, matches ...string) bool {
	for _, match := range matches {
		if match == "" {
			continue
		}
		if strings.Contains(msg, match) {
			return true
		}
	}
	return false
}

func containsAll(msg string, matches ...string) bool {
	for _, match := range matches {
		if match == "" {
			continue
		}
		if !strings.Contains(msg, match) {
			return false
		}
	}
	return len(matches) > 0
}
