package reportcomposition

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
)

var compositionVersionPattern = regexp.MustCompile(`^v([1-9][0-9]*)$`)

func canonicalJSON(value any) ([]byte, error) {
	return json.Marshal(value)
}

func hashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func formatCompositionVersion(version int64) string {
	return "v" + strconv.FormatInt(version, 10)
}

func parseCompositionVersion(value string) (int64, bool) {
	matches := compositionVersionPattern.FindStringSubmatch(value)
	if len(matches) != 2 {
		return 0, false
	}
	parsed, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil || parsed <= 0 {
		return 0, false
	}
	return parsed, true
}

func canonicalComposition(record ResourceRecord, compositionVersion int64) (json.RawMessage, string, error) {
	base := map[string]any{
		"schema_id":                    CompositionSchemaID,
		"composition_id":               record.CompositionID.String(),
		"composition_version":          formatCompositionVersion(compositionVersion),
		"incident_id":                  record.IncidentID.String(),
		"template_id":                  record.TemplateID,
		"template_version":             record.TemplateVersion,
		"authored_against_snapshot_id": optionalStringForJSON(record.AuthoredAgainstSnapshotID),
		"deck_ops":                     rawJSONValue(record.DeckOps),
		"diagram_decls":                rawJSONValue(record.DiagramDecls),
		"authored_texts":               rawJSONValue(record.AuthoredTexts),
	}
	digestBytes, err := canonicalJSON(base)
	if err != nil {
		return nil, "", err
	}
	digest := hashHex(digestBytes)
	base["composition_sha256"] = digest
	canonicalBytes, err := canonicalJSON(base)
	if err != nil {
		return nil, "", err
	}
	if recomputed, err := digestFromCompositionBytes(canonicalBytes); err != nil || recomputed != digest {
		if err != nil {
			return nil, "", err
		}
		return nil, "", fmt.Errorf("composition digest mismatch")
	}
	return json.RawMessage(canonicalBytes), digest, nil
}

func digestFromCompositionBytes(data []byte) (string, error) {
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return "", err
	}
	delete(decoded, "composition_sha256")
	digestBytes, err := canonicalJSON(decoded)
	if err != nil {
		return "", err
	}
	return hashHex(digestBytes), nil
}

func previewSource(resource ResourceRecord, version *VersionRecord, request PreviewRequest) (json.RawMessage, string, *int64, *string, error) {
	var draftVersion *int64
	var compositionVersion any
	var compositionSHA any
	authoredAgainst := resource.AuthoredAgainstSnapshotID
	deckOps := resource.DeckOps
	diagramDecls := resource.DiagramDecls
	authoredTexts := resource.AuthoredTexts
	var compositionSHAPtr *string
	if version == nil {
		draft := resource.DraftVersion
		draftVersion = &draft
		compositionVersion = nil
		compositionSHA = nil
	} else {
		var decoded map[string]any
		if err := json.Unmarshal(version.CanonicalComposition, &decoded); err != nil {
			return nil, "", nil, nil, err
		}
		authoredAgainst = stringPtrFromAny(decoded["authored_against_snapshot_id"])
		deckOps = mustRawFromAny(decoded["deck_ops"])
		diagramDecls = mustRawFromAny(decoded["diagram_decls"])
		authoredTexts = mustRawFromAny(decoded["authored_texts"])
		compositionVersion = formatCompositionVersion(version.CompositionVersion)
		compositionSHA = version.CompositionSHA256
		sha := version.CompositionSHA256
		compositionSHAPtr = &sha
	}
	base := map[string]any{
		"schema_id":                    PreviewSourceSchemaID,
		"source_kind":                  request.SourceKind,
		"incident_id":                  resource.IncidentID.String(),
		"composition_id":               resource.CompositionID.String(),
		"draft_version":                draftVersion,
		"composition_version":          compositionVersion,
		"composition_sha256":           compositionSHA,
		"template_id":                  resource.TemplateID,
		"template_version":             resource.TemplateVersion,
		"authored_against_snapshot_id": optionalStringForJSON(authoredAgainst),
		"deck_ops":                     rawJSONValue(deckOps),
		"diagram_decls":                rawJSONValue(diagramDecls),
		"authored_texts":               rawJSONValue(authoredTexts),
	}
	digestBytes, err := canonicalJSON(base)
	if err != nil {
		return nil, "", nil, nil, err
	}
	digest := hashHex(digestBytes)
	base["preview_source_sha256"] = digest
	canonicalBytes, err := canonicalJSON(base)
	if err != nil {
		return nil, "", nil, nil, err
	}
	return json.RawMessage(canonicalBytes), digest, draftVersion, compositionSHAPtr, nil
}

func validateDraft(deckOps json.RawMessage, diagramDecls json.RawMessage, authoredTexts json.RawMessage, version *VersionRecord) map[string]any {
	issues := []map[string]any{}
	issues = append(issues, duplicateIDIssues(deckOps, "op_id", "composition_op_id")...)
	issues = append(issues, duplicateIDIssues(diagramDecls, "decl_id", "diagram_id")...)
	issues = append(issues, duplicateIDIssues(authoredTexts, "authored_text_id", "authored_text_id")...)
	if version != nil {
		if digest, err := digestFromCompositionBytes(version.CanonicalComposition); err != nil || digest != version.CompositionSHA256 {
			issues = append(issues, issue("composition_digest_mismatch", map[string]any{
				"composition_id":      version.CompositionID.String(),
				"composition_version": formatCompositionVersion(version.CompositionVersion),
			}))
		}
	}
	stage := any(nil)
	if len(issues) > 0 {
		stage = "schema_validation"
	}
	if version != nil && len(issues) > 0 && issues[0]["code"] == "composition_digest_mismatch" {
		stage = "canonical_digest_validation"
	}
	return map[string]any{
		"valid":               len(issues) == 0,
		"stage":               stage,
		"issues":              issues,
		"composition_id":      nil,
		"composition_version": nil,
		"composition_sha256":  nil,
	}
}

func validationSummaryForResource(record ResourceRecord) map[string]any {
	summary := validateDraft(record.DeckOps, record.DiagramDecls, record.AuthoredTexts, nil)
	summary["composition_id"] = record.CompositionID.String()
	return summary
}

func validationSummaryForVersion(version VersionRecord) map[string]any {
	summary := validateDraft(nil, nil, nil, &version)
	summary["composition_id"] = version.CompositionID.String()
	summary["composition_version"] = formatCompositionVersion(version.CompositionVersion)
	summary["composition_sha256"] = version.CompositionSHA256
	return summary
}

func summaryValid(summary map[string]any) bool {
	valid, _ := summary["valid"].(bool)
	return valid
}

func duplicateIDIssues(raw json.RawMessage, idField string, detailKey string) []map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		return []map[string]any{issue("composition_schema_invalid", map[string]any{"field": idField})}
	}
	seen := map[string]struct{}{}
	for _, item := range items {
		id, ok := item[idField].(string)
		if !ok || id == "" {
			return []map[string]any{issue("composition_schema_invalid", map[string]any{"field": idField})}
		}
		if _, ok := seen[id]; ok {
			return []map[string]any{issue("composition_duplicate_id", map[string]any{detailKey: id})}
		}
		seen[id] = struct{}{}
	}
	return nil
}

func issue(code string, details map[string]any) map[string]any {
	return map[string]any{
		"severity":     "error",
		"code":         code,
		"message_key":  "report_composition." + code,
		"safe_details": details,
	}
}

func mustRawFromAny(value any) json.RawMessage {
	encoded, _ := canonicalJSON(value)
	return json.RawMessage(encoded)
}

func stringPtrFromAny(value any) *string {
	parsed, ok := value.(string)
	if !ok {
		return nil
	}
	return &parsed
}
