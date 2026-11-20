package openapiassembly

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	notFoundCodeMember    = "x-cartulary-not-found-code"
	notFoundDetailsMember = "x-cartulary-not-found-details"
)

// preResolutionErrorCodes are the route-resolution rejections of Core 01
// REQ-01-707. They precede resolution and belong to no operation
// (REQ-01-702), so no operation may declare them.
var preResolutionErrorCodes = map[string]struct{}{
	"method_not_allowed": {},
	"route_not_found":    {},
}

// pathParameterFormats are the named formats a path parameter may declare
// instead of a pattern.
var pathParameterFormats = map[string]struct{}{"uuid": {}}

// pathParameterSchemaMembers are the schema members a path parameter may use:
// a string type, its format or pattern with optional length bounds, and
// descriptive members.
var pathParameterSchemaMembers = map[string]struct{}{
	"type": {}, "format": {}, "pattern": {}, "minLength": {}, "maxLength": {},
	"description": {}, "title": {}, "example": {}, "examples": {},
}

// validatePathParameterDeclarations enforces Core 01 REQ-01-707 on the
// assembled document. Every path placeholder of every operation resolves,
// directly or through a component, to a parameter whose schema declares the
// value's format (format uuid, or an anchored pattern that compiles, with
// optional length bounds) and whose x-cartulary-not-found-code names the code
// of the response a malformed value receives. That code must be registered,
// must not be a pre-resolution code, and must be declared by the operation's
// response at the code's registered status. A not-found response that is not
// a 404 declares its fixed details through x-cartulary-not-found-details.
// A component parameter is checked against every operation that uses it.
func validatePathParameterDeclarations(document *orderedValue, loadRegistry func() (map[string]int, error)) error {
	paths, ok := objectMember(document, "paths")
	if !ok {
		return nil
	}
	var registry map[string]int
	for _, path := range paths.object {
		placeholders := pathPlaceholders(path.name)
		if len(placeholders) == 0 || path.value.kind != objectKind {
			continue
		}
		for _, method := range path.value.object {
			if _, ok := operationKeys[method.name]; !ok || method.value.kind != objectKind {
				continue
			}
			label := strings.ToUpper(method.name) + " " + path.name
			if registry == nil {
				var err error
				if registry, err = loadRegistry(); err != nil {
					return err
				}
			}
			parameters, err := resolvedPathParameters(document, path.value, method.value)
			if err != nil {
				return fmt.Errorf("%s: %w", label, err)
			}
			responses, _ := objectMember(method.value, "responses")
			for _, placeholder := range placeholders {
				parameter, declared := parameters[placeholder]
				if !declared {
					return fmt.Errorf("%s: path parameter %q has no declaration", label, placeholder)
				}
				if err := validatePathParameter(document, parameter, responses, registry); err != nil {
					return fmt.Errorf("%s: path parameter %q: %w", label, placeholder, err)
				}
			}
		}
	}
	return nil
}

// resolvedPathParameters returns the operation's path parameters by name. An
// operation-level declaration overrides a path-item declaration of the same
// name, as OpenAPI specifies.
func resolvedPathParameters(document, pathItem, operation *orderedValue) (map[string]*orderedValue, error) {
	resolved := make(map[string]*orderedValue)
	for _, container := range []*orderedValue{pathItem, operation} {
		parameters, ok := objectMember(container, "parameters")
		if !ok {
			continue
		}
		if parameters.kind != arrayKind {
			return nil, errors.New("parameters must be an array")
		}
		for _, parameter := range parameters.array {
			if ref, referenced := objectMember(parameter, "$ref"); referenced && ref.kind == stringKind {
				target, found := resolvePointer(document, ref.scalar.(string))
				if !found {
					return nil, fmt.Errorf("unresolved parameter reference %q", ref.scalar)
				}
				parameter = target
			}
			if parameter.kind != objectKind {
				continue
			}
			name, hasName := objectMember(parameter, "name")
			in, hasIn := objectMember(parameter, "in")
			if hasName && hasIn && name.kind == stringKind && in.kind == stringKind && in.scalar.(string) == "path" {
				resolved[name.scalar.(string)] = parameter
			}
		}
	}
	return resolved, nil
}

func validatePathParameter(document, parameter, responses *orderedValue, registry map[string]int) error {
	if err := validatePathParameterSchema(document, parameter); err != nil {
		return err
	}
	codeValue, ok := objectMember(parameter, notFoundCodeMember)
	if !ok {
		return fmt.Errorf("requires %s", notFoundCodeMember)
	}
	if codeValue.kind != stringKind || codeValue.scalar.(string) == "" {
		return fmt.Errorf("%s must be a non-empty string", notFoundCodeMember)
	}
	code := codeValue.scalar.(string)
	if _, preResolution := preResolutionErrorCodes[code]; preResolution {
		return fmt.Errorf("%s names %q, which belongs to no operation", notFoundCodeMember, code)
	}
	status, registered := registry[code]
	if !registered {
		return fmt.Errorf("%s names unregistered code %q", notFoundCodeMember, code)
	}
	if !responseDeclaresCode(document, responses, status, code) {
		return fmt.Errorf("%s %q must be declared by the operation's %d response", notFoundCodeMember, code, status)
	}
	details, hasDetails := objectMember(parameter, notFoundDetailsMember)
	if hasDetails {
		if details.kind != objectKind || len(details.object) == 0 {
			return fmt.Errorf("%s must be a non-empty object", notFoundDetailsMember)
		}
		for _, member := range details.object {
			if member.name == "" || member.value.kind != stringKind || member.value.scalar.(string) == "" {
				return fmt.Errorf("%s member %q must be a non-empty string", notFoundDetailsMember, member.name)
			}
		}
	}
	if status != 404 && !hasDetails {
		return fmt.Errorf("%s %q has status %d, so the parameter must declare %s", notFoundCodeMember, code, status, notFoundDetailsMember)
	}
	return nil
}

func validatePathParameterSchema(document, parameter *orderedValue) error {
	schema, ok := objectMember(parameter, "schema")
	if !ok {
		return errors.New("requires a schema")
	}
	if ref, referenced := objectMember(schema, "$ref"); referenced && ref.kind == stringKind {
		target, found := resolvePointer(document, ref.scalar.(string))
		if !found {
			return fmt.Errorf("unresolved schema reference %q", ref.scalar)
		}
		schema = target
	}
	if schema.kind != objectKind {
		return errors.New("schema must be an object")
	}
	for _, member := range schema.object {
		if _, allowed := pathParameterSchemaMembers[member.name]; !allowed {
			return fmt.Errorf("schema member %q is not allowed", member.name)
		}
	}
	if typeValue, ok := objectMember(schema, "type"); !ok || typeValue.kind != stringKind || typeValue.scalar.(string) != "string" {
		return errors.New("schema type must be string")
	}
	format, hasFormat := objectMember(schema, "format")
	pattern, hasPattern := objectMember(schema, "pattern")
	switch {
	case hasFormat == hasPattern:
		return errors.New("schema must declare exactly one of format and pattern")
	case hasFormat:
		if format.kind != stringKind {
			return errors.New("schema format must be a string")
		}
		if _, supported := pathParameterFormats[format.scalar.(string)]; !supported {
			return fmt.Errorf("schema format %q is not supported", format.scalar)
		}
	default:
		if pattern.kind != stringKind {
			return errors.New("schema pattern must be a string")
		}
		expression := pattern.scalar.(string)
		if !strings.HasPrefix(expression, "^") || !strings.HasSuffix(expression, "$") || strings.HasSuffix(expression, `\$`) {
			return fmt.Errorf("schema pattern %q must be anchored with ^ and $", expression)
		}
		if _, err := regexp.Compile(expression); err != nil {
			return fmt.Errorf("schema pattern %q does not compile: %w", expression, err)
		}
	}
	bounds := make(map[string]int, 2)
	for _, name := range []string{"minLength", "maxLength"} {
		value, present := objectMember(schema, name)
		if !present {
			continue
		}
		bound, valid := nonNegativeInteger(value)
		if !valid {
			return fmt.Errorf("schema %s must be a non-negative integer", name)
		}
		bounds[name] = bound
	}
	if maximum, hasMaximum := bounds["maxLength"]; hasMaximum && maximum == 0 {
		return errors.New("schema maxLength must be positive")
	}
	if minimum, hasMinimum := bounds["minLength"]; hasMinimum {
		if maximum, hasMaximum := bounds["maxLength"]; hasMaximum && minimum > maximum {
			return errors.New("schema minLength exceeds maxLength")
		}
	}
	return nil
}

func nonNegativeInteger(value *orderedValue) (int, bool) {
	if value.kind != numberKind {
		return 0, false
	}
	text := fmt.Sprint(value.scalar)
	bound, err := strconv.Atoi(text)
	return bound, err == nil && bound >= 0
}

// responseDeclaresCode reports whether the response for status declares code,
// directly or through a component response.
func responseDeclaresCode(document, responses *orderedValue, status int, code string) bool {
	if responses == nil {
		return false
	}
	response, ok := objectMember(responses, strconv.Itoa(status))
	if !ok {
		return false
	}
	if ref, referenced := objectMember(response, "$ref"); referenced && ref.kind == stringKind {
		resolved, found := resolvePointer(document, ref.scalar.(string))
		if !found {
			return false
		}
		response = resolved
	}
	codes, ok := objectMember(response, errorCodesMember)
	if !ok || codes.kind != arrayKind {
		return false
	}
	for _, value := range codes.array {
		if value.kind == stringKind && value.scalar.(string) == code {
			return true
		}
	}
	return false
}
