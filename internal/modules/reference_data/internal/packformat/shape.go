package packformat

import (
	"bytes"
	"encoding/json"
	"regexp"
	"slices"
	"strings"

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

// ProjectionDigest binds producer evidence to the exact compiled owner schema.
func ProjectionDigest(name string) string { return Digest(projection(name)) }

// shape is the private compiled representation of this owner's closed schema
// vocabulary. It is not a general JSON Schema service. Unknown validation
// keywords in trusted projections fail the build initialization rather than
// silently weakening admission.
type shape struct {
	nullable, unique                                       bool
	nativeDateTime                                         bool
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
	anyDispatch, oneDispatch                               *unionDispatch
}

func compileProjection(name string) *shape {
	return compileProjectionChain(name, map[string]bool{})
}

func compileProjectionChain(name string, ancestors map[string]bool) *shape {
	return compileReference(name, "", nil, ancestors)
}

var projectionReferencePattern = regexp.MustCompile(`^(?:profiles/)?[a-z0-9_.-]+\.schema\.json$`)
var definitionReferencePattern = regexp.MustCompile(`^/\$defs/[a-z0-9_-]+$`)

// References resolve exclusively within compiled owner artifacts. The bounded
// definition vocabulary has no network, filesystem, traversal or dynamic IDs.
func compileReference(ref, documentName string, document map[string]any, ancestors map[string]bool) *shape {
	name, fragment, hasFragment := strings.Cut(ref, "#")
	if name != "" {
		if !projectionReferencePattern.MatchString(name) {
			panic("unsupported reference pack schema reference")
		}
		documentName = name
		document = nil
		if err := json.Unmarshal(projection(name), &document); err != nil {
			panic("invalid reference pack schema projection")
		}
	}
	if document == nil || (hasFragment && !definitionReferencePattern.MatchString(fragment)) {
		panic("unsupported reference pack schema definition")
	}
	identity := documentName + "#" + fragment
	if ancestors[identity] {
		panic("cyclic reference pack schema projection")
	}
	ancestors[identity] = true
	defer delete(ancestors, identity)
	value := document
	if hasFragment {
		definitions, ok := document["$defs"].(map[string]any)
		if !ok {
			panic("missing reference pack schema definitions")
		}
		value, ok = definitions[strings.TrimPrefix(fragment, "/$defs/")].(map[string]any)
		if !ok {
			panic("missing reference pack schema definition")
		}
	}
	return compileShapeIn(value, documentName, document, ancestors)
}

func compileShape(value map[string]any, ancestors map[string]bool) *shape {
	return compileShapeIn(value, "inline", value, ancestors)
}

func compileShapeIn(value map[string]any, documentName string, document map[string]any, ancestors map[string]bool) *shape {
	if ref, exists := value["$ref"]; exists {
		name, ok := ref.(string)
		if !ok || len(value) != 1 {
			panic("unsupported reference pack schema reference")
		}
		return compileReference(name, documentName, document, ancestors)
	}

	s := &shape{properties: map[string]*shape{}}
	definition := value
	for key, value := range value {
		switch key {
		case "$schema", "$id", "$comment", "description":
		case "$defs":
			if _, ok := value.(map[string]any); !ok {
				panic("invalid reference pack definitions")
			}
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
			if !slices.Contains([]string{"null", "boolean", "integer", "string", "array", "object"}, s.kind) {
				panic("unsupported reference pack schema type")
			}
		case "uniqueItems":
			s.unique = value.(bool)
		case "x-preserveNativeDateTime":
			if value != true || definition["format"] != "date-time" {
				panic("invalid native timestamp binding")
			}
			s.nativeDateTime = true
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
				s.properties[name] = compileShapeIn(child.(map[string]any), documentName, document, ancestors)
			}
		case "required":
			for _, name := range value.([]any) {
				s.required = append(s.required, name.(string))
			}
		case "additionalProperties":
			if allowed, ok := value.(bool); ok {
				s.closed = !allowed
			} else {
				s.additional = compileShapeIn(value.(map[string]any), documentName, document, ancestors)
			}
		case "propertyNames":
			s.propertyNames = compileShapeIn(value.(map[string]any), documentName, document, ancestors)
		case "minProperties":
			v := int(value.(float64))
			s.minProperties = &v
		case "maxProperties":
			v := int(value.(float64))
			s.maxProperties = &v
		case "items":
			s.items = compileShapeIn(value.(map[string]any), documentName, document, ancestors)
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
				s.anyOf = append(s.anyOf, compileShapeIn(child.(map[string]any), documentName, document, ancestors))
			}
		case "not":
			s.not = compileShapeIn(value.(map[string]any), documentName, document, ancestors)
		case "oneOf":
			for _, child := range value.([]any) {
				s.oneOf = append(s.oneOf, compileShapeIn(child.(map[string]any), documentName, document, ancestors))
			}
		default:
			panic("unsupported reference pack schema keyword: " + key)
		}
	}
	s.anyDispatch = compileUnion(s.anyOf)
	s.oneDispatch = compileUnion(s.oneOf)
	if s.limitID != "" {
		expected, exists := fixedShapeLimits[s.limitID]
		bounds := []*int{s.maxItems, s.maxProperties, s.maxLength, s.maxUTF8Bytes}
		boundCount := 0
		for _, bound := range bounds {
			if bound != nil {
				boundCount++
				if *bound != expected {
					exists = false
				}
			}
		}
		if !exists || boundCount != 1 {
			panic("invalid reference pack schema limit binding: " + s.limitID)
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
