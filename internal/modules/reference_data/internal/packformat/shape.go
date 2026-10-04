package packformat

import (
	"bytes"
	"encoding/json"
	"regexp"

	"github.com/JochiRaider/cartulary/internal/gen/contractreferencepacks"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func projection(name string) []byte {
	for _, artifact := range contractreferencepacks.Artifacts {
		if artifact.Path == "contracts/reference-packs/"+name {
			return []byte(artifact.JSON)
		}
	}
	panic("reference pack build is missing a required typed projection: " + name)
}

// shape is the private compiled representation of this owner's closed schema
// vocabulary. It is not a general JSON Schema service. Unknown validation
// keywords in trusted projections fail the build initialization rather than
// silently weakening admission.
type shape struct {
	nullable, unique                                       bool
	format                                                 string
	kind                                                   string
	limitID                                                string
	constant                                               any
	hasConstant                                            bool
	enum                                                   []any
	properties                                             map[string]*shape
	additional, propertyNames                              *shape
	minProperties, maxProperties                           *int
	required                                               []string
	closed                                                 bool
	items                                                  *shape
	minimum, maximum                                       *float64
	minLength, maxLength, maxUTF8Bytes, minItems, maxItems *int
	pattern                                                *regexp.Regexp
	not                                                    *shape
	anyOf, oneOf                                           []*shape
}

func compileProjection(name string) *shape {
	return compileProjectionChain(name, map[string]bool{})
}

func compileProjectionChain(name string, ancestors map[string]bool) *shape {
	if ancestors[name] {
		panic("cyclic reference pack schema projection")
	}
	ancestors[name] = true
	defer delete(ancestors, name)
	var value map[string]any
	if err := json.Unmarshal(projection(name), &value); err != nil {
		panic("invalid reference pack schema projection")
	}
	return compileShape(value, ancestors)
}

var projectionReferencePattern = regexp.MustCompile(`^(?:profiles/)?[a-z0-9_.-]+\.schema\.json$`)

func compileShape(value map[string]any, ancestors map[string]bool) *shape {
	if ref, exists := value["$ref"]; exists {
		name, ok := ref.(string)
		if !ok || len(value) != 1 || !projectionReferencePattern.MatchString(name) {
			panic("unsupported reference pack schema reference")
		}
		// References resolve only compiled owner projections. No network,
		// filesystem, fragment, traversal or user-supplied schema is admitted.
		return compileProjectionChain(name, ancestors)
	}
	s := &shape{properties: map[string]*shape{}}
	for key, value := range value {
		switch key {
		case "$schema", "$id", "description":
		case "type":
			if name, ok := value.(string); ok {
				s.kind = name
			} else {
				kinds, ok := value.([]any)
				if !ok || len(kinds) != 2 || kinds[1] != "null" {
					panic("unsupported reference pack type union")
				}
				s.kind = kinds[0].(string)
				s.nullable = true
			}
		case "uniqueItems":
			s.unique = value.(bool)
		case "format":
			s.format = value.(string)
			if s.format != "uuid" && s.format != "date-time" {
				panic("unsupported reference pack string format")
			}
		case "const":
			s.constant = value
			s.hasConstant = true
		case "enum":
			s.enum = value.([]any)
		case "properties":
			for name, child := range value.(map[string]any) {
				s.properties[name] = compileShape(child.(map[string]any), ancestors)
			}
		case "required":
			for _, name := range value.([]any) {
				s.required = append(s.required, name.(string))
			}
		case "additionalProperties":
			if allowed, ok := value.(bool); ok {
				s.closed = !allowed
			} else {
				s.additional = compileShape(value.(map[string]any), ancestors)
			}
		case "propertyNames":
			s.propertyNames = compileShape(value.(map[string]any), ancestors)
		case "minProperties":
			v := int(value.(float64))
			s.minProperties = &v
		case "maxProperties":
			v := int(value.(float64))
			s.maxProperties = &v
		case "items":
			s.items = compileShape(value.(map[string]any), ancestors)
		case "minimum":
			v := value.(float64)
			s.minimum = &v
		case "maximum":
			v := value.(float64)
			s.maximum = &v
		case "minLength":
			v := int(value.(float64))
			s.minLength = &v
		case "x-limitId":
			s.limitID = value.(string)
		case "x-maxUtf8Bytes":
			v := int(value.(float64))
			s.maxUTF8Bytes = &v
		case "maxLength":
			v := int(value.(float64))
			s.maxLength = &v
		case "minItems":
			v := int(value.(float64))
			s.minItems = &v
		case "maxItems":
			v := int(value.(float64))
			s.maxItems = &v
		case "pattern":
			s.pattern = regexp.MustCompile(value.(string))
		case "anyOf":
			for _, child := range value.([]any) {
				s.anyOf = append(s.anyOf, compileShape(child.(map[string]any), ancestors))
			}
		case "not":
			s.not = compileShape(value.(map[string]any), ancestors)
		case "oneOf":
			for _, child := range value.([]any) {
				s.oneOf = append(s.oneOf, compileShape(child.(map[string]any), ancestors))
			}
		default:
			panic("unsupported reference pack schema keyword: " + key)
		}
	}
	return s
}
func equalJSON(a, b any) bool {
	left, e1 := canonicaljson.Marshal(a)
	right, e2 := canonicaljson.Marshal(b)
	return e1 == nil && e2 == nil && bytes.Equal(left, right)
}
func (s *shape) matches(value any) bool {
	return s.walk(value, "$", func(Finding) error { return errShapeMismatch }) == nil
}
