package savedviews

import "strings"

type Scope string

const (
	ScopePrivate Scope = "private"
	ScopeShared  Scope = "shared"
	ScopeSystem  Scope = "system"
)

func ParseScope(value string) (Scope, bool) {
	switch Scope(value) {
	case ScopePrivate, ScopeShared, ScopeSystem:
		return Scope(value), true
	default:
		return "", false
	}
}

func DefaultCreateScope(value *string) (Scope, bool) {
	if value == nil {
		return ScopePrivate, true
	}
	return ParseScope(*value)
}

func IsOrdinaryCreateScope(scope Scope) bool {
	return scope == ScopePrivate || scope == ScopeShared
}

func NormalizeDisplayName(value string) (string, bool) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", false
	}
	return normalized, true
}
