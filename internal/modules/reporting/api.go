package reporting

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	ProfileID = "snapshot_reporting"

	SnapshotsRouteContributionID          = "snapshot_reporting.snapshots_route"
	ReleasesRouteContributionID           = "snapshot_reporting.releases_route"
	ReportCompositionsRouteContributionID = "snapshot_reporting.report_compositions_route"
	RenderExportContributionID            = "snapshot_reporting.render_export"
	RenderExportParticipantID             = "snapshot_reporting.render_export_v1"
	RenderExportParticipantKind           = "snapshot_reporting"
	RenderExportOperationKind             = "emit"
	RenderExportContextSchemaID           = "cartulary.extension_snapshot_reporting_participant_context.v1"
	RenderExportResultSchemaID            = "cartulary.extension_snapshot_reporting_participant_result.v1"
	RenderExportAlgorithmID               = "snapshot_reporting.render_export_v1"
	RenderExportOrderingAlgorithmID       = "materialize_reporting_export_model_v1"
	JobWorkerKind                         = "snapshot_reporting.job_worker_v1"
	SnapshotCreateJobKind                 = "snapshot_reporting.snapshot_create_v1"
	SnapshotCreateOperationKind           = "snapshot_reporting.snapshot_create"
	ReleaseCreateJobKind                  = "snapshot_reporting.release_create_v1"
	ReleaseCreateOperationKind            = "snapshot_reporting.release_create"
	CompositionPreviewJobKind             = "snapshot_reporting.composition_preview_v1"
	CompositionPreviewOperationKind       = "snapshot_reporting.composition_preview"

	DerivationVersion         = "cartulary.reporting_derivation_profile.v1"
	ExportModelSchemaID       = "cartulary.reporting_export_model.v1"
	LegacyDerivationVersion   = "cartulary.snapshot_export_model.v3"
	LegacyExportModelSchemaID = "cartulary.export_model.v3"
	OutputOptionsSchemaID     = "cartulary.reporting_render_request_options.v1"
	SourceBoundaryTokenPrefix = "cartulary.source_boundary.v1:"

	DefaultTemplateID      = "cartulary.report.default"
	DefaultTemplateVersion = "1"

	OutputKindSlidev  = "slidev"
	OutputKindMermaid = "mermaid"

	ReleaseScopeInternalDraft  = "internal_draft"
	ReleaseScopeInternalReview = "internal_review"
	ReleaseScopeExternal       = "external_release"

	ReleaseStatePendingApproval = "pending_approval"
	ReleaseStateApproved        = "approved"
	ReleaseStatePublished       = "published"
	ReleaseStateInvalidated     = "invalidated"
	ReleaseStateRenderFailed    = "render_failed"
)

var (
	sha256HexPattern          = regexp.MustCompile(`^[a-f0-9]{64}$`)
	compositionVersionPattern = regexp.MustCompile(`^v[1-9][0-9]*$`)
	outputKindVocabulary      = []string{
		OutputKindSlidev,
		OutputKindMermaid,
	}
	releaseScopeVocabulary = []string{
		ReleaseScopeInternalDraft,
		ReleaseScopeInternalReview,
		ReleaseScopeExternal,
	}
)

func supportedOutputKinds() []string {
	return append([]string(nil), outputKindVocabulary...)
}

func supportedReleaseScopes() []string {
	return append([]string(nil), releaseScopeVocabulary...)
}

type CreateSnapshotRequest struct {
	IncidentID                   uuid.UUID
	ClientTxnID                  string
	SourceChangeSetHighWatermark *string
	Normalized                   []byte
}

type CreateReleaseRequest struct {
	SnapshotID              uuid.UUID
	ClientTxnID             string
	TemplateID              string
	TemplateVersion         string
	RedactionProfileID      string
	RedactionProfileVersion string
	OutputKind              string
	ReleaseScope            string
	RecipientPartitionRefs  []string
	OutputOptions           json.RawMessage
	GraphProjectionRefs     json.RawMessage
	CompositionJSON         json.RawMessage
	CompositionID           *uuid.UUID
	CompositionVersion      *string
	CompositionSHA256       *string
	Normalized              []byte
}

type ReleaseActionRequest struct {
	ClientTxnID string
	Reason      *string
	Normalized  []byte
}

func DecodeCreateSnapshotRequest(reader io.Reader) (CreateSnapshotRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader, "invalid_snapshot_request")
	if apiErr != nil {
		return CreateSnapshotRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"incident_id":                      {},
		"client_txn_id":                    {},
		"source_change_set_high_watermark": {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return CreateSnapshotRequest{}, invalidSnapshotRequest(key, "unknown_field")
		}
	}
	var request CreateSnapshotRequest
	if value, ok := raw["incident_id"]; !ok {
		return CreateSnapshotRequest{}, invalidSnapshotRequest("incident_id", "missing_required_field")
	} else if bytesEqualJSONNull(value) {
		return CreateSnapshotRequest{}, invalidSnapshotRequest("incident_id", "field_not_nullable")
	} else {
		var rawID string
		if err := json.Unmarshal(value, &rawID); err != nil {
			return CreateSnapshotRequest{}, invalidSnapshotRequest("incident_id", "invalid_value")
		}
		parsed, err := uuid.Parse(rawID)
		if err != nil {
			return CreateSnapshotRequest{}, invalidSnapshotRequest("incident_id", "invalid_value")
		}
		request.IncidentID = parsed
	}
	clientTxnID, apiErr := requiredStringField(raw, "client_txn_id", "invalid_snapshot_request")
	if apiErr != nil {
		return CreateSnapshotRequest{}, apiErr
	}
	request.ClientTxnID = clientTxnID
	if value, ok := raw["source_change_set_high_watermark"]; ok {
		if bytesEqualJSONNull(value) {
			return CreateSnapshotRequest{}, invalidSnapshotRequest("source_change_set_high_watermark", "field_not_nullable")
		}
		var watermark string
		if err := json.Unmarshal(value, &watermark); err != nil || strings.TrimSpace(watermark) == "" {
			return CreateSnapshotRequest{}, invalidSnapshotRequest("source_change_set_high_watermark", "invalid_value")
		}
		request.SourceChangeSetHighWatermark = &watermark
	}
	normalized, err := normalizeSnapshotRequest(request.IncidentID, request.ClientTxnID, request.SourceChangeSetHighWatermark)
	if err != nil {
		return CreateSnapshotRequest{}, internalAPIError(err)
	}
	request.Normalized = normalized
	return request, nil
}

func DecodeCreateReleaseRequest(reader io.Reader) (CreateReleaseRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader, "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"snapshot_id":               {},
		"client_txn_id":             {},
		"template_id":               {},
		"template_version":          {},
		"redaction_profile_id":      {},
		"redaction_profile_version": {},
		"output_kind":               {},
		"release_scope":             {},
		"recipient_partition_refs":  {},
		"output_options":            {},
		"graph_projection_refs":     {},
		"composition_id":            {},
		"composition_version":       {},
		"composition_sha256":        {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return CreateReleaseRequest{}, invalidReleaseRequest(key, "unknown_field")
		}
	}
	var request CreateReleaseRequest
	request.GraphProjectionRefs = json.RawMessage(`[]`)
	if value, ok := raw["snapshot_id"]; !ok {
		return CreateReleaseRequest{}, invalidReleaseRequest("snapshot_id", "missing_required_field")
	} else if bytesEqualJSONNull(value) {
		return CreateReleaseRequest{}, invalidReleaseRequest("snapshot_id", "field_not_nullable")
	} else {
		var rawID string
		if err := json.Unmarshal(value, &rawID); err != nil {
			return CreateReleaseRequest{}, invalidReleaseRequest("snapshot_id", "invalid_value")
		}
		parsed, err := uuid.Parse(rawID)
		if err != nil {
			return CreateReleaseRequest{}, invalidReleaseRequest("snapshot_id", "invalid_value")
		}
		request.SnapshotID = parsed
	}
	clientTxnID, apiErr := requiredStringField(raw, "client_txn_id", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	request.ClientTxnID = clientTxnID
	request.ReleaseScope = ReleaseScopeInternalDraft
	templateID, apiErr := requiredStringField(raw, "template_id", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	request.TemplateID = templateID
	templateVersion, apiErr := requiredStringField(raw, "template_version", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	request.TemplateVersion = templateVersion
	outputKind, apiErr := requiredStringField(raw, "output_kind", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	request.OutputKind = outputKind
	if value, ok := raw["release_scope"]; ok {
		value, apiErr := optionalNonNullString(value, "release_scope", "invalid_release_request")
		if apiErr != nil {
			return CreateReleaseRequest{}, apiErr
		}
		request.ReleaseScope = value
	}
	if value, ok := raw["recipient_partition_refs"]; ok {
		refs, apiErr := optionalStringSet(value, "recipient_partition_refs", "invalid_release_request")
		if apiErr != nil {
			return CreateReleaseRequest{}, apiErr
		}
		request.RecipientPartitionRefs = refs
	}
	if request.RecipientPartitionRefs == nil {
		request.RecipientPartitionRefs = []string{}
	}
	if value, ok := raw["output_options"]; ok {
		request.OutputOptions, apiErr = materializeOutputOptions(value, request.OutputKind, request.ReleaseScope)
		if apiErr != nil {
			return CreateReleaseRequest{}, apiErr
		}
	}
	if request.OutputOptions == nil {
		request.OutputOptions, apiErr = materializeOutputOptions(nil, request.OutputKind, request.ReleaseScope)
		if apiErr != nil {
			return CreateReleaseRequest{}, apiErr
		}
	}
	if value, ok := raw["graph_projection_refs"]; ok {
		request.GraphProjectionRefs, apiErr = optionalJSONArray(value, "graph_projection_refs", "invalid_release_request")
		if apiErr != nil {
			return CreateReleaseRequest{}, apiErr
		}
		if apiErr := validateSourceProjectionRefs(request.GraphProjectionRefs, "invalid_release_request"); apiErr != nil {
			return CreateReleaseRequest{}, apiErr
		}
	}
	compositionID, compositionIDSet, apiErr := optionalUUIDField(raw, "composition_id", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	compositionVersion, compositionVersionSet, apiErr := optionalCompositionVersionField(raw, "composition_version", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	compositionSHA, compositionSHASet, apiErr := optionalSHA256Field(raw, "composition_sha256", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	switch setCount := boolInt(compositionIDSet) + boolInt(compositionVersionSet) + boolInt(compositionSHASet); setCount {
	case 0:
	case 3:
		request.CompositionID = compositionID
		request.CompositionVersion = compositionVersion
		request.CompositionSHA256 = compositionSHA
	default:
		return CreateReleaseRequest{}, invalidReleaseRequest("composition_id", "composition_tuple_incomplete")
	}
	redactionProfileID, apiErr := requiredStringField(raw, "redaction_profile_id", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	request.RedactionProfileID = redactionProfileID
	redactionProfileVersion, apiErr := requiredStringField(raw, "redaction_profile_version", "invalid_release_request")
	if apiErr != nil {
		return CreateReleaseRequest{}, apiErr
	}
	request.RedactionProfileVersion = redactionProfileVersion
	normalized, err := json.Marshal(map[string]any{
		"snapshot_id":               request.SnapshotID.String(),
		"client_txn_id":             request.ClientTxnID,
		"template_id":               request.TemplateID,
		"template_version":          request.TemplateVersion,
		"redaction_profile_id":      request.RedactionProfileID,
		"redaction_profile_version": request.RedactionProfileVersion,
		"output_kind":               request.OutputKind,
		"release_scope":             request.ReleaseScope,
		"recipient_partition_refs":  request.RecipientPartitionRefs,
		"output_options":            rawJSONValue(request.OutputOptions),
		"graph_projection_refs":     rawJSONValue(request.GraphProjectionRefs),
		"composition_id":            optionalUUIDStringForHash(request.CompositionID),
		"composition_version":       optionalStringForHash(request.CompositionVersion),
		"composition_sha256":        optionalStringForHash(request.CompositionSHA256),
	})
	if err != nil {
		return CreateReleaseRequest{}, internalAPIError(err)
	}
	request.Normalized = normalized
	return request, nil
}

func DecodeReleaseActionRequest(reader io.Reader) (ReleaseActionRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader, "invalid_release_request")
	if apiErr != nil {
		return ReleaseActionRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"client_txn_id": {},
		"reason":        {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return ReleaseActionRequest{}, invalidReleaseRequest(key, "unknown_field")
		}
	}
	clientTxnID, apiErr := requiredStringField(raw, "client_txn_id", "invalid_release_request")
	if apiErr != nil {
		return ReleaseActionRequest{}, apiErr
	}
	var reason *string
	if value, ok := raw["reason"]; ok {
		if bytesEqualJSONNull(value) {
			reason = nil
		} else {
			parsed, apiErr := optionalNonNullString(value, "reason", "invalid_release_request")
			if apiErr != nil {
				return ReleaseActionRequest{}, apiErr
			}
			trimmed := strings.TrimSpace(parsed)
			if trimmed != "" {
				reason = &trimmed
			}
		}
	}
	normalized, err := json.Marshal(map[string]any{
		"client_txn_id": clientTxnID,
		"reason":        optionalStringForHash(reason),
	})
	if err != nil {
		return ReleaseActionRequest{}, internalAPIError(err)
	}
	return ReleaseActionRequest{ClientTxnID: clientTxnID, Reason: reason, Normalized: normalized}, nil
}

func decodeJSONObject(reader io.Reader, errorCode string) (map[string]json.RawMessage, *httpapi.APIError) {
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err == nil {
		return raw, nil
	}
	reasonCode := "request_not_object"
	if errors.Is(err, httpapi.ErrStrictJSONDuplicateMember) {
		reasonCode = "duplicate_object_member"
	}
	return nil, &httpapi.APIError{Status: http.StatusBadRequest, Code: errorCode, Details: map[string]any{"reason_code": reasonCode}}
}

func requiredStringField(raw map[string]json.RawMessage, field string, errorCode string) (string, *httpapi.APIError) {
	value, ok := raw[field]
	if !ok {
		return "", invalidRequest(errorCode, field, "missing_required_field")
	}
	if bytesEqualJSONNull(value) {
		return "", invalidRequest(errorCode, field, "field_not_nullable")
	}
	var parsed string
	if err := json.Unmarshal(value, &parsed); err != nil || strings.TrimSpace(parsed) == "" {
		return "", invalidRequest(errorCode, field, "invalid_value")
	}
	return parsed, nil
}

func optionalNonNullString(raw json.RawMessage, field string, errorCode string) (string, *httpapi.APIError) {
	if bytesEqualJSONNull(raw) {
		return "", invalidRequest(errorCode, field, "field_not_nullable")
	}
	var parsed string
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", invalidRequest(errorCode, field, "invalid_value")
	}
	return parsed, nil
}

func optionalStringSet(raw json.RawMessage, field string, errorCode string) ([]string, *httpapi.APIError) {
	if bytesEqualJSONNull(raw) {
		return nil, invalidRequest(errorCode, field, "field_not_nullable")
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, invalidRequest(errorCode, field, "invalid_value")
	}
	seen := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return nil, invalidRequest(errorCode, field, "invalid_value")
		}
		seen[value] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for value := range seen {
		out = append(out, value)
	}
	sort.Strings(out)
	return out, nil
}

func materializeOutputOptions(raw json.RawMessage, outputKind string, releaseScope string) (json.RawMessage, *httpapi.APIError) {
	options := map[string]any{
		"schema_id":         OutputOptionsSchemaID,
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
			return nil, invalidRequest("invalid_release_request", "output_options", "field_not_nullable")
		}
		parsed, err := httpapi.DecodeStrictJSONObject(bytes.NewReader(raw))
		if err != nil {
			reasonCode := "invalid_value"
			if errors.Is(err, httpapi.ErrStrictJSONDuplicateMember) {
				reasonCode = "duplicate_object_member"
			}
			return nil, invalidRequest("invalid_release_request", "output_options", reasonCode)
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
				return nil, invalidRequest("invalid_release_request", "output_options", "unknown_field")
			}
			if key == "schema_id" {
				var schemaID string
				if err := json.Unmarshal(value, &schemaID); err != nil || schemaID != OutputOptionsSchemaID {
					return nil, invalidRequest("invalid_release_request", "output_options", "invalid_value")
				}
				continue
			}
			var parsedBool bool
			if err := json.Unmarshal(value, &parsedBool); err != nil {
				return nil, invalidRequest("invalid_release_request", "output_options", "invalid_value")
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
			return nil, invalidRequest("invalid_release_request", "output_options", "source_only_external_release_invalid")
		}
		for _, key := range []string{"pdf", "svg", "png", "pptx", "rendered_diagrams"} {
			if topLevelSelectorsValid && explicitTrue[key] {
				return nil, invalidRequest("invalid_release_request", "output_options", "source_only_conflict")
			}
			options[key] = false
		}
	}
	if topLevelSelectorsValid {
		if outputKind == OutputKindMermaid && options["pptx"] == true {
			return nil, invalidRequest("invalid_release_request", "output_options", "unsupported_output_option")
		}
		if options["png"] == true || options["pptx"] == true {
			return nil, invalidRequest("invalid_release_request", "output_options", "unsupported_output_option")
		}
		if releaseScope == ReleaseScopeExternal {
			switch {
			case outputKind == OutputKindSlidev && options["pdf"] == false:
				return nil, invalidRequest("invalid_release_request", "output_options", "required_output_omitted")
			case outputKind == OutputKindMermaid && options["svg"] == false:
				return nil, invalidRequest("invalid_release_request", "output_options", "required_output_omitted")
			}
		}
	}
	canonical, err := canonicalJSON(options)
	if err != nil {
		return nil, internalAPIError(err)
	}
	return json.RawMessage(canonical), nil
}

func optionalJSONArray(raw json.RawMessage, field string, errorCode string) (json.RawMessage, *httpapi.APIError) {
	if bytesEqualJSONNull(raw) {
		return nil, invalidRequest(errorCode, field, "field_not_nullable")
	}
	var parsed []any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, invalidRequest(errorCode, field, "invalid_value")
	}
	canonical, err := canonicalJSON(parsed)
	if err != nil {
		return nil, internalAPIError(err)
	}
	return json.RawMessage(canonical), nil
}

func optionalUUIDField(raw map[string]json.RawMessage, field string, errorCode string) (*uuid.UUID, bool, *httpapi.APIError) {
	value, ok := raw[field]
	if !ok || bytesEqualJSONNull(value) {
		return nil, false, nil
	}
	var rawID string
	if err := json.Unmarshal(value, &rawID); err != nil {
		return nil, false, invalidRequest(errorCode, field, "invalid_value")
	}
	parsed, err := uuid.Parse(rawID)
	if err != nil {
		return nil, false, invalidRequest(errorCode, field, "invalid_value")
	}
	return &parsed, true, nil
}

func optionalCompositionVersionField(raw map[string]json.RawMessage, field string, errorCode string) (*string, bool, *httpapi.APIError) {
	value, ok := raw[field]
	if !ok || bytesEqualJSONNull(value) {
		return nil, false, nil
	}
	parsed, apiErr := optionalNonNullString(value, field, errorCode)
	if apiErr != nil {
		return nil, false, apiErr
	}
	if !compositionVersionPattern.MatchString(parsed) {
		return nil, false, invalidRequest(errorCode, field, "invalid_value")
	}
	return &parsed, true, nil
}

func optionalSHA256Field(raw map[string]json.RawMessage, field string, errorCode string) (*string, bool, *httpapi.APIError) {
	value, ok := raw[field]
	if !ok || bytesEqualJSONNull(value) {
		return nil, false, nil
	}
	parsed, apiErr := optionalNonNullString(value, field, errorCode)
	if apiErr != nil {
		return nil, false, apiErr
	}
	if !sha256HexPattern.MatchString(parsed) {
		return nil, false, invalidRequest(errorCode, field, "invalid_value")
	}
	return &parsed, true, nil
}

func validateSourceProjectionRefs(raw json.RawMessage, errorCode string) *httpapi.APIError {
	var refs []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &refs); err != nil {
		return invalidRequest(errorCode, "graph_projection_refs", "invalid_value")
	}
	seen := map[string]struct{}{}
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		allowed := map[string]struct{}{
			"projection_schema_id":     {},
			"graph_view_id":            {},
			"source_snapshot_id":       {},
			"projection_run_id":        {},
			"projection_version":       {},
			"projection_config_digest": {},
			"projection_source_digest": {},
			"projection_output_digest": {},
		}
		for key := range ref {
			if _, ok := allowed[key]; !ok {
				return invalidRequest(errorCode, "graph_projection_refs", "invalid_value")
			}
		}
		requiredStrings := []string{
			"projection_schema_id",
			"graph_view_id",
			"projection_run_id",
			"source_snapshot_id",
			"projection_version",
			"projection_config_digest",
			"projection_source_digest",
			"projection_output_digest",
		}
		values := map[string]string{}
		for _, field := range requiredStrings {
			value, ok := ref[field]
			if !ok {
				return invalidRequest(errorCode, "graph_projection_refs", "invalid_value")
			}
			var parsed string
			if err := json.Unmarshal(value, &parsed); err != nil || strings.TrimSpace(parsed) == "" {
				return invalidRequest(errorCode, "graph_projection_refs", "invalid_value")
			}
			values[field] = parsed
		}
		if values["projection_schema_id"] != "graph_projection.v1" ||
			!sha256HexPattern.MatchString(values["projection_config_digest"]) ||
			!sha256HexPattern.MatchString(values["projection_source_digest"]) ||
			!sha256HexPattern.MatchString(values["projection_output_digest"]) {
			return invalidRequest(errorCode, "graph_projection_refs", "invalid_value")
		}
		graphViewID := values["graph_view_id"]
		if _, ok := seen[graphViewID]; ok {
			return invalidRequest(errorCode, "graph_projection_refs", "graph_projection_ambiguous")
		}
		seen[graphViewID] = struct{}{}
		ids = append(ids, graphViewID)
	}
	if !sort.StringsAreSorted(ids) {
		return invalidRequest(errorCode, "graph_projection_refs", "invalid_value")
	}
	return nil
}

func normalizeSnapshotRequest(incidentID uuid.UUID, clientTxnID string, watermark *string) ([]byte, error) {
	return json.Marshal(map[string]any{
		"incident_id":                      incidentID.String(),
		"client_txn_id":                    clientTxnID,
		"source_change_set_high_watermark": optionalStringForHash(watermark),
	})
}

func optionalStringForHash(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func optionalUUIDStringForHash(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func rawJSONValue(raw json.RawMessage) any {
	var value any
	_ = json.Unmarshal(raw, &value)
	return value
}

func validOutputKind(kind string) bool {
	return slices.Contains(outputKindVocabulary, kind)
}

func validReleaseScope(scope string) bool {
	return slices.Contains(releaseScopeVocabulary, scope)
}

func validateCreateReleaseRequestSemantics(request CreateReleaseRequest) (TemplateContract, *httpapi.APIError) {
	if !validOutputKind(request.OutputKind) {
		return TemplateContract{}, invalidReleaseRequest("output_kind", "unsupported_output_kind")
	}
	if !validReleaseScope(request.ReleaseScope) {
		return TemplateContract{}, invalidReleaseRequest("release_scope", "unsupported_release_scope")
	}
	if request.ReleaseScope != ReleaseScopeExternal && len(request.RecipientPartitionRefs) > 0 {
		return TemplateContract{}, invalidReleaseRequest("recipient_partition_refs", "recipient_partitions_not_allowed")
	}
	templateContract, ok := ResolveTemplateContract(request.TemplateID, request.TemplateVersion)
	if !ok {
		return TemplateContract{}, unsupportedTemplateError(request.TemplateID, request.TemplateVersion)
	}
	if !isSupportedRedactionProfileSelector(request.RedactionProfileID, request.RedactionProfileVersion) {
		return TemplateContract{}, unsupportedRedactionProfileError(request.RedactionProfileID, request.RedactionProfileVersion)
	}
	return templateContract, nil
}

func validateCreateReleaseRecipientPartitions(request CreateReleaseRequest, model ExportModel) *httpapi.APIError {
	if request.ReleaseScope != ReleaseScopeExternal {
		return nil
	}
	if len(request.RecipientPartitionRefs) == 0 {
		return invalidReleaseRequest("recipient_partition_refs", "recipient_partition_profile_mismatch")
	}
	snapshotPartyPartitions := map[string]struct{}{}
	for _, field := range model.RedactionFields() {
		if field.SourceFamily != "party" {
			continue
		}
		for _, ref := range field.DisclosurePartitionRefs {
			if strings.HasPrefix(ref, "party:") && validPartyPartitionRef(ref) {
				snapshotPartyPartitions[ref] = struct{}{}
			}
		}
	}
	for _, ref := range request.RecipientPartitionRefs {
		if !validPartyPartitionRef(ref) {
			return invalidReleaseRequest("recipient_partition_refs", "invalid_recipient_partition_ref")
		}
		if _, ok := snapshotPartyPartitions[ref]; !ok {
			return invalidReleaseRequest("recipient_partition_refs", "unknown_recipient_partition")
		}
	}
	profile, _, err := ResolveRedactionProfile(request.RedactionProfileID, request.RedactionProfileVersion, request.RecipientPartitionRefs)
	if err != nil {
		return invalidReleaseRequest("redaction_profile_id", "unsupported_redaction_profile")
	}
	profilePartyRefs := map[string]struct{}{}
	for _, ref := range profile.AllowedDisclosurePartitionRefs {
		if strings.HasPrefix(ref, "party:") {
			profilePartyRefs[ref] = struct{}{}
		}
	}
	if len(profilePartyRefs) != len(request.RecipientPartitionRefs) {
		return invalidReleaseRequest("recipient_partition_refs", "recipient_partition_profile_mismatch")
	}
	for _, ref := range request.RecipientPartitionRefs {
		if _, ok := profilePartyRefs[ref]; !ok {
			return invalidReleaseRequest("recipient_partition_refs", "recipient_partition_profile_mismatch")
		}
	}
	return nil
}

func validPartyPartitionRef(ref string) bool {
	const prefix = "party:"
	if !strings.HasPrefix(ref, prefix) {
		return false
	}
	segment := strings.TrimPrefix(ref, prefix)
	if segment == "" || !utf8.ValidString(segment) {
		return false
	}
	runeCount := 0
	for _, r := range segment {
		runeCount++
		if r == ':' || r == '/' || r == '\\' || r == '#' || unicode.IsSpace(r) || r == 0 || r < 0x20 || (r >= 0x7f && r <= 0x9f) {
			return false
		}
	}
	return runeCount <= 128
}

func isSupportedRedactionProfileSelector(id string, version string) bool {
	switch id {
	case InternalRedactionProfileID, ExternalRedactionProfileID, TokenizedRedactionProfileID:
		return version == "1"
	default:
		return false
	}
}

func hashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func invalidSnapshotRequest(field string, reasonCode string) *httpapi.APIError {
	return invalidRequest("invalid_snapshot_request", field, reasonCode)
}

func invalidReleaseRequest(field string, reasonCode string) *httpapi.APIError {
	return invalidRequest("invalid_release_request", field, reasonCode)
}

func invalidRequest(code string, field string, reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status: http.StatusBadRequest,
		Code:   code,
		Details: map[string]any{
			"field":       field,
			"reason_code": reasonCode,
		},
	}
}

func clientTxnConflict(clientTxnID string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "client_txn_conflict", Details: map[string]any{"client_txn_id": clientTxnID}}
}

func bytesEqualJSONNull(value json.RawMessage) bool {
	return strings.TrimSpace(string(value)) == "null"
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}

func releaseStateConflict(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "release_state_conflict", Details: map[string]any{"reason_code": reasonCode}}
}

func releaseApprovalRejected(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "release_approval_rejected", Details: map[string]any{"reason_code": reasonCode}}
}

func unsupportedTemplateError(id string, version string) *httpapi.APIError {
	return &httpapi.APIError{
		Status: http.StatusBadRequest,
		Code:   "invalid_release_request",
		Details: map[string]any{
			"field":            "template_id",
			"reason_code":      "unsupported_template",
			"template_id":      id,
			"template_version": version,
		},
	}
}

func unsupportedRedactionProfileError(id string, version string) *httpapi.APIError {
	return &httpapi.APIError{
		Status: http.StatusBadRequest,
		Code:   "invalid_release_request",
		Details: map[string]any{
			"field":                     "redaction_profile_id",
			"reason_code":               "unsupported_redaction_profile",
			"redaction_profile_id":      id,
			"redaction_profile_version": version,
		},
	}
}

func snapshotBoundaryMismatch(expected string, actual string) *httpapi.APIError {
	return &httpapi.APIError{
		Status: http.StatusConflict,
		Code:   "snapshot_source_boundary_conflict",
		Details: map[string]any{
			"expected_source_change_set_high_watermark": expected,
			"current_source_change_set_high_watermark":  actual,
		},
	}
}
