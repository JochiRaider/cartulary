package savedviews

import "github.com/JochiRaider/cartulary/internal/platform/fieldnorm"

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
	normalized, ok := fieldnorm.NormalizeLine(value)
	if !ok || countRunes(normalized) > 256 {
		return "", false
	}
	return normalized, true
}

func countRunes(value string) int {
	count := 0
	for range value {
		count++
	}
	return count
}
