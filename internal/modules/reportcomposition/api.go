package reportcomposition

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

var sha256HexPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func DecodeCreateDraftRequest(reader io.Reader) (CreateDraftRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return CreateDraftRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"client_txn_id":                {},
		"template_id":                  {},
		"template_version":             {},
		"authored_against_snapshot_id": {},
		"deck_ops":                     {},
		"diagram_decls":                {},
		"authored_texts":               {},
	}
	if apiErr := rejectUnknown(raw, allowed); apiErr != nil {
		return CreateDraftRequest{}, apiErr
	}
	request := CreateDraftRequest{
		DeckOps:       cloneRaw(emptyJSONArray),
		DiagramDecls:  cloneRaw(emptyJSONArray),
		AuthoredTexts: cloneRaw(emptyJSONArray),
	}
	clientTxnID, apiErr := requiredString(raw, "client_txn_id")
	if apiErr != nil {
		return CreateDraftRequest{}, apiErr
	}
	request.ClientTxnID = clientTxnID
	templateID, apiErr := requiredString(raw, "template_id")
	if apiErr != nil {
		return CreateDraftRequest{}, apiErr
	}
	request.TemplateID = templateID
	templateVersion, apiErr := requiredString(raw, "template_version")
	if apiErr != nil {
		return CreateDraftRequest{}, apiErr
	}
	request.TemplateVersion = templateVersion
	if value, ok := raw["authored_against_snapshot_id"]; ok {
		request.AuthoredAgainstSnapshotID, apiErr = nullableString(value, "authored_against_snapshot_id")
		if apiErr != nil {
			return CreateDraftRequest{}, apiErr
		}
	}
	if value, ok := raw["deck_ops"]; ok {
		request.DeckOps, apiErr = requiredArray(value, "deck_ops")
		if apiErr != nil {
			return CreateDraftRequest{}, apiErr
		}
	}
	if value, ok := raw["diagram_decls"]; ok {
		request.DiagramDecls, apiErr = requiredArray(value, "diagram_decls")
		if apiErr != nil {
			return CreateDraftRequest{}, apiErr
		}
	}
	if value, ok := raw["authored_texts"]; ok {
		request.AuthoredTexts, apiErr = requiredArray(value, "authored_texts")
		if apiErr != nil {
			return CreateDraftRequest{}, apiErr
		}
	}
	if summary := validateDraft(request.DeckOps, request.DiagramDecls, request.AuthoredTexts, nil); !summaryValid(summary) {
		return CreateDraftRequest{}, validationSummaryError(http.StatusBadRequest, "invalid_request", summary)
	}
	request.Normalized, _ = canonicalJSON(map[string]any{
		"client_txn_id":                request.ClientTxnID,
		"template_id":                  request.TemplateID,
		"template_version":             request.TemplateVersion,
		"authored_against_snapshot_id": optionalStringForJSON(request.AuthoredAgainstSnapshotID),
		"deck_ops":                     rawJSONValue(request.DeckOps),
		"diagram_decls":                rawJSONValue(request.DiagramDecls),
		"authored_texts":               rawJSONValue(request.AuthoredTexts),
	})
	return request, nil
}

func DecodeUpdateDraftRequest(reader io.Reader) (UpdateDraftRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return UpdateDraftRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"client_txn_id":                {},
		"base_draft_version":           {},
		"authored_against_snapshot_id": {},
		"deck_ops":                     {},
		"diagram_decls":                {},
		"authored_texts":               {},
	}
	if apiErr := rejectUnknown(raw, allowed); apiErr != nil {
		return UpdateDraftRequest{}, apiErr
	}
	clientTxnID, apiErr := requiredString(raw, "client_txn_id")
	if apiErr != nil {
		return UpdateDraftRequest{}, apiErr
	}
	base, apiErr := requiredPositiveInt(raw, "base_draft_version")
	if apiErr != nil {
		return UpdateDraftRequest{}, apiErr
	}
	request := UpdateDraftRequest{ClientTxnID: clientTxnID, BaseDraftVersion: base}
	if value, ok := raw["authored_against_snapshot_id"]; ok {
		request.AuthoredAgainstPresent = true
		request.AuthoredAgainstSnapshotID, apiErr = nullableString(value, "authored_against_snapshot_id")
		if apiErr != nil {
			return UpdateDraftRequest{}, apiErr
		}
	}
	if value, ok := raw["deck_ops"]; ok {
		parsed, apiErr := requiredArray(value, "deck_ops")
		if apiErr != nil {
			return UpdateDraftRequest{}, apiErr
		}
		request.DeckOps = &parsed
	}
	if value, ok := raw["diagram_decls"]; ok {
		parsed, apiErr := requiredArray(value, "diagram_decls")
		if apiErr != nil {
			return UpdateDraftRequest{}, apiErr
		}
		request.DiagramDecls = &parsed
	}
	if value, ok := raw["authored_texts"]; ok {
		parsed, apiErr := requiredArray(value, "authored_texts")
		if apiErr != nil {
			return UpdateDraftRequest{}, apiErr
		}
		request.AuthoredTexts = &parsed
	}
	request.Normalized, _ = canonicalJSON(map[string]any{
		"client_txn_id":                request.ClientTxnID,
		"base_draft_version":           request.BaseDraftVersion,
		"authored_against_snapshot_id": optionalStringForJSON(request.AuthoredAgainstSnapshotID),
		"authored_against_present":     request.AuthoredAgainstPresent,
		"deck_ops":                     optionalRawJSONValue(request.DeckOps),
		"diagram_decls":                optionalRawJSONValue(request.DiagramDecls),
		"authored_texts":               optionalRawJSONValue(request.AuthoredTexts),
	})
	return request, nil
}

func DecodeDraftVersionRequest(reader io.Reader) (DraftVersionRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return DraftVersionRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"client_txn_id":      {},
		"base_draft_version": {},
	}
	if apiErr := rejectUnknown(raw, allowed); apiErr != nil {
		return DraftVersionRequest{}, apiErr
	}
	clientTxnID, apiErr := requiredString(raw, "client_txn_id")
	if apiErr != nil {
		return DraftVersionRequest{}, apiErr
	}
	base, apiErr := requiredPositiveInt(raw, "base_draft_version")
	if apiErr != nil {
		return DraftVersionRequest{}, apiErr
	}
	normalized, _ := canonicalJSON(map[string]any{
		"client_txn_id":      clientTxnID,
		"base_draft_version": base,
	})
	return DraftVersionRequest{ClientTxnID: clientTxnID, BaseDraftVersion: base, Normalized: normalized}, nil
}

func DecodeValidateRequest(reader io.Reader) (ValidateRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return ValidateRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"source_kind":         {},
		"composition_version": {},
		"inline_composition":  {},
		"validation_context":  {},
	}
	if apiErr := rejectUnknown(raw, allowed); apiErr != nil {
		return ValidateRequest{}, apiErr
	}
	request := ValidateRequest{SourceKind: SourceKindDraft}
	if value, ok := raw["source_kind"]; ok {
		sourceKind, apiErr := requiredStringValue(value, "source_kind")
		if apiErr != nil {
			return ValidateRequest{}, apiErr
		}
		request.SourceKind = sourceKind
	}
	if value, ok := raw["composition_version"]; ok {
		version, apiErr := requiredCompositionVersionValue(value, "composition_version")
		if apiErr != nil {
			return ValidateRequest{}, apiErr
		}
		request.CompositionVersion = &version
	}
	if value, ok := raw["inline_composition"]; ok {
		if bytesEqualJSONNull(value) {
			return ValidateRequest{}, schemaFieldError("inline_composition", "field_not_nullable")
		}
		parsed, apiErr := requiredObject(value, "inline_composition")
		if apiErr != nil {
			return ValidateRequest{}, apiErr
		}
		request.InlineComposition = &parsed
	}
	if value, ok := raw["validation_context"]; ok {
		request.ValidationContextIs = true
		if bytesEqualJSONNull(value) {
			request.ValidationContext = nil
		} else {
			parsed, apiErr := requiredObject(value, "validation_context")
			if apiErr != nil {
				return ValidateRequest{}, apiErr
			}
			request.ValidationContext = &parsed
		}
	}
	if !validValidateSourceSelection(request) {
		return ValidateRequest{}, validationCodeError(http.StatusBadRequest, "invalid_request", "composition_source_invalid")
	}
	return request, nil
}

func DecodePreviewRequest(reader io.Reader) (PreviewRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return PreviewRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"client_txn_id":                 {},
		"source_kind":                   {},
		"composition_version":           {},
		"snapshot_id":                   {},
		"derivation_version":            {},
		"template_id":                   {},
		"template_version":              {},
		"redaction_profile_id":          {},
		"redaction_profile_version":     {},
		"redaction_profile_sha256":      {},
		"render_environment_profile_id": {},
		"output_kind":                   {},
		"output_options":                {},
		"recipient_partition_refs":      {},
		"graph_projection_refs":         {},
	}
	if apiErr := rejectUnknown(raw, allowed); apiErr != nil {
		return PreviewRequest{}, apiErr
	}
	request := PreviewRequest{
		SourceKind:             SourceKindDraft,
		OutputOptions:          cloneRaw(emptyJSONObject),
		RecipientPartitionRefs: cloneRaw(emptyJSONArray),
		GraphProjectionRefs:    cloneRaw(emptyJSONArray),
	}
	fields := []struct {
		name   string
		target *string
	}{
		{"client_txn_id", &request.ClientTxnID},
		{"snapshot_id", &request.SnapshotID},
		{"derivation_version", &request.DerivationVersion},
		{"template_id", &request.TemplateID},
		{"template_version", &request.TemplateVersion},
		{"redaction_profile_id", &request.RedactionProfileID},
		{"redaction_profile_version", &request.RedactionProfileVersion},
		{"redaction_profile_sha256", &request.RedactionProfileSHA256},
		{"render_environment_profile_id", &request.RenderEnvironmentProfile},
		{"output_kind", &request.OutputKind},
	}
	for _, field := range fields {
		value, apiErr := requiredString(raw, field.name)
		if apiErr != nil {
			return PreviewRequest{}, apiErr
		}
		*field.target = value
	}
	if !sha256HexPattern.MatchString(request.RedactionProfileSHA256) {
		return PreviewRequest{}, schemaFieldError("redaction_profile_sha256", "invalid_value")
	}
	if !slices.Contains([]string{OutputKindSlidev, OutputKindMermaid}, request.OutputKind) {
		return PreviewRequest{}, schemaFieldError("output_kind", "unsupported_output_kind")
	}
	if value, ok := raw["source_kind"]; ok {
		sourceKind, apiErr := requiredStringValue(value, "source_kind")
		if apiErr != nil {
			return PreviewRequest{}, apiErr
		}
		request.SourceKind = sourceKind
	}
	if value, ok := raw["composition_version"]; ok {
		version, apiErr := requiredCompositionVersionValue(value, "composition_version")
		if apiErr != nil {
			return PreviewRequest{}, apiErr
		}
		request.CompositionVersion = &version
	}
	if !validPreviewSourceSelection(request) {
		return PreviewRequest{}, validationCodeError(http.StatusBadRequest, "invalid_request", "composition_source_invalid")
	}
	if value, ok := raw["output_options"]; ok {
		request.OutputOptions, apiErr = requiredObject(value, "output_options")
		if apiErr != nil {
			return PreviewRequest{}, apiErr
		}
	}
	if value, ok := raw["recipient_partition_refs"]; ok {
		request.RecipientPartitionRefs, apiErr = requiredArray(value, "recipient_partition_refs")
		if apiErr != nil {
			return PreviewRequest{}, apiErr
		}
	}
	if !rawArrayEmpty(request.RecipientPartitionRefs) {
		return PreviewRequest{}, validationCodeError(http.StatusBadRequest, "invalid_request", "composition_schema_invalid")
	}
	if value, ok := raw["graph_projection_refs"]; ok {
		request.GraphProjectionRefs, apiErr = requiredArray(value, "graph_projection_refs")
		if apiErr != nil {
			return PreviewRequest{}, apiErr
		}
	}
	if apiErr := validateGraphRefs(request.GraphProjectionRefs); apiErr != nil {
		return PreviewRequest{}, apiErr
	}
	request.Normalized, _ = canonicalJSON(map[string]any{
		"client_txn_id":                 request.ClientTxnID,
		"source_kind":                   request.SourceKind,
		"composition_version":           optionalIntVersionForJSON(request.CompositionVersion),
		"snapshot_id":                   request.SnapshotID,
		"derivation_version":            request.DerivationVersion,
		"template_id":                   request.TemplateID,
		"template_version":              request.TemplateVersion,
		"redaction_profile_id":          request.RedactionProfileID,
		"redaction_profile_version":     request.RedactionProfileVersion,
		"redaction_profile_sha256":      request.RedactionProfileSHA256,
		"render_environment_profile_id": request.RenderEnvironmentProfile,
		"output_kind":                   request.OutputKind,
		"output_options":                rawJSONValue(request.OutputOptions),
		"recipient_partition_refs":      rawJSONValue(request.RecipientPartitionRefs),
		"graph_projection_refs":         rawJSONValue(request.GraphProjectionRefs),
	})
	return request, nil
}

func decodeJSONObject(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err == nil {
		return raw, nil
	}
	reasonCode := "request_not_object"
	if errors.Is(err, httpapi.ErrStrictJSONDuplicateMember) {
		reasonCode = "duplicate_object_member"
	}
	return nil, schemaFieldError("", reasonCode)
}

func rejectUnknown(raw map[string]json.RawMessage, allowed map[string]struct{}) *httpapi.APIError {
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return schemaFieldError(key, "unknown_field")
		}
	}
	return nil
}

func requiredString(raw map[string]json.RawMessage, field string) (string, *httpapi.APIError) {
	value, ok := raw[field]
	if !ok {
		return "", schemaFieldError(field, "missing_required_field")
	}
	return requiredStringValue(value, field)
}

func requiredStringValue(value json.RawMessage, field string) (string, *httpapi.APIError) {
	if bytesEqualJSONNull(value) {
		return "", schemaFieldError(field, "field_not_nullable")
	}
	var parsed string
	if err := json.Unmarshal(value, &parsed); err != nil || strings.TrimSpace(parsed) == "" {
		return "", schemaFieldError(field, "invalid_value")
	}
	return parsed, nil
}

func nullableString(value json.RawMessage, field string) (*string, *httpapi.APIError) {
	if bytesEqualJSONNull(value) {
		return nil, nil
	}
	parsed, apiErr := requiredStringValue(value, field)
	if apiErr != nil {
		return nil, apiErr
	}
	return &parsed, nil
}

func requiredPositiveInt(raw map[string]json.RawMessage, field string) (int64, *httpapi.APIError) {
	value, ok := raw[field]
	if !ok {
		return 0, schemaFieldError(field, "missing_required_field")
	}
	if bytesEqualJSONNull(value) {
		return 0, schemaFieldError(field, "field_not_nullable")
	}
	var parsed int64
	if err := json.Unmarshal(value, &parsed); err != nil || parsed <= 0 {
		return 0, schemaFieldError(field, "invalid_positive_integer")
	}
	return parsed, nil
}

func requiredCompositionVersionValue(value json.RawMessage, field string) (int64, *httpapi.APIError) {
	rawVersion, apiErr := requiredStringValue(value, field)
	if apiErr != nil {
		return 0, apiErr
	}
	version, ok := parseCompositionVersion(rawVersion)
	if !ok {
		return 0, schemaFieldError(field, "invalid_value")
	}
	return version, nil
}

func requiredArray(value json.RawMessage, field string) (json.RawMessage, *httpapi.APIError) {
	if bytesEqualJSONNull(value) {
		return nil, schemaFieldError(field, "field_not_nullable")
	}
	var parsed []any
	if err := json.Unmarshal(value, &parsed); err != nil {
		return nil, schemaFieldError(field, "invalid_value")
	}
	canonical, err := canonicalJSON(parsed)
	if err != nil {
		return nil, internalAPIError(err)
	}
	return json.RawMessage(canonical), nil
}

func requiredObject(value json.RawMessage, field string) (json.RawMessage, *httpapi.APIError) {
	if bytesEqualJSONNull(value) {
		return nil, schemaFieldError(field, "field_not_nullable")
	}
	var parsed map[string]any
	if err := json.Unmarshal(value, &parsed); err != nil || parsed == nil {
		return nil, schemaFieldError(field, "invalid_value")
	}
	canonical, err := canonicalJSON(parsed)
	if err != nil {
		return nil, internalAPIError(err)
	}
	return json.RawMessage(canonical), nil
}

func validValidateSourceSelection(request ValidateRequest) bool {
	switch request.SourceKind {
	case SourceKindDraft:
		return request.CompositionVersion == nil && request.InlineComposition == nil
	case SourceKindVersion:
		return request.CompositionVersion != nil && request.InlineComposition == nil
	case SourceKindInline:
		return request.CompositionVersion == nil && request.InlineComposition != nil
	default:
		return false
	}
}

func validPreviewSourceSelection(request PreviewRequest) bool {
	switch request.SourceKind {
	case SourceKindDraft:
		return request.CompositionVersion == nil
	case SourceKindVersion:
		return request.CompositionVersion != nil
	default:
		return false
	}
}

func validateGraphRefs(raw json.RawMessage) *httpapi.APIError {
	var refs []map[string]any
	if err := json.Unmarshal(raw, &refs); err != nil {
		return schemaFieldError("graph_projection_refs", "invalid_value")
	}
	seen := map[string]struct{}{}
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		value, ok := ref["graph_view_id"].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return validationCodeError(http.StatusBadRequest, "invalid_request", "composition_validation_context_invalid")
		}
		if _, ok := seen[value]; ok {
			return validationCodeError(http.StatusBadRequest, "invalid_request", "composition_validation_context_invalid")
		}
		seen[value] = struct{}{}
		ids = append(ids, value)
	}
	if !sort.StringsAreSorted(ids) {
		return validationCodeError(http.StatusBadRequest, "invalid_request", "composition_validation_context_invalid")
	}
	return nil
}

func bytesEqualJSONNull(value json.RawMessage) bool {
	return strings.TrimSpace(string(value)) == "null"
}

func rawArrayEmpty(value json.RawMessage) bool {
	var parsed []any
	if err := json.Unmarshal(value, &parsed); err != nil {
		return false
	}
	return len(parsed) == 0
}

func cloneRaw(value json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), value...)
}

func optionalStringForJSON(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func optionalIntVersionForJSON(value *int64) any {
	if value == nil {
		return nil
	}
	return formatCompositionVersion(*value)
}

func rawJSONValue(raw json.RawMessage) any {
	var value any
	_ = json.Unmarshal(raw, &value)
	return value
}

func optionalRawJSONValue(raw *json.RawMessage) any {
	if raw == nil {
		return nil
	}
	return rawJSONValue(*raw)
}

func schemaFieldError(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{
		"validation_code": "composition_schema_invalid",
		"reason_code":     reasonCode,
	}
	if field != "" {
		details["field"] = field
	}
	return &httpapi.APIError{Status: http.StatusBadRequest, Code: "invalid_request", Details: details}
}

func validationCodeError(status int, code string, validationCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: status, Code: code, Details: map[string]any{"validation_code": validationCode}}
}

func validationSummaryError(status int, code string, summary map[string]any) *httpapi.APIError {
	return &httpapi.APIError{Status: status, Code: code, Details: map[string]any{"validation_summary": summary}}
}

func internalAPIError(err error) *httpapi.APIError {
	if err == nil {
		err = fmt.Errorf("internal error")
	}
	return &httpapi.APIError{Status: http.StatusInternalServerError, Code: "internal_error", Message: err.Error(), Details: map[string]any{}}
}
