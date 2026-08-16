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
	if asciiPathIdentifier.MatchString(member) {
		return parent + "." + member
	}
	encoded, _ := canonicalJSON(member)
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

func definitionMembers(idKey, outputKey string) map[string]memberSpec {
	return map[string]memberSpec{
		idKey:                  {kind: kindString, required: true},
		"target_scope":         {kind: kindString, required: true},
		"target_kind":          {kind: kindString, required: true},
		"source_field_path":    {kind: kindString, required: true},
		outputKey:              {kind: kindString, required: true},
		"projected_type":       {kind: kindString, required: true},
		"required":             {kind: kindBoolean},
		"default_value":        {kind: kindAny, nullable: true},
		"missing_behavior":     {kind: kindString},
		"source_null_behavior": {kind: kindString},
		"null_output_policy":   {kind: kindString},
		"merge_behavior":       {kind: kindString},
	}
}
