package graphprojection

import (
	"encoding/json"
	"regexp"
)

type jsonKind uint8

const (
	kindAny jsonKind = iota
	kindString
	kindObject
	kindArray
	kindBoolean
	kindNumber
)

type memberSpec struct {
	kind     jsonKind
	required bool
	nullable bool
}

var asciiPathIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func canonicalInputMemberPath(parent, member string) string {
	if len(parent)+len(member)+1 > graphProjectionLimits.MaxValidationMessageBytes {
		return parent
	}
	if asciiPathIdentifier.MatchString(member) {
		return parent + "." + member
	}
	encoded, _ := canonicalJSON(member)
	if len(parent)+len(encoded)+2 > graphProjectionLimits.MaxValidationMessageBytes {
		return parent
	}
	return parent + "[" + string(encoded) + "]"
}

func matchesJSONKind(value any, kind jsonKind) bool {
	switch kind {
	case kindAny:
		return true
	case kindString:
		_, ok := value.(string)
		return ok
	case kindObject:
		_, ok := value.(map[string]any)
		return ok
	case kindArray:
		_, ok := value.([]any)
		return ok
	case kindBoolean:
		_, ok := value.(bool)
		return ok
	case kindNumber:
		_, ok := value.(json.Number)
		return ok
	default:
		return false
	}
}
