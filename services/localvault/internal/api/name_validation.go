package api

import "regexp"

var validResourceName = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

func isValidResourceName(name string) bool {
	return validResourceName.MatchString(name)
}

const (
	maxResourceNameLen = 128
	maxPayloadBytes    = 1 << 20 // 1 MiB
	maxBulkItems       = 200
)

func validateResourceName(name string) string {
	if name == "" {
		return "name is required"
	}
	if len(name) > maxResourceNameLen {
		return "name exceeds maximum length"
	}
	if !validResourceName.MatchString(name) {
		return "name must match [A-Z_][A-Z0-9_]*"
	}
	return ""
}
