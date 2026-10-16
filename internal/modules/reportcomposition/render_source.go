package reportcomposition

import (
	"bytes"
	"encoding/json"
	"strings"
	"unicode"
	"unicode/utf8"
)

// RenderSource is the companion-owned, validated input projection consumed by
// Reporting. A preview descriptor and an immutable version remain distinct.
type RenderSource struct {
	SchemaID            string
	IncidentID          string
	CompositionID       string
	CompositionVersion  *string
	CompositionSHA256   *string
	TemplateID          string
	TemplateVersion     string
	SourceKind          string
	DraftVersion        *int64
	PreviewSourceSHA256 string
	DeckOps             json.RawMessage
	DiagramDecls        json.RawMessage
	AuthoredTexts       json.RawMessage
}

type RenderSourceError struct{ ReasonCode string }

func (e *RenderSourceError) Error() string { return "reportcomposition: " + e.ReasonCode }

// DecodeRenderSource validates retained bytes before consumers project away
// members. It neither repairs admitted content nor consults mutable draft state.
func DecodeRenderSource(raw []byte) (RenderSource, error) {
	return decodeBoundRenderSource(raw, nil)
}

// Binding is checked after schema admission and before digest or semantic checks.
// All consumers share this decoder so local validation cannot admit a different body.
func decodeBoundRenderSource(raw []byte, binding func(RenderSource) string) (RenderSource, error) {
	invalid := func(reason string) (RenderSource, error) {
		return RenderSource{}, &RenderSourceError{ReasonCode: reason}
	}
	object, err := decodeCompositionObject(raw)
	if err != nil {
		return invalid("composition_schema_invalid")
	}
	canonical, err := canonicalJSON(object)
	if err != nil || !bytes.Equal(raw, canonical) {
		return invalid("composition_schema_invalid")
	}
	schema, _ := object["schema_id"].(string)
	keys := []string{"schema_id", "composition_id", "composition_version", "composition_sha256", "incident_id", "template_id", "template_version", "authored_against_snapshot_id", "deck_ops", "diagram_decls", "authored_texts"}
	if schema == PreviewSourceSchemaID {
		keys = append(keys, "source_kind", "draft_version", "preview_source_sha256")
	} else if schema != CompositionSchemaID {
		return invalid("composition_schema_invalid")
	}
	if len(object) != len(keys) || !objectHasOnly(object, keys...) {
		return invalid("composition_schema_invalid")
	}
	for _, key := range []string{"composition_id", "incident_id", "template_id", "template_version"} {
		if !validCompositionIdentifier(object[key]) {
			return invalid("composition_schema_invalid")
		}
	}
	if object["authored_against_snapshot_id"] != nil && !validCompositionIdentifier(object["authored_against_snapshot_id"]) {
		return invalid("composition_schema_invalid")
	}
	for _, key := range []string{"deck_ops", "diagram_decls", "authored_texts"} {
		if _, ok := object[key].([]any); !ok {
			return invalid("composition_schema_invalid")
		}
		original, err := canonicalJSON(object[key])
		if err != nil {
			return invalid("composition_schema_invalid")
		}
		materialized, err := materializeCompositionArray(original, key)
		if err != nil || !bytes.Equal(original, materialized) {
			return invalid("composition_schema_invalid")
		}
	}
	result := RenderSource{
		SchemaID: schema, IncidentID: object["incident_id"].(string), CompositionID: object["composition_id"].(string),
		TemplateID: object["template_id"].(string), TemplateVersion: object["template_version"].(string),
		DeckOps: mustRawFromAny(object["deck_ops"]), DiagramDecls: mustRawFromAny(object["diagram_decls"]), AuthoredTexts: mustRawFromAny(object["authored_texts"]),
	}
	digestMember := "composition_sha256"
	if schema == PreviewSourceSchemaID {
		digestMember = "preview_source_sha256"
		result.SourceKind, _ = object["source_kind"].(string)
		result.PreviewSourceSHA256, _ = object[digestMember].(string)
		if result.SourceKind == "draft" {
			version, ok := int64Value(object["draft_version"])
			if !ok || version < 1 || version > 9007199254740991 || object["composition_version"] != nil || object["composition_sha256"] != nil {
				return invalid("composition_source_invalid")
			}
			result.DraftVersion = &version
		} else if result.SourceKind != "version" || object["draft_version"] != nil {
			return invalid("composition_source_invalid")
		}
	}
	if schema == CompositionSchemaID || result.SourceKind == "version" {
		version, ok := object["composition_version"].(string)
		sha, shaOK := object["composition_sha256"].(string)
		if !ok || !validCompositionIdentifier(version) || !compositionVersionPattern.MatchString(version) || !shaOK || !sha256HexPattern.MatchString(sha) {
			return invalid("composition_schema_invalid")
		}
		result.CompositionVersion, result.CompositionSHA256 = &version, &sha
	}
	summary := validateDraft(result.DeckOps, result.DiagramDecls, result.AuthoredTexts)
	if summary["stage"] == "schema_validation" {
		return invalid(firstCompositionIssueCode(summary))
	}
	if binding != nil {
		if code := binding(result); code != "" {
			return invalid(code)
		}
	}
	digest, err := compositionDocumentDigest(raw, digestMember)
	if err != nil || object[digestMember] != digest {
		return invalid("composition_digest_mismatch")
	}
	if result.SourceKind == "version" {
		// A version preview must bind the immutable document, not merely a
		// self-consistent preview containing an unrelated composition digest.
		base := make(map[string]any, len(object)-3)
		for key, value := range object {
			if key != "source_kind" && key != "draft_version" && key != "preview_source_sha256" {
				base[key] = value
			}
		}
		base["schema_id"] = CompositionSchemaID
		delete(base, "composition_sha256")
		encoded, err := canonicalJSON(base)
		if err != nil || hashHex(encoded) != *result.CompositionSHA256 {
			return invalid("composition_digest_mismatch")
		}
	}
	if !summaryValid(summary) {
		return invalid(firstCompositionIssueCode(summary))
	}
	return result, nil
}

func validCompositionIdentifier(value any) bool {
	text, ok := value.(string)
	if !ok || !utf8.ValidString(text) || utf8.RuneCountInString(text) < 1 || utf8.RuneCountInString(text) > 128 || strings.TrimSpace(text) != text || strings.ContainsAny(text, "/\\#") {
		return false
	}
	for _, char := range text {
		if unicode.IsControl(char) {
			return false
		}
	}
	return true
}

func sameOptionalString(a, b *string) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}
func sameOptionalInt64(a, b *int64) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}

func firstCompositionIssueCode(summary map[string]any) string {
	issues, _ := summary["issues"].([]map[string]any)
	if len(issues) > 0 {
		if code, ok := issues[0]["code"].(string); ok {
			return code
		}
	}
	return "composition_schema_invalid"
}

func compositionResourceBinding(record ResourceRecord, source RenderSource) string {
	if source.SchemaID != CompositionSchemaID {
		return "composition_source_invalid"
	}
	if source.IncidentID != record.IncidentID.String() {
		return "composition_incident_mismatch"
	}
	if source.CompositionID != record.CompositionID.String() {
		return "composition_id_mismatch"
	}
	if source.TemplateID != record.TemplateID || source.TemplateVersion != record.TemplateVersion {
		return "composition_template_mismatch"
	}
	return ""
}
