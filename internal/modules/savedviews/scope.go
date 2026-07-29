package savedviews

import (
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

type scope string

const (
	scopePrivate scope = "private"
	scopeShared  scope = "shared"
	scopeSystem  scope = "system"
)

func parseScope(value string) (scope, bool) {
	switch scope(value) {
	case scopePrivate, scopeShared, scopeSystem:
		return scope(value), true
	default:
		return "", false
	}
}

func defaultCreateScope(value *string) (scope, bool) {
	if value == nil {
		return scopePrivate, true
	}
	return parseScope(*value)
}

func isOrdinaryCreateScope(value scope) bool {
	return value == scopePrivate || value == scopeShared
}

func canMutate(record savedViewRecord, actorUserID uuid.UUID, membershipRole string) bool {
	if record.Scope == scopeSystem {
		return false
	}
	if membershipRole == "admin" {
		return true
	}
	return record.OwnerUserID != nil && *record.OwnerUserID == actorUserID
}

func normalizeDisplayName(value string) (string, bool) {
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
