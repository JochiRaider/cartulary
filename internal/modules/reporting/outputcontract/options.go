// Package outputcontract owns Reporting output-option admission and immutable
// consumption for both release and composition-preview routes.
package outputcontract

import (
	"bytes"
	"encoding/json"
	"errors"

	reportingjson "github.com/JochiRaider/cartulary/internal/modules/reporting/canonicaljson"
	strictjson "github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const (
	SchemaID                   = "cartulary.reporting_render_request_options.v1"
	OutputKindSlidev           = "slidev"
	OutputKindMermaid          = "mermaid"
	ReleaseScopeInternalDraft  = "internal_draft"
	ReleaseScopeInternalReview = "internal_review"
	ReleaseScopeExternal       = "external_release"
)

type Options struct {
	SchemaID         string `json:"schema_id"`
	SourceOnly       bool   `json:"source_only"`
	PDF              bool   `json:"pdf"`
	SVG              bool   `json:"svg"`
	PNG              bool   `json:"png"`
	PPTX             bool   `json:"pptx"`
	RenderedDiagrams bool   `json:"rendered_diagrams"`
}

type Error struct{ ReasonCode string }

func (e *Error) Error() string { return "reporting output options: " + e.ReasonCode }

func bytesEqualJSONNull(raw []byte) bool { return bytes.Equal(bytes.TrimSpace(raw), []byte("null")) }
func validOutputKind(kind string) bool   { return kind == OutputKindSlidev || kind == OutputKindMermaid }
func validReleaseScope(scope string) bool {
	return scope == ReleaseScopeInternalDraft || scope == ReleaseScopeInternalReview || scope == ReleaseScopeExternal
}

// Materialize applies route defaults before admission and canonical identity.
func Materialize(raw json.RawMessage, outputKind string, releaseScope string) (json.RawMessage, *Error) {
	options := map[string]any{
		"schema_id":         SchemaID,
		"source_only":       false,
		"pdf":               outputKind == OutputKindSlidev,
		"svg":               outputKind == OutputKindMermaid,
		"png":               false,
		"pptx":              false,
		"rendered_diagrams": true,
	}
	explicitTrue := map[string]bool{}
	if raw != nil {
		if bytesEqualJSONNull(raw) {
			return nil, &Error{ReasonCode: "field_not_nullable"}
		}
		value, err := strictjson.DecodeStrict(raw)
		if err != nil {
			reasonCode := "invalid_value"
			if errors.Is(err, strictjson.ErrDuplicateMember) {
				reasonCode = "duplicate_object_member"
			}
			return nil, &Error{ReasonCode: reasonCode}
		}
		parsed, ok := value.(map[string]any)
		if !ok {
			return nil, &Error{ReasonCode: "invalid_value"}
		}
		allowed := map[string]struct{}{
			"schema_id":         {},
			"source_only":       {},
			"pdf":               {},
			"svg":               {},
			"png":               {},
			"pptx":              {},
			"rendered_diagrams": {},
		}
		for key, value := range parsed {
			if _, ok := allowed[key]; !ok {
				return nil, &Error{ReasonCode: "unknown_field"}
			}
			if value == nil {
				return nil, &Error{ReasonCode: "field_not_nullable"}
			}
			if key == "schema_id" {
				if schemaID, ok := value.(string); !ok || schemaID != SchemaID {
					return nil, &Error{ReasonCode: "invalid_value"}
				}
				continue
			}
			parsedBool, ok := value.(bool)
			if !ok {
				return nil, &Error{ReasonCode: "invalid_value"}
			}
			options[key] = parsedBool
			if parsedBool {
				explicitTrue[key] = true
			}
		}
	}
	sourceOnly, _ := options["source_only"].(bool)
	topLevelSelectorsValid := validOutputKind(outputKind) && validReleaseScope(releaseScope)
	if sourceOnly {
		if topLevelSelectorsValid && releaseScope == ReleaseScopeExternal {
			return nil, &Error{ReasonCode: "source_only_external_release_invalid"}
		}
		for _, key := range []string{"pdf", "svg", "png", "pptx", "rendered_diagrams"} {
			if topLevelSelectorsValid && explicitTrue[key] {
				return nil, &Error{ReasonCode: "source_only_conflict"}
			}
			options[key] = false
		}
	}
	if topLevelSelectorsValid {
		if outputKind == OutputKindMermaid && options["pptx"] == true {
			return nil, &Error{ReasonCode: "unsupported_output_option"}
		}
		if options["png"] == true || options["pptx"] == true {
			return nil, &Error{ReasonCode: "unsupported_output_option"}
		}
		if releaseScope == ReleaseScopeExternal {
			switch {
			case outputKind == OutputKindSlidev && options["pdf"] == false:
				return nil, &Error{ReasonCode: "required_output_omitted"}
			case outputKind == OutputKindMermaid && options["svg"] == false:
				return nil, &Error{ReasonCode: "required_output_omitted"}
			}
		}
	}
	canonical, err := reportingjson.Marshal(options)
	if err != nil {
		return nil, &Error{ReasonCode: "invalid_value"}
	}
	return json.RawMessage(canonical), nil
}

// Render consumers accept only the complete admitted value. Defaults belong to
// request admission and must never repair missing or altered retained input.
func DecodeAdmitted(raw json.RawMessage, kind string, releaseScope string) (Options, *Error) {
	fail := func(reason string) (Options, *Error) {
		return Options{}, &Error{ReasonCode: reason}
	}
	canonical, err := reportingjson.Canonicalize(raw)
	if err != nil {
		return fail("invalid_scalar_value")
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &members); err != nil || members == nil {
		return fail("invalid_type")
	}
	keys := []string{"schema_id", "source_only", "pdf", "svg", "png", "pptx", "rendered_diagrams"}
	allowed := map[string]bool{}
	for _, key := range keys {
		allowed[key] = true
	}
	for key := range members {
		if !allowed[key] {
			return fail("unknown_object_member")
		}
	}
	for _, key := range keys {
		value, present := members[key]
		if !present {
			return fail("required_member_missing")
		}
		if bytesEqualJSONNull(value) {
			return fail("invalid_null")
		}
	}
	var options Options
	if err := json.Unmarshal(canonical, &options); err != nil {
		return fail("invalid_type")
	}
	if options.SchemaID != SchemaID {
		return fail("invalid_schema_id")
	}
	if !validOutputKind(kind) || !validReleaseScope(releaseScope) {
		return fail("invalid_scalar_value")
	}
	validated, apiErr := Materialize(canonical, kind, releaseScope)
	if apiErr != nil {
		return fail(apiErr.ReasonCode)
	}
	if !bytes.Equal(validated, canonical) {
		return fail("invalid_scalar_value")
	}
	return options, nil
}
