package openapicompat

import "strings"

// appendSubtreeChange keeps the structural change and separately accounts for
// security declarations hidden when the structural diff stops at this object.
func appendSubtreeChange(changes *[]Change, kind, pointer string, before, after any) {
	appendChange(changes, kind, pointer, before, after)
	left, right := map[string]any{}, map[string]any{}
	collectSecurityDeclarations(pointer, before, left)
	collectSecurityDeclarations(pointer, after, right)
	for child, beforeValue := range left {
		if child == pointer {
			continue // Already recorded by the structural diff.
		}
		afterValue, present := right[child]
		if !present {
			appendChange(changes, "remove", child, beforeValue, nil)
		} else if !equalJSON(beforeValue, afterValue) {
			appendChange(changes, "change", child, beforeValue, afterValue)
		}
	}
	for child, afterValue := range right {
		if _, present := left[child]; child != pointer && !present {
			appendChange(changes, "add", child, nil, afterValue)
		}
	}
}

func collectSecurityDeclarations(pointer string, value any, declarations map[string]any) {
	parts := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	if isSecurityDeclaration(parts) {
		declarations[pointer] = value
		return
	}
	if object, ok := value.(map[string]any); ok {
		for key, child := range object {
			collectSecurityDeclarations(pointer+"/"+escapePointer(key), child, declarations)
		}
	}
}

func isSecurityDeclaration(parts []string) bool {
	if len(parts) == 1 && parts[0] == "security" {
		return true
	}
	if len(parts) == 3 && parts[0] == "components" && parts[1] == "securitySchemes" {
		return true
	}
	if len(parts) >= 4 && isOperationPointer(parts[:len(parts)-1]) {
		return parts[len(parts)-1] == "security" || parts[len(parts)-1] == "x-cartulary-state-changing"
	}
	return false
}

func classifySecurityChange(kind, pointer string, before, after any) (Classification, bool) {
	parts := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	for size := 1; size <= len(parts); size++ {
		if isSecurityDeclaration(parts[:size]) {
			// Only a scheme's actual descriptive fields are non-behavioral;
			// a scheme or requirement named "description" is still security.
			if size == 3 && parts[0] == "components" && len(parts) > size {
				if isNonBehavioralPointer("/" + parts[size]) {
					return "", false
				}
				// Scope keys identify authorization capabilities; their string
				// values are descriptions. Only an existing description edit
				// gets this exception, never a scope addition or removal.
				if kind == "change" && len(parts) == 7 && parts[3] == "flows" && parts[5] == "scopes" {
					switch parts[4] {
					case "implicit", "password", "clientCredentials", "authorizationCode":
						_, beforeString := before.(string)
						_, afterString := after.(string)
						if beforeString && afterString {
							return NonBehavioral, true
						}
					}
				}
			}
			return SecuritySensitive, true
		}
	}
	return "", false
}

// Recognize OpenAPI operation locations, including reusable path items and
// callbacks. Schema properties and examples with similar names are unrelated.
func isOperationPointer(parts []string) bool {
	if len(parts) < 3 {
		return false
	}
	switch parts[len(parts)-1] {
	case "get", "put", "post", "delete", "options", "head", "patch", "trace":
	default:
		return false
	}
	if len(parts) == 3 {
		return parts[0] == "webhooks" || (parts[0] == "paths" && strings.HasPrefix(parts[1], "~1"))
	}
	if parts[0] == "components" {
		if len(parts) == 4 && parts[1] == "pathItems" || len(parts) == 5 && parts[1] == "callbacks" {
			return true
		}
	}
	return len(parts) >= 7 && parts[len(parts)-4] == "callbacks" && isOperationPointer(parts[:len(parts)-4])
}
